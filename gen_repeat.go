package genex

import (
	"bytes"
	"fmt"
	"math/big"
)

type Repeat struct {
	inner      Generator
	rmin, rmax int

	count    *big.Int
	min, max int
	rng      RNG
}

func (g *Repeat) Complexity() int {
	return 1 + g.inner.Complexity()
}

func (g *Repeat) Count() *big.Int {
	return new(big.Int).Set(g.count)
}

func (g *Repeat) Bounds() (int, int) {
	return g.min, g.max
}

func (g *Repeat) Iterate() *Iterator {
	ln := g.rmin
	its := make([]*Iterator, g.rmax)

	for i := range g.rmax {
		its[i] = g.inner.Iterate()
	}

	return &Iterator{
		get: func(w *bytes.Buffer) {
			for i := range min(ln, g.rmax) {
				its[i].get(w)
			}
		},
		next: func() bool {
			if ln == 0 {
				ln++
				return ln <= g.rmax
			}
			for i := ln - 1; i >= 0; i-- {
				if its[i].next() {
					break
				}

				its[i].reset()

				if i == 0 {
					if ln++; ln > g.rmax {
						return false
					}
				}
			}
			return true
		},
		reset: func() {
			for _, it := range its {
				it.reset()
			}
			ln = g.rmin
		},
	}
}

func (g *Repeat) Sample(w *bytes.Buffer) {
	count := g.rmin + int(g.rng()%uint64(g.rmax-g.rmin+1))

	for range count {
		g.inner.Sample(w)
	}
}

func (g *Repeat) String() string {
	if g.rmin == 0 && g.rmax == 1 {
		return fmt.Sprintf("%s?", g.inner)
	}
	if g.rmin == g.rmax {
		return fmt.Sprintf("%s{\033[35m%d\033[0m}", g.inner, g.rmin)
	}
	return fmt.Sprintf("%s{\033[35m%d\033[0m,\033[35m%d\033[0m}", g.inner, g.rmin, g.rmax)
}

func NewRepeat(g Generator, rmin, rmax int) Generator {
	if rmin < 0 {
		rmin = 0
	}

	if rmax < 0 {
		rmax = 0
	}

	if rmin > rmax {
		rmin, rmax = rmax, rmin
	}

	if rmin == 1 && rmax == 1 {
		return g
	}

	if rmin == 0 && rmax == 0 {
		return NewFixed(nil)
	}

	if f, ok := g.(*Fixed); ok {
		if rmin == rmax {
			res := make([]byte, 0, len(f.text)*rmin)
			for range rmin {
				res = append(res, f.text...)
			}
			return NewFixed(res)
		}
		if len(f.text) == 0 {
			return f
		}
	}

	c := big.NewInt(0)
	inner := g.Count()
	for i := rmin; i <= rmax; i++ {
		cur := big.NewInt(int64(i))
		cur.Exp(inner, cur, nil) // cur = inner.Count() ^ i
		c.Add(c, cur)
	}

	min, max := g.Bounds()
	min *= rmin
	max *= rmax

	return &Repeat{
		inner: g,
		rmin:  rmin,
		rmax:  rmax,

		count: c,
		min:   min,
		max:   max,
		rng:   FastRand,
	}
}
