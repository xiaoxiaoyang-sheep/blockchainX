package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xiaoxiaoyang-sheep/blockchainX/core"
	"github.com/xiaoxiaoyang-sheep/blockchainX/crypto"
	"github.com/xiaoxiaoyang-sheep/blockchainX/network"
)

func main() {
	trLocal := network.NewLocalTransport("LOCAL")
	trRemoteA := network.NewLocalTransport("REMOTE_A")
	trRemoteB := network.NewLocalTransport("REMOTE_B")
	trRemoteC := network.NewLocalTransport("REMOTE_C")

	trLocal.Connect(trRemoteA)
	trRemoteA.Connect(trRemoteB)
	trRemoteB.Connect(trRemoteC)

	trRemoteA.Connect(trLocal)

	initRemoteServers([]network.Transport{trRemoteA, trRemoteB, trRemoteC})

	go func() {
		for {
			// trRemote.SendMessage(trLocal.Addr(), []byte("hello world"))
			if err := sendTransaction(trRemoteA, trLocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		time.Sleep(7 * time.Second)

		trLater := network.NewLocalTransport("Later_REMOTE")
		trRemoteC.Connect(trLater)
		laterServer := makeServer(string(trLater.Addr()), trLater, nil)

		go laterServer.Start()
	}()

	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", trLocal, &privKey)
	localServer.Start()

}

func initRemoteServers(trs []network.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, trs[i], nil)
		go s.Start()
	}
}

func makeServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
	opts := network.ServerOps{
		PrivateKey: pk,
		ID:         id,
		Transports: []network.Transport{tr},
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return nil
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())
}
