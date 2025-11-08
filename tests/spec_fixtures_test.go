package toon_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/toon-format/toon-go"
)

type fixtureFile struct {
	Version     string        `json:"version"`
	Category    string        `json:"category"`
	Description string        `json:"description"`
	Tests       []fixtureCase `json:"tests"`
}

type fixtureCase struct {
	Name        string          `json:"name"`
	Input       json.RawMessage `json:"input"`
	Expected    json.RawMessage `json:"expected"`
	Options     map[string]any  `json:"options"`
	ShouldError bool            `json:"shouldError"`
	SpecSection string          `json:"specSection"`
	Note        string          `json:"note"`
}

func TestSpecEncodeFixtures(t *testing.T) {
	t.Helper()
	root := filepath.Join("spec", "tests", "fixtures", "encode")
	for _, path := range listFixtureFiles(t, root) {
		path := path
		fixture := loadFixtureFile(t, path)
		if fixture.Category != "encode" {
			t.Fatalf("%s: unexpected category %q", filepath.Base(path), fixture.Category)
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			for _, tc := range fixture.Tests {
				tc := tc
				t.Run(tc.Name, func(t *testing.T) {
					input := decodeEncodeInput(t, tc.Input)
					opts := encoderOptionsFromFixture(t, tc.Options)
					doc, err := toon.MarshalString(input, opts...)
					if tc.ShouldError {
						if err == nil {
							t.Fatalf("expected error, got %q", doc)
						}
						return
					}
					if err != nil {
						t.Fatalf("MarshalString: %v", err)
					}
					want := decodeFixtureString(t, tc.Expected)
					if doc != want {
						t.Fatalf("output mismatch\ngot:\n%s\nwant:\n%s", doc, want)
					}
				})
			}
		})
	}
}

func TestSpecDecodeFixtures(t *testing.T) {
	t.Helper()
	root := filepath.Join("spec", "tests", "fixtures", "decode")
	for _, path := range listFixtureFiles(t, root) {
		path := path
		fixture := loadFixtureFile(t, path)
		if fixture.Category != "decode" {
			t.Fatalf("%s: unexpected category %q", filepath.Base(path), fixture.Category)
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			for _, tc := range fixture.Tests {
				tc := tc
				t.Run(tc.Name, func(t *testing.T) {
					input := decodeFixtureString(t, tc.Input)
					opts := decoderOptionsFromFixture(t, tc.Options)
					value, err := toon.DecodeString(input, opts...)
					if tc.ShouldError {
						if err == nil {
							t.Fatalf("expected error, got value %#v", value)
						}
						return
					}
					if err != nil {
						t.Fatalf("DecodeString: %v", err)
					}
					want := decodeFixtureValue(t, tc.Expected)
					if !reflect.DeepEqual(value, want) {
						t.Fatalf("decoded value mismatch\n got: %#v\nwant: %#v", value, want)
					}
				})
			}
		})
	}
}

func listFixtureFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(paths)
	return paths
}

func loadFixtureFile(t *testing.T, path string) fixtureFile {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	var fixture fixtureFile
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("Unmarshal %s: %v", path, err)
	}
	return fixture
}

func decodeEncodeInput(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := parseOrderedJSON(decoder)
	if err != nil {
		t.Fatalf("decode encode input: %v", err)
	}
	return value
}

func parseOrderedJSON(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch token := tok.(type) {
	case json.Delim:
		switch token {
		case '{':
			fields := make([]toon.Field, 0)
			for dec.More() {
				keyToken, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, fmt.Errorf("expected string key, got %T", keyToken)
				}
				value, err := parseOrderedJSON(dec)
				if err != nil {
					return nil, err
				}
				fields = append(fields, toon.Field{Key: key, Value: value})
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return toon.NewObject(fields...), nil
		case '[':
			var items []any
			for dec.More() {
				value, err := parseOrderedJSON(dec)
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return items, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter %v", token)
		}
	case json.Number:
		return token, nil
	case string:
		return token, nil
	case bool:
		return token, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported token %T", token)
	}
}

func decodeFixtureString(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("decode string: %v", err)
	}
	return s
}

func decodeFixtureValue(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	if raw == nil {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode expected value: %v", err)
	}
	return v
}

func encoderOptionsFromFixture(t *testing.T, options map[string]any) []toon.EncoderOption {
	t.Helper()
	if len(options) == 0 {
		return nil
	}
	var result []toon.EncoderOption
	if value, ok := options["indent"]; ok {
		result = append(result, toon.WithIndent(asPositiveInt(t, value, "indent")))
	}
	if value, ok := options["delimiter"]; ok {
		marker := asDelimiter(t, value)
		result = append(result,
			toon.WithDocumentDelimiter(marker),
			toon.WithArrayDelimiter(marker),
		)
	}
	if value, ok := options["lengthMarker"]; ok {
		if marker, ok := value.(string); ok && marker == "#" {
			result = append(result, toon.WithLengthMarkers(true))
		}
	}
	return result
}

func decoderOptionsFromFixture(t *testing.T, options map[string]any) []toon.DecoderOption {
	t.Helper()
	if len(options) == 0 {
		return nil
	}
	var result []toon.DecoderOption
	if value, ok := options["indent"]; ok {
		result = append(result, toon.WithDecoderIndent(asPositiveInt(t, value, "indent")))
	}
	if value, ok := options["strict"]; ok {
		strict, ok := value.(bool)
		if !ok {
			t.Fatalf("strict option must be boolean, got %T", value)
		}
		result = append(result, toon.WithStrictMode(strict))
	}
	return result
}

func asPositiveInt(t *testing.T, value any, name string) int {
	t.Helper()
	switch v := value.(type) {
	case float64:
		if math.Trunc(v) != v {
			t.Fatalf("%s option must be integer, got %v", name, v)
		}
		if v < 0 {
			t.Fatalf("%s option must be positive, got %v", name, v)
		}
		return int(v)
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			t.Fatalf("%s option: %v", name, err)
		}
		if n < 0 {
			t.Fatalf("%s option must be positive, got %d", name, n)
		}
		return int(n)
	default:
		t.Fatalf("%s option must be numeric, got %T", name, value)
	}
	return 0
}

func asDelimiter(t *testing.T, value any) toon.Delimiter {
	t.Helper()
	str, ok := value.(string)
	if !ok {
		t.Fatalf("delimiter option must be string, got %T", value)
	}
	switch str {
	case ",":
		return toon.DelimiterComma
	case "\t":
		return toon.DelimiterTab
	case "|":
		return toon.DelimiterPipe
	default:
		t.Fatalf("unsupported delimiter %q", str)
	}
	return toon.DelimiterComma
}
