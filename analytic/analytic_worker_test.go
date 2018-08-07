package analytic

import (
	"fmt"
	"testing"
	"time"

	"github.com/BaritoLog/go-boilerplate/timekit"
	"github.com/Shopify/sarama"
	"github.com/golang/mock/gomock"

	models "github.com/go-squads/floodgate-worker/influxdb-handler"
	mock "github.com/go-squads/floodgate-worker/mock"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	want := &sarama.ConsumerMessage{
		Key:            []byte{},
		Value:          []byte{10, 20},
		Topic:          "analytic-test",
		Partition:      0,
		Offset:         0,
		Timestamp:      time.Now(),
		BlockTimestamp: time.Now(),
		Headers:        nil,
	}

	consumer := mock.NewMockClusterAnalyser(ctrl)

	consumer.EXPECT().Messages().AnyTimes().Return(sampleMsg(want))
	consumer.EXPECT().MarkOffset(gomock.Any(), gomock.Any())
	consumer.EXPECT().Close()

	influxDB := models.NewInfluxService(8086, "localhost", "analyticsdbtest", "gopayadmin", "gopayadmin")
	influxDB.InitDB()

	worker := NewAnalyticWorker(consumer, influxDB, configLogLevelMapping())

	var got *sarama.ConsumerMessage

	worker.Start(func(message *sarama.ConsumerMessage) {
		got = message
	})
	defer worker.Stop()

	timekit.Sleep("1ms")
	fmt.Println("------" + fmt.Sprint(want))
	fmt.Println("======" + fmt.Sprint(got))
	assert.Equal(t, want, got, "they should be equal")
}

func TestAnalyticWorkerZeroParam_ErrorBadkafka(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	want := &sarama.ConsumerMessage{
		Key:            []byte{},
		Value:          []byte{10, 20},
		Topic:          "analytic-test",
		Partition:      0,
		Offset:         0,
		Timestamp:      time.Now(),
		BlockTimestamp: time.Now(),
		Headers:        nil,
	}

	consumer := mock.NewMockClusterAnalyser(ctrl)

	consumer.EXPECT().Messages().AnyTimes().Return(sampleMsg(want))
	consumer.EXPECT().Close()

	influxDB := models.NewInfluxService(8086, "localhost", "analyticsdbtest", "gopayadmin", "gopayadmin")
	influxDB.InitDB()

	worker := NewAnalyticWorker(consumer, influxDB, configLogLevelMapping())

	var got *sarama.ConsumerMessage
	worker.OnSuccess(func(message *sarama.ConsumerMessage) { got = message })

	worker.Start()
	defer worker.Stop()

	timekit.Sleep("1ms")
	fmt.Println("------" + fmt.Sprint(want))
	fmt.Println("======" + fmt.Sprint(got))
	assert.NotEqual(t, want, got, "they should not be equal")
}

func TestMessageConvertedToInfluxFieldRight(t *testing.T) {
	testMessage := []byte(`{"Method":"POST","Code":"200","@timestamp":"test"}`)

	test := &sarama.ConsumerMessage{
		Key:            []byte{},
		Value:          testMessage,
		Topic:          "analytic-test",
		Partition:      0,
		Offset:         0,
		Timestamp:      time.Now(),
		BlockTimestamp: time.Now(),
		Headers:        nil,
	}

	colName, value := ConvertMessageToInfluxField(test, "")
	assert.Equal(t, "200_POST", colName, "Column name should be 200_POST")
	assert.Equal(t, 1, value, "Value should be one")
}

func TestBadMessageConvertToInfluxField(t *testing.T) {
	test := &sarama.ConsumerMessage{
		Key:            []byte{},
		Value:          []byte{10, 20},
		Topic:          "analytic-test",
		Partition:      0,
		Offset:         0,
		Timestamp:      time.Now(),
		BlockTimestamp: time.Now(),
		Headers:        nil,
	}

	colName, value := ConvertMessageToInfluxField(test, "")
	assert.Equal(t, "", colName, "Column Name returned should be empty")
	assert.Equal(t, 0, value, "Value returned should be zero")
}

func sampleMsg(message ...*sarama.ConsumerMessage) <-chan *sarama.ConsumerMessage {
	newMsg := make(chan *sarama.ConsumerMessage)
	go func() {
		for _, v := range message {
			newMsg <- v
		}
	}()

	timekit.Sleep("1s")
	return newMsg
}

func TestExtractCorrectLogLabel(t *testing.T) {
	worker := NewAnalyticWorker(nil, nil, configLogLevelMapping())
	var testMap = make(map[string]interface{})
	testMap["LOG_LEVEL"] = "ERROR"
	result, exist := worker.getLogLabel(testMap)
	if !exist || result != "LOG_LEVEL" {
		t.Error("Failed to extract correct log label")
	}
}

func TestIfInvalidLogLabel(t *testing.T) {
	worker := NewAnalyticWorker(nil, nil, configLogLevelMapping())
	var testMap = make(map[string]interface{})
	testMap["ERROR"] = "ERROR"
	result, exist := worker.getLogLabel(testMap)
	if exist || result == "ERROR" {
		t.Error("Failed to detect it's an invalid log level")
	}
}
