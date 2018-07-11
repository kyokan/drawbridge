package protocol

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/go-errors/errors"
	"github.com/kyokan/drawbridge/internal/p2p"
)

type PingPongHandler struct {}

func (*PingPongHandler) CanAccept(msg lnwire.Message) bool {
	return msg.MsgType() == lnwire.MsgPing || msg.MsgType() == lnwire.MsgPong
}

func (*PingPongHandler) Accept(envelope *p2p.Envelope) (lnwire.Message, error) {
	msg := envelope.Msg
	switch msg.MsgType() {
	case lnwire.MsgPing:
		return lnwire.NewPong(msg.(*lnwire.Ping).PaddingBytes), nil
	case lnwire.MsgPong:
		return nil, nil
	default:
		return nil, errors.New("invalid message type")
	}
}