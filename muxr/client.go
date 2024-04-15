package muxr

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a client that connects to a server.
type Client struct {
	sync.Mutex
	serverAddr     string
	connAdaptor    *ConnAdaptor
	isClosed       bool
	counter        uint32
	streamsManager StreamManager
}

// NewClient creates and returns a new instance of Client with the provided server address.
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

// Start establishes a connection and initiates communication with the server.
// It returns an error if connection setup fails.
func (c *Client) Start() error {
	c.Lock()
	defer c.Unlock()

	header := make(http.Header)
	header.Set("websocket-mux", VERSION)

	conn, _, err := websocket.DefaultDialer.Dial(c.serverAddr, header)
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

			typ, length, id := ParseHeader(data[:NUM_BYTES_HEADER])

			switch typ {
			case TYPE_DATA:
				if stream, ok := c.streamsManager.Get(id); ok {
					func(stream *Stream) {
						stream.Lock()
						defer stream.Unlock()
						if !stream.isClosed {
							select {
							case stream.ReceiverChan <- data[NUM_BYTES_HEADER : NUM_BYTES_HEADER+length]:
							default:
								fmt.Println("muxr: stream buffer is full")
							}
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

// StartForever Begin client and launch a goroutine with a loop to continuously restart the tunnel if it closes.
func (c *Client) StartForever() (closer func(), err error) {
	exitChan := make(chan struct{}, 1)
	err = c.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-exitChan:
				fmt.Println("muxr StartForEver: closed")
				return
			default:
				time.Sleep(time.Second / 10)
				if c.isClosed {
					fmt.Println("muxr StartForEver: restarting")
					err = c.Start()
					if err != nil {
						fmt.Println("muxr StartForEver error:", err)
						time.Sleep(time.Second)
					}
				}
			}
		}
	}()

	return func() {
		exitChan <- struct{}{}
		c.Stop()
	}, nil
}

// Stop closes the client's connection and releases associated resources.
func (c *Client) Stop() {
	c.Lock()
	defer c.Unlock()
	if !c.isClosed {
		closeHandler := c.connAdaptor.Conn.CloseHandler()
		closeHandler(websocket.CloseNormalClosure, "normal")
	}
	c.isClosed = true
	c.streamsManager.KillAll()
}

func (c *Client) getStreamId() uint32 {
	c.Lock()
	defer c.Unlock()

	current := c.counter
	c.counter += 1
	return current
}

// Dial establishes a new stream with the server.
// It returns a pointer to the created stream and any error encountered.
func (c *Client) Dial() (*Stream, error) {
	streamId := c.getStreamId()

	loopCounter := 0
	for ; c.isClosed && loopCounter < 10; loopCounter++ {
		time.Sleep(time.Second / 10)
	}

	if c.isClosed {
		return nil, ErrTunnelClosed
	}

	stream := newStream(streamId, c.connAdaptor, RESIVER_CHANNEL_SIZE)
	stream.ConnAdaptor.WritePacket(TYPE_INITIAL, streamId, []byte{})
	c.streamsManager.Set(streamId, stream)
	return stream, nil
}
