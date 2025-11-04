package toon_test

import (
	"strings"
	"testing"

	"github.com/toon-format/toon-go"
)

func TestMarshalStructOmitEmpty(t *testing.T) {
	user := profile{ID: 42, Name: "Grace", Active: true}

	doc, err := toon.MarshalString(user)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}

	expectLines(t, doc,
		"id: 42",
		"name: Grace",
		"active: true",
	)

	email := "grace@example.com"
	user.Email = &email
	doc, err = toon.MarshalString(user)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	lines := strings.Split(doc, "\n")
	if !containsLine(lines, "email: grace@example.com") {
		t.Fatalf("email field missing: %s", doc)
	}
}

func TestUnmarshalStructNested(t *testing.T) {
	doc := strings.Join([]string{
		"users[2]{id,name,active}:",
		"  1,Ada,true",
		"  2,Bob,false",
		"count: 2",
	}, "\n")

	var payload usersPayload
	if err := toon.UnmarshalString(doc, &payload); err != nil {
		t.Fatalf("UnmarshalString: %v", err)
	}
	if len(payload.Users) != 2 || payload.Users[1].Name != "Bob" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestUnmarshalTypedSlice(t *testing.T) {
	doc := strings.Join([]string{
		"events[2]:",
		"  - type: metric",
		"    values[2]: 1,2",
		"  - type: metric",
		"    values[2]: 3,4",
	}, "\n")

	var envelope typedEnvelope
	if err := toon.UnmarshalString(doc, &envelope); err != nil {
		t.Fatalf("UnmarshalString: %v", err)
	}
	if len(envelope.Events) != 2 || envelope.Events[0].Values[1] != 2 {
		t.Fatalf("unexpected events: %#v", envelope.Events)
	}
}

func TestPointerOmitEmptyRoundTrip(t *testing.T) {
	type pointerPayload struct {
		Name *string `toon:"name,omitempty"`
		Age  *int    `toon:"age,omitempty"`
		Flag bool    `toon:"flag"`
	}

	pp := pointerPayload{Flag: true}
	doc, err := toon.MarshalString(pp)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	expectLines(t, doc, "flag: true")

	name := "Jo"
	age := 7
	pp.Name = &name
	pp.Age = &age
	doc, err = toon.MarshalString(pp)
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	lines := strings.Split(doc, "\n")
	if !containsLine(lines, "name: Jo") || !containsLine(lines, "age: 7") {
		t.Fatalf("pointer fields missing: %s", doc)
	}

	var decoded pointerPayload
	if err := toon.UnmarshalString(doc, &decoded); err != nil {
		t.Fatalf("UnmarshalString: %v", err)
	}
	if decoded.Name == nil || *decoded.Name != "Jo" {
		t.Fatalf("name decode mismatch: %#v", decoded.Name)
	}
	if decoded.Age == nil || *decoded.Age != 7 {
		t.Fatalf("age decode mismatch: %#v", decoded.Age)
	}
}
