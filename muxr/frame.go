package muxr

import (
	"encoding/binary"
)

// Frame represents a muxr frame.
// 7-byte header + payload
//
// {1byteType}{2byteLength}{4byteID}{payload}
type Frame []byte

// Packing packs a frame with the provided ID, type, and payload.
func Packing(id uint32, typ uint8, payload []byte) []byte {

	length := len(payload)
	frame := make([]byte, NUM_BYTES_HEADER+length)

	frame[0] = typ
	binary.BigEndian.PutUint16(frame[1:3], uint16(length))
	binary.BigEndian.PutUint32(frame[3:], id)

	copy(frame[NUM_BYTES_HEADER:], payload)
	return frame
}

// ParseHeader parses the header of a frame and returns the type, length, and stream ID.
func ParseHeader(header []byte) (uint8, uint16, uint32) {
	return uint8(header[0]), // type
		binary.BigEndian.Uint16(header[1:3]), // length
		binary.BigEndian.Uint32(header[3:]) // stream id
}

// GetPayload extracts the payload from a frame.
func GetPayload(frame Frame, buf []byte) (n int) {
	n = copy(buf, frame[NUM_BYTES_HEADER:])
	return
}
