package analytic

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	ErrorFlag   = "ERROR"
	InfoFlag    = "INFO"
	WarningFlag = "WARNING"
	DebugFlag   = "DEBUG"
)

func configLogLevelMapping() map[string]string {
	err := godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/go-squads/floodgate-worker/.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var logMap = make(map[string]string)
	mapToFlag(logMap, "ERROR_LEVELS", ErrorFlag)
	mapToFlag(logMap, "INFO_LEVELS", InfoFlag)
	mapToFlag(logMap, "WARNING_LEVELS", WarningFlag)
	mapToFlag(logMap, "DEBUG_LEVELS", DebugFlag)
	return logMap
}

func mapToFlag(mapObject map[string]string, envString string, flagResult string) {
	for _, logLevel := range strings.Split(os.Getenv(envString), ",") {
		mapObject[logLevel] = flagResult
	}
}
