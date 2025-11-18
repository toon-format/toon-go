package main

import (
	"encoding/json"
	"fmt"

	"github.com/toon-format/toon-go"
)

type Config struct {
	AppName  string    `toon:"app_name"`
	Version  int       `toon:"version"`
	Settings toon.TOON `toon:"settings"`
}

func main() {
	fmt.Println("=== TOON Type Example ===")
	fmt.Println()

	// Example 1: Creating a Config with embedded TOON
	fmt.Println("1. Creating a Config with TOON field:")
	cfg := Config{
		AppName:  "MyApp",
		Version:  1,
		Settings: toon.TOON("timeout: 30\nretries: 3\nverbose: true"),
	}

	// Marshal to TOON
	toonDoc, err := toon.MarshalString(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Marshaled to TOON:\n%s\n\n", toonDoc)

	// Example 2: Unmarshaling back
	fmt.Println("2. Unmarshaling back to struct:")
	var decoded Config
	err = toon.UnmarshalString(toonDoc, &decoded)
	if err != nil {
		panic(err)
	}
	fmt.Printf("App Name: %s\n", decoded.AppName)
	fmt.Printf("Version: %d\n", decoded.Version)
	fmt.Printf("Settings: %s\n\n", decoded.Settings.String())

	// Example 3: TOON type with JSON
	fmt.Println("3. Using TOON type with JSON:")
	type Wrapper struct {
		Data toon.TOON `json:"data"`
	}
	w := Wrapper{
		Data: toon.TOON("key: value\nother: 123"),
	}
	jsonData, err := json.Marshal(w)
	if err != nil {
		panic(err)
	}
	fmt.Printf("JSON output: %s\n", jsonData)

	var wDecoded Wrapper
	err = json.Unmarshal(jsonData, &wDecoded)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decoded TOON data: %s\n\n", wDecoded.Data.String())

	// Example 4: TOON type helper methods
	fmt.Println("4. TOON type helper methods:")
	var empty toon.TOON
	fmt.Printf("Empty TOON IsNil(): %v\n", empty.IsNil())

	nonEmpty := toon.TOON("test: data")
	fmt.Printf("Non-empty TOON IsNil(): %v\n", nonEmpty.IsNil())
	fmt.Printf("TOON String(): %s\n", nonEmpty.String())

	// Example 5: Direct MarshalText/UnmarshalText
	fmt.Println("\n5. Direct MarshalText/UnmarshalText:")
	var raw toon.TOON
	err = raw.UnmarshalText([]byte("status: active\ncount: 42"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Unmarshaled: %s\n", raw.String())

	marshaled, err := raw.MarshalText()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Marshaled: %s\n", string(marshaled))
}
