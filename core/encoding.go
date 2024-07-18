package core

import "io"

type Encoder[T any] interface {
	Encoder(io.Writer, T) error
}

type Decoder[T any] interface {
	Decoder(io.Reader, T) error
}
