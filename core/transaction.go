package core

import (
	"fmt"

	"github.com/xiaoxiaoyang-sheep/blockchainX/crypto"
)

type Transaction struct {
	Date      []byte
	PublicKey crypto.PublicKey
	Signature *crypto.Signature
}

func (tx *Transaction) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(tx.Date)
	if err != nil {
		return err
	}

	tx.PublicKey = privKey.PublicKey()
	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	if !tx.Signature.Verify(tx.PublicKey, tx.Date) {
		return fmt.Errorf("invalid transaction signature")
	}

	return nil
}
