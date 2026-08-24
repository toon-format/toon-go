# TOON Format for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/toon-format/toon-go.svg)](https://pkg.go.dev/github.com/toon-format/toon-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/toon-format/toon-go)](https://goreportcard.com/report/github.com/toon-format/toon-go)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

**Token-Oriented Object Notation** is a compact, human-readable format designed for passing structured data to Large Language Models with significantly reduced token usage.

`toon-spec: 4.1` — this implementation targets specification v4.1 and passes the full conformance fixture suite of [toon-format/spec](https://github.com/toon-format/spec) v4.1.1.

## Example

**JSON** (verbose):
```json
{
  "users": [
    { "id": 1, "name": "Alice", "role": "admin" },
    { "id": 2, "name": "Bob", "role": "user" }
  ]
}
```

**TOON** (compact):
```
users[2]{id,name,role}:
  1,Alice,admin
  2,Bob,user
```

## Specification Coverage

| Feature | Section | Status |
| --- | --- | --- |
| Inline primitive arrays, list form, tabular form | §9.1–§9.4 | supported |
| Nested field groups: `orders[2]{id,customer{name,country},total}:` | §6, §9.3 | supported |
| Keyed tabular form: `users[2:]{age,city}:` with one entry row per key | §6, §9.5 | supported |
| Comment lines removed in a lexical pre-pass | §5.1 | supported |
| Canonical number form, decoder number grammar | §2, §4 | supported |
| Explicit empty arrays `key: []` / `[]` (legacy `key[0]:` accepted) | §9.1 | supported |
| Byte-order mark removal, CRLF input, trailing-space stripping | §12 | supported |
| Strict-mode diagnostics, duplicate-key last-write-wins | §14 | supported |
| Comma, tab, and pipe delimiters | §11 | supported |

Two features are deliberately absent because the specification removed them:

- `[#N]` length markers were removed in spec 2.0. `WithLengthMarkers` is retained as a deprecated no-op and the decoder rejects `[#N]`.
- Key folding and path expansion (`keyFolding`, `flattenDepth`, `expandPaths`) were removed in spec 4.0. Dotted keys are single literal keys. Flattening nested objects is now expressed by nested field groups in tabular and keyed tabular headers (§9.3, §9.5), shown below.

### Implementation-defined behavior

The specification requires these choices to be documented:

- **Key order.** Decoded objects are `map[string]any`, which does not retain insertion order, so document key order is not preserved on decode (§2). Encoding preserves the encounter order of `toon.Object` fields; Go maps are encoded in sorted key order.
- **Numeric domain.** Numbers decode to `float64`. A token outside that domain decodes to its nearest `float64`. On encode, integers beyond IEEE 754 exact range are emitted as quoted plain-decimal strings (§2).
- **Tabs in indentation.** Rejected in strict mode. In non-strict mode each leading tab counts as one indentation level (§12).

## Nested Objects in Tabular Form

An array of uniform objects whose columns are themselves uniform objects declares
nested field groups once in the header; the rows stay flat:

```go
doc, _ := toon.MarshalString(payload)
```

```
orders[2]{id,customer{name,country},total}:
  1,Ada,UK,9.5
  2,Bob,ES,14
```

An object whose values are uniform objects collapses into keyed tabular form,
where each entry carries its own key:

```
users[2:]{age,city}:
  alice: 30,Madrid
  bob: 41,Lisboa
```

## Usage

### Marshal and Unmarshal

```go
package main

import (
    "fmt"

    "github.com/toon-format/toon-go"
)

type User struct {
    ID    int    `toon:"id"`
    Name  string `toon:"name"`
    Role  string `toon:"role"`
}

type Payload struct {
    Users []User `toon:"users"`
}

func main() {
    in := Payload{
        Users: []User{
            {ID: 1, Name: "Alice", Role: "admin"},
            {ID: 2, Name: "Bob", Role: "user"},
        },
    }

    encoded, err := toon.Marshal(in)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(encoded))

    var out Payload
    if err := toon.Unmarshal(encoded, &out); err != nil {
        panic(err)
    }
    fmt.Printf("first user: %+v\n", out.Users[0])
}
```

### Unmarshal into Maps

`Unmarshal` can populate dynamic maps, mimicking the `encoding/json` package:

```go
var doc map[string]any
if err := toon.Unmarshal(encoded, &doc); err != nil {
    panic(err)
}
fmt.Printf("users: %#v\n", doc["users"])
```

### Decode Without Structs

If you do not have a destination struct, use `Decode` for a dynamic representation:

```go
package main

import (
    "fmt"
    "github.com/toon-format/toon-go"
)

func main() {
    raw := []byte("users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user\n")
    decoded, err := toon.Decode(raw)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", decoded)
}
```

For more runnable samples, explore the programs in `./examples`.

## Resources

- [TOON Specification](https://github.com/toon-format/spec/blob/main/SPEC.md)
- [Main Repository](https://github.com/toon-format/toon)
- [Benchmarks & Performance](https://github.com/toon-format/toon#benchmarks)
- [Other Language Implementations](https://github.com/toon-format/toon#other-implementations)

## Contributing

Interested in implementing TOON for Go? Check out the [specification](https://github.com/toon-format/spec/blob/main/SPEC.md) and feel free to contribute!

## License

MIT License © 2025-PRESENT [Johann Schopplich](https://github.com/johannschopplich)
