package analytic

import "testing"

func TestIfNewTopicIsDetected(t *testing.T) {
	testService := NewAnalyticServices([]string{"localhost:9092"})
	value := testService.checkIfTopicAlreadySubscribed("thisisnotatopic")
	if value {
		t.Error("Failed to detect new topic")
	}
}
