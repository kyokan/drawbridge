package wire

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type FundingLocked struct {
	ChannelID common.Hash
}

func (msg *FundingLocked) MsgType() lnwire.MessageType {
	return MsgFundingLocked
}

func (msg *FundingLocked) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *FundingLocked) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.ChannelID,
	)
}

func (msg *FundingLocked) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.ChannelID,
	)
}
