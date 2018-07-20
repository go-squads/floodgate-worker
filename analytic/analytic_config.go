package analytic

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type AnalyticConfig interface {
	SetUpConfig() *cluster.Config
	SetUpClient(brokers []string, config *sarama.Config) (*sarama.Client, error)
}
