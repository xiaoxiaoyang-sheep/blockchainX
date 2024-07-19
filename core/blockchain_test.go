package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(RandomBlock(0))
	assert.Nil(t, err)
	return bc
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockchainWithGenesis(t)

	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlockWithSignature(t, uint32(i+1))
		assert.Nil(t, bc.AddBlock(block))
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)

	assert.NotNil(t, bc.AddBlock(RandomBlock(100)))
}

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := NewBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}
