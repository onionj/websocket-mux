package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/onionj/websocket-mux/muxr"
)

var (
	client bool
	server bool
)

// simple ping pong application
func main() {
	flag.BoolVar(&client, "client", false, "start a client")
	flag.BoolVar(&server, "server", false, "start a server")
	flag.Parse()

	if client {
		clientRunner()
	} else if server {
		serverRunner()
	} else {
		flag.Usage()
	}
}

func serverRunner() {
	server := muxr.NewServer(":8080")
	server.Handle("/", func(stream *muxr.Stream) {
		data, err := stream.Read()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("server get :", string(data), ",streamId:", stream.Id())

		msg := []byte("Pong")
		err = stream.Write(msg)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("server send:", string(msg), ",streamId:", stream.Id())
	})

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}
}

func clientRunner() {

	client := muxr.NewClient("ws://127.0.0.1:8080/")
	client.Start()
	defer client.Stop()

	wg := sync.WaitGroup{}

	for i := 0; i < 15; i++ {
		wg.Add(1)

		go func(client *muxr.Client) {
			defer wg.Done()

			// create new stream
			stream, err := client.Dial()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer stream.Close()

			msg := []byte("Ping")

			// write to stream
			if err = stream.Write(msg); err != nil {
				fmt.Println(err.Error(), 1, stream.Id())
				return
			}
			fmt.Println("client send:", string(msg), ",streamId:", stream.Id())

			// read pong from server
			data, err := stream.Read()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("client get :", string(data), ",streamId:", stream.Id())
		}(client)
	}
	wg.Wait()
}
