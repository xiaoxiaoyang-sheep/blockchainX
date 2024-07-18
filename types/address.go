package types

import (
	"encoding/hex"
	"fmt"
)

type Address [20]uint8

func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		msg := fmt.Sprintf("given bytes with length %d should be 20", len(b))
		panic(msg)
	}

	value := make([]uint8, 20)
	copy(value, b[:])

	return Address(value)
}

func (addr Address) ToSlice() []byte {
	b := make([]byte, 20)
	copy(b, b[:])
	return b
}

func (addr Address) String() string {
	return hex.EncodeToString(addr.ToSlice())
}
