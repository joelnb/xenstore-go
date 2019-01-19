package xenstore

import (
	"io"
	"net"
	"os"
)

// Transport is an interface for sending and receiving data from XenStore.
type Transport interface {
	// Send a packet to the XenStore backend.
	Send(*Packet) error
	// Receive a packet from the XenStore backend.
	Receive() (*Packet, error)
	// Close if required by the underlying implementation.
	Close() error
}

// ReadWriteTransport is an implementation of the Transport interface which works for any
// io.ReadWriteCloser..
type ReadWriteTransport struct {
	rw   io.ReadWriteCloser
	open bool
}

func (r *ReadWriteTransport) Close() error {
	if r.open {
		r.open = false
		return r.rw.Close()
	}

	// Possibly this should return an error as it is already closed
	return nil
}

func (r *ReadWriteTransport) Send(p *Packet) error {
	if !r.open {
		panic("Send on closed transport")
	}

	return p.Pack(r.rw)
}

func (r *ReadWriteTransport) Receive() (*Packet, error) {
	if !r.open {
		panic("Receive on closed transport")
	}

	p := &Packet{}
	return p, p.Unpack(r.rw)
}

// Check if the underlying io.ReadWriteCloser has been closed yet.
func (r *ReadWriteTransport) IsOpen() bool {
	return r.open
}

// UnixSocketTransport is an implementation of Transport which sends/receives data from
// XenStore using a unix socket.
type UnixSocketTransport struct {
	*ReadWriteTransport

	Path string
}

// NewUnixSocketTransport creates a new connected UnixSocketTransport.
func NewUnixSocketTransport(path string) (*UnixSocketTransport, error) {
	c, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}

	return &UnixSocketTransport{
		&ReadWriteTransport{
			rw:   c,
			open: true,
		},
		path,
	}, nil
}

// XenBusTransport is an implementation of Transport which sends/receives data from
// XenStore using the special XenBus device on Linux (and possibly other Unix operating
// systems)
type XenBusTransport struct {
	*ReadWriteTransport

	Path string
}

// Create a new connected XenBusTransport.
func NewXenBusTransport(path string) (*XenBusTransport, error) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModeCharDevice|os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &XenBusTransport{
		&ReadWriteTransport{
			rw:   file,
			open: true,
		},
		path,
	}, nil
}
