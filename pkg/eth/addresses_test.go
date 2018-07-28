package eth

import (
	"testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestSortAddresses(t *testing.T) {
	a := common.HexToAddress("0x414636aea6d47594007c1064dc91a1af4fc3b6f3")
	b := common.HexToAddress("0xfbb5fbd8c5046705fc41e378e7559dfa7e13202b")

	sortA, sortB := SortAddresses(a, b)
	assert.Equal(t, sortA, a)
	assert.Equal(t, sortB, b)

	sortA, sortB = SortAddresses(b, a)
	assert.Equal(t, sortA, a)
	assert.Equal(t, sortB, b)
}