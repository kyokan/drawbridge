package db

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ETHOutput struct {
	ID              *big.Int
	ContractAddress common.Address
	Amount          *big.Int
	BlockNumber     uint64
	TxHash          common.Hash
	Script          []byte
	Type            uint8
	IsSpent         bool
	IsWithdrawn     bool
}