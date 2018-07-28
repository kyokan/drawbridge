package p2p

import (
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"net"
	"github.com/btcsuite/btcd/connmgr"
	"time"
	"sync"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/lightningnetwork/lnd/brontide"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

var nLog *zap.SugaredLogger

func init() {
	nLog = logger.Logger.Named("node")
}

type Node struct {
	reactor        *Reactor
	connMgr        *connmgr.ConnManager
	peerBook       *PeerBook
	bootstrapPeers []string
	addr           string
	port           string
	lndIdentity    *crypto.PublicKey
	lndHost        string
}

type NodeConfig struct {
	Reactor        *Reactor
	PeerBook       *PeerBook
	P2PAddr        string
	P2PPort        string
	BootstrapPeers []string
	LNDIdentity    *crypto.PublicKey
	LNDHost        string
}

func NewNode(config *NodeConfig) (*Node, error) {
	return &Node{
		reactor:        config.Reactor,
		peerBook:       config.PeerBook,
		bootstrapPeers: config.BootstrapPeers,
		addr:           config.P2PAddr,
		port:           config.P2PPort,
		lndIdentity:    config.LNDIdentity,
		lndHost:        config.LNDHost,
	}, nil
}

func (n *Node) Start(identityKey *btcec.PrivateKey) error {
	nLog.Infow("starting p2p node",
		"p2pIp", n.addr,
		"p2pPort", n.port,
		"identityKey", hexutil.Encode(identityKey.PubKey().SerializeCompressed()),
	)

	listenAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(n.addr, n.port))

	if err != nil {
		nLog.Panicw("failed to parse TCP address", "err", err)
	}

	listener, err := brontide.NewListener(identityKey, listenAddr.String())

	if err != nil {
		nLog.Panicw("failed to listen to TCP address", "err", err, "addr", listenAddr.String())
	}

	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners: []net.Listener{
			listener,
		},
		OnAccept:       n.onAccept,
		RetryDuration:  time.Second * 5,
		TargetOutbound: 100,
		Dial: func(a net.Addr) (net.Conn, error) {
			if a == nil || a == (*lnwire.NetAddress)(nil) {
				return nil, errors.New("addr is nil")
			}

			return brontide.Dial(identityKey, a.(*lnwire.NetAddress), func(network string, address string) (net.Conn, error) {
				return net.Dial(network, address)
			})
		},
		OnConnection:    n.onConnection,
		OnDisconnection: n.onDisconnection,
	})

	if err != nil {
		nLog.Panicw("failed to start p2p node", "err", err)
	}

	n.connMgr = cmgr

	cmgr.Start()

	if len(n.bootstrapPeers) > 0 {
		addrs, err := ResolveAddrs(n.bootstrapPeers)

		if err != nil {
			nLog.Errorw("failed to resolve bootstrap peers", "err", err, "bootstrapPeers", n.bootstrapPeers)
		}

		go n.bootstrap(addrs)
	}

	return nil
}

func (n *Node) FindPeer(pub *crypto.PublicKey) *Peer {
	return n.peerBook.FindPeer(pub)
}

func (n *Node) SendPeer(pub *crypto.PublicKey, msg lnwire.Message) error {
	peer := n.peerBook.FindPeer(pub)

	if peer == nil {
		return errors.New("no peer with id " + pub.CompressedHex() + " found")
	}

	return peer.Send(msg)
}

func (n *Node) onConnection(req *connmgr.ConnReq, conn net.Conn) {
	noiseConn := conn.(*brontide.Conn)
	peer, err := NewPeer(n.reactor, noiseConn, true)

	if err != nil {
		nLog.Errorw("failed to create peer", "err", err.Error())
		return
	}

	nLog.Infow("established outbound peer connection", "conn", peer.Identity.CompressedHex())

	if n.peerBook.AddPeer(peer) {
		peer.Start(n.lndIdentity, n.lndHost)
	}
}

func (n *Node) onAccept(conn net.Conn) {
	noiseConn := conn.(*brontide.Conn)
	peer, err := NewPeer(n.reactor, noiseConn, false)

	if err != nil {
		nLog.Errorw("failed to create peer", "err", err.Error())
		return
	}

	nLog.Infow("established inbound peer connection", "conn", peer.Identity.CompressedHex())

	if n.peerBook.AddPeer(peer) {
		peer.Start(n.lndIdentity, n.lndHost)
	}
}

func (n *Node) onDisconnection(req *connmgr.ConnReq) {
	addr := req.Addr.(*lnwire.NetAddress)

	pub, err := crypto.PublicFromBTCEC(addr.IdentityKey)

	if err != nil {
		nLog.Errorw("failed to wrap public key", "err", err.Error())
		return
	}

	nLog.Infow("peer disconnected", "conn", pub.CompressedHex())
	n.peerBook.RemovePeer(pub)
}

func (n *Node) bootstrap(addrs []*lnwire.NetAddress) {
	var wg sync.WaitGroup

	for _, addr := range addrs {
		wg.Add(1)
		go (func() {
			req := &connmgr.ConnReq{
				Addr:      addr,
				Permanent: true,
			}

			n.connMgr.Connect(req)
			wg.Done()
		})()
	}

	wg.Wait()
}
