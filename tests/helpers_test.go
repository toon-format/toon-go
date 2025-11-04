package toon_test

import (
	"strings"
	"testing"

	"github.com/toon-format/toon-go"
)

func expectLines(t *testing.T, doc string, want ...string) {
	t.Helper()
	lines := strings.Split(doc, "\n")
	if len(lines) != len(want) {
		t.Fatalf("line count mismatch: got %d, want %d\nGot:\n%s\nWant:\n%s",
			len(lines), len(want), doc, strings.Join(want, "\n"))
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("line %d mismatch:\n got: %q\nwant: %q\nFull:\n%s",
				i+1, lines[i], want[i], doc)
		}
	}
}

func containsLine(lines []string, target string) bool {
	for _, line := range lines {
		if line == target {
			return true
		}
	}
	return false
}

func decodeMap(t *testing.T, doc string, opts ...toon.DecoderOption) map[string]any {
	t.Helper()
	value, err := toon.DecodeString(doc, opts...)
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	root, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map root, got %T", value)
	}
	return root
}
