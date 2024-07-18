package network

import (
	"fmt"
	"time"
)

type ServerOps struct {
	Transports []Transport
}

type Server struct {
	ServerOps
	rpcCh  chan Rpc
	quitCh chan struct{}
}

func NewServer(ops ServerOps) *Server {
	return &Server{
		ServerOps: ops,
		rpcCh:     make(chan Rpc),
		quitCh:    make(chan struct{}),
	}
}

func (s *Server) Start() {
	s.initTransports()
	trick := time.NewTicker(5 * time.Second)

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			fmt.Printf("%v \n", rpc.Payload)
		case <-s.quitCh:
			break free
		case <-trick.C:
			fmt.Println("do stuff every x seconds")
		}
	}
}

func (s *Server) initTransports() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}
