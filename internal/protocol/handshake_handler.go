package protocol

import (
	"github.com/kyokan/drawbridge/internal/lndclient"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/pkg/wire"
	"github.com/go-errors/errors"
)

type HandshakeHandler struct {
	lndClient        *lndclient.LNDClient
}

func NewHandshakeHandler(lndClient *lndclient.LNDClient) *HandshakeHandler {
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

func (h *HandshakeHandler) Accept(msg lnwire.Message) (lnwire.Message, error) {
	switch msg.MsgType() {
	case wire.MsgInit:
		return h.acceptInit(msg.(*wire.Init))
	default:
		return nil, errors.New("unknown message type")
	}
}

func (h *HandshakeHandler) acceptInit(msg *wire.Init) (lnwire.Message, error) {
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

	return nil, nil
}
