package analytic

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Shopify/sarama"
	influx "github.com/go-squads/floodgate-worker/influxdb-handler"
	"github.com/go-squads/floodgate-worker/mailer"
	log "github.com/sirupsen/logrus"
)

type AnalyticWorker interface {
	Start(f ...func(*sarama.ConsumerMessage))
	Stop()
	OnSuccess(f func(*sarama.ConsumerMessage))
}

type analyticWorker struct {
	consumer       ClusterAnalyser
	signalToStop   chan int
	onSuccessFunc  func(*sarama.ConsumerMessage)
	refreshTopics  func()
	databaseClient influx.InfluxDB
	isRunning      bool
	logMap         map[string]string
}

func NewAnalyticWorker(consumer ClusterAnalyser, databaseCon influx.InfluxDB, errorMap map[string]string) *analyticWorker {
	return &analyticWorker{
		consumer:       consumer,
		signalToStop:   make(chan int),
		databaseClient: databaseCon,
		logMap:         errorMap,
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

func (w *analyticWorker) Start(f ...func(*sarama.ConsumerMessage)) {
	if f != nil {
		w.OnSuccess(f[0])
	} else {
		w.OnSuccess(w.storeMessageToDB)
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
			log.Info("Stopped")
			return
		}
	}
}

func (w *analyticWorker) storeMessageToDB(message *sarama.ConsumerMessage) {
	messageVal := make(map[string]interface{})
	_ = json.Unmarshal(message.Value, &messageVal)

	timeToParse, _ := time.Parse(os.Getenv("TIME_LAYOUT"), fmt.Sprint(messageVal["@timestamp"]))
	roundedTime := time.Date(timeToParse.Year(), timeToParse.Month(), timeToParse.Day(),
		timeToParse.Hour(), 0, 0, 0, timeToParse.Location())
	logLevelLabel, exist := w.getLogLabel(messageVal)
	if !exist {
		w.databaseClient.InsertToInflux(message.Topic, UnknownFlag, 1, roundedTime)
	} else {
		value := fmt.Sprint(messageVal[logLevelLabel])
		w.parseAndStoreLogLevel(logLevelLabel, value, message, roundedTime)
	}
	return
}

func (w *analyticWorker) parseAndStoreLogLevel(logLevelLabel string, logLevelValue string, message *sarama.ConsumerMessage, messageTime time.Time) {
	_, exist := w.logMap[logLevelValue]
	if !exist || w.logMap[logLevelValue] == LevelFlag {
		w.databaseClient.InsertToInflux(message.Topic, UnknownFlag, 1, messageTime)
		return
	}

	w.databaseClient.InsertToInflux(message.Topic, w.logMap[logLevelValue], 1, messageTime)

	if w.logMap[logLevelValue] == ErrorFlag {
		methodColumnName, incValue := ConvertMessageToInfluxField(message, logLevelLabel)
		topicErrorName := message.Topic + "_Errors"
		w.databaseClient.InsertToInflux(topicErrorName, methodColumnName, incValue, messageTime)
		w.checkErrorThreshold(message.Topic, messageTime)
	}
	return
}

func (w *analyticWorker) getLogLabel(message map[string]interface{}) (string, bool) {
	keys := make([]string, len(message))
	for k := range message {
		keys = append(keys, k)
	}

	for _, label := range keys {
		logLabel, exist := w.logMap[label]
		if exist && w.logMap[label] == LevelFlag {
			return logLabel, true
		}
	}
	return "", false
}

// Use when log levels = Error
func ConvertMessageToInfluxField(message *sarama.ConsumerMessage, logLabel string) (string, int) {
	messageVal := make(map[string]interface{})
	_ = json.Unmarshal(message.Value, &messageVal)

	delete(messageVal, logLabel)
	delete(messageVal, "@timestamp")
	delete(messageVal, "_ctx")
	var listOfValues []string
	for _, v := range messageVal {
		listOfValues = append(listOfValues, fmt.Sprint(v))
	}

	sort.Strings(listOfValues)
	var columnName string
	for _, v := range listOfValues {
		columnName += "_" + v
	}

	if len(columnName) > 0 {
		return columnName[1:len(columnName)], 1
	} else {
		return "", 0
	}
}

func (w *analyticWorker) checkErrorThreshold(topic string, messageTime time.Time) {
	errorValue := w.databaseClient.GetFieldValueIfExist(ErrorFlag, topic, messageTime)
	trafficValue := w.databaseClient.GetFieldValueIfExist(InfoFlag, topic, messageTime)

	if trafficValue+errorValue > 100 {
		errorPercentage := (errorValue / (trafficValue + errorValue)) * 100
		if errorPercentage >= errorThreshold {
			log.Info("SENT ERROR MAIL NOTIFICATION")
			mailer.SendMail(topic, ErrorFlag)
		}
	}
}

func (w *analyticWorker) checkIfRunning() bool {
	return w.isRunning
}
