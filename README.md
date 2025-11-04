# TOON Format for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/toon-format/toon-go.svg)](https://pkg.go.dev/github.com/toon-format/toon-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/toon-format/toon-go)](https://goreportcard.com/report/github.com/toon-format/toon-go)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

**Token-Oriented Object Notation** is a compact, human-readable format designed for passing structured data to Large Language Models with significantly reduced token usage.

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

    encoded, err := toon.Marshal(in, toon.WithLengthMarkers(true))
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
