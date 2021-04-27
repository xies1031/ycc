package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/panjf2000/gnet"
)

type server struct {
	*gnet.EventServer
	connectedSockets sync.Map
	ch               chan []byte
}

func (s *server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Push server is listening on %s (multi-cores: %t, loops: %d)", srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (s *server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("Socket with addr: %s has been opened...\n", c.RemoteAddr().String())
	s.connectedSockets.Store(c.RemoteAddr().String(), c)
	return
}

func (s *server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Printf("Socket with addr: %s is closing...\n", c.RemoteAddr().String())
	s.connectedSockets.Delete(c.RemoteAddr().String())
	return
}

// func (s *server) Tick() (delay time.Duration, action gnet.Action) {
// 	log.Println("It's time to push data to clients!!!")
// 	s.connectedSockets.Range(func(key, value interface{}) bool {
// 		addr := key.(string)
// 		c := value.(gnet.Conn)
// 		c.AsyncWrite([]byte(fmt.Sprintf("heart beating to %s\n", addr)))
// 		return true
// 	})
// 	delay = s.tick
// 	return
// }

func (s *server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	s.ch <- frame
	return
}

func (s *server) forward() {
	for {
		data := <-s.ch
		s.connectedSockets.Range(func(key, value interface{}) bool {
			c := value.(gnet.Conn)
			c.AsyncWrite(data)
			return true
		})
	}
}

func main() {
	var port int
	var multicore bool

	flag.IntVar(&port, "port", 9000, "server port")
	flag.BoolVar(&multicore, "multicore", true, "multicore")
	flag.Parse()
	s := &server{ch: make(chan []byte, 16)}
	go s.forward()
	log.Fatal(gnet.Serve(s, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(multicore)))
}
