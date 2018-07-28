package wire

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
	"math/big"
)

type FundingCreated struct {
	PendingChannelID [32]byte
	InputID *big.Int
	Sig []byte
}

func (msg *FundingCreated) MsgType() lnwire.MessageType {
	return MsgFundingCreated
}

func (msg *FundingCreated) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *FundingCreated) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.PendingChannelID,
		&msg.InputID,
		&msg.Sig,
	)
}

func (msg *FundingCreated) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.PendingChannelID,
		msg.InputID,
		msg.Sig,
	)
}