package srpc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"

	"github.com/golang/protobuf/proto"
)

const (
	MsgSizeSize = 8
)

var (
	ErrBytesRead  = errors.New("number of bytes read is invalid")
	ErrBytesWrite = errors.New("number of bytes written is invalid")
	ErrMessageSz  = errors.New("message size is invalid")
	ErrNoMethod   = errors.New("method not found on server")
)

func readData(c net.Conn, pb proto.Message) (err error) {
	sbytes := make([]byte, MsgSizeSize, MsgSizeSize)
	sbuff := bytes.NewReader(sbytes)

	n, err := c.Read(sbytes)
	if err != nil {
		return err
	} else if n != MsgSizeSize {
		return ErrBytesRead
	}

	var sz int64
	err = binary.Read(sbuff, binary.BigEndian, &sz)
	if err != nil {
		return err
	}

	if sz < 0 {
		return err
	}

	dbytes := make([]byte, sz, sz)
	n, err = c.Read(dbytes)
	if err != nil {
		return err
	} else if int64(n) != sz {
		return ErrBytesRead
	}

	return proto.Unmarshal(dbytes, pb)
}

func writeData(c net.Conn, pb proto.Message) (err error) {
	dbytes, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	sz := int64(len(dbytes))
	sbuff := new(bytes.Buffer)
	err = binary.Write(sbuff, binary.BigEndian, sz)
	if err != nil {
		return err
	}

	sbytes := sbuff.Bytes()
	n, err := c.Write(sbytes)
	if n != MsgSizeSize {
		return ErrBytesWrite
	}

	n, err = c.Write(dbytes)
	if int64(n) != sz {
		return ErrBytesWrite
	}

	return nil
}
