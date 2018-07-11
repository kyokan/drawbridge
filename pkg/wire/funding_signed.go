package wire

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
)

type FundingSigned struct {
	ChannelID common.Hash
	Sig       crypto.Signature
}

func (msg *FundingSigned) MsgType() lnwire.MessageType {
	return MsgFundingSigned
}

func (msg *FundingSigned) MaxPayloadLength(uint32) uint32 {
	return 65535
}

func (msg *FundingSigned) Decode(r io.Reader, pver uint32) error {
	return readElements(
		r,
		&msg.ChannelID,
		&msg.Sig,
	)
}

func (msg *FundingSigned) Encode(w io.Writer, pver uint32) error {
	return writeElements(
		w,
		msg.ChannelID,
		msg.Sig,
	)
}
