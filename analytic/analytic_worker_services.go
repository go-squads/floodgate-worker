package analytic

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/go-squads/floodgate-worker/config"
	"github.com/go-squads/floodgate-worker/mongo"
	log "github.com/sirupsen/logrus"
	"github.com/go-squads/floodgate-worker/buffer"
)

type AnalyserServices interface {
	Start() error
	Close()
	SetUpConfig() cluster.Config
	SetUpClient(config *sarama.Config) (sarama.Client, error)
	NewClusterConsumer(groupID string, topic string) (*cluster.Consumer, error)
}

type analyticServices struct {
	isClosed   bool
	brokers    []string
	topicList  []string
	workerList map[string]AnalyticWorker
	errorMap   map[string]string

	database      mongo.Connector
	clusterConfig cluster.Config
	client        sarama.Client
	brokersConfig sarama.Config
	buffer buffer.Buffer
	newTopicEventName    string
	newTopicToCreate     string
	TopicRefresherWorker AnalyticWorker
}

func (a *analyticServices) SetUpConfig() cluster.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_2_0

	clusterConfig := cluster.NewConfig()
	clusterConfig.Config = *config
	return *clusterConfig
}

func (a *analyticServices) SetUpClient(config *sarama.Config) (sarama.Client, error) {
	client, err := sarama.NewClient(a.brokers, config)
	return client, err
}

func (a *analyticServices) NewClusterConsumer(groupID string, topic string) (*cluster.Consumer, error) {
	analyserCluster, err := cluster.NewConsumer(a.brokers, groupID, []string{topic}, &a.clusterConfig)
	if err != nil {
		log.Error("Failed to create a cluster of analyser")
	}
	return analyserCluster, err
}

func (a *analyticServices) SetBrokerAndTopics() error {
	brokerConfig := sarama.NewConfig()
	analyserClusterConfig := a.SetUpConfig()
	brokerClient, err := a.SetUpClient(brokerConfig)
	if err != nil {
		log.Error("Failed to connect to broker...")
	}
	topicList, _ := brokerClient.Topics()
	a.clusterConfig = analyserClusterConfig
	a.topicList = topicList
	a.client = brokerClient
	a.brokersConfig = *brokerConfig
	return err
}

func NewAnalyticServices(brokers []string) AnalyserServices {
	newTopicEventTopic := os.Getenv("NEW_TOPIC_EVENT")
	db, err := mongo.New("mongodb://localhost:27017", "floodgate-worker")
	buffer := buffer.New(db)
	if err != nil {
		log.Fatal(err)
	}

	return &analyticServices{
		brokers:           brokers,
		workerList:        make(map[string]AnalyticWorker),
		database:          db,
		buffer:				buffer,
		newTopicEventName: newTopicEventTopic,
		errorMap:          config.LogLevelMapping(),
	}
}

func (a *analyticServices) spawnNewAnalyser(topic string, initialOffset int64) error {
	a.clusterConfig.Config.Consumer.Offsets.Initial = initialOffset
	// Mock this part
	analyserCluster, err := a.NewClusterConsumer(os.Getenv("CONSUMER_GROUP_ID"), topic)
	if err != nil {
		log.Error("Cluster consumer analyser creation failure")
	}
	worker := NewAnalyticWorker(analyserCluster, a.errorMap, topic)
	a.workerList[topic] = worker
	log.Info("Spawned worker for " + topic)
	return err
}

func (a *analyticServices) checkIfTopicAlreadySubscribed(topic string) bool {
	if strings.HasSuffix(topic, "_logs") {
		_, exist := a.workerList[topic]
		if !exist {
			return false
		}
	}
	return true
}

func (a *analyticServices) spawnNewAnalyserForNewTopic() error {
	err := a.spawnNewAnalyser(a.newTopicToCreate, sarama.OffsetOldest)
	return err
}

func (a *analyticServices) Start() error {
	err := a.SetBrokerAndTopics()
	if err != nil {
		return err
	}
	a.isClosed = false
	if err != nil {
		return err
	}
	err = a.spawnTopicRefresher()
	if err != nil {
		return err
	}

	topicList, _ := a.client.Topics()
	for _, topic := range topicList {
		if !a.checkIfTopicAlreadySubscribed(topic) {
			a.spawnNewAnalyser(topic, sarama.OffsetNewest)
			worker := a.workerList[topic]
			worker.Start()
		}
	}

	return err
}

func (a *analyticServices) spawnTopicRefresher() error {
	a.clusterConfig.Config.Consumer.Offsets.Initial = sarama.OffsetNewest
	refresherCluster, err := a.NewClusterConsumer(os.Getenv("TOPIC_REFRESHER_GROUP_ID"), os.Getenv("NEW_TOPIC_EVENT"))
	if err != nil {
		log.Error("Failed to create a topic refresher")
	}

	topicRefresher := NewAnalyticWorker(refresherCluster, a.errorMap, os.Getenv("NEW_TOPIC_EVENT"))
	log.Info("Spawned a topic refresher")
	topicRefresher.Start(a.OnNewTopicEvent)
	return err
}

func (a *analyticServices) OnNewTopicEvent(message *sarama.ConsumerMessage) {
	topicToCheck := string(message.Value)

	_, exist := a.workerList[topicToCheck]
	if exist {
		return
	}

	a.newTopicToCreate = topicToCheck
	err := a.spawnNewAnalyserForNewTopic()
	if err != nil {
		log.Error("Failed to spawn new topic")
	}
	worker := a.workerList[a.newTopicToCreate]
	worker.Start()
}

func (a *analyticServices) Close() {
	a.isClosed = true
	topicList, _ := a.client.Topics()
	for _, topic := range topicList {
		worker, exist := a.workerList[topic]
		if exist {
			worker.Stop()
		}
	}
	a.buffer.Close()
	a.client.Close()
	return
}

func (a *analyticServices) checkIfClosed() bool {
	return a.isClosed
}

func (a *analyticServices) showWorkerList() map[string]AnalyticWorker {
	return a.workerList
}
