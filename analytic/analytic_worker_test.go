package analytic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MyMockedObject struct {
	mock.Mock
}

func TestIfErrorWhenCreatingNewWorker(t *testing.T) {
	assert := assert.New(t)

	brokers := []string{"localhost:9092"}
	topics := []string{"test_logs"}
	testWorker, err := NewAnalyticWorker(brokers, "groupAnalytic", topics)
	if err != nil {
		t.Errorf("Something went wrong when creating the worker")
	}
	assert.NotNil(testWorker.consumer)
	testWorker.Start()

}
