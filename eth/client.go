package eth

import (
	"github.com/ethereum/go-ethereum/rpc"
	logrus "github.com/sirupsen/logrus"
)

// Connect to blockchain node, either using HTTP or Websocket connection
// depending upon true/ false, passed to function, respectively
func GetClient(rpcUrl string) *rpc.Client {

	rpcClient, err := rpc.Dial(rpcUrl)
	if err != nil {
		logrus.Panic("Cannot connect to blockchain node, err: ", err)
	}

	return rpcClient
}
