package xenstore

import (
	"sync"
)

// NewRouter creates a new instance of the Router struct for Transport t with all
// of the correct defaults set.
func NewRouter(t Transport) *Router {
	return &Router{
		transport:  t,
		channelMap: map[uint32]chan *Packet{},
		lock:       sync.Mutex{},
		loop:       true,
	}
}

// Router provides a way of sending a Packet and receiving the reply in return.
// It does ths by intercepting all packets over a Transport and forwarding them
// to listeners over channels.
type Router struct {
	transport  Transport
	channelMap map[uint32]chan *Packet
	lock       sync.Mutex
	loop       bool
}

// Loop starts the Router's internal event loop.
func (r *Router) Loop() error {
	r.loop = true

	for r.loop {
		p, err := r.transport.Receive()
		if err != nil {
			return err
		}

		r.sendToChannel(p)
	}

	return nil
}

// Send sends a Packet to XenStore and returns a channel which the response Packet
// will be sent over when it is received.
func (r *Router) Send(pkt *Packet) (chan *Packet, error) {
	c := make(chan *Packet)

	r.lock.Lock()
	defer r.lock.Unlock()

	if err := r.transport.Send(pkt); err != nil {
		return nil, err
	}

	r.channelMap[pkt.Header.RqId] = c

	return c, nil
}

// Stop ends the internal event loop as soon as the next packet has been received
func (r *Router) Stop() {
	r.loop = false
}

func (r *Router) removeChannel(id uint32) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.channelMap, id)
}

func (r *Router) sendToChannel(pkt *Packet) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if chnl, ok := r.channelMap[pkt.Header.RqId]; ok {
		chnl <- pkt
	} else {
		panic("no channel to send to!")
	}
}
