package genex

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

type Charset struct {
	chrs []byte
	repr [][2]byte
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
		g.chrs[g.rng()%uint64(len(g.chrs))],
	)
}

func (g *Charset) String() string {
	vars := make([]string, len(g.repr))
	for i, r := range g.repr {
		if r[0] == r[1] {
			vars[i] = string(r[0])
		} else {
			vars[i] = fmt.Sprintf("%c-%c", r[0], r[1])
		}
	}
	return fmt.Sprintf("[\033[32m%s\033[0m]", strings.Join(vars, ""))
}

func NewCharset(c ...byte) Generator {
	if len(c) == 0 {
		return NewFixed(nil)
	}

	if len(c)%2 == 1 {
		c = append(c, c[len(c)-1])
	}

	if len(c) == 1 && c[0] == c[1] {
		return NewFixed([]byte{c[0]})
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
	repr := make([][2]byte, 0, len(c)/2)

	for i := 0; i < len(c); i += 2 {
		start := c[i]
		end := c[i+1]

		for j := start; j <= end; j++ {
			expand = append(expand, j)
		}

		repr = append(repr, [2]byte{start, end})
	}

	sort.Slice(expand, func(i, j int) bool {
		return expand[i] < expand[j]
	})

	pruned := make([]byte, 0, len(expand))
	pruned = append(pruned, expand[0])

	for i := range len(expand) - 1 {
		if expand[i] != expand[i+1] {
			pruned = append(pruned, expand[i+1])
		}
	}

	switch len(pruned) {
	case 0:
		return NewFixed(nil)
	case 1:
		return NewFixed(pruned)
	default:
		return &Charset{
			chrs: pruned,
			repr: repr,
			rng:  FastRand,
		}
	}

}
