package syncer

import (
	"context"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

type JobArgs struct {
	BlockNumbers         []uint64
	Client               *rpc.Client
	Db                   *bun.DB
	Step                 uint
	CallTimeoutInSeconds uint
}

type JobResult struct {
	Blocks       []*db.Block
	Transactions []*db.Transaction
}

var (
	execFn = func(ctx context.Context, args interface{}) interface{} {
		jobArgs, ok := args.(JobArgs)
		if !ok {
			logrus.Panic("Wrong type for args parameter")
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

		return JobResult{Blocks: dbBlocks, Transactions: dbTransactions}
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

	step := jobArgs.Step
	if len(elems) != 0 {
		totalCounter := uint(math.Ceil(float64(len(elems)) / float64(step)))
		var i uint
		for i = 0; i < totalCounter; i++ {

			from := i * step
			to := int(math.Min(float64(len(elems)), float64((i+1)*step)))

			elemSlice := elems[from:to]
			for {
				ioErr := batchCallWithTimeout(&elemSlice, *jobArgs.Client, jobArgs.CallTimeoutInSeconds, ctx)
				if ioErr != nil {
					logrus.Error("Cannot get transactions from blockchain, err: ", ioErr)
					continue
				}
				if transactions[0].Hash != "" {
					break
				}
			}
		}
	}

	for _, e := range errors {
		if e != nil {
			logrus.Error("Error during batch call, err: ", e.Error())
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

	for {
		ioErr := batchCallWithTimeout(&elems, *jobArgs.Client, jobArgs.CallTimeoutInSeconds, ctx)
		if ioErr != nil {
			logrus.Error("Cannot get blocks from blockchain, err: ", ioErr)
			continue
		}
		if blocks[0].Number != "" {
			break
		}
	}

	for _, e := range errors {
		if e != nil {
			logrus.Error("Error during batch call, err: ", e.Error())
		}
	}

	return blocks
}

func batchCallWithTimeout(elems *[]rpc.BatchElem, client rpc.Client, callTimeoutInSeconds uint, ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(callTimeoutInSeconds)*time.Second)
	defer cancel()
	return client.BatchCallContext(ctxWithTimeout, *elems)
}
