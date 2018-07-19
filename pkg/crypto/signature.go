package crypto

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signature []byte

func SignatureFromWire(wire [64]byte) Signature {
	var b = make([]byte, 65)
	copy(b, wire[:])
	b[64] = 0
	return b
}

func SignatureFromHex(hex string) (Signature, error) {
	b, err := hexutil.Decode(hex)

	if err != nil {
		return nil, err
	}

	if len(b) != 65 {
		return nil, errors.New("length of signature must be 65")
	}

	return b, nil
}

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