package main

import (
	"ethernal/explorer/config"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"ethernal/explorer/syncer"
	"log"
	"os"
	"time"
)

type Block struct {
	Number string
}

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

	rpcClient := eth.GetClient(config.RPCUrl)

	startingAt := time.Now().UTC()
	syncer.SyncMissingBlocks(rpcClient, db, config)
	log.Println("Took: ", time.Now().UTC().Sub(startingAt))
	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
}
