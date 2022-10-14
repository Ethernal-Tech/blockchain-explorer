package syncer

import (
	"context"
	"ethernal/explorer/workers"
	"log"
	"math"

	"github.com/ethereum/go-ethereum/rpc"
	bundb "github.com/uptrace/bun"
)

func SyncMissingBlocks(missingBlocks []uint64, client *rpc.Client, db *bundb.DB) {
	wp := workers.New(32)
	// wp := workers.New(2)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go wp.GenerateFrom(createJobs(missingBlocks, client, db))

	go wp.Run(ctx)

	// dbBlocks := []*pgdb.Block{}
	// dbAllTransactions := []*pgdb.Transaction{}

	for {
		select {
		case result, ok := <-wp.Results():
			if !ok {
				//log.Println("[ERROR] ", r.Err)
				continue
			}

			val := result.Value.(JobResult)
			log.Println(len(val.Blocks))

			db.NewInsert().Model(&val.Blocks).Exec(ctx)

		case <-wp.Done:
			return
		default:
		}
	}

}

func createJobs(missingBlocks []uint64, client *rpc.Client, db *bundb.DB) []workers.Job {
	step := 1000
	jobsCount := int(math.Ceil(float64(len(missingBlocks)) / float64(step)))
	jobs := make([]workers.Job, jobsCount)

	for i := 0; i < jobsCount; i++ {
		jobs[i] = workers.Job{
			ExecFn: execFn,
			Args: JobArgs{
				BlockNumbers: missingBlocks[i*step : (i+1)*step],
				Client:       client,
				Db:           db,
			},
		}
	}
	return jobs
}
