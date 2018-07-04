package msghandler

import "github.com/lightningnetwork/lnd/lnwire"

func HandlePing(ping *lnwire.Ping) *lnwire.Pong {
	return lnwire.NewPong(ping.PaddingBytes)
}