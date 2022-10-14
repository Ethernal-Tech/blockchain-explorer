package eth

import (
	"ethernal/explorer/db"
	"log"
	"strconv"
)

type Block struct {
	Hash             string
	ParentHash       string
	Sha3Uncles       string
	Miner            string
	StateRoot        string
	TransactionsRoot string
	Size             string
	ReceiptsRoot     string
	LogsBloom        string
	TotalDifficulty  string
	Number           string
	GasLimit         string
	GasUsed          string
	Timestamp        string
	ExtraData        string
	MixHash          string
	Nonce            string
	Transactions     []Transaction
}

type Transaction struct {
	BlockHash        string
	BlockNumber      string
	From             string
	Gas              string
	GasPrice         string
	Hash             string
	Input            string
	Nonce            string
	To               string
	TransactionIndex string
	Value            string
	V                string
	S                string
	R                string
}

type TransactionReceipt struct {
	Status          string
	Root            string
	ContractAddress string
}

func (b *Block) ToDbBlock() *db.Block {
	return &db.Block{
		Hash:       b.Hash,
		Number:     ToUint64(b.Number),
		Time:       ToUint64(b.Timestamp),
		ParentHash: b.ParentHash,
		Difficulty: b.TotalDifficulty,
		GasUsed:    ToUint64(b.GasUsed),
		GasLimit:   ToUint64(b.GasLimit),
		Nonce:      b.Nonce,
		Miner:      b.Miner,
		// Size:          ToFloat64(b.Size),
		StateRootHash: b.StateRoot,
		// UncleHash: b.Sha3Uncles,
		TransactionRootHash: b.TransactionsRoot,
		ReceiptRootHash:     b.ReceiptsRoot,
		ExtraData:           []byte(b.ExtraData),
	}

}

func ToUint64(str string) uint64 {
	if len(str) <= 2 {
		return 0
	}

	res, err := strconv.ParseUint(str[2:], 16, 64)
	if err != nil {
		log.Printf("Error converting %s to uint64 ", str)
		return 0
	}
	return res
}

func ToFloat64(str string) float64 {
	if len(str) <= 2 {
		return 0
	}

	res, err := strconv.ParseFloat(str[2:], 64)
	if err != nil {
		log.Printf("Error converting %s to float64 ", str)
		return 0
	}
	return res
}
