package db

import (
	"context"
	"database/sql"
	"ethernal/explorer/config"
	"fmt"
	"log"
	"path/filepath"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func InitDb() *bun.DB {
	configFile, err := filepath.Abs(".env")

	if err != nil {
		log.Fatalf("[!] Failed to find `.env` : %s\n", err.Error())
	}

	err = config.Read(configFile)

	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		config.Get("DB_USER"), config.Get("DB_PASSWORD"), config.Get("DB_HOST"),
		config.Get("DB_PORT"), config.Get("DB_NAME"))

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connString)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	ctx := context.Background()
	if _, err := db.NewCreateTable().Model((*Block)(nil)).Exec(ctx); err != nil {
		log.Println(err)
	}

	if _, err := db.NewCreateTable().Model((*Transaction)(nil)).Exec(ctx); err != nil {
		log.Println(err)
	}

	if _, err := db.NewCreateTable().Model((*Event)(nil)).Exec(ctx); err != nil {
		log.Println(err)
	}

	return db
}

var _ bun.BeforeCreateTableHook = (*Event)(nil)

func (*Event) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("transaction_hash") REFERENCES "transactions" ("hash")`)
	query.ForeignKey(`("block_hash") REFERENCES "blocks" ("hash")`)
	return nil
}

var _ bun.BeforeCreateTableHook = (*Transaction)(nil)

func (*Transaction) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("block_hash") REFERENCES "blocks" ("hash")`)
	return nil
}

var _ bun.AfterCreateTableHook = (*Event)(nil)

func (*Event) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	_, err := query.DB().NewCreateIndex().
		Model((*Event)(nil)).
		Index("origin_idx").
		Column("origin").
		NewCreateIndex().
		Model((*Event)(nil)).
		Index("topics_idx").
		Column("topics").
		Model((*Event)(nil)).
		Index("transaction_hash_idx").
		Column("transaction_hash").
		Exec(ctx)
	return err
}

var _ bun.AfterCreateTableHook = (*Transaction)(nil)

func (*Transaction) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	_, err := query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("from_idx").
		Column("from").
		NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("to_idx").
		Column("to").
		Model((*Transaction)(nil)).
		Index("block_hash_idx").
		Column("block_hash").
		Exec(ctx)
	return err
}

var _ bun.AfterCreateTableHook = (*Block)(nil)

func (*Block) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	_, err := query.DB().NewCreateIndex().
		Model((*Block)(nil)).
		Index("number_idx").
		Column("number").
		Exec(ctx)
	return err
}
