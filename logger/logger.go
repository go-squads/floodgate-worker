package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

func Init() {
	SetTimestamp()
	SetLevel(os.Getenv("LOG_LEVEL"))
}

func SetTimestamp(){
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}

func SetLevel(level string) {
	logrus.Infof("Setting log level with %v", level)
	switch level {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.Infof("Setted log level to %v", logrus.GetLevel())
}
