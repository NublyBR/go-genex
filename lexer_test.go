package genex

import (
	"reflect"
	"testing"
)

func TestLexer(t *testing.T) {
	type test struct {
		Pattern string
		Expect  []Token
	}

	var tests = []test{
		{
			Pattern: `[a-zA-Z0-9!-\]\n\x1d]{32}`,
			Expect: []Token{
				NewToken(TokBraOpen, []byte("[")),
				NewToken(TokRaw, []byte("a-zA-Z0-9!-]\n\x1d")),
				NewToken(TokBraClose, []byte("]")),
				NewToken(TokCurOpen, []byte("{")),
				NewToken(TokRaw, []byte("32")),
				NewToken(TokCurClose, []byte("}")),
			},
		},
		{
			Pattern: `<5>`,
			Expect: []Token{
				NewToken(TokTagOpen, []byte("<")),
				NewToken(TokRaw, []byte("5")),
				NewToken(TokTagClose, []byte(">")),
			},
		},
	}

	for _, cur := range tests {
		result := Tokenize([]byte(cur.Pattern))

		if !reflect.DeepEqual(cur.Expect, result) {
			t.Errorf("expected Tokenize(%q) to be %+v, got %+v", cur.Pattern, cur.Expect, result)
		}
	}

	CompileTree(
		Tokenize([]byte(
			`[a-zA-Z0-9]{32}`,
		)),
	)
}
