package syncer

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

type JobArgs struct {
	BlockNumber uint64
	EthClient   *ethclient.Client
}

var (
	errDefault = errors.New("wrong argument type")
	execFn     = func(ctx context.Context, args interface{}) (interface{}, error) {
		jobArgs, ok := args.(JobArgs)
		if !ok {
			return nil, errDefault
		}

		for {
			block, err := jobArgs.EthClient.BlockByNumber(ctx, big.NewInt(int64(jobArgs.BlockNumber)))
			if err != nil {
				log.Println(err)
			} else {
				return block, nil
			}
		}
	}
)
