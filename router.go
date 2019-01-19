package xenstore

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// NewRouter creates a new instance of the Router struct for Transport t with all
// of the correct defaults set.
func NewRouter(t Transport) *Router {
	return &Router{
		transport:  t,
		channelMap: map[uint32]chan *Packet{},
		watchMap:   map[string][]chan *Packet{},
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
	watchMap   map[string][]chan *Packet
	lock       sync.Mutex
	loop       bool
}

// Start starts the Router's internal event loop.
func (r *Router) Start() error {
	r.loop = true

OUTER:
	for r.loop {
		p, err := r.transport.Receive()
		if err != nil {
			if !r.loop {
				// If the error is that the file was already closed then it likely
				// means that we closed it so swallow this specific error.
				switch v := err.(type) {
				case *os.PathError:
					if v.Err == os.ErrClosed {
						break OUTER
					}
				}
			}

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

	if pkt.Header.Op == XsWatch {
		payloadParts := strings.Split(pkt.payloadString(), "\u0000")

		if _, ok := r.watchMap[payloadParts[1]]; !ok {
			r.watchMap[payloadParts[1]] = []chan *Packet{}
		}

		r.watchMap[payloadParts[1]] = append(r.watchMap[payloadParts[1]], c)
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

func (r *Router) removeWatchChannel(token string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.watchMap, token)
}

func (r *Router) sendToChannel(pkt *Packet) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if pkt.Header.Op == XsWatchEvent {
		payloadParts := strings.Split(pkt.payloadString(), "\u0000")
		watchToken := payloadParts[1]

		if channels, ok := r.watchMap[watchToken]; ok {
			for _, chnl := range channels {
				chnl <- pkt
			}
		} else {
			panic(fmt.Sprintf("no channel(s) to send packet for '%s' to!", watchToken))
		}
	} else {
		if chnl, ok := r.channelMap[pkt.Header.RqId]; ok {
			chnl <- pkt
		} else {
			panic(fmt.Sprintf("no channel to send packet for %d to!", pkt.Header.RqId))
		}
	}
}
