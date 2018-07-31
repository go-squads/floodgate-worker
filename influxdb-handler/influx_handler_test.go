package dbhandler

import (
	"fmt"
	"testing"

	mock_dbhandler "github.com/go-squads/floodgate-worker/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInfluxDB_invalidDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := "Get http://local:host:8086/ping?wait_for_leader=10s: invalid URL port \"host:8086\""
	influxDb := mock_dbhandler.NewMockInfluxDB(ctrl)
	influxDb.EXPECT().InitDB().Return(fmt.Errorf(expectedErr))

	influx := NewInfluxService(8086, "local:host", "analytics-test", "gopayadmin", "gopayadmin")

	assert.Equal(t, influxDb.InitDB().Error(), influx.InitDB().Error())
}

func TestInsertToInflux(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    influxDb := mock_dbhandler.NewMockInfluxDB(ctrl)
    influxDb.EXPECT().InsertToInflux(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

    influx := NewInfluxService(8086, "localhost", "analytics-test", "gopayadmin", "gopayadmin")

    assert.Equal(t, influx.InsertToInflux("test", "gopay-test", "500_GET", 1, time.Now()), nil)
}
