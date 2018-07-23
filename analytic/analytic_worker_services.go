package analytic

import (
	"fmt"
	"log"
	"strings"

	"github.com/BaritoLog/go-boilerplate/timekit"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	influx "github.com/go-squads/floodgate-worker/influxdb-handler"
)

const (
	GroupID = "StreamAnalyser"
)

type analyticServices struct {
	clusterConfig cluster.Config
	client        sarama.Client
	brokers       []string
	workerList    map[string]analyticWorker
	database      influx.InfluxDB
	topicList     []string
	brokersConfig sarama.Config
	isClosed      bool
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
	topicList, _ := brokerClient.Topics()
	if err != nil {
		log.Printf("Failed to connect to broker...")
	}
	return analyticServices{
		clusterConfig: analyserClusterConfig,
		client:        brokerClient,
		brokers:       brokers,
		workerList:    make(map[string]analyticWorker),
		database:      influx.NewInfluxService(8086, "localhost", "analyticsKafkaDB", "", ""),
		topicList:     topicList,
		brokersConfig: *brokerConfig,
	}
}

func (a *analyticServices) spawnNewAnalyser(topic string, initialOffset int64) error {
	a.clusterConfig.Config.Consumer.Offsets.Initial = initialOffset
	analyserCluster, err := cluster.NewConsumer(a.brokers, GroupID, []string{topic}, &a.clusterConfig)
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
	topicList, _ := a.client.Topics()
	for _, topic := range topicList {
		if strings.HasSuffix(topic, "_logs") {
			if !a.checkIfTopicAlreadySubscribed(topic) {
				a.spawnNewAnalyser(topic, sarama.OffsetNewest)
				worker := a.workerList[topic]
				worker.Start()
			}
		}
	}
	a.refreshForNewTopics()
	return
}

func (a *analyticServices) refreshForNewTopics() {
	for !a.isClosed {
		newClient, err := setUpClient(a.brokers, &a.brokersConfig)
		if err != nil {
			fmt.Println("Error" + err.Error())
		}
		clientTopics, _ := newClient.Topics()
		if len(a.topicList) != len(clientTopics) {
			for _, topic := range clientTopics {
				exist := a.checkIfTopicAlreadySubscribed(topic)
				if !exist {
					a.spawnNewAnalyserForNewTopic(topic, sarama.OffsetOldest)
					a.topicList = append(a.topicList, topic)
				}
			}
		}
		newClient.Close()
		timekit.Sleep("5s")
	}
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
