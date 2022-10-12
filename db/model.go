package db

import (
	"github.com/lib/pq"
)

// Events - Events emitted from smart contracts to be held in this table
type Event struct {
	BlockHash       string         `bun:",pk"`
	Index           uint           `bun:",pk,type:integer"`
	Origin          string         `bun:"type:char(42),notnull"`
	Topics          pq.StringArray `bun:"type:text[],notnull"`
	Data            []byte
	TransactionHash string `bun:"type:char(66),notnull"`
}

// Transactions - Blockchain transaction holder table model
type Transaction struct {
	Hash      string `bun:",pk,type:char(66)"`
	From      string `bun:"type:char(42),notnull"`
	To        string `bun:"type:char(42)"`
	Contract  string `bun:"type:char(42)"`
	Value     string `bun:"type:varchar"`
	Data      []byte
	Gas       uint64 `bun:"type:bigint,notnull"`
	GasPrice  string `bun:"type:varchar,notnull"`
	Cost      string `bun:"type:varchar,notnull"`
	Nonce     uint64 `bun:"type:bigint,notnull"`
	State     uint64 `bun:"type:smallint,notnull"`
	BlockHash string `bun:"type:char(66),notnull"`
	Event     Event
}

// Blocks - Mined block info holder table model
type Block struct {
	Hash                string  `bun:",pk,type:char(66)"`
	Number              uint64  `bun:"type:bigint,notnull,unique"`
	Time                uint64  `bun:"type:bigint,notnull"`
	ParentHash          string  `bun:"type:char(66),notnull"`
	Difficulty          string  `bun:"type:varchar,notnull"`
	GasUsed             uint64  `bun:"type:bigint,notnull"`
	GasLimit            uint64  `bun:"type:bigint,notnull"`
	Nonce               string  `bun:"type:varchar,notnull"`
	Miner               string  `bun:"type:char(42),notnull"`
	Size                float64 `bun:"type:float(8),notnull"`
	StateRootHash       string  `bun:"type:char(66),notnull"`
	UncleHash           string  `bun:"type:char(66),notnull"`
	TransactionRootHash string  `bun:"type:char(66),notnull"`
	ReceiptRootHash     string  `bun:"type:char(66),notnull"`
	ExtraData           []byte  `bun:"type:bytea"`
	Transaction         Transaction
	Event               Event
}
