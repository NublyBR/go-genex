# 💠 go-genex

> [!NOTE]
> This project is under development. Breaking changes are to be expected.

Compile regex-inspired patterns into fast generators for producing matching values.

Compiled generators sample efficiently, making `go-genex` a good fit for tests, fixtures, fuzzing, and synthetic data generation.

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

Install package:

```sh
go get github.com/NublyBR/go-genex
```

Install CLI:

```sh
go install github.com/NublyBR/go-genex/genex@latest
```

## Randomness

By default, `go-genex` uses a fast non-cryptographic PRNG for performance.

If you need values generated with cryptographically secure randomness, provide a secure RNG explicitly:

```go
gen, err := genex.Compile("...", genex.OptionRNG(genex.SecureRand))
if err != nil {
    panic(err)
}
```

## Features

### Combinatronics

Each compiled generator can give you useful information about itself.

* `min, max = gen.Bounds()` - Get the minimum and maximum size of generated strings.
* `count = gen.Count()` - Get a `*big.Int` representing how many values exist in the generator's search-space.
* `complexity = gen.Complexity()` - Get how many nodes are in the AST. Does not necessarily reflect the time complexity to generate values.

### Fast sampling

Compiled generators can be sampled efficiently, making `go-genex` suitable for tests, fuzzing inputs, fixtures, and synthetic data generation.

### Zero-allocation generation

Sampling can be done without heap allocations when writing into a reused buffer.

```text
goos: linux
goarch: amd64
pkg: github.com/NublyBR/go-genex
cpu: Intel(R) Core(TM) i5-9600K CPU @ 3.70GHz
=== RUN   BenchmarkRandom
BenchmarkRandom-6        5525199               211.4 ns/op             0 B/op          0 allocs/op
=== RUN   BenchmarkIter
BenchmarkIter-6         11856430               110.7 ns/op             0 B/op          0 allocs/op
```

> Keep in mind that while the generator itself is allocation-free, calling `buf.String()` allocates memory on the heap.
> You may use `buf.Bytes()` to get the underneath buffer and use that, but only for as long as you do not call `iter.Step()`.

## Use cases

* Generating test data from structured patterns
* Producing reproducible synthetic identifiers
* Filling templates with constrained random values
* Fuzzing parsers and validators with valid-shaped inputs

## Pattern syntax

| Characters                                                | Meaning                                                                                                                                                                      |
| --------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `[xyz]`<br>`[a-c]`<br>`[\n\x1f]`                          | **Character class:** Generates one character from the set. Supports literal characters, ranges, and hexadecimal escape sequences.                                            |
| `(hello)`                                                 | **Group:** Groups expressions together, mainly for composition with repeaters such as `(...){32}`.                                                                           |
| `a\|b`<br>`(a\|b)`                                        | **Disjunction:** Generates one of multiple alternatives.                                                                                                                     |
| `<64>`<br>`<base:start/end>`<br>`<start/end>`<br>`<!end>` | **Numeric expression:** Generates numbers from a fixed value or numeric range. Supports explicit base selection and optional zero-padding to the width of the maximum value. |
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

// Generate a strong password using all visible ASCII characters
`.{64}`
```

## Non-goals

* Not intended for regex matching
* No backreferences
* No lookahead / lookbehind
* Not PCRE-compatible