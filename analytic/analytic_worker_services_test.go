package analytic

import (
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
	mock "github.com/go-squads/floodgate-worker/mock"
	"github.com/golang/mock/gomock"
)

func TestIfNewTopicIsDetected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAnalyserServices(ctrl)
	mockClient.EXPECT().SetUpClient(gomock.Any()).AnyTimes().Return(nil, fmt.Errorf(""))
	var v interface{} = NewAnalyticServices([]string{"localhost:9092"})
	testService := v.(*analyticServices)
	value := testService.checkIfTopicAlreadySubscribed("thisisdefinitelynotatopic_logs")
	if value {
		t.Error("Failed to detect new topic")
	}
}

func TestIfTopicWithoutProperSuffixIgnored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAnalyserServices(ctrl)
	mockClient.EXPECT().SetUpClient(gomock.Any()).AnyTimes().Return(nil, fmt.Errorf(""))
	var v interface{} = NewAnalyticServices([]string{"localhost:9092"})
	testService := v.(*analyticServices)
	value := testService.checkIfTopicAlreadySubscribed("thisisdefinitelynotatopic")
	if !value {
		t.Error("Failed to detect it doesn't ignore topics with wrong suffix")
	}
}

func TestIfOldTopicIsDetected(t *testing.T) {
	// testService := NewAnalyticServices([]string{"localhost:9092"})

}

func TestIfNewWorkerIsProperlyMapped(t *testing.T) {
	var v interface{} = NewAnalyticServices([]string{"localhost:9092"})
	testService := v.(*analyticServices)
	testService.spawnNewAnalyser("test", sarama.OffsetNewest)
	_, exist := testService.workerList["test"]
	if !exist {
		t.Error("Worker not properly mapped")
	}
}

// Add tests for Close()
// func TestIfMessageValueIsMapped(t * testing.T) {
// 	testMessage := &sarama.ConsumerMessage{
// 		Key:            nil,
// 		Value:          {"Method:" ,
// 		Topic:          "analytic-test",
// 		Partition:      0,
// 		Offset:         0,
// 		Timestamp:      time.Now(),
// 		BlockTimestamp: time.Now(),
// 		Headers:        nil,
// 	}
// }
