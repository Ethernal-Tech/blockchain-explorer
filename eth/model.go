package eth

import (
	"ethernal/explorer/db"
	"ethernal/explorer/utils"

	"github.com/sirupsen/logrus"
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
	Hash             string
	BlockHash        string
	BlockNumber      string
	From             string
	To               string
	Gas              string
	GasPrice         string
	Input            string
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
	Logs              []Log
	// LogsBloom         string
	// Root   string
	Status string
}

type Log struct {
	Address          string
	Topics           []string
	Data             string
	BlockNumber      string
	TransactionHash  string
	TransactionIndex string
	BlockHash        string
	LogIndex         string
	//Removed          bool
}

func CreateDbBlock(block *Block) *db.Block {
	return &db.Block{
		Hash:              block.Hash,
		Number:            utils.ToUint64(block.Number),
		ParentHash:        block.ParentHash,
		Nonce:             block.Nonce,
		Miner:             block.Miner,
		Difficulty:        block.Difficulty,
		TotalDifficulty:   block.TotalDifficulty,
		ExtraData:         []byte(block.ExtraData),
		Size:              utils.ToUint64(block.Size),
		GasLimit:          utils.ToUint64(block.GasLimit),
		GasUsed:           utils.ToUint64(block.GasUsed),
		Timestamp:         utils.ToUint64(block.Timestamp),
		TransactionsCount: len(block.Transactions),
	}
}

func CreateDbTransaction(transaction *Transaction, receipt *TransactionReceipt) *db.Transaction {
	if transaction.BlockHash != receipt.BlockHash ||
		transaction.BlockNumber != receipt.BlockNumber ||
		transaction.TransactionIndex != receipt.TransactionIndex ||
		transaction.Hash != receipt.TransactionHash {
		logrus.Panic("Error converting transaction and receipt to DbTransaction")
		return &db.Transaction{}
	}

	return &db.Transaction{
		Hash:             transaction.Hash,
		BlockHash:        transaction.BlockHash,
		BlockNumber:      utils.ToUint64(transaction.BlockNumber),
		From:             transaction.From,
		To:               transaction.To,
		Gas:              utils.ToUint64(transaction.Gas),
		GasUsed:          utils.ToUint64(receipt.GasUsed),
		GasPrice:         utils.ToUint64(transaction.GasPrice),
		Nonce:            utils.ToUint64(transaction.Nonce),
		TransactionIndex: utils.ToUint64(transaction.TransactionIndex),
		Value:            transaction.Value,
		ContractAddress:  receipt.ContractAddress,
		Status:           utils.ToUint64(receipt.Status),
		Timestamp:        utils.ToUint64(transaction.Timestamp),
		InputData:        transaction.Input,
	}
}

func CreateDbLog(transaction *Transaction, receipt *TransactionReceipt) []*db.Log {
	if transaction.BlockHash != receipt.BlockHash ||
		transaction.BlockNumber != receipt.BlockNumber ||
		transaction.TransactionIndex != receipt.TransactionIndex ||
		transaction.Hash != receipt.TransactionHash {
		logrus.Panic("Error converting transaction and receipt to DbLog")
		return []*db.Log{}
	}

	var logs []*db.Log

	for i := 0; i < len(receipt.Logs); i++ {
		log := &db.Log{
			Address:         receipt.Logs[i].Address,
			Data:            receipt.Logs[i].Data,
			BlockNumber:     utils.ToUint64(receipt.Logs[i].BlockNumber),
			TransactionHash: receipt.Logs[i].TransactionHash,
			BlockHash:       receipt.Logs[i].BlockHash,
			Index:           utils.ToUint32(receipt.Logs[i].LogIndex),
		}
		for j, topic := range receipt.Logs[i].Topics {
			switch j {
			case 0:
				log.Topic0 = topic
			case 1:
				log.Topic1 = topic
			case 2:
				log.Topic2 = topic
			case 3:
				log.Topic3 = topic
			}
		}
		logs = append(logs, log)
	}

	return logs
}
