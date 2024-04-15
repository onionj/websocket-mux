package muxr

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestPingPongWithMuxrClientAndServer tests ping-pong communication between a muxr client and server.
func TestPingPongWithMuxrClientAndServer(t *testing.T) {
	go func() {
		server := NewServer(":19881")
		server.Handle("/", func(stream *Stream) {
			ping, err := stream.Read()
			assert.Equal(t, nil, err)
			assert.Equal(t, []byte("ping"), ping)
			err = stream.Write([]byte("pong"))
			assert.Equal(t, nil, err)

		})

		err := server.ListenAndServe()
		assert.Equal(t, nil, err)
	}()

	client := NewClient("ws://127.0.0.1:19881/")

	time.Sleep(time.Second / 20)

	err := client.Start()
	assert.Equal(t, nil, err)

	stream, err := client.Dial()
	assert.Equal(t, nil, err)

	defer stream.Close()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ping := []byte("ping")

	err = stream.Write(ping)
	assert.Equal(t, nil, err)

	pong, err := stream.Read()
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("pong"), pong)

	err = stream.Write(ping)
	assert.Equal(t, ErrStreamClosed, err)
}

// TestPingPongWithGorillaClientAndMuxrServer tests ping-pong communication between a Gorilla WebSocket client and a muxr server.
func TestPingPongWithGorillaClientAndMuxrServer(t *testing.T) {
	totalLoop := 100

	go func() {
		server := NewServer(":19882")
		server.Handle("/", func(stream *Stream) {
			for i := 0; i < totalLoop; i++ {
				assert.Equal(t, uint32(0), stream.Id())
				ping, err := stream.Read()
				assert.Equal(t, nil, err)
				assert.Equal(t, []byte("ping"), ping)
				err = stream.Write([]byte("pong"))
				assert.Equal(t, nil, err)
			}
		})

		err := server.ListenAndServe()
		assert.Equal(t, nil, err)
	}()

	time.Sleep(time.Second / 20)

	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:19882/", nil)
	assert.Equal(t, nil, err)

	defer func() {
		closeHandler := conn.CloseHandler()
		_ = closeHandler(websocket.CloseNormalClosure, "")
	}()

	for i := 0; i < totalLoop; i++ {
		ping := []byte("ping")
		err = conn.WriteMessage(websocket.BinaryMessage, ping)
		assert.Equal(t, nil, err)

		_, data, err := conn.ReadMessage()
		assert.Equal(t, nil, err)
		assert.Equal(t, []byte("pong"), data)
	}

	_, _, err = conn.ReadMessage()
	assert.NotEqual(t, nil, err)
}
