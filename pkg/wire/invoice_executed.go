package wire

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type InvoiceExecuted struct {
	SwapID [32]byte
}

func (msg *InvoiceExecuted) MsgType() lnwire.MessageType {
	return MsgInvoiceExecuted
}

func (msg *InvoiceExecuted) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *InvoiceExecuted) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.SwapID,
	)
}

func (msg *InvoiceExecuted) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.SwapID,
	)
}
 