package p2p

import "github.com/lightningnetwork/lnd/lnwire"

type Envelope struct {
	Peer *Peer
	Msg  lnwire.Message
}

func NewEnvelope(peer *Peer, msg lnwire.Message) (*Envelope) {
	return &Envelope{
		Peer: peer,
		Msg:  msg,
	}
}
