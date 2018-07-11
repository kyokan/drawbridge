package p2p

import (
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"net"
	"github.com/roasbeef/btcd/connmgr"
	"time"
	"sync"
	"errors"
	"github.com/kyokan/drawbridge/pkg"
	"github.com/roasbeef/btcd/btcec"
	"github.com/lightningnetwork/lnd/brontide"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/internal/conv"
)

var nLog *zap.SugaredLogger

func init() {
	nLog = logger.Logger.Named("node")
}

type Node struct {
	reactor  *Reactor
	connMgr  *connmgr.ConnManager
	peerBook *PeerBook
	bootstrapPeers []string
	addr string
	port string
}

func NewNode(reactor *Reactor, config *pkg.Config) (*Node, error) {
	return &Node{
		reactor: reactor,
		peerBook: NewPeerBook(),
		addr: config.P2PAddr,
		port: config.P2PPort,
		bootstrapPeers: config.BootstrapPeers,
	}, nil
}

func (n *Node) Start(identityKey *btcec.PrivateKey) error {
	nLog.Infow("starting p2p node", "p2pIp", n.addr, "p2pPort", n.port, "identityKey", conv.PubKeyToHex(identityKey.PubKey()))

	listenAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(n.addr, n.port))

	if err != nil {
		nLog.Panicw("failed to parse TCP address", "err", err)
	}

	listener, err := brontide.NewListener(identityKey, listenAddr.String())

	if err != nil {
		nLog.Panicw("failed to listen to TCP address", "err", err, "addr", listenAddr.String())
	}

	node := &Node{
		reactor: n.reactor,
		peerBook: NewPeerBook(),
	}

	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners: []net.Listener{
			listener,
		},
		OnAccept:       node.onAccept,
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
		OnConnection:    node.onConnection,
		OnDisconnection: node.onDisconnection,
	})

	if err != nil {
		nLog.Panicw("failed to start p2p node", "err", err)
	}

	node.connMgr = cmgr

	cmgr.Start()

	if len(n.bootstrapPeers) > 0 {
		addrs, err := ResolveAddrs(n.bootstrapPeers)

		if err != nil {
			nLog.Errorw("failed to resolve bootstrap peers", "err", err, "bootstrapPeers", n.bootstrapPeers)
		}

		go node.bootstrap(addrs)
	}

	return nil
}

func (n *Node) SendPeer(pub *btcec.PublicKey, msg lnwire.Message) error {
	peer := n.peerBook.FindPeer(pub)

	if peer == nil {
		return errors.New("no peer with id " + conv.PubKeyToHex(pub) + " found")
	}

	return peer.Send(msg)
}

func (n *Node) onConnection(req *connmgr.ConnReq, conn net.Conn) {
	noiseConn := conn.(*brontide.Conn)
	nLog.Infow("established outbound peer connection", "conn", conv.PubKeyToHex(noiseConn.RemotePub()))
	peer := NewPeer(n.reactor, noiseConn, true)

	if n.peerBook.AddPeer(peer) {
		peer.Start()
	}
}

func (n *Node) onAccept(conn net.Conn) {
	noiseConn := conn.(*brontide.Conn)
	nLog.Infow("established inbound peer connection", "conn", conv.PubKeyToHex(noiseConn.RemotePub()))
	peer := NewPeer(n.reactor, noiseConn, false)

	if n.peerBook.AddPeer(peer) {
		peer.Start()
	}
}

func (n *Node) onDisconnection(req *connmgr.ConnReq) {
	addr := req.Addr.(*lnwire.NetAddress)
	nLog.Infow("peer disconnected", "conn", conv.PubKeyToHex(addr.IdentityKey))
	n.peerBook.RemovePeer(addr.IdentityKey)
}

func (n *Node) bootstrap(addrs []*lnwire.NetAddress) {
	var wg sync.WaitGroup

	for _, addr := range addrs {
		wg.Add(1)
		go (func() {
			req := &connmgr.ConnReq{
				Addr: addr,
				Permanent: true,
			}

			n.connMgr.Connect(req)
			wg.Done()
		})()
	}

	wg.Wait()
}
