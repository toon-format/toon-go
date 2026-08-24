// Package format implements the string, key, and number rendering rules of the
// TOON specification (§2, §7).
package format

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Context carries the delimiter that governs delimiter-aware quoting (§11.1).
// Conforming encoders declare the document delimiter as the active delimiter of
// every header they emit, so a single value covers both roles.
type Context struct {
	Delimiter rune
}

func (c Context) delimiter() rune {
	return c.Delimiter
}

// numericLike matches the §7.2 numeric-like quoting trigger. It accepts a
// leading plus and leading zeros, both of which encoders must quote even though
// the decoder grammar of §4 rejects them as numbers.
var numericLike = regexp.MustCompile(`^[+-]?[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

// unquotedKey matches the §7.3 pattern for keys that may be emitted bare. The
// character class is ASCII-only, as the specification spells it out explicitly.
var unquotedKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)

// decoderNumber is the normative decoder number grammar of §4, before the
// forbidden-leading-zero check.
var decoderNumber = regexp.MustCompile(`^-?[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

// FormatString applies the TOON quoting rules to s.
func FormatString(s string, ctx Context) (string, error) {
	if err := ValidateString(s); err != nil {
		return "", err
	}
	if NeedsQuoting(s, ctx) {
		return QuoteString(s)
	}
	return s, nil
}

// NeedsQuoting reports whether the §7.2 rules require s to be quoted.
func NeedsQuoting(s string, ctx Context) bool {
	if s == "" {
		return true
	}
	if hasEdgeWhitespace(s) {
		return true
	}
	switch s {
	case "true", "false", "null":
		return true
	}
	if numericLike.MatchString(s) {
		return true
	}
	if strings.ContainsAny(s, ":\"\\[]{}") {
		return true
	}
	if hasControl(s) {
		return true
	}
	if d := ctx.delimiter(); d != 0 && strings.ContainsRune(s, d) {
		return true
	}
	switch s[0] {
	case '-', '#':
		return true
	}
	return false
}

func hasEdgeWhitespace(s string) bool {
	first := s[0]
	last := s[len(s)-1]
	return first == ' ' || first == '\t' || last == ' ' || last == '\t'
}

func hasControl(s string) bool {
	for _, r := range s {
		if r < 0x20 {
			return true
		}
	}
	return false
}

// QuoteString escapes s per the §7.1 encoder column and wraps it in quotes.
func QuoteString(s string) (string, error) {
	if err := ValidateString(s); err != nil {
		return "", err
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
				continue
			}
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String(), nil
}

// ValidateString rejects host strings that are not sequences of Unicode scalar
// values (§3). In Go this surfaces as invalid UTF-8, which includes the WTF-8
// encoding of an unpaired surrogate.
func ValidateString(s string) error {
	if utf8.ValidString(s) {
		return nil
	}
	return fmt.Errorf("toon: string is not valid UTF-8 (unpaired surrogate or malformed sequence)")
}

// EncodeKey renders a key, entry key, or field name per §7.3.
func EncodeKey(key string) (string, error) {
	if IsUnquotedKey(key) {
		return key, nil
	}
	return QuoteString(key)
}

// IsUnquotedKey reports whether key matches the §7.3 unquoted-key pattern.
func IsUnquotedKey(key string) bool {
	return unquotedKey.MatchString(key)
}

// FormatNumber renders f in the canonical form of §2. Values inside the
// canonical range use plain decimal notation; outside it, JSON exponent
// notation with a lowercase "e" and an explicit sign is emitted.
func FormatNumber(f float64) string {
	if f == 0 {
		return "0"
	}
	abs := math.Abs(f)
	if abs >= 1e-6 && abs < 1e21 {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return normalizeExponent(strconv.FormatFloat(f, 'e', -1, 64))
}

// normalizeExponent strips the zero padding Go adds to exponents so that
// 1e-07 becomes 1e-7, matching the reference implementation byte for byte.
func normalizeExponent(s string) string {
	idx := strings.IndexAny(s, "eE")
	if idx < 0 {
		return s
	}
	mantissa := s[:idx]
	exp := s[idx+1:]
	sign := "+"
	if len(exp) > 0 && (exp[0] == '+' || exp[0] == '-') {
		sign = string(exp[0])
		exp = exp[1:]
	}
	exp = strings.TrimLeft(exp, "0")
	if exp == "" {
		exp = "0"
	}
	return mantissa + "e" + sign + exp
}

// ParseNumberToken applies the normative decoder number grammar of §4. The
// boolean result reports whether the token is a number at all; tokens that are
// not are plain strings.
func ParseNumberToken(token string) (float64, bool) {
	if !decoderNumber.MatchString(token) {
		return 0, false
	}
	if hasForbiddenLeadingZeros(token) {
		return 0, false
	}
	f, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return 0, false
	}
	if f == 0 {
		// -0 decodes to 0 (§4).
		return 0, true
	}
	return f, true
}

// hasForbiddenLeadingZeros reports whether the integer part of token carries
// leading zeros that §4 forbids. A single "0" integer part followed by a
// fraction or exponent is allowed.
func hasForbiddenLeadingZeros(token string) bool {
	digits := strings.TrimPrefix(token, "-")
	if len(digits) < 2 || digits[0] != '0' {
		return false
	}
	return digits[1] >= '0' && digits[1] <= '9'
}
