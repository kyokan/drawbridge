package conv

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"math/big"
)

func TestStringToBig(t *testing.T) {
	n1, err := StringToBig("100")
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100).Cmp(n1), 0)

	n2, err := StringToBig("nope")
	assert.NotNil(t, err)
	assert.Nil(t, n2)
}