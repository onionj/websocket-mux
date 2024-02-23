
# Muxr: Efficient WebSocket Multiplexing for Go

Muxr is a Go package designed to simplify WebSocket communication by offering efficient multiplexing capabilities. It enables you to manage multiple streams over a single WebSocket connection, where each stream can handle multiple requests and responses, effectively streamlining communication between clients and servers.

> **Note**: If user connect to a muxr server with a regular WebSocket client, the muxr server handles it in single stream mode.

## Features

- **Stream Multiplexing**: Handle multiple streams concurrently over a single WebSocket connection.
- **Flexible Communication**: Each stream can manage multiple requests and responses independently.
- **Simplified Integration**: Seamlessly integrate into existing Go applications for WebSocket communication.

## Installation

Install Muxr using `go get`:

```bash
go get github.com/onionj/websocket-mux/muxr
```

## Usage

### Muxr Server

```go
package main

import (
    "fmt"
    "github.com/onionj/websocket-mux/muxr"
)

func main() {    
    // Create a new Muxr server listening on port 8080.
    server := muxr.NewWsServer(":8080")

    // Handle WebSocket connections on the "/api" endpoint.
    server.Handle("/api", func(stream *muxr.Stream) {
        for {
            // Read data from the client.
            data, err := stream.Read()
            if err != nil {
                fmt.Println("Error reading from client:", err)
                return
            }
            fmt.Println("Server received:", string(data))

            // Send a response back to the client.
            msg := []byte("Pong")
            err = stream.Write(msg)
            if err != nil {
                fmt.Println("Error writing to client:", err)
                return
            }
            fmt.Println("Server sent:", string(msg))
        }
    })

    // Start the Muxr server.
    err := server.ListenAndServe()
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
}
```

> **Note**: For handling TLS, use: `server.ListenAndServeTLS(certFile, keyFile)`

### Muxr Client

```go
package main

import (
    "fmt"
    "github.com/onionj/websocket-mux/muxr"
)

func main() {
    // Create a new Muxr client connecting to the server.
    client := muxr.NewClient("ws://127.0.0.1:8080/api")
    client.Start()
    defer client.Stop()

    // Establish a stream for communication with the server.
    stream, err := client.Dial()
    if err != nil {
        fmt.Println("Error establishing stream:", err)
        return
    }
    defer stream.Close()

    // Send a message to the server.
    msg := []byte("Ping")
    if err = stream.Write(msg); err != nil {
        fmt.Println("Error sending message:", err)
        return
    }
    fmt.Println("Client sent:", string(msg))

    // Read the response from the server.
    data, err := stream.Read()
    if err != nil {
        fmt.Println("Error reading response:", err)
        return
    }
    fmt.Println("Client received:", string(data))
}

```

## Examples
Explore additional examples provided in the [example](./example/) directory

## License
This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
