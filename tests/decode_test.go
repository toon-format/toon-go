package toon_test

import (
	"strings"
	"testing"

	"github.com/toon-format/toon-go"
)

func TestDecodeRootForms(t *testing.T) {
	value, err := toon.DecodeString("42")
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	if value.(float64) != 42 {
		t.Fatalf("expected 42, got %v", value)
	}

	arr, err := toon.DecodeString("[2]: 1,2")
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	slice := arr.([]any)
	if len(slice) != 2 || slice[0].(float64) != 1 || slice[1].(float64) != 2 {
		t.Fatalf("unexpected root array: %#v", slice)
	}
}

func TestDecodeStrictErrors(t *testing.T) {
	cases := []struct {
		name string
		doc  string
	}{
		{name: "length mismatch", doc: "items[2]: 1"},
		{name: "indent mismatch", doc: "key:\n  child:\n   grand: value"},
		{name: "blank line", doc: "items[1]:\n  - ready\n\n  - set"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if _, err := toon.DecodeString(tc.doc); err == nil {
				t.Fatalf("expected strict error for %q", tc.doc)
			}
		})
	}
}

func TestDecodePermissive(t *testing.T) {
	doc := "items[2]: 1,2,3"
	if _, err := toon.DecodeString(doc, toon.WithStrictMode(false)); err != nil {
		t.Fatalf("permissive decode failed: %v", err)
	}
}

func TestDecodeWithCustomDocumentDelimiter(t *testing.T) {
	doc := strings.Join([]string{
		"records[2|]:",
		"  - id: a|b",
		"  - id: c|d",
	}, "\n")

	root := decodeMap(t, doc, toon.WithDecoderDocumentDelimiter(toon.DelimiterPipe))
	records := root["records"].([]any)
	if len(records) != 2 {
		t.Fatalf("expected records length 2, got %d", len(records))
	}
}

func TestDecoderIndentOption(t *testing.T) {
	doc := strings.Join([]string{
		"items[1]:",
		"\t- item",
	}, "\n")

	if _, err := toon.DecodeString(doc); err == nil {
		t.Fatalf("expected strict indentation error with tabs")
	}

	if _, err := toon.DecodeString(doc, toon.WithStrictMode(false), toon.WithDecoderIndent(1)); err != nil {
		t.Fatalf("permissive tab decode failed: %v", err)
	}
}
