# Muxr

Muxr is a websocket multiplexer. muxr multiplexes one websocket connection with a virtual channel called a stream. muxr sends and receives multiple requests and responses in parallel using a single websocket connection. Therefore, our application will be fast and simple.

# Simple Usage:

## Muxr server:

```go
server := muxr.NewWsServer(":8080")

server.Handle("/api", func(stream *muxr.Stream) {
        data, err := stream.Read()
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        fmt.Println("server get :", string(data))

		msg := []byte("Pong")
        err = stream.Write(msg)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        fmt.Println("server send:", string(msg))
	})

server.ListenAndServe()
```
> For TLS: Replace the last line with this line:
```go
server.ListenAndServeTLS(certFile, keyFile)
```

## Muxr Client:

```go
client := muxr.NewClient()
client.Start("ws://127.0.0.1:8080/api")
defer client.Stop()

stream, err := client.Dial()
if err != nil {
    fmt.Println(err.Error())
    return
}
defer stream.Close()

msg := []byte("Ping")

if err = stream.Write(msg); err != nil {
    fmt.Println(err.Error(), 1, stream.Id())
    return
}
fmt.Println("client send:", string(msg))

data, err := stream.Read()
if err != nil {
	fmt.Println(err.Error())
	return
}
fmt.Println("client get :", string(data)
```

#### Run server:

```bash
go run ./server.go
```

#### Run client:

```bash
go run ./client.go
```

#### server logs:
```log
ListenAndServe on :8080
websocket: Open
server get : Ping
server send: Pong
websocket: close 1000 (normal)
```

#### client logs:
```log
client send: Ping
client get : Pong
```

# Examples: you can find more examples here: [Examples Folder](./examples/)
