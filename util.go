package genex

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"
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

var state = uint64(time.Now().UnixNano())

func FastRand() int64 {
	z := atomic.AddUint64(&state, 0x9e3779b97f4a7c15)
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return int64((z ^ (z >> 31)) & 0x7fff_ffff_ffff_ffff)
}

func SecureRand() int64 {
	var buf [8]byte
	rand.Read(buf[:])
	return int64(binary.BigEndian.Uint64(buf[:]) & 0x7fff_ffff_ffff_ffff)
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
