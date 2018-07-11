package p2p

import (
	"net"
	"github.com/lightningnetwork/lnd/lnwire"
	"strings"
	"errors"
	"github.com/roasbeef/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func ResolveAddrs(addrs []string) ([]*lnwire.NetAddress, error) {
	out := make([]*lnwire.NetAddress, len(addrs))

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

		keyBytes, err := hexutil.Decode(pub)

		if err != nil {
			return nil, err
		}

		identityKey, err := btcec.ParsePubKey(keyBytes, btcec.S256())

		if err != nil {
			return nil, err
		}

		addr := &lnwire.NetAddress{
			IdentityKey: identityKey,
			Address: resolved,
		}

		out = append(out, addr)
	}

	return out, nil
}