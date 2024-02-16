package muxr

import (
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	sync.Mutex
	serverAddr     string
	connAdaptor    *ConnAdaptor
	isClosed       bool
	counter        uint32
	streamsManager StreamManager
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		isClosed:   true,
		counter:    1,
		streamsManager: StreamManager{
			Streams: make(map[uint32]*Stream),
		},
	}
}

func (c *Client) Start() error {
	c.Lock()
	defer c.Unlock()

	conn, _, err := websocket.DefaultDialer.Dial(c.serverAddr, nil)
	if err != nil {
		return err
	}
	c.connAdaptor = &ConnAdaptor{Conn: conn}
	c.isClosed = false

	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Println("ReadMessage err:", err)
				break
			}

			typ, lenght, id := ParseHeader(data[:NUM_BYTES_HEADER])

			switch typ {
			case TYPE_DATA:
				if stream, ok := c.streamsManager.Get(id); ok {
					func(stream *Stream) {
						stream.Lock()
						defer stream.Unlock()
						if !stream.isClosed {
							stream.reciverChannel <- data[NUM_BYTES_HEADER : NUM_BYTES_HEADER+lenght]
						}
					}(stream)
				}
			case TYPE_CLOSE:
				stream, ok := c.streamsManager.Get(id)
				if ok {
					go func() {
						stream.Kill()
						c.streamsManager.Delete(id)
					}()
				}
			}
		}
		c.Stop()
	}()
	return nil
}

func (c *Client) Stop() {
	c.Lock()
	defer c.Unlock()
	c.isClosed = true
	c.streamsManager.KillAll()
	closeHandler := c.connAdaptor.Conn.CloseHandler()
	closeHandler(websocket.CloseNormalClosure, "normal")
}

func (c *Client) getStreamId() uint32 {
	c.Lock()
	defer c.Unlock()

	current := c.counter
	c.counter += 1
	return current
}

func (c *Client) Dial() (*Stream, error) {
	streamId := c.getStreamId()

	if c.isClosed {
		return nil, errors.New("the tunnel is not ready")
	}

	stream := newStream(streamId, c.connAdaptor)
	stream.connAdaptor.WritePacket(TYPE_INITIAL, streamId, []byte{})
	c.streamsManager.Set(streamId, stream)
	return stream, nil
}
