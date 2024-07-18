package network

type NetAddr string

type Rpc struct {
	From    NetAddr
	Payload []byte
}

type Transport interface {
	Consume() <-chan Rpc
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	Addr() NetAddr
}
