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
	Transactions     []string
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

func CreateDbBlock(block *Block) *db.Block {
	return &db.Block{
		Hash:       block.Hash,
		Number:     ToUint64(block.Number),
		Time:       ToUint64(block.Timestamp),
		ParentHash: block.ParentHash,
		Difficulty: block.TotalDifficulty,
		GasUsed:    ToUint64(block.GasUsed),
		GasLimit:   ToUint64(block.GasLimit),
		Nonce:      block.Nonce,
		Miner:      block.Miner,
		// Size:          ToFloat64(b.Size),
		StateRootHash: block.StateRoot,
		// UncleHash: b.Sha3Uncles,
		TransactionRootHash: block.TransactionsRoot,
		ReceiptRootHash:     block.ReceiptsRoot,
		ExtraData:           []byte(block.ExtraData),
	}
}

func CreateDbTransaction(transaction *Transaction, receipt *TransactionReceipt) *db.Transaction {
	return &db.Transaction{
		Hash:      transaction.Hash,
		BlockHash: transaction.BlockHash,
		From:      transaction.From,
		Gas:       ToUint64(transaction.Gas),
		GasPrice:  transaction.GasPrice,
		//Input - Data?
		Nonce: ToUint64(transaction.Nonce),
		To:    transaction.To,
		//Index:
		Value:    transaction.Value,
		State:    ToUint64(receipt.Status),
		Contract: receipt.ContractAddress,
		//Root: receipt.Root,
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
