package p2p

import (
	"github.com/roasbeef/btcd/btcec"
	"github.com/kyokan/drawbridge/internal/conv"
	"sync"
)

type PeerBook struct {
	peerIndices map[string]uint16
	peers       map[uint16]*Peer
	mut         *sync.Mutex
	lastIdx     uint16
}

func NewPeerBook() *PeerBook {
	return &PeerBook{
		peerIndices: make(map[string]uint16),
		peers:       make(map[uint16]*Peer),
		mut:         &sync.Mutex{},
		lastIdx:     0,
	}
}

func (p *PeerBook) FindPeer(pub *btcec.PublicKey) (*Peer) {
	p.mut.Lock()
	defer p.mut.Unlock()

	peerIdx := p.peerIndices[conv.PubKeyToHex(pub)]

	if peerIdx == 0 {
		return nil
	}

	return p.peers[peerIdx]
}

func (p *PeerBook) AddPeer(peer *Peer) bool {
	p.mut.Lock()
	defer p.mut.Unlock()

	keyStr := conv.PubKeyToHex(peer.Identity)

	if p.peerIndices[keyStr] != 0 {
		return false
	}

	p.lastIdx++
	p.peerIndices[keyStr] = p.lastIdx
	p.peers[p.lastIdx] = peer
	return true
}

func (p *PeerBook) RemovePeer(pub *btcec.PublicKey) bool {
	p.mut.Lock()
	defer p.mut.Unlock()

	keyStr := conv.PubKeyToHex(pub)

	if p.peerIndices[keyStr] == 0 {
		return false
	}

	peerIdx := p.peerIndices[keyStr]
	delete(p.peerIndices, keyStr)
	delete(p.peers, peerIdx)
	return true
}
