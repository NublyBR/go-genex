package genex

import (
	"fmt"
	"regexp"
	"strconv"
)

type TokenType int

const (
	TokEOF TokenType = iota - 1
	TokRaw
	TokPipe
	TokOpt
	TokPlus
	TokStar
	TokAny
	TokSpecial

	TokParOpen
	TokParClose

	TokBraOpen
	TokBraClose

	TokCurOpen
	TokCurClose

	TokTagOpen
	TokTagClose
)

type Token struct {
	Type  TokenType
	Value []byte
}

func (t Token) String() string {
	return fmt.Sprintf("(%d %q)", t.Type, t.Value)
}

var (
	special = `\|\(\)\[\]\{\}\<\>\?\+\*\.\\`
	escape  = `\\(x[0-9a-fA-F]{2}|[^wWsSdhH])`
	escapeR = `\\(?:x[0-9a-fA-F]{2}|.)`
	// reSpecial = regexp.MustCompile(`^[` + special + `]`)
	reRaw     = regexp.MustCompile(`^(` + escape + `|[^` + special + `])+`)
	reEscape  = regexp.MustCompile(escape)
	reSpecial = regexp.MustCompile(`^\\[wWsSdhH]`)
)

func NewToken(typ TokenType, val []byte) Token {
	return Token{Type: typ, Value: val}
}

func Unescape(b []byte) []byte {
	return reEscape.ReplaceAllFunc(b, func(b []byte) []byte {
		switch b[1] {
		case 'a':
			return []byte{'\a'}
		case 'b':
			return []byte{'\b'}
		case 'f':
			return []byte{'\f'}
		case 'n':
			return []byte{'\n'}
		case 'r':
			return []byte{'\r'}
		case 't':
			return []byte{'\t'}
		case 'v':
			return []byte{'\v'}
		case 'x':
			val, _ := strconv.ParseUint(string(b[2:]), 16, 8)
			return []byte{byte(val)}
		default:
			return b[1:]
		}
	})
}

var tokMap = [256]TokenType{
	'|': TokPipe,
	'?': TokOpt,
	'+': TokPlus,
	'*': TokStar,
	'(': TokParOpen,
	')': TokParClose,
	'[': TokBraOpen,
	']': TokBraClose,
	'{': TokCurOpen,
	'}': TokCurClose,
	'<': TokTagOpen,
	'>': TokTagClose,
	'.': TokAny,
}

func Tokenize(data []byte) []Token {
	var tk []Token

	for len(data) > 0 {
		if found := tokMap[data[0]]; found != TokRaw {
			tk = append(tk, NewToken(found, data[:1]))
			data = data[1:]
			continue
		}

		if res := reRaw.Find(data); res != nil {
			tk = append(tk, NewToken(TokRaw, Unescape(res)))
			data = data[len(res):]
			continue
		}

		if res := reSpecial.Find(data); res != nil {
			tk = append(tk, NewToken(TokSpecial, res[1:]))
			data = data[len(res):]
			continue
		}

		data = data[1:]
	}

	return tk
}
