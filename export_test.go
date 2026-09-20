package genex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestExportImportSampler(t *testing.T) {
	for _, test := range []struct {
		name, pattern, index string
		seed                 int64
	}{
		{"start", `[a-z]`, "0", 0},
		{"resume", `(x|[ab]){0,2}`, "5", -42},
		{"numeric", `<!16:a/ff/2>`, "3", 42},
		{"huge index", `.{512}`, new(big.Int).Lsh(big.NewInt(1), 1000).String(), math.MaxInt64},
		{"minimum seed", `[a-z]`, "3", math.MinInt64},
		{"last", `[a-z]`, "25", 42},
		{"exhausted", `[a-z]`, "26", 42},
		{"past end", `[a-z]`, "27", 42},
		{"negative position", `[a-z]`, "-1", 42},
		{"empty value", ``, "0", 42},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, path := range []string{"direct", "json"} {
				t.Run(path, func(t *testing.T) {
					original := NewSampler(MustCompile(test.pattern), test.seed)
					if _, ok := original.Index.SetString(test.index, 10); !ok {
						t.Fatal("invalid test index")
					}
					exported := ExportSampler(original)
					if path == "json" {
						data, err := json.Marshal(exported)
						if err != nil {
							t.Fatal(err)
						}
						exported = nil
						if err := json.Unmarshal(data, &exported); err != nil {
							t.Fatal(err)
						}
					}
					imported, err := ImportSampler(exported)
					if err != nil {
						t.Fatal(err)
					}
					if imported.Gen.String() != original.Gen.String() || imported.Seed != original.Seed || imported.Index.Cmp(&original.Index) != 0 {
						t.Fatal("sampler state changed during round trip")
					}
					for range 4 {
						want, got := bytes.NewBufferString("prefix:"), bytes.NewBufferString("prefix:")
						wantOK, gotOK := original.Next(want), imported.Next(got)
						if gotOK != wantOK || got.String() != want.String() || imported.Index.Cmp(&original.Index) != 0 {
							t.Fatal("restored sampler did not continue the original sequence")
						}
					}
				})
			}
		})
	}
}

func TestImportSamplerDefaults(t *testing.T) {
	for _, value := range []map[string]any{
		{"gen": nil},
		{"gen": nil, "index": "0"},
		{"gen": nil, "seed": "0"},
	} {
		sampler, err := ImportSampler(value)
		if err != nil {
			t.Fatal(err)
		}
		if sampler.Seed != 0 || sampler.Index.Sign() != 0 || sampler.Gen.String() != NewFixed(nil).String() {
			t.Fatal("incorrect defaults")
		}
	}
}

func TestImportSamplerInvalid(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"wrong type", "sampler"},
		{"missing generator", map[string]any{}},
		{"invalid generator", map[string]any{"gen": true}},
		{"index wrong type", map[string]any{"gen": nil, "index": 1}},
		{"index null", map[string]any{"gen": nil, "index": nil}},
		{"index empty", map[string]any{"gen": nil, "index": ""}},
		{"index invalid", map[string]any{"gen": nil, "index": "abc"}},
		{"index fractional", map[string]any{"gen": nil, "index": "1.5"}},
		{"seed wrong type", map[string]any{"gen": nil, "seed": 1}},
		{"seed null", map[string]any{"gen": nil, "seed": nil}},
		{"seed empty", map[string]any{"gen": nil, "seed": ""}},
		{"seed invalid", map[string]any{"gen": nil, "seed": "abc"}},
		{"seed overflow", map[string]any{"gen": nil, "seed": "9223372036854775808"}},
		{"seed underflow", map[string]any{"gen": nil, "seed": "-9223372036854775809"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if sampler, err := ImportSampler(test.value); err == nil || sampler != nil {
				t.Fatalf("got (%v, %v), want nil sampler and error", sampler, err)
			}
		})
	}
}

func TestExportImport(t *testing.T) {
	for _, test := range []struct {
		name string
		gen  Generator
	}{
		{"fixed", NewFixed([]byte("hello"))},
		{"fixed binary", NewFixed([]byte{0, '\n', '"', '\\', 0xff})},
		{"fixed empty", NewFixed([]byte{})},
		{"fixed nil", NewFixed(nil)},
		{"charset", NewCharset('a', 'z', 'A', 'Z', '0', '9')},
		{"choice", NewChoice(NewFixed([]byte("hello")), NewFixed([]byte("world")))},
		{"concat", NewConcat(NewFixed([]byte("prefix-")), NewCharset('a', 'z'))},
		{"numeric defaults", NewNumeric(10, 0, 100, 1, false)},
		{"numeric base step padding", NewNumeric(16, 8, 256, 2, true)},
		{"numeric original pattern", MustCompile(`<!64/1/100>`)},
		{"repeat fixed", NewRepeat(NewCharset('a', 'z'), 3, 3)},
		{"repeat variable", NewRepeat(NewCharset('a', 'z'), 0, 3)},
		{"nested", MustCompile(`prefix-(hello|[a-z]{2,4})-<!16:a/ff/2>`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			want := test.gen.String()
			for _, path := range []string{"direct", "json"} {
				t.Run(path, func(t *testing.T) {
					exported := Export(test.gen)
					if path == "json" {
						data, err := json.Marshal(exported)
						if err != nil {
							t.Fatalf("marshal exported generator: %v", err)
						}
						// Decode into an empty interface so the original Go types
						// cannot influence the JSON round trip.
						exported = nil
						if err := json.Unmarshal(data, &exported); err != nil {
							t.Fatalf("unmarshal exported generator: %v", err)
						}
					}

					imported, err := Import(exported)
					if err != nil {
						t.Fatalf("import exported value %#v: %v", exported, err)
					}
					if got := imported.String(); got != want {
						t.Errorf("string after round trip = %q, want %q", got, want)
					}
				})
			}
		})
	}
}

func TestReadInt(t *testing.T) {
	type namedInt int64
	type namedFloat float64
	for _, value := range []any{
		int(42), int8(42), int16(42), int32(42), int64(42),
		uint(42), uint8(42), uint16(42), uint32(42), uint64(42),
		float32(42), float64(42), namedInt(42), namedFloat(42),
		json.Number("42"), json.Number("42.0"), json.Number("4.2e1"),
	} {
		for _, want := range []any{int(42), int64(42), uint64(42), namedInt(42)} {
			t.Run(fmt.Sprintf("%T(%v) to %T", value, value, want), func(t *testing.T) {
				ptr := reflect.New(reflect.TypeOf(want))
				if err := readInt(ptr.Interface(), value); err != nil {
					t.Fatal(err)
				}
				if got := ptr.Elem().Interface(); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v, want %v", got, want)
				}
			})
		}
	}

	for _, test := range []struct {
		name  string
		value any
		want  any
	}{
		{"negative", int64(-42), int(-42)},
		{"min int64", json.Number("-9223372036854775808"), int64(math.MinInt64)},
		{"max int64", uint64(math.MaxInt64), int64(math.MaxInt64)},
		{"max uint64", json.Number("18446744073709551615"), uint64(math.MaxUint64)},
		{"max uint64 integer", uint64(math.MaxUint64), uint64(math.MaxUint64)},
		{"beyond float precision", json.Number("9007199254740993"), uint64(9007199254740993)},
		{"negative float boundary", float64(math.MinInt64), int64(math.MinInt64)},
		{"large unsigned float", math.Ldexp(1, 63), uint64(1 << 63)},
		{"negative zero", math.Copysign(0, -1), uint64(0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			ptr := reflect.New(reflect.TypeOf(test.want))
			if err := readInt(ptr.Interface(), test.value); err != nil {
				t.Fatal(err)
			}
			if got := ptr.Elem().Interface(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestReadIntInvalid(t *testing.T) {
	for _, test := range []struct {
		name    string
		value   any
		initial any
	}{
		{"nil", nil, int(7)},
		{"string", "42", int(7)},
		{"bool", true, int(7)},
		{"fraction", 1.5, int(7)},
		{"negative unsigned", -1, uint64(7)},
		{"negative unsigned float", -1.0, uint64(7)},
		{"NaN", math.NaN(), int(7)},
		{"positive infinity", math.Inf(1), int(7)},
		{"negative infinity", math.Inf(-1), int(7)},
		{"signed overflow", uint64(math.MaxUint64), int64(7)},
		{"int overflow", math.Ldexp(1, strconv.IntSize-1), int(7)},
		{"signed float overflow", math.Ldexp(1, 63), int64(7)},
		{"signed float underflow", math.Nextafter(float64(math.MinInt64), math.Inf(-1)), int64(7)},
		{"unsigned float overflow", math.Ldexp(1, 64), uint64(7)},
		{"narrow overflow", 128, int8(7)},
		{"number unsigned overflow", json.Number("18446744073709551616"), uint64(7)},
		{"number signed underflow", json.Number("-9223372036854775809"), int64(7)},
		{"number fraction", json.Number("1.01"), int(7)},
		{"number precise fraction", json.Number("1.0000000000000000001"), int(7)},
		{"number fractional exponent", json.Number("1e-3"), int(7)},
		{"number invalid", json.Number("garbage"), int(7)},
		{"number non-JSON syntax", json.Number("0x10"), int(7)},
	} {
		t.Run(test.name, func(t *testing.T) {
			ptr := reflect.New(reflect.TypeOf(test.initial))
			ptr.Elem().Set(reflect.ValueOf(test.initial))
			if err := readInt(ptr.Interface(), test.value); err == nil {
				t.Fatal("expected conversion error")
			}
			if got := ptr.Elem().Interface(); !reflect.DeepEqual(got, test.initial) {
				t.Fatalf("failed conversion changed destination to %v", got)
			}
		})
	}
	for _, ptr := range []any{nil, (*int)(nil), 0, new(string), new(float64)} {
		if err := readInt(ptr, 42); err == nil {
			t.Errorf("expected error for destination %T", ptr)
		}
	}
}

func TestImportJSONNumber(t *testing.T) {
	gen := NewNumeric(10, 9007199254740993, 9007199254741001, 2, true)
	data, err := json.Marshal(Export(gen))
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	var exported any
	if err := decoder.Decode(&exported); err != nil {
		t.Fatal(err)
	}
	imported, err := Import(exported)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Export(imported), Export(gen); !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
}
