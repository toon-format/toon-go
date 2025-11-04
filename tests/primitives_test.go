package toon_test

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/toon-format/toon-go"
)

func TestMarshalPrimitiveRoot(t *testing.T) {
	doc, err := toon.MarshalString("hello")
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	if doc != "hello" {
		t.Fatalf("unexpected output %q", doc)
	}
}

func TestMarshalNormalization(t *testing.T) {
	payload := struct {
		Timestamp time.Time `toon:"timestamp"`
		NotANum   float64   `toon:"nan"`
		Big       *big.Int  `toon:"big"`
	}{
		Timestamp: time.Date(2025, 10, 31, 12, 0, 0, 0, time.UTC),
		NotANum:   math.NaN(),
		Big:       big.NewInt(0).Exp(big.NewInt(10), big.NewInt(6), nil),
	}

	doc, err := toon.MarshalString(payload)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	lines := strings.Split(doc, "\n")
	if !containsLine(lines, "timestamp: \"2025-10-31T12:00:00Z\"") {
		t.Fatalf("timestamp line missing: %v", lines)
	}
	if !containsLine(lines, "nan: null") {
		t.Fatalf("NaN normalization missing: %v", lines)
	}
	if !containsLine(lines, "big: 1000000") {
		t.Fatalf("big int normalization missing: %v", lines)
	}
}

func TestMarshalLargeIntegerPrecision(t *testing.T) {
	payload := map[string]any{
		"safe":  int64(9007199254740991),
		"large": int64(9007199254740993),
		"huge":  big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil),
	}

	doc, err := toon.MarshalString(payload)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	lines := strings.Split(doc, "\n")
	if !containsLine(lines, "safe: 9007199254740991") {
		t.Fatalf("safe integer should remain numeric: %v", lines)
	}
	if !containsLine(lines, "large: \"9007199254740993\"") {
		t.Fatalf("large integer should be quoted: %v", lines)
	}
	if !containsLine(lines, "huge: \"1000000000000000000\"") {
		t.Fatalf("huge integer should be quoted: %v", lines)
	}

	value, err := toon.DecodeString(doc)
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	root := value.(map[string]any)
	if root["large"] != "9007199254740993" {
		t.Fatalf("large integer decode mismatch: %#v", root["large"])
	}
	if root["huge"] != "1000000000000000000" {
		t.Fatalf("huge integer decode mismatch: %#v", root["huge"])
	}
}

func TestMarshalWithObjectHelper(t *testing.T) {
	doc, err := toon.MarshalString(toon.NewObject(
		toon.Field{Key: "first", Value: 1},
		toon.Field{Key: "second", Value: "value"},
	))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc,
		"first: 1",
		"second: value",
	)
}

func TestMarshalCustomTimeFormatter(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 4, 5, 6, time.UTC)
	doc, err := toon.MarshalString(map[string]any{"ts": ts}, toon.WithTimeFormatter(func(t time.Time) string {
		return t.Format(time.RFC822)
	}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	lines := strings.Split(doc, "\n")
	if !containsLine(lines, "ts: \"02 Jan 24 03:04 UTC\"") {
		t.Fatalf("time formatter not applied: %v", lines)
	}
}

func TestMarshalWithIndentOption(t *testing.T) {
	payload := map[string]any{
		"outer": map[string]any{
			"inner": []int{1, 2},
		},
	}
	doc, err := toon.MarshalString(payload, toon.WithIndent(4))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc,
		"outer:",
		"    inner[2]: 1,2",
	)
}

func TestStringerNormalization(t *testing.T) {
	t.Run("custom stringer", func(t *testing.T) {
		val := struct {
			ID fmt.Stringer `toon:"id"`
		}{
			ID: stringer("abc-123"),
		}
		doc, err := toon.MarshalString(val)
		if err != nil {
			t.Fatalf("MarshalString: %v", err)
		}
		expectLines(t, doc, "id: abc-123")
	})
}

type stringer string

func (s stringer) String() string {
	return string(s)
}
