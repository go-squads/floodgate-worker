package analytic

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/go-squads/floodgate-worker/mailer"
	"github.com/go-squads/floodgate-worker/mongo"
	log "github.com/sirupsen/logrus"
)

type AnalyticWorker interface {
	Start(f ...func(*sarama.ConsumerMessage))
	Stop()
	OnSuccess(f func(*sarama.ConsumerMessage))
}

type incomingLog struct {
	lvl    string
	method string
	path   string
	code   string
}

type analyticWorker struct {
	consumer        ClusterAnalyser
	signalToStop    chan int
	onSuccessFunc   func(*sarama.ConsumerMessage)
	refreshTopics   func()
	databaseClient  mongo.Collection
	isRunning       bool
	logMap          map[string]string
	subscribedTopic string
	mailerService   mailer.MailerService
}

var logCount map[incomingLog]int

func NewAnalyticWorker(consumer ClusterAnalyser, databaseCon mongo.Collection, errorMap map[string]string, topic string) *analyticWorker {
	return &analyticWorker{
		consumer:        consumer,
		signalToStop:    make(chan int),
		databaseClient:  databaseCon,
		logMap:          errorMap,
		subscribedTopic: topic,
		mailerService:   mailer.NewMailerService(),
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
	logCount = make(map[incomingLog]int)
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
	data := incomingLog{
		lvl:    fmt.Sprint(messageVal["lvl"]),
		method: fmt.Sprint(messageVal["method"]),
		path:   fmt.Sprint(messageVal["path"]),
		code:   fmt.Sprint(messageVal["code"]),
	}
	logCount[data]++
	log.Debugf("%v: %v", data, logCount[data])
}

func (w *analyticWorker) checkIfRunning() bool {
	return w.isRunning
}
