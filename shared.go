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
	MsgSizeMax  = 1024 * 1024 * 16
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

	if sz < 0 || sz > MsgSizeMax {
		return ErrMessageSz
	}

	dbytes := make([]byte, sz, sz)
	toRead := dbytes[:]

	for len(toRead) > 0 {
		n, err = c.Read(toRead)
		if err != nil {
			return err
		}

		toRead = toRead[n:]
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

	toWrite := dbytes[:]
	for len(toWrite) > 0 {
		n, err = c.Write(toWrite)
		if err != nil {
			return err
		}

		toWrite = toWrite[n:]
	}

	return nil
}
