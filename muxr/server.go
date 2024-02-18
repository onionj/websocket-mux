package muxr

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Handler func(*Stream)

type wsServer struct {
	bindAddr string
	handlers map[string]Handler
}

func NewServer(bindAddr string) *wsServer {
	server := &wsServer{bindAddr: bindAddr, handlers: make(map[string]Handler)}
	return server
}

func (s wsServer) Handle(pattern string, handler Handler) {
	s.handlers[pattern] = handler
}

func (s *wsServer) ListenAndServe() error {
	fmt.Println("ListenAndServe on", s.bindAddr)

	httpServeMux := http.NewServeMux()

	for pattern := range s.handlers {
		httpServeMux.HandleFunc(pattern, s.wsServerHandler)
	}

	if err := http.ListenAndServe(s.bindAddr, httpServeMux); err != nil {
		return err
	}
	return nil
}

func (s *wsServer) ListenAndServeTLS(certFile string, keyFile string) error {
	fmt.Println("ListenAndServeTLS on", s.bindAddr)

	httpServeMux := http.NewServeMux()

	for pattern := range s.handlers {
		httpServeMux.HandleFunc(pattern, s.wsServerHandler)
	}

	if err := http.ListenAndServeTLS(s.bindAddr, certFile, keyFile, httpServeMux); err != nil {
		return err
	}
	return nil
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func (s *wsServer) wsServerHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)

	wsMuxVersion := request.Header.Get("websocket-mux")
	IsMuxClient := wsMuxVersion != ""

	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	defer func() {
		closeHandler := conn.CloseHandler()
		closeHandler(websocket.CloseNormalClosure, "")
	}()

	fmt.Println("websocket: Open, IsMuxClient:", IsMuxClient, wsMuxVersion)

	connAdaptor := &ConnAdaptor{Conn: conn}
	streamsManager := StreamManager{Streams: make(map[uint32]*Stream)}
	defer streamsManager.KillAll()

	// Handle muxr conn (in case the client is muxr.client)
	if IsMuxClient {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				break
			}

			typ, lenght, id := ParseHeader(data[:NUM_BYTES_HEADER])

			switch typ {
			case TYPE_INITIAL:
				stream := newStream(id, connAdaptor)
				streamsManager.Set(id, stream)
				go func() {
					defer func() {
						streamsManager.Delete(id)
						stream.Close()
					}()
					if parsedURL, err := url.Parse(request.RequestURI); err == nil {
						s.handlers[parsedURL.Path](stream)
					}
				}()
			case TYPE_DATA:
				stream, ok := streamsManager.Get(id)
				if ok {
					func() {
						stream.Lock()
						defer stream.Unlock()
						if !stream.isClosed {
							stream.reciverChannel <- data[NUM_BYTES_HEADER : NUM_BYTES_HEADER+lenght]
						}
					}()
				}
			case TYPE_CLOSE:
				stream, ok := streamsManager.Get(id)
				if ok {
					stream.Kill()
					streamsManager.Delete(id)
				}
				continue
			}
		}
	} else { // Single stream mode (handle normal websocket clients)
		streamId := uint32(0)
		stream := newStream(streamId, connAdaptor)
		streamsManager.Set(streamId, stream)
		parsedURL, err := url.Parse(request.RequestURI)
		if err != nil {
			return
		}
		go func() {
			defer func() {
				stream.Kill()
				closeHandler := conn.CloseHandler()
				closeHandler(websocket.CloseNormalClosure, "")
			}()

			s.handlers[parsedURL.Path](stream)
		}()

		isAlive := true
		for isAlive {
			func() {
				_, data, err := conn.ReadMessage()
				if err != nil {
					fmt.Println(err)
					isAlive = false
					return
				}

				stream.Lock()
				defer stream.Unlock()

				if stream.isClosed {
					fmt.Println("websocket: stream was closed")
					isAlive = false
					return
				}
				stream.reciverChannel <- data
			}()
		}
	}
}
