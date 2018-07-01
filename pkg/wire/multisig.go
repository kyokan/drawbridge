package wire

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"crypto/sha256"
)

type Multisig struct {
	InputA [32]byte
	InputB [32]byte
	SignerA common.Address
	SignerB common.Address
}

func (m *Multisig) Hashable() ([]byte) {
	buf := new(bytes.Buffer)
	buf.Write(m.InputA[:])
	buf.Write(m.InputB[:])
	buf.Write(m.SignerA[:])
	buf.Write(m.SignerB[:])
	return buf.Bytes()
}

func (m *Multisig) Hash() ([32]byte) {
	hashable := m.Hashable()
	return sha256.Sum256(hashable)
}