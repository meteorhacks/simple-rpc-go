package main

import (
	"log"
	"net/http"

	_ "net/http/pprof"

	"github.com/meteorhacks/simple-rpc-go/srpc"
)

const (
	concurrency = 8
)

func main() {
	for i := 0; i < concurrency; i++ {
		go startClient()
	}

	startPPROF()
}

func startClient() {
	c := srpc.NewClient("localhost:12345")
	p := make([]byte, 1024, 1024)

	if err := c.Connect(); err != nil {
		panic(err)
	}

	for {
		if _, err := c.Call("echo", p); err != nil {
			panic(err)
		}
	}
}

func startPPROF() {
	log.Println("PPROF:  listening on :6061")
	log.Println(http.ListenAndServe(":6061", nil))
}
