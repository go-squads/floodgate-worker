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

	influxDB := models.NewInfluxService(8086, "localhost", "analyticsdb", "gopayadmin", "gopayadmin")
	influxDB.InitDB()

	worker := NewAnalyticWorker(consumer, influxDB)

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
	consumer.EXPECT().MarkOffset(gomock.Any(), gomock.Any())
	consumer.EXPECT().Close()

	influxDB := models.NewInfluxService(8086, "localhost", "analyticsdb", "gopayadmin", "gopayadmin")
	influxDB.InitDB()

	worker := NewAnalyticWorker(consumer, influxDB)

	var got *sarama.ConsumerMessage

	worker.Start()
	defer worker.Stop()

	timekit.Sleep("1ms")
	fmt.Println("------" + fmt.Sprint(want))
	fmt.Println("======" + fmt.Sprint(got))
	assert.Equal(t, want, got, "they should be equal")
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
