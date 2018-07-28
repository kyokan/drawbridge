package wire

import (
	"math/big"
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type InitiateSwap struct {
	ReceivingChainID uint16
	ReceivingChainAmount *big.Int
	SendingChainID uint16
	SendingChainAmount *big.Int
	PaymentHash [32]byte
}

func (msg *InitiateSwap) MsgType() lnwire.MessageType {
	return MsgInitiateSwap
}

func (msg *InitiateSwap) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *InitiateSwap) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.ReceivingChainID,
		&msg.ReceivingChainAmount,
		&msg.SendingChainID,
		&msg.SendingChainAmount,
		&msg.PaymentHash,
	)
}

func (msg *InitiateSwap) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.ReceivingChainID,
		msg.ReceivingChainAmount,
		msg.SendingChainID,
		msg.SendingChainAmount,
		msg.PaymentHash,
	)
}