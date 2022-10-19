package syncer

import (
	"context"
	"ethernal/explorer/config"
	"ethernal/explorer/eth"
	"ethernal/explorer/utils"
	"ethernal/explorer/workers"
	"log"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	bundb "github.com/uptrace/bun"
)

var step uint

func SyncMissingBlocks(client *rpc.Client, db *bundb.DB, config config.Config) {

	missingBlocks := []uint64{}
	var blocks uint64 = 100000
	var i uint64
	for i = 0; i < blocks; i++ {
		missingBlocks = append(missingBlocks, i+1)
	}

	step = config.Step
	//wp := workers.New(config.WorkersCount)

	//TEST START

	//missingBlocks = []uint64{15000000}

	//TEST END

	// totalCounter := int(math.Ceil(float64(len(missingBlocks)) / float64(step)))
	// counter := 0

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	getMissingBlocks(client, db, ctx, config.CallTimeoutInSeconds)

	// var wg sync.WaitGroup

	// go wp.GenerateFrom(createJobs(missingBlocks, client, db, config.CallTimeoutInSeconds))

	// go wp.Run(ctx, &wg)

	// for {
	// 	select {
	// 	case result, ok := <-wp.Results():
	// 		if !ok {
	// 			log.Println("[ERROR] ", result.Err)
	// 			continue
	// 		}

	// 		counter++
	// 		val := result.Value.(JobResult)

	// 		_, blockError := db.NewInsert().Model(&val.Blocks).Exec(ctx)
	// 		if blockError != nil {
	// 			log.Println(blockError)
	// 		}

	// 		if len(val.Transactions) != 0 {
	// 			_, transError := db.NewInsert().Model(&val.Transactions).Exec(ctx)
	// 			if transError != nil {
	// 				log.Println(transError)
	// 			}
	// 		}

	// 		//log.Println("Counter result after: ", counter)
	// 		if counter == totalCounter {
	// 			wg.Done()
	// 		}
	// 	case <-wp.Done:
	// 		log.Println("DONE")
	// 		return
	// 	}
	// }

}

func createJobs(missingBlocks []uint64, client *rpc.Client, db *bundb.DB, callTimeoutInSeconds uint) []workers.Job {
	jobsCount := uint(math.Ceil(float64(len(missingBlocks)) / float64(step)))
	jobs := make([]workers.Job, jobsCount)
	var i uint

	for i = 0; i < jobsCount; i++ {

		end := int(math.Min(float64(len(missingBlocks)), float64((i+1)*step)))

		jobs[i] = workers.Job{
			ExecFn: execFn,
			Args: JobArgs{
				BlockNumbers:         missingBlocks[i*step : end],
				Client:               client,
				Db:                   db,
				CallTimeoutInSeconds: callTimeoutInSeconds,
			},
		}
	}
	return jobs
}

func getMissingBlocks(client *rpc.Client, db *bundb.DB, ctx context.Context, callTimeoutInSeconds uint) []uint64 {
	missingBlocks := []uint64{}
	var blockNumber uint64

	for {
		block, err := getLatestBlockFromChainWithTimeout(client, callTimeoutInSeconds, ctx)
		if err != nil {
			log.Println("Get latest block IO Error: ", err)
			continue
		}
		if block != nil {
			blockNumber = utils.ToUint64(block.Number)
			break
		}

		log.Println("Retry ")
	}
	log.Println("Latest block from chain: ", blockNumber)

	// blockNumbersFromDb := []BlockNumber{}
	// db.NewSelect().Table("blocks").Column("number").Model(&blockNumbersFromDb).Order("number ASC").Scan(ctx)

	return missingBlocks
}

func getLatestBlockFromChainWithTimeout(client *rpc.Client, callTimeoutInSeconds uint, ctx context.Context) (*eth.Block, error) {
	var block *eth.Block
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(callTimeoutInSeconds)*time.Second)
	defer cancel()
	err := client.CallContext(ctxWithTimeout, block, "eth_getBlockByNumber", "latest", false)
	return block, err
}

type BlockNumber struct {
	Number uint64
}
