package logger

// How to use
// myGoFile.go
//  import 	"gitlab.solaredge.com/portialinuxdevelopers/eos/services/network-service/pkg/utils/logger"
//	logger.InitLogger(true)  // true for enable debug level
//  logger.Logger.Info("My message")

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type CustomFormatter struct {
	logrus.TextFormatter
}

var levelList = []string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
}

func (mf *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	level := levelList[int(entry.Level)]
	strList := strings.Split(entry.Caller.File, "/")
	fileName := strList[len(strList)-1]
	b.WriteString(
		fmt.Sprintf("[%s][%s:%d][%s]: %s\n",
			entry.Time.Format("2006-01-02 15:04:05"),
			fileName,
			entry.Caller.Line,
			level,
			entry.Message,
		),
	)
	//TODO Example how to add color
	// return []byte(fmt.Sprintf("[%s] - \x1b[%dm%s\x1b[0m - %s\n", entry.Time.Format(f.TimestampFormat), levelColor, strings.ToUpper(entry.Level.String()), entry.Message)), nil

	return b.Bytes(), nil
}

func InitLogger(level string) {
	Logger = logrus.New()
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&CustomFormatter{})
	fmt.Printf("Setting logger to level: %s\n", level)
	switch level {
	case "DEBUG":
		Logger.SetLevel(logrus.DebugLevel)
	case "ERROR":
		Logger.SetLevel(logrus.ErrorLevel)
	case "INFO":
		Logger.SetLevel(logrus.InfoLevel)
	}
}
