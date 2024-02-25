package muxr

import "errors"

const (
	VERSION          string = "v0.4.4"
	NUM_BYTES_HEADER        = 7
	TYPE_INITIAL     uint8  = 1 // 0000 0001
	TYPE_DATA        uint8  = 2 // 0000 0010
	TYPE_CLOSE       uint8  = 4 // 0000 0100

	RESIVER_CHANNEL_SIZE = 200
)

var ErrTunnelClosed = errors.New("tunnel closed")
