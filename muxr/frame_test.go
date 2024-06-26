package muxr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFrameOpen tests the creation and parsing of a data frame.
func TestFrameOpen(t *testing.T) {
	payload := []byte{6, 6, 6, 6, 6, 6, 6, 6, 6}
	length := uint16(len(payload))
	streamId := uint32(1332)

	frame := Packing(streamId, TYPE_DATA, payload)

	_typ, _length, _id := ParseHeader(frame[:NUM_BYTES_HEADER])

	assert.Equal(t, TYPE_DATA, _typ)
	assert.Equal(t, length, _length)
	assert.Equal(t, streamId, _id)

	_payload := make([]byte, _length)
	GetPayload(frame, _payload)
	assert.Equal(t, payload, _payload)
}

// TestFrameClose tests the creation and parsing of a close frame.
func TestFrameClose(t *testing.T) {

	length := uint16(0)
	streamId := uint32(1332)

	frame := Packing(streamId, TYPE_CLOSE, []byte{})

	_typ, _length, _id := ParseHeader(frame[:NUM_BYTES_HEADER])

	assert.Equal(t, TYPE_CLOSE, _typ)
	assert.Equal(t, length, _length)
	assert.Equal(t, streamId, _id)

	_payload := make([]byte, _length)
	GetPayload(frame, _payload)
	assert.Equal(t, []byte{}, _payload)
}
