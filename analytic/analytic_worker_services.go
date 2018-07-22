package analytic

import (
	"log"
	"strings"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

const (
	GroupID = "StreamAnalyser"
)

type analyticServices struct {
	clusterConfig cluster.Config
	client        sarama.Client
	brokers       []string
	workerList    map[string]analyticWorker
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
		log.Printf("Failed to connect to broker...")
	}
	return analyticServices{
		clusterConfig: analyserClusterConfig,
		client:        brokerClient,
		brokers:       brokers,
		workerList:    make(map[string]analyticWorker),
	}
}

func (a *analyticServices) spawnNewAnalyser(topic string) error {
	analyserCluster, err := cluster.NewConsumer(a.brokers, GroupID, []string{topic}, &a.clusterConfig)
	if err != nil {
		log.Fatalf("Failed to create a cluster of analyser")
	}
	worker := NewAnalyticWorker(analyserCluster)
	a.workerList[topic] = *worker
	return err
}

// Start
// make a new client
// check for new topics
// make an array of workers - then iterate through it to start it?
func (a *analyticServices) checkIfTopicAlreadySubscribed(topic string) bool {
	_, exist := a.workerList[topic]
	if !exist {
		return false
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
	topicList, _ := a.client.Topics()
	for _, topic := range topicList {
		if strings.HasSuffix(topic, "_logs") {
			if !a.checkIfTopicAlreadySubscribed(topic) {
				a.spawnNewAnalyserForNewTopic(topic)
			}
		}
	}
	return
}

func (a *analyticServices) Close() {
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
