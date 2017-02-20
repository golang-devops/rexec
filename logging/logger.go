package logging

import (
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/multi"
	"github.com/apex/log/handlers/text"
	"github.com/natefinch/lumberjack"
)

var logger *log.Entry

func Logger() *log.Entry {
	return logger
}

func init() {
	logJsonFilePath := strings.TrimSpace(os.Getenv("LOG_JSON_FILE"))
	if logJsonFilePath == "" {
		logJsonFilePath = "./log.json"
	}

	rollingLogger := &lumberjack.Logger{
		Filename:   logJsonFilePath,
		MaxSize:    20, // megabytes
		MaxBackups: 20,
		MaxAge:     28, //days
	}

	hostName, err := os.Hostname()
	if err != nil {
		hostName = "UNKNOWN"
	}

	logLevel := log.DebugLevel
	tmpLogger := &log.Logger{
		Handler: multi.New(
			text.New(os.Stdout),
			json.New(rollingLogger),
		),
		Level: logLevel,
	}
	logger = tmpLogger.WithField("hostname", hostName)
}
