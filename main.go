package main

import (
	"bytes"
	"ethernal/explorer/common"
	"ethernal/explorer/config"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"ethernal/explorer/listener"
	"ethernal/explorer/syncer"
	"fmt"

	"os"

	logrus "github.com/sirupsen/logrus"
)

type MyFormatter struct{}

var levelList = []string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
}

func (MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//strList := strings.Split(entry.Caller.File, "/")
	//fileName := strList[len(strList)-1]

	b.WriteString(fmt.Sprintf(" %s - %s (line:%d)\n[%s] %s\n\n",
		entry.Time.Format("2006-01-02 15:04:05"), entry.Caller.File,
		entry.Caller.Line, levelList[int(entry.Level)], entry.Message))
	return b.Bytes(), nil
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(MyFormatter{})

	// log.StandardLogger().Formatter = &easy.Formatter{
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// 	LogFormat:       "[%lvl%]: %time% - %msg%\n",
	// }
}

func main() {
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logrus.Panic("Error opening or creating file, err: ", err)
	}
	defer f.Close()

	logrus.SetOutput(f)

	os.Stderr = f

	config, err := config.LoadConfig()
	if err != nil {
		logrus.Panic("Failed to load config, err: ", err.Error())
	}

	db := db.InitDb(config)

	switch config.Mode {
	case common.Manual:
		// HTTP connection to blockchain
		connection := eth.BlockchainNodeConnection{
			HTTP: eth.GetClient(config.HTTPUrl),
		}
		syncer.SyncMissingBlocks(connection.HTTP, db, config)
	case common.Automatic:
		// both HTTP and WebSocket connection to blockchain
		connection := eth.BlockchainNodeConnection{
			HTTP:      eth.GetClient(config.HTTPUrl),
			WebSocket: eth.GetClient(config.WebSocketUrl),
		}
		listener.ListenForNewBlocks(&connection, db, config)
	default:
		logrus.Info("Mode %s is not provided", config.Mode)
	}
}
