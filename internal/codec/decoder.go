package codec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	formatpkg "github.com/toon-format/toon-go/internal/format"
	parsepkg "github.com/toon-format/toon-go/internal/parse"
)

// Decoder parses TOON documents into Go values that match the data model of §2.
// Numbers are returned as float64, objects as map[string]any, and arrays as
// []any. Because Go maps do not retain insertion order, the decoder does not
// preserve document key order; §2 requires this deviation to be documented.
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
	return d.DecodeString(string(data))
}

// DecodeString parses the provided TOON document.
func (d *Decoder) DecodeString(doc string) (any, error) {
	lines, err := prepareLines(doc, d.cfg)
	if err != nil {
		return nil, err
	}
	p := &parser{lines: lines, cfg: d.cfg}
	return p.parseDocument()
}

// Decode uses a temporary decoder configured with opts.
func Decode(data []byte, opts ...DecoderOption) (any, error) {
	return NewDecoder(opts...).Decode(data)
}

// DecodeString decodes s using a temporary decoder.
func DecodeString(s string, opts ...DecoderOption) (any, error) {
	return NewDecoder(opts...).DecodeString(s)
}

// docLine is one line of the comment-stripped document (§5.1).
type docLine struct {
	number  int
	depth   int
	content string
	blank   bool
}

// prepareLines removes the byte-order mark, normalizes line terminators, strips
// comment lines and trailing spaces, and computes each line's depth (§5.1, §12).
func prepareLines(input string, cfg decoderOptions) ([]docLine, error) {
	input = strings.TrimPrefix(input, "\ufeff")
	raw := strings.Split(input, "\n")
	lines := make([]docLine, 0, len(raw))

	for i, text := range raw {
		number := i + 1
		text = strings.TrimSuffix(text, "\r")
		text = strings.TrimRight(text, " ")

		if isCommentLine(text) {
			continue
		}
		if text == "" {
			lines = append(lines, docLine{number: number, blank: true})
			continue
		}

		depth, content, err := lineDepth(text, cfg)
		if err != nil {
			return nil, errorWrap(number, err)
		}
		lines = append(lines, docLine{number: number, depth: depth, content: content})
	}

	// A final newline produces a trailing empty line that carries no meaning.
	if n := len(lines); n > 0 && lines[n-1].blank {
		lines = lines[:n-1]
	}
	return lines, nil
}

// isCommentLine reports whether the line's first character after zero or more
// spaces is "#" (§5.1). A tab in the leading whitespace disqualifies it.
func isCommentLine(text string) bool {
	i := 0
	for i < len(text) && text[i] == ' ' {
		i++
	}
	return i < len(text) && text[i] == '#'
}

func lineDepth(text string, cfg decoderOptions) (int, string, error) {
	spaces, tabs, offset := 0, 0, 0
scan:
	for offset < len(text) {
		switch text[offset] {
		case ' ':
			spaces++
		case '\t':
			tabs++
		default:
			break scan
		}
		offset++
	}
	if cfg.strict {
		if tabs > 0 {
			return 0, "", errors.New("tabs are not allowed in indentation")
		}
		if spaces%cfg.indentSize != 0 {
			return 0, "", fmt.Errorf("indentation must be a multiple of %d spaces", cfg.indentSize)
		}
		return spaces / cfg.indentSize, text[offset:], nil
	}
	// Non-strict depth for tab indentation is implementation-defined; each
	// leading tab counts as one level (§12).
	return spaces/cfg.indentSize + tabs, text[offset:], nil
}

type parser struct {
	lines []docLine
	pos   int
	cfg   decoderOptions
	spans []*spanState
}

// spanState tracks an open header span: the scope's content depth and how many
// items, rows, or entry rows it has consumed so far (§12).
type spanState struct {
	contentDepth int
	consumed     int
}

func (p *parser) pushSpan(contentDepth int) *spanState {
	span := &spanState{contentDepth: contentDepth}
	p.spans = append(p.spans, span)
	return span
}

func (p *parser) popSpan() {
	p.spans = p.spans[:len(p.spans)-1]
}

// checkBlank rejects a blank line that falls inside an open header span. The
// span reaches through the last line of the scope's content, which may sit
// deeper than the scope itself, so every open span is consulted.
func (p *parser) checkBlank() error {
	if !p.cfg.strict {
		return nil
	}
	idx, ok := p.nextNonBlank(p.pos + 1)
	if !ok {
		return nil
	}
	next := p.lines[idx]
	for _, span := range p.spans {
		if span.consumed > 0 && next.depth >= span.contentDepth {
			return errorAt(p.lines[p.pos].number, "blank line inside a header span")
		}
	}
	return nil
}

// scopeEndsAtBlank reports whether the scope at contentDepth ends at the blank
// line under the cursor.
func (p *parser) scopeEndsAtBlank(contentDepth int) bool {
	idx, ok := p.nextNonBlank(p.pos + 1)
	return !ok || p.lines[idx].depth < contentDepth
}

func (p *parser) nextNonBlank(from int) (int, bool) {
	for i := from; i < len(p.lines); i++ {
		if !p.lines[i].blank {
			return i, true
		}
	}
	return 0, false
}

func (p *parser) countNonBlank() int {
	count := 0
	for _, line := range p.lines {
		if !line.blank {
			count++
		}
	}
	return count
}

// parseDocument applies the root-form discovery rules of §5.
func (p *parser) parseDocument() (any, error) {
	idx, ok := p.nextNonBlank(0)
	if !ok {
		return map[string]any{}, nil
	}
	p.pos = idx
	line := p.lines[idx]

	if line.content == "[]" {
		p.pos++
		if err := p.checkTrailing(); err != nil {
			return nil, err
		}
		return []any{}, nil
	}

	hdr, status := parseHeaderSyntax(line.content, p.cfg.strict)
	hdr.number = line.number
	switch status {
	case headerOK:
		if !hdr.hasKey {
			p.pos++
			var value any
			var err error
			if hdr.keyed {
				value, err = p.parseKeyedRows(hdr, line.depth+1)
			} else {
				value, err = p.parseArrayBody(hdr, line.depth+1)
			}
			if err != nil {
				return nil, err
			}
			if err := p.checkTrailing(); err != nil {
				return nil, err
			}
			return value, nil
		}
	case headerMalformed:
		if p.cfg.strict {
			return nil, errorAt(line.number, "malformed array header")
		}
	}

	if status == headerNotHeader {
		if p.countNonBlank() == 1 && parsepkg.IndexUnquoted(line.content, ':') < 0 {
			p.pos++
			value, err := decodeValueToken(line.content)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			return value, nil
		}
	}

	result := map[string]any{}
	seen := map[string]bool{}
	if err := p.parseObjectInto(result, seen, 0); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *parser) checkTrailing() error {
	idx, ok := p.nextNonBlank(p.pos)
	if !ok {
		return nil
	}
	if p.cfg.strict {
		return errorAt(p.lines[idx].number, "trailing content after the root form")
	}
	p.pos = len(p.lines)
	return nil
}

// parseObjectInto fills result with the fields of an object scope whose content
// stands at depth (§8).
func (p *parser) parseObjectInto(result map[string]any, seen map[string]bool, depth int) error {
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if line.blank {
			if err := p.checkBlank(); err != nil {
				return err
			}
			p.pos++
			continue
		}
		if line.depth < depth {
			return nil
		}
		if line.depth > depth {
			if isScalarLine(line.content, p.cfg.strict) {
				// A scalar line outside root primitive position is an error in
				// any mode (§5.2, §14.2).
				return errorAt(line.number, "unexpected scalar line")
			}
			if p.cfg.strict {
				return errorAt(line.number, "unexpected indentation")
			}
			p.pos++
			continue
		}

		hdr, status := parseHeaderSyntax(line.content, p.cfg.strict)
		hdr.number = line.number
		if status == headerOK && hdr.hasKey {
			p.pos++
			var value any
			var err error
			if hdr.keyed {
				value, err = p.parseKeyedRows(hdr, depth+1)
			} else {
				value, err = p.parseArrayBody(hdr, depth+1)
			}
			if err != nil {
				return err
			}
			if err := p.assign(result, seen, hdr.key, value, line.number); err != nil {
				return err
			}
			continue
		}
		if p.cfg.strict {
			switch {
			case status == headerOK && !hdr.hasKey:
				return errorAt(line.number, "keyless array header is valid only at the document root")
			case status == headerMalformed:
				return errorAt(line.number, "malformed array header")
			}
		}

		key, value, err := p.parseKeyValueLine(line, depth)
		if err != nil {
			return err
		}
		if err := p.assign(result, seen, key, value, line.number); err != nil {
			return err
		}
	}
	return nil
}

// parseKeyValueLine consumes a key-value line whose key stands at depth and
// returns its decoded key and value (§8).
func (p *parser) parseKeyValueLine(line docLine, depth int) (string, any, error) {
	colon := parsepkg.IndexUnquoted(line.content, ':')
	if colon < 0 {
		// A scalar line outside root primitive position is an error in any mode.
		return "", nil, errorAt(line.number, "expected a key-value line")
	}
	key, err := decodeKeyToken(strings.Trim(line.content[:colon], " "))
	if err != nil {
		return "", nil, errorWrap(line.number, err)
	}
	rest := strings.Trim(line.content[colon+1:], " ")
	p.pos++

	switch rest {
	case "":
		nested, err := p.parseNestedObject(depth + 1)
		if err != nil {
			return "", nil, err
		}
		return key, nested, nil
	case "[]":
		return key, []any{}, nil
	}

	value, err := decodeValueToken(rest)
	if err != nil {
		return "", nil, errorWrap(line.number, err)
	}
	return key, value, nil
}

func (p *parser) parseNestedObject(depth int) (map[string]any, error) {
	result := map[string]any{}
	idx, ok := p.nextNonBlank(p.pos)
	if !ok || p.lines[idx].depth < depth {
		return result, nil
	}
	if p.lines[idx].depth > depth && p.cfg.strict {
		return nil, errorAt(p.lines[idx].number, "indentation depth jump")
	}
	seen := map[string]bool{}
	if err := p.parseObjectInto(result, seen, depth); err != nil {
		return nil, err
	}
	return result, nil
}

// assign stores a decoded field, applying the duplicate-key rules of §14.3.
func (p *parser) assign(result map[string]any, seen map[string]bool, key string, value any, number int) error {
	if seen[key] {
		if p.cfg.strict {
			return errorAtf(number, "duplicate sibling key %q", key)
		}
	}
	seen[key] = true
	result[key] = value
	return nil
}

// parseArrayBody decodes the value of a non-keyed header whose scope content
// stands at contentDepth (§9.1–§9.4).
func (p *parser) parseArrayBody(hdr header, contentDepth int) (any, error) {
	if hdr.fields != nil {
		return p.parseTabularRows(hdr, contentDepth)
	}
	if hdr.inline != "" {
		return p.parseInlineValues(hdr)
	}
	return p.parseListItems(hdr, contentDepth)
}

func (p *parser) parseInlineValues(hdr header) (any, error) {
	tokens, err := parsepkg.SplitDelimited(hdr.inline, hdr.delimiter.rune())
	if err != nil {
		return nil, errorWrap(hdr.number, err)
	}
	values := make([]any, 0, len(tokens))
	for _, token := range tokens {
		value, err := decodeValueToken(token)
		if err != nil {
			return nil, errorWrap(hdr.number, err)
		}
		values = append(values, value)
	}
	if p.cfg.strict && len(values) != hdr.length {
		return nil, errorAtf(hdr.number, "inline array declares %d values but carries %d", hdr.length, len(values))
	}
	return values, nil
}

func (p *parser) parseTabularRows(hdr header, contentDepth int) (any, error) {
	leaves := leafCount(hdr.fields)
	rows := []any{}
	delimiter := hdr.delimiter.rune()
	span := p.pushSpan(contentDepth)
	defer p.popSpan()

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if line.blank {
			if err := p.checkBlank(); err != nil {
				return nil, err
			}
			ends := p.scopeEndsAtBlank(contentDepth)
			p.pos++
			if ends {
				break
			}
			continue
		}
		if line.depth < contentDepth {
			break
		}
		if line.depth > contentDepth {
			if p.cfg.strict {
				return nil, errorAt(line.number, "unexpected indentation in tabular scope")
			}
			p.pos++
			continue
		}
		if !isRowLine(line.content, delimiter) {
			break
		}

		p.pos++
		cells, err := parsepkg.SplitDelimited(line.content, delimiter)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		if p.cfg.strict && len(cells) != leaves {
			return nil, errorAtf(line.number, "tabular row carries %d cells but the header declares %d", len(cells), leaves)
		}
		index := 0
		row, err := buildFromCells(hdr.fields, cells, &index)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		rows = append(rows, row)
		span.consumed++
	}

	if p.cfg.strict && len(rows) != hdr.length {
		return nil, errorAtf(hdr.number, "tabular array declares %d rows but carries %d", hdr.length, len(rows))
	}
	return rows, nil
}

// isRowLine applies the §9.3 row versus key-value disambiguation.
func isRowLine(content string, delimiter rune) bool {
	colon := parsepkg.IndexUnquoted(content, ':')
	if colon < 0 {
		return true
	}
	delim := parsepkg.IndexUnquoted(content, delimiter)
	return delim >= 0 && delim < colon
}

func (p *parser) parseKeyedRows(hdr header, contentDepth int) (any, error) {
	leaves := leafCount(hdr.fields)
	result := map[string]any{}
	seen := map[string]bool{}
	delimiter := hdr.delimiter.rune()
	count := 0
	span := p.pushSpan(contentDepth)
	defer p.popSpan()

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if line.blank {
			if err := p.checkBlank(); err != nil {
				return nil, err
			}
			ends := p.scopeEndsAtBlank(contentDepth)
			p.pos++
			if ends {
				break
			}
			continue
		}
		if line.depth < contentDepth {
			break
		}
		if line.depth > contentDepth {
			if p.cfg.strict {
				return nil, errorAt(line.number, "unexpected indentation in keyed tabular scope")
			}
			p.pos++
			continue
		}

		colon := parsepkg.IndexUnquoted(line.content, ':')
		if colon < 0 {
			if p.cfg.strict {
				return nil, errorAt(line.number, "entry row is missing a colon")
			}
			p.pos++
			continue
		}

		p.pos++
		key, err := decodeKeyToken(strings.Trim(line.content[:colon], " "))
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		cells, err := parsepkg.SplitDelimited(strings.Trim(line.content[colon+1:], " "), delimiter)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		if p.cfg.strict && len(cells) != leaves {
			return nil, errorAtf(line.number, "entry row carries %d cells but the header declares %d", len(cells), leaves)
		}
		index := 0
		value, err := buildFromCells(hdr.fields, cells, &index)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		if err := p.assign(result, seen, key, value, line.number); err != nil {
			return nil, err
		}
		count++
		span.consumed++
	}

	if p.cfg.strict && count != hdr.length {
		return nil, errorAtf(hdr.number, "keyed tabular object declares %d entries but carries %d", hdr.length, count)
	}
	return result, nil
}

func (p *parser) parseListItems(hdr header, contentDepth int) (any, error) {
	items := []any{}
	span := p.pushSpan(contentDepth)
	defer p.popSpan()

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if line.blank {
			if err := p.checkBlank(); err != nil {
				return nil, err
			}
			ends := p.scopeEndsAtBlank(contentDepth)
			p.pos++
			if ends {
				break
			}
			continue
		}
		if line.depth < contentDepth {
			break
		}
		if line.depth > contentDepth {
			if p.cfg.strict {
				return nil, errorAt(line.number, "unexpected indentation in list scope")
			}
			p.pos++
			continue
		}
		if line.content != "-" && !strings.HasPrefix(line.content, "- ") {
			break
		}

		p.pos++
		span.consumed++
		item, err := p.parseListItem(line, contentDepth)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if p.cfg.strict && len(items) != hdr.length {
		return nil, errorAtf(hdr.number, "array declares %d items but carries %d", hdr.length, len(items))
	}
	return items, nil
}

// parseListItem decodes one list item whose hyphen stands at itemDepth (§9.4,
// §10). A keyed first field carried on the hyphen line stands at itemDepth+1,
// so any scope it opens has its content at itemDepth+2.
func (p *parser) parseListItem(line docLine, itemDepth int) (any, error) {
	rest := strings.TrimLeft(strings.TrimPrefix(line.content, "-"), " ")

	switch rest {
	case "":
		return map[string]any{}, nil
	case "[]":
		return []any{}, nil
	}

	hdr, status := parseHeaderSyntax(rest, p.cfg.strict)
	hdr.number = line.number

	if status == headerOK {
		if !hdr.hasKey {
			if hdr.fields != nil || hdr.keyed {
				if p.cfg.strict {
					return nil, errorAt(line.number, "a keyless fields-bearing header is valid only at the document root")
				}
			} else {
				return p.parseArrayBody(hdr, itemDepth+1)
			}
		} else {
			result := map[string]any{}
			seen := map[string]bool{}
			var value any
			var err error
			if hdr.keyed {
				value, err = p.parseKeyedRows(hdr, itemDepth+2)
			} else {
				value, err = p.parseArrayBody(hdr, itemDepth+2)
			}
			if err != nil {
				return nil, err
			}
			if err := p.assign(result, seen, hdr.key, value, line.number); err != nil {
				return nil, err
			}
			if err := p.parseObjectInto(result, seen, itemDepth+1); err != nil {
				return nil, err
			}
			return result, nil
		}
	} else if status == headerMalformed && p.cfg.strict {
		return nil, errorAt(line.number, "malformed array header")
	}

	colon := parsepkg.IndexUnquoted(rest, ':')
	if colon < 0 {
		value, err := decodeValueToken(rest)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		return value, nil
	}

	key, err := decodeKeyToken(strings.Trim(rest[:colon], " "))
	if err != nil {
		return nil, errorWrap(line.number, err)
	}
	valueToken := strings.Trim(rest[colon+1:], " ")

	var value any
	switch valueToken {
	case "":
		nested, nErr := p.parseNestedObject(itemDepth + 2)
		if nErr != nil {
			return nil, nErr
		}
		value = nested
	case "[]":
		value = []any{}
	default:
		value, err = decodeValueToken(valueToken)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
	}

	result := map[string]any{}
	seen := map[string]bool{}
	if err := p.assign(result, seen, key, value, line.number); err != nil {
		return nil, err
	}
	if err := p.parseObjectInto(result, seen, itemDepth+1); err != nil {
		return nil, err
	}
	return result, nil
}

// isScalarLine reports whether content is a scalar line under the §5.2 line
// classification: not a list item, not a header, and carrying no unquoted colon.
func isScalarLine(content string, strict bool) bool {
	if content == "-" || strings.HasPrefix(content, "- ") {
		return false
	}
	if _, status := parseHeaderSyntax(content, strict); status == headerOK {
		return false
	}
	return parsepkg.IndexUnquoted(content, ':') < 0
}

// buildFromCells materializes one row or entry value by walking the field list
// depth-first (§9.3).
func buildFromCells(nodes []fieldNode, cells []string, index *int) (map[string]any, error) {
	result := make(map[string]any, len(nodes))
	for _, node := range nodes {
		if node.children != nil {
			sub, err := buildFromCells(node.children, cells, index)
			if err != nil {
				return nil, err
			}
			result[node.name] = sub
			continue
		}
		if *index < len(cells) {
			value, err := decodeValueToken(cells[*index])
			if err != nil {
				return nil, err
			}
			result[node.name] = value
		}
		*index++
	}
	return result, nil
}

func leafCount(nodes []fieldNode) int {
	total := 0
	for _, node := range nodes {
		if node.children == nil {
			total++
			continue
		}
		total += leafCount(node.children)
	}
	return total
}

func decodeKeyToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("missing key before colon")
	}
	if parsepkg.IsQuotedToken(token) {
		if err := parsepkg.ValidateQuotedToken(token); err != nil {
			return "", err
		}
		return parsepkg.UnquoteString(token)
	}
	return token, nil
}

// decodeValueToken maps a primitive token to a host value per §4. The empty
// array form is positional and is therefore handled by the caller.
func decodeValueToken(token string) (any, error) {
	if token == "" {
		return "", nil
	}
	if parsepkg.IsQuotedToken(token) {
		if err := parsepkg.ValidateQuotedToken(token); err != nil {
			return nil, err
		}
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
	if number, ok := formatpkg.ParseNumberToken(token); ok {
		return number, nil
	}
	return token, nil
}

// header is a parsed array, tabular, or keyed header (§6).
type header struct {
	key       string
	hasKey    bool
	length    int
	keyed     bool
	delimiter Delimiter
	fields    []fieldNode
	inline    string
	number    int
}

type headerStatus int

const (
	// headerNotHeader means the line is not a header and falls through to the
	// key-value class (§5.2).
	headerNotHeader headerStatus = iota
	// headerMalformed means the line opens like a header but violates §6.
	headerMalformed
	headerOK
)

// parseHeaderSyntax applies the header grammar of §6 to a line's content.
func parseHeaderSyntax(content string, strict bool) (header, headerStatus) {
	bracket := parsepkg.IndexUnquoted(content, '[')
	if bracket < 0 {
		return header{}, headerNotHeader
	}
	if colon := parsepkg.IndexUnquoted(content, ':'); colon >= 0 && colon < bracket {
		return header{}, headerNotHeader
	}

	hdr := header{delimiter: DelimiterComma}
	keyPart := content[:bracket]
	if strings.TrimRight(keyPart, " \t") != keyPart {
		// Whitespace between a key and its bracket segment (§6).
		return header{}, headerMalformed
	}
	if keyPart != "" {
		key, err := decodeKeyToken(keyPart)
		if err != nil {
			return header{}, headerMalformed
		}
		hdr.key = key
		hdr.hasKey = true
	}

	rest := content[bracket+1:]
	close := strings.IndexByte(rest, ']')
	if close < 0 {
		return header{}, headerMalformed
	}
	length, keyed, delimiter, ok := parseBracketSegment(rest[:close])
	if !ok {
		return header{}, headerMalformed
	}
	hdr.length = length
	hdr.keyed = keyed
	hdr.delimiter = delimiter

	after := rest[close+1:]
	if strings.HasPrefix(after, "{") {
		end := matchBrace(after)
		if end < 0 {
			return header{}, headerMalformed
		}
		fields, err := parseFieldList(after[1:end], delimiter, strict)
		if err != nil {
			return header{}, headerMalformed
		}
		hdr.fields = fields
		after = after[end+1:]
	}

	if !strings.HasPrefix(after, ":") {
		return header{}, headerMalformed
	}
	hdr.inline = strings.Trim(after[1:], " ")

	if hdr.keyed && hdr.fields == nil {
		return header{}, headerMalformed
	}
	if hdr.fields != nil && hdr.inline != "" {
		// A fields-bearing header carries no inline content (§6).
		return header{}, headerMalformed
	}
	return hdr, headerOK
}

// parseBracketSegment parses "[N]", "[N<delim>]", "[N:]", or "[N:<delim>]".
func parseBracketSegment(segment string) (int, bool, Delimiter, bool) {
	digits := 0
	for digits < len(segment) && segment[digits] >= '0' && segment[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false, DelimiterComma, false
	}
	if digits > 1 && segment[0] == '0' {
		// Leading zeros are not a canonical length (§6).
		return 0, false, DelimiterComma, false
	}
	length, err := strconv.Atoi(segment[:digits])
	if err != nil {
		return 0, false, DelimiterComma, false
	}

	rest := segment[digits:]
	keyed := false
	if strings.HasPrefix(rest, ":") {
		keyed = true
		rest = rest[1:]
	}

	delimiter := DelimiterComma
	switch rest {
	case "":
	case "\t":
		delimiter = DelimiterTab
	case "|":
		delimiter = DelimiterPipe
	default:
		return 0, false, DelimiterComma, false
	}
	return length, keyed, delimiter, true
}

// matchBrace returns the index of the "}" that closes the field list starting at
// index 0, ignoring braces inside quoted names (§6).
func matchBrace(s string) int {
	depth := 0
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuotes {
			switch c {
			case '\\':
				i++
			case '"':
				inQuotes = false
			}
			continue
		}
		switch c {
		case '"':
			inQuotes = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// parseFieldList parses the entries of a field list, recursing into nested
// field groups (§6, §9.3).
func parseFieldList(body string, delimiter Delimiter, strict bool) ([]fieldNode, error) {
	if body == "" {
		return nil, errors.New("empty field list")
	}
	if err := checkForeignDelimiters(body, delimiter); err != nil {
		return nil, err
	}
	entries, err := splitFieldEntries(body, delimiter.rune())
	if err != nil {
		return nil, err
	}

	nodes := make([]fieldNode, 0, len(entries))
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry == "" {
			return nil, errors.New("empty field entry")
		}
		namePart := entry
		var children []fieldNode
		if brace := topLevelBrace(entry); brace >= 0 {
			end := matchBrace(entry[brace:])
			if end < 0 || brace+end != len(entry)-1 {
				return nil, errors.New("malformed nested field group")
			}
			nested, err := parseFieldList(entry[brace+1:len(entry)-1], delimiter, strict)
			if err != nil {
				return nil, err
			}
			children = nested
			namePart = entry[:brace]
		}
		name, err := decodeKeyToken(namePart)
		if err != nil {
			return nil, err
		}
		if seen[name] && strict {
			// Non-strict mode keeps the duplicate so that the field walk of
			// §9.3 yields last-write-wins in every decoded element (§14.3).
			return nil, fmt.Errorf("duplicate field name %q", name)
		}
		seen[name] = true
		nodes = append(nodes, fieldNode{name: name, children: children})
	}
	return nodes, nil
}

// checkForeignDelimiters rejects a field list that uses an unquoted delimiter
// character other than the one declared by the bracket segment (§6, §14.2).
func checkForeignDelimiters(body string, delimiter Delimiter) error {
	for _, candidate := range []Delimiter{DelimiterComma, DelimiterTab, DelimiterPipe} {
		if candidate == delimiter {
			continue
		}
		if parsepkg.IndexUnquoted(body, candidate.rune()) >= 0 {
			return fmt.Errorf("field list uses %s but the header declares %s", candidate, delimiter)
		}
	}
	return nil
}

// topLevelBrace returns the index of the entry's own nested field group, or -1.
func topLevelBrace(entry string) int {
	inQuotes := false
	for i := 0; i < len(entry); i++ {
		c := entry[i]
		if inQuotes {
			switch c {
			case '\\':
				i++
			case '"':
				inQuotes = false
			}
			continue
		}
		switch c {
		case '"':
			inQuotes = true
		case '{':
			return i
		}
	}
	return -1
}

// splitFieldEntries splits a field list on delimiter, honouring quotes and
// nested field groups.
func splitFieldEntries(body string, delimiter rune) ([]string, error) {
	entries := make([]string, 0, 4)
	var current strings.Builder
	depth := 0
	inQuotes := false

	for i := 0; i < len(body); i++ {
		c := body[i]
		if inQuotes {
			current.WriteByte(c)
			switch c {
			case '\\':
				if i+1 < len(body) {
					i++
					current.WriteByte(body[i])
				}
			case '"':
				inQuotes = false
			}
			continue
		}
		switch {
		case c == '"':
			inQuotes = true
			current.WriteByte(c)
		case c == '{':
			depth++
			current.WriteByte(c)
		case c == '}':
			depth--
			if depth < 0 {
				return nil, errors.New("unmatched brace in field list")
			}
			current.WriteByte(c)
		case depth == 0 && c < 0x80 && rune(c) == delimiter:
			entries = append(entries, current.String())
			current.Reset()
		default:
			current.WriteByte(c)
		}
	}
	if inQuotes {
		return nil, errors.New("unterminated quoted field name")
	}
	if depth != 0 {
		return nil, errors.New("unmatched brace in field list")
	}
	entries = append(entries, current.String())
	return entries, nil
}
