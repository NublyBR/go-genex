package genex

import (
	"bytes"
	"fmt"
	"math/big"
)

type Fixed struct {
	text []byte
}

func (g *Fixed) Complexity() int {
	return 1
}

func (g *Fixed) Count() *big.Int {
	return big.NewInt(1)
}

func (g *Fixed) Bounds() (int, int) {
	return len(g.text), len(g.text)
}

func (g *Fixed) Iterate() *Iterator {
	return &Iterator{
		get:   func(w *bytes.Buffer) { w.Write(g.text) },
		next:  func() bool { return false },
		reset: func() {},
	}
}

func (g *Fixed) Sample(w *bytes.Buffer) {
	w.Write(g.text)
}

func (g *Fixed) String() string {
	esc := fmt.Sprintf("%q", g.text)
	return fmt.Sprintf("\033[33m%s\033[0m", esc[1:len(esc)-1])
}

func NewFixed(s []byte) Generator {
	return &Fixed{text: s}
}
