package wire

import (
	"math/big"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type OpenChannel struct {
	PendingChannelID [32]byte
	FundingAmount    *big.Int
	CsvDelay         uint16
	MaxAcceptedHTLCs uint16
	FundingKey       *crypto.PublicKey
}

func (msg *OpenChannel) MsgType() lnwire.MessageType {
	return MsgOpenChannel
}

func (msg *OpenChannel) MaxPayloadLength(uint32) uint32 {
	return 129
}

func (msg *OpenChannel) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.PendingChannelID,
		&msg.FundingAmount,
		&msg.CsvDelay,
		&msg.MaxAcceptedHTLCs,
		&msg.FundingKey,
	)
}

func (msg *OpenChannel) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.PendingChannelID,
		msg.FundingAmount,
		msg.CsvDelay,
		msg.MaxAcceptedHTLCs,
		msg.FundingKey,
	)
}