package muxr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameOpen(t *testing.T) {

	payload := []byte{6, 6, 6, 6, 6, 6, 6, 6, 6}
	lenght := uint16(len(payload))
	streamId := uint32(1332)

	frame := Packing(streamId, TYPE_DATA, payload)

	_typ, _lenght, _id := ParseHeader(frame[:NUM_BYTES_HEADER])

	assert.Equal(t, TYPE_DATA, _typ)
	assert.Equal(t, lenght, _lenght)
	assert.Equal(t, streamId, _id)

	_payload := make([]byte, _lenght)
	GetPayload(frame, _payload)
	assert.Equal(t, payload, _payload)
}

func TestFrameClose(t *testing.T) {

	lenght := uint16(0)
	streamId := uint32(1332)

	frame := Packing(streamId, TYPE_CLOSE, []byte{})

	_typ, _lenght, _id := ParseHeader(frame[:NUM_BYTES_HEADER])

	assert.Equal(t, TYPE_CLOSE, _typ)
	assert.Equal(t, lenght, _lenght)
	assert.Equal(t, streamId, _id)

	_payload := make([]byte, _lenght)
	GetPayload(frame, _payload)
	assert.Equal(t, []byte{}, _payload)
}
