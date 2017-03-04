package xenstore

import (
	"bytes"
)

type BufferTransport struct {
	*ReadWriteTransport
}

func NewBufferTransport() *BufferTransport {
	var buf = bytes.NewBuffer([]byte{})

	return &BufferTransport{
		&ReadWriteTransport{
			rw:   BufCloser{buf},
			open: true,
		},
	}
}

type BufCloser struct {
	*bytes.Buffer
}

func (b BufCloser) Close() error {
	return nil
}
