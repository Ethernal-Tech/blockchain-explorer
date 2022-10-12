package main

import (
	"ethernal/explorer/db"
	"log"
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

	db := db.InitDb()

	log.Println(db != nil)

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
