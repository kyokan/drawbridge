package crypto

import (
	"github.com/ethereum/go-ethereum/crypto"
)

type Signature []byte

func (s Signature) Bytes() ([]byte) {
	return s
}

func (s Signature) Wire() ([64]byte) {
	var out [64]byte
	copy(out[:], s[:64])
	return out
}

func VerifySignature(data [32]byte, expectedPub *PublicKey, sig Signature) bool {
	actualPub, err := crypto.SigToPub(data[:], sig.Bytes())

	if err != nil {
		return false
	}

	pub, err := PublicFromOtherPublic(actualPub)

	if err != nil {
		return false
	}

	return pub.Equal(expectedPub)
}