package main

import (
	"ethernal/explorer/config"
	"ethernal/explorer/eth"
	"ethernal/explorer/syncer"
	"log"
	"path/filepath"
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

func main() {
	log.Println("Hello world")

	configFile, err := filepath.Abs(".env")

	if err != nil {
		log.Fatalf("[!] Failed to find `.env` : %s\n", err.Error())
	}

	err = config.Read(configFile)

	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}
	// db := db.InitDb()

	// log.Println(db != nil)

	ethClient := eth.GetClient()
	missingBlocks := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	//missingBlocks := []uint64{1, 2}
	startingAt := time.Now().UTC()
	syncer.SyncMissingBlocks(missingBlocks, ethClient)
	log.Println("Took: ", time.Now().UTC().Sub(startingAt))
	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// bookNum := 50000

	// var books []Book

	// for i := 0; i < bookNum; i++ {
	// 	b := &Book{
	// 		Name:       "book_test",
	// 		CategoryID: 1,
	// 	}
	// 	books = append(books, *b)
	// }

	// startingAt := time.Now().UTC()

	// for _, book := range books {
	// 	// if _, err := db.NewInsert().Model(&book).Exec(ctx); err != nil {
	// 	// 	log.Println(err)
	// 	// }
	// 	//db.NewInsert().Model(&[]Book{book}).Exec(ctx)
	// 	db.NewInsert().Model(&book).Exec(ctx)
	// }

	// //db.NewInsert().Model(&books).Exec(ctx)
	// log.Println("Took: ", time.Now().UTC().Sub(startingAt))
}
