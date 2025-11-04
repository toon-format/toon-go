package codec

import (
	"fmt"

	formatpkg "github.com/toon-format/toon-go/internal/format"
)

type formatContext struct {
	active   Delimiter
	document Delimiter
	inArray  bool
}

func (c formatContext) toInternal() formatpkg.Context {
	return formatpkg.Context{
		Active:   c.active.rune(),
		Document: c.document.rune(),
		InArray:  c.inArray,
	}
}

func formatPrimitive(value normalizedValue, ctx formatContext) (string, error) {
	switch v := value.(type) {
	case nil:
		return "null", nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case string:
		return formatpkg.FormatString(v, ctx.toInternal())
	case numberValue:
		return v.literal, nil
	default:
		return "", fmt.Errorf("toon: unsupported primitive %T", value)
	}
}

func encodeKey(key string) (string, error) {
	return formatpkg.EncodeKey(key)
}
