# 💠 go-genex

> [!NOTE]
> This project is under development. Breaking changes are to be expected.

Compile regex-inspired patterns into fast generators for producing matching values.

Compiled generators support random sampling, exhaustive iteration, and direct indexing. Seeded samplers produce reproducible sequences, and generators and sampler positions can be exported for later use.

> `go-genex` is not a full regex engine. It implements a regex-inspired language designed for efficient generation rather than matching.

## Example

```go
gen, err := genex.Compile(`Sample-[a-zA-Z0-9]{32}`)
if err != nil {
    panic(err)
}

_, maxSize := gen.Bounds()

buf := bytes.NewBuffer(make([]byte, 0, maxSize))

for i := 0; i < 8; i++ {
    gen.Sample(buf)
    fmt.Println(buf.String())
    buf.Reset()
}
```

Example output:

```text
Sample-5nbDamOZltDlWvnvfdeXabxxCTirtUsC
Sample-OxDucBaSPF2bea7lflfYe1mO4VKHEdEp
Sample-xBEA8MWcmKRQJ7cgdZLMREpw8M8KyKzj
Sample-oXbnHimbOsJzRDvytjXoAwCyhgHAF2wi
Sample-cbRCSmqMQ149HAzJN5ka78QqWu5a4KOi
Sample-tUzistAbUxxRMIue8enrYMYDmyy3xRI3
Sample-x0iAoGpWQi74TpC9GdxZVUOp9WM3gtnm
Sample-zdeO6h8MysaVPfPT8Fa3KVq5958Q2R2N
```

## Installation

Requires Go 1.24.2 or later.

Install package:

```sh
go get github.com/NublyBR/go-genex
```

Install CLI:

```sh
go install github.com/NublyBR/go-genex/genex@latest
```

## Randomness

By default, `gen.Sample` uses a fast non-cryptographic PRNG for performance.

If you need samples generated with cryptographically secure randomness, provide a secure RNG explicitly:

```go
gen, err := genex.Compile("...", genex.OptionRNG(genex.SecureRand))
if err != nil {
    panic(err)
}
```

## Features

### Combinatorics

Each compiled generator can give you useful information about itself.

* `min, max = gen.Bounds()` - Get the minimum and maximum size of generated strings in bytes.
* `count = gen.Count()` - Get a `*big.Int` representing how many positions exist in the generator's search space. Overlapping alternatives can produce the same value at different positions.
* `complexity = gen.Complexity()` - Get how many nodes are in the AST. Does not necessarily reflect the time complexity to generate values.

### Fast sampling

Compiled generators can be sampled efficiently, making `go-genex` suitable for tests, fuzzing inputs, fixtures, and synthetic data generation.

`gen.Sample(buf)` appends a random value to the buffer on each call. Values may repeat. Reset the buffer between samples as shown in the example above.

### Exhaustive iteration

Use `gen.Iterate()` to visit every result in order:

```go
gen := genex.MustCompile(`[a-c][0-1]`)

for value := range gen.Iterate() {
    fmt.Println(string(value))
}
// a0, a1, b0, b1, c0, c1
```

`Iterate()` returns an `iter.Seq[[]byte]`. Each sequence keeps its position; call `gen.Iterate()` again to start over. The yielded bytes share a buffer, so copy them with `bytes.Clone(value)` if you want to keep them after the next iteration.

### Direct indexing

Use `gen.Index(buf, idx)` to generate a value at a specific position without visiting the preceding values. Indices use `*big.Int`, so they can address search spaces larger than a 64-bit integer.

```go
gen := genex.MustCompile(`<255>\.<255>\.<255>\.<255>`)
idx := big.NewInt(0x04_03_02_01)
var buf bytes.Buffer

gen.Index(&buf, idx)
fmt.Println(buf.String()) // 1.2.3.4
```

Indices are zero-based. `Index` appends to the buffer and modifies `idx`, leaving the quotient after division by `gen.Count()`. For an index in `[0, gen.Count())`, the remaining value is zero. Larger nonnegative indices wrap around the search space. Pass `new(big.Int).Set(idx)` if you need to preserve the original index.

In concatenations and fixed repetitions, the leftmost component changes fastest. This differs from the order produced by `Iterate()`.

### Reproducible samplers

Use `genex.NewSampler(gen, seed)` to visit generator indices in a reproducible pseudo-random order. A sampler visits each index once without storing the entire sequence, making it useful even for patterns such as `.{512}`.

```go
gen := genex.MustCompile(`[a-z]{3}`)
sampler := genex.NewSampler(gen, 42)
var buf bytes.Buffer

for i := 0; i < 8 && sampler.Next(&buf); i++ {
    fmt.Println(buf.String())
    buf.Reset()
}
```

The same generator and seed produce the same sequence. `Next` appends a value and advances `sampler.Index`. At the end of the sequence it returns `false` without changing the buffer or index. Overlapping alternatives can still produce duplicate values, since different indices may represent the same output.

Set `sampler.Index` to seek to a position, or use `sampler.Index.SetInt64(0)` to replay the sequence. A sampler's position refers to its shuffled sequence; it is different from the index passed to `gen.Index`.

### Exporting and importing generators

`genex.Export(gen)` converts a generator into byte slices, numbers, booleans, lists, and maps that can be serialized as JSON, YAML, or another suitable format. This is useful for storing large expressions without managing them as pattern strings.

```go
gen := genex.MustCompile(`user-(admin|staff|guest)-<100/999>`)

data, err := json.Marshal(genex.Export(gen))
if err != nil {
    panic(err)
}

decoder := json.NewDecoder(bytes.NewReader(data))
decoder.UseNumber()

var exported any
if err := decoder.Decode(&exported); err != nil {
    panic(err)
}

restored, err := genex.Import(exported)
if err != nil {
    panic(err)
}
fmt.Println(restored.String() == gen.String()) // true
```

Use `decoder.UseNumber()` to preserve large integers when decoding JSON. Default decoding into `any` uses `float64`, which can lose precision. For other formats, decode objects as `map[string]any` and lists as `[]any`, and preserve byte slices or encode them as base64 strings.

You can also pass the exported value directly to `genex.Import(genex.Export(gen))`. Custom RNG functions are not serialized; `Import` accepts options such as `genex.OptionRNG(genex.SecureRand)` to configure the restored generator.

### Saving and restoring samplers

`genex.ExportSampler(sampler)` exports the generator, seed, and current position so you can resume the sequence later. `genex.ImportSampler` accepts the exported object directly or after serialization.

```go
sampler := genex.NewSampler(genex.MustCompile(`[a-z]{3}`), 42)
var buf bytes.Buffer
sampler.Next(&buf) // Consume the first value before saving.
buf.Reset()

data, err := json.Marshal(genex.ExportSampler(sampler))
if err != nil {
    panic(err)
}

decoder := json.NewDecoder(bytes.NewReader(data))
decoder.UseNumber()

var exported any
if err := decoder.Decode(&exported); err != nil {
    panic(err)
}

restored, err := genex.ImportSampler(exported)
if err != nil {
    panic(err)
}
if restored.Next(&buf) {
    fmt.Println(buf.String()) // The next value in the saved sequence.
}
```

The exported index and seed are decimal strings to preserve their precision. `ImportSampler` also accepts generator options, just like `Import`.

### Zero-allocation generation

Random sampling with `gen.Sample(buf)` can be done without heap allocations when writing into a reused buffer with enough capacity. Indexing, seeded samplers, and serialization may allocate.

Example benchmark results for random sampling and iteration:

```text
goos: linux
goarch: amd64
pkg: github.com/NublyBR/go-genex
cpu: Intel(R) Core(TM) i5-9600K CPU @ 3.70GHz
=== RUN   BenchmarkRandom
BenchmarkRandom-6        3617576               323.5 ns/op             0 B/op          0 allocs/op
=== RUN   BenchmarkIter
BenchmarkIter-6         12079502               109.5 ns/op             0 B/op          0 allocs/op
```

> Converting a buffer to a string can allocate. `buf.Bytes()` shares the buffer's storage and is only valid until the buffer is modified.

Run the benchmarks on your machine:

```sh
go test -run '^$' -bench 'Benchmark(Random|Iter)$' -benchmem
```

## Use cases

* Generating test data from structured patterns
* Producing reproducible synthetic identifiers
* Filling templates with constrained random values
* Fuzzing parsers and validators with valid-shaped inputs
* Saving and resuming reproducible generation jobs

## Pattern syntax

| Characters                                                | Meaning                                                                                                                                                                      |
| --------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `[xyz]`<br>`[a-c]`<br>`[\n\x1f]`                          | **Character class:** Generates one character from the set. Supports literal characters, ranges, and hexadecimal escape sequences.                                            |
| `.` | **Visible ASCII:** Generates one byte from `!` through `~`. |
| `(hello)`                                                 | **Group:** Groups expressions together, mainly for composition with repeaters such as `(...){32}`.                                                                           |
| `a\|b`<br>`(a\|b)`                                        | **Disjunction:** Generates one of multiple alternatives.                                                                                                                     |
| `<64>`<br>`<base:start/end>`<br>`<start/end/step>`<br>`<!end>` | **Numeric expression:** Generates numbers from an inclusive range (`<64>` means 0 through 64). Supports explicit base selection, a step, and optional zero-padding. |
| `...{64}`<br>`...{32,64}`                                 | **Repeater:** Repeats the previous value `n` times. Also accepts a minimum and maximum value.                                                                                |
| `x*`<br>`x+`<br>`x?`                                      | **Quantifiers:** Shorthand for specific repeater setups.<br>`x*` = `x{0,8}`<br>`x+` = `x{1,8}`<br>`x?` = `x{0,1}`                                                            |

### Syntax examples

```go
// Generate padded hexadecimal numbers from 0000 to ffff
`<!16:ffff>`

// Generate unpadded octal numbers from 0 to 777
`<8:777>`

// Generate sample usernames such as 'user-admin-159'
`user-(admin|staff|guest)-<100/999>`

// Generate IPv4 addresses
`(<255>\.){3}<255>`

// Generate IPv6 addresses
`(<!16:ffff>:){7}<!16:ffff>`

// Generate UUIDv4's
`\h{8}-\h{4}-4\h{3}-[89ab]\h{3}-\h{12}`

// Generate a 64-character value using all visible ASCII characters
`.{64}`
```

## Non-goals

* Not intended for regex matching
* No backreferences
* No lookahead / lookbehind
* Not PCRE-compatible
