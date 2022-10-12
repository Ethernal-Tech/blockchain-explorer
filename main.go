package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Book struct {
	ID         int64 `bun:",pk,autoincrement"`
	Name       string
	CategoryID int64
}

var _ bun.AfterCreateTableHook = (*Book)(nil)

func (*Book) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	_, err := query.DB().NewCreateIndex().
		Model((*Book)(nil)).
		Index("category_id_idx").
		Column("category_id").
		Exec(ctx)
	return err
}

func main() {
	log.Println("Hello world")
	configFile, err := filepath.Abs(".env")

	if err != nil {
		log.Fatalf("[!] Failed to find `.env` : %s\n", err.Error())
	}

	err = Read(configFile)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		Get("DB_USER"), Get("DB_PASSWORD"), Get("DB_HOST"),
		Get("DB_PORT"), Get("DB_NAME"))

	log.Println(dsn)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	ctx := context.Background()
	if _, err := db.NewCreateTable().Model((*Book)(nil)).Exec(ctx); err != nil {
		log.Println(err)
	}

	bookNum := 50000

	var books []Book

	for i := 0; i < bookNum; i++ {
		b := &Book{
			Name:       "book_test",
			CategoryID: 1,
		}
		books = append(books, *b)
	}

	startingAt := time.Now().UTC()

	for _, book := range books {
		// if _, err := db.NewInsert().Model(&book).Exec(ctx); err != nil {
		// 	log.Println(err)
		// }
		//db.NewInsert().Model(&[]Book{book}).Exec(ctx)
		db.NewInsert().Model(&book).Exec(ctx)
	}

	//db.NewInsert().Model(&books).Exec(ctx)
	log.Println("Took: ", time.Now().UTC().Sub(startingAt))
}
