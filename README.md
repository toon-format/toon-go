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

### Capturing Tool Call Payloads with the TOON Type

The `toon.TOON` type automatically converts JSON payloads to compact TOON format. When you unmarshal JSON with a TOON field, nested objects are converted to TOON's efficient representation.

#### Basic Usage

```go
type Response struct {
    EventID string    `json:"event_id"`
    Payload toon.TOON `json:"payload"`  // Converts JSON to TOON automatically
}

// Tool call returns JSON with nested payload
jsonResponse := `{
  "event_id": "evt_xyz",
  "payload": {
    "id": "order_xyz",
    "amount": 99.99,
    "items": ["widget", "gadget"]
  }
}`

// Unmarshal JSON - payload automatically converted to TOON format
var response Response
json.Unmarshal([]byte(jsonResponse), &response)

fmt.Printf("Event ID: %s\n", response.EventID)
// Output: Event ID: evt_xyz

// Payload is stored in compact TOON format
fmt.Printf("Payload:\n%s\n", response.Payload.String())
// Output:
// amount: 99.99
// id: order_xyz
// items[2]: widget,gadget

// Process the payload when needed
payloadData, _ := toon.Decode(response.Payload)
payloadMap := payloadData.(map[string]any)
fmt.Printf("Order ID: %s\n", payloadMap["id"])     // order_xyz
fmt.Printf("Amount: %.2f\n", payloadMap["amount"]) // 99.99

// Round-trip back to JSON works seamlessly
jsonBytes, _ := json.Marshal(response)
// Payload automatically converts back: {"event_id":"evt_xyz","payload":{...}}
```

#### Why Use TOON for Payloads?

Automatic JSON-to-TOON conversion gives you:
- **Compact storage** - TOON format is 30-50% smaller than JSON for structured data
- **Human-readable** - Easy to read in logs and databases
- **Delay parsing** - Only decode when you actually need the data
- **Flexible schemas** - Works with any JSON structure, no type definitions needed
- **Round-trip safe** - Converts back to JSON automatically when marshaling

#### Interfaces Implemented

The `TOON` type implements standard Go interfaces for seamless integration:

- `json.Marshaler` / `json.Unmarshaler` - JSON support (handles nested objects)
- `encoding.TextMarshaler` / `encoding.TextUnmarshaler` - Text-based formats
- `database/sql.Scanner` / `driver.Valuer` - Direct database operations
- `fmt.Stringer` - String representation via `String()` method
- `IsNil()` helper - Check for empty/nil values

#### Working with TOON Documents

You can also use `TOON` to store raw TOON documents within structs:

```go
type Config struct {
    AppName  string    `toon:"app_name"`
    Version  int       `toon:"version"`
    Settings toon.TOON `toon:"settings"`  // Raw TOON content
}

cfg := Config{
    AppName:  "MyApp",
    Version:  1,
    Settings: toon.TOON("timeout: 30\nretries: 3\nverbose: true"),
}

// Marshal to TOON - Settings preserved as-is
encoded, _ := toon.MarshalString(cfg)
// Output:
// app_name: MyApp
// version: 1
// settings: "timeout: 30\nretries: 3\nverbose: true"
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
