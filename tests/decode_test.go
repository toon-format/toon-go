package toon_test

import (
	"strings"
	"testing"
	"time"
)

func TestDecodeManyBlankLinesInLinearTime(t *testing.T) {
	doc := "a: 1\n" + strings.Repeat("\n", 200_000) + "b: 2"
	start := time.Now()
	decodeMap(t, doc)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("decoding 200k blank lines took %v, want under 1s", elapsed)
	}
}
