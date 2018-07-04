package p2p

import (
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"net"
	"github.com/btcsuite/btcd/connmgr"
	"time"
	"sync"
	"errors"
)

var nLog *zap.SugaredLogger

func init() {
	nLog = logger.Logger.Named("node")
}

type Node struct {
	reactor  *Reactor
	connMgr  *connmgr.ConnManager
	peerBook *PeerBook
}

func StartNode(reactor *Reactor, addr string, port string, bootstrapPeers []string) (*Node) {
	nLog.Infow("starting p2p node", "p2pIp", addr, "p2pPort", port)

	listenAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, port))

	if err != nil {
		nLog.Panicw("failed to parse TCP address", "err", err)
	}

	listener, err := net.ListenTCP("tcp", listenAddr)

	if err != nil {
		nLog.Panicw("failed to listen to TCP address", "err", err, "addr", listenAddr.String())
	}

	node := &Node{
		reactor: reactor,
	}

	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners: []net.Listener{
			listener,
		},
		OnAccept:       node.OnAccept,
		RetryDuration:  time.Second * 5,
		TargetOutbound: 100,
		Dial: func(a net.Addr) (net.Conn, error) {
			if a == nil {
				return nil, errors.New("addr is nil")
			}

			return net.Dial("tcp", a.String())
		},
		OnConnection:    node.OnConnection,
		OnDisconnection: node.OnDisconnection,
		GetNewAddress:   node.GetNextAddress,
	})

	if err != nil {
		nLog.Panicw("failed to start p2p node", "err", err)
	}

	node.connMgr = cmgr

	if len(bootstrapPeers) > 0 {
		addrs, err := ResolveTCPAddrs(bootstrapPeers)

		if err != nil {
			nLog.Errorw("failed to resolve bootstrap peers", "err", err, "bootstrapPeers", bootstrapPeers)
		}

		node.peerBook = NewPeerBook(addrs)
	} else {
		node.peerBook = NewPeerBook(nil)
	}

	cmgr.Start()
	go node.bootstrap()

	return node
}

func (n *Node) GetNextAddress() (net.Addr, error) {
	return n.peerBook.PopDisconnectedPeer(), nil
}

func (n *Node) OnConnection(req *connmgr.ConnReq, conn net.Conn) {
	nLog.Infow("established outbound peer connection", "conn", conn.RemoteAddr().String())
	n.peerBook.PushConnectedPeer(conn.RemoteAddr())

	peer := NewPeer(n.reactor, conn, true)
	peer.Start()
}

func (n *Node) OnAccept(conn net.Conn) {
	nLog.Infow("established inbound peer connection", "conn", conn.RemoteAddr().String())
	n.peerBook.PushConnectedPeer(conn.RemoteAddr())

	peer := NewPeer(n.reactor, conn, false)
	peer.Start()
}

func (n *Node) OnDisconnection(req *connmgr.ConnReq) {
	nLog.Infow("peer disconnected", "conn", req.Addr.String())
}

func (n *Node) bootstrap() {
	var wg sync.WaitGroup

	for i := 0; i < n.peerBook.DisconnectedCount(); i++ {
		wg.Add(1)
		go (func() {
			n.connMgr.NewConnReq()
			wg.Done()
		})()
	}

	wg.Wait()
}
