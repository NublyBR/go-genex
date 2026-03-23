package genex

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type TokenConsumer struct {
	Tokens     []Token
	Count, Idx int
	opt        func(Generator) Generator
}

func errUnexpected(tok Token) error {
	if tok.Type == TokEOF {
		return fmt.Errorf("unexpected EOF")
	}
	return fmt.Errorf("unexpected token %q", tok.Value)
}

func (t *TokenConsumer) Peek(offset int) Token {
	if t.Idx+offset-1 >= t.Count {
		return NewToken(TokEOF, nil)
	}

	return t.Tokens[t.Idx+offset-1]
}

func (t *TokenConsumer) Get() Token {
	if t.Idx+1 > t.Count {
		return NewToken(TokEOF, nil)
	}

	t.Idx++
	return t.Tokens[t.Idx-1]
}

func (t *TokenConsumer) Has() bool {
	return t.Idx < t.Count
}

func (t *TokenConsumer) Mul() (int, int, error) {
	switch next := t.Peek(1); next.Type {
	case TokOpt:
		t.Get()
		return 0, 1, nil

	case TokPlus:
		t.Get()
		return 1, 8, nil

	case TokStar:
		t.Get()
		return 0, 8, nil

	case TokCurOpen:
		t.Get()

		param := t.Get()
		if param.Type != TokRaw {
			return 0, 0, errUnexpected(param)
		}

		if close := t.Get(); close.Type != TokCurClose {
			return 0, 0, errUnexpected(close)
		}

		nums := make([]int, 0, 2)
		split := bytes.Split(param.Value, []byte(","))
		if len(split) == 0 || len(split) > 2 {
			return 0, 0, errors.New("unexpected amount of parameters")
		}

		for _, arg := range split {
			if len(arg) == 0 {
				nums = append(nums, -1)
				continue
			}

			num, err := strconv.ParseUint(string(arg), 10, 32)
			if err != nil {
				return 0, 0, err
			}
			nums = append(nums, int(num))
		}

		switch len(nums) {
		case 1:
			if nums[0] == -1 {
				return 0, 0, errors.New("invalid parameters")
			}
			return nums[0], nums[0], nil
		case 2:
			if nums[0] == -1 && nums[1] == -1 {
				return 0, 0, errors.New("invalid parameters")
			}
			if nums[0] == -1 {
				return 0, nums[1], nil
			}
			if nums[1] == -1 {
				return nums[0], nums[0], nil
			}
			return nums[0], nums[1], nil
		default:
			panic(0)
		}

	default:
		return 1, 1, nil
	}
}

func (t *TokenConsumer) Multify(g Generator) (Generator, error) {
	min, max, err := t.Mul()
	if err != nil {
		return nil, err
	}

	if min == 1 && max == 1 {
		return g, nil
	}

	return t.opt(NewRepeat(g, min, max)), nil
}

var (
	reCharset = regexp.MustCompile(`(` + escapeR + `|[^-])-(` + escapeR + `|[^-])|(` + escapeR + `|.)`)
)

func (t *TokenConsumer) Parse() (Generator, error) {
	switch next := t.Get(); next.Type {
	case TokRaw:
		if min, max, err := t.Mul(); err != nil {
			return nil, err
		} else if min != 1 || max != 1 {
			idx := len(next.Value) - 1

			return t.opt(NewConcat(
				t.opt(NewFixed(next.Value[:idx])),
				t.opt(NewRepeat(
					t.opt(NewFixed(next.Value[idx:])),
					min, max,
				)),
			)), nil
		}
		return t.opt(NewFixed(next.Value)), nil

	case TokAny:
		return t.Multify(t.opt(NewCharset('!', '~')))

	case TokSpecial:
		inner := NewFixed(nil)

		switch next.Value[0] {
		case 'w':
			inner = NewCharset('a', 'z', 'A', 'Z', '0', '9', '_', '_')
		case 'W':
			inner = NewCharset('!', '/', ':', '@', '[', '^', '`', '`', '{', '~')
		case 's':
			inner = NewCharset(' ', ' ', '\t', '\t', '\f', '\f', '\r', '\r', '\n', '\n')
		case 'S':
			inner = NewCharset('!', '~')
		case 'd':
			inner = NewCharset('0', '9')
		case 'h':
			inner = NewCharset('0', '9', 'a', 'f')
		case 'H':
			inner = NewCharset('0', '9', 'A', 'F')
		}

		return t.Multify(t.opt(inner))

	case TokParOpen:
		choice := []Generator{}
		concat := []Generator{}

	parLoop:
		for {
			switch inner := t.Peek(1); inner.Type {
			case TokParClose:
				t.Get()
				break parLoop

			case TokPipe:
				t.Get()
				choice = append(choice, t.opt(NewConcat(concat...)))
				concat = []Generator{}

			case TokEOF:
				return nil, errUnexpected(inner)

			default:
				gen, err := t.Parse()
				if err != nil {
					return nil, err
				}
				concat = append(concat, gen)
			}
		}

		if len(concat) > 0 {
			choice = append(choice, t.opt(NewConcat(concat...)))
		}

		return t.Multify(t.opt(NewChoice(choice...)))

	case TokBraOpen:
		opts := make([]byte, 0, 2*16)
		buf := bytes.NewBuffer(nil)

		commit := func() {
			if buf.Len() == 0 {
				return
			}
			defer buf.Reset()

			mt := reCharset.FindAllSubmatch(buf.Bytes(), -1)
			for _, cur := range mt {
				if len(cur[3]) > 0 {
					val := Unescape(cur[3])
					opts = append(opts, val[0], val[0])
				} else {
					min := Unescape(cur[1])
					max := Unescape(cur[2])
					opts = append(opts, min[0], max[0])
				}
			}
		}

	bracketLoop:
		for {
			next := t.Get()

			switch next.Type {
			case TokBraClose:
				break bracketLoop

			case TokRaw, TokAny:
				buf.Write(next.Value)

			case TokSpecial:
				commit()

				switch next.Value[0] {
				case 'w':
					opts = append(opts, 'a', 'z', 'A', 'Z', '0', '9', '_', '_')
				case 'W':
					opts = append(opts, '!', '/', ':', '@', '[', '^', '`', '`', '{', '~')
				case 's':
					opts = append(opts, ' ', ' ', '\t', '\t', '\f', '\f', '\r', '\r', '\n', '\n')
				case 'S':
					opts = append(opts, '!', '~')
				case 'd':
					opts = append(opts, '0', '9')
				case 'h':
					opts = append(opts, '0', '9', 'a', 'f')
				case 'H':
					opts = append(opts, '0', '9', 'A', 'F')
				}

			default:
				return nil, errUnexpected(next)
			}
		}

		commit()

		return t.Multify(t.opt(NewCharset(opts...)))

	case TokTagOpen:
		inner := t.Get()
		if inner.Type != TokRaw {
			return nil, errUnexpected(inner)
		}

		if close := t.Get(); close.Type != TokTagClose {
			return nil, errUnexpected(close)
		}

		var (
			base  = int(10)
			start = uint64(0)
			end   = uint64(0)
			step  = uint64(1)
			pad   = false

			ptr = inner.Value
		)

		if ptr[0] == '!' {
			pad = true
			ptr = ptr[1:]
		}

		if idx := bytes.IndexByte(ptr, ':'); idx != -1 {
			bs, err := strconv.ParseUint(string(ptr[:idx]), 10, 32)
			if err != nil {
				return nil, err
			}
			base = int(bs)
			if base < 2 || base > numBaseMax {
				return nil, fmt.Errorf("invalid base %d, expected %d-%d", base, 2, numBaseMax)
			}
			ptr = ptr[idx+1:]
		}

		parts := bytes.Split(ptr, []byte("/"))
		if len(parts) > 3 {
			return nil, fmt.Errorf("expected maximum of 3 arguments, got %d", len(parts))
		}

		ints := make([]uint64, len(parts))
		for i, part := range parts {
			if len(part) == 0 {
				continue
			}

			cur, err := numParse(part, base)
			if err != nil {
				return nil, err
			}

			ints[i] = cur
		}

		switch len(ints) {
		case 1:
			end = ints[0]

		case 2:
			start = ints[0]
			end = ints[1]

		case 3:
			start = ints[0]
			end = ints[1]
			step = ints[2]
		}

		if step < 1 {
			return nil, fmt.Errorf("invalid step %d", step)
		}

		return t.Multify(t.opt(NewNumeric(base, start, end, step, pad)))

	default:
		return nil, errUnexpected(next)
	}
}

func NewTokenConsumer(t []Token) *TokenConsumer {
	return &TokenConsumer{Tokens: t, Count: len(t)}
}

func CompileTree(t []Token, opts ...Option) (Generator, error) {
	opt := optionApply(opts...)

	switch len(t) {
	case 0:
		return opt(NewFixed(nil)), nil
	case 1:
		if t[0].Type == TokRaw {
			return opt(NewFixed(t[0].Value)), nil
		}
	}

	actual := make([]Token, 0, len(t)+2)
	actual = append(actual, NewToken(TokParOpen, []byte("(")))
	actual = append(actual, t...)
	actual = append(actual, NewToken(TokParClose, []byte(")")))
	con := NewTokenConsumer(actual)
	con.opt = opt

	res, err := con.Parse()
	if err != nil {
		return nil, err
	}

	if con.Has() {
		return nil, errUnexpected(con.Get())
	}

	return res, nil
}

func Compile(s string, opts ...Option) (Generator, error) {
	tokens := Tokenize([]byte(s))
	return CompileTree(tokens, opts...)
}
