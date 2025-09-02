package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// InitLogger initializes the global logger with the specified level
func InitLogger(level string) {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger("info")
	}
	return log
}
