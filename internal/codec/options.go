package codec

import (
	"fmt"
	"time"
)

// Delimiter identifies the character used to split values inside array scopes.
type Delimiter rune

const (
	// DelimiterComma is the default delimiter. It is omitted from brackets.
	DelimiterComma Delimiter = ','
	// DelimiterTab uses HTAB for delimiting values.
	DelimiterTab Delimiter = '\t'
	// DelimiterPipe uses the '|' character for delimiting values.
	DelimiterPipe Delimiter = '|'
)

func (d Delimiter) String() string {
	switch d {
	case DelimiterComma:
		return "comma"
	case DelimiterTab:
		return "tab"
	case DelimiterPipe:
		return "pipe"
	default:
		return fmt.Sprintf("delimiter(%q)", rune(d))
	}
}

func (d Delimiter) rune() rune {
	switch d {
	case DelimiterComma:
		return ','
	case DelimiterTab:
		return '\t'
	case DelimiterPipe:
		return '|'
	default:
		return ','
	}
}

// EncoderOption mutates encoding behaviour.
type EncoderOption func(*encoderOptions)

type encoderOptions struct {
	indentSize         int
	documentDelimiter  Delimiter
	arrayDelimiter     Delimiter
	includeLengthMarks bool
	timeFormatter      func(time.Time) string
}

func defaultEncoderOptions() encoderOptions {
	return encoderOptions{
		indentSize:        2,
		documentDelimiter: DelimiterComma,
		arrayDelimiter:    DelimiterComma,
		timeFormatter: func(t time.Time) string {
			return t.UTC().Format(time.RFC3339Nano)
		},
	}
}

// WithIndent configures the number of spaces used per indentation level.
func WithIndent(spaces int) EncoderOption {
	return func(o *encoderOptions) {
		if spaces > 0 {
			o.indentSize = spaces
		}
	}
}

// WithDocumentDelimiter configures the delimiter that influences quoting
// decisions outside array scopes.
func WithDocumentDelimiter(delimiter Delimiter) EncoderOption {
	return func(o *encoderOptions) {
		if delimiter == DelimiterComma || delimiter == DelimiterTab || delimiter == DelimiterPipe {
			o.documentDelimiter = delimiter
		}
	}
}

// WithArrayDelimiter configures the default delimiter declared for arrays that
// do not explicitly override the active delimiter.
func WithArrayDelimiter(delimiter Delimiter) EncoderOption {
	return func(o *encoderOptions) {
		if delimiter == DelimiterComma || delimiter == DelimiterTab || delimiter == DelimiterPipe {
			o.arrayDelimiter = delimiter
		}
	}
}

// WithLengthMarkers enables emitting optional # markers in array headers.
func WithLengthMarkers(enabled bool) EncoderOption {
	return func(o *encoderOptions) {
		o.includeLengthMarks = enabled
	}
}

// WithTimeFormatter specifies the formatter used for time.Time normalization.
func WithTimeFormatter(formatter func(time.Time) string) EncoderOption {
	return func(o *encoderOptions) {
		if formatter != nil {
			o.timeFormatter = formatter
		}
	}
}

// DecoderOption mutates decoder behaviour.
type DecoderOption func(*decoderOptions)

type decoderOptions struct {
	indentSize    int
	strict        bool
	documentDelim Delimiter
}

func defaultDecoderOptions() decoderOptions {
	return decoderOptions{
		indentSize:    2,
		strict:        true,
		documentDelim: DelimiterComma,
	}
}

// WithStrictMode toggles the strict-mode diagnostics.
func WithStrictMode(strict bool) DecoderOption {
	return func(o *decoderOptions) {
		o.strict = strict
	}
}

// WithDecoderIndent configures the expected indentation step.
func WithDecoderIndent(spaces int) DecoderOption {
	return func(o *decoderOptions) {
		if spaces > 0 {
			o.indentSize = spaces
		}
	}
}

// WithDecoderDocumentDelimiter configures the delimiter that influences
// delimiter-aware string parsing when no array header is active.
func WithDecoderDocumentDelimiter(delimiter Delimiter) DecoderOption {
	return func(o *decoderOptions) {
		if delimiter == DelimiterComma || delimiter == DelimiterTab || delimiter == DelimiterPipe {
			o.documentDelim = delimiter
		}
	}
}
