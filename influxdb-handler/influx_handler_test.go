package dbhandler

import (
	"testing"
)

func TestIfConnectToInfluxFails(t *testing.T) {
	testDBService := NewInfluxService(0, "", "", "", "")
	_, err := testDBService.connectToInflux()
	if err == nil {
		t.Error("Failed to detect error")
	}
}
