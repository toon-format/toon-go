package toon_test

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/toon-format/toon-go"
)

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
	if !slices.Contains(lines, "timestamp: \"2025-10-31T12:00:00Z\"") {
		t.Fatalf("timestamp line missing: %v", lines)
	}
	if !slices.Contains(lines, "nan: null") {
		t.Fatalf("NaN normalization missing: %v", lines)
	}
	if !slices.Contains(lines, "big: 1000000") {
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
	if !slices.Contains(lines, "safe: 9007199254740991") {
		t.Fatalf("safe integer should remain numeric: %v", lines)
	}
	if !slices.Contains(lines, "large: \"9007199254740993\"") {
		t.Fatalf("large integer should be quoted: %v", lines)
	}
	if !slices.Contains(lines, "huge: \"1000000000000000000\"") {
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

func TestMarshalCustomTimeFormatter(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 4, 5, 6, time.UTC)
	doc, err := toon.MarshalString(map[string]any{"ts": ts}, toon.WithTimeFormatter(func(t time.Time) string {
		return t.Format(time.RFC822)
	}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	lines := strings.Split(doc, "\n")
	if !slices.Contains(lines, "ts: \"02 Jan 24 03:04 UTC\"") {
		t.Fatalf("time formatter not applied: %v", lines)
	}
}

func TestStringerNormalization(t *testing.T) {
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
}

func TestMarshalNilStringerPointer(t *testing.T) {
	val := struct {
		When *time.Time `toon:"when"`
		ID   *stringer  `toon:"id"`
	}{}
	doc, err := toon.MarshalString(val)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc, "when: null", "id: null")
}

type stringer string

func (s stringer) String() string {
	return string(s)
}
