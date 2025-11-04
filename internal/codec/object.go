package codec

// Field represents a single key/value pair in an ordered object.
type Field struct {
	Key   string
	Value any
}

// Object preserves the encounter order of its fields, ensuring deterministic
// emission by the encoder.
type Object struct {
	Fields []Field
}

// NewObject constructs an ordered Object from the provided key/value pairs.
func NewObject(fields ...Field) Object {
	return Object{Fields: append([]Field(nil), fields...)}
}

// Len reports the number of fields.
func (o Object) Len() int {
	return len(o.Fields)
}

// IsEmpty reports whether the object has no fields.
func (o Object) IsEmpty() bool {
	return len(o.Fields) == 0
}
