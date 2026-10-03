package toon_test

import (
	"strings"
	"testing"
	"time"

	"github.com/toon-format/toon-go"
)

func TestDecodePermissive(t *testing.T) {
	doc := "items[2]: 1,2,3"
	if _, err := toon.DecodeString(doc, toon.WithStrictMode(false)); err != nil {
		t.Fatalf("permissive decode failed: %v", err)
	}
}

func TestDecoderIndentOption(t *testing.T) {
	doc := strings.Join([]string{
		"items[1]:",
		"\t- item",
	}, "\n")

	if _, err := toon.DecodeString(doc, toon.WithStrictMode(false), toon.WithDecoderIndent(1)); err != nil {
		t.Fatalf("permissive tab decode failed: %v", err)
	}
}

func TestDecodeManyBlankLinesInLinearTime(t *testing.T) {
	doc := "a: 1\n" + strings.Repeat("\n", 200_000) + "b: 2"
	start := time.Now()
	decodeMap(t, doc)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("decoding 200k blank lines took %v, want under 1s", elapsed)
	}
}
