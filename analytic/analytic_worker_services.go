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

type analyticServices struct {
	isClosed   bool
	brokers    []string
	topicList  []string
	workerList map[string]analyticWorker

	database      influx.InfluxDB
	clusterConfig cluster.Config
	client        sarama.Client
	brokersConfig sarama.Config

	newTopicEventName    string
	lastNewTopic         string
	TopicRefresherWorker AnalyticWorker
}

func setUpConfig() cluster.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_2_0

	clusterConfig := cluster.NewConfig()
	clusterConfig.Config = *config
	return *clusterConfig
}

func setUpClient(brokers []string, config *sarama.Config) (sarama.Client, error) {
	client, err := sarama.NewClient(brokers, config)
	return client, err
}

func NewAnalyticServices(brokers []string) analyticServices {
	brokerConfig := sarama.NewConfig()
	analyserClusterConfig := setUpConfig()
	brokerClient, err := setUpClient(brokers, brokerConfig)
	if err != nil {
		log.Fatal("Failed to connect to broker...")
	}
	topicList, _ := brokerClient.Topics()

	err = godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/go-squads/floodgate-worker/.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	newTopicEventTopic := os.Getenv("NEW_TOPIC_EVENT")

	return analyticServices{
		clusterConfig:     analyserClusterConfig,
		client:            brokerClient,
		brokers:           brokers,
		workerList:        make(map[string]analyticWorker),
		database:          connectToInflux(),
		topicList:         topicList,
		brokersConfig:     *brokerConfig,
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
	analyserCluster, err := cluster.NewConsumer(a.brokers, os.Getenv("CONSUMER_GROUP_ID"), []string{topic}, &a.clusterConfig)
	if err != nil {
		log.Fatalf("Failed to create a cluster of analyser")
	}
	worker := NewAnalyticWorker(analyserCluster, a.database)
	a.workerList[topic] = *worker
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

func (a *analyticServices) Start() {
	a.isClosed = false
	err := a.database.InitDB()
	if err != nil {
		log.Printf("Failed to init InfluxDB")
	}
	err = a.spawnTopicRefresher()
	if err != nil {
		log.Printf("Failure in creating topic refresher")
	}

	topicList, _ := a.client.Topics()
	for _, topic := range topicList {
		if !a.checkIfTopicAlreadySubscribed(topic) {
			a.spawnNewAnalyser(topic, sarama.OffsetNewest)
			worker := a.workerList[topic]
			worker.Start()
		}
	}

	return
}

func (a *analyticServices) spawnTopicRefresher() error {
	a.clusterConfig.Config.Consumer.Offsets.Initial = sarama.OffsetNewest
	refresherCluster, err := cluster.NewConsumer(a.brokers, os.Getenv("TOPIC_REFRESHER_GROUP_ID"),
		[]string{os.Getenv("NEW_TOPIC_EVENT")}, &a.clusterConfig)

	if err != nil {
		log.Fatalf("Failed to create a topic refresher")
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
