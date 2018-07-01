package wire

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"github.com/ethereum/go-ethereum/common/math"
)

type Encumbrance struct {
	InputId  [32]byte
	LockTime *big.Int
	ValueA   *big.Int
	ValueB   *big.Int
	HashLock [32]byte
	SigA     []byte
	SigB     []byte
}

func (e *Encumbrance) Hashable() ([]byte) {
	buf := new(bytes.Buffer)
	buf.Write(e.InputId[:])
	buf.Write(math.PaddedBigBytes(e.LockTime, 32))
	buf.Write(math.PaddedBigBytes(e.ValueA, 32))
	buf.Write(math.PaddedBigBytes(e.ValueB, 32))
	buf.Write(e.HashLock[:])
	return buf.Bytes()
}

func (e *Encumbrance) Hash() ([32]byte) {
	hashable := e.Hashable()
	return sha256.Sum256(hashable)
}
