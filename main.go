package main

import (
	"ethernal/explorer/common"
	"ethernal/explorer/config"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"ethernal/explorer/listener"
	"ethernal/explorer/syncer"
	"log"
	"os"
)

func main() {

	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[!] Failed to load config : %s\n", err.Error())
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
		log.Printf("Mode %s is not provided", config.Mode)
	}
}
