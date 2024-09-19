package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func SetupLogger() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	logLevel := logrus.InfoLevel //read from config
	log.SetLevel(logLevel)
}

func GetLogger() *logrus.Logger {
	return log
}

type Fields logrus.Fields

func WithFields(fields Fields) *logrus.Entry {
	return log.WithFields(logrus.Fields(fields))
}
