package analytic

import (
	"log"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type AnalyticWorker interface {
	Start()
	Stop()
	IsStart() bool
	OnError(f func(error))
	OnSuccess(f func(*sarama.ConsumerMessage))
	OnNotification(f func(*cluster.Notification))
}

type analyticWorker struct {
	Consumer cluster.Consumer
}

func NewAnalyticWorker(brokers []string, clusterID string,
	topicList []string) (analyticWorker, error) {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	clusterConsumer, err := cluster.NewConsumer(brokers, "analytic", topicList, config)

	if err != nil {
		panic(err)
	}

	return analyticWorker{
		Consumer: *clusterConsumer,
	}, nil
}

func (worker *analyticWorker) loopError() {
	for err := range worker.Consumer.Errors() {
		log.Printf("Error: %s\n", err.Error())
	}
}

func (worker *analyticWorker) checkAndSubscribeTopics(topics []string) {

}
