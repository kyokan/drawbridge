package wire

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type SwapAccepted struct {
	SwapID       [32]byte
	BTCChannelID uint64
}

func (msg *SwapAccepted) MsgType() lnwire.MessageType {
	return MsgSwapAccepted
}

func (msg *SwapAccepted) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *SwapAccepted) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.SwapID,
		&msg.BTCChannelID,
	)
}

func (msg *SwapAccepted) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.SwapID,
		msg.BTCChannelID,
	)
}
