package genex

import (
	"testing"
)

func TestCompiler(t *testing.T) {
	_, err := Compile("<64>{1}[a-z]{2}(a{3}b|c(d)){4}")
	if err != nil {
		t.Fatal(err)
	}
}
