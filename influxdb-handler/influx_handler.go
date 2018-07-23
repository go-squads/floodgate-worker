package dbhandler

import (
	"fmt"
	"log"
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
	Host               string `json:"host"`
	Port               int    `json:"port"`
	DatabaseName       string `json:"databaseName"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	DatabaseConnection client.Client
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
	fmt.Println("Trying to connect to " + influxDb.DatabaseName + " database")
	addr, err := url.Parse(fmt.Sprintf("http://%s:%d", influxDb.Host, influxDb.Port))
	if err != nil {
		println("InfluxDB : Invalid Url,Please check domain name given\nError Details: ", err.Error())
		return err
	}

	influxDb.DatabaseConnection, err = client.NewHTTPClient(client.HTTPConfig{

		Addr:     addr.String(),
		Username: influxDb.Username,
		Password: influxDb.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	_, _, err = influxDb.DatabaseConnection.Ping(10 * time.Second)

	if err != nil {
		println("InfluxDB : Failed to connect to Database . Please check the details entered\nError Details: ", err.Error())
		return err
	}

	createDbErr := influxDb.createDatabase(influxDb.DatabaseName)
	if createDbErr != nil {
		if createDbErr.Error() != "database already exists" {
			println("InfluxDB : Failed to create Database")
			return createDbErr
		}

	}

	fmt.Println("Successfully connected to " + influxDb.DatabaseName + " database!")
	return nil
}

func (influxDb *influxDb) InsertToInflux(MyDB string, measurement string, columnName string, value int, roundedTime time.Time) {
	return
}

func (influxDb *influxDb) GetFieldValueIfExist(MyDB string, columnName string, measurement string, roundedTime time.Time) int {
	return 0
}

func (influxDb *influxDb) createDatabase(databaseName string) error {
	command := fmt.Sprintf("create database %s", databaseName)
	q := client.Query{
		Command:  command,
		Database: "",
	}
	if response, err := influxDb.DatabaseConnection.Query(q); err == nil {
		if response.Error() != nil {
			fmt.Println(response.Error())
			return response.Error()
		}
	}
	return nil
}

func (influxDb *influxDb) GetDatabaseName() string {
	return influxDb.DatabaseName
}
