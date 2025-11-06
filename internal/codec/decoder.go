package codec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	formatpkg "github.com/toon-format/toon-go/internal/format"
	parsepkg "github.com/toon-format/toon-go/internal/parse"
)

// Decoder parses TOON documents into Go values that match the data model from
// Section 2. Numbers are returned as float64, objects as map[string]any, and
// arrays as []any. Strings are unescaped per Section 7.1.
type Decoder struct {
	cfg decoderOptions
}

// NewDecoder constructs a Decoder with the given options.
func NewDecoder(opts ...DecoderOption) *Decoder {
	cfg := defaultDecoderOptions()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Decoder{cfg: cfg}
}

// Decode parses the provided TOON document.
func (d *Decoder) Decode(data []byte) (any, error) {
	parser, err := newParser(string(data), d.cfg)
	if err != nil {
		return nil, err
	}
	value, err := parser.parseDocument()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// DecodeString is a convenience wrapper around Decode.
func (d *Decoder) DecodeString(doc string) (any, error) {
	return d.Decode([]byte(doc))
}

// Decode uses a temporary decoder configured with opts.
func Decode(data []byte, opts ...DecoderOption) (any, error) {
	return NewDecoder(opts...).Decode(data)
}

// DecodeString decodes s using a temporary decoder.
func DecodeString(s string, opts ...DecoderOption) (any, error) {
	return NewDecoder(opts...).DecodeString(s)
}

type parser struct {
	lines []parsedLine
	pos   int
	cfg   decoderOptions
}

type parsedLine struct {
	number  int
	indent  int
	content string
	raw     string
	blank   bool
}

func newParser(input string, cfg decoderOptions) (*parser, error) {
	rawLines := splitLines(input)
	lines := make([]parsedLine, 0, len(rawLines))
	for idx, raw := range rawLines {
		if raw == "" {
			lines = append(lines, parsedLine{
				number:  idx + 1,
				indent:  0,
				content: "",
				raw:     "",
				blank:   true,
			})
			continue
		}
		indent, content, err := computeIndent(raw, cfg)
		if err != nil {
			return nil, errorWrap(idx+1, err)
		}
		lines = append(lines, parsedLine{
			number:  idx + 1,
			indent:  indent,
			content: content,
			raw:     raw,
			blank:   strings.TrimSpace(content) == "",
		})
	}
	return &parser{
		lines: lines,
		cfg:   cfg,
	}, nil
}

func splitLines(input string) []string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	lines := strings.Split(input, "\n")
	// Drop trailing empty line caused by final newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func computeIndent(line string, cfg decoderOptions) (int, string, error) {
	indent := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			indent++
		case '\t':
			if cfg.strict {
				return 0, "", errors.New("tabs are not allowed in indentation (strict mode)")
			}
			indent++
		default:
			content := line[i:]
			if cfg.strict && indent%cfg.indentSize != 0 {
				return 0, "", fmt.Errorf("indentation must be a multiple of %d spaces", cfg.indentSize)
			}
			return indent / cfg.indentSize, content, nil
		}
	}
	// Entire line whitespace.
	return 0, "", nil
}

func (p *parser) parseDocument() (any, error) {
	p.skipBlankLinesOutsideArrays()
	if p.pos >= len(p.lines) {
		return map[string]any{}, nil
	}

	nonBlank := p.countRemainingNonBlank()
	first := p.current()

	header, ok, err := tryParseHeader(first.content)
	if err != nil {
		return nil, errorWrap(first.number, err)
	}

	if nonBlank == 1 && !ok && !isKeyValue(first.content) {
		token := strings.TrimSpace(first.content)
		value, err := decodePrimitiveToken(token)
		if err != nil {
			return nil, errorWrap(first.number, err)
		}
		p.pos++
		return value, nil
	}

	if ok && first.indent == 0 && header.key == "" {
		p.pos++
		return p.parseArray(header, 0)
	}

	return p.parseObject(0)
}

func (p *parser) parseObject(depth int) (map[string]any, error) {
	result := make(map[string]any)
	for p.pos < len(p.lines) {
		line := p.current()
		if line.blank {
			p.pos++
			continue
		}
		if line.indent < depth {
			break
		}
		if line.indent > depth {
			return nil, errorAt(line.number, "unexpected indentation")
		}
		header, isHeader, err := tryParseHeader(line.content)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		if isHeader {
			if header.key == "" {
				return nil, errorAt(line.number, "arrays within objects must have a key")
			}
			p.pos++
			value, err := p.parseArray(header, depth)
			if err != nil {
				return nil, err
			}
			result[header.key] = value
			continue
		}

		key, rest, err := splitKeyValue(line.content)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		p.pos++
		if rest == "" {
			nextValue, err := p.parseObject(depth + 1)
			if err != nil {
				return nil, err
			}
			result[key] = nextValue
			continue
		}

		value, err := decodePrimitiveToken(rest)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		result[key] = value
	}
	return result, nil
}

func (p *parser) parseArray(header parsedHeader, depth int) (any, error) {
	delimiter := header.delimiter.rune()
	var values []any
	ctx := p.cfg

	if len(header.inlineValues) > 0 {
		raw, err := parsepkg.SplitInlineValues(header.inlineValues, delimiter)
		if err != nil {
			return nil, errorWrap(p.lines[p.pos-1].number, err)
		}
		for _, token := range raw {
			value, err := decodePrimitiveToken(token)
			if err != nil {
				return nil, errorWrap(p.lines[p.pos-1].number, err)
			}
			values = append(values, value)
		}
		if ctx.strict && len(values) != header.length {
			return nil, errorAtf(p.lines[p.pos-1].number, "inline array length mismatch; expected %d, got %d", header.length, len(values))
		}
		return values, nil
	}

	if len(header.fields) > 0 {
		rows := make([]any, 0, header.length)
		for p.pos < len(p.lines) {
			line := p.current()
			if line.blank {
				if ctx.strict {
					if nextIndent, ok := p.nextNonBlankIndent(p.pos); !ok || nextIndent <= depth {
						break
					}
					return nil, errorAt(line.number, "blank line inside tabular array")
				}
				p.pos++
				continue
			}
			if line.indent <= depth {
				break
			}
			if line.indent != depth+1 {
				return nil, errorAt(line.number, "invalid indentation for tabular row")
			}
			trimmed := strings.TrimSpace(line.content)
			if indexOutsideQuotes(trimmed, ':') != -1 {
				break
			}
			p.pos++
			raw, err := parsepkg.SplitInlineValues(trimmed, delimiter)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			if ctx.strict && len(raw) != len(header.fields) {
				return nil, errorAt(line.number, "tabular row width mismatch")
			}
			row := make(map[string]any, len(header.fields))
			for idx, field := range header.fields {
				if idx >= len(raw) {
					break
				}
				value, err := decodePrimitiveToken(raw[idx])
				if err != nil {
					return nil, errorWrap(line.number, err)
				}
				row[field] = value
			}
			rows = append(rows, row)
			if ctx.strict && len(rows) > header.length {
				return nil, errorAtf(line.number, "too many tabular rows (expected %d)", header.length)
			}
		}
		if ctx.strict && len(rows) != header.length {
			return nil, errorAtf(p.lines[p.pos-1].number, "tabular length mismatch; expected %d rows", header.length)
		}
		return rows, nil
	}

	values = make([]any, 0, header.length)
	for p.pos < len(p.lines) {
		line := p.current()
		if line.blank {
			if ctx.strict {
				if nextIndent, ok := p.nextNonBlankIndent(p.pos); !ok || nextIndent <= depth {
					break
				}
				return nil, errorAt(line.number, "blank line inside list array")
			}
			p.pos++
			continue
		}
		if line.indent <= depth {
			break
		}
		if line.indent != depth+1 {
			return nil, errorAt(line.number, "invalid indentation for list item")
		}
		if !strings.HasPrefix(line.content, "-") {
			break
		}
		itemContent := strings.TrimSpace(line.content[1:])
		p.pos++
		if itemContent == "" {
			values = append(values, map[string]any{})
			continue
		}

		if strings.HasPrefix(itemContent, "[") {
			itemHeader, ok, err := tryParseHeader(itemContent)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			if !ok {
				return nil, errorAt(line.number, "invalid array header in list item")
			}
			itemValue, err := p.parseArray(itemHeader, depth+1)
			if err != nil {
				return nil, err
			}
			values = append(values, itemValue)
			continue
		}

		if header, isHeader, err := tryParseHeader(itemContent); err != nil {
			return nil, errorWrap(line.number, err)
		} else if isHeader {
			if header.key == "" {
				return nil, errorAt(line.number, "arrays within objects must have a key")
			}
			arrayValue, err := p.parseArray(header, depth+1)
			if err != nil {
				return nil, err
			}
			obj := map[string]any{header.key: arrayValue}
			if err := p.collectObjectListSiblings(obj, depth); err != nil {
				return nil, err
			}
			values = append(values, obj)
			continue
		}

		if isKeyValue(itemContent) {
			key, rest, err := splitKeyValue(itemContent)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			if rest == "" {
				obj, err := p.parseObject(depth + 3)
				if err != nil {
					return nil, err
				}
				values = append(values, map[string]any{key: obj})
				continue
			}
			val, err := decodePrimitiveToken(rest)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			obj := map[string]any{key: val}
			if err := p.collectObjectListSiblings(obj, depth); err != nil {
				return nil, err
			}
			values = append(values, obj)
			continue
		}

		value, err := decodePrimitiveToken(itemContent)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		values = append(values, value)
	}

	if ctx.strict && len(values) != header.length {
		return nil, errorAtf(p.lines[p.pos-1].number, "list length mismatch; expected %d items", header.length)
	}
	return values, nil
}

func (p *parser) current() parsedLine {
	return p.lines[p.pos]
}

func (p *parser) skipBlankLinesOutsideArrays() {
	for p.pos < len(p.lines) {
		if !p.lines[p.pos].blank {
			break
		}
		p.pos++
	}
}

func (p *parser) countRemainingNonBlank() int {
	count := 0
	for _, line := range p.lines[p.pos:] {
		if !line.blank {
			count++
		}
	}
	return count
}

func (p *parser) nextNonBlankIndent(from int) (int, bool) {
	for i := from + 1; i < len(p.lines); i++ {
		if !p.lines[i].blank {
			return p.lines[i].indent, true
		}
	}
	return 0, false
}

func (p *parser) collectObjectListSiblings(obj map[string]any, depth int) error {
	for p.pos < len(p.lines) {
		next := p.current()
		if next.blank {
			if p.cfg.strict {
				if nextIndent, ok := p.nextNonBlankIndent(p.pos); !ok || nextIndent <= depth+1 {
					break
				}
				return errorAt(next.number, "blank line inside object list item")
			}
			p.pos++
			continue
		}
		if next.indent <= depth+1 {
			break
		}
		if next.indent != depth+2 {
			return errorAt(next.number, "invalid indentation for object list sibling")
		}
		if header, isHeader, err := tryParseHeader(next.content); err != nil {
			return errorWrap(next.number, err)
		} else if isHeader {
			p.pos++
			value, err := p.parseArray(header, depth+1)
			if err != nil {
				return err
			}
			if header.key == "" {
				return errorAt(next.number, "arrays within objects must have a key")
			}
			obj[header.key] = value
			continue
		}
		key, rest, err := splitKeyValue(next.content)
		if err != nil {
			return errorWrap(next.number, err)
		}
		p.pos++
		if rest == "" {
			nested, err := p.parseObject(depth + 3)
			if err != nil {
				return err
			}
			obj[key] = nested
		} else {
			value, err := decodePrimitiveToken(rest)
			if err != nil {
				return errorWrap(next.number, err)
			}
			obj[key] = value
		}
	}
	return nil
}

type parsedHeader struct {
	key          string
	length       int
	delimiter    Delimiter
	fields       []string
	inlineValues string
}

func tryParseHeader(content string) (parsedHeader, bool, error) {
	colon := indexOutsideQuotes(content, ':')
	if colon == -1 {
		return parsedHeader{}, false, nil
	}
	left := strings.TrimSpace(content[:colon])
	right := strings.TrimSpace(content[colon+1:])
	if left == "" {
		return parsedHeader{}, false, nil
	}
	bracketStart := indexOutsideQuotes(left, '[')
	if bracketStart == -1 {
		return parsedHeader{}, false, nil
	}
	rest := left[bracketStart+1:]
	bracketOffset := indexOutsideQuotes(rest, ']')
	if bracketOffset == -1 {
		return parsedHeader{}, false, errors.New("missing closing bracket in array header")
	}
	keyPart := strings.TrimSpace(left[:bracketStart])
	bracketSegment := rest[:bracketOffset]
	fieldSegment := strings.TrimSpace(rest[bracketOffset+1:])

	header := parsedHeader{
		key:       "",
		delimiter: DelimiterComma,
	}

	if keyPart != "" {
		key, err := decodeKeyToken(keyPart)
		if err != nil {
			return parsedHeader{}, false, err
		}
		header.key = key
	}

	length, delim, err := parseBracketSegment(bracketSegment)
	if err != nil {
		return parsedHeader{}, false, err
	}
	header.length = length
	header.delimiter = delim

	if fieldSegment != "" {
		if !strings.HasPrefix(fieldSegment, "{") || !strings.HasSuffix(fieldSegment, "}") {
			return parsedHeader{}, false, errors.New("invalid field segment in array header")
		}
		inner := fieldSegment[1 : len(fieldSegment)-1]
		if inner != "" {
			rawFields, err := parsepkg.SplitInlineValues(inner, delim.rune())
			if err != nil {
				return parsedHeader{}, false, err
			}
			fields := make([]string, 0, len(rawFields))
			for _, token := range rawFields {
				field, err := decodeKeyToken(token)
				if err != nil {
					return parsedHeader{}, false, err
				}
				fields = append(fields, field)
			}
			header.fields = fields
		}
	}

	header.inlineValues = right
	return header, true, nil
}

func parseBracketSegment(segment string) (int, Delimiter, error) {
	useMarker := false
	if strings.HasPrefix(segment, "#") {
		useMarker = true
		segment = segment[1:]
	}
	if segment == "" {
		return 0, DelimiterComma, errors.New("missing array length")
	}
	var digits strings.Builder
	var delim = DelimiterComma
	for _, r := range segment {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
			continue
		}
		switch r {
		case '\t':
			delim = DelimiterTab
		case '|':
			delim = DelimiterPipe
		default:
			return 0, DelimiterComma, fmt.Errorf("invalid delimiter symbol %q", r)
		}
	}
	lengthStr := digits.String()
	if lengthStr == "" {
		return 0, DelimiterComma, errors.New("missing digits in array length")
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return 0, DelimiterComma, err
	}
	_ = useMarker // marker is ignored semantically.
	return length, delim, nil
}

func splitKeyValue(content string) (string, string, error) {
	colon := indexOutsideQuotes(content, ':')
	if colon == -1 {
		return "", "", errors.New("missing colon after key")
	}
	keyToken := strings.TrimSpace(content[:colon])
	valueToken := strings.TrimSpace(content[colon+1:])
	key, err := decodeKeyToken(keyToken)
	if err != nil {
		return "", "", err
	}
	return key, valueToken, nil
}

func decodeKeyToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty key")
	}
	if token[0] == '"' {
		return parsepkg.UnquoteString(token)
	}
	if !formatpkg.IsValidUnquotedKey(token) {
		return "", fmt.Errorf("invalid unquoted key %q", token)
	}
	return token, nil
}

func decodePrimitiveToken(token string) (any, error) {
	if token == "" {
		return "", nil
	}
	if token[0] == '"' {
		return parsepkg.UnquoteString(token)
	}
	switch token {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null":
		return nil, nil
	}
	if hasForbiddenLeadingZeros(token) {
		return token, nil
	}
	if formatpkg.LooksNumeric(token) {
		num, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return nil, err
		}
		if num == 0 {
			num = 0
		}
		return num, nil
	}
	return token, nil
}

func hasForbiddenLeadingZeros(token string) bool {
	if len(token) < 2 {
		return false
	}
	if token[0] != '0' && (len(token) <= 1 || token[0] != '-' || token[1] != '0') {
		return false
	}
	// tokens like -0.x are legitimate numbers.
	if strings.Contains(token, ".") || strings.ContainsAny(token, "eE") {
		return false
	}
	if token[0] == '-' {
		return len(token) > 2 && token[1] == '0' && unicode.IsDigit(rune(token[2]))
	}
	return unicode.IsDigit(rune(token[1]))
}

func isKeyValue(content string) bool {
	return indexOutsideQuotes(content, ':') > 0
}

func indexOutsideQuotes(s string, target rune) int {
	inQuotes := false
	escaped := false
	for idx, r := range s {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inQuotes:
			escaped = true
		case r == '"':
			inQuotes = !inQuotes
		case !inQuotes && r == target:
			return idx
		}
	}
	return -1
}
