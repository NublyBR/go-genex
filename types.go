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
	String() string
	Complexity() int

	iterate() *iterator
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
