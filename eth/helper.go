package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Erc721Transfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
}

type Erc1155TransferBatch struct {
	Operator common.Address
	From     common.Address
	To       common.Address
	Ids      []*big.Int
	Values   []*big.Int
}

type Erc1155TransferSingle struct {
	Operator common.Address
	From     common.Address
	To       common.Address
	Id       *big.Int
	Value    *big.Int
}
