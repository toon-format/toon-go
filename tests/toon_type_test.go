package toon_test

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"

	"github.com/toon-format/toon-go"
)

func TestTOONMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    toon.TOON
		expected string
	}{
		{
			name:     "simple TOON document",
			input:    toon.TOON("key: value\nother: 123"),
			expected: "key: value\nother: 123",
		},
		{
			name:     "empty TOON",
			input:    toon.TOON(""),
			expected: "",
		},
		{
			name:     "nil TOON",
			input:    nil,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText failed: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestTOONUnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple TOON document",
			input:    []byte("key: value\nother: 123"),
			expected: "key: value\nother: 123",
		},
		{
			name:     "empty bytes",
			input:    []byte(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc toon.TOON
			err := doc.UnmarshalText(tt.input)
			if err != nil {
				t.Fatalf("UnmarshalText failed: %v", err)
			}
			if string(doc) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(doc))
			}
		})
	}
}

func TestTOONUnmarshalTextNilPointer(t *testing.T) {
	var doc *toon.TOON
	err := doc.UnmarshalText([]byte("test"))
	if err == nil {
		t.Fatal("expected error for nil pointer, got nil")
	}
	if !strings.Contains(err.Error(), "nil pointer") {
		t.Errorf("expected 'nil pointer' error, got: %v", err)
	}
}

func TestTOONScan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{
			name:     "scan from bytes",
			input:    []byte("key: value"),
			expected: "key: value",
			wantErr:  false,
		},
		{
			name:     "scan from string",
			input:    "key: value",
			expected: "key: value",
			wantErr:  false,
		},
		{
			name:     "scan from nil",
			input:    nil,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "scan from invalid type",
			input:    123,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc toon.TOON
			err := doc.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}
			if string(doc) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(doc))
			}
		})
	}
}

func TestTOONScanNilPointer(t *testing.T) {
	var doc *toon.TOON
	err := doc.Scan([]byte("test"))
	if err == nil {
		t.Fatal("expected error for nil pointer, got nil")
	}
	if !strings.Contains(err.Error(), "nil pointer") {
		t.Errorf("expected 'nil pointer' error, got: %v", err)
	}
}

func TestTOONValue(t *testing.T) {
	tests := []struct {
		name     string
		input    toon.TOON
		expected driver.Value
	}{
		{
			name:     "non-nil TOON",
			input:    toon.TOON("key: value"),
			expected: []byte("key: value"),
		},
		{
			name:     "empty TOON",
			input:    toon.TOON(""),
			expected: []byte(""),
		},
		{
			name:     "nil TOON",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.Value()
			if err != nil {
				t.Fatalf("Value failed: %v", err)
			}
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			resultBytes, ok := result.([]byte)
			if !ok {
				t.Fatalf("expected []byte, got %T", result)
			}
			expectedBytes := tt.expected.([]byte)
			if string(resultBytes) != string(expectedBytes) {
				t.Errorf("expected %q, got %q", expectedBytes, resultBytes)
			}
		})
	}
}

func TestTOONString(t *testing.T) {
	tests := []struct {
		name     string
		input    toon.TOON
		expected string
	}{
		{
			name:     "simple TOON",
			input:    toon.TOON("key: value"),
			expected: "key: value",
		},
		{
			name:     "empty TOON",
			input:    toon.TOON(""),
			expected: "",
		},
		{
			name:     "nil TOON",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTOONIsNil(t *testing.T) {
	tests := []struct {
		name     string
		input    toon.TOON
		expected bool
	}{
		{
			name:     "non-empty TOON",
			input:    toon.TOON("key: value"),
			expected: false,
		},
		{
			name:     "empty TOON",
			input:    toon.TOON(""),
			expected: true,
		},
		{
			name:     "nil TOON",
			input:    nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.IsNil()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTOONInStruct(t *testing.T) {
	type Config struct {
		Version  int       `toon:"version"`
		Settings toon.TOON `toon:"settings"`
	}

	t.Run("marshal struct with TOON field", func(t *testing.T) {
		cfg := Config{
			Version:  1,
			Settings: toon.TOON("timeout: 30"),
		}

		result, err := toon.MarshalString(cfg)
		if err != nil {
			t.Fatalf("MarshalString failed: %v", err)
		}

		if !strings.Contains(result, "version: 1") {
			t.Errorf("result missing version field: %s", result)
		}
		if !strings.Contains(result, "timeout: 30") {
			t.Errorf("result missing settings content: %s", result)
		}
	})

	t.Run("unmarshal struct with TOON field", func(t *testing.T) {
		input := "version: 2\nsettings: timeout: 60"

		var cfg Config
		err := toon.UnmarshalString(input, &cfg)
		if err != nil {
			t.Fatalf("UnmarshalString failed: %v", err)
		}

		if cfg.Version != 2 {
			t.Errorf("expected version 2, got %d", cfg.Version)
		}

		settingsStr := cfg.Settings.String()
		if !strings.Contains(settingsStr, "timeout") {
			t.Errorf("settings missing timeout: %s", settingsStr)
		}
	})

	t.Run("round-trip struct with TOON field", func(t *testing.T) {
		original := Config{
			Version:  3,
			Settings: toon.TOON("mode: fast\ndebug: false"),
		}

		marshaled, err := toon.MarshalString(original)
		if err != nil {
			t.Fatalf("MarshalString failed: %v", err)
		}

		var decoded Config
		err = toon.UnmarshalString(marshaled, &decoded)
		if err != nil {
			t.Fatalf("UnmarshalString failed: %v", err)
		}

		if decoded.Version != original.Version {
			t.Errorf("version mismatch: expected %d, got %d", original.Version, decoded.Version)
		}

		if decoded.Settings.String() == "" {
			t.Error("settings lost during round-trip")
		}
	})
}

func TestTOONWithJSON(t *testing.T) {
	t.Run("JSON unmarshal string into TOON", func(t *testing.T) {
		type Wrapper struct {
			Data toon.TOON `json:"data"`
		}

		input := `{"data":"simple string"}`

		var w Wrapper
		err := json.Unmarshal([]byte(input), &w)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		// JSON string is converted to TOON format
		if w.Data.String() != "simple string" {
			t.Errorf("expected 'simple string', got %q", w.Data.String())
		}
	})

	t.Run("JSON unmarshal nested object into TOON", func(t *testing.T) {
		type Response struct {
			EventID string    `json:"event_id"`
			Payload toon.TOON `json:"payload"`
		}

		jsonInput := `{
			"event_id": "evt_xyz",
			"payload": {
				"id": "order_xyz",
				"amount": 99.99,
				"items": ["widget", "gadget"]
			}
		}`

		var response Response
		err := json.Unmarshal([]byte(jsonInput), &response)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if response.EventID != "evt_xyz" {
			t.Errorf("expected event_id 'evt_xyz', got %q", response.EventID)
		}

		// Payload is stored as TOON format
		if !strings.Contains(response.Payload.String(), "id: order_xyz") {
			t.Errorf("payload missing expected TOON content: %s", response.Payload.String())
		}

		// Decode the TOON payload
		payloadData, err := toon.Decode(response.Payload)
		if err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		payloadMap := payloadData.(map[string]any)
		if payloadMap["id"] != "order_xyz" {
			t.Errorf("expected order_xyz, got %v", payloadMap["id"])
		}
	})

	t.Run("JSON marshal TOON field back to JSON", func(t *testing.T) {
		type Response struct {
			EventID string    `json:"event_id"`
			Payload toon.TOON `json:"payload"`
		}

		// Start with TOON data
		response := Response{
			EventID: "evt_123",
			Payload: toon.TOON("key: value\nother: 42"),
		}

		// Marshal to JSON - TOON should convert back to JSON
		result, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		// Parse result to verify structure
		var parsed map[string]any
		if err := json.Unmarshal(result, &parsed); err != nil {
			t.Fatalf("failed to parse result: %v", err)
		}

		payload := parsed["payload"].(map[string]any)
		if payload["key"] != "value" {
			t.Errorf("expected key='value', got %v", payload["key"])
		}
		if payload["other"] != float64(42) {
			t.Errorf("expected other=42, got %v", payload["other"])
		}
	})
}

func TestTOONNested(t *testing.T) {
	type Document struct {
		Title    string    `toon:"title"`
		Metadata toon.TOON `toon:"metadata"`
	}

	t.Run("nested TOON objects", func(t *testing.T) {
		doc := Document{
			Title:    "Test Doc",
			Metadata: toon.TOON("author: Alice"),
		}

		marshaled, err := toon.MarshalString(doc)
		if err != nil {
			t.Fatalf("MarshalString failed: %v", err)
		}

		var decoded Document
		err = toon.UnmarshalString(marshaled, &decoded)
		if err != nil {
			t.Fatalf("UnmarshalString failed: %v", err)
		}

		if decoded.Title != doc.Title {
			t.Errorf("title mismatch: expected %q, got %q", doc.Title, decoded.Title)
		}

		if decoded.Metadata.IsNil() {
			t.Error("metadata lost during unmarshal")
		}
	})
}
