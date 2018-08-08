package dbhandler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

type InfluxDB interface {
	InitDB() error
	InsertToInflux(measurement string, columnName string, value int, roundedTime time.Time) error
	GetFieldValueIfExist(columnName string, measurement string, roundedTime time.Time) int
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
	log.Info("Trying to connect to " + influxDb.DatabaseName + " database")
	addr, err := url.Parse(fmt.Sprintf("http://%s:%d", influxDb.Host, influxDb.Port))
	if err != nil {
		log.Error("InfluxDB : Invalid Url,Please check domain name given\nError Details: ", err.Error())
		return err
	}

	influxDb.DatabaseConnection, err = client.NewHTTPClient(client.HTTPConfig{

		Addr:     addr.String(),
		Username: influxDb.Username,
		Password: influxDb.Password,
	})
	if err != nil {
		return err
	}
	_, _, err = influxDb.DatabaseConnection.Ping(10 * time.Second)

	if err != nil {
		log.Error("InfluxDB : Failed to connect to Database . Please check the details entered\nError Details: ", err.Error())
		return err
	}

	createDbErr := influxDb.createDatabase(influxDb.DatabaseName)
	if createDbErr != nil {
		if createDbErr.Error() != "database already exists" {
			log.Warn("InfluxDB : Failed to create Database")
			return createDbErr
		}

	}

	log.Info("Successfully connected to " + influxDb.DatabaseName + " database!")
	return nil
}

func (influxDb *influxDb) InsertToInflux(measurement string, columnName string, value int, roundedTime time.Time) error {
	if value != 0 {
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxDb.DatabaseName,
			Precision: "s",
		})
		if err != nil {
			return err
		}

		value = value + influxDb.GetFieldValueIfExist(columnName, measurement, roundedTime)

		pt, err := client.NewPoint(measurement, nil, map[string]interface{}{columnName: value}, roundedTime)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		fmt.Println(columnName + "-----" + fmt.Sprint(value))
		if err := influxDb.DatabaseConnection.Write(bp); err != nil {
			return err
		}
	}
	return nil
}

func (influxDb *influxDb) GetFieldValueIfExist(columnName string, measurement string, roundedTime time.Time) int {
	toMili := roundedTime.UnixNano() / int64(time.Nanosecond)
	q := client.Query{
		Command:  fmt.Sprintf("select \"%s\" from \"%s\" where time =%d", columnName, measurement, toMili),
		Database: influxDb.DatabaseName,
	}

	resp, err := influxDb.DatabaseConnection.Query(q)
	if err != nil {
		return 0
	}
	if resp.Error() != nil {
		return 0
	} else {
		if len(resp.Results[0].Series) != 0 {
			res, err := resp.Results[0].Series[0].Values[0][1].(json.Number).Int64()
			if err != nil {
				return 0
			}
			return int(res)
		}
	}
	return 0
}

func (influxDb *influxDb) createDatabase(databaseName string) error {
	command := fmt.Sprintf("create database %s", databaseName)
	q := client.Query{
		Command:  command,
		Database: "",
	}
	return influxDb.basicQuery(q)
}

func (influxDb *influxDb) DeleteDatabase() error {
	command := fmt.Sprintf("drop database %s", influxDb.DatabaseName)
	q := client.Query{
		Command:  command,
		Database: "",
	}
	return influxDb.basicQuery(q)
}

func (influxDb *influxDb) GetDatabaseName() string {
	return influxDb.DatabaseName
}

func (influxDb *influxDb) basicQuery(query client.Query) error {
	if response, err := influxDb.DatabaseConnection.Query(query); err == nil {
		if response.Error() != nil {
			fmt.Println(response.Error())
			return response.Error()
		}
	}
	return nil
}
