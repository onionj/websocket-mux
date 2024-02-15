package muxr

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Handler func(*Stream)

type serverHandlers struct {
	Handlers map[string]Handler
}

func (sh *serverHandlers) handle(pattern string, handler Handler) {
	if sh.Handlers == nil {
		sh.Handlers = make(map[string]Handler)
	}
	sh.Handlers[pattern] = handler
}

var defaultServerHandlers = &serverHandlers{}

type wsServer struct {
	addr string
}

func NewWsServer(addr string) *wsServer {
	server := &wsServer{addr: addr}
	return server
}

func (s *wsServer) Handle(pattern string, handler Handler) {
	defaultServerHandlers.handle(pattern, handler)
}

func (s *wsServer) ListenAndServe() {
	fmt.Println("ListenAndServe on", s.addr)

	for pattern := range defaultServerHandlers.Handlers {
		http.HandleFunc(pattern, wsServerHandler)
	}

	if err := http.ListenAndServe(s.addr, nil); err != nil {
		panic(err)
	}
}

func (s *wsServer) ListenAndServeTLS(certFile string, keyFile string) {
	fmt.Println("ListenAndServeTLS on", s.addr)

	for pattern := range defaultServerHandlers.Handlers {
		http.HandleFunc(pattern, wsServerHandler)
	}

	if err := http.ListenAndServeTLS(s.addr, certFile, keyFile, nil); err != nil {
		panic(err)
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func wsServerHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	defer conn.Close()
	fmt.Println("websocket: Open")

	connAdaptor := &ConnAdaptor{Conn: conn}
	streamsManager := StreamManager{Streams: make(map[uint32]*Stream)}

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
					time.Sleep(time.Second * 3)
					stream.Close()
				}()
				defaultServerHandlers.Handlers[request.RequestURI](stream)
			}()
		case TYPE_DATA:
			stream, ok := streamsManager.Get(id)
			if ok {
				go func() {
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
	streamsManager.KillAll()
}
