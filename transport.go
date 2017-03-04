package xenstore

import (
	"io"
	"net"
	"os"
)

type Transport interface {
	Send(*Packet) error
	Receive() (*Packet, error)
	Close() error
}

type ReadWriteTransport struct {
	rw   io.ReadWriteCloser
	open bool
}

func (r *ReadWriteTransport) Send(p *Packet) error {
	return p.Pack(r.rw)
}

func (r *ReadWriteTransport) Receive() (*Packet, error) {
	p := &Packet{}
	return p, p.Unpack(r.rw)
}

func (r *ReadWriteTransport) IsOpen() bool {
	return r.open
}

func (r *ReadWriteTransport) Close() error {
	if r.open {
		r.open = false
		return r.rw.Close()
	}

	// Possibly this should return an error as it is already closed
	return nil
}

type UnixSocketTransport struct {
	*ReadWriteTransport

	Path string
}

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

type XenBusTransport struct {
	*ReadWriteTransport

	Path string
}

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
