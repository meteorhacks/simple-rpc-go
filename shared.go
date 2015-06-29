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
	sbuff := bytes.NewBuffer(sbytes)

	n, err := c.Read(sbytes)
	if err != nil {
		return err
	} else if n != MsgSizeSize {
		return ErrBytesRead
	}

	sz, err := binary.ReadVarint(sbuff)
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
	sbytes := make([]byte, MsgSizeSize, MsgSizeSize)
	dbytes, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	sz := int64(len(dbytes))
	binary.PutVarint(sbytes, sz)

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
