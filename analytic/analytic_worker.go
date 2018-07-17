package analytic

import (
	cluster "github.com/bsm/sarama-cluster"
)

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
