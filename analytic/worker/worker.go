package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-squads/floodgate-worker/buffer"
	log "github.com/sirupsen/logrus"
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
	log.Debugf("\nTopic: %s, Partition: %d, Offset: %d, Key: %s, MessageVal: %s,\n",
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
	data := buffer.IncomingLog{
		Level:     fmt.Sprint(messageVal["lvl"]),
		Method:    fmt.Sprint(messageVal["method"]),
		Path:      fmt.Sprint(messageVal["path"]),
		Code:      fmt.Sprint(messageVal["code"]),
		Timestamp: roundTime(messageVal["@timestamp"]).UTC().Format(time.RFC3339),
	}
	buffer.GetBuffer().Add(w.subscribedTopic, data)
}

func (w *analyticWorker) checkIfRunning() bool {
	return w.isRunning
}

func roundTime(timestamp interface{}) time.Time {
	timeToParse, _ := time.Parse(os.Getenv("TIME_LAYOUT"), fmt.Sprint(timestamp))
	roundedTime := time.Date(timeToParse.Year(), timeToParse.Month(), timeToParse.Day(),
		timeToParse.Hour(), timeToParse.Minute(), 0, 0, timeToParse.Location())
	log.Debugf("Parsed Time: %v, %v", timeToParse, roundedTime)
	return roundedTime
}
