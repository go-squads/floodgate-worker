package analytic

import (
	"testing"

	"github.com/Shopify/sarama"
)

func TestIfNewTopicIsDetected(t *testing.T) {
	testService := NewAnalyticServices([]string{"localhost:9092"})
	value := testService.checkIfTopicAlreadySubscribed("thisisdefinitelynotatopic")
	if value {
		t.Error("Failed to detect new topic")
	}
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
