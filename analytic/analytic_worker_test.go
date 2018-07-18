package analytic

import (
	"testing"
)

func TestIfErrorWhenCreatingNewWorker(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topics := []string{"test_logs"}
	testWorker, err := NewAnalyticWorker(brokers, "groupAnalytic", topics)
	if err != nil {
		t.Errorf("Something went wrong when creating the worker")
	}
	testWorker.loopError()
}
