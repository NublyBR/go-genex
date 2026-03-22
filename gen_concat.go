package genex

import (
	"bytes"
	"fmt"
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

func (g *Concat) Iterate() *Iterator {
	ln1 := len(g.items) - 1
	its := make([]*Iterator, len(g.items))

	for i, gen := range g.items {
		its[i] = gen.Iterate()
	}

	return &Iterator{
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

func (g *Concat) Sample(w *bytes.Buffer) {
	for _, opt := range g.items {
		opt.Sample(w)
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

	for {
		retry := false
		flat := make([]Generator, 0, len(o))

		for _, item := range o {
			switch cast := item.(type) {
			case *Concat:
				flat = append(flat, cast.items...)
				retry = true
			case *Fixed:
				if len(flat) > 0 {
					if fixed, ok := flat[len(flat)-1].(*Fixed); ok {
						nw := make([]byte, len(fixed.text)+len(cast.text))
						copy(nw, fixed.text)
						copy(nw[len(fixed.text):], cast.text)
						fixed.text = nw
						continue
					}
				}
				if len(cast.text) > 0 {
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
