package analytic

import (
	"log"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

const (
	GroupID = "StreamAnalyser"
)

type analyticServices struct {
	clusterConfig cluster.Config
	Client        sarama.Client
}

func (a *analyticServices) SetUpConfig() *cluster.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_2_0

	clusterConfig := cluster.NewConfig()
	clusterConfig.Config = *config
	return clusterConfig
}

func (a *analyticServices) SetUpClient(brokers []string, config *sarama.Config) (sarama.Client, error) {
	client, err := sarama.NewClient(brokers, config)
	return client, err
}

// Call NewConsumer
func (a *analyticServices) SpawnAnalyticWorker(brokers []string) {
	brokerConfig := sarama.NewConfig()
	analyserClusterConfig := a.SetUpConfig()
	brokerClient, err := a.SetUpClient(brokers, brokerConfig)
	if err != nil {
		log.Printf("Failed to connect to broker...")
	}
	topicList, err := brokerClient.Topics()
	workerCluster, err := cluster.NewConsumer(brokers, GroupID, topicList, analyserClusterConfig)
	if err != nil {
		log.Printf("Failed to init consumer")
	}
	// use worker made to call NewAnalyticWorker
	worker := NewAnalyticWorker(workerCluster)
	worker.Start()
}
