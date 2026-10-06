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

func TestDecodeRejectsInvalidUTF8(t *testing.T) {
	doc := []byte("a: 1\nb: x\xffy")
	for _, strict := range []bool{true, false} {
		_, err := toon.Decode(doc, toon.WithStrictMode(strict))
		if err == nil {
			t.Fatalf("strict=%v: expected an error for ill-formed UTF-8", strict)
		}
		if got, want := err.Error(), "line 2: invalid UTF-8 sequence"; got != want {
			t.Fatalf("strict=%v: error = %q, want %q", strict, got, want)
		}
	}
}
