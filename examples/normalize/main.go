package main

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/toon-format/toon-go"
)

func main() {
	queue := map[string]any{
		"timestamp": time.Date(2025, 10, 31, 12, 0, 0, 0, time.UTC),
		"nan":       math.NaN(),
		"big":       big.NewInt(0).Exp(big.NewInt(10), big.NewInt(30), nil),
	}

	encoded, err := toon.MarshalString(queue)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded)
}
