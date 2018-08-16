package buffer

import (
	"time"
	"github.com/go-squads/floodgate-worker/mongo"
	"os"
	"log"
	"gopkg.in/robfig/cron.v2"
	"github.com/sirupsen/logrus"
)

type IncomingLog struct {
	Level  string
	Method string
	Path   string
	Code   string
}

type StoreLog struct {
	Level     string    `json:"lvl"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Code      string    `json:"code"`
	Count     int       `json:"count"`
	Timestamp string	`json"timestamp"`
}

type Buffer interface {
	Add(topic string, log IncomingLog)
	StartCron()
	Flush()
}

type buffer struct {
	buff map[string]map[IncomingLog]int
	db mongo.Connector
}

var bufferObj Buffer

func GetBuffer() Buffer {
	if bufferObj == nil {
		log.Fatal("Please instantiate the buffer first")
		os.Exit(1)
	}
	return bufferObj
}

func createStoreLog(log IncomingLog, count int) *StoreLog {
	return &StoreLog{
		Level: log.Level,
		Method: log.Method,
		Path: log.Path,
		Code: log.Code,
		Count: count,
		Timestamp: time.Now().Format(os.Getenv("TIME_LAYOUT")),
	}
}

func New(connector mongo.Connector) Buffer {
	bufferObj = &buffer{buff: make(map[string]map[IncomingLog]int), db: connector}
	bufferObj.StartCron()
	return bufferObj
}

func (s *buffer) Add(topic string, log IncomingLog) {
	logrus.Debug("Incoming data from ", topic)
	if s.buff[topic] == nil {
		s.buff[topic] = make(map[IncomingLog]int)
	}
	s.buff[topic][log]++
	logrus.Debug("Adding data to buffer", s.buff)
}

func (s *buffer)Flush() {
	log.Println("Flushing data to database")
	for k,v := range s.buff {
		log.Println("Flushing data",k,v)
		col := s.db.GetCollection(k)
		for kk,vv := range v {
			sl := createStoreLog(kk,vv)
			col.Insert(sl)
		}
	}
	s.buff = make(map[string]map[IncomingLog]int)
}

func (s *buffer)StartCron() {
	c := cron.New()
	c.AddFunc("0 * * * * *", s.Flush)
	c.Start()
}
