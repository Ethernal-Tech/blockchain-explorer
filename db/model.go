package db

// Blocks - Mined block info holder table model
type Block struct {
	Hash            string `bun:",pk,type:char(66)"`
	Number          uint64 `bun:"type:bigint,notnull,unique"`
	ParentHash      string `bun:"type:char(66),notnull"`
	Nonce           string `bun:"type:varchar,notnull"`
	Miner           string `bun:"type:char(42),notnull"`
	Difficulty      string `bun:"type:varchar,notnull"`
	TotalDifficulty string `bun:"type:varchar,notnull"`
	ExtraData       []byte `bun:"type:bytea"`
	Size            uint64 `bun:"type:bigint,notnull"`
	GasLimit        uint64 `bun:"type:bigint,notnull"`
	GasUsed         uint64 `bun:"type:bigint,notnull"`
	Timestamp       uint64 `bun:"type:bigint,notnull"`
	// Block reward - zbir fee-jeva svih transakcija iz blocka
}

// Transactions - Blockchain transaction holder table model
type Transaction struct {
	Hash             string `bun:",pk,type:char(66)"`
	BlockHash        string `bun:"type:char(66),notnull"`
	BlockNumber      uint64 `bun:"type:bigint,notnull"`
	From             string `bun:"type:char(42),notnull"`
	To               string `bun:"type:char(42)"`
	Gas              uint64 `bun:"type:bigint,notnull"`
	GasUsed          uint64 `bun:"type:bigint,notnull"`
	GasPrice         uint64 `bun:"type:bigint,notnull"`
	Nonce            uint64 `bun:"type:bigint,notnull"`
	TransactionIndex uint64 `bun:"type:integer,notnull"`
	Value            string `bun:"type:varchar"`
	ContractAddress  string `bun:"type:char(42)"`
	Status           uint64 `bun:"type:smallint,notnull"`
	Timestamp        uint64 `bun:"type:bigint,notnull"` // copy block timestamp
	// TransactionAction - Istražiti šta je ovo
}

// Events - Events emitted from smart contracts to be held in this table
// type Event struct {
// 	BlockHash       string         `bun:",pk"`
// 	Index           uint           `bun:",pk,type:integer"`
// 	Origin          string         `bun:"type:char(42),notnull"`
// 	Topics          pq.StringArray `bun:"type:text[],notnull"`
// 	Data            []byte
// 	TransactionHash string `bun:"type:char(66),notnull"`
// }
