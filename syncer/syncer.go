package syncer

import (
	"context"
	"ethernal/explorer/workers"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/uptrace/bun"
)

func SyncMissingBlocks(missingBlocks []uint64, ethClient *ethclient.Client, db *bun.DB) {
	wp := workers.New(4)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go wp.GenerateFrom(createJobs(missingBlocks, ethClient, db))

	go wp.Run(ctx)

	for {
		select {
		case r, ok := <-wp.Results():
			if !ok {
				//log.Println("[ERROR] ", r.Err)
				continue
			}

			val := r.Value.(*types.Block)
			log.Println(val.Hash())

		case <-wp.Done:
			return
		default:
		}
	}

}

func createJobs(missingBlocks []uint64, ethClient *ethclient.Client, db *bun.DB) []workers.Job {
	jobsCount := len(missingBlocks)
	jobs := make([]workers.Job, jobsCount)

	for i := 0; i < jobsCount; i++ {
		jobs[i] = workers.Job{
			ExecFn: execFn,
			Args: JobArgs{
				BlockNumber: missingBlocks[i],
				EthClient:   ethClient,
				Db:          db,
			},
		}
	}
	return jobs
}
