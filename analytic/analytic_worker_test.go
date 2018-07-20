package analytic

import (
	"testing"
	"time"

	"github.com/BaritoLog/go-boilerplate/timekit"
	"github.com/Shopify/sarama"
	mock_analytic "github.com/go-squads/floodgate-worker/mock"
	"github.com/golang/mock/gomock"
)

func TestAnalyticWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	want := &sarama.ConsumerMessage{
		Key:            nil,
		Value:          nil,
		Topic:          "analytic-test",
		Partition:      0,
		Offset:         0,
		Timestamp:      time.Now(),
		BlockTimestamp: time.Now(),
		Headers:        nil,
	}

	mockCluster := mock_analytic.NewMockClusterAnalyser(ctrl)

	mockCluster.EXPECT().Messages().AnyTimes().Return(sendSampleMessage(want))
	mockCluster.EXPECT().MarkOffset(gomock.Any(), gomock.Any())
	mockCluster.EXPECT().Close()

	analyserTest := NewAnalyticWorker(mockCluster)

	analyserTest.Start()
	defer analyserTest.Stop()

	// Write TESTS!

	timekit.Sleep("1ms")
}

func sendSampleMessage(message ...*sarama.ConsumerMessage) <-chan *sarama.ConsumerMessage {
	newMsg := make(chan *sarama.ConsumerMessage)
	go func() {
		for _, v := range message {
			newMsg <- v
		}
	}()

	timekit.Sleep("1s")
	return newMsg
}
