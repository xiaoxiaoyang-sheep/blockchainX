package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack(128)

	s.Push(1)
	s.Push(2)

	value := s.Pop()
	assert.Equal(t, value, 1)

	value = s.Pop()
	assert.Equal(t, value, 2)
}

func TestVM(t *testing.T) {
	// data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	// vm := NewVM(data)
	// assert.Nil(t, vm.Run())

	// result := vm.stack.Pop()
	// assert.Equal(t, 3, result)

	// data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d}
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0e}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().(int)
	// assert.Equal(t, "FOO", string(result))
	assert.Equal(t, 1, result)
}
