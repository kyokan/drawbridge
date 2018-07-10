package p2p

import "github.com/lightningnetwork/lnd/lnwire"

type Ping struct {}

func (*Ping) HandlePing(ping *lnwire.Ping) (*lnwire.Pong, error) {
	return lnwire.NewPong(ping.PaddingBytes), nil
}