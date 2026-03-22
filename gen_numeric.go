package genex

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
)

type Numeric struct {
	base             int
	start, end, step uint64
	pad              bool
	buf              []byte

	count    uint64
	min, max int
	rng      RNG
}

func (g *Numeric) Complexity() int {
	return 1
}

func (g *Numeric) Count() *big.Int {
	return new(big.Int).SetUint64(g.count)
}

func (g *Numeric) Bounds() (int, int) {
	return g.min, g.max
}

func (g *Numeric) Iterate() *Iterator {
	n := g.start
	buf := make([]byte, g.max)

	return &Iterator{
		get: func(w *bytes.Buffer) {
			ptr := numEncode(buf, n, g.base)
			if g.pad {
				w.Write(buf)
			} else {
				w.Write(ptr)
			}
		},
		next: func() bool {
			n += g.step
			if n > g.end {
				n -= g.step
				return false
			}
			return true
		},
		reset: func() {
			n = g.start
		},
	}
}

func (g *Numeric) Sample(w *bytes.Buffer) {
	n := g.start + (g.rng()%g.count)*g.step
	if g.pad {
		for i := range g.buf {
			g.buf[i] = numBase[0]
		}
	}

	ptr := numEncode(g.buf, n, g.base)
	if g.pad {
		w.Write(g.buf)
	} else {
		w.Write(ptr)
	}
}

func (g *Numeric) String() string {
	res := new(strings.Builder)

	res.WriteString("<\033[32m")
	if g.base != 10 {
		fmt.Fprintf(res, "%d:", g.base)
	}
	switch {
	case g.start == 0 && g.step == 1:
		fmt.Fprintf(res, "%d", g.end)
	case g.start == 0 && g.step != 1:
		fmt.Fprintf(res, "/%d/%d", g.end, g.step)
	case g.step == 1:
		fmt.Fprintf(res, "%d/%d", g.start, g.end)
	default:
		fmt.Fprintf(res, "%d/%d/%d", g.start, g.end, g.step)
	}
	res.WriteString("\033[0m>")

	return res.String()
}

func NewNumeric(base int, start, end, step uint64, pad bool) Generator {
	if start == 0 && end == 0xffff_ffff_ffff_ffff {
		// Due to the current implementation, the full uint64 range causes a panic due to an overflow.
		end--
	}

	count := uint64((end-start)/step + 1)
	min := numSize(start, base)
	max := numSize(start+(count-1)*step, base)

	return &Numeric{
		base:  base,
		start: start,
		end:   end,
		step:  step,
		pad:   pad,
		buf:   make([]byte, max),

		count: count,
		min:   min,
		max:   max,
		rng:   FastRand,
	}
}
