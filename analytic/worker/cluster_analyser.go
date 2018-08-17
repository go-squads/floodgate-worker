package worker

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type ClusterAnalyser interface {
	Messages() <-chan *sarama.ConsumerMessage
	Errors() <-chan error
	Notifications() <-chan *cluster.Notification
	MarkOffset(message *sarama.ConsumerMessage, metadata string)
	Close() error
}
