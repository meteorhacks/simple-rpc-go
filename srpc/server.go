package srpc

import (
	"net"
)

// Server is a server only interface to simple-rpc session
// It is there for those who are familiar with client-server model
type Server interface {
	SetHandler(name string, fn Handler)
	Listen() (err error)
	Close() (err error)
}

type server struct {
	address  string
	handlers map[string]Handler
	listener net.Listener
}

// NewServer creates a new simple-rpc server
func NewServer(address string) (S Server) {
	return &server{
		address:  address,
		handlers: make(map[string]Handler),
	}
}

func (s *server) SetHandler(name string, fn Handler) {
	s.handlers[name] = fn
}

func (s *server) Listen() (err error) {
	ch, l, err := Serve(s.address)
	if err != nil {
		return err
	}

	s.listener = l
	for ss := range ch {
		ss.Handle(s.handlers)
	}

	return nil
}

func (s *server) Close() (err error) {
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}
