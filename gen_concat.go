package genex

import (
	"bytes"
	"fmt"
	"iter"
	"math/big"
	"strings"
)

type Concat struct {
	items []Generator

	count    *big.Int
	min, max int
}

func (g *Concat) Complexity() int {
	ret := 1
	for _, item := range g.items {
		ret += item.Complexity()
	}
	return ret
}

func (g *Concat) Count() *big.Int {
	return new(big.Int).Set(g.count)
}

func (g *Concat) Bounds() (int, int) {
	return g.min, g.max
}

func (g *Concat) Iterate() iter.Seq[[]byte] {
	return makeSeq(g)
}

func (g *Concat) iterate() *iterator {
	ln1 := len(g.items) - 1
	its := make([]*iterator, len(g.items))

	for i, gen := range g.items {
		its[i] = gen.iterate()
	}

	return &iterator{
		get: func(w *bytes.Buffer) {
			for _, it := range its {
				it.get(w)
			}
		},
		next: func() bool {
			for i := ln1; i >= 0; i-- {
				if its[i].next() {
					break
				}

				its[i].reset()

				if i == 0 {
					return false
				}
			}
			return true
		},
		reset: func() {
			for _, it := range its {
				it.reset()
			}
		},
	}
}

func (g *Concat) export() any {
	ret := make([]any, 0, len(g.items))
	for _, item := range g.items {
		ret = append(ret, item.export())
	}
	return ret
}

func (g *Concat) Sample(w *bytes.Buffer) {
	for _, opt := range g.items {
		opt.Sample(w)
	}
}

func (g *Concat) Index(w *bytes.Buffer, idx *big.Int) {
	for _, opt := range g.items {
		opt.Index(w, idx)
	}
}

func (g *Concat) String() string {
	vars := make([]string, len(g.items))
	for i, cur := range g.items {
		vars[i] = cur.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(vars, ""))
}

func NewConcat(o ...Generator) Generator {
	switch len(o) {
	case 0:
		return NewFixed(nil)
	case 1:
		return o[0]
	}

	tryFlat := false

	for i := range len(o) {
		if _, ok := o[i].(*Concat); ok {
			tryFlat = true
			break
		}

		if i == len(o)-1 {
			break
		}

		if _, ok := o[i].(Fixed); ok {
			if _, ok := o[i+1].(Fixed); ok {
				tryFlat = true
				break
			}
		}
	}

	if tryFlat {
		for {
			retry := false
			flat := make([]Generator, 0, len(o))

			for _, item := range o {
				switch cast := item.(type) {
				case *Concat:
					flat = append(flat, cast.items...)
					retry = true
				case Fixed:
					if len(flat) > 0 {
						if fixed, ok := flat[len(flat)-1].(Fixed); ok {
							nw := make([]byte, len(fixed)+len(cast))
							copy(nw, fixed)
							copy(nw[len(fixed):], cast)
							flat[len(flat)-1] = Fixed(nw)
							continue
						}
					}
					if len(cast) > 0 {
						flat = append(flat, cast)
					}
				default:
					flat = append(flat, cast)
				}
			}

			o = flat

			if !retry {
				break
			}
		}
	}

	switch len(o) {
	case 0:
		return NewFixed(nil)
	case 1:
		return o[0]
	default:
		c := big.NewInt(1)
		for _, elem := range o {
			c.Mul(c, elem.Count())
		}

		min, max := 0, 0
		for _, elem := range o {
			cmin, cmax := elem.Bounds()
			min += cmin
			max += cmax
		}

		return &Concat{
			items: o,

			count: c,
			min:   min,
			max:   max,
		}
	}
}
