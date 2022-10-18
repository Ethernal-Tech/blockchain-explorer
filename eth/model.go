package eth

import (
	"ethernal/explorer/db"
	"log"
	"strconv"
)

type Block struct {
	Hash       string
	Number     string
	ParentHash string
	Nonce      string
	// Sha3Uncles string
	// LogsBloom        string
	// TransactionsRoot string
	// StateRoot        string
	// ReceiptsRoot     string
	Miner           string
	Difficulty      string
	TotalDifficulty string
	ExtraData       string
	Size            string
	GasLimit        string
	GasUsed         string
	Timestamp       string
	Transactions    []string
	// Uncles           []string
	// MixHash string
}

type Transaction struct {
	Hash        string
	BlockHash   string
	BlockNumber string
	From        string
	To          string
	Gas         string
	GasPrice    string
	// Input            string
	Nonce            string
	TransactionIndex string
	Value            string
	// V                string
	// S                string
	// R                string
	Timestamp string // For DB only
}

type TransactionReceipt struct {
	TransactionHash  string
	TransactionIndex string
	BlockHash        string
	BlockNumber      string
	// From             string
	// To                string
	CumulativeGasUsed string
	GasUsed           string
	ContractAddress   string
	// Logs              string
	// LogsBloom         string
	// Root   string
	Status string
}

func CreateDbBlock(block *Block) *db.Block {
	return &db.Block{
		Hash:            block.Hash,
		Number:          ToUint64(block.Number),
		ParentHash:      block.ParentHash,
		Nonce:           block.Nonce,
		Miner:           block.Miner,
		Difficulty:      block.Difficulty,
		TotalDifficulty: block.TotalDifficulty,
		ExtraData:       []byte(block.ExtraData),
		Size:            ToUint64(block.Size),
		GasLimit:        ToUint64(block.GasLimit),
		GasUsed:         ToUint64(block.GasUsed),
		Timestamp:       ToUint64(block.Timestamp),
	}
}

func CreateDbTransaction(transaction *Transaction, receipt *TransactionReceipt) *db.Transaction {
	if transaction.BlockHash != receipt.BlockHash ||
		transaction.BlockNumber != receipt.BlockNumber ||
		transaction.TransactionIndex != receipt.TransactionIndex ||
		transaction.Hash != receipt.TransactionHash {
		log.Println("Error converting transaction and receipt to DbTransaction")
		return &db.Transaction{}
	}

	return &db.Transaction{
		Hash:             transaction.Hash,
		BlockHash:        transaction.BlockHash,
		BlockNumber:      ToUint64(transaction.BlockNumber),
		From:             transaction.From,
		To:               transaction.To,
		Gas:              ToUint64(transaction.Gas),
		GasUsed:          ToUint64(receipt.GasUsed),
		GasPrice:         ToUint64(transaction.GasPrice),
		Nonce:            ToUint64(transaction.Nonce),
		TransactionIndex: ToUint64(transaction.TransactionIndex),
		Value:            transaction.Value,
		ContractAddress:  receipt.ContractAddress,
		Status:           ToUint64(receipt.Status),
		Timestamp:        ToUint64(transaction.Timestamp),
	}
}

func ToUint64(str string) uint64 {
	if len(str) <= 2 {
		return 0
	}

	var res uint64
	var err error

	if str[0:2] == "0x" {
		res, err = strconv.ParseUint(str[2:], 16, 64)
		if err != nil {
			log.Printf("Error converting %s to uint64. %s", str, err)
			return 0
		}
	} else {
		res, err = strconv.ParseUint(str, 10, 64)
		if err != nil {
			log.Printf("Error converting %s to uint64. %s", str, err)
			return 0
		}
	}
	return res
}
