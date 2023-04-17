package syncer

import (
	"context"
	"database/sql"
	"ethernal/explorer/common"
	"ethernal/explorer/config"
	"ethernal/explorer/db"
	"ethernal/explorer/eth"
	"ethernal/explorer/utils"
	"ethernal/explorer/workers"
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	bundb "github.com/uptrace/bun"
)

// SyncMissingBlocks keeps the database in sync with the blockchain.
func SyncMissingBlocks(client *rpc.Client, db *bundb.DB, config *config.Config) {
	startingAt := time.Now().UTC()
	logrus.Info("Synchronization started")
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

	missingBlocks, latestBlock := getMissingBlocks(ctx, client, db, config.CallTimeoutInSeconds, config.Checkpoint)
	logrus.Info("Number of missing blocks: ", len(missingBlocks))
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
				logrus.Error("err: ", result.Err)
				continue
			}

			counter++
			val, isOk := result.Value.(JobResult)
			if !isOk {
				if counter == totalCounter {
					wg.Done()
				}
				continue
			}

			// inserting blocks and transactions in one transaction scope
			_ = db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bundb.Tx) error {
				_, blockError := tx.NewInsert().Model(&val.Blocks).Exec(ctx)
				if blockError != nil {
					var numbers []uint64
					for _, b := range val.Blocks {
						numbers = append(numbers, b.Number)
					}

					logrus.Error("Error during inserting blocks with numbers ", numbers, " in DB, err: ", blockError)
					return blockError
				}

				if len(val.Transactions) != 0 {
					_, transError := tx.NewInsert().Model(&val.Transactions).Exec(ctx)
					if transError != nil {
						logrus.Error("Error during inserting transactions in DB, err: ", transError)
						return transError
					}
				}

				if len(val.Contracts) != 0 {
					_, contractsError := tx.NewInsert().Model(&val.Contracts).Exec(ctx)
					if contractsError != nil {
						logrus.Error("Error during inserting contracts in DB, err: ", contractsError)
						return contractsError
					}
				}

				if len(val.Logs) != 0 {
					_, logsError := tx.NewInsert().Model(&val.Logs).Exec(ctx)
					if logsError != nil {
						logrus.Error("Error during inserting logs in DB, err: ", logsError)
						return logsError
					}
				}

				return nil
			})

			if counter == totalCounter {
				wg.Done()
			}
		case <-wp.Done:
			// set a new checkpoint, if there are enough new blocks since the last checkpoint
			if config.Mode == common.Automatic {
				if (latestBlock - config.Checkpoint) > (uint64)(config.CheckpointWindow) {
					findNewCheckPoint(client, db, ctx, config, latestBlock)
				}
			}
			logrus.Info("Synchronization DONE")
			logrus.Info("Took: ", time.Now().UTC().Sub(startingAt))
			return
		}
	}
}

func createJobs(missingBlocks []uint64, client *rpc.Client, db *bundb.DB, config *config.Config) []workers.Job {
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
				EthLogs:              config.EthLogs,
			},
		}
	}

	logrus.Info("The number of created jobs is ", len(jobs))

	return jobs
}

// getMissingBlock returns the numbers of the missing blocks in the database and the number of the latest block on the blockchain.
func getMissingBlocks(ctx context.Context, client *rpc.Client, db *bundb.DB, callTimeoutInSeconds uint, checkpoint uint64) ([]uint64, uint64) {
	blockNumberFromChain := getLastBlockFromChain(ctx, client, callTimeoutInSeconds)
	blockNumbersFromDb := []uint64{}
	db.NewSelect().Table("blocks").Column("number").Order("number ASC").Where("number >= ?", checkpoint).Scan(ctx, &blockNumbersFromDb)
	mb := findMissingBlocks(blockNumberFromChain, &blockNumbersFromDb, checkpoint)

	return mb, blockNumberFromChain
}

func getLastBlockFromChain(ctx context.Context, client *rpc.Client, callTimeoutInSeconds uint) uint64 {
	var latestBlock uint64 = 0
	for {
		block, err := getLatestBlockFromChainWithTimeout(ctx, client, callTimeoutInSeconds)
		if err != nil {
			logrus.Error("Cannot get the latest block, err: ", err)
			continue
		}
		if block.Number != "" {
			latestBlock = utils.ToUint64(block.Number)
			logrus.Info("The number of latest block is ", latestBlock)
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

// findMissingBlock determines the missing blocks in the database and returns their numbers.
func findMissingBlocks(blockNumberFromChain uint64, blockNumbersFromDb *[]uint64, checkpoint uint64) []uint64 {
	missingBlocks := []uint64{}

	var i uint64
	if len(*blockNumbersFromDb) == 0 {
		for i = checkpoint; i <= blockNumberFromChain-1; i++ {
			missingBlocks = append(missingBlocks, i)
		}
		return missingBlocks
	}

	// check if any block is missing until the last block in the database
	counter := 0
	for i = checkpoint; i <= (*blockNumbersFromDb)[len(*blockNumbersFromDb)-1]; i++ {
		if i < (*blockNumbersFromDb)[counter] {
			missingBlocks = append(missingBlocks, i)
		} else {
			counter++
		}
	}

	for i = (*blockNumbersFromDb)[len(*blockNumbersFromDb)-1] + 1; i <= blockNumberFromChain-1; i++ {
		missingBlocks = append(missingBlocks, i)
	}

	return missingBlocks
}

// findNewCheckPoint determines the new checkpoint - starting block for the next synch.
func findNewCheckPoint(client *rpc.Client, database *bundb.DB, ctx context.Context, config *config.Config, latestBlock uint64) {
	startingAt := time.Now().UTC()
	maxBlock := latestBlock - uint64(config.CheckpointDistance)
	blocksFromDb := []db.Block{}
	// fetch numbers and hashes of the specified number of blocks
	database.NewSelect().Table("blocks").Column("number", "hash").Order("number ASC").Where("number >= ? AND number <= ?", config.Checkpoint, maxBlock).Limit(int(config.CheckpointWindow)).Scan(ctx, &blocksFromDb)
	// not enough blocks added to the database to move the checkpoint
	if (len(blocksFromDb)) <= 1 {
		return
	}

	blockNumbers := []uint64{}
	for _, block := range blocksFromDb {
		blockNumbers = append(blockNumbers, block.Number)
	}

	jobArgs := JobArgs{
		BlockNumbers:         blockNumbers,
		Client:               client,
		Db:                   database,
		Step:                 config.Step,
		CallTimeoutInSeconds: config.CallTimeoutInSeconds,
	}
	// fetch specified blocks from the blockchain
	blocksFromBlockchain := GetBlocks(jobArgs, ctx)
	if blocksFromBlockchain == nil {
		return
	}

	numbersToDelete := []uint64{}
	// compare hashes of blocks in the database with hashes on the blockchain
	// if they do not match, the block number is added for deletion
	for i := range blockNumbers {
		if blocksFromDb[i].Hash != blocksFromBlockchain[i].Hash {
			numbersToDelete = append(numbersToDelete, blocksFromDb[i].Number)
		}
	}

	if len(numbersToDelete) != 0 {
		logrus.Info("Deleting blocks: ", numbersToDelete)
		startDeletingAt := time.Now().UTC()
		// deleting blocks, transactions and logs in one transaction scope
		_ = database.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bundb.Tx) error {
			_, logError := tx.NewDelete().Table("logs").Where("block_number IN (?)", bundb.In(numbersToDelete)).Exec(ctx)
			if logError != nil {
				logrus.Error("Error during deleting logs from DB, err: ", logError)
				return logError
			}

			_, transError := tx.NewDelete().Table("transactions").Where("block_number IN (?)", bundb.In(numbersToDelete)).Exec(ctx)
			if transError != nil {
				logrus.Error("Error during deleting transactions from DB, err: ", transError)
				return transError
			}

			_, blockError := tx.NewDelete().Table("blocks").Where("number IN (?)", bundb.In(numbersToDelete)).Exec(ctx)
			if blockError != nil {
				logrus.Error("Error during deleting blocks from DB, err: ", blockError)
				return blockError
			}

			return nil
		})
		logrus.Info("Deleting took: ", time.Now().UTC().Sub(startDeletingAt))
		logrus.Info("Validation took: ", time.Now().UTC().Sub(startingAt))
		return
	}

	var i uint64
	counter := 0
	for i = config.Checkpoint; i <= (blockNumbers)[len(blockNumbers)-1]; i++ {
		if i < (blockNumbers)[counter] {
			config.Checkpoint = i
			logrus.Info("Checkpoint: ", config.Checkpoint)
			logrus.Info("Validation took: ", time.Now().UTC().Sub(startingAt))
			return
		} else {
			counter++
		}
	}

	config.Checkpoint = (blockNumbers)[len(blockNumbers)-1]
	logrus.Info("Checkpoint: ", config.Checkpoint)
	logrus.Info("Validation took: ", time.Now().UTC().Sub(startingAt))
}
