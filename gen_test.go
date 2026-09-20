package genex

import (
	"testing"
)

func TestIterator(t *testing.T) {
	gen, err := Compile(`[a-z]`)
	if err != nil {
		t.Fatal(err)
	}

	i := 'a'

	for next := range gen.Iterate() {
		if i > 'z' {
			t.Fatalf("unexpected extra byte: %v", next)
		}
		if len(next) != 1 || next[0] != byte(i) {
			t.Fatalf("unexpected byte at %c: %v", i, next)
		}
		i++
	}
}
