package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/go-squads/floodgate-worker/analytic"
	"github.com/urfave/cli"
)

func main() {

	app := cli.App{
		Name:    "analyser-services",
		Usage:   "Provide kafka producer or consumer for Barito project",
		Version: "0.1.0",
		Commands: []cli.Command{
			{
				Name:      "analyser",
				ShortName: "aw",
				Usage:     "start barito-flow stream analyser",
				Action:    ActionAnalyserService,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}

func ActionAnalyserService(cli *cli.Context) (err error) {
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
		spawnNewWorker(brokers, clusterConfig, v)
	}
	return
}

func spawnNewWorker(brokers []string, clusterConfig *cluster.Config, topic string) {
	consumer, err := cluster.NewConsumer(brokers, "analytic-testing", []string{topic}, clusterConfig)
	fmt.Println(topic)
	if err != nil {
		fmt.Println(err.Error())
	}
	worker := analytic.NewAnalyticWorker(consumer)
	worker.Start()
}
