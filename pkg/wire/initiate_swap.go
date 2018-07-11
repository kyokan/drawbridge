package wire

import (
		"github.com/lightningnetwork/lnd/lnwire"
	"io"
		"math/big"
	"github.com/kyokan/drawbridge/pkg/crypto"
	)

type InitiateSwap struct {
	SwapID [32]byte
	PaymentHash [32]byte
	ETHChannelID [32]byte
	ETHAmount *big.Int
	ETHCommitmentSignature crypto.Signature
	SendingAddress *crypto.PublicKey
	RequestedAmount *big.Int
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
		&msg.SwapID,
		&msg.PaymentHash,
		&msg.ETHChannelID,
		&msg.ETHAmount,
		&msg.ETHCommitmentSignature,
		&msg.SendingAddress,
		&msg.RequestedAmount,
	)
}

func (msg *InitiateSwap) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.SwapID,
		msg.PaymentHash,
		msg.ETHChannelID,
		msg.ETHAmount,
		msg.ETHCommitmentSignature,
		msg.SendingAddress,
		msg.RequestedAmount,
	)
}