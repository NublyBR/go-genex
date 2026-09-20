package genex

import (
	"bytes"
	"math/big"
	"testing"
)

func TestIterator(t *testing.T) {
	g := MustCompile(`[a-z]`)

	i := 'a'

	for next := range g.Iterate() {
		if i > 'z' {
			t.Fatalf("unexpected extra byte: %v", next)
		}
		if len(next) != 1 || next[0] != byte(i) {
			t.Fatalf("unexpected byte at %c: %v", i, next)
		}
		i++
	}
}

func TestIndex(t *testing.T) {
	var tests = []struct {
		Pattern  string
		Index    int64
		Expected string
	}{
		{
			Pattern:  `<255>\.<255>\.<255>\.<255>`,
			Index:    0x04_03_02_01,
			Expected: "1.2.3.4",
		},
		{
			Pattern:  `[a-z]{5}`,
			Index:    0x64af93, // hex(sum([(ord(n) - ord('a')) * (26**i) for i, n in enumerate('hello')]))
			Expected: "hello",
		},
		{
			Pattern:  `(ab|cd|ef|gh)`,
			Index:    2,
			Expected: "ef",
		},
		{
			Pattern:  `(x|[ab])`,
			Index:    2,
			Expected: "b",
		},
		{
			Pattern:  `[ab]{0,2}`,
			Index:    6,
			Expected: "bb",
		},
		{
			Pattern:  `<!16:a/f/2>`,
			Index:    2,
			Expected: "e",
		},
		{
			Pattern:  `<!10:8/12/2>`,
			Index:    0,
			Expected: "08",
		},
	}

	buf := bytes.NewBuffer(make([]byte, 0, 255))

	for _, test := range tests {
		g := MustCompile(test.Pattern)

		buf.Reset()
		g.Index(buf, big.NewInt(test.Index))

		if actual := buf.String(); actual != test.Expected {
			t.Errorf("Generating pattern %q at index %d expected %q, got %q",
				test.Pattern, test.Index, test.Expected, actual)
		}
	}

	// Check every index in small search spaces, including wrapping and the
	// quotient left for subsequent components. A prefix checks append behavior.
	for _, test := range []struct {
		pattern string
		values  []string
	}{
		{`fixed`, []string{"fixed"}},
		{`[a-c]`, []string{"a", "b", "c"}},
		{`<2/8/3>`, []string{"2", "5", "8"}},
		{`(x|[ab])`, []string{"x", "a", "b"}},
		{`([ab]|x)`, []string{"a", "b", "x"}},
		{`(a[01]|b[01])`, []string{"a0", "b0", "a1", "b1"}},
		{`(x|[ab])[01]`, []string{"x0", "a0", "b0", "x1", "a1", "b1"}},
		{`[ab]{2}`, []string{"aa", "ba", "ab", "bb"}},
		{`[ab]{0,2}`, []string{"", "a", "b", "aa", "ba", "ab", "bb"}},
		{`[ab]{1,2}`, []string{"a", "b", "aa", "ba", "ab", "bb"}},
		{`x{0,2}`, []string{"", "x", "xx"}},
		{`[ab]?[01]`, []string{"0", "a0", "b0", "1", "a1", "b1"}},
		{`(x|[ab]){1,2}`, []string{"x", "a", "b", "xx", "ax", "bx", "xa", "aa", "ba", "xb", "ab", "bb"}},
		{`(x|[ab]{0,1})[01]`, []string{"x0", "0", "a0", "b0", "x1", "1", "a1", "b1"}},
	} {
		t.Run(test.pattern, func(t *testing.T) {
			gen := MustCompile(test.pattern)
			count := gen.Count()
			if count.Cmp(big.NewInt(int64(len(test.values)))) != 0 {
				t.Fatalf("count = %s, want %d", count, len(test.values))
			}
			for _, quotient := range []int64{-1, 0, 2} {
				for i, expected := range test.values {
					idx := new(big.Int).Mul(count, big.NewInt(quotient))
					idx.Add(idx, big.NewInt(int64(i)))
					buf := bytes.NewBufferString("prefix:")
					gen.Index(buf, idx)
					if buf.String() != "prefix:"+expected || idx.Cmp(big.NewInt(quotient)) != 0 {
						t.Errorf("offset %d, cycle %d: got (%q, remainder %s), want (%q, remainder %d)",
							i, quotient, buf.String(), idx, "prefix:"+expected, quotient)
					}
				}
			}
		})
	}

	t.Run("large indices", func(t *testing.T) {
		for _, test := range []struct {
			pattern, index, expected, remainder string
		}{
			{`<9223372036854775807>`, "9223372036854775807", "9223372036854775807", "0"},
			{`<9223372036854775807>`, "9223372036854775808", "0", "1"},
			{`<9223372036854775807>[a-c]`, "9223372036854775808", "0b", "0"},
			{`<9223372036854775807>[a-c]`, "27670116110564327424", "0a", "1"},
			{`[ab]{65}`, "18446744073709551616", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab", "0"},
			{`(x|<9223372036854775807>)[ab]`, "9223372036854775809", "xb", "0"},
			{`[ab]{0,64}[01]`, "36893488147419103231", "1", "0"},
		} {
			gen := MustCompile(test.pattern)
			idx, ok := new(big.Int).SetString(test.index, 10)
			if !ok {
				t.Fatalf("invalid test index %q", test.index)
			}
			var buf bytes.Buffer
			gen.Index(&buf, idx)
			if buf.String() != test.expected || idx.String() != test.remainder {
				t.Errorf("%s at %s: got (%q, remainder %s), want (%q, remainder %s)",
					test.pattern, test.index, buf.String(), idx, test.expected, test.remainder)
			}
		}
	})
}
