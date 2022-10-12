package eth

import (
	"ethernal/explorer/config"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Connect to blockchain node, either using HTTP or Websocket connection
// depending upon true/ false, passed to function, respectively
func GetClient() *ethclient.Client {
	log.Println(config.Get("RPCUrl"))
	client, err := ethclient.Dial(config.Get("RPCUrl"))

	if err != nil {
		log.Fatalf("[!] Failed to connect to blockchain : %s\n", err.Error())
	}

	return client
}
