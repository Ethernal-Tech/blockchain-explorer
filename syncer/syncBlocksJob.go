package syncer

import (
	"context"
	"errors"
	"log"
	"math/big"

	"ethernal/explorer/db"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/uptrace/bun"
)

type JobArgs struct {
	BlockNumber uint64
	EthClient   *ethclient.Client
	Db          *bun.DB
}

var (
	errDefault = errors.New("wrong argument type")
	execFn     = func(ctx context.Context, args interface{}) (interface{}, error) {
		jobArgs, ok := args.(JobArgs)
		if !ok {
			return nil, errDefault
		}

		var block *types.Block
		var err error

		for {
			block, err = jobArgs.EthClient.BlockByNumber(ctx, big.NewInt(int64(jobArgs.BlockNumber)))
			if err != nil {
				log.Println(err)
			} else {
				break
			}
		}

		dbBlock := db.Block{
			Hash:   block.Hash().String(),
			Number: block.NumberU64(),
		}
		jobArgs.Db.NewInsert().Model(&dbBlock).Exec(ctx)

		return block, nil
	}
)
