package codec

import (
	"fmt"
	"strconv"
	"strings"
)

// Encoder serializes Go values as TOON documents.
type Encoder struct {
	cfg encoderOptions
}

// NewEncoder constructs an Encoder using the supplied options. Absent options
// default to the TOON Core Profile recommendations (Section 19).
func NewEncoder(opts ...EncoderOption) *Encoder {
	cfg := defaultEncoderOptions()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Encoder{cfg: cfg}
}

// Marshal renders v into a TOON document. Values are first normalized to the
// TOON data model (Section 2), then encoded using the concrete syntax rules
// in Sections 5–12.
func (e *Encoder) Marshal(v any) ([]byte, error) {
	normalized, err := normalize(v, e.cfg)
	if err != nil {
		return nil, err
	}
	state := &encodeState{cfg: e.cfg}
	if err := state.encodeRoot(normalized); err != nil {
		return nil, err
	}
	output := strings.Join(state.lines, "\n")
	return []byte(output), nil
}

// MarshalString is equivalent to Marshal but returns a string.
func (e *Encoder) MarshalString(v any) (string, error) {
	data, err := e.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Marshal encodes v using a temporary encoder.
func Marshal(v any, opts ...EncoderOption) ([]byte, error) {
	return NewEncoder(opts...).Marshal(v)
}

// MarshalString encodes v as a TOON document string.
func MarshalString(v any, opts ...EncoderOption) (string, error) {
	return NewEncoder(opts...).MarshalString(v)
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

func (s *encodeState) encodeRoot(value normalizedValue) error {
	switch val := value.(type) {
	case nil, bool, string, numberValue:
		token, err := formatPrimitive(val, formatContext{
			active:   s.cfg.arrayDelimiter,
			document: s.cfg.documentDelimiter,
			inArray:  false,
		})
		if err != nil {
			return err
		}
		s.emit(token)
	case Object:
		if err := s.encodeObject(val, 0); err != nil {
			return err
		}
	case []normalizedValue:
		if err := s.encodeArray("", val, 0, true); err != nil {
			return err
		}
	default:
		return fmt.Errorf("toon: unsupported root value %T", value)
	}
	return nil
}

func (s *encodeState) encodeObject(obj Object, depth int) error {
	if depth == 0 && obj.IsEmpty() {
		return nil
	}
	indent := s.indent(depth)
	for _, field := range obj.Fields {
		switch val := field.Value.(type) {
		case nil, bool, string, numberValue:
			keyLiteral, err := encodeKey(field.Key)
			if err != nil {
				return err
			}
			token, err := formatPrimitive(val, formatContext{
				active:   s.cfg.arrayDelimiter,
				document: s.cfg.documentDelimiter,
				inArray:  false,
			})
			if err != nil {
				return err
			}
			s.emit(indent + keyLiteral + ": " + token)
		case Object:
			keyLiteral, err := encodeKey(field.Key)
			if err != nil {
				return err
			}
			s.emit(indent + keyLiteral + ":")
			if err := s.encodeObject(val, depth+1); err != nil {
				return err
			}
		case []normalizedValue:
			if err := s.encodeArray(field.Key, val, depth, false); err != nil {
				return err
			}
		default:
			return fmt.Errorf("toon: unsupported object field %s of type %T", field.Key, val)
		}
	}
	return nil
}

func (s *encodeState) encodeArray(key string, values []normalizedValue, depth int, root bool) error {
	indent := s.indent(depth)
	delimiter := s.cfg.arrayDelimiter
	ctx := formatContext{
		active:   delimiter,
		document: s.cfg.documentDelimiter,
		inArray:  true,
	}

	keyLiteral := ""
	var err error
	if key != "" {
		keyLiteral, err = encodeKey(key)
		if err != nil {
			return err
		}
	}

	if isPrimitiveArray(values) {
		header := renderHeader(keyLiteral, len(values), delimiter, s.cfg.includeLengthMarks, nil)
		line := indent + header
		if len(values) > 0 {
			inline := make([]string, 0, len(values))
			for _, v := range values {
				token, err := formatPrimitive(v, ctx)
				if err != nil {
					return err
				}
				inline = append(inline, token)
			}
			line += " " + strings.Join(inline, string(delimiter.rune()))
		}
		s.emit(line)
		return nil
	}

	if fields, ok := detectTabular(values); ok {
		header := renderHeader(keyLiteral, len(values), delimiter, s.cfg.includeLengthMarks, fields)
		s.emit(indent + header)
		for _, row := range values {
			obj := row.(Object)
			rowLine := s.indent(depth + 1)
			rowValues := make([]string, 0, len(fields))
			for _, field := range fields {
				token, err := formatPrimitive(objField(obj, field), ctx)
				if err != nil {
					return err
				}
				rowValues = append(rowValues, token)
			}
			rowLine += strings.Join(rowValues, string(delimiter.rune()))
			s.emit(rowLine)
		}
		return nil
	}

	header := renderHeader(keyLiteral, len(values), delimiter, s.cfg.includeLengthMarks, nil)
	s.emit(indent + header)
	for _, item := range values {
		if root {
			if err := s.encodeListItem(item, depth+1, ctx); err != nil {
				return err
			}
			continue
		}
		if err := s.encodeArrayItem(item, depth+1, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *encodeState) encodeArrayItem(item normalizedValue, depth int, ctx formatContext) error {
	switch v := item.(type) {
	case nil, bool, string, numberValue:
		token, err := formatPrimitive(v, ctx)
		if err != nil {
			return err
		}
		s.emit(s.indent(depth) + "- " + token)
	case Object:
		if err := s.encodeObjectListItem(v, depth, ctx); err != nil {
			return err
		}
	case []normalizedValue:
		return s.encodeArrayForObjectListItem("", v, depth, ctx)
	default:
		return fmt.Errorf("toon: unsupported array item %T", v)
	}
	return nil
}

func (s *encodeState) encodeListItem(item normalizedValue, depth int, ctx formatContext) error {
	switch v := item.(type) {
	case nil, bool, string, numberValue:
		token, err := formatPrimitive(v, ctx)
		if err != nil {
			return err
		}
		s.emit(s.indent(depth) + "- " + token)
	case Object:
		if err := s.encodeObjectListItem(v, depth, ctx); err != nil {
			return err
		}
	case []normalizedValue:
		return s.encodeArrayForObjectListItem("", v, depth, ctx)
	default:
		return fmt.Errorf("toon: unsupported list item %T", v)
	}
	return nil
}

func (s *encodeState) encodeObjectListItem(obj Object, depth int, ctx formatContext) error {
	if obj.IsEmpty() {
		s.emit(s.indent(depth) + "- {}")
		return nil
	}
	first := obj.Fields[0]
	if isPrimitive(first.Value) {
		keyLiteral, err := encodeKey(first.Key)
		if err != nil {
			return err
		}
		token, err := formatPrimitive(first.Value, ctx)
		if err != nil {
			return err
		}
		s.emit(s.indent(depth) + "- " + keyLiteral + ": " + token)
		if len(obj.Fields) > 1 {
			if err := s.encodeObject(Object{Fields: obj.Fields[1:]}, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if arr, ok := first.Value.([]normalizedValue); ok {
		keyLiteral, err := encodeKey(first.Key)
		if err != nil {
			return err
		}
		if err := s.encodeArrayForObjectListItem(keyLiteral, arr, depth, ctx); err != nil {
			return err
		}
		if len(obj.Fields) > 1 {
			if err := s.encodeObject(Object{Fields: obj.Fields[1:]}, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	s.emit(s.indent(depth) + "-")
	return s.encodeObject(obj, depth+1)
}

func (s *encodeState) encodeArrayForObjectListItem(keyLiteral string, values []normalizedValue, depth int, ctx formatContext) error {
	delimiter := ctx.active
	indent := s.indent(depth)

	if fields, ok := detectTabular(values); ok {
		header := renderHeader(keyLiteral, len(values), delimiter, s.cfg.includeLengthMarks, fields)
		s.emit(indent + "- " + header)
		for _, row := range values {
			obj := row.(Object)
			rowLine := s.indent(depth + 1)
			rowValues := make([]string, 0, len(fields))
			for _, field := range fields {
				token, err := formatPrimitive(objField(obj, field), ctx)
				if err != nil {
					return err
				}
				rowValues = append(rowValues, token)
			}
			s.emit(rowLine + strings.Join(rowValues, string(delimiter.rune())))
		}
		return nil
	}

	if isPrimitiveArray(values) {
		header := renderHeader(keyLiteral, len(values), delimiter, s.cfg.includeLengthMarks, nil)
		line := indent + "- " + header
		if len(values) > 0 {
			inline := make([]string, 0, len(values))
			for _, v := range values {
				token, err := formatPrimitive(v, ctx)
				if err != nil {
					return err
				}
				inline = append(inline, token)
			}
			line += " " + strings.Join(inline, string(delimiter.rune()))
		}
		s.emit(line)
		return nil
	}

	header := renderHeader(keyLiteral, len(values), delimiter, s.cfg.includeLengthMarks, nil)
	s.emit(indent + "- " + header)
	for _, item := range values {
		if err := s.encodeListItem(item, depth+1, ctx); err != nil {
			return err
		}
	}
	return nil
}

func detectTabular(values []normalizedValue) ([]string, bool) {
	if len(values) == 0 {
		return nil, false
	}
	first, ok := values[0].(Object)
	if !ok || first.IsEmpty() {
		return nil, false
	}
	fields := make([]string, len(first.Fields))
	fieldSet := make(map[string]struct{}, len(first.Fields))
	for i, field := range first.Fields {
		if !isPrimitive(field.Value) {
			return nil, false
		}
		fields[i] = field.Key
		fieldSet[field.Key] = struct{}{}
	}
	for _, value := range values[1:] {
		obj, ok := value.(Object)
		if !ok {
			return nil, false
		}
		if len(obj.Fields) != len(fields) {
			return nil, false
		}
		seen := make(map[string]struct{}, len(fields))
		for _, field := range obj.Fields {
			if _, ok := fieldSet[field.Key]; !ok || !isPrimitive(field.Value) {
				return nil, false
			}
			seen[field.Key] = struct{}{}
		}
		if len(seen) != len(fields) {
			return nil, false
		}
	}
	return fields, true
}

func objField(obj Object, key string) normalizedValue {
	for _, field := range obj.Fields {
		if field.Key == key {
			return field.Value
		}
	}
	return nil
}

func isPrimitive(value normalizedValue) bool {
	switch value.(type) {
	case nil, bool, string, numberValue:
		return true
	default:
		return false
	}
}

func isPrimitiveArray(values []normalizedValue) bool {
	for _, v := range values {
		if !isPrimitive(v) {
			return false
		}
	}
	return true
}

func renderHeader(keyLiteral string, length int, delimiter Delimiter, includeMarker bool, fields []string) string {
	var b strings.Builder
	if keyLiteral != "" {
		b.WriteString(keyLiteral)
	}
	b.WriteByte('[')
	if includeMarker {
		b.WriteByte('#')
	}
	b.WriteString(strconv.Itoa(length))
	if delimiter != DelimiterComma {
		b.WriteRune(delimiter.rune())
	}
	b.WriteByte(']')
	if len(fields) > 0 {
		b.WriteByte('{')
		for i, field := range fields {
			if i > 0 {
				b.WriteRune(delimiter.rune())
			}
			fieldLiteral, _ := encodeKey(field)
			b.WriteString(fieldLiteral)
		}
		b.WriteByte('}')
	}
	b.WriteByte(':')
	return b.String()
}
