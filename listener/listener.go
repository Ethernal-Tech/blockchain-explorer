package listener

import (
	"context"
	"ethernal/explorer/config"
	"ethernal/explorer/eth"
	"ethernal/explorer/syncer"
	"ethernal/explorer/utils"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rpc"

	bundb "github.com/uptrace/bun"
)

type BlockHeader struct {
	Number string
}

// ListenForNewBlocks listens for new blocks on the blockchain and then processes them.
func ListenForNewBlocks(connection *eth.BlockchainNodeConnection, db *bundb.DB, config config.Config) {

	// synch signal ensures that only one trigger can perform synchronization at a time
	synch := syncer.GetSignalSynchInstance()
	// adding signal to channel ensures that the first trigger performs synchronization
	synch.Done <- struct{}{}
	// channel for new blocks
	blocks := make(chan BlockHeader)

	// subscription to newHeads event run in a goroutine
	go func() {
		// loop to re-establish the subscription if it has ended
		for i := 0; ; i++ {
			if i > 0 {
				time.Sleep(2 * time.Second)
			}
			subscribeBlocks(connection.WebSocket, blocks, config.CallTimeoutInSeconds)
		}
	}()

	// listen on channel for new blocks
	for block := range blocks {
		log.Println("New block:", utils.ToUint64(block.Number))
		// check if the trigger can start sync or it will be ignored
		select {
		// if channel Done contains sync signal, start sync
		case <-synch.Done:
			go syncer.SyncMissingBlocks(connection.HTTP, db, config)
		// ignore synch
		default:
		}
	}
}

// SubscribeBlocks maintains a subscription for new blocks.
func subscribeBlocks(client *rpc.Client, blocks chan BlockHeader, timeout uint) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// subscribe to newHeads event
	subscription, err := client.EthSubscribe(ctx, blocks, "newHeads")
	if err != nil {
		log.Println("Error subscribing to newHeads event:", err)
		return
	}

	// The subscription will deliver events to the channel. Wait for the
	// subscription to end for any reason, then loop around to re-establish
	// the connection.
	log.Println("Connection with subscription to newHeads event lost: ", <-subscription.Err())
}
