package analytic

import (
	"testing"

	"github.com/Shopify/sarama"
	mock "github.com/go-squads/floodgate-worker/mock"
	"github.com/golang/mock/gomock"
)

func TestIfNewTopicIsDetected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockBrokerServices(ctrl)
	mockClient.EXPECT().SetUpConfig().AnyTimes().Return(gomock.Any())
	mockClient.EXPECT().SetUpClient(gomock.Any(), gomock.Any()).AnyTimes().Return(gomock.Any())
	testService := NewAnalyticServices([]string{"localhost:9092"})
	value := testService.checkIfTopicAlreadySubscribed("thisisdefinitelynotatopic")
	if !value {
		t.Error("Failed to detect new topic")
	}
}

func TestIfOldTopicIsDetected(t *testing.T) {
	// testService := NewAnalyticServices([]string{"localhost:9092"})

}

func TestIfNewWorkerIsProperlyMapped(t *testing.T) {
	testService := NewAnalyticServices([]string{"localhost:9092"})
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
