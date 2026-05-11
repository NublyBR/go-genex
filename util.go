package genex

import (
	"bytes"
	"fmt"
	"math/big"
)

type byteWriter interface {
	WriteByte(byte) error
}

func Readable(b *big.Int) string {
	if b.IsInt64() {
		i := b.Int64()

		if i < 0 {
			i = -i
		}

		if i < 10_000_000 {
			return fmt.Sprint(i)
		}
	}

	return fmt.Sprintf("%e", new(big.Float).SetInt(b))
}

func SampleString(gen Generator) string {
	_, max := gen.Bounds()
	buf := bytes.NewBuffer(make([]byte, 0, max))
	gen.Sample(buf)
	return buf.String()
}

var sp = [256]byte{
	'-':  '-',
	'|':  '|',
	'?':  '?',
	'+':  '+',
	'*':  '*',
	'(':  '(',
	')':  ')',
	'[':  '[',
	']':  ']',
	'{':  '{',
	'}':  '}',
	'<':  '<',
	'>':  '>',
	'.':  '.',
	'\a': 'a',
	'\b': 'b',
	'\f': 'f',
	'\n': 'n',
	'\r': 'r',
	'\t': 't',
	'\v': 'v',
}

func writeSpecial(w *bytes.Buffer, c byte) {
	const hex = "0123456789abcdef"

	if sp[c] != 0 {
		w.WriteByte('\\')
		w.WriteByte(sp[c])
		return
	}

	if c < '\x19' || c > '\x7e' {
		w.WriteByte('\\')
		w.WriteByte('x')
		w.WriteByte(hex[c>>4])
		w.WriteByte(hex[c&0xF])
		return
	}

	w.WriteByte(c)
}

func apply[T any](t []T, f func(T) T) []T {
	ret := make([]T, len(t))

	for i := range t {
		ret[i] = f(t[i])
	}

	return ret
}

func Clone(g Generator, opts ...Option) Generator {
	switch cast := g.(type) {
	case *Charset:
		cpy := *cast
		return optionApplyFn(&cpy, opts...)

	case *Choice:
		cpy := *cast
		cpy.items = apply(cpy.items, func(other Generator) Generator {
			return Clone(other, opts...)
		})
		return optionApplyFn(&cpy, opts...)

	case *Repeat:
		cpy := *cast
		cpy.item = Clone(cpy.item, opts...)
		return optionApplyFn(&cpy, opts...)

	case *Numeric:
		cpy := *cast
		return optionApplyFn(&cpy, opts...)

	case *Concat:
		cpy := *cast
		cpy.items = apply(cpy.items, func(other Generator) Generator {
			return Clone(other, opts...)
		})
		return optionApplyFn(&cpy, opts...)

	case *Fixed:
		cpy := *cast
		return optionApplyFn(&cpy, opts...)

	default:
		return g
	}
}
