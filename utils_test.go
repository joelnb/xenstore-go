package xenstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	requestCounter = 0x0

	assert.Equal(t, uint32(0), RequestID())
	assert.Equal(t, uint32(1), RequestID())
	assert.Equal(t, uint32(2), RequestID())

	maxUint32 := ^uint32(0)
	requestCounter = maxUint32
	assert.Equal(t, maxUint32, RequestID())
	assert.Equal(t, uint32(0), RequestID())
}
