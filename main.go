package main

import (
	"time"

	"github.com/xiaoxiaoyang-sheep/blockchainX/network"
)

func main() {
	trLocal := network.NewLocalTransport("LOCAL")
	trRemote := network.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			trRemote.SendMessage(trLocal.Addr(), []byte("hello world"))
			time.Sleep(1 * time.Second)
		}
	}()

	ops := network.ServerOps{
		Transports: []network.Transport{trLocal},
	}

	s := network.NewServer(ops)
	s.Start()

}
