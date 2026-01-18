package xenstore

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func expectPacket(t *testing.T, p *Packet, res string) {
	buf := bytes.NewBuffer([]byte{})

	if err := p.Pack(buf); err != nil {
		t.Fatal(err)
	}

	b, err := readBytes(buf, len(res))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, res, string(b))
}

func readBytes(r io.Reader, n int) ([]byte, error) {
	result := []byte{}

	for n > 0 {
		b := make([]byte, n)

		nr, err := r.Read(b)
		if err != nil {
			return []byte{}, err
		}

		result = append(result, b[:nr]...)
		n -= nr
	}

	return result, nil
}

func TestPacket(t *testing.T) {
	requestCounter = 0x0

	p1, err := NewPacket(XsDebug, []byte("test"), 0x0)
	if err != nil {
		t.Fatal(err)
	}

	expectPacket(t, p1,
		"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x04\x00\x00\x00test")

	p2, err := NewPacket(XsDebug, []byte("/local/domain/0/name"), 0x0)
	if err != nil {
		t.Fatal(err)
	}

	expectPacket(t, p2,
		"\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x14\x00\x00\x00/local/domain/0/name")
}

func TestTransactionPacket(t *testing.T) {
	requestCounter = 0x0

	p1, err := NewPacket(XsDebug, []byte("test"), 0x9)
	if err != nil {
		t.Fatal(err)
	}

	expectPacket(t, p1,
		"\x00\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00\x04\x00\x00\x00test")

	p2, err := NewPacket(XsDebug, []byte("/local/domain/0/name"), 0x6)
	if err != nil {
		t.Fatal(err)
	}

	expectPacket(t, p2,
		"\x00\x00\x00\x00\x01\x00\x00\x00\x06\x00\x00\x00\x14\x00\x00\x00/local/domain/0/name")
}

func TestPacketPackUnpack(t *testing.T) {
	p1, err := NewPacket(XsDebug, []byte("/local/domain/0"), 0x7)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBuffer([]byte{})
	if err := p1.Pack(b); err != nil {
		t.Fatal(err)
	}

	p2 := &Packet{}

	if err := p2.Unpack(b); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, p1.Header.Op, p2.Header.Op)
	assert.Equal(t, p1.Header.RqId, p2.Header.RqId)
	assert.Equal(t, p1.Header.TxId, p2.Header.TxId)
	assert.Equal(t, p1.Header.Length, p2.Header.Length)

	assert.Equal(t, p1.Payload, p2.Payload)
}
