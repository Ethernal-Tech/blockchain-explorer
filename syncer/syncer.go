package syncer

import (
	"context"
	"ethernal/explorer/config"
	"ethernal/explorer/workers"
	"log"
	"math"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	bundb "github.com/uptrace/bun"
)

var step int

func SyncMissingBlocks(client *rpc.Client, db *bundb.DB, config config.Config) {

	missingBlocks := []uint64{}
	var blocks uint64 = 1000000
	var i uint64
	for i = 0; i < blocks; i++ {
		missingBlocks = append(missingBlocks, i+1)
	}

	step = config.Step
	wp := workers.New(config.WorkersCount)

	totalCounter := int(math.Ceil(float64(blocks) / float64(step)))
	counter := 0

	//TEST START

	// missingBlocks = []uint64{7785381}
	// totalCounter = 1

	//TEST END

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	var wg sync.WaitGroup

	go wp.GenerateFrom(createJobs(missingBlocks, client, db))

	go wp.Run(ctx, &wg)

	for {
		select {
		case result, ok := <-wp.Results():
			if !ok {
				log.Println("[ERROR] ", result.Err)
				continue
			}

			counter++
			val := result.Value.(JobResult)

			_, blockError := db.NewInsert().Model(&val.Blocks).Exec(ctx)
			if blockError != nil {
				log.Println(blockError)
			}

			if len(val.Transactions) != 0 {
				_, transError := db.NewInsert().Model(&val.Transactions).Exec(ctx)
				if transError != nil {
					log.Println(transError)
				}
			}

			//log.Println("Counter result after: ", counter)
			if counter == totalCounter {
				wg.Done()
			}
		case <-wp.Done:
			log.Println("DONE")
			return
		}
	}

}

func createJobs(missingBlocks []uint64, client *rpc.Client, db *bundb.DB) []workers.Job {
	jobsCount := int(math.Ceil(float64(len(missingBlocks)) / float64(step)))
	jobs := make([]workers.Job, jobsCount)

	for i := 0; i < jobsCount; i++ {

		end := int(math.Min(float64(len(missingBlocks)), float64((i+1)*step)))

		jobs[i] = workers.Job{
			ExecFn: execFn,
			Args: JobArgs{
				BlockNumbers: missingBlocks[i*step : end],
				Client:       client,
				Db:           db,
			},
		}
	}
	return jobs
}
