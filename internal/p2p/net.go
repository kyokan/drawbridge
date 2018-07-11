package p2p

import (
	"net"
	"github.com/lightningnetwork/lnd/lnwire"
	"strings"
	"errors"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

func ResolveAddrs(addrs []string) ([]*lnwire.NetAddress, error) {
	var out []*lnwire.NetAddress

	for _, a := range addrs {
		splits := strings.Split(a, "|")

		if len(splits) != 2 {
			return nil, errors.New("invalid peer: " + a)
		}

		host := splits[0]
		pub := splits[1]

		resolved, err := net.ResolveTCPAddr("tcp", host)

		if err != nil {
			return nil, err
		}

		identityKey, err := crypto.PublicFromCompressedHex(pub)

		if err != nil {
			return nil, err
		}

		addr := &lnwire.NetAddress{
			IdentityKey: identityKey.BTCEC(),
			Address: resolved,
		}

		out = append(out, addr)
	}

	return out, nil
}