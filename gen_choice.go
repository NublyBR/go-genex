package genex

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

type Choice struct {
	items []Generator

	count    *big.Int
	min, max int
	rng      RNG
}

func (g *Choice) Complexity() int {
	ret := 1
	for _, item := range g.items {
		ret += item.Complexity()
	}
	return ret
}

func (g *Choice) Count() *big.Int {
	return new(big.Int).Set(g.count)
}

func (g *Choice) Bounds() (int, int) {
	return g.min, g.max
}

func (g *Choice) Iterate() *Iterator {
	ln := len(g.items)
	its := make([]*Iterator, ln)
	idx := 0

	for i, gen := range g.items {
		its[i] = gen.Iterate()
	}

	return &Iterator{
		get: func(w *bytes.Buffer) {
			if idx < ln {
				its[idx].get(w)
			}
		},
		next: func() bool {
			if st := its[idx].next(); !st {
				its[idx].reset()

				if idx++; idx >= ln {
					return false
				}
			}
			return true
		},
		reset: func() {
			for _, it := range its {
				it.reset()
			}
			idx = 0
		},
	}
}

func (g *Choice) Sample(w *bytes.Buffer) {
	g.items[g.rng()%int64(len(g.items))].Sample(w)
}

func (g *Choice) String() string {
	vars := make([]string, len(g.items))
	for i, cur := range g.items {
		vars[i] = cur.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(vars, "|"))
}

func NewChoice(o ...Generator) Generator {
	switch len(o) {
	case 0:
		return NewFixed(nil)
	case 1:
		return o[0]
	}

	all := true
	chr := make([]byte, 0, 2*len(o))
	for _, gen := range o {
		if f, ok := gen.(*Fixed); ok && len(f.text) == 1 {
			chr = append(chr, f.text[0], f.text[0])
			continue
		}

		all = false
		break
	}
	if all {
		return NewCharset(chr...)
	}

	for {
		retry := false
		flat := make([]Generator, 0, len(o))
		nore := map[string]bool{}

		for _, item := range o {
			switch cast := item.(type) {
			case *Choice:
				flat = append(flat, cast.items...)
				retry = true
			case *Fixed:
				txt := hex.EncodeToString(cast.text)
				if !nore[txt] {
					nore[txt] = true
					flat = append(flat, cast)
				}
			default:
				flat = append(flat, item)
			}
		}

		o = flat

		if !retry {
			break
		}
	}

	switch len(o) {
	case 0:
		return NewFixed(nil)
	case 1:
		return o[0]
	default:
		c := big.NewInt(0)
		for _, elem := range o {
			c.Add(c, elem.Count())
		}

		min, max := o[0].Bounds()
		for _, elem := range o[1:] {
			cmin, cmax := elem.Bounds()
			if cmin < min {
				min = cmin
			}
			if cmax > max {
				max = cmax
			}
		}

		return &Choice{
			items: o,

			count: c,
			min:   min,
			max:   max,
			rng:   FastRand,
		}
	}

}

func NewChoiceFixed(o ...[]byte) Generator {
	og := make([]Generator, len(o))
	for i := range og {
		og[i] = NewFixed(o[i])
	}
	return NewChoice(og...)
}
