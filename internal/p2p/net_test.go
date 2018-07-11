package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"bytes"
)

func TestResolveAddrs(t *testing.T) {
	keyA, err := hexutil.Decode("0x02ce7edc292d7b747fab2f23584bbafaffde5c8ff17cf689969614441e0527b900")
	keyB, err := hexutil.Decode("0x02785a891f323acd6cef0fc509bb14304410595914267c50467e51c87142acbb5e")

	inputAddrs := []string {
		"127.0.0.1:8080|0x02ce7edc292d7b747fab2f23584bbafaffde5c8ff17cf689969614441e0527b900",
		"8.8.8.8:8080|0x02785a891f323acd6cef0fc509bb14304410595914267c50467e51c87142acbb5e",
	}

	resolved, err := ResolveAddrs(inputAddrs)
	assert.Nil(t, err)

	assert.True(t, bytes.Equal(resolved[0].IdentityKey.SerializeCompressed(), keyA))
	assert.True(t, bytes.Equal(resolved[1].IdentityKey.SerializeCompressed(), keyB))
}