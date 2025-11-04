package toon_test

import (
	"testing"
	"time"

	"github.com/toon-format/toon-go"
)

func TestEncoderReusability(t *testing.T) {
	enc := toon.NewEncoder(
		toon.WithArrayDelimiter(toon.DelimiterPipe),
		toon.WithLengthMarkers(true),
	)

	first, err := enc.MarshalString(usersPayload{
		Users: []profile{{ID: 1, Name: "Ada", Active: true}},
		Count: 1,
	})
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	second, err := enc.MarshalString(usersPayload{
		Users: []profile{{ID: 2, Name: "Bob", Active: false}},
		Count: 1,
	})
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	if first == second {
		t.Fatalf("encoder produced identical output for different inputs")
	}
}

func TestDecoderOptionsCombination(t *testing.T) {
	doc := "items[2|]: 1|2|3"
	dec := toon.NewDecoder(
		toon.WithStrictMode(false),
		toon.WithDecoderDocumentDelimiter(toon.DelimiterPipe),
	)
	if _, err := dec.DecodeString(doc); err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
}

func TestTimeFormatterOptionDoesNotLeak(t *testing.T) {
	enc := toon.NewEncoder(toon.WithTimeFormatter(func(time.Time) string {
		return "custom"
	}))
	doc, err := enc.MarshalString(map[string]any{"ts": time.Now()})
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	if doc != "ts: custom" {
		t.Fatalf("unexpected doc: %s", doc)
	}

	other, err := toon.MarshalString(map[string]any{"ts": time.Unix(0, 0)})
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	if other == doc {
		t.Fatalf("time formatter leaked into default encoder")
	}
}
