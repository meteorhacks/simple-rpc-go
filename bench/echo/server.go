package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/meteorhacks/simple-rpc-go"
)

var (
	counter = 0
	mutex   = sync.Mutex{}
)

func main() {
	go logRate()
	s := srpc.NewServer(":12345")
	s.SetHandler("echo", handler)
	fmt.Println(s.Listen())
}

func handler(req []byte) (res []byte, err error) {
	mutex.Lock()
	counter++
	mutex.Unlock()

	return req, nil
}

func logRate() {
	for {
		time.Sleep(time.Second)
		fmt.Println(counter)
		mutex.Lock()
		counter = 0
		mutex.Unlock()
	}
}
