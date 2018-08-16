package analytic

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
		log "github.com/sirupsen/logrus"
	"github.com/go-squads/floodgate-worker/buffer"
)

type AnalyticWorker interface {
	Start(f ...func(*sarama.ConsumerMessage))
	Stop()
	OnSuccess(f func(*sarama.ConsumerMessage))
}

type analyticWorker struct {
	consumer        ClusterAnalyser
	signalToStop    chan int
	onSuccessFunc   func(*sarama.ConsumerMessage)
	refreshTopics   func()
	isRunning       bool
	logMap          map[string]string
	subscribedTopic string
}


func NewAnalyticWorker(consumer ClusterAnalyser, errorMap map[string]string, topic string) *analyticWorker {
	return &analyticWorker{
		consumer:        consumer,
		signalToStop:    make(chan int),
		logMap:          errorMap,
		subscribedTopic: topic,
	}
}

func (w *analyticWorker) OnSuccess(f func(*sarama.ConsumerMessage)) {
	w.onSuccessFunc = f
}

func (w *analyticWorker) successReadMessage(message *sarama.ConsumerMessage) {
	log.Infof("\nTopic: %s, Partition: %d, Offset: %d, Key: %s, MessageVal: %s,\n",
		message.Topic, message.Partition, message.Offset, message.Key, message.Value)
	if w.onSuccessFunc != nil {
		w.onSuccessFunc(message)
	}
}

func (w *analyticWorker) Start(f ...func(*sarama.ConsumerMessage)) {
	if f != nil {
		w.OnSuccess(f[0])
	} else {
		w.OnSuccess(w.onNewMessage)
	}

	w.isRunning = true
	go w.consumeMessage()
}

func (w *analyticWorker) Stop() {
	if w.consumer != nil {
		w.consumer.Close()
	}

	go func() {
		w.signalToStop <- 1
	}()
}

func (w *analyticWorker) consumeMessage() {
	for {
		select {
		case message, ok := <-w.consumer.Messages():
			if ok {
				w.successReadMessage(message)
				w.consumer.MarkOffset(message, "")
			}
		case <-w.signalToStop:
			w.isRunning = false
			log.Infof("Stopped worker with %v topic", w.subscribedTopic)
			return
		}
	}
}

func (w *analyticWorker) onNewMessage(message *sarama.ConsumerMessage) {
	messageVal := make(map[string]interface{})
	_ = json.Unmarshal(message.Value, &messageVal)
	log.Debugf("storeMessageToDB: %v", messageVal)
	data := buffer.IncomingLog{
		Level:    fmt.Sprint(messageVal["lvl"]),
		Method: fmt.Sprint(messageVal["method"]),
		Path:   fmt.Sprint(messageVal["path"]),
		Code:   fmt.Sprint(messageVal["code"]),
	}
	buffer.GetBuffer().Add(w.subscribedTopic, data)
}

func (w *analyticWorker) checkIfRunning() bool {
	return w.isRunning
}
