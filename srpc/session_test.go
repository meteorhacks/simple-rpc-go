package srpc

import (
	"net"
	"reflect"
	"testing"
)

const (
	TestAddress  = ":12000"
	BenchAddress = ":12001"
)

func init() {
	ch, _, err := Serve(BenchAddress)
	if err != nil {
		panic(err)
	}

	handlers := make(map[string]Handler)
	handlers["echo"] = func(req []byte) (res []byte, err error) {
		return req, nil
	}

	go func() {
		for s := range ch {
			s.Handle(handlers)
		}
	}()
}

func TestServe(t *testing.T) {
	ch, l, err := Serve(TestAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		defer l.Close()
	}

	done := make(chan bool)

	go func() {
		_ = <-ch
		done <- true
	}()

	c, err := net.Dial("tcp", TestAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		defer c.Close()
	}

	<-done
}

func TestDial(t *testing.T) {
	ch, l, err := Serve(TestAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		defer l.Close()
	}

	done := make(chan bool)

	go func() {
		_ = <-ch
		done <- true
	}()

	s, err := Dial(TestAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		defer s.Close()
	}

	<-done
}

func TestHandler(t *testing.T) {
	ch, l, err := Serve(TestAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		defer l.Close()
	}

	handlers := make(map[string]Handler)
	handlers["echo"] = func(req []byte) (res []byte, err error) {
		return req, nil
	}

	go func() {
		s := <-ch
		s.Handle(handlers)
	}()

	s, err := Dial(TestAddress)
	if err != nil {
		t.Fatal(err)
	} else {
		defer s.Close()
	}

	pld := []byte{1, 2, 3}
	out, err := s.Call("echo", pld)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(pld, out) {
		t.Fatal("incorrect result")
	}
}

func BenchRoundtripSz(b *testing.B, sz int) {
	s, err := Dial(BenchAddress)
	if err != nil {
		b.Fatal(err)
	} else {
		defer s.Close()
	}

	pld := make([]byte, sz)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Call("echo", pld)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundtripSz1k(b *testing.B)  { BenchRoundtripSz(b, 1024) }
func BenchmarkRoundtripSz10k(b *testing.B) { BenchRoundtripSz(b, 10240) }

func BenchPRoundtripSz(b *testing.B, p int, sz int) {
	pld := make([]byte, sz)

	b.SetParallelism(p)
	b.RunParallel(func(pb *testing.PB) {
		s, err := Dial(BenchAddress)
		if err != nil {
			b.Fatal(err)
		} else {
			defer s.Close()
		}

		for pb.Next() {
			_, err := s.Call("echo", pld)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkP5RoundtripSz1k(b *testing.B)   { BenchPRoundtripSz(b, 5, 1024) }
func BenchmarkP5RoundtripSz10k(b *testing.B)  { BenchPRoundtripSz(b, 5, 10240) }
func BenchmarkP10RoundtripSz1k(b *testing.B)  { BenchPRoundtripSz(b, 5, 1024) }
func BenchmarkP10RoundtripSz10k(b *testing.B) { BenchPRoundtripSz(b, 5, 10240) }
