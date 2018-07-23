package dbhandler

import (
	"fmt"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

type InfluxDB interface {
	InitDB() error
	InsertToInflux(MyDB string, measurement string, columnName string, value int, roundedTime time.Time)
	GetFieldValueIfExist(MyDB string, columnName string, measurement string, roundedTime time.Time) int
}

type influxDb struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"databaseName"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func NewInfluxService(port int, host, dbname, username, password string) InfluxDB {
	return &influxDb{
		Host:         host,
		Port:         port,
		DatabaseName: dbname,
		Username:     username,
		Password:     password,
	}
}

func (influxDb *influxDb) InitDB() error {
	_, err := url.Parse(fmt.Sprintf("http://%s:%d", influxDb.Host, influxDb.Port))
	return err
}

func (influxDb *influxDb) InsertToInflux(MyDB string, measurement string, columnName string, value int, roundedTime time.Time) {
	return
}

func (influxDb *influxDb) GetFieldValueIfExist(MyDB string, columnName string, measurement string, roundedTime time.Time) int {
	return 0
}

func (influxDb *influxDb) connectToInflux() (client.Client, error) {
	fmt.Println("Trying to connect to " + influxDb.DatabaseName + " database")
	addr, err := url.Parse(fmt.Sprintf("http://%s:%d", influxDb.Host, influxDb.Port))
	if err != nil {
		println("InfluxDB : Invalid Url,Please check domain name given\nError Details: ", err.Error())
		return nil, err
	}

	influxDBconnection, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr.String(),
		Username: influxDb.Username,
		Password: influxDb.Password,
	})
	return influxDBconnection, err
}
