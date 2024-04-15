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

// main is the entry point of the application.
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

// serverRunner runs the server application.
func serverRunner() {
	// Create a new WebSocket multiplexer server listening on port 8080.
	server := muxr.NewServer(":8080")

	// Handle incoming WebSocket connections on the root path.
	server.Handle("/", func(stream *muxr.Stream) {
		// Read the incoming message from the client.
		data, err := stream.Read()
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}
		fmt.Println("Server received:", string(data), ", Stream ID:", stream.Id())

		// Send a response back to the client.
		msg := []byte("Pong")
		err = stream.Write(msg)
		if err != nil {
			fmt.Println("Error writing to client:", err)
			return
		}
		fmt.Println("Server sent:", string(msg), ", Stream ID:", stream.Id())
	})

	// Start the server and listen for incoming WebSocket connections.
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}

// clientRunner runs the client application.
func clientRunner() {

	client := muxr.NewClient("ws://127.0.0.1:8080/")
	closerFunc, err := client.StartForever()
	if err != nil {
		panic(err)
	}
	defer closerFunc()

	// Create a wait group to synchronize goroutines.
	wg := sync.WaitGroup{}

	// Spawn multiple goroutines to simulate concurrent clients.
	for i := 0; i < 15; i++ {
		wg.Add(1)

		go func(client *muxr.Client) {
			defer wg.Done()

			// Create a new stream for communication with the server.
			stream, err := client.Dial()
			if err != nil {
				fmt.Println("Error creating stream:", err)
				return
			}
			defer stream.Close()

			// Send a message to the server.
			msg := []byte("Ping")
			if err = stream.Write(msg); err != nil {
				fmt.Println("Error writing to server:", err)
				return
			}
			fmt.Println("Client sent:", string(msg), ", Stream ID:", stream.Id())

			// Read the response from the server.
			data, err := stream.Read()
			if err != nil {
				fmt.Println("Error reading from server:", err)
				return
			}
			fmt.Println("Client received:", string(data), ", Stream ID:", stream.Id())
		}(client)
	}
	wg.Wait()
}
