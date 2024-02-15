package muxr

import "errors"

const (
	MAX_STREAM_ID          = 1 << 31 // 2,147,483,648 // TODO Add type reset stream
	NUM_BYTES_HEADER       = 7
	TYPE_INITIAL     uint8 = 0
	TYPE_DATA        uint8 = 1
	TYPE_CLOSE       uint8 = 2
)

var ErrTunnelClosed = errors.New("tunnel closed")
