package toon

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	formatpkg "github.com/toon-format/toon-go/internal/format"
)

// Encoder serializes Go values as TOON documents. The zero value encodes with
// the default options. An Encoder is safe for concurrent use.
type Encoder struct {
	opts []EncoderOption
}

// NewEncoder constructs an Encoder using the supplied options.
func NewEncoder(opts ...EncoderOption) *Encoder {
	return &Encoder{opts: slices.Clone(opts)}
}

// Marshal renders v as a TOON document.
func (e *Encoder) Marshal(v any) ([]byte, error) {
	cfg := defaultEncoderOptions()
	for _, opt := range e.opts {
		opt(&cfg)
	}
	normalized, err := normalize(v, cfg)
	if err != nil {
		return nil, err
	}
	state := &encodeState{cfg: cfg}
	if err := state.encodeRoot(normalized); err != nil {
		return nil, err
	}
	return []byte(strings.Join(state.lines, "\n")), nil
}

// MarshalString is equivalent to Marshal but returns a string.
func (e *Encoder) MarshalString(v any) (string, error) {
	data, err := e.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Marshal returns the TOON encoding of v.
func Marshal(v any, opts ...EncoderOption) ([]byte, error) {
	return NewEncoder(opts...).Marshal(v)
}

// MarshalString encodes v as a TOON document string.
func MarshalString(v any, opts ...EncoderOption) (string, error) {
	return NewEncoder(opts...).MarshalString(v)
}

// fieldNode is one field entry of a header's field list. A node without
// children is a leaf field; a node with children is a nested field group
// declaring a nested-uniform column.
type fieldNode struct {
	name     string
	children []fieldNode
}

type encodeState struct {
	cfg   encoderOptions
	lines []string
}

func (s *encodeState) emit(line string) {
	s.lines = append(s.lines, line)
}

func (s *encodeState) indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	return strings.Repeat(" ", depth*s.cfg.indentSize)
}

func (s *encodeState) delim() string {
	return string(s.cfg.delimiter.rune())
}

func (s *encodeState) formatPrimitive(value normalizedValue) (string, error) {
	switch v := value.(type) {
	case nil:
		return "null", nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case string:
		return formatpkg.FormatString(v, formatpkg.Context{Delimiter: s.cfg.delimiter.rune()})
	case numberValue:
		return v.literal, nil
	default:
		return "", fmt.Errorf("toon: unsupported primitive %T", value)
	}
}

func (s *encodeState) encodeRoot(value normalizedValue) error {
	switch val := value.(type) {
	case nil, bool, string, numberValue:
		token, err := s.formatPrimitive(val)
		if err != nil {
			return err
		}
		// Unquoted, a decoder would remove a leading U+FEFF as a byte-order mark.
		if strings.HasPrefix(token, "\uFEFF") {
			if token, err = formatpkg.QuoteString(val.(string)); err != nil {
				return err
			}
		}
		s.emit(token)
		return nil
	case Object:
		if val.IsEmpty() {
			// An empty object at the root yields an empty document.
			return nil
		}
		if nodes, ok := detectKeyedTabular(val); ok {
			return s.encodeKeyedTabular("", val, nodes, 0)
		}
		return s.encodeObjectBody(val, 0)
	case []normalizedValue:
		if len(val) == 0 {
			s.emit("[]")
			return nil
		}
		return s.encodeArray("", val, 0)
	default:
		return fmt.Errorf("toon: unsupported root value %T", value)
	}
}

func (s *encodeState) encodeObjectBody(obj Object, depth int) error {
	for _, field := range obj.Fields {
		if err := s.encodeObjectField(field, depth); err != nil {
			return err
		}
	}
	return nil
}

// encodeObjectField renders one object field, whose opening line stands at
// depth and whose nested content, if any, stands at depth+1.
func (s *encodeState) encodeObjectField(field Field, depth int) error {
	keyLit, err := formatpkg.EncodeKey(field.Key)
	if err != nil {
		return err
	}
	indent := s.indent(depth)

	switch val := field.Value.(type) {
	case nil, bool, string, numberValue:
		token, err := s.formatPrimitive(val)
		if err != nil {
			return err
		}
		s.emit(indent + keyLit + ": " + token)
		return nil
	case Object:
		if nodes, ok := detectKeyedTabular(val); ok {
			return s.encodeKeyedTabular(keyLit, val, nodes, depth)
		}
		s.emit(indent + keyLit + ":")
		if val.IsEmpty() {
			return nil
		}
		return s.encodeObjectBody(val, depth+1)
	case []normalizedValue:
		if len(val) == 0 {
			s.emit(indent + keyLit + ": []")
			return nil
		}
		return s.encodeArray(keyLit, val, depth)
	default:
		return fmt.Errorf("toon: unsupported object field %s of type %T", field.Key, val)
	}
}

// encodeArray renders a non-empty array whose header stands at depth. keyLit is
// empty for a root array.
func (s *encodeState) encodeArray(keyLit string, values []normalizedValue, depth int) error {
	indent := s.indent(depth)

	if allPrimitive(values) {
		header, err := s.renderHeader(keyLit, len(values), false, nil)
		if err != nil {
			return err
		}
		cells := make([]string, 0, len(values))
		for _, v := range values {
			token, err := s.formatPrimitive(v)
			if err != nil {
				return err
			}
			cells = append(cells, token)
		}
		s.emit(indent + header + " " + strings.Join(cells, s.delim()))
		return nil
	}

	if nodes, ok := detectTabular(values); ok {
		header, err := s.renderHeader(keyLit, len(values), false, nodes)
		if err != nil {
			return err
		}
		s.emit(indent + header)
		return s.emitRows(values, nodes, depth+1)
	}

	header, err := s.renderHeader(keyLit, len(values), false, nil)
	if err != nil {
		return err
	}
	s.emit(indent + header)
	for _, item := range values {
		if err := s.encodeListItem(item, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *encodeState) emitRows(values []normalizedValue, nodes []fieldNode, rowDepth int) error {
	indent := s.indent(rowDepth)
	for _, value := range values {
		obj, ok := value.(Object)
		if !ok {
			return fmt.Errorf("toon: tabular row is %T, not an object", value)
		}
		cells, err := s.rowCells(obj, nodes)
		if err != nil {
			return err
		}
		s.emit(indent + strings.Join(cells, s.delim()))
	}
	return nil
}

// encodeKeyedTabular renders an object in keyed tabular form. keyLit is empty
// when the object is the document root.
func (s *encodeState) encodeKeyedTabular(keyLit string, obj Object, nodes []fieldNode, depth int) error {
	header, err := s.renderHeader(keyLit, obj.Len(), true, nodes)
	if err != nil {
		return err
	}
	s.emit(s.indent(depth) + header)
	rowIndent := s.indent(depth + 1)
	for _, field := range obj.Fields {
		entryKey, err := formatpkg.EncodeKey(field.Key)
		if err != nil {
			return err
		}
		value, ok := field.Value.(Object)
		if !ok {
			return fmt.Errorf("toon: keyed tabular entry %s is %T, not an object", field.Key, field.Value)
		}
		cells, err := s.rowCells(value, nodes)
		if err != nil {
			return err
		}
		s.emit(rowIndent + entryKey + ": " + strings.Join(cells, s.delim()))
	}
	return nil
}

func (s *encodeState) encodeListItem(item normalizedValue, depth int) error {
	indent := s.indent(depth)

	switch val := item.(type) {
	case nil, bool, string, numberValue:
		token, err := s.formatPrimitive(val)
		if err != nil {
			return err
		}
		s.emit(indent + "- " + token)
		return nil

	case []normalizedValue:
		if len(val) == 0 {
			// The key: [] form does not apply to list items.
			s.emit(indent + "- [0" + s.cfg.delimiter.symbol() + "]:")
			return nil
		}
		header, err := s.renderHeader("", len(val), false, nil)
		if err != nil {
			return err
		}
		if allPrimitive(val) {
			cells := make([]string, 0, len(val))
			for _, v := range val {
				token, err := s.formatPrimitive(v)
				if err != nil {
					return err
				}
				cells = append(cells, token)
			}
			s.emit(indent + "- " + header + " " + strings.Join(cells, s.delim()))
			return nil
		}
		// A keyless fields-bearing header is valid only at the document root,
		// so nested arrays of objects use list form here.
		s.emit(indent + "- " + header)
		for _, nested := range val {
			if err := s.encodeListItem(nested, depth+1); err != nil {
				return err
			}
		}
		return nil

	case Object:
		if val.IsEmpty() {
			s.emit(indent + "-")
			return nil
		}
		// The first field is carried on the hyphen line and stands at depth+1
		// for all scope purposes, so it is rendered at depth+1 and its opening
		// line is then rewritten to carry the marker.
		start := len(s.lines)
		if err := s.encodeObjectField(val.Fields[0], depth+1); err != nil {
			return err
		}
		nested := s.indent(depth + 1)
		s.lines[start] = indent + "- " + strings.TrimPrefix(s.lines[start], nested)
		for _, field := range val.Fields[1:] {
			if err := s.encodeObjectField(field, depth+1); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("toon: unsupported list item %T", item)
	}
}

// rowCells walks the field list depth-first and renders one cell per leaf
// field.
func (s *encodeState) rowCells(obj Object, nodes []fieldNode) ([]string, error) {
	cells := make([]string, 0, len(nodes))
	for _, node := range nodes {
		value, ok := obj.get(node.name)
		if !ok {
			return nil, fmt.Errorf("toon: row is missing field %q", node.name)
		}
		if node.children == nil {
			token, err := s.formatPrimitive(value)
			if err != nil {
				return nil, err
			}
			cells = append(cells, token)
			continue
		}
		sub, ok := value.(Object)
		if !ok {
			return nil, fmt.Errorf("toon: nested column %q holds %T, not an object", node.name, value)
		}
		nestedCells, err := s.rowCells(sub, node.children)
		if err != nil {
			return nil, err
		}
		cells = append(cells, nestedCells...)
	}
	return cells, nil
}

// renderHeader builds an array, tabular, or keyed header.
func (s *encodeState) renderHeader(keyLit string, length int, keyed bool, nodes []fieldNode) (string, error) {
	var b strings.Builder
	b.WriteString(keyLit)
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(length))
	if keyed {
		b.WriteByte(':')
	}
	b.WriteString(s.cfg.delimiter.symbol())
	b.WriteByte(']')
	if len(nodes) > 0 {
		fields, err := s.renderFieldList(nodes)
		if err != nil {
			return "", err
		}
		b.WriteString(fields)
	}
	b.WriteByte(':')
	return b.String(), nil
}

func (s *encodeState) renderFieldList(nodes []fieldNode) (string, error) {
	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		name, err := formatpkg.EncodeKey(node.name)
		if err != nil {
			return "", err
		}
		if node.children != nil {
			nested, err := s.renderFieldList(node.children)
			if err != nil {
				return "", err
			}
			name += nested
		}
		parts = append(parts, name)
	}
	return "{" + strings.Join(parts, s.delim()) + "}", nil
}

// detectTabular reports the shared columns of an array of objects.
func detectTabular(values []normalizedValue) ([]fieldNode, bool) {
	if len(values) == 0 {
		return nil, false
	}
	objects := make([]Object, 0, len(values))
	for _, value := range values {
		obj, ok := value.(Object)
		if !ok {
			return nil, false
		}
		objects = append(objects, obj)
	}
	return detectColumns(objects)
}

// detectKeyedTabular reports the shared columns of an object whose values are
// objects.
func detectKeyedTabular(obj Object) ([]fieldNode, bool) {
	if obj.Len() < 2 {
		return nil, false
	}
	values := make([]Object, 0, obj.Len())
	for _, field := range obj.Fields {
		value, ok := field.Value.(Object)
		if !ok {
			return nil, false
		}
		values = append(values, value)
	}
	return detectColumns(values)
}

// detectColumns reports the shared field structure of a sequence of objects, or
// false when any column is neither uniform-primitive nor nested-uniform.
func detectColumns(objects []Object) ([]fieldNode, bool) {
	if len(objects) == 0 {
		return nil, false
	}
	first := objects[0]
	if first.IsEmpty() {
		return nil, false
	}
	for _, obj := range objects {
		if obj.Len() != first.Len() {
			return nil, false
		}
		for _, field := range first.Fields {
			if _, ok := obj.get(field.Key); !ok {
				return nil, false
			}
		}
	}

	nodes := make([]fieldNode, 0, first.Len())
	for _, field := range first.Fields {
		column := make([]normalizedValue, 0, len(objects))
		for _, obj := range objects {
			value, _ := obj.get(field.Key)
			column = append(column, value)
		}
		if allPrimitive(column) {
			nodes = append(nodes, fieldNode{name: field.Key})
			continue
		}
		nested := make([]Object, 0, len(column))
		for _, value := range column {
			sub, ok := value.(Object)
			if !ok || sub.IsEmpty() {
				return nil, false
			}
			nested = append(nested, sub)
		}
		children, ok := detectColumns(nested)
		if !ok {
			return nil, false
		}
		nodes = append(nodes, fieldNode{name: field.Key, children: children})
	}
	return nodes, true
}

func isPrimitive(value normalizedValue) bool {
	switch value.(type) {
	case nil, bool, string, numberValue:
		return true
	default:
		return false
	}
}

func allPrimitive(values []normalizedValue) bool {
	for _, v := range values {
		if !isPrimitive(v) {
			return false
		}
	}
	return true
}
