package wallet

import (
	"github.com/lightningnetwork/lnd/shachain"
	"github.com/roasbeef/btcd/btcec"
	"github.com/lightningnetwork/lnd/lnwallet"
	"crypto/rand"
)

func GenCommitmentRoot() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func CommitmentAtIndex(root []byte, index uint64) (*btcec.PublicKey, error) {
	producer, err := shachain.NewRevocationProducerFromBytes(root)

	if err != nil {
		return nil, err
	}

	preimage, err := producer.AtIndex(index)

	if err != nil {
		return nil, err
	}

	return lnwallet.ComputeCommitmentPoint(preimage[:]), nil
}
