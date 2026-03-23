package genex

import (
	"bytes"
	"math/big"
	"sort"
)

type Charset struct {
	chrs []byte
	repr string
	rng  RNG
}

func (g *Charset) Complexity() int {
	return 1
}

func (g *Charset) Count() *big.Int {
	return big.NewInt(int64(len(g.chrs)))
}

func (g *Charset) Bounds() (int, int) {
	return 1, 1
}

func (g *Charset) Iterate() *Iterator {
	ln := len(g.chrs)
	idx := 0

	return &Iterator{
		get: func(w *bytes.Buffer) {
			if idx < ln {
				w.WriteByte(g.chrs[idx])
			}
		},
		next: func() bool {
			idx++
			return idx < ln
		},
		reset: func() {
			idx = 0
		},
	}
}

func (g *Charset) Sample(w *bytes.Buffer) {
	w.WriteByte(
		g.chrs[g.rng()%int64(len(g.chrs))],
	)
}

func (g *Charset) String() string {
	return g.repr
}

func NewCharset(c ...byte) Generator {
	if len(c) == 0 {
		return NewFixed(nil)
	}

	if len(c) == 1 || (len(c) == 2 && c[0] == c[1]) {
		return NewFixed(c[:1])
	}

	if len(c)%2 == 1 {
		c = append(c, c[len(c)-1])
	}

	count := 0

	for i := 0; i < len(c); i += 2 {
		start := &c[i]
		end := &c[i+1]

		if *end < *start {
			*start, *end = *end, *start
		}

		count += int(*end - *start)
	}

	expand := make([]byte, 0, count)

	for i := 0; i < len(c); i += 2 {
		start := c[i]
		end := c[i+1]

		for j := start; j <= end; j++ {
			expand = append(expand, j)
		}
	}

	sort.Slice(expand, func(i, j int) bool {
		return expand[i] < expand[j]
	})

	repr := bytes.NewBuffer(make([]byte, 0, len(expand)*2+20))
	repr.WriteString("\033[32m[\033[0m")

	pushrepr := func(a, b byte) {
		if a == b {
			writeSpecial(repr, a)
			return
		}

		writeSpecial(repr, a)
		repr.WriteString("\033[32m-\033[0m")
		writeSpecial(repr, b)
	}

	first := expand[0]
	j := 0

	for i := range len(expand) - 1 {
		if expand[i] != expand[i+1] {
			j++
			expand[j] = expand[i+1]

			if expand[j-1] != expand[j]-1 {
				pushrepr(first, expand[j-1])
				first = expand[j]
			}
		}
	}

	pushrepr(first, expand[len(expand)-1])
	repr.WriteString("\033[32m]\033[0m")

	expand = expand[:j]

	switch len(expand) {
	case 0:
		return NewFixed(nil)
	case 1:
		return NewFixed(expand)
	default:
		return &Charset{
			chrs: expand,
			repr: repr.String(),
			rng:  FastRand,
		}
	}

}
