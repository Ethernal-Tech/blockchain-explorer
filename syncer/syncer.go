package syncer

import (
	"context"
	"ethernal/explorer/common"
	"ethernal/explorer/config"
	"ethernal/explorer/eth"
	"ethernal/explorer/utils"
	"ethernal/explorer/workers"
	"log"
	"math"
	"sync"
	"time"

	"database/sql"

	"github.com/ethereum/go-ethereum/rpc"
	bundb "github.com/uptrace/bun"
)

func SyncMissingBlocks(client *rpc.Client, db *bundb.DB, config config.Config) {
	startingAt := time.Now().UTC()
	log.Print("Synchronization started")
	// only for automatic mode - when synch is finished send a signal in channel Done
	if config.Mode == common.Automatic {
		defer func() {
			synch := GetSignalSynchInstance()
			synch.Done <- struct{}{}
		}()
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	wp := workers.New(config.WorkersCount)

	missingBlocks := getMissingBlocks(ctx, client, db, config.CallTimeoutInSeconds)
	log.Println("Number of missing blocks: ", len(missingBlocks))
	if len(missingBlocks) == 0 {
		return
	}

	totalCounter := int(math.Ceil(float64(len(missingBlocks)) / float64(config.Step)))
	counter := 0

	var wg sync.WaitGroup

	go wp.GenerateFrom(createJobs(missingBlocks, client, db, config))

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

			// inserting blocks and transactions in one transaction scope
			err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bundb.Tx) error {
				_, blockError := tx.NewInsert().Model(&val.Blocks).Exec(ctx)
				if blockError != nil {
					return blockError
				}

				if len(val.Transactions) != 0 {
					_, transError := tx.NewInsert().Model(&val.Transactions).Exec(ctx)
					if transError != nil {
						return transError
					}
				}

				return nil
			})

			if err != nil {
				log.Println(err)
			}

			if counter == totalCounter {
				wg.Done()
			}
		case <-wp.Done:
			log.Println("Synchronization DONE")
			log.Printf("Took: %s", time.Now().UTC().Sub(startingAt))
			return
		}
	}
}

func createJobs(missingBlocks []uint64, client *rpc.Client, db *bundb.DB, config config.Config) []workers.Job {
	step := config.Step
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
				Step:                 config.Step,
				CallTimeoutInSeconds: config.CallTimeoutInSeconds,
			},
		}
	}
	return jobs
}

func getMissingBlocks(ctx context.Context, client *rpc.Client, db *bundb.DB, callTimeoutInSeconds uint) []uint64 {
	blockNumberFromChain := getLastBlockFromChain(ctx, client, callTimeoutInSeconds)
	blockNumbersFromDb := []uint64{}
	db.NewSelect().Table("blocks").Column("number").Order("number ASC").Scan(ctx, &blockNumbersFromDb)
	mb := findMissingBlocks(blockNumberFromChain, &blockNumbersFromDb)

	// log.Println("Missing blocks", len(mb))
	// log.Println(mb[0])
	// log.Println(mb[len(mb)-1])

	return mb
}

func getLastBlockFromChain(ctx context.Context, client *rpc.Client, callTimeoutInSeconds uint) uint64 {
	var latestBlock uint64 = 0
	for {
		block, err := getLatestBlockFromChainWithTimeout(ctx, client, callTimeoutInSeconds)
		if err != nil {
			log.Println("Get latest block IO Error: ", err)
			continue
		}
		if block.Number != "" {
			latestBlock = utils.ToUint64(block.Number)
			break
		}
	}
	return latestBlock
}

func getLatestBlockFromChainWithTimeout(ctx context.Context, client *rpc.Client, callTimeoutInSeconds uint) (eth.Block, error) {
	block := eth.Block{}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(callTimeoutInSeconds)*time.Second)
	defer cancel()
	err := client.CallContext(ctxWithTimeout, &block, "eth_getBlockByNumber", "latest", false)
	return block, err
}

func findMissingBlocks(blockNumberFromChain uint64, blockNumbersFromDb *[]uint64) []uint64 {
	missingBlocks := []uint64{}

	var i uint64
	if len(*blockNumbersFromDb) == 0 {
		for i = 1; i <= blockNumberFromChain; i++ {
			missingBlocks = append(missingBlocks, i)
		}
		return missingBlocks
	}

	counter := 0
	for i = 1; i <= (*blockNumbersFromDb)[len(*blockNumbersFromDb)-1]; i++ {
		if i < (*blockNumbersFromDb)[counter] {
			missingBlocks = append(missingBlocks, i)
		} else {
			counter++
		}
	}

	for i = (*blockNumbersFromDb)[len(*blockNumbersFromDb)-1] + 1; i <= blockNumberFromChain; i++ {
		missingBlocks = append(missingBlocks, i)
	}

	return missingBlocks
}
