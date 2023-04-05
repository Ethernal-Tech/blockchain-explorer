package loger

import (
	"bytes"
	"fmt"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

type MyFormatter struct{}

// log levels which logrus supports
var levelList = []string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
}

// creating custom formatter
func (MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if entry.Level == logrus.DebugLevel {
		b.WriteString(fmt.Sprintf(" %s - %s (line:%d)\n[%s] %s\n\n",
			entry.Time.Format("2006-01-02 15:04:05"), entry.Caller.File,
			entry.Caller.Line, levelList[int(entry.Level)], entry.Message))
		return b.Bytes(), nil
	} else {
		b.WriteString(fmt.Sprintf(" %s [%s] %s\n\n",
			entry.Time.Format("2006-01-02 15:04:05"), levelList[int(entry.Level)], entry.Message))
		return b.Bytes(), nil
	}
}

func Setup() {
	logrus.SetReportCaller(true)       // this line is for logging filename and line number
	logrus.SetLevel(logrus.InfoLevel)  // setting log level
	logrus.SetFormatter(MyFormatter{}) // setting custom formatter
}

func Rotate() {
	writer, err := rotatelogs.New(
		"logfile.%Y.%m.%d",
		//rotatelogs.WithLinkName("logfile"),
		rotatelogs.WithRotationTime(7*24*time.Hour),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(4),
	)

	if err != nil {
		logrus.Fatalf("Failed to Initialize Log File %s", err)
	}
	logrus.SetOutput(writer)
}
