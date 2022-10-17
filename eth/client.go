package eth

import (
	"log"

	"github.com/ethereum/go-ethereum/rpc"
)

// Connect to blockchain node, either using HTTP or Websocket connection
// depending upon true/ false, passed to function, respectively
func GetClient(rpcUrl string) *rpc.Client {

	rpcClient, err := rpc.Dial(rpcUrl)
	if err != nil {
		log.Println(err)
	}

	return rpcClient
}
