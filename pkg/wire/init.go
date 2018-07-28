package wire

import (
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type Init struct {
	LNDIdentificationKey *crypto.PublicKey
	LNDHost           string
}

func NewInit(ident *crypto.PublicKey, host string) (*Init) {
	return &Init {
		LNDIdentificationKey: ident,
		LNDHost: host,
	}
}

func (msg *Init) MsgType() lnwire.MessageType {
	return MsgInit
}

func (msg *Init) MaxPayloadLength(uint32) uint32 {
	return 40
}

func (msg *Init) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.LNDIdentificationKey,
		&msg.LNDHost,
	)
}

func (msg *Init) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.LNDIdentificationKey,
		msg.LNDHost,
	)
}
