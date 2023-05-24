package eth

import (
	"ethernal/explorer/common"
	"ethernal/explorer/db"
	"ethernal/explorer/utils"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereumCommon "github.com/ethereum/go-ethereum/common"
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

func CreateDbContract(receipt *TransactionReceipt) db.Contract {
	return db.Contract{
		Address:         receipt.ContractAddress,
		TransactionHash: receipt.TransactionHash,
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

func CreateDbNfts(transaction *Transaction, receipt *TransactionReceipt) ([]*db.Nft, error) {
	var nfts []*db.Nft
	for _, log := range receipt.Logs {
		if len(log.Topics) == 4 && log.Topics[0] == common.Erc721TransferEvent.Signature {
			parsedLog := &Erc721Transfer{}
			if err := parseLog(parsedLog, log, common.Erc721TransferEvent.Name, common.Erc721TransferEvent.Abi); err != nil {
				return nil, err
			}

			nft := &db.Nft{
				BlockHash:       log.BlockHash,
				Index:           utils.ToUint32(log.LogIndex),
				TransactionHash: log.TransactionHash,
				Address:         log.Address,
				From:            parsedLog.From.String(),
				To:              parsedLog.To.String(),
				TokenId:         parsedLog.TokenId.String(),
				TokenTypeId:     common.ERC721Type,
			}
			nfts = append(nfts, nft)
		} else if len(log.Topics) == 4 && log.Topics[0] == common.Erc1155TransferSingleEvent.Signature {
			parsedLog := &Erc1155TransferSingle{}
			if err := parseLog(parsedLog, log, common.Erc1155TransferSingleEvent.Name, common.Erc1155TransferSingleEvent.Abi); err != nil {
				return nil, err
			}
			nft := &db.Nft{
				BlockHash:       log.BlockHash,
				Index:           utils.ToUint32(log.LogIndex),
				TransactionHash: log.TransactionHash,
				Address:         log.Address,
				From:            parsedLog.From.String(),
				To:              parsedLog.To.String(),
				TokenId:         parsedLog.Id.String(),
				Value:           parsedLog.Value.String(),
				TokenTypeId:     common.ERC1155Type,
			}
			nfts = append(nfts, nft)
		} else if len(log.Topics) == 4 && log.Topics[0] == common.Erc1155TransferBatchEvent.Signature {
			parsedLog := &Erc1155TransferBatch{}

			if err := parseLog(parsedLog, log, common.Erc1155TransferBatchEvent.Name, common.Erc1155TransferBatchEvent.Abi); err != nil {
				return nil, err
			}

			for index, id := range parsedLog.Ids {
				nft := &db.Nft{
					BlockHash:       log.BlockHash,
					Index:           utils.ToUint32(log.LogIndex),
					TransactionHash: log.TransactionHash,
					Address:         log.Address,
					From:            parsedLog.From.String(),
					To:              parsedLog.To.String(),
					TokenId:         id.String(),
					Value:           parsedLog.Values[index].String(),
					TokenTypeId:     common.ERC1155Type,
				}
				nfts = append(nfts, nft)
			}
		} else {
			continue
		}

	}
	return nfts, nil
}

func parseLog(out interface{}, log Log, eventName string, eventAbi string) error {
	parsedAbi, _ := abi.JSON(strings.NewReader("[" + eventAbi + "]"))
	event := parsedAbi.Events[eventName]
	if ethereumCommon.BytesToHash(ethereumCommon.Hex2Bytes((log.Topics[0][2:]))) != event.ID {
		return fmt.Errorf("event signature mismatch")
	}
	if len(log.Data) > 0 {
		if err := parsedAbi.UnpackIntoInterface(out, eventName, ethereumCommon.Hex2Bytes(log.Data[2:])); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range event.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	var topics []ethereumCommon.Hash
	for i := 1; i < len(log.Topics); i++ {
		topics = append(topics, ethereumCommon.BytesToHash(ethereumCommon.Hex2Bytes(log.Topics[i][2:])))
	}
	return abi.ParseTopics(out, indexed, topics)
}
