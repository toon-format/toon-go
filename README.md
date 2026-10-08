# TOON for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/toon-format/toon-go.svg)](https://pkg.go.dev/github.com/toon-format/toon-go)
[![SPEC v4.4](https://img.shields.io/badge/spec-v4.4-lightgrey)](https://github.com/toon-format/spec/blob/v4.4.0/SPEC.md)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

Encodes Go values to [TOON (Token-Oriented Object Notation)](https://github.com/toon-format/toon) and decodes TOON back. TOON is a compact, indentation-based encoding of the JSON data model for LLM input.

## Installation

```bash
go get github.com/toon-format/toon-go
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/toon-format/toon-go"
)

type User struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
	Role string `toon:"role"`
}

type Payload struct {
	Users []User `toon:"users"`
}

func main() {
	in := Payload{Users: []User{{1, "Ada", "admin"}, {2, "Bob", "user"}}}

	encoded, err := toon.MarshalString(in)
	if err != nil {
		panic(err)
	}
	fmt.Println(encoded)
	// users[2]{id,name,role}:
	//   1,Ada,admin
	//   2,Bob,user

	var out Payload
	if err := toon.UnmarshalString(encoded, &out); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", out)
	// {Users:[{ID:1 Name:Ada Role:admin} {ID:2 Name:Bob Role:user}]}
}
```

Without a destination type, `Decode` and `DecodeString` return `map[string]any`, `[]any`, and primitives. Pass encoder options to `Marshal` and decoder options to `Unmarshal` or `Decode`:

| Option | Default | Description |
| ------ | ------- | ----------- |
| `WithIndent(n)` | `2` | Spaces per indentation level when encoding |
| `WithDelimiter(d)` | `DelimiterComma` | Delimiter for array values and tabular rows: `DelimiterComma`, `DelimiterTab`, or `DelimiterPipe` |
| `WithTimeFormatter(f)` | RFC 3339 in UTC | Formats `time.Time` values |
| `WithStrictMode(b)` | `true` | `false` accepts declared counts that don't match, duplicate keys (the last wins), tab or misaligned indentation, blank lines inside an array, and a block indented too deep; every other decoding error stays |
| `WithDecoderIndent(n)` | `2` | Expected spaces per indentation level when decoding |

## Specification

Targets [TOON spec v4.4](https://github.com/toon-format/spec/blob/v4.4.0/SPEC.md), and the test suite runs that version's conformance fixtures.

- **Numbers decode to `float64`** – integers beyond 2^53 lose precision and a token that overflows `float64` (e.g. `1e999`) decodes as a string; on encode, integers beyond ±(2^53 − 1) and `*big.Int` values become quoted decimal strings ([§4](https://github.com/toon-format/spec/blob/v4.4.0/SPEC.md#4-decoding-interpretation-reference-decoder))
- **Structs encode as objects keyed by their `toon` tags** – `toon:"name"` renames, `toon:"name,omitempty"` skips zero values, `toon:"-"` skips the field, a field without a `toon` tag falls back to its `json` tag, and an untagged field uses its Go name; maps need string keys, `time.Time` and `fmt.Stringer` values become strings, and `NaN` and `±Inf` become `null` ([§3](https://github.com/toon-format/spec/blob/v4.4.0/SPEC.md#3-encoding-normalization-reference-encoder))
- **Decoded objects are `map[string]any`, so document key order is lost** – on encode, `toon.Object` keeps its field order and Go maps are written in sorted key order ([§2](https://github.com/toon-format/spec/blob/v4.4.0/SPEC.md#2-data-model))

## Resources

- **Specification:** [SPEC.md](https://github.com/toon-format/spec/blob/main/SPEC.md) – Normative rules and conformance checklists
- **Format Overview:** [toonformat.dev](https://toonformat.dev/guide/format-overview) – Every form with examples
- **Other Implementations:** [toonformat.dev](https://toonformat.dev/ecosystem/implementations) – TOON in other languages
- **API Reference:** [pkg.go.dev](https://pkg.go.dev/github.com/toon-format/toon-go) – Every exported function and option

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the development setup and pull request guidelines.

## License

[MIT](./LICENSE) License © 2025-PRESENT [Bintang Pradana Erlangga Putra](https://github.com/bpradana) and [Johann Schopplich](https://github.com/johannschopplich)
