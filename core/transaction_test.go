package core

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiaoxiaoyang-sheep/blockchainX/crypto"
)

func TestSignTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Date: []byte("foo"),
	}
	assert.Nil(t, tx.Sign(privKey))
}

func TestVerifyTransaction(t *testing.T) {
	tx := randomTxWithSignature(t)
	assert.Nil(t, tx.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()

	assert.NotNil(t, tx.Verify())
}

func TestTxEncodeDecode(t *testing.T) {
	tx := randomTxWithSignature(t)
	buf := &bytes.Buffer{}
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecoded := new(Transaction)
	assert.Nil(t, txDecoded.Decode(NewGobTxDecoder(buf)))

	assert.Equal(t, tx, txDecoded)
}

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Date: []byte("foo"),
	}
	assert.Nil(t, tx.Sign(privKey))
	return tx
}
