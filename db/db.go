package db

import (
	"context"
	"database/sql"
	"ethernal/explorer/config"
	"fmt"
	"time"

	logrusbun "github.com/oiime/logrusbun"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func InitDb(config config.Config) *bun.DB {

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName, config.DbSSL)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connString), pgdriver.WithTimeout(0*time.Second)))

	err := sqldb.Ping()
	if err != nil {
		logrus.Panic("Cannot connect to DB, err: ", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{
		Logger:     logrus.StandardLogger(),
		QueryLevel: logrus.DebugLevel,
		ErrorLevel: logrus.ErrorLevel,
		SlowLevel:  logrus.WarnLevel,
	}))

	ctx := context.Background()
	if _, err := db.NewCreateTable().Model((*Block)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table Block, err: ", err)
	}

	if _, err := db.NewCreateTable().Model((*Transaction)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table Transaction, err: ", err)
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
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("to_idx").
		Column("to").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("block_hash_idx").
		Column("block_hash").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("block_number_idx").
		Column("block_number").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Transaction)(nil)).
		Index("contract_address_idx").
		Column("contract_address").
		IfNotExists().
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
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Block)(nil)).
		Index("miner_idx").
		Column("miner").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	return err
}
