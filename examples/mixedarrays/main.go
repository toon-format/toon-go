package main

import (
	"fmt"

	"github.com/toon-format/toon-go"
)

type MetricEvent struct {
	Type   string `toon:"type"`
	Values []int  `toon:"values"`
}

type EventEnvelope struct {
	Events []any `toon:"events"`
}

func main() {
	feed := []any{
		"ready",
		MetricEvent{Type: "metric", Values: []int{1, 2, 3}},
		[]string{"nested", "list"},
	}

	payload := EventEnvelope{Events: feed}

	encoded, err := toon.MarshalString(payload)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded)

	var decoded EventEnvelope
	if err := toon.UnmarshalString(encoded, &decoded); err != nil {
		panic(err)
	}
	fmt.Printf("event[1] type=%s\n", decoded.Events[1].(map[string]any)["type"])
}
