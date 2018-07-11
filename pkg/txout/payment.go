package txout

import (
	"io"
	"github.com/ethereum/go-ethereum/common"
)

type PaymentSigType uint8

type Payment struct {
	Recipient common.Address
}

func NewPayment(recipient common.Address) *Payment {
	return &Payment{
		Recipient: recipient,
	}
}

func (p *Payment) Decode(r io.Reader, pver uint32) error {
	var addr common.Address
	_, err := io.ReadFull(r, addr[:])
	if err != nil {
		return err
	}

	p.Recipient = addr
	return nil
}

func (p *Payment) Encode(w io.Writer, pver uint32) error {
	var b [1]byte
	b[0] = byte(OutputPayment)
	_, err := w.Write(b[:])
	if err != nil {
		return nil
	}

	_, err = w.Write(p.Recipient.Bytes())
	return err
}

func (p *Payment) OutputType() OutputType {
	return OutputPayment
}

type PaymentWitness struct {
}

func NewPaymentWitness() *PaymentWitness {
	return &PaymentWitness{}
}

func (p *PaymentWitness) Encode(w io.Writer) error {
	var b [1]byte
	b[0] = 0
	_, err := w.Write(b[:])
	return err
}
