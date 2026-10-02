package toon_test

import (
	"testing"

	"github.com/toon-format/toon-go"
)

func TestUnmarshalNilTarget(t *testing.T) {
	err := toon.Unmarshal(nil, nil)
	if err == nil {
		t.Fatalf("expected error for nil target")
	}
	if err.Error() != "toon: Unmarshal nil target" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalNonPointer(t *testing.T) {
	var value any
	err := toon.Unmarshal([]byte("foo: bar"), value)
	if err == nil {
		t.Fatalf("expected error for non-pointer target")
	}
}

func TestDecodeRejectsInvalidUTF8InStrictMode(t *testing.T) {
	doc := []byte("a: 1\nb: x\xffy")
	_, err := toon.Decode(doc)
	if err == nil {
		t.Fatal("expected an error for ill-formed UTF-8")
	}
	if got, want := err.Error(), "line 2: invalid UTF-8 sequence"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}

	if _, err := toon.Decode(doc, toon.WithStrictMode(false)); err != nil {
		t.Fatalf("non-strict Decode: %v", err)
	}
}
