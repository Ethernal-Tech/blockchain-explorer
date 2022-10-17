package main

import (
	"ethernal/explorer/config"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"ethernal/explorer/syncer"
	"log"
	"time"
)

// type Book struct {
// 	ID         int64 `bun:",pk,autoincrement"`
// 	Name       string
// 	CategoryID int64
// 	Author     string
// }

// var _ bun.AfterCreateTableHook = (*Book)(nil)

// func (*Book) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
// 	_, err := query.DB().NewCreateIndex().
// 		Model((*Book)(nil)).
// 		Index("category_id_idx").
// 		Column("category_id").
// 		Exec(ctx)
// 	return err
// }

type Block struct {
	Number string
}

func main() {

	// f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// log.SetOutput(f)
	// log.Println("This is a test log entry")

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
