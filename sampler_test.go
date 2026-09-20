package genex

import (
	"bytes"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"testing"
)

func TestSampler(t *testing.T) {
	for _, test := range []struct {
		name, pattern string
		want          []string
	}{
		{"empty value", ``, []string{""}},
		{"singleton", `fixed`, []string{"fixed"}},
		{"power of two", `[a-p]`, strings.Split("abcdefghijklmnop", "")},
		{"just above power of two", `[a-q]`, strings.Split("abcdefghijklmnopq", "")},
		{"alphabet", `[a-z]`, strings.Split("abcdefghijklmnopqrstuvwxyz", "")},
		{"unequal branches", `(x|[ab])`, []string{"x", "a", "b"}},
		{"variable repeat", `[ab]{0,2}`, []string{"", "a", "b", "aa", "ab", "ba", "bb"}},
		// Sampling permutes indices, so overlapping alternatives retain duplicates.
		{"overlapping branches", `(a[bc]|[ab]c)`, []string{"ab", "ac", "ac", "bc"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			gen := MustCompile(test.pattern)
			want := slices.Clone(test.want)
			slices.Sort(want)
			for _, seed := range []int64{0, 1, -1} {
				t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
					sampler := NewSampler(gen, seed)
					samples := samplerSamples(t, sampler, len(want))
					assertSamplerExhausted(t, sampler)
					ordered := slices.Clone(samples)
					slices.Sort(ordered)
					if !slices.Equal(ordered, want) {
						t.Fatalf("sorted samples = %q, want %q", ordered, want)
					}

					sampler.Index.SetInt64(0)
					if again := samplerSamples(t, sampler, len(want)); !slices.Equal(again, samples) {
						t.Fatalf("reset sequence = %q, want %q", again, samples)
					}
					fresh := NewSampler(MustCompile(test.pattern), seed)
					if again := samplerSamples(t, fresh, len(want)); !slices.Equal(again, samples) {
						t.Fatalf("fresh sequence = %q, want %q", again, samples)
					}

					mid := len(want) / 2
					sampler.Index.SetInt64(int64(mid))
					if suffix := samplerSamples(t, sampler, len(want)-mid); !slices.Equal(suffix, samples[mid:]) {
						t.Fatalf("seek sequence = %q, want %q", suffix, samples[mid:])
					}
					assertSamplerExhausted(t, sampler)
				})
			}
		})
	}
}

func TestSamplerSeeds(t *testing.T) {
	gen := MustCompile(`[a-z]`)
	a := samplerSamples(t, NewSampler(gen, 0), 26)
	b := samplerSamples(t, NewSampler(gen, 1), 26)
	if slices.Equal(a, b) {
		t.Fatal("seeds 0 and 1 produced the same alphabet sequence")
	}
}

func TestSamplerBounds(t *testing.T) {
	sampler := NewSampler(MustCompile(`[a-z]`), 0)
	for _, index := range []int64{-1, 26, 27} {
		t.Run(fmt.Sprintf("index=%d", index), func(t *testing.T) {
			sampler.Index.SetInt64(index)
			assertSamplerExhausted(t, sampler)
		})
	}
	sampler.Index.SetInt64(0)
	samplerSamples(t, sampler, 26)
	assertSamplerExhausted(t, sampler)
}

func TestSamplerLargeIndices(t *testing.T) {
	// 600 bits also exercises expansion beyond a single round-function digest.
	for _, bits := range []int{65, 600} {
		t.Run(fmt.Sprintf("bits=%d", bits), func(t *testing.T) {
			gen := MustCompile(fmt.Sprintf(`[ab]{%d}`, bits))
			sampler := NewSampler(gen, -42)
			start := new(big.Int).Lsh(big.NewInt(1), 64)
			sampler.Index.Set(start)
			samples := samplerSamples(t, sampler, 32)
			seen := make(map[string]bool, len(samples))
			for _, sample := range samples {
				if len(sample) != bits || strings.Trim(sample, "ab") != "" {
					t.Fatalf("invalid sample %q", sample)
				}
				if seen[sample] {
					t.Fatalf("duplicate sample %q", sample)
				}
				seen[sample] = true
			}
			fresh := NewSampler(gen, -42)
			fresh.Index.Set(start)
			if again := samplerSamples(t, fresh, 32); !slices.Equal(again, samples) {
				t.Fatal("large-index sequence is not reproducible")
			}
			sampler.Index.Sub(gen.Count(), big.NewInt(3))
			samplerSamples(t, sampler, 3)
			assertSamplerExhausted(t, sampler)
		})
	}
}

func TestSamplerHugeCount(t *testing.T) {
	gen := MustCompile(`.{512}`)
	count := gen.Count()
	wantCount := new(big.Int).Exp(big.NewInt(94), big.NewInt(512), nil)
	if count.Cmp(wantCount) != 0 {
		t.Fatalf("count = %s, want 94^512", count)
	}
	t.Logf("search space: %d-bit count (%d decimal digits)", count.BitLen(), len(count.String()))

	for _, test := range []struct {
		name  string
		start *big.Int
	}{
		{"start", new(big.Int)},
		{"middle", new(big.Int).Rsh(count, 1)},
		{"end", new(big.Int).Sub(count, big.NewInt(8))},
	} {
		t.Run(test.name, func(t *testing.T) {
			sampler := NewSampler(gen, 42)
			sampler.Index.Set(test.start)
			samples := samplerSamples(t, sampler, 8)
			seen := make(map[string]bool, len(samples))
			for _, sample := range samples {
				if len(sample) != 512 {
					t.Fatalf("sample length = %d, want 512", len(sample))
				}
				for i := range sample {
					if sample[i] < '!' || sample[i] > '~' {
						t.Fatalf("invalid byte %q at offset %d", sample[i], i)
					}
				}
				if seen[sample] {
					t.Fatal("duplicate sample")
				}
				seen[sample] = true
			}
			fresh := NewSampler(gen, 42)
			fresh.Index.Set(test.start)
			if again := samplerSamples(t, fresh, 8); !slices.Equal(again, samples) {
				t.Fatal("huge-count sequence is not reproducible")
			}
			if test.name == "end" {
				assertSamplerExhausted(t, sampler)
			}
		})
	}
}

func samplerSamples(t *testing.T, sampler *Sampler, count int) []string {
	t.Helper()
	samples := make([]string, 0, count)
	wantIndex := new(big.Int).Set(&sampler.Index)
	for range count {
		buf := bytes.NewBufferString("prefix:")
		if !sampler.Next(buf) {
			t.Fatalf("sampler finished early at index %s", &sampler.Index)
		}
		wantIndex.Add(wantIndex, big.NewInt(1))
		if sampler.Index.Cmp(wantIndex) != 0 {
			t.Fatalf("index = %s, want %s", &sampler.Index, wantIndex)
		}
		sample, ok := strings.CutPrefix(buf.String(), "prefix:")
		if !ok {
			t.Fatal("Next overwrote the existing buffer contents")
		}
		samples = append(samples, sample)
	}
	return samples
}

func assertSamplerExhausted(t *testing.T, sampler *Sampler) {
	t.Helper()
	index := new(big.Int).Set(&sampler.Index)
	buf := bytes.NewBufferString("untouched")
	for range 2 {
		if sampler.Next(buf) {
			t.Fatalf("Next returned true at invalid index %s", index)
		}
		if buf.String() != "untouched" || sampler.Index.Cmp(index) != 0 {
			t.Fatalf("exhaustion changed buffer or index: (%q, %s)", buf.String(), &sampler.Index)
		}
	}
}
