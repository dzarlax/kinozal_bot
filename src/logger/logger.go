package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()

	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	log.SetLevel(logrus.InfoLevel)

	file, err := os.OpenFile("logs/combined.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Warn("Failed to log to file, using default stderr")
	}
}

func Debug(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Debug(message)
	} else {
		log.Debug(message)
	}
}

func Info(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Info(message)
	} else {
		log.Info(message)
	}
}

func Warn(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Warn(message)
	} else {
		log.Warn(message)
	}
}

func Error(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Error(message)
	} else {
		log.Error(message)
	}
}