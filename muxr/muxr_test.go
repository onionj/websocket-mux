package muxr

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPingPong(t *testing.T) {
	go func() {
		server := NewWsServer(":8080")

		server.Handle("/api", func(stream *Stream) {

			ping, err := stream.Read()
			assert.Equal(t, nil, err)
			assert.Equal(t, []byte("ping"), ping)
			err = stream.Write([]byte("pong"))
			assert.Equal(t, nil, err)

		})

		server.ListenAndServe()
	}()

	time.Sleep(time.Second / time.Duration(2))

	client := NewClient()
	client.Start("ws://127.0.0.1:8080/api")

	stream, err := client.Dial()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
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

	time.Sleep(time.Second * time.Duration(4))
	err = stream.Write(ping)
	assert.Equal(t, ErrStreamClosed, err)
}
