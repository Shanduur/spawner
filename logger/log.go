package logger

import (
	"bytes"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is default instance of logger used in all other packages
// instead of global scope logrus.Logger.
var Logger *logrus.Logger

var logFile *os.File

var Buffer = bytes.NewBuffer([]byte{})

const (
	From = "from"
)

func init() {
	Logger = logrus.New()

	configLogger()
}

func configLogger() {
	Log().SetFormatter(&logrus.JSONFormatter{})

	logFile, err := os.OpenFile("spawner.err", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Log().SetOutput(io.MultiWriter(Buffer))
	} else {
		Log().SetOutput(io.MultiWriter(logFile, Buffer))
	}

	switch os.Getenv("LOG_LEVEL") {
	case logrus.PanicLevel.String():
		Log().SetLevel(logrus.PanicLevel)

	case logrus.FatalLevel.String():
		Log().SetLevel(logrus.FatalLevel)

	case logrus.ErrorLevel.String():
		Log().SetLevel(logrus.ErrorLevel)

	case logrus.WarnLevel.String():
		Log().SetLevel(logrus.WarnLevel)

	case logrus.InfoLevel.String():
		Log().SetLevel(logrus.InfoLevel)

	case logrus.DebugLevel.String():
		Log().SetLevel(logrus.DebugLevel)

	case logrus.TraceLevel.String():
		Log().SetLevel(logrus.TraceLevel)
	}

	Log().Infof("level set to: %s", Log().GetLevel().String())
}

// Log is used to return the default Logger.
func Log() *logrus.Logger {
	return Logger
}
