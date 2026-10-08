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
