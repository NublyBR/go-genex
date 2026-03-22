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

func BenchmarkAll(b *testing.B) {
	// Some of these initialize the structs directly to skip over the optimization paths.
	var tests = []struct {
		Name string
		Init func(b *testing.B) Generator
	}{
		{
			Name: "Fixed",
			Init: func(b *testing.B) Generator {
				return NewFixed([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
			},
		},
		{
			Name: "Charset",
			Init: func(b *testing.B) Generator {
				return NewCharset('a', 'z')
			},
		},
		{
			Name: "Choice",
			Init: func(b *testing.B) Generator {
				return &Choice{
					items: []Generator{
						NewFixed([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
						NewFixed([]byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
						NewFixed([]byte("cccccccccccccccccccccccccccccccc")),
					},
					min: 32,
					max: 32,
					rng: FastRand,
				}
			},
		},
		{
			Name: "Concat",
			Init: func(b *testing.B) Generator {
				return &Concat{
					items: []Generator{
						NewFixed([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
						NewFixed([]byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
						NewFixed([]byte("cccccccccccccccccccccccccccccccc")),
					},
					min: 32 * 3,
					max: 32 * 3,
				}
			},
		},
		{
			Name: "Numeric",
			Init: func(b *testing.B) Generator {
				return NewNumeric(62, 0, 0xffff_ffff_ffff_ffff, 1, false)
			},
		},
		{
			Name: "Repeat",
			Init: func(b *testing.B) Generator {
				return &Repeat{
					inner: NewFixed([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),

					rmin: 3,
					rmax: 3,

					min: 32 * 3,
					max: 32 * 3,

					rng: FastRand,
				}
			},
		},
	}

	for _, test := range tests {
		b.Run(test.Name, func(b *testing.B) {
			gen := test.Init(b)
			min, max := gen.Bounds()
			b.SetBytes(int64(min+max) / 2)

			buf := bytes.NewBuffer(make([]byte, max))

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				gen.Sample(buf)
				buf.Reset()
			}
		})
	}
}
