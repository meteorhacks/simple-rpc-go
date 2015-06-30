package srpc

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

type Client interface {
	Connect() (err error)
	Call(name string, pld []byte) (out []byte, err error)
}

type client struct {
	nextId  int64
	address string
	mcalls  map[int64]chan *Response
	conn    net.Conn
	idMutx  sync.Mutex
	wMutx   sync.Mutex
}

func NewClient(address string) (C Client) {
	c := client{
		address: address,
		mcalls:  make(map[int64]chan *Response),
		idMutx:  sync.Mutex{},
		wMutx:   sync.Mutex{},
	}

	return &c
}

func (c *client) Connect() (err error) {
	c.conn, err = net.Dial("tcp", c.address)
	if err != nil {
		return err
	}

	go c.handleConnection()

	return err
}

func (c *client) Call(name string, pld []byte) (out []byte, err error) {
	id := c.getID()
	ch := make(chan *Response)
	c.mcalls[id] = ch

	req := &Request{Id: id, Method: name, Payload: pld}
	c.wMutx.Lock()
	err = writeData(c.conn, req)
	c.wMutx.Unlock()

	if err != nil {
		return nil, err
	}

	res := <-ch

	if res.Error != "" {
		// TODO: improve errors
		return nil, errors.New(res.Error)
	}

	return res.Payload, nil
}

func (c *client) handleConnection() {
	defer c.conn.Close()

	for {
		res := &Response{}
		if err := readData(c.conn, res); err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error reading message", err)
			break
		}

		go c.handleResponse(res)
	}
}

func (c *client) handleResponse(res *Response) {
	ch, ok := c.mcalls[res.Id]
	if !ok {
		log.Println("Invalid response ID")
		return
	}

	delete(c.mcalls, res.Id)
	ch <- res
}

func (c *client) getID() (id int64) {
	c.idMutx.Lock()
	id = c.nextId
	c.nextId++
	c.idMutx.Unlock()
	return id
}
