package buffer

import (
	"os"

	"github.com/go-squads/floodgate-worker/mongo"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	cron "gopkg.in/robfig/cron.v2"
)

type IncomingLog struct {
	Level     string
	Method    string
	Path      string
	Code      string
	Timestamp string
}

type StoreLog struct {
	Level     string `json:"lvl"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Code      string `json:"code"`
	Count     int    `json:"count"`
	Timestamp string `json"timestamp"`
}

type Buffer interface {
	Add(topic string, log IncomingLog)
	StartCron()
	Flush()
	Close()
}

type buffer struct {
	buff map[string]map[IncomingLog]int
	db   mongo.Connector
	cron *cron.Cron
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
		Level:     log.Level,
		Method:    log.Method,
		Path:      log.Path,
		Code:      log.Code,
		Count:     count,
		Timestamp: log.Timestamp,
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

func (s *buffer) Flush() {
	log.Debug("Flushing data to database")
	for k, v := range s.buff {
		log.Println("Flushing data", k, v)
		col := s.db.GetCollection(k)
		for kk, vv := range v {
			sl := createStoreLog(kk, vv)
			col.Insert(sl)
		}
	}
	s.buff = make(map[string]map[IncomingLog]int)
}

func (s *buffer) StartCron() {
	s.cron = cron.New()
	s.cron.AddFunc(os.Getenv("CRON_INTERVAL"), s.Flush)
	s.cron.Start()
}

func (s *buffer) Close() {
	log.Info("Stopping")
	s.cron.Stop()
}
