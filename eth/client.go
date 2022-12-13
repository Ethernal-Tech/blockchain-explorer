package eth

import (
	"github.com/ethereum/go-ethereum/rpc"
	logrus "github.com/sirupsen/logrus"
)

type BlockchainNodeConnection struct {
	HTTP      *rpc.Client
	WebSocket *rpc.Client
}

// Connect to blockchain node, either using HTTP or Websocket connection depending on URL passed to function
func GetClient(rpcUrl string) *rpc.Client {

	rpcClient, err := rpc.Dial(rpcUrl)
	if err != nil {
		logrus.Panic("Cannot connect to blockchain node, err: ", err)
	}

	return rpcClient
}
