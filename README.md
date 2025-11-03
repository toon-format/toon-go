# TOON Format for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/toon-format/toon-go.svg)](https://pkg.go.dev/github.com/toon-format/toon-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/toon-format/toon-go)](https://goreportcard.com/report/github.com/toon-format/toon-go)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

**Token-Oriented Object Notation** is a compact, human-readable format designed for passing structured data to Large Language Models with significantly reduced token usage.

## Status

🚧 **This package is currently a namespace reservation.** Full implementation coming soon!

### Example

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

## Resources

- [TOON Specification](https://github.com/toon-format/spec/blob/main/SPEC.md)
- [Main Repository](https://github.com/toon-format/toon)
- [Benchmarks & Performance](https://github.com/toon-format/toon#benchmarks)
- [Other Language Implementations](https://github.com/toon-format/toon#other-implementations)

## Future Usage

Once implemented, the package will provide:

```go
package main

import (
    "fmt"
    "github.com/toon-format/toon-go"
)

func main() {
    data := map[string]interface{}{
        "users": []map[string]interface{}{
            {"id": 1, "name": "Alice", "role": "admin"},
            {"id": 2, "name": "Bob", "role": "user"},
        },
    }

    // Encode to TOON
    encoded, err := toon.Encode(data, nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(encoded)

    // Decode from TOON
    decoded, err := toon.Decode(encoded, nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", decoded)
}
```

## Contributing

Interested in implementing TOON for Go? Check out the [specification](https://github.com/toon-format/spec/blob/main/SPEC.md) and feel free to contribute!

## License

MIT License © 2025-PRESENT [Johann Schopplich](https://github.com/johannschopplich)
