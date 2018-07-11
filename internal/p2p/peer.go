package p2p

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"bytes"
	"time"
	"io"
	"sync/atomic"
	"sync"
	"github.com/lightningnetwork/lnd/brontide"
	"github.com/roasbeef/btcd/btcec"
	"github.com/kyokan/drawbridge/internal/conv"
	"fmt"
)

var pLog *zap.SugaredLogger

const idleTimeout = time.Minute * 5
const pingInterval = time.Second * 5

func init() {
	pLog = logger.Logger.Named("peer")
}

type Peer struct {
	reactor        *Reactor
	conn           *brontide.Conn
	selfOriginated bool
	writeBuf       *[65535]byte
	incomingQueue  chan *Envelope
	outgoingQueue  chan *Envelope
	disconnected   uint32
	wg             *sync.WaitGroup

	Identity *btcec.PublicKey
}

func NewPeer(reactor *Reactor, conn *brontide.Conn, selfOriginated bool) *Peer {
	return &Peer{
		reactor:        reactor,
		conn:           conn,
		selfOriginated: selfOriginated,
		writeBuf:       new([65535]byte),
		incomingQueue:  make(chan *Envelope, 10),
		outgoingQueue:  make(chan *Envelope, 10),
		disconnected:   0,
		wg:             new(sync.WaitGroup),
		Identity:       conn.RemotePub(),
	}
}

func (p *Peer) Start() {
	fmt.Println("start peer")
	p.reactor.AddEnvelopeChan(p.incomingQueue, p.outgoingQueue)

	go p.readHandler()
	go p.writeHandler()
	go p.pingHandler()

	globalFeats := lnwire.NewRawFeatureVector()
	localFeats := lnwire.NewRawFeatureVector()
	msg := lnwire.NewInitMessage(globalFeats, localFeats)
	p.outgoingQueue <- NewEnvelope(p, msg)
}

func (p *Peer) Stop() (error) {
	atomic.StoreUint32(&p.disconnected, 1)
	p.wg.Wait()
	close(p.incomingQueue)
	close(p.outgoingQueue)
	return p.conn.Close()
}

func (p * Peer) Send(msg lnwire.Message) error {
	p.outgoingQueue <- NewEnvelope(p, msg)
	return nil
}

func (p *Peer) readHandler() {
	p.wg.Add(1)

	idleTimer := time.AfterFunc(idleTimeout, func() {
		pLog.Errorf("peer timed out", "peer", p)
	})

	for {
		if atomic.LoadUint32(&p.disconnected) == 1 {
			p.wg.Done()
			return
		}

		idleTimer.Stop()
		nextMessage, err := p.readMessage()

		if err != nil {
			if err == io.EOF {
				pLog.Infow("remote end hung up", "peer", p, "err", err)
				p.Stop()
			} else {
				pLog.Infow("failed to read message", "peer", p, "err", err)
			}

			continue
		}

		pLog.Infow("received message", "peer", p, "wireMsg", nextMessage.MsgType().String())

		p.incomingQueue <- NewEnvelope(p, nextMessage)
		idleTimer.Reset(idleTimeout)
	}
}

func (p *Peer) writeHandler() {
	p.wg.Add(1)

	for {
		if atomic.LoadUint32(&p.disconnected) == 1 {
			p.wg.Done()
			return
		}

		select {
		case envelope := <-p.outgoingQueue:
			pLog.Infow("writing lnwire message", "peer", p, "wireMsg", envelope.Msg.MsgType().String())
			err := p.writeMessage(envelope.Msg)

			if err != nil {
				pLog.Errorw("failed to write message", "peer", p, "err", err)
			}
		}
	}
}

func (p *Peer) pingHandler() {
	p.wg.Add(1)

	tick := time.NewTicker(pingInterval)
	defer tick.Stop()

	for {
		if atomic.LoadUint32(&p.disconnected) == 1 {
			p.wg.Done()
			return
		}

		select {
		case <-tick.C:
			p.outgoingQueue <- NewEnvelope(p, lnwire.NewPing(16))
		}
	}
}

func (p *Peer) readMessage() (lnwire.Message, error) {
	rawMsg, err := p.conn.ReadNextMessage()

	if err != nil {
		return nil, err
	}

	msg, err := lnwire.ReadMessage(bytes.NewReader(rawMsg), 0)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (p *Peer) writeMessage(msg lnwire.Message) error {
	b := bytes.NewBuffer(p.writeBuf[0:0:len(p.writeBuf)])
	_, err := lnwire.WriteMessage(b, msg, 0)

	if err != nil {
		return err
	}

	p.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = p.conn.Write(b.Bytes())
	return err
}

func (p *Peer) String() string {
	return conv.PubKeyToHex(p.Identity)
}
