package crypto

import (
	"testing"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/common"
)

func TestPublicKey_ETHAddress(t *testing.T) {
	priv, err := crypto.HexToECDSA("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3")
	assert.Nil(t, err)
	pub, err := PublicFromOtherPublic(priv.Public())
	assert.Nil(t, err)
	addr := common.HexToAddress("0x627306090abab3a6e1400e9345bc60c78a8bef57")
	assert.Equal(t, addr, pub.ETHAddress())
}

func TestPublicKey_Equal(t *testing.T) {
	priv1, err := crypto.HexToECDSA("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3")
	assert.Nil(t, err)
	priv2, err := crypto.HexToECDSA("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3")
	assert.Nil(t, err)
	priv3, err := crypto.HexToECDSA("ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f")
	assert.Nil(t, err)

	pub1, err := PublicFromOtherPublic(priv1.Public())
	assert.Nil(t, err)
	pub2, err := PublicFromOtherPublic(priv2.Public())
	assert.Nil(t, err)
	pub3, err := PublicFromOtherPublic(priv3.Public())
	assert.Nil(t, err)

	assert.True(t, pub1.Equal(pub2))
	assert.False(t, pub1.Equal(pub3))
}