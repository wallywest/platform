package platform

import (
	"io"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	defaultLogger()
}

func defaultLogger() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	logger.Level = logrus.InfoLevel
	Logger = logger
}

func SetLogOutput(v io.Writer) {
	Logger.Out = v
}

func SetLogLevel(level logrus.Level) {
	Logger.Level = level
}

func TestLogger() *logrus.Logger {
	defaultLogger()
	Logger.Out = ioutil.Discard
	return Logger
}
