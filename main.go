package main

import (
	"ethernal/explorer/common"
	"ethernal/explorer/config"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"ethernal/explorer/listener"
	"ethernal/explorer/loger"
	"ethernal/explorer/syncer"

	"github.com/sirupsen/logrus"
)

func main() {
	loger.Setup()
	loger.Rotate()

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
