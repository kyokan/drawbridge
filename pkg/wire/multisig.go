package wire

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/drawbridge/internal/eth"
	"github.com/kyokan/drawbridge/pkg/types"
)

type Multisig struct {
	InputA  [32]byte
	InputB  [32]byte
	SignerA common.Address
	SignerB common.Address
}

func NewMultisig(inputId [32]byte, ourAddress common.Address, theirAddress common.Address) (*Multisig) {
	a, _ := eth.SortAddresses(ourAddress, theirAddress)

	if bytes.Equal(ourAddress[:], a[:]) {
		return &Multisig{
			InputA:  inputId,
			InputB:  types.ZeroUTXOID,
			SignerA: ourAddress,
			SignerB: theirAddress,
		}
	}

	return &Multisig{
		InputA:  types.ZeroUTXOID,
		InputB:  inputId,
		SignerA: theirAddress,
		SignerB: ourAddress,
	}
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
	return crypto.Keccak256Hash(hashable)
}

func (m *Multisig) ID() [32]byte {
	buf := new(bytes.Buffer)
	buf.Write(m.InputA[:])
	buf.Write(m.InputB[:])
	buf.Write(m.SignerA[:])
	buf.Write(m.SignerB[:])
	return crypto.Keccak256Hash(buf.Bytes())
}
