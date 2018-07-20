package main

import (
	"fmt"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/go-squads/floodgate-worker/analytic"
)

func main() {
	brokers := []string{
		"localhost:9092",
	}
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_2_0

	clusterConfig := cluster.NewConfig()
	clusterConfig.Config = *config

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		fmt.Println(err.Error())
	}
	topicList, err := client.Topics()

	for _, v := range topicList {
		fmt.Println("call function")
		go spawnNewWorker(brokers, clusterConfig, v)
	}
}

func spawnNewWorker(brokers []string, clusterConfig *cluster.Config, topic string) {
	consumer, err := cluster.NewConsumer(brokers, "analytic-testing", []string{topic}, clusterConfig)
	fmt.Println(topic)
	if err != nil {
		fmt.Println(err.Error())
	}
	worker := analytic.NewAnalyticWorker(consumer)
	worker.Start()
	worker.Stop()
}
