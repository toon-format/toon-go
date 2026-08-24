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

// §7.4 obliges decoders to accept any unquoted key token as a literal key, even
// one an encoder would have had to quote.
func TestDecodeAcceptsNonEncoderKeys(t *testing.T) {
	value, err := toon.DecodeString("1invalid: value")
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	doc, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected an object, got %T", value)
	}
	if doc["1invalid"] != "value" {
		t.Fatalf("unexpected decoded document: %#v", doc)
	}
}

func TestDecodeInvalidQuotedString(t *testing.T) {
	doc := "name: \"unterminated"
	if _, err := toon.DecodeString(doc); err == nil {
		t.Fatalf("expected quoted string error")
	}
}
