package p2p

import "github.com/lightningnetwork/lnd/lnwire"

type MsgHandler interface {
	CanAccept(msg lnwire.Message) bool
	Accept(msg lnwire.Message) (lnwire.Message, error)
}