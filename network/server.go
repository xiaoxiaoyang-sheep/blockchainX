package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/xiaoxiaoyang-sheep/blockchainX/core"
	"github.com/xiaoxiaoyang-sheep/blockchainX/crypto"
	"github.com/xiaoxiaoyang-sheep/blockchainX/types"
)

var defaultBlockTime = 5 * time.Second

type ServerOps struct {
	ID             string
	Transport      Transport
	Logger         log.Logger
	RPCDecodedFunc RPCDecodedFunc
	RPCProcessor   RPCProcessor
	Transports     []Transport
	BlockTime      time.Duration
	PrivateKey     *crypto.PrivateKey
}

type Server struct {
	ServerOps
	blockTime   time.Duration
	mempool     *TxPool
	chain       *core.Blockchain
	IsValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOps) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodedFunc == nil {
		opts.RPCDecodedFunc = DecodeRPCDecodedFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}

	chain, err := core.NewBlockchain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOps:   opts,
		blockTime:   opts.BlockTime,
		chain:       chain,
		mempool:     NewTxPool(1000),
		IsValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}),
	}

	// if we dont get any processor from the server options, we going to use
	// the sever as default
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	if s.IsValidator {
		go s.validatorLoop()
	}

	// for _, tr := range s.Transports {
	// 	s.sendGetStatusMessage(tr)
	// }
	return s, nil
}

func (s *Server) Start() {
	s.initTransports()

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodedFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err := s.Logger.Log("error", err); err != nil {
					s.Logger.Log("send get status error", err)
				}
			}
		case <-s.quitCh:
			break free
		}
	}

	s.Logger.Log("msg", "Server is shutting down")

}

func (s *Server) validatorLoop() {
	trick := time.NewTicker(s.blockTime)

	s.Logger.Log("msg", "Starting validator loop", "blockTime", s.blockTime)

	for {
		<-trick.C
		s.createNewBlock()
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From, t)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	}
	return nil
}

// TODO(@Yanko): Remoeve the logic from the main funciton to here
// Normally Transport which is our own transport should do the trick
func (s *Server) sendGetStatusMessage(tr Transport) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	// msg := NewMessage(MessageTypeGetStatus, buf.Bytes())

	return nil
}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return nil
		}
	}
	return nil
}

func (s *Server) processStatusMessage(from NetAddr, data *StatusMessage) error {

	fmt.Printf("received Getstatus response msg from %s => %+v\n", from, data)
	return nil
}

func (s *Server) processGetStatusMessage(from NetAddr, data *GetStatusMessage) error {
	fmt.Printf("received Getstatus msg from %s => %+v\n", from, data)

	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.ID,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return s.Transport.SendMessage(from, msg.Bytes())
}

func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	go s.broadcastBlock(b)
	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.mempool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	// s.Logger.Log(
	// 	"msg", "adding new tx to mempool",
	// 	"hash", hash,
	// 	"mempoolLength", s.mempool.PendingCount(),
	// )

	// broadcast this tx to peers
	go s.broadcoatTx(tx)

	s.mempool.Add(tx)

	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcoatTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
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

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// For now we are going to use all transactions that are in the mempool
	// Later on when we know the internal structure of our transaction
	// we will implement some kind of complexity funciton to determine how
	// many transcations can be included in a block
	txx := s.mempool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	// TODO(@Yanko): pending pool of tx should only reflect on validator nodes.
	// Right now "normal nodes" does not have their pending pool cleared.
	s.mempool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}

	b, _ := core.NewBlock(header, nil)

	return b
}
