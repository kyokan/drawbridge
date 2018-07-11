package wire

import (
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type AcceptChannel struct {
	PendingChannelID [32]byte
	CsvDelay         uint16
	MaxAcceptedHTLCs uint16
	FundingKey       *crypto.PublicKey
}

func (msg *AcceptChannel) MsgType() lnwire.MessageType {
	return MsgAcceptChannel
}

func (msg *AcceptChannel) MaxPayloadLength(uint32) uint32 {
	return 129
}

func (msg *AcceptChannel) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.PendingChannelID,
		&msg.CsvDelay,
		&msg.MaxAcceptedHTLCs,
		&msg.FundingKey,
	)
}

func (msg *AcceptChannel) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.PendingChannelID,
		msg.CsvDelay,
		msg.MaxAcceptedHTLCs,
		msg.FundingKey,
	)
}