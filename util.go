package genex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
)

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
		w.Write([]byte{
			'\\',
			sp[c],
		})
		return
	}

	if c < '\x19' || c > '\x7e' {
		w.Write([]byte{
			'\\',
			'x',
			hex[c>>4],
			hex[c&0xF],
		})
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

// readInt converts numeric values without truncation or overflow. Use
// json.Decoder.UseNumber when decoding large integers: precision already lost
// by decoding into float64 cannot be recovered here.
func readInt(ptr, val any) error {
	dst := reflect.ValueOf(ptr)
	if !dst.IsValid() || dst.Kind() != reflect.Pointer || dst.IsNil() {
		return fmt.Errorf("invalid integer destination: %T", ptr)
	}
	dst = dst.Elem()
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return fmt.Errorf("invalid integer destination: %T", ptr)
	}

	var number big.Int
	if raw, ok := val.(json.Number); ok {
		// A rational preserves decimal and exponent notation exactly, including
		// integral values such as 1.0 or 1e3, without rounding through float64.
		var rational big.Rat
		if !json.Valid([]byte(raw)) {
			return fmt.Errorf("invalid number: %q", raw)
		}
		if _, ok := rational.SetString(string(raw)); !ok || !rational.IsInt() {
			return fmt.Errorf("invalid integer: %q", raw)
		}
		number.Set(rational.Num())
	} else {
		src := reflect.ValueOf(val)
		switch src.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			number.SetInt64(src.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			number.SetUint64(src.Uint())
		case reflect.Float32, reflect.Float64:
			f := src.Float()
			if math.IsNaN(f) || math.IsInf(f, 0) || math.Trunc(f) != f {
				return fmt.Errorf("invalid integer: %v", val)
			}
			new(big.Float).SetFloat64(f).Int(&number)
		default:
			return fmt.Errorf("invalid number: %v (%T)", val, val)
		}
	}

	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if number.IsInt64() && !dst.OverflowInt(number.Int64()) {
			dst.SetInt(number.Int64())
			return nil
		}
	default:
		if number.IsUint64() && !dst.OverflowUint(number.Uint64()) {
			dst.SetUint(number.Uint64())
			return nil
		}
	}
	return fmt.Errorf("number %v overflows %s", val, dst.Type())
}
