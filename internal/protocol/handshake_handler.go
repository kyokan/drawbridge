package protocol

import (
	"github.com/kyokan/drawbridge/internal/lndclient"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/pkg/wire"
	"github.com/go-errors/errors"
	"github.com/kyokan/drawbridge/internal/p2p"
)

type HandshakeHandler struct {
	lndClient        *lndclient.Client
}

func NewHandshakeHandler(lndClient *lndclient.Client) *HandshakeHandler {
	return &HandshakeHandler{
		lndClient: lndClient,
	}
}

func (h *HandshakeHandler) CanAccept(msg lnwire.Message) bool {
	switch msg.MsgType() {
	case wire.MsgInit:
		return true
	default:
		return false
	}
}

func (h *HandshakeHandler) Accept(envelope *p2p.Envelope) (lnwire.Message, error) {
	msg := envelope.Msg
	switch msg.MsgType() {
	case wire.MsgInit:
		return h.acceptInit(msg.(*wire.Init), envelope.Peer)
	default:
		return nil, errors.New("unknown message type")
	}
}

func (h *HandshakeHandler) acceptInit(msg *wire.Init, peer *p2p.Peer) (lnwire.Message, error) {
	log.Infow(
		"connecting lnd peers",
		"pubkey",
		msg.LNDIdentificationKey.CompressedHex(),
		"addr",
		msg.LNDHost,
	)
	err := h.lndClient.ConnectPeer(msg.LNDIdentificationKey, msg.LNDHost)
	if err != nil {
		return nil, err
	}
	peer.LNDIdentity = msg.LNDIdentificationKey
	return nil, nil
}
