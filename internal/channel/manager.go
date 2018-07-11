package channel

import (
	"github.com/kyokan/drawbridge/internal/p2p"
	"github.com/roasbeef/btcd/btcec"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/kyokan/drawbridge/internal/wallet"
)

type Manager struct {
	node *p2p.Node
	peerBook *p2p.PeerBook
}

func NewManager(node *p2p.Node) *Manager {
	return &Manager{
		node: node,
	}
}

func (m *Manager) OpenChannel(pub *btcec.PublicKey, amount *big.Int) error {
	peer := m.peerBook.FindPeer(pub)

	if peer == nil {
		return errors.New("peer not found")
	}

	_, err := wallet.GenCommitmentRoot()

	if err != nil {
		return err
	}

	return nil
}