package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/xiaoxiaoyang-sheep/blockchainX/core"
)

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
	MessageTypeStatus    MessageType = 0x4
	MessageTypeGetStatus MessageType = 0x5
)

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{
		Header: t,
		Data:   data,
	}
}

func (m *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(m)
	return buf.Bytes()
}

type DecodedMessage struct {
	From NetAddr
	Data any
}

type RPCDecodedFunc func(RPC) (*DecodedMessage, error)

func DecodeRPCDecodedFunc(rpc RPC) (*DecodedMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode the message from %s: %v", rpc.From, err)
	}

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageTypeGetStatus:
		return &DecodedMessage{
			From: rpc.From,
			Data: &GetStatusMessage{},
		}, nil
	case MessageTypeStatus:
		statusMessage := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(statusMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: statusMessage,
		}, nil
	default:
		return nil, fmt.Errorf("invalid message header %x", msg.Header)
	}
}

type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
