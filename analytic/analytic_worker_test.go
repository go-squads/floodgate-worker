package analytic

import (
	"fmt"
	"testing"
)

func TestCreateWorker(t *testing.T) {
	brokers := []string{"192.168.1.1"}
	topics := []string{"test1", "test2"}
	testWorker, err := NewAnalyticWorker(brokers, "groupTest", topics)

	if testWorker.Consumer.groupID != "groupTest" {
		fmt.Println("Nope")
	}
}
