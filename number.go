package genex

import "fmt"

const (
	numBase    = `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	numBaseMax = len(numBase)
)

var (
	numMap = [256]int{}
)

func init() {
	for i := range 256 {
		numMap[i] = -1
	}
	for i, c := range numBase {
		numMap[c] = i
	}
}

func numSize(n uint64, base int) int {
	if n == 0 {
		return 1
	}

	d := 0
	for n > 0 {
		n /= uint64(base)
		d++
	}
	return d
}

func numParse(buf []byte, base int) (uint64, error) {
	res := uint64(0)
	for _, chr := range buf {
		val := numMap[chr]
		if val == -1 {
			return 0, fmt.Errorf("invalid number %q base %d", buf, base)
		}

		res = res*uint64(base) + uint64(val)
	}
	return res, nil
}

func numEncode(buf []byte, n uint64, base int) []byte {
	idx := len(buf)
	if n == 0 {
		buf[idx-1] = numBase[0]
		return buf[idx-1:]
	}
	for n > 0 {
		idx--
		buf[idx] = numBase[int(n%uint64(base))]
		n /= uint64(base)
	}
	return buf[idx:]
}
