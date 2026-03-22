package genex

import (
	"bytes"
	"testing"
)

const benchmarkPattern = `((([0-9]|([1-9][0-9])|(1[0-9]{2})|(2[0-4][0-9])|(25[0-5])).){3}([0-9]|([1-9][0-9])|(1[0-9]{2})|(2[0-4][0-9])|(25[0-5])))`

func BenchmarkRandom(b *testing.B) {
	g, err := Compile(benchmarkPattern)
	if err != nil {
		b.Fatal(err)
	}

	_, max := g.Bounds()

	buf := bytes.NewBuffer(make([]byte, 0, max))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.Sample(buf)
		buf.Reset()
	}
}

func BenchmarkIter(b *testing.B) {
	g, err := Compile(benchmarkPattern)
	if err != nil {
		b.Fatal(err)
	}

	_, max := g.Bounds()

	buf := bytes.NewBuffer(make([]byte, 0, max))

	iter := g.Iterate()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		iter.Next()
		iter.Get(buf)
		buf.Reset()
	}
}

func BenchmarkCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Compile(benchmarkPattern)
	}
}
