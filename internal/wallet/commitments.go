package wallet

import (
	"github.com/lightningnetwork/lnd/shachain"
	"github.com/btcsuite/btcd/btcec"
	"github.com/lightningnetwork/lnd/lnwallet"
	"crypto/rand"
)

func Rand32() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func Rand32Array() ([32]byte, error) {
	var out [32]byte
	res, err := Rand32()

	if err != nil {
		return out, err
	}

	copy(out[:], res)
	return out, nil
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

func FirstCommitmentPoint() ([]byte,  *btcec.PublicKey, error) {
	seed, err := Rand32()

	if err != nil {
		return nil, nil, err
	}

	commitment, err := CommitmentAtIndex(seed, 0)

	if err != nil {
		return nil, nil, err
	}

	return seed, commitment, nil
}