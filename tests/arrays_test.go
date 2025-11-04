package toon_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/toon-format/toon-go"
)

func TestMarshalTabularArray(t *testing.T) {
	payload := usersPayload{
		Users: []profile{
			{ID: 1, Name: "Ada", Active: true},
			{ID: 2, Name: "Bob", Active: false},
		},
		Count: 2,
	}

	doc, err := toon.MarshalString(payload)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	expectLines(t, doc,
		"users[2]{id,name,active}:",
		"  1,Ada,true",
		"  2,Bob,false",
		"count: 2",
	)
}

func TestMarshalMixedArray(t *testing.T) {
	payload := mixedEnvelope{
		Events: []any{
			"ready",
			metricEvent{Type: "metric", Values: []int{1, 2, 3}},
			[]string{"nested", "list"},
		},
	}

	doc, err := toon.MarshalString(payload)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	expectLines(t, doc,
		"events[3]:",
		"  - ready",
		"  - type: metric",
		"    values[3]: 1,2,3",
		"  - [2]: nested,list",
	)
}

func TestMarshalDelimitersAndLengthMarkers(t *testing.T) {
	payload := usersPayload{
		Users: []profile{{ID: 1, Name: "Ada", Active: true}},
		Count: 1,
	}

	doc, err := toon.MarshalString(payload,
		toon.WithDocumentDelimiter(toon.DelimiterPipe),
		toon.WithArrayDelimiter(toon.DelimiterPipe),
		toon.WithLengthMarkers(true),
	)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	expectLines(t, doc,
		"users[#1|]{id|name|active}:",
		"  1|Ada|true",
		"count: 1",
	)
}

func TestNestedDelimiterScopes(t *testing.T) {
	payload := struct {
		Buckets []struct {
			Name   string   `toon:"name"`
			Values []string `toon:"values"`
		} `toon:"buckets"`
	}{
		Buckets: []struct {
			Name   string   `toon:"name"`
			Values []string `toon:"values"`
		}{
			{Name: "alpha", Values: []string{"a", "b"}},
			{Name: "beta", Values: []string{"c", "d"}},
		},
	}

	doc, err := toon.MarshalString(payload, toon.WithArrayDelimiter(toon.DelimiterPipe))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	expectLines(t, doc,
		"buckets[2|]:",
		"  - name: alpha",
		"    values[2|]: a|b",
		"  - name: beta",
		"    values[2|]: c|d",
	)

	root := decodeMap(t, doc)
	buckets := root["buckets"].([]any)
	first := buckets[0].(map[string]any)
	vals := first["values"].([]any)
	if !reflect.DeepEqual(vals, []any{"a", "b"}) {
		t.Fatalf("unexpected values: %#v", vals)
	}
}

func TestDecodeTabular(t *testing.T) {
	doc := strings.Join([]string{
		"users[2]{id,name,active}:",
		"  1,Ada,true",
		"  2,Bob,false",
		"count: 2",
	}, "\n")

	root := decodeMap(t, doc)
	if root["count"] != float64(2) {
		t.Fatalf("count mismatch: %v", root["count"])
	}
	users := root["users"].([]any)
	first := users[0].(map[string]any)
	if first["id"] != float64(1) || first["name"] != "Ada" || first["active"] != true {
		t.Fatalf("unexpected first user: %#v", first)
	}
}

func TestDecodeMixedArray(t *testing.T) {
	doc := strings.Join([]string{
		"events[3]:",
		"  - ready",
		"  - type: metric",
		"    values[3]: 1,2,3",
		"  - [2]: nested,list",
	}, "\n")

	root := decodeMap(t, doc)
	events := root["events"].([]any)
	if len(events) != 3 {
		t.Fatalf("events length = %d", len(events))
	}
	second := events[1].(map[string]any)
	if second["type"] != "metric" {
		t.Fatalf("unexpected second event: %#v", second)
	}
	values := second["values"].([]any)
	if !reflect.DeepEqual(values, []any{float64(1), float64(2), float64(3)}) {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestRoundTripObjectListArrayFirstField(t *testing.T) {
	payload := bucketSet{
		Buckets: []bucket{
			{Values: []int{1, 2}, Label: "alpha"},
			{Values: []int{3, 4}, Label: "beta"},
		},
	}

	doc, err := toon.MarshalString(payload)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	expectLines(t, doc,
		"buckets[2]:",
		"  - values[2]: 1,2",
		"    label: alpha",
		"  - values[2]: 3,4",
		"    label: beta",
	)

	root := decodeMap(t, doc)
	buckets := root["buckets"].([]any)
	first := buckets[0].(map[string]any)
	vals := first["values"].([]any)
	if !reflect.DeepEqual(vals, []any{float64(1), float64(2)}) {
		t.Fatalf("unexpected values: %#v", vals)
	}

	var decoded bucketSet
	if err := toon.UnmarshalString(doc, &decoded); err != nil {
		t.Fatalf("UnmarshalString: %v", err)
	}
	if decoded.Buckets[1].Label != "beta" || decoded.Buckets[0].Values[1] != 2 {
		t.Fatalf("unexpected decoded buckets: %#v", decoded.Buckets)
	}
}
