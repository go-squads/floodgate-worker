package analytic

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	influx "github.com/go-squads/floodgate-worker/influxdb-handler"

	"github.com/joho/godotenv"
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

	database      influx.InfluxDB
	clusterConfig cluster.Config
	client        sarama.Client
	brokersConfig sarama.Config

	newTopicEventName    string
	lastNewTopic         string
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
		log.Printf("Failed to create a cluster of analyser")
	}
	return analyserCluster, err
}

func (a *analyticServices) SetBrokerAndTopics() error {
	brokerConfig := sarama.NewConfig()
	analyserClusterConfig := a.SetUpConfig()
	brokerClient, err := a.SetUpClient(brokerConfig)
	if err != nil {
		log.Print("Failed to connect to broker...")
	}
	topicList, _ := brokerClient.Topics()
	a.clusterConfig = analyserClusterConfig
	a.topicList = topicList
	a.client = brokerClient
	a.brokersConfig = *brokerConfig
	return err
}

func NewAnalyticServices(brokers []string) AnalyserServices {
	err := godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/go-squads/floodgate-worker/.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	newTopicEventTopic := os.Getenv("NEW_TOPIC_EVENT")

	return &analyticServices{
		brokers:           brokers,
		workerList:        make(map[string]AnalyticWorker),
		database:          connectToInflux(),
		newTopicEventName: newTopicEventTopic,
	}
}

func connectToInflux() influx.InfluxDB {
	influxPort, _ := strconv.Atoi(os.Getenv("INFLUX_PORT"))
	influxHost := os.Getenv("INFLUX_HOST")
	influxDbName := os.Getenv("INFLUX_DB")
	influxUsername := os.Getenv("INFLUX_USERNAME")
	influxPassword := os.Getenv("INFLUX_PASSWORD")
	return influx.NewInfluxService(influxPort, influxHost, influxDbName, influxUsername, influxPassword)
}

func (a *analyticServices) spawnNewAnalyser(topic string, initialOffset int64) error {
	a.clusterConfig.Config.Consumer.Offsets.Initial = initialOffset
	// Mock this part
	analyserCluster, err := a.NewClusterConsumer(os.Getenv("CONSUMER_GROUP_ID"), topic)
	if err != nil {
		log.Printf("Cluster consumer analyser creation failure")
	}
	worker := NewAnalyticWorker(analyserCluster, a.database)
	a.workerList[topic] = worker
	fmt.Println("Spawned worker for " + topic)
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

func (a *analyticServices) spawnNewAnalyserForNewTopic(topic string, messageOffset int64) {
	err := a.spawnNewAnalyser(topic, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Failed to create new worker for new topic")
	}
	newWorker := a.workerList[topic]
	newWorker.Start()
	return
}

func (a *analyticServices) Start() error {
	err := a.SetBrokerAndTopics()
	if err != nil {
		return err
	}
	a.isClosed = false
	err = a.database.InitDB()
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
		log.Printf("Failed to create a topic refresher")
	}

	topicRefresher := NewAnalyticWorker(refresherCluster, a.database)
	fmt.Println("Spawned a topic refresher")
	topicRefresher.Start(a.OnNewTopicEvent)
	return err
}

func (a *analyticServices) OnNewTopicEvent(message *sarama.ConsumerMessage) {
	topicToCheck := string(message.Value)

	_, exist := a.workerList[topicToCheck]
	if exist {
		return
	}

	a.spawnNewAnalyserForNewTopic(topicToCheck, sarama.OffsetOldest)
	a.lastNewTopic = topicToCheck
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
	a.client.Close()
	return
}

func (a *analyticServices) checkIfClosed() bool {
	return a.isClosed
}

func (a *analyticServices) showWorkerList() map[string]AnalyticWorker {
	return a.workerList
}
