package main

import (
	"reflect"

	"github.com/meteorhacks/simple-rpc-go"
)

func main() {
	c := srpc.NewClient("localhost:12345")
	p := make([]byte, 1024, 1024)

	if err := c.Connect(); err != nil {
		panic(err)
	}

	for {
		o, err := c.Call("echo", p)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(p, o) {
			panic("invalid response")
		}
	}
}
