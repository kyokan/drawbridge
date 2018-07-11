package crypto

import (
	"github.com/ethereum/go-ethereum/crypto"
	"fmt"
)

const SignaturePreamble = "\x19Ethereum Signed Message:\n"


type Signature []byte

func (s Signature) Bytes() ([]byte) {
	return s
}

func (s Signature) Wire() ([64]byte) {
	var out [64]byte
	copy(out[:], s[:64])
	return out
}

func (s Signature) Verify(data []byte, expectedPub *PublicKey) bool {
	hash := GethHash(data)
	actualPub, err := crypto.SigToPub(hash, s)
	if err != nil {
		return false
	}

	pub, err := PublicFromOtherPublic(actualPub)
	if err != nil {
		return false
	}

	return pub.Equal(expectedPub)
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

func GethHash(data []byte) []byte {
	msg := fmt.Sprintf("%s%d%s", SignaturePreamble, len(data), data)
	return crypto.Keccak256([]byte(msg))
}