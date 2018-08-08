package analytic

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	ErrorFlag                   = "ERROR"
	InfoFlag                    = "INFO"
	WarningFlag                 = "WARNING"
	DebugFlag                   = "DEBUG"
	LevelFlag                   = "LOG_LEVEL"
	UnknownFlag                 = "UNKNOWN"
	defaultErrorThreshold       = 10
	defaultWarningThreshold     = 30
	defaultMinimumDataThreshold = 100
)

var (
	errorThreshold       = defaultErrorThreshold
	warningThreshold     = defaultWarningThreshold
	minimumDataThreshold = defaultMinimumDataThreshold
)

func configLogLevelMapping() map[string]string {
	LoadEnviromentConfig()

	var logMap = make(map[string]string)
	mapToFlag(logMap, "ERROR_LEVELS", ErrorFlag)
	mapToFlag(logMap, "INFO_LEVELS", InfoFlag)
	mapToFlag(logMap, "WARNING_LEVELS", WarningFlag)
	mapToFlag(logMap, "DEBUG_LEVELS", DebugFlag)
	mapToFlag(logMap, "LOG_LEVEL_KEY_NAME", LevelFlag)
	return logMap
}

func mapToFlag(mapObject map[string]string, envString string, flagResult string) {
	for _, logLevel := range strings.Split(os.Getenv(envString), ",") {
		mapObject[logLevel] = flagResult
	}
}

func LoadEnviromentConfig() {
	err := godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/go-squads/floodgate-worker/.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	errorThreshold = loadEnviromentThresholds("ERROR_THRESHOLD", defaultErrorThreshold)
	warningThreshold = loadEnviromentThresholds("WARNING_THRESHOLD", defaultWarningThreshold)
	minimumDataThreshold = loadEnviromentThresholds("MINIMUM_DATA", defaultMinimumDataThreshold)
}

func loadEnviromentThresholds(thresholdLabel string, defaultValue int) int {
	resultThreshold, err := strconv.Atoi(os.Getenv(thresholdLabel))
	if err != nil {
		log.Error("Error in converting the .env threshold")
		resultThreshold = defaultValue
	}
	return resultThreshold
}
