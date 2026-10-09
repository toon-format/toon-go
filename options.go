package toon

import (
	"fmt"
	"time"
)

// Delimiter identifies the character used to split field entries, inline array
// values, tabular row cells, and keyed entry-row cells.
type Delimiter rune

const (
	// DelimiterComma is the default delimiter. It is omitted from brackets.
	DelimiterComma Delimiter = ','
	// DelimiterTab uses HTAB for delimiting values.
	DelimiterTab Delimiter = '\t'
	// DelimiterPipe uses the '|' character for delimiting values.
	DelimiterPipe Delimiter = '|'
)

// String returns "comma", "tab", or "pipe", or delimiter('x') for any other
// rune x.
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
	case DelimiterComma, DelimiterTab, DelimiterPipe:
		return rune(d)
	default:
		return ','
	}
}

// symbol returns the delimiter symbol as it appears inside a bracket segment.
// Comma is implied by its absence.
func (d Delimiter) symbol() string {
	if d == DelimiterComma {
		return ""
	}
	return string(d.rune())
}

func validDelimiter(d Delimiter) bool {
	return d == DelimiterComma || d == DelimiterTab || d == DelimiterPipe
}

// EncoderOption configures an Encoder.
type EncoderOption func(*encoderOptions)

type encoderOptions struct {
	indentSize    int
	delimiter     Delimiter
	timeFormatter func(time.Time) string
}

func defaultEncoderOptions() encoderOptions {
	return encoderOptions{
		indentSize: 2,
		delimiter:  DelimiterComma,
		timeFormatter: func(t time.Time) string {
			return t.UTC().Format(time.RFC3339Nano)
		},
	}
}

// WithIndent configures the number of spaces used per indentation level. A
// value below 1 is ignored.
func WithIndent(spaces int) EncoderOption {
	return func(o *encoderOptions) {
		if spaces > 0 {
			o.indentSize = spaces
		}
	}
}

// WithDelimiter sets the delimiter that every emitted header declares. A value
// other than DelimiterComma, DelimiterTab, or DelimiterPipe is ignored.
func WithDelimiter(delimiter Delimiter) EncoderOption {
	return func(o *encoderOptions) {
		if validDelimiter(delimiter) {
			o.delimiter = delimiter
		}
	}
}

// WithTimeFormatter specifies the formatter used for time.Time normalization. A
// nil formatter is ignored.
func WithTimeFormatter(formatter func(time.Time) string) EncoderOption {
	return func(o *encoderOptions) {
		if formatter != nil {
			o.timeFormatter = formatter
		}
	}
}

// DecoderOption configures a Decoder.
type DecoderOption func(*decoderOptions)

type decoderOptions struct {
	indentSize int
	strict     bool
}

func defaultDecoderOptions() decoderOptions {
	return decoderOptions{
		indentSize: 2,
		strict:     true,
	}
}

// WithStrictMode toggles strict decoding. With false, the decoder accepts
// exactly five deviations: a declared length that doesn't match the content,
// duplicate keys and field names (the last one wins), tab or misaligned
// indentation (each tab one level, spaces floored), blank lines inside an
// array or keyed tabular object, and a block whose first line is indented too
// deep, which then sets the block's depth. Every other decoding error applies
// in both modes.
func WithStrictMode(strict bool) DecoderOption {
	return func(o *decoderOptions) {
		o.strict = strict
	}
}

// WithDecoderIndent configures the expected indentation step. A value below 1
// is ignored.
func WithDecoderIndent(spaces int) DecoderOption {
	return func(o *decoderOptions) {
		if spaces > 0 {
			o.indentSize = spaces
		}
	}
}
