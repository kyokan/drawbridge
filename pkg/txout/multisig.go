package txout

import (
	"github.com/ethereum/go-ethereum/common"
	"io"
	"github.com/kyokan/drawbridge/pkg/eth"
)

type Multisig struct {
	Alice common.Address
	Bob   common.Address
}

func NewMultisig(alice common.Address, bob common.Address) *Multisig {
	a, b := eth.SortAddresses(alice, bob)

	return &Multisig{
		Alice: a,
		Bob:   b,
	}
}

func (m *Multisig) OutputType() OutputType {
	return OutputMultisig
}

func (m *Multisig) Decode(r io.Reader, pver uint32) error {
	var alice common.Address
	var bob common.Address
	_, err := io.ReadFull(r, alice[:])
	if err != nil {
		return err
	}
	_, err = io.ReadFull(r, bob[:])
	if err != nil {
		return err
	}

	m.Alice = alice
	m.Bob = bob
	return nil
}

func (m *Multisig) Encode(w io.Writer, pver uint32) error {
	var b [1]byte
	b[0] = byte(OutputMultisig)
	_, err := w.Write(b[:])
	if err != nil {
		return err
	}
	_, err = w.Write(m.Alice.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(m.Bob.Bytes())
	return err
}

type MultisigWitness struct {
}

func NewMultisigWitness() *MultisigWitness {
	return &MultisigWitness{}
}

func (m *MultisigWitness) Encode(w io.Writer) error {
	var b [1]byte
	if _, err := w.Write(b[:]); err != nil {
		return err
	}
	return nil
}
