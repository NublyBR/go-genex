package genex

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

// Export represents a generator such that it can safely serialized as JSON, Yaml, or other formats.
// Useful for very large expressions that could become cumbersome to manage as text.
func Export(g Generator) any {
	return g.export()
}

// Import restores a generator from an exported value.
// Use json.Decoder.UseNumber when decoding JSON to preserve large integers.
func Import(v any, options ...Option) (Generator, error) {
	switch cast := v.(type) {
	case nil:
		return optionApplyFn(NewFixed(nil), options...), nil

	case []byte:
		return optionApplyFn(NewFixed(cast), options...), nil

	case string:
		dec, err := base64.StdEncoding.DecodeString(cast)
		if err != nil {
			return nil, err
		}
		return optionApplyFn(NewFixed(dec), options...), nil

	case []any:
		ret := make([]Generator, 0, len(cast))
		for _, item := range cast {
			cur, err := Import(item, options...)
			if err != nil {
				return nil, err
			}
			ret = append(ret, cur)
		}
		return optionApplyFn(NewConcat(ret...), options...), nil

	case map[string]any:
		if charset, ok := cast["charset"]; ok {
			var chars []byte

			switch c := charset.(type) {
			case []byte:
				chars = c
			case string:
				dec, err := base64.StdEncoding.DecodeString(c)
				if err != nil {
					return nil, err
				}
				chars = dec
			default:
				return nil, errors.New("invalid charset")
			}

			if len(chars)%2 != 0 {
				return nil, errors.New("invalid charset")
			}

			return optionApplyFn(NewCharset(chars...), options...), nil
		}

		if choice, ok := cast["choice"]; ok {
			choices, ok := choice.([]any)
			if !ok {
				return nil, errors.New("invalid choices")
			}

			ret := make([]Generator, 0, len(choices))
			for _, item := range choices {
				cur, err := Import(item, options...)
				if err != nil {
					return nil, err
				}
				ret = append(ret, cur)
			}
			return optionApplyFn(NewChoice(ret...), options...), nil
		}

		if repeat, ok := cast["repeat"]; ok {
			var rmin, rmax int

			if slice, ok := repeat.([]any); ok {
				if len(slice) != 2 {
					return nil, fmt.Errorf("invalid repeat: %v", slice)
				}

				if err := readInt(&rmin, slice[0]); err != nil {
					return nil, fmt.Errorf("invalid repeat: %s", err)
				}

				if err := readInt(&rmax, slice[1]); err != nil {
					return nil, fmt.Errorf("invalid repeat: %s", err)
				}
			} else {
				if err := readInt(&rmin, repeat); err != nil {
					return nil, fmt.Errorf("invalid repeat: %s", err)
				}

				rmax = rmin
			}

			item, ok := cast["item"]
			if !ok {
				return nil, fmt.Errorf("missing repeat item")
			}

			sub, err := Import(item, options...)
			if err != nil {
				return nil, err
			}

			return optionApplyFn(NewRepeat(sub, rmin, rmax), options...), nil
		}

		if _, ok := cast["start"]; ok {
			var (
				base  int = 10
				start uint64
				end   uint64
				step  uint64 = 1
				pad   bool
			)

			if nbase, ok := cast["base"]; ok {
				if err := readInt(&base, nbase); err != nil {
					return nil, fmt.Errorf("invalid numeric base: %s", err)
				}
			}

			if err := readInt(&start, cast["start"]); err != nil {
				return nil, fmt.Errorf("invalid numeric start: %s", err)
			}

			if err := readInt(&end, cast["end"]); err != nil {
				return nil, fmt.Errorf("invalid numeric end: %s", err)
			}

			if nstep, ok := cast["step"]; ok {
				if err := readInt(&step, nstep); err != nil {
					return nil, fmt.Errorf("invalid numeric step: %s", err)
				}
			}

			if cast["pad"] == true {
				pad = true
			}

			return optionApplyFn(NewNumeric(base, start, end, step, pad), options...), nil
		}
	}

	return nil, errors.New("no valid generator found")
}

// ExportSampler stores the generator and sampler position. Index and seed are
// decimal strings so serializer round trips preserve their full precision.
func ExportSampler(s *Sampler) any {
	return map[string]any{
		"gen":   s.Gen.export(),
		"index": s.Index.String(),
		"seed":  strconv.FormatInt(s.Seed, 10),
	}
}

// ImportSampler restores an exported sampler. Omitted index and seed default
// to zero; present values must be decimal strings. Out-of-range positions are
// preserved, and Next returns false for them as usual.
func ImportSampler(v any, options ...Option) (*Sampler, error) {
	mp, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("invalid sampler: expected an object")
	}

	value, ok := mp["gen"]
	if !ok {
		return nil, errors.New("missing sampler generator")
	}
	gen, err := Import(value, options...)
	if err != nil {
		return nil, fmt.Errorf("invalid sampler generator: %w", err)
	}

	var index big.Int
	seed := int64(0)

	if value, present := mp["index"]; present {
		str, ok := value.(string)
		if !ok {
			return nil, errors.New("invalid sampler index: expected a decimal string")
		}
		if _, ok := index.SetString(str, 10); !ok {
			return nil, fmt.Errorf("invalid sampler index: %q", str)
		}
	}
	if value, present := mp["seed"]; present {
		str, ok := value.(string)
		if !ok {
			return nil, errors.New("invalid sampler seed: expected a decimal string")
		}
		seed, err = strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sampler seed: %w", err)
		}
	}
	sampler := NewSampler(gen, seed)
	sampler.Index.Set(&index)
	return sampler, nil
}
