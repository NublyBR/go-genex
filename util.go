package genex

import (
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
