package eth

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"ethernal/explorer/common"
	"ethernal/explorer/db"
	"ethernal/explorer/utils"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereumCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	bundb "github.com/uptrace/bun"
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

type NftMetadata struct {
	Name                  string                 `json:"name"`
	Image                 string                 `json:"image"`
	Description           string                 `json:"description"`
	NftMetadataAttributes []NftMetadataAttribute `json:"attributes"`
}

type NftMetadataAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
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

func CreateDbNfts(receipt *TransactionReceipt, client rpc.Client, timeout uint, ipfsGateway string, bunDb *bundb.DB, ctx context.Context) ([]*db.NftTransfer, error) {
	var dbNftTransfers []*db.NftTransfer
	for _, log := range receipt.Logs {
		if len(log.Topics) == 4 && log.Topics[0] == common.Erc721TransferEvent.Signature {
			parsedLog := &Erc721Transfer{}
			if err := parseLog(parsedLog, log, common.Erc721TransferEvent.Name, common.Erc721TransferEvent.Abi); err != nil {
				return nil, err
			}

			nftTransfer := &db.NftTransfer{
				BlockHash:       log.BlockHash,
				Index:           utils.ToUint32(log.LogIndex),
				BlockNumber:     utils.ToUint64(log.BlockNumber),
				TransactionHash: log.TransactionHash,
				Address:         log.Address,
				From:            parsedLog.From.String(),
				To:              parsedLog.To.String(),
				TokenId:         parsedLog.TokenId.String(),
				TokenTypeId:     common.ERC721Type,
			}

			dbNftTransfers = append(dbNftTransfers, nftTransfer)

			// nft mint
			if parsedLog.From.String() == "0x0000000000000000000000000000000000000000" {
				getNftMetadata(nftTransfer, client, timeout, ipfsGateway, bunDb, ctx)
			}
		} else if len(log.Topics) == 4 && log.Topics[0] == common.Erc1155TransferSingleEvent.Signature {
			parsedLog := &Erc1155TransferSingle{}
			if err := parseLog(parsedLog, log, common.Erc1155TransferSingleEvent.Name, common.Erc1155TransferSingleEvent.Abi); err != nil {
				return nil, err
			}
			nftTransfer := &db.NftTransfer{
				BlockHash:       log.BlockHash,
				Index:           utils.ToUint32(log.LogIndex),
				BlockNumber:     utils.ToUint64(log.BlockNumber),
				TransactionHash: log.TransactionHash,
				Address:         log.Address,
				From:            parsedLog.From.String(),
				To:              parsedLog.To.String(),
				TokenId:         parsedLog.Id.String(),
				Value:           parsedLog.Value.String(),
				TokenTypeId:     common.ERC1155Type,
			}

			dbNftTransfers = append(dbNftTransfers, nftTransfer)

			// nft mint
			if parsedLog.From.String() == "0x0000000000000000000000000000000000000000" {
				getNftMetadata(nftTransfer, client, timeout, ipfsGateway, bunDb, ctx)
			}
		} else if len(log.Topics) == 4 && log.Topics[0] == common.Erc1155TransferBatchEvent.Signature {
			parsedLog := &Erc1155TransferBatch{}

			if err := parseLog(parsedLog, log, common.Erc1155TransferBatchEvent.Name, common.Erc1155TransferBatchEvent.Abi); err != nil {
				return nil, err
			}

			mint := parsedLog.From.String() == "0x0000000000000000000000000000000000000000"

			for index, id := range parsedLog.Ids {
				nftTransfer := &db.NftTransfer{
					BlockHash:       log.BlockHash,
					Index:           utils.ToUint32(log.LogIndex),
					BlockNumber:     utils.ToUint64(log.BlockNumber),
					TransactionHash: log.TransactionHash,
					Address:         log.Address,
					From:            parsedLog.From.String(),
					To:              parsedLog.To.String(),
					TokenId:         id.String(),
					Value:           parsedLog.Values[index].String(),
					TokenTypeId:     common.ERC1155Type,
				}
				dbNftTransfers = append(dbNftTransfers, nftTransfer)

				// nft mint
				if mint {
					getNftMetadata(nftTransfer, client, timeout, ipfsGateway, bunDb, ctx)
				}
			}
		} else {
			continue
		}
	}
	return dbNftTransfers, nil
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

func getNftMetadata(nftTransfer *db.NftTransfer, client rpc.Client, timeout uint, ipfsGateway string, bunDb *bundb.DB, ctx context.Context) {
	var metadataId *uint64
	bunDb.NewSelect().Table("nft_metadata").Column("id").Where("token_id = ? AND address = ?", nftTransfer.TokenId, nftTransfer.Address).Scan(ctx, &metadataId)

	// if it does not exist in the database, check if it is in the dictionary
	if metadataId == nil {
		dictionary := GetMetadataDictionaryInstance()
		key := nftTransfer.TokenId + "-" + nftTransfer.Address

		// we start processing metadata only if it has been added to the dictionary (if another goroutine has not already started processing metadata for the same nft)
		if added := dictionary.TryAdd(key, true); added {
			go processNftMetadata(nftTransfer, client, timeout, ipfsGateway, bunDb)
		}
	}
}

func processNftMetadata(nftTransfer *db.NftTransfer, client rpc.Client, timeout uint, ipfsGateway string, bunDb *bundb.DB) {
	metadata := &NftMetadata{
		NftMetadataAttributes: make([]NftMetadataAttribute, 0),
	}
	dbNftMetadataAttributes := []*db.NftMetadataAttribute{}
	var elems []rpc.BatchElem
	var metadataUrl string
	var data []byte
	type params struct {
		To   string `json:"to"`
		Data string `json:"data"`
	}

	if nftTransfer.TokenTypeId == common.ERC721Type {
		parsedAbi, _ := abi.JSON(strings.NewReader("[" + common.TokenUriMethod.Abi + "]"))
		tokenId := new(big.Int)
		tokenId.SetString(nftTransfer.TokenId, 10)
		data, _ = parsedAbi.Pack(common.TokenUriMethod.Name, tokenId)
	} else if nftTransfer.TokenTypeId == common.ERC1155Type {
		parsedAbi, _ := abi.JSON(strings.NewReader("[" + common.UriMethod.Abi + "]"))
		tokenId := new(big.Int)
		tokenId.SetString(nftTransfer.TokenId, 10)
		data, _ = parsedAbi.Pack(common.UriMethod.Name, tokenId)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	elems = append(elems, rpc.BatchElem{
		Method: "eth_call",
		Args:   []interface{}{params{nftTransfer.Address, "0x" + hex.EncodeToString(data)}, "latest"},
		Result: &metadataUrl,
	})

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	err := client.BatchCallContext(ctxWithTimeout, elems)
	if err != nil {
		logrus.Error("Cannot get metadata url from blockchain, err: ", err)
	}

	if len(metadataUrl) > 0 && metadataUrl[0:2] == "0x" {
		metadataUrl = metadataUrl[2:]
	}

	bs, _ := hex.DecodeString(metadataUrl)
	re := regexp.MustCompile("[^a-zA-Z0-9:// -.]+")
	str := re.ReplaceAllString(string(bs), "")
	r, _ := regexp.Compile(`(?P<protocol>\w+):\/\/(?P<route>.*)`)
	m := r.FindStringSubmatch(str)

	if len(m) > 0 {
		result := make(map[string]string)
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = m[i]
			}
		}
		protocol := result["protocol"]

		if strings.Contains(protocol, "ipfs") {
			url := ipfsGateway + result["route"]
			getJson(url, metadata, timeout)
		} else if strings.Contains(protocol, "http") {
			url := "http://" + result["route"]
			getJson(url, metadata, timeout)
		} else if strings.Contains(protocol, "https") {
			url := "https://" + result["route"]
			getJson(url, metadata, timeout)
		}
	}
	dbNftMetadata := db.NftMetadata{
		TokenId:     nftTransfer.TokenId,
		Address:     nftTransfer.Address,
		Name:        metadata.Name,
		Image:       metadata.Image,
		Description: metadata.Description,
	}

	for _, attribute := range metadata.NftMetadataAttributes {
		dbNftAttribute := &db.NftMetadataAttribute{
			TraitType:     attribute.TraitType,
			Value:         attribute.Value,
			NftMetadataId: &dbNftMetadata.Id,
		}
		dbNftMetadataAttributes = append(dbNftMetadataAttributes, dbNftAttribute)
	}
	dictionary := GetMetadataDictionaryInstance()
	dictionary.itemsData <- itemsData{metadata: &dbNftMetadata, attributes: dbNftMetadataAttributes}
}

func getJson(url string, target interface{}, timeout uint) error {
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	response, err := client.Get(url)
	if err != nil {
		logrus.Error("Cannot get metadata from "+url+", err: ", err)
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logrus.Error("Cannot get metadata from "+url+", err: ", err)
		return err
	}
	read, _ := ioutil.ReadAll(response.Body)
	return json.Unmarshal([]byte(read), target)
}

func SyncNftMetadata(bunDb *bundb.DB) {
	dictionary := GetMetadataDictionaryInstance()
	ctx := context.TODO()
	for job := range dictionary.itemsData {
		_ = bunDb.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bundb.Tx) error {
			_, nftMetadataError := tx.NewInsert().Model(job.metadata).Exec(ctx)
			if nftMetadataError != nil {
				logrus.Error("Error during inserting nft metadata in DB, err: ", nftMetadataError)
				return nftMetadataError
			}

			if len(job.attributes) != 0 {
				_, nftMetadataAttributeError := tx.NewInsert().Model(&job.attributes).Exec(ctx)
				if nftMetadataAttributeError != nil {
					logrus.Error("Error during inserting nft metadata attributes in DB, err: ", nftMetadataAttributeError)
					return nftMetadataAttributeError
				}
			}
			return nil
		})
		dictionary.TryRemove(job.metadata.TokenId + "-" + job.metadata.Address)
	}
}
