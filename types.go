package genex

import (
	"bytes"
	"iter"
	"math/big"
)

type Channel <-chan string

type Generator interface {
	Count() *big.Int
	Bounds() (int, int)
	Iterate() iter.Seq[[]byte]
	Sample(*bytes.Buffer)
	Index(*bytes.Buffer, *big.Int)
	String() string
	Complexity() int

	iterate() *iterator
	export() any
}

type RNG func() int64

var _ = []Generator{
	&Concat{},
	&Fixed{},
	&Charset{},
	&Choice{},
	&Repeat{},
	&Numeric{},
}
