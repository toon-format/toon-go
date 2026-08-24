package toon_test

import (
	"reflect"
	"testing"

	"github.com/toon-format/toon-go"
)

// §9.3: a nested-uniform column is declared as a nested field group while the
// rows stay flat delimiter-separated primitives.
func TestMarshalNestedFieldGroups(t *testing.T) {
	orders := toon.NewObject(toon.Field{Key: "orders", Value: []any{
		toon.NewObject(
			toon.Field{Key: "id", Value: 1},
			toon.Field{Key: "customer", Value: toon.NewObject(
				toon.Field{Key: "name", Value: "Ada"},
				toon.Field{Key: "country", Value: "UK"},
			)},
			toon.Field{Key: "total", Value: 9.5},
		),
		toon.NewObject(
			toon.Field{Key: "id", Value: 2},
			toon.Field{Key: "customer", Value: toon.NewObject(
				toon.Field{Key: "name", Value: "Bob"},
				toon.Field{Key: "country", Value: "ES"},
			)},
			toon.Field{Key: "total", Value: 14},
		),
	}})

	doc, err := toon.MarshalString(orders)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc,
		"orders[2]{id,customer{name,country},total}:",
		"  1,Ada,UK,9.5",
		"  2,Bob,ES,14",
	)

	decoded, err := toon.DecodeString(doc)
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	want := map[string]any{"orders": []any{
		map[string]any{"id": float64(1), "customer": map[string]any{"name": "Ada", "country": "UK"}, "total": 9.5},
		map[string]any{"id": float64(2), "customer": map[string]any{"name": "Bob", "country": "ES"}, "total": float64(14)},
	}}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("round trip mismatch:\n got: %#v\nwant: %#v", decoded, want)
	}
}

// §9.3: nesting depth is unbounded.
func TestMarshalDeeplyNestedFieldGroups(t *testing.T) {
	row := func(city string) toon.Object {
		return toon.NewObject(
			toon.Field{Key: "a", Value: toon.NewObject(
				toon.Field{Key: "b", Value: toon.NewObject(
					toon.Field{Key: "c", Value: city},
				)},
			)},
		)
	}
	doc, err := toon.MarshalString(toon.NewObject(toon.Field{
		Key:   "rows",
		Value: []any{row("x"), row("y")},
	}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc,
		"rows[2]{a{b{c}}}:",
		"  x",
		"  y",
	)
}

// §9.5: an object whose values are uniform objects collapses into keyed tabular
// form, both in object-field position and at the document root.
func TestMarshalKeyedTabular(t *testing.T) {
	users := toon.NewObject(
		toon.Field{Key: "alice", Value: toon.NewObject(
			toon.Field{Key: "age", Value: 30},
			toon.Field{Key: "city", Value: "Madrid"},
		)},
		toon.Field{Key: "bob", Value: toon.NewObject(
			toon.Field{Key: "age", Value: 41},
			toon.Field{Key: "city", Value: "Lisboa"},
		)},
	)

	doc, err := toon.MarshalString(toon.NewObject(toon.Field{Key: "users", Value: users}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc,
		"users[2:]{age,city}:",
		"  alice: 30,Madrid",
		"  bob: 41,Lisboa",
	)

	rootDoc, err := toon.MarshalString(users)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, rootDoc,
		"[2:]{age,city}:",
		"  alice: 30,Madrid",
		"  bob: 41,Lisboa",
	)

	decoded, err := toon.DecodeString(rootDoc)
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	want := map[string]any{
		"alice": map[string]any{"age": float64(30), "city": "Madrid"},
		"bob":   map[string]any{"age": float64(41), "city": "Lisboa"},
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("round trip mismatch:\n got: %#v\nwant: %#v", decoded, want)
	}
}

// §9.5: detection needs at least two entries, so a single-entry object nests.
func TestSingleEntryObjectDoesNotCollapse(t *testing.T) {
	doc, err := toon.MarshalString(toon.NewObject(toon.Field{
		Key: "users",
		Value: toon.NewObject(toon.Field{Key: "alice", Value: toon.NewObject(
			toon.Field{Key: "age", Value: 30},
		)}),
	}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc,
		"users:",
		"  alice:",
		"    age: 30",
	)
}

// §5.1: comment lines are removed in a lexical pre-pass and never terminate a
// scope or count as a row.
func TestDecodeStripsComments(t *testing.T) {
	doc := "# leading\nitems[2]{id}:\n  1\n  # between rows\n  2\n# trailing"
	value, err := toon.DecodeString(doc)
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	want := map[string]any{"items": []any{
		map[string]any{"id": float64(1)},
		map[string]any{"id": float64(2)},
	}}
	if !reflect.DeepEqual(value, want) {
		t.Fatalf("decoded mismatch:\n got: %#v\nwant: %#v", value, want)
	}
}

// §7.2: a string that starts with "#" must be quoted so that encoder output
// never contains a line that reads as a comment.
func TestEncodeQuotesCommentLikeStrings(t *testing.T) {
	doc, err := toon.MarshalString(toon.NewObject(toon.Field{Key: "note", Value: "#1 pick"}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc, `note: "#1 pick"`)
}

// §9.1: empty arrays use the explicit form, never the legacy header.
func TestEncodeEmptyArrays(t *testing.T) {
	doc, err := toon.MarshalString(toon.NewObject(toon.Field{Key: "items", Value: []any{}}))
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc, "items: []")

	rootDoc, err := toon.MarshalString([]any{})
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, rootDoc, "[]")
}

// §14.2: the [#N] length-marker syntax was removed in TOON 2.0.
func TestDecodeRejectsLengthMarker(t *testing.T) {
	if _, err := toon.DecodeString("items[#2]: 1,2"); err == nil {
		t.Fatalf("expected an error for the removed [#N] syntax")
	}
}
