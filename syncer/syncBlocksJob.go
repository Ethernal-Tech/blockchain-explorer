package syncer

import (
	"context"
	"errors"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/uptrace/bun"
)

type JobArgs struct {
	BlockNumbers         []uint64
	Client               *rpc.Client
	Db                   *bun.DB
	CallTimeoutInSeconds uint
}

type JobResult struct {
	Blocks       []*db.Block
	Transactions []*db.Transaction
}

var (
	errDefault = errors.New("wrong argument type")
	execFn     = func(ctx context.Context, args interface{}) (interface{}, error) {
		jobArgs, ok := args.(JobArgs)
		if !ok {
			return nil, errDefault
		}

		blocks := GetBlocks(jobArgs, ctx)
		transactions, receipts := GetTransactions(blocks, jobArgs, ctx)

		dbBlocks := make([]*db.Block, len(blocks))
		for i, b := range blocks {
			dbBlocks[i] = eth.CreateDbBlock(b)
		}

		dbTransactions := make([]*db.Transaction, len(transactions))
		for i, t := range transactions {
			dbTransactions[i] = eth.CreateDbTransaction(t, receipts[i])
		}

		return JobResult{Blocks: dbBlocks, Transactions: dbTransactions}, nil
	}
)

func GetTransactions(blocks []*eth.Block, jobArgs JobArgs, ctx context.Context) ([]*eth.Transaction, []*eth.TransactionReceipt) {
	var transactions []*eth.Transaction
	var receipts []*eth.TransactionReceipt
	var errors []error
	var elems []rpc.BatchElem

	for _, block := range blocks {
		if len(block.Transactions) == 0 {
			continue
		}

		for _, transHash := range block.Transactions {
			transaction := &eth.Transaction{
				Timestamp: block.Timestamp,
			}
			receipt := &eth.TransactionReceipt{}
			err1 := error(nil)
			err2 := error(nil)

			elems = append(elems, rpc.BatchElem{
				Method: "eth_getTransactionByHash",
				Args:   []interface{}{transHash},
				Result: transaction,
				Error:  err1,
			})
			elems = append(elems, rpc.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{transHash},
				Result: receipt,
				Error:  err2,
			})

			transactions = append(transactions, transaction)
			receipts = append(receipts, receipt)
			errors = append(errors, err1)
			errors = append(errors, err2)
		}
	}

	if len(elems) != 0 {
		for {
			ioErr := batchCallWithTimeout(&elems, *jobArgs.Client, jobArgs.CallTimeoutInSeconds, ctx)
			if ioErr != nil {
				log.Println("Get transations IO Error", ioErr)
			}
			if transactions[0].Hash != "" {
				break
			}
		}
	}

	for _, e := range errors {
		if e != nil {
			log.Println("Error batch call: ", e.Error())
		}
	}

	return transactions, receipts
}

func GetBlocks(jobArgs JobArgs, ctx context.Context) []*eth.Block {
	var blocks []*eth.Block
	errors := make([]error, 0, len(jobArgs.BlockNumbers))
	elems := make([]rpc.BatchElem, 0, len(jobArgs.BlockNumbers))

	for _, blockNumber := range jobArgs.BlockNumbers {
		block := &eth.Block{}
		err := error(nil)

		elems = append(elems, rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{string(hexutil.EncodeBig(big.NewInt(int64(blockNumber)))), false},
			Result: block,
			Error:  err,
		})

		blocks = append(blocks, block)
		errors = append(errors, err)
	}

	//log.Println("Before batch call: [", jobArgs.BlockNumbers[0], ":", jobArgs.BlockNumbers[len(jobArgs.BlockNumbers)-1], "]")

	for {
		ioErr := batchCallWithTimeout(&elems, *jobArgs.Client, jobArgs.CallTimeoutInSeconds, ctx)
		if ioErr != nil {
			log.Println("Get blocks IO Error: ", ioErr)
			continue
		}
		if blocks[0].Number != "" {
			break
		}
	}

	//log.Println("After batch call: ", jobArgs.BlockNumbers[0])

	for _, e := range errors {
		if e != nil {
			log.Println("Error batch call: ", e.Error())
		}
	}

	return blocks
}

func batchCallWithTimeout(elems *[]rpc.BatchElem, client rpc.Client, callTimeoutInSeconds uint, ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(callTimeoutInSeconds)*time.Second)
	defer cancel()
	return client.BatchCallContext(ctxWithTimeout, *elems)
}
