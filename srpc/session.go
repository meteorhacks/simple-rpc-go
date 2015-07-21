package srpc

import (
	"encoding/binary"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

const (
	// MaxMessageSize is the maximum valid size for a message. (limit: 16MB)
	MaxMessageSize = 1024 * 1024 * 16
)

var (
	// ErrDisconnected is returned when the connection is lost
	ErrDisconnected = Error{1, "disconnected"}

	// ErrNoSuchMethod is returned when the requested method doesn't exist
	ErrNoSuchMethod = Error{2, "method not found"}

	// ErrReadFailed is returned when a message could not be read
	// from the network socket completely and successfully.
	ErrReadFailed = Error{3, "failed to read message"}

	// ErrWriteFailed is returned when a message could not be written
	// to the network socket completely and successfully.
	ErrWriteFailed = Error{4, "failed to write message"}

	// ResDisconnected is returned as response when the connection
	// is lost before completing a method call and getting a response.
	ResDisconnected = Response{Error: &ErrDisconnected}

	// ResNoSuchMethod is returned as response when the requested
	// method does not exist on the server session.
	ResNoSuchMethod = Response{Error: &ErrNoSuchMethod}

	// MsgNoSuchMethod is returned as response when the requested
	// method does not exist on the server session.
	MsgNoSuchMethod = Message{Response: &ResNoSuchMethod}
)

// Handler receives a method call (for a method name) and responds to it.
// If a Handler is not available, an ErrMethodMissing will be returned.
type Handler func(req []byte) (res []byte, err error)

// Session is a simple rpc peer which can act as both server and client.
// A session can call methods on other peers and also handle method calls.
// Sessions can be made when a client dials in to a server and whenever a
// server receives a new connection.
type Session interface {
	Call(name string, pld []byte) (out []byte, err error)
	Handle(handlers map[string]Handler)
	Close() (err error)
}

type session struct {
	nextID     uint32
	handlers   map[string]Handler
	inflight   map[uint32]chan *Response
	connection net.Conn
	writeMutx  *sync.Mutex
}

func newSession(c net.Conn) (s *session) {
	return &session{
		handlers:   make(map[string]Handler),
		inflight:   make(map[uint32]chan *Response),
		connection: c,
		writeMutx:  &sync.Mutex{},
	}
}

// Dial connects to `address` and creates a new `Session`
func Dial(address string) (s Session, err error) {
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	ss := newSession(c)
	go ss.handleConnection()

	return ss, nil
}

// Serve accepts new connections on `address` and creates sessions
// Use `for s := range ch { ... }` to get created sessions.
func Serve(address string) (ch chan Session, l net.Listener, err error) {
	l, err = net.Listen("tcp", address)
	if err != nil {
		return nil, nil, err
	}

	ch = make(chan Session)

	go func() {
		defer close(ch)

		for {
			c, err := l.Accept()
			if err != nil {
				break
			}

			s := newSession(c)
			ch <- s

			go s.handleConnection()
		}
	}()

	return ch, l, nil
}

func (s *session) Call(name string, pld []byte) (out []byte, err error) {
	id := s.getNextID()
	ch := make(chan *Response)
	s.inflight[id] = ch

	req := &Request{Id: id, Method: name, Payload: pld}
	msg := &Message{Request: req}

	err = s.write(msg)
	if err != nil {
		return nil, err
	}

	res := <-ch
	if res.Error != nil {
		return nil, res.Error
	}

	return res.Payload, nil
}

func (s *session) Handle(handlers map[string]Handler) {
	s.handlers = handlers
}

func (s *session) Close() (err error) {
	return s.connection.Close()
}

func (s *session) getNextID() (id uint32) {
	return atomic.AddUint32(&s.nextID, 1)
}

func (s *session) handleConnection() {
	defer s.Close()

	for {
		msg, err := s.read()
		if err != nil {
			break
		}

		switch {
		case msg.Request != nil:
			go s.handleRequest(msg.Request)
		case msg.Response != nil:
			go s.handleResponse(msg.Response)
		}
	}

	// end inflight method calls
	for id, ch := range s.inflight {
		delete(s.inflight, id)
		ch <- &ResDisconnected
	}
}

func (s *session) handleResponse(res *Response) {
	ch, ok := s.inflight[res.Id]
	if !ok {
		return
	}

	delete(s.inflight, res.Id)
	ch <- res
}

func (s *session) handleRequest(req *Request) {
	handler, ok := s.handlers[req.Method]
	if !ok {
		s.write(&MsgNoSuchMethod)
		return
	}

	res := &Response{Id: req.Id}

	var err error
	res.Payload, err = handler(req.Payload)
	if err != nil {
		res.Error = &Error{0, err.Error()}
	}

	msg := &Message{Response: res}
	err = s.write(msg)
	if err != nil {
		log.Println(err)
	}
}

func (s *session) read() (msg *Message, err error) {
	var msgSize int64
	err = binary.Read(s.connection, binary.LittleEndian, &msgSize)
	if err != nil {
		return nil, err
	}

	// TODO: The `msgSize > MaxMessageSize` test is used to avoid large `msgSize`.
	// An extremely large value for `msgSize` can cause a memory error and crash.
	// Find out why sometimes we receive extremely high values for `msgSize`.
	// Anyways, this test can still help protect servers from DOS attacks.
	if msgSize < 0 || msgSize > MaxMessageSize {
		return nil, &ErrReadFailed
	}

	buffer := make([]byte, msgSize)
	toRead := buffer[:]

	for len(toRead) > 0 {
		n, err := s.connection.Read(toRead)
		if err != nil {
			return nil, err
		}

		toRead = toRead[n:]
	}

	msg = &Message{}
	err = proto.Unmarshal(buffer, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *session) write(msg *Message) (err error) {
	buffer, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	s.writeMutx.Lock()
	defer s.writeMutx.Unlock()

	msgSize := int64(len(buffer))
	err = binary.Write(s.connection, binary.LittleEndian, msgSize)
	if err != nil {
		return err
	}

	toWrite := buffer[:]
	for len(toWrite) > 0 {
		n, err := s.connection.Write(toWrite)
		if err != nil {
			return err
		}

		toWrite = toWrite[n:]
	}

	return nil
}
