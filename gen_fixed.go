package genex

import (
	"bytes"
	"math/big"
)

type Fixed []byte

func (g Fixed) Complexity() int {
	return 1
}

func (g Fixed) Count() *big.Int {
	return big.NewInt(1)
}

func (g Fixed) Bounds() (int, int) {
	return len(g), len(g)
}

func (g Fixed) Iterate() *Iterator {
	return &Iterator{
		get:   func(w *bytes.Buffer) { w.Write(g) },
		next:  func() bool { return false },
		reset: func() {},
	}
}

func (g Fixed) Sample(w *bytes.Buffer) {
	w.Write(g)
}

func (g Fixed) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, len(g)*3))
	buf.WriteString("\033[33m")
	for _, b := range g {
		writeSpecial(buf, b)
	}
	buf.WriteString("\033[0m")

	return buf.String()
}

func NewFixed(s []byte) Generator {
	return Fixed(s)
}
