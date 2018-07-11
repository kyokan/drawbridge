package p2p

import (
	"sync"
	"github.com/lightningnetwork/lnd/lnwire"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"time"
	"github.com/kyokan/drawbridge/pkg/wire"
)

type Reactor struct {
	chans    map[uint64]*reactorChannel
	toAdd    map[uint64]*reactorChannel
	toRemove []uint64
	id       uint64
	mut      *sync.Mutex
	msgHandlers []MsgHandler
}

type reactorChannel struct {
	in  chan *Envelope
	out chan *Envelope
}

var rLog *zap.SugaredLogger

func init() {
	rLog = logger.Logger.Named("reactor")
}

func NewReactor(msgHandlers []MsgHandler) *Reactor {
	return &Reactor{
		chans:       make(map[uint64]*reactorChannel),
		toAdd:       make(map[uint64]*reactorChannel),
		toRemove:    make([]uint64, 10),
		id:          0,
		mut:         new(sync.Mutex),
		msgHandlers: msgHandlers,
	}
}

func (r *Reactor) AddEnvelopeChan(in chan *Envelope, out chan *Envelope) uint64 {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.id += 1
	r.toAdd[r.id] = &reactorChannel{in: in, out: out}
	return r.id
}

func (r *Reactor) RemoveEnvelopeChan(id uint64) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.toRemove = append(r.toRemove, id)
}

func (r *Reactor) Run() {
	for {
		r.manageMembership()

		for _, ch := range r.chans {
			select {
			case in := <-ch.in:
				res := r.handle(in)

				if res != nil {
					ch.out <- NewEnvelope(in.Peer, res)
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func (r *Reactor) manageMembership() {
	r.mut.Lock()
	defer r.mut.Unlock()

	for id, ch := range r.toAdd {
		r.chans[id] = ch
	}

	r.toAdd = make(map[uint64]*reactorChannel)

	for _, id := range r.toRemove {
		delete(r.chans, id)
	}

	r.toRemove = make([]uint64, 10)
}

func (r *Reactor) handle(envelope *Envelope) lnwire.Message {
	msg := envelope.Msg
	var res lnwire.Message
	var err error

	for _, handler := range r.msgHandlers {
		if handler.CanAccept(msg) {
			res, err = handler.Accept(envelope)
			break
		}
	}

	if err != nil {
		rLog.Warnw("caught error processing message", "msgType", wire.MessageName(msg.MsgType()),
			"err", err.Error())
		return nil
	}

	return res
}
