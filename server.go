package srpc

import (
	"io"
	"log"
	"net"
	"sync"
)

// Method handler receives a request of []byte
// and responds to it with a []byte
type Handler func(req []byte) (res []byte, err error)

type Server interface {
	SetHandler(name string, fn Handler)
	Listen() (err error)
}

type server struct {
	address string
	methods map[string]Handler
	wMutx   sync.Mutex
}

func NewServer(address string) (S Server) {
	s := server{
		address: address,
		methods: make(map[string]Handler),
		wMutx:   sync.Mutex{},
	}

	return &s
}

// Set a method handler for name 'name'
func (s *server) SetHandler(name string, fn Handler) {
	s.methods[name] = fn
}

// Start listening on given address
func (s *server) Listen() (err error) {
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go s.handleConnection(c)
	}

	return nil
}

func (s *server) handleConnection(c net.Conn) {
	defer c.Close()

	for {
		req := &Request{}
		if err := readData(c, req); err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error reading message", err)
			break
		}

		go s.handleRequest(c, req)
	}
}

func (s *server) handleRequest(c net.Conn, req *Request) {
	res := &Response{Id: req.Id}
	handler, ok := s.methods[req.Method]
	if !ok {
		res.Error = ErrNoMethod.Error()
	} else {
		resData, err := handler(req.Payload)
		if err != nil {
			res.Error = err.Error()
		} else if resData != nil {
			res.Payload = resData
		}
	}

	s.wMutx.Lock()
	err := writeData(c, res)
	s.wMutx.Unlock()

	if err != nil {
		log.Println("Error writing message", err)
	}
}
