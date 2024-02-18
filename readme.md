
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
    server := muxr.NewWsServer(":8080")

    server.Handle("/api", func(stream *muxr.Stream) {
        for {
            data, err := stream.Read()
            if err != nil {
                fmt.Println(err.Error())
                return
            }
            fmt.Println("Server received:", string(data))

            msg := []byte("Pong")
            err = stream.Write(msg)
            if err != nil {
                fmt.Println(err.Error())
                return
            }
            fmt.Println("Server sent:", string(msg))
        }
    })

    err := server.ListenAndServe()
    if err != nil {
        fmt.Println(err)
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
    client := muxr.NewClient("ws://127.0.0.1:8080/api")
    client.Start()
    defer client.Stop()

    stream, err := client.Dial()
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    defer stream.Close()

    msg := []byte("Ping")

    if err = stream.Write(msg); err != nil {
        fmt.Println(err.Error())
        return
    }
    fmt.Println("Client sent:", string(msg))

    data, err := stream.Read()
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    fmt.Println("Client received:", string(data))
}
```

## Examples
There are other examples provided in the [example](./example/) directory

## License
This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
