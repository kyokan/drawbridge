package p2p

import (
	"sync"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type PeerBook struct {
	peerIndices map[string]uint16
	peers       map[uint16]*Peer
	mut         sync.Mutex
	lastIdx     uint16
}

func NewPeerBook() *PeerBook {
	return &PeerBook{
		peerIndices: make(map[string]uint16),
		peers:       make(map[uint16]*Peer),
		lastIdx:     0,
	}
}

func (p *PeerBook) FindPeer(pub *crypto.PublicKey) (*Peer) {
	p.mut.Lock()
	defer p.mut.Unlock()

	peerIdx, exists := p.peerIndices[pub.CompressedHex()]

	if !exists {
		return nil
	}

	return p.peers[peerIdx]
}

func (p *PeerBook) AddPeer(peer *Peer) bool {
	p.mut.Lock()
	defer p.mut.Unlock()

	keyStr := peer.Identity.CompressedHex()

	if _, exists := p.peerIndices[keyStr]; exists {
		return false
	}

	p.lastIdx++
	p.peerIndices[keyStr] = p.lastIdx
	p.peers[p.lastIdx] = peer
	return true
}

func (p *PeerBook) RemovePeer(pub *crypto.PublicKey) bool {
	p.mut.Lock()
	defer p.mut.Unlock()

	keyStr := pub.CompressedHex()

	if _, exists := p.peerIndices[keyStr]; !exists {
		return false
	}

	peerIdx := p.peerIndices[keyStr]
	delete(p.peerIndices, keyStr)
	delete(p.peers, peerIdx)
	return true
}
