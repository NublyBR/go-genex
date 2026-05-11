package genex

import (
	"bytes"
	"testing"
)

func TestIterator(t *testing.T) {
	gen, err := Compile(`[a-z]`)
	if err != nil {
		t.Fatal(err)
	}

	it := gen.Iterate()
	buf := bytes.NewBuffer(make([]byte, 0, 1))

	for i := 'a'; i <= 'z'; i++ {
		buf.Reset()

		if ok := it.Next(); !ok {
			t.Fatalf("unexpected end of iterator at %c", i)
		}

		it.Get(buf)

		ret := buf.Bytes()
		if len(ret) != 1 || ret[0] != byte(i) {
			t.Fatalf("unexpected byte at %c: %v", i, ret)
		}
	}
}
