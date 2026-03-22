package genex

import "bytes"

type (
	ItGet   func(w *bytes.Buffer)
	ItNext  func() bool
	ItReset func()
)

type Iterator struct {
	state int // 0:start 1:iter 2:end
	get   ItGet
	next  ItNext
	reset ItReset
}

func (i *Iterator) Get(w *bytes.Buffer) {
	if i.state != 1 {
		panic("genex: Get called before Next or after iteration ended")
	}
	i.get(w)
}

func (i *Iterator) Next() bool {
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

func (i *Iterator) Reset() {
	i.reset()
	i.state = 0
}

var DummyIterator = &Iterator{
	get:   func(w *bytes.Buffer) {},
	next:  func() bool { return false },
	reset: func() {},
}
