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

func InitDb(config *config.Config) *bun.DB {

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

	if _, err := db.NewCreateTable().Model((*Contract)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table Contract, err: ", err)
	}

	if _, err := db.NewCreateTable().Model((*Log)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table Log, err: ", err)
	}

	if _, err := db.NewCreateTable().Model((*AbiType)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table AbiType, err: ", err)
	}

	if _, err := db.NewCreateTable().Model((*Abi)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table Abi, err: ", err)
	}

	if _, err := db.NewCreateTable().Model((*TokenType)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table TokenType, err: ", err)
	}

	if _, err := db.NewCreateTable().Model((*Nft)(nil)).IfNotExists().Exec(ctx); err != nil {
		logrus.Panic("Error while creating the table Nft, err: ", err)
	}
	return db
}

// ---------------Contract Table---------------------------------
var _ bun.BeforeCreateTableHook = (*Contract)(nil)

func (*Contract) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("transaction_hash") REFERENCES "transactions" (hash)`)
	return nil
}

var _ bun.AfterCreateTableHook = (*Contract)(nil)

func (*Contract) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

	_, err = query.DB().NewCreateIndex().
		Model((*Contract)(nil)).
		Index("contracts_transaction_hash_idx").
		Column("transaction_hash").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// -----------------AbiType Table-----------------------------
var _ bun.AfterCreateTableHook = (*AbiType)(nil)

func (*AbiType) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {

	var count int = 0
	count, err := query.DB().NewSelect().Model(&AbiType{}).Count(ctx)
	if count == 0 {
		if err != nil {
			logrus.Panic("Error while checking count of rows in the AbiType table, err: ", err)
			return err
		}
		abiTypes := []*AbiType{
			{Id: 1, Name: "Constructor"},
			{Id: 2, Name: "Event"},
			{Id: 3, Name: "Function"},
		}
		if _, err := query.DB().NewInsert().Model(&abiTypes).Exec(ctx); err != nil {
			logrus.Panic("Error while inserting data into the AbiType table, err: ", err)
			return err
		}
	}
	return nil
}

// ------------------Abi Table---------------------------------
var _ bun.BeforeCreateTableHook = (*Abi)(nil)

func (*Abi) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("address") REFERENCES "contracts" (address)`)
	query.ForeignKey(`("abi_type_id") REFERENCES "abi_types" (id)`)
	return nil
}

var _ bun.AfterCreateTableHook = (*Abi)(nil)

func (*Abi) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

	_, err = query.DB().NewCreateIndex().
		Model((*Abi)(nil)).
		Index("hash_idx").
		Column("hash").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Abi)(nil)).
		Index("abis_address_idx").
		Column("address").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ----------------Log Table---------------------------------------
var _ bun.BeforeCreateTableHook = (*Log)(nil)

func (*Log) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("block_hash") REFERENCES "blocks" (hash)`)
	query.ForeignKey(`("transaction_hash") REFERENCES "transactions" (hash)`)
	return nil
}

var _ bun.AfterCreateTableHook = (*Log)(nil)

func (*Log) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

	_, err = query.DB().NewCreateIndex().
		Model((*Log)(nil)).
		Index("logs_address_idx").
		Column("address").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = query.DB().NewCreateIndex().
		Model((*Log)(nil)).
		Index("logs_transaction_hash_idx").
		Column("transaction_hash").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// --------------Transaction Table-----------------------------------------
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

// --------------Block Table-------------------------------
var _ bun.AfterCreateTableHook = (*Block)(nil)

func (*Block) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

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

// -----------------TokenType Table-----------------------------
var _ bun.AfterCreateTableHook = (*TokenType)(nil)

func (*TokenType) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {

	var count int = 0
	count, err := query.DB().NewSelect().Model(&TokenType{}).Count(ctx)
	if count == 0 {
		if err != nil {
			logrus.Panic("Error while checking count of rows in the TokenType table, err: ", err)
			return err
		}
		tokenTypes := []*TokenType{
			{Id: 1, Name: "ERC-20"},
			{Id: 2, Name: "ERC-721"},
			{Id: 3, Name: "ERC-1155"},
		}
		if _, err := query.DB().NewInsert().Model(&tokenTypes).Exec(ctx); err != nil {
			logrus.Panic("Error while inserting data into the TokenType table, err: ", err)
			return err
		}
	}
	return nil
}

// ---------------Nft Table---------------------------------
var _ bun.BeforeCreateTableHook = (*Nft)(nil)

func (*Nft) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("block_hash", "index") REFERENCES "logs" ("block_hash", "index")`)
	query.ForeignKey(`("transaction_hash") REFERENCES "transactions" (hash)`)
	query.ForeignKey(`("token_type_id") REFERENCES "token_types" (id)`)
	return nil
}

var _ bun.AfterCreateTableHook = (*Nft)(nil)

func (*Nft) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	var err error

	_, err = query.DB().NewCreateIndex().
		Model((*Nft)(nil)).
		Index("nfts_block_number_idx").
		Column("block_number").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
