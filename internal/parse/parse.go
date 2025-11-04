package parse

import (
	"errors"
	"fmt"
	"strings"
)

// UnquoteString removes surrounding quotes and unescapes TOON strings.
func UnquoteString(token string) (string, error) {
	if len(token) < 2 || token[0] != '"' || token[len(token)-1] != '"' {
		return "", errors.New("invalid quoted string")
	}
	var b strings.Builder
	b.Grow(len(token) - 2)
	escaped := false
	for i := 1; i < len(token)-1; i++ {
		ch := token[i]
		if escaped {
			switch ch {
			case '\\', '"':
				b.WriteByte(ch)
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			default:
				return "", fmt.Errorf("invalid escape sequence \\%c", ch)
			}
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		b.WriteByte(ch)
	}
	if escaped {
		return "", errors.New("unterminated escape sequence")
	}
	return b.String(), nil
}

// SplitInlineValues tokenizes a delimiter-separated list, respecting quoted segments.
func SplitInlineValues(segment string, delimiter rune) ([]string, error) {
	if strings.TrimSpace(segment) == "" {
		return nil, nil
	}
	var tokens []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for _, r := range segment {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\' && inQuotes:
			current.WriteRune(r)
			escaped = true
		case r == '"':
			current.WriteRune(r)
			inQuotes = !inQuotes
		case r == delimiter && !inQuotes:
			tokens = append(tokens, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if inQuotes {
		return nil, errors.New("unterminated string in delimited values")
	}
	tokens = append(tokens, strings.TrimSpace(current.String()))
	return tokens, nil
}
