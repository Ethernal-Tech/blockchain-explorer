package db

import (
	"context"
	"database/sql"
	"ethernal/explorer/config"
	"fmt"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func InitDb(config config.Config) *bun.DB {

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connString)))
	db := bun.NewDB(sqldb, pgdialect.New())

	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	ctx := context.Background()
	if _, err := db.NewCreateTable().Model((*Block)(nil)).Exec(ctx); err != nil {
		log.Println(err)
	}

	if _, err := db.NewCreateTable().Model((*Transaction)(nil)).Exec(ctx); err != nil {
		log.Println(err)
	}

	return db
}

var _ bun.BeforeCreateTableHook = (*Transaction)(nil)

func (*Transaction) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("block_hash") REFERENCES "blocks" ("hash")`)
	return nil
}

var _ bun.AfterCreateTableHook = (*Transaction)(nil)

func (*Transaction) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("from_idx").
		Column("from").
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("to_idx").
		Column("to").
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("block_hash_idx").
		Column("block_hash").
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("block_number_idx").
		Column("block_number").
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("contract_address_idx").
		Column("contract_address").
		Exec(ctx)
	if err != nil {
		return err
	}

	return err
}

var _ bun.AfterCreateTableHook = (*Block)(nil)

func (*Block) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

	_, err = query.DB().NewCreateIndex().
		Model((*Block)(nil)).
		Index("number_idx").
		Column("number").
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Block)(nil)).
		Index("miner_idx").
		Column("miner").
		Exec(ctx)
	if err != nil {
		return err
	}

	return err
}
