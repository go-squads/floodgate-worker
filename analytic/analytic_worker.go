package analytic

import (
	"fmt"
	"os"

	"github.com/Shopify/sarama"
)

type AnalyticWorker interface {
	OnSuccess(f func(*sarama.ConsumerMessage))
}

type analyticWorker struct {
	consumer      ClusterAnalyser
	signalToStop  chan int
	onSuccessFunc func(*sarama.ConsumerMessage)
}

func NewAnalyticWorker(consumer ClusterAnalyser) *analyticWorker {
	return &analyticWorker{
		consumer:     consumer,
		signalToStop: make(chan int),
	}
}

func (w *analyticWorker) OnSuccess(f func(*sarama.ConsumerMessage)) {
	w.onSuccessFunc = f
}

func (w *analyticWorker) successReadMessage(message *sarama.ConsumerMessage) {
	fmt.Fprintf(os.Stdout, "\nTopic: %s, Partition: %d, Offset: %d, Key: %s, MessageVal: %s,\n",
		message.Topic, message.Partition, message.Offset, message.Key, message.Value)
	if w.onSuccessFunc != nil {
		w.onSuccessFunc(message)
	}
}

func (w *analyticWorker) Start() {
	go w.consumeMessage()
}

func (w *analyticWorker) Stop() {
	if w.consumer != nil {
		w.consumer.Close()
	}

	go func() {
		fmt.Println("sent stop signal")
		w.signalToStop <- 1
	}()
}

func (w *analyticWorker) consumeMessage() {
	for {
		fmt.Println("Looking for logs:..")
		select {
		case message, ok := <-w.consumer.Messages():
			if ok {
				w.successReadMessage(message)
				w.consumer.MarkOffset(message, "")
			}
		case <-w.signalToStop:
			return
		}
	}
}
