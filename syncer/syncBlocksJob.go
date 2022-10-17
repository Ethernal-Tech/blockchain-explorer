package syncer

import (
	"context"
	"errors"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/uptrace/bun"
)

type JobArgs struct {
	BlockNumbers []uint64
	Client       *rpc.Client
	Db           *bun.DB
}

type JobResult struct {
	Blocks []*db.Block
}

var (
	errDefault = errors.New("wrong argument type")
	execFn     = func(ctx context.Context, args interface{}) (interface{}, error) {
		jobArgs, ok := args.(JobArgs)
		if !ok {
			return nil, errDefault
		}

		var blocks []*eth.Block

		elems := make([]rpc.BatchElem, 0, len(jobArgs.BlockNumbers))
		errors := make([]error, 0, len(jobArgs.BlockNumbers))

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
			//log.Println("Enter: ", jobArgs.BlockNumbers[0])
			ioErr := jobArgs.Client.BatchCall(elems)
			//log.Println("Exit: ", jobArgs.BlockNumbers[0])
			if ioErr != nil {
				log.Println("Error", ioErr)
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

		dbBlocks := make([]*db.Block, len(blocks))

		for i, b := range blocks {
			dbBlocks[i] = b.ToDbBlock()
		}

		//log.Println("SENT: ", jobArgs.BlockNumbers[0])
		return JobResult{Blocks: dbBlocks}, nil
	}
)
