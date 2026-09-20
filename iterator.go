package genex

import (
	"bytes"
	"iter"
)

type iterator struct {
	state int // 0:start 1:iter 2:end
	get   func(w *bytes.Buffer)
	next  func() bool
	reset func()
}

func (i *iterator) Get(w *bytes.Buffer) {
	if i.state != 1 {
		panic("genex: Get called before Next or after iteration ended")
	}
	i.get(w)
}

func (i *iterator) Next() bool {
	switch {
	case i.state == 0:
		i.state = 1
		return true

	case i.state == 2:
		return false

	case i.next():
		return true

	default:
		i.state = 2
		return false
	}
}

func (i *iterator) Reset() {
	i.reset()
	i.state = 0
}

func makeSeq(g Generator) iter.Seq[[]byte] {
	_, max := g.Bounds()
	buf := bytes.NewBuffer(make([]byte, 0, max))
	it := g.iterate()
	return func(yield func([]byte) bool) {
		for it.Next() {
			buf.Reset()
			it.Get(buf)
			if !yield(buf.Bytes()) {
				return
			}
		}
	}
}
