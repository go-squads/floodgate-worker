package analytic

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	client "github.com/influxdata/influxdb/client/v2"
)

const (
	GroupID = "StreamAnalyser"
)

type analyticServices struct {
	clusterConfig cluster.Config
	client        sarama.Client
	brokers       []string
	workerList    map[string]analyticWorker
	database      client.Client
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

func connectToInfluxDB() client.Client {
	clientToDB, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: "vin0298",
		Password: "1450Id",
	})

	if err != nil {
		fmt.Println("Failed connection ", err)
	}

	return clientToDB
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
		database:      connectToInfluxDB(),
		topicList:     topicList,
		brokersConfig: *brokerConfig,
	}
}

func (a *analyticServices) spawnNewAnalyser(topic string) error {
	analyserCluster, err := cluster.NewConsumer(a.brokers, GroupID, []string{topic}, &a.clusterConfig)
	if err != nil {
		log.Fatalf("Failed to create a cluster of analyser")
	}
	worker := NewAnalyticWorker(analyserCluster)
	a.workerList[topic] = *worker
	fmt.Println("Spawned worker for " + topic)
	return err
}

// Start
// make a new client
// check for new topics
// make an array of workers - then iterate through it to start it?
func (a *analyticServices) checkIfTopicAlreadySubscribed(topic string) bool {
	if strings.HasSuffix(topic, "_logs") {
		_, exist := a.workerList[topic]
		if !exist {
			return false
		}
	}
	return true
}

func (a *analyticServices) spawnNewAnalyserForNewTopic(topic string) {
	err := a.spawnNewAnalyser(topic)
	if err != nil {
		log.Printf("Failed to create new worker for new topic")
	}
	newWorker := a.workerList[topic]
	newWorker.Start()
	return
}

// Check for new topic when new message with new topic appears
// sarama.CreateTopicsRequest
func (a *analyticServices) Start() {
	a.isClosed = false
	topicList, _ := a.client.Topics()
	for _, topic := range topicList {
		if strings.HasSuffix(topic, "_logs") {
			if !a.checkIfTopicAlreadySubscribed(topic) {
				a.spawnNewAnalyserForNewTopic(topic)
				fmt.Println("Topic: " + topic)
			}
		}
	}
	a.testLoop()
	return
}

func (a *analyticServices) testLoop() {
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
					a.spawnNewAnalyserForNewTopic(topic)
					a.topicList = append(a.topicList, topic)
				}
			}
		}
		newClient.Close()
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

// Parse the Message
// Increment the count
// Get hour : use message.Timestamp.Hour()
func (a *analyticServices) getMessageContent(message *sarama.ConsumerMessage) map[string]string {
	messageContent := make(map[string]string)
	err := json.Unmarshal(message.Value, &messageContent)
	if err != nil {
		log.Printf("Failed conversion")
	}
	return messageContent
}

// func (a *analyticServics) storeToDB(message *sarama.ConsumerMessage) {
// 	messageContent := getMessageContent(message)

// }
