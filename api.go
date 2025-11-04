// Package toon implements the Token-Oriented Object Notation (TOON)
// encoder and decoder described in docs/SPEC.md. TOON is a compact,
// human-readable serialization format targeting LLM workflows where predictable
// structure and reduced token counts are important. The package exposes a small
// public API while keeping implementation details inside internal packages.
package toon

import (
	"time"

	"github.com/toon-format/toon-go/internal/codec"
)

// Delimiter identifies the character used to split values inside array scopes.
type Delimiter = codec.Delimiter

const (
	// DelimiterComma is the default delimiter. It is omitted from brackets.
	DelimiterComma = codec.DelimiterComma
	// DelimiterTab uses HTAB for delimiting values.
	DelimiterTab = codec.DelimiterTab
	// DelimiterPipe uses the '|' character for delimiting values.
	DelimiterPipe = codec.DelimiterPipe
)

// EncoderOption mutates encoding behaviour.
type EncoderOption = codec.EncoderOption

// DecoderOption mutates decoder behaviour.
type DecoderOption = codec.DecoderOption

// Field represents a single key/value pair in an ordered object.
type Field = codec.Field

// Object preserves the encounter order of its fields, ensuring deterministic
// emission by the encoder.
type Object = codec.Object

// NewObject constructs an ordered Object from the provided key/value pairs.
func NewObject(fields ...Field) Object {
	return codec.NewObject(fields...)
}

// Encoder serializes Go values as TOON documents.
type Encoder = codec.Encoder

// NewEncoder constructs an Encoder using the supplied options. Absent options
// default to the TOON Core Profile recommendations (Section 19).
func NewEncoder(opts ...EncoderOption) *Encoder {
	return codec.NewEncoder(opts...)
}

// Marshal renders v into a TOON document using a temporary encoder.
func Marshal(v any, opts ...EncoderOption) ([]byte, error) {
	return codec.Marshal(v, opts...)
}

// MarshalString renders v as a TOON document string.
func MarshalString(v any, opts ...EncoderOption) (string, error) {
	return codec.MarshalString(v, opts...)
}

// WithIndent configures the number of spaces used per indentation level.
func WithIndent(spaces int) EncoderOption {
	return codec.WithIndent(spaces)
}

// WithDocumentDelimiter configures the delimiter that influences quoting
// decisions outside array scopes.
func WithDocumentDelimiter(delimiter Delimiter) EncoderOption {
	return codec.WithDocumentDelimiter(delimiter)
}

// WithArrayDelimiter configures the default delimiter declared for arrays that
// do not explicitly override the active delimiter.
func WithArrayDelimiter(delimiter Delimiter) EncoderOption {
	return codec.WithArrayDelimiter(delimiter)
}

// WithLengthMarkers enables emitting optional # markers in array headers.
func WithLengthMarkers(enabled bool) EncoderOption {
	return codec.WithLengthMarkers(enabled)
}

// WithTimeFormatter specifies the formatter used for time.Time normalization.
func WithTimeFormatter(formatter func(time.Time) string) EncoderOption {
	return codec.WithTimeFormatter(formatter)
}

// Decoder parses TOON documents into Go values that match the data model from
// Section 2. Numbers are returned as float64, objects as map[string]any, and
// arrays as []any. Strings are unescaped per Section 7.1.
type Decoder = codec.Decoder

// NewDecoder constructs a Decoder with the given options.
func NewDecoder(opts ...DecoderOption) *Decoder {
	return codec.NewDecoder(opts...)
}

// Decode parses the provided TOON document using a temporary decoder.
func Decode(data []byte, opts ...DecoderOption) (any, error) {
	return codec.Decode(data, opts...)
}

// DecodeString parses a TOON document string using a temporary decoder.
func DecodeString(s string, opts ...DecoderOption) (any, error) {
	return codec.DecodeString(s, opts...)
}

// WithStrictMode toggles the strict-mode diagnostics.
func WithStrictMode(strict bool) DecoderOption {
	return codec.WithStrictMode(strict)
}

// WithDecoderIndent configures the expected indentation step.
func WithDecoderIndent(spaces int) DecoderOption {
	return codec.WithDecoderIndent(spaces)
}

// WithDecoderDocumentDelimiter configures the delimiter that influences
// delimiter-aware string parsing when no array header is active.
func WithDecoderDocumentDelimiter(delimiter Delimiter) DecoderOption {
	return codec.WithDecoderDocumentDelimiter(delimiter)
}

// Unmarshal decodes the TOON document in data into v, which must be a non-nil
// pointer. Struct fields use `toon` struct tags for naming and omitempty
// semantics, mirroring Marshal behaviour.
func Unmarshal(data []byte, v any, opts ...DecoderOption) error {
	return codec.Unmarshal(data, v, opts...)
}

// UnmarshalString decodes the TOON document in s into v.
func UnmarshalString(s string, v any, opts ...DecoderOption) error {
	return codec.UnmarshalString(s, v, opts...)
}
