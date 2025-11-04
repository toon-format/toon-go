package main

import (
	"fmt"

	"github.com/toon-format/toon-go"
)

func main() {
	doc := "items[2]: 1,2,3"

	_, err := toon.DecodeString(doc)
	fmt.Printf("strict decode error: %v\n", err)

	value, err := toon.DecodeString(doc, toon.WithStrictMode(false))
	if err != nil {
		panic(err)
	}
	root := value.(map[string]any)
	items := root["items"].([]any)
	fmt.Printf("permissive decode length=%d\n", len(items))
}
