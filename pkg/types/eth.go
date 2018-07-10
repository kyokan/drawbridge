package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var ZeroUTXOID = *new([32]byte)

type EthUTXO struct {
	Owner       common.Address
	Value       *big.Int
	BlockNumber *big.Int
	TxHash      common.Hash
	ID          [32]byte
	InputID     [32]byte
	IsWithdrawn bool
	IsSpent     bool
}
