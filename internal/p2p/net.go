package p2p

import "net"

func ResolveTCPAddrs(addrs []string) ([]net.Addr, error) {
	out := make([]net.Addr, len(addrs))

	for _, a := range addrs {
		resolved, err := net.ResolveTCPAddr("tcp", a)

		if err != nil {
			return nil, err
		}

		out = append(out, resolved)
	}

	return out, nil
}