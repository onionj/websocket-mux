package muxr

import (
	"errors"
	"io"
	"sync"
)

type Stream struct {
	sync.Mutex
	id             uint32
	reciverChannel chan []byte
	isClosed       bool
	connAdaptor    *ConnAdaptor
}

var ErrStreamClosed = errors.New("stream closed")

// expirationTime: If the stream remains inactive for a certain period, it will be automatically terminated or destroyed.
func newStream(
	id uint32,
	connAdaptor *ConnAdaptor,
) *Stream {
	return &Stream{
		id:             id,
		reciverChannel: make(chan []byte),
		isClosed:       false,
		connAdaptor:    connAdaptor,
	}
}

func (st *Stream) Read() ([]byte, error) {
	data, ok := <-st.reciverChannel
	if !ok {
		return nil, io.EOF
	}
	return data, nil
}

func (st *Stream) Write(data []byte) error {
	st.Lock()
	defer st.Unlock()
	if st.isClosed {
		return ErrStreamClosed
	}
	err := st.connAdaptor.WritePacket(TYPE_DATA, st.id, data)
	if err != nil {
		return ErrStreamClosed
	}
	return nil
}

func (st *Stream) Close() {
	st.connAdaptor.WritePacket(TYPE_CLOSE, st.id, []byte{})
	st.Kill()
}

func (st *Stream) Kill() {

	// Clear the channel
	select {
	case <-st.reciverChannel:
	default:
	}

	st.Lock()
	defer st.Unlock()

	if st.isClosed {
		return
	}

	st.isClosed = true
	close(st.reciverChannel)
}

func (st *Stream) IsClose() bool {
	return st.isClosed
}

func (st *Stream) Id() uint32 {
	return st.id
}

type StreamManager struct {
	sync.Mutex
	Streams map[uint32]*Stream
}

func (sm *StreamManager) Get(id uint32) (*Stream, bool) {
	sm.Lock()
	defer sm.Unlock()
	stream, ok := sm.Streams[id]
	return stream, ok
}

func (sm *StreamManager) Set(id uint32, stream *Stream) {
	sm.Lock()
	defer sm.Unlock()
	sm.Streams[id] = stream
}

func (sm *StreamManager) Delete(id uint32) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.Streams, id)
}

func (sm *StreamManager) KillAll() {
	sm.Lock()
	defer sm.Unlock()
	for id, st := range sm.Streams {
		st.Kill()
		delete(sm.Streams, id)
	}
}
