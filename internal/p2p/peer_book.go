package p2p

import (
	"net"
	"sync"
)

type AddrSet = map[net.Addr]bool

type PeerBook struct {
	all          AddrSet
	connected    AddrSet
	disconnected AddrSet
	mut          *sync.Mutex
}

func NewPeerBook(initialPeers []net.Addr) (*PeerBook) {
	all := make(map[net.Addr]bool)
	disconnected := make(map[net.Addr]bool)

	for _, p := range initialPeers {
		all[p] = true
		disconnected[p] = true
	}

	return &PeerBook{
		all:          all,
		connected:    make(map[net.Addr]bool),
		disconnected: disconnected,
		mut:          &sync.Mutex{},
	}
}

func (p *PeerBook) PushConnectedPeer(addr net.Addr) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.connected[addr] = true
}

func (p *PeerBook) DisconnectPeer(addr net.Addr) {
	p.mut.Lock()
	defer p.mut.Unlock()
	delete(p.connected, addr)
}

func (p *PeerBook) DisconnectedCount() (int) {
	p.mut.Lock()
	defer p.mut.Unlock()
	return len(p.disconnected)
}

func (p *PeerBook) PopDisconnectedPeer() (net.Addr) {
	p.mut.Lock()
	defer p.mut.Unlock()

	for k := range p.disconnected {
		delete(p.disconnected, k)
		delete(p.all, k)
		return k
	}

	return nil
}
