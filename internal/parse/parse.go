// Package parse implements the lexical helpers the TOON decoder needs:
// unescaping quoted tokens (§7.1), scanning for unquoted characters (§5.2), and
// splitting delimiter-separated cell sequences (§11.2).
package parse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// UnquoteString removes the surrounding quotes from token and applies the
// decoder column of the §7.1 escape table.
func UnquoteString(token string) (string, error) {
	if len(token) < 2 || token[0] != '"' || token[len(token)-1] != '"' {
		return "", errors.New("unterminated quoted string")
	}
	body := token[1 : len(token)-1]
	var b strings.Builder
	b.Grow(len(body))
	for i := 0; i < len(body); {
		ch := body[i]
		if ch != '\\' {
			r, size := utf8.DecodeRuneInString(body[i:])
			if r == utf8.RuneError && size == 1 {
				return "", errors.New("invalid UTF-8 in quoted string")
			}
			b.WriteString(body[i : i+size])
			i += size
			continue
		}
		i++
		if i >= len(body) {
			return "", errors.New("unterminated escape sequence")
		}
		switch body[i] {
		case '\\':
			b.WriteByte('\\')
		case '"':
			b.WriteByte('"')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'u':
			if i+4 >= len(body) {
				return "", errors.New(`\u escape requires four hex digits`)
			}
			hex := body[i+1 : i+5]
			code, err := strconv.ParseUint(hex, 16, 32)
			if err != nil {
				return "", fmt.Errorf(`invalid \u escape %q`, hex)
			}
			if code >= 0xD800 && code <= 0xDFFF {
				return "", fmt.Errorf(`surrogate escape \u%s is not allowed`, hex)
			}
			b.WriteRune(rune(code))
			i += 4
		default:
			return "", fmt.Errorf("invalid escape sequence \\%c", body[i])
		}
		i++
	}
	return b.String(), nil
}

// IsQuotedToken reports whether token starts with a double quote, which per
// §7.4 obliges it to be a complete quoted token.
func IsQuotedToken(token string) bool {
	return len(token) > 0 && token[0] == '"'
}

// ValidateQuotedToken enforces the §7.4 quoted-token boundary rule: the closing
// quote must be the token's last character.
func ValidateQuotedToken(token string) error {
	if !IsQuotedToken(token) {
		return nil
	}
	end := closingQuote(token)
	if end < 0 {
		return errors.New("unterminated quoted string")
	}
	if end != len(token)-1 {
		return errors.New("unexpected content after closing quote")
	}
	return nil
}

// closingQuote returns the index of the quote that closes the quoted token
// starting at index 0, or -1 when the token is unterminated.
func closingQuote(token string) int {
	for i := 1; i < len(token); i++ {
		switch token[i] {
		case '\\':
			i++
		case '"':
			return i
		}
	}
	return -1
}

// IndexUnquoted reports the byte index of the first occurrence of target that
// lies outside a quoted region, or -1.
func IndexUnquoted(s string, target rune) int {
	inQuotes := false
	for i := 0; i < len(s); {
		c := s[i]
		if inQuotes {
			switch c {
			case '\\':
				i += 2
			case '"':
				inQuotes = false
				i++
			default:
				i++
			}
			continue
		}
		if c == '"' {
			inQuotes = true
			i++
			continue
		}
		if c < 0x80 && rune(c) == target {
			return i
		}
		i++
	}
	return -1
}

// SplitDelimited splits a cell sequence on delimiter, honouring quoted regions,
// preserving empty tokens, and trimming the surrounding U+0020 of each token
// (§11.2, §12). A sequence that trims to nothing yields zero cells.
func SplitDelimited(segment string, delimiter rune) ([]string, error) {
	if strings.Trim(segment, " ") == "" {
		return nil, nil
	}
	tokens := make([]string, 0, 4)
	var current strings.Builder
	inQuotes := false
	escaped := false
	for _, r := range segment {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case inQuotes && r == '\\':
			current.WriteRune(r)
			escaped = true
		case r == '"':
			current.WriteRune(r)
			inQuotes = !inQuotes
		case !inQuotes && r == delimiter:
			tokens = append(tokens, strings.Trim(current.String(), " "))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if inQuotes {
		return nil, errors.New("unterminated quoted string")
	}
	tokens = append(tokens, strings.Trim(current.String(), " "))
	return tokens, nil
}
