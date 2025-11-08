package format

import (
	"fmt"
	"strings"
	"unicode"
)

// Context captures delimiter information for quoting decisions.
type Context struct {
	Active   rune
	Document rune
	InArray  bool
}

// FormatString applies TOON quoting rules to the provided string.
func FormatString(s string, ctx Context) (string, error) {
	if err := ValidateCharacters(s); err != nil {
		return "", err
	}
	if NeedsQuoting(s, ctx) {
		return QuoteString(s)
	}
	return s, nil
}

// NeedsQuoting reports whether the string must be quoted in the supplied context.
func NeedsQuoting(s string, ctx Context) bool {
	if len(s) == 0 {
		return true
	}
	if strings.TrimSpace(s) != s {
		return true
	}
	switch s {
	case "true", "false", "null":
		return true
	}
	if LooksNumeric(s) {
		return true
	}
	if HasLeadingZeroDecimal(s) {
		return true
	}
	if strings.ContainsAny(s, ":\\\"[]{}") {
		return true
	}
	if strings.ContainsRune(s, '\n') || strings.ContainsRune(s, '\r') || strings.ContainsRune(s, '\t') {
		return true
	}
	if strings.HasPrefix(s, "-") {
		return true
	}
	if ctx.InArray && ctx.Active != 0 && strings.ContainsRune(s, ctx.Active) {
		return true
	}
	if !ctx.InArray && ctx.Document != 0 && strings.ContainsRune(s, ctx.Document) {
		return true
	}
	return false
}

// QuoteString escapes and wraps the string in double quotes.
func QuoteString(s string) (string, error) {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			if r < 0x20 {
				return "", fmt.Errorf("toon: unsupported control character U+%04X in string", r)
			}
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String(), nil
}

// ValidateCharacters ensures the string does not contain unsupported control characters.
func ValidateCharacters(s string) error {
	for _, r := range s {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("toon: unsupported control character U+%04X in string", r)
		}
	}
	return nil
}

// LooksNumeric reports whether the string resembles a numeric literal.
func LooksNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	i := 0
	if s[0] == '-' {
		i++
		if i == len(s) {
			return false
		}
	}
	digits := 0
	for i < len(s) && isDigit(s[i]) {
		i++
		digits++
	}
	if digits == 0 {
		return false
	}
	if i < len(s) && s[i] == '.' {
		i++
		if i == len(s) || !isDigit(s[i]) {
			return false
		}
		for i < len(s) && isDigit(s[i]) {
			i++
		}
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if i == len(s) || !isDigit(s[i]) {
			return false
		}
		for i < len(s) && isDigit(s[i]) {
			i++
		}
	}
	return i == len(s)
}

// HasLeadingZeroDecimal reports whether the string is a decimal with forbidden leading zeros.
func HasLeadingZeroDecimal(s string) bool {
	if len(s) < 2 {
		return false
	}
	if s[0] != '0' {
		return false
	}
	return s[1] >= '0' && s[1] <= '9'
}

// EncodeKey applies TOON key quoting rules.
func EncodeKey(key string) (string, error) {
	if key == "" {
		return QuoteString(key)
	}
	if IsValidUnquotedKey(key) {
		return key, nil
	}
	return QuoteString(key)
}

// IsValidUnquotedKey reports whether the key satisfies the identifier pattern.
func IsValidUnquotedKey(key string) bool {
	if key == "" {
		return false
	}
	for pos, r := range key {
		if pos == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
