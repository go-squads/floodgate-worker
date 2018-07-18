package analytic

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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
	consumer cluster.Consumer
	signals  chan int
	isStart  bool
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

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	return analyticWorker{
		consumer: *clusterConsumer,
	}, nil
}

func (worker *analyticWorker) loopError() {
	for err := range worker.consumer.Errors() {
		log.Printf("Error: %s\n", err.Error())
	}
}

func (worker *analyticWorker) loopNotification() {
	for notification := range worker.consumer.Notifications() {
		log.Printf("Rebalanced: %+v\n", notification)
	}
}

func (w *analyticWorker) loopMain() {
	fmt.Println("here we go looping the main")
	w.isStart = true
	for {
		fmt.Println("loopmain new worker")
		select {
		case message, ok := <-w.consumer.Messages():
			if ok {
				fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\t%s\n", message.Topic, message.Partition, message.Offset, message.Key, message.Value)
				w.consumer.MarkOffset(message, "")
			}
		case <-w.signals:
			fmt.Println("SIGNALS: ")
			w.isStart = false
			return
		}
	}
}

func (w *analyticWorker) Start() {
	fmt.Println("start the worker")
	go w.loopError()
	go w.loopNotification()
	go w.loopMain()
	fmt.Println("end of start")
}
