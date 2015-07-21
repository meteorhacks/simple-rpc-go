package srpc

// Client is a client only interface to simple-rpc session
// It is there for those who are familiar with client-server model
type Client interface {
	Connect() (err error)
	Call(name string, pld []byte) (out []byte, err error)
	Close() (err error)
}

type client struct {
	address string
	session Session
}

// NewClient creates a new simple-rpc client
func NewClient(address string) (C Client) {
	return &client{address: address}
}

func (c *client) Connect() (err error) {
	c.session, err = Dial(c.address)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) Call(name string, pld []byte) (out []byte, err error) {
	return c.session.Call(name, pld)
}

func (c *client) Close() (err error) {
	if c.session != nil {
		return c.session.Close()
	}

	return nil
}
