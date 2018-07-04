package p2p

import (
	"sync"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/internal/p2p/msghandler"
)

type Reactor struct {
	chans map[uint64]*reactorChannel
	toAdd map[uint64]*reactorChannel
	toRemove []uint64
	id    uint64
	mut   *sync.Mutex
}

type reactorChannel struct {
	in  chan *Envelope
	out chan *Envelope
}

func NewReactor() *Reactor {
	return &Reactor{
		chans: make(map[uint64]*reactorChannel),
		toAdd: make(map[uint64]*reactorChannel),
		toRemove: make([]uint64, 10),
		id:    0,
		mut:   new(sync.Mutex),
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
			case in := <- ch.in:
				res := r.handle(in)

				if res != nil {
					ch.out <- NewEnvelope(in.Peer, res)
				}
			}
		}
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

	switch msg.MsgType() {
	case lnwire.MsgPing:
		res = msghandler.HandlePing(msg.(*lnwire.Ping))
	}

	return res
}