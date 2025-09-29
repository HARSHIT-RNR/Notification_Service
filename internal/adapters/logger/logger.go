package logger

import (
	"github.com/sirupsen/logrus"
)

// NewLogger initializes and returns a new logrus logger instance.
func NewLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
	return log
}
