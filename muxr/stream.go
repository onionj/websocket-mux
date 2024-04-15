package muxr

import (
	"errors"
	"io"
	"sync"
)

type Stream struct {
	sync.Mutex
	id           uint32
	isClosed     bool
	ReceiverChan chan []byte
	ConnAdaptor  *ConnAdaptor
}

var ErrStreamClosed = errors.New("stream closed")

// newStream creates a new Stream instance with the provided ID and connection adaptor.
func newStream(
	id uint32,
	connAdaptor *ConnAdaptor,
	receiverChanSize int,
) *Stream {
	return &Stream{
		id:           id,
		isClosed:     false,
		ReceiverChan: make(chan []byte, receiverChanSize),
		ConnAdaptor:  connAdaptor,
	}
}

// Read reads data from the stream's receiver channel.
func (st *Stream) Read() ([]byte, error) {
	data, ok := <-st.ReceiverChan
	if !ok {
		return nil, io.EOF
	}
	return data, nil
}

// Write writes data to the stream.
func (st *Stream) Write(data []byte) error {
	st.Lock()
	defer st.Unlock()
	if st.isClosed {
		return ErrStreamClosed
	}
	err := st.ConnAdaptor.WritePacket(TYPE_DATA, st.id, data)
	if err != nil {
		return ErrTunnelClosed
	}
	return nil
}

// Close closes the stream.
func (st *Stream) Close() {
	_ = st.ConnAdaptor.WritePacket(TYPE_CLOSE, st.id, []byte{})
	st.Kill()
}

// Kill marks the stream as closed and closes its receiver channel.
func (st *Stream) Kill() {

	st.Lock()
	defer st.Unlock()

	if st.isClosed {
		return
	}

	st.isClosed = true
	close(st.ReceiverChan)
}

// IsClose returns true if the stream is closed, otherwise false.
func (st *Stream) IsClose() bool {
	return st.isClosed
}

// Id returns the ID of the stream.
func (st *Stream) Id() uint32 {
	return st.id
}

type StreamManager struct {
	sync.Mutex
	Streams map[uint32]*Stream
}

// Get retrieves a stream from the stream manager by its ID.
func (sm *StreamManager) Get(id uint32) (*Stream, bool) {
	sm.Lock()
	defer sm.Unlock()
	stream, ok := sm.Streams[id]
	return stream, ok
}

// Set adds a stream to the stream manager.
func (sm *StreamManager) Set(id uint32, stream *Stream) {
	sm.Lock()
	defer sm.Unlock()
	sm.Streams[id] = stream
}

// Delete removes a stream from the stream manager by its ID.
func (sm *StreamManager) Delete(id uint32) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.Streams, id)
}

// KillAll closes all streams in the stream manager.
func (sm *StreamManager) KillAll() {
	sm.Lock()
	defer sm.Unlock()
	for id, st := range sm.Streams {
		st.Kill()
		delete(sm.Streams, id)
	}
}
