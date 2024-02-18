package muxr

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnAdaptor struct {
	sync.Mutex
	Conn *websocket.Conn
}

func (c *ConnAdaptor) WritePacket(typ uint8, id uint32, data []byte) error {
	c.Lock()
	defer c.Unlock()
	if id == 0 {
		// handle single stream mode (in case the client is not a Muxr client)
		return c.Conn.WriteMessage(websocket.BinaryMessage, data)
	}
	return c.Conn.WriteMessage(websocket.BinaryMessage, Packing(id, typ, data))
}
