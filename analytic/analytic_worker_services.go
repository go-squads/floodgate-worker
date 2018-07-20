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
		Client:        brokerClient,
	}
}

// Call NewConsumer
