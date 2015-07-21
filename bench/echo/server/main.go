package main

import (
	"log"
	"sync/atomic"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/meteorhacks/simple-rpc-go/srpc"
)

var (
	counter uint32
)

func main() {
	go startPPROF()
	go logRate()

	s := srpc.NewServer(":12345")
	s.SetHandler("echo", handler)
	log.Println(s.Listen())
}

func handler(req []byte) (res []byte, err error) {
	atomic.AddUint32(&counter, 1)
	return req, nil
}

func logRate() {
	for {
		time.Sleep(time.Second * 10)
		log.Println(counter/10, "req/s")
		atomic.StoreUint32(&counter, 0)
	}
}

func startPPROF() {
	log.Println("PPROF:  listening on :6060")
	log.Println(http.ListenAndServe(":6060", nil))
}
