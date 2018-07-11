package ethclient

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
)

func SortAddresses(a common.Address, b common.Address) (common.Address, common.Address) {
	cmp := bytes.Compare(a[:], b[:])

	if cmp == 1 {
		return b, a
	}

	if cmp == -1 {
		return a, b
	}

	return a, b
}