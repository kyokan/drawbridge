package pkg

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type Config struct {
	ChainHashes    *ChainHashes
	SigningPubkey  *crypto.PublicKey
	P2PAddr        string
	P2PPort        string
	BootstrapPeers []string
}

type ChainHashes struct {
	TestToken chainhash.Hash
}

func NewChainHashes() (*ChainHashes, error) {
	testTokenHash, err := chainhash.NewHashFromStr("438a269b9ef6d3204e0056bc58c7afcaf4fd3524fd6da063fe6e5408dc696f73")

	if err != nil {
		return nil, err
	}

	return &ChainHashes{
		TestToken: *testTokenHash,
	}, nil
}

func (h *ChainHashes) ValidChainHash(hash chainhash.Hash) bool {
	switch hash {
	case h.TestToken:
		return true
	default:
		return false
	}
}
