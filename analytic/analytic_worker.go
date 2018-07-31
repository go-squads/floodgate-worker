package analytic

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Shopify/sarama"
	influx "github.com/go-squads/floodgate-worker/influxdb-handler"
)

type AnalyticWorker interface {
	OnSuccess(f func(*sarama.ConsumerMessage))
}

type analyticWorker struct {
	consumer       ClusterAnalyser
	signalToStop   chan int
	onSuccessFunc  func(*sarama.ConsumerMessage)
	refreshTopics  func()
	databaseClient influx.InfluxDB
}

func NewAnalyticWorker(consumer ClusterAnalyser, databaseCon influx.InfluxDB) *analyticWorker {
	return &analyticWorker{
		consumer:       consumer,
		signalToStop:   make(chan int),
		databaseClient: databaseCon,
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

func (w *analyticWorker) storeMessageToDB(message *sarama.ConsumerMessage) {
	timeVal := make(map[string]string)
	_ = json.Unmarshal(message.Value, &timeVal)
	timeToParse, _ := time.Parse(os.Getenv("TIME_LAYOUT"), timeVal["@timestamp"])
	columnName, value := ConvertMessageToInfluxField(message)
	fmt.Println(columnName)

	roundedTime := time.Date(timeToParse.Year(), timeToParse.Month(), timeToParse.Day(),
		timeToParse.Hour(), 0, 0, 0, timeToParse.Location())
	fmt.Println("Time:" + fmt.Sprint(roundedTime))
	w.databaseClient.InsertToInflux(message.Topic, columnName, value, roundedTime)
	return
}

func ConvertMessageToInfluxField(message *sarama.ConsumerMessage) (string, int) {
	messageVal := make(map[string]string)
	_ = json.Unmarshal(message.Value, &messageVal)

	delete(messageVal, "@timestamp")
	delete(messageVal, "_ctx")

	var listOfValues []string
	for _, v := range messageVal {
		listOfValues = append(listOfValues, v)
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
