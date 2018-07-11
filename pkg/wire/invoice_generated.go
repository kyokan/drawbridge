package wire

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type InvoiceGenerated struct {
	SwapID         [32]byte
	PaymentRequest string
}

func (msg *InvoiceGenerated) MsgType() lnwire.MessageType {
	return MsgInvoiceGenerated
}

func (msg *InvoiceGenerated) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *InvoiceGenerated) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.SwapID,
		&msg.PaymentRequest,
	)
}

func (msg *InvoiceGenerated) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.SwapID,
		msg.PaymentRequest,
	)
}
