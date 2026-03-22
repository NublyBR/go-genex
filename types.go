package genex

import (
	"bytes"
	"math/big"
)

type Channel <-chan string

type Generator interface {
	Count() *big.Int
	Bounds() (int, int)
	Iterate() *Iterator
	Sample(*bytes.Buffer)
	String() string
	Complexity() int
}

type RNG func() uint64

var _ = []Generator{
	&Concat{},
	&Fixed{},
	&Charset{},
	&Choice{},
	&Repeat{},
	&Numeric{},
}
