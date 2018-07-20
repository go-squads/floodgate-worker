package analytic

import (
	"log"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type analyticServices struct {
	Client         *sarama.Client
	Config         *cluster.Config
	AnalyticConfig *AnalyticConfig
	GroupID        string
}

func (a *analyticServices) SetUpConfig() *cluster.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_2_0

	clusterConfig := cluster.NewConfig()
	clusterConfig.Config = *config
	return clusterConfig
}

func (a *analyticServices) SetUpClient(brokers []string, config *sarama.Config) (*sarama.Client, error) {
	client, err := *sarama.NewClient(brokers, config)
	if err != nil {
		log.Printf("Failed to connect to broker")
	}
	return client, err
}
