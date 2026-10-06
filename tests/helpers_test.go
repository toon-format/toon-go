package toon_test

import (
	"strings"
	"testing"
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
