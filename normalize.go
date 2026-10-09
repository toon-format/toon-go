package toon

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strconv"
	"time"

	formatpkg "github.com/toon-format/toon-go/internal/format"
)

// normalize converts a Go value to the TOON data model, ready for encoding. The
// returned value is one of:
//   - nil
//   - bool
//   - string
//   - numberValue
//   - Object
//   - []normalizedValue
//
// Big integers that exceed IEEE 754 precision are converted to decimal strings.
func normalize(v any, cfg encoderOptions) (normalizedValue, error) {
	// A nil pointer becomes null before the cases below can call its methods,
	// which may dereference it.
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() == reflect.Pointer && rv.IsNil() {
		return nil, nil
	}

	switch val := v.(type) {
	case string:
		return val, nil
	case bool:
		return val, nil
	case json.Number:
		return normalizeNumberString(val.String())
	case float32:
		// Formatting at float64 precision would print float32(0.1) as 0.10000000149011612.
		return normalizeNumberString(strconv.FormatFloat(float64(val), 'g', -1, 32))
	case float64:
		return normalizeFloat(val)
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(val).Int()
		if i > maxSafeInteger || i < -maxSafeInteger {
			return strconv.FormatInt(i, 10), nil
		}
		return numberValue{literal: strconv.FormatInt(i, 10)}, nil
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(val).Uint()
		if u > maxSafeInteger {
			return strconv.FormatUint(u, 10), nil
		}
		return numberValue{literal: strconv.FormatUint(u, 10)}, nil
	case *big.Int:
		if val.IsInt64() {
			return normalize(val.Int64(), cfg)
		}
		return val.String(), nil
	case big.Int:
		return normalize(&val, cfg)
	case time.Time:
		return cfg.timeFormatter(val), nil
	case *time.Time:
		return normalize(*val, cfg)
	case *json.Number:
		return normalize(*val, cfg)
	case fmt.Stringer:
		return val.String(), nil
	case Object:
		return normalizeObjectFields(val.Fields, cfg)
	case Field:
		return normalizeObjectFields([]Field{val}, cfg)
	}

	switch rv.Kind() {
	case reflect.Pointer:
		return normalize(rv.Elem().Interface(), cfg)
	case reflect.Slice, reflect.Array:
		length := rv.Len()
		result := make([]normalizedValue, 0, length)
		for i := range length {
			item, err := normalize(rv.Index(i).Interface(), cfg)
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}
		return result, nil
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("toon: unsupported map key type %s", rv.Type().Key())
		}
		iter := rv.MapRange()
		var fields []Field
		for iter.Next() {
			fieldValue, err := normalize(iter.Value().Interface(), cfg)
			if err != nil {
				return nil, err
			}
			fields = append(fields, Field{
				Key:   iter.Key().String(),
				Value: fieldValue,
			})
		}
		slices.SortFunc(fields, func(a, b Field) int {
			if a.Key < b.Key {
				return -1
			}
			if a.Key > b.Key {
				return 1
			}
			return 0
		})
		return Object{Fields: fields}, nil
	case reflect.Struct:
		return normalizeStructValue(rv, cfg)
	}

	return nil, fmt.Errorf("toon: unsupported value of type %T", v)
}

func normalizeStructValue(val reflect.Value, cfg encoderOptions) (Object, error) {
	meta := cachedStructMeta(val.Type())
	fields := make([]Field, 0, len(meta.fields))
	for _, field := range meta.fields {
		childValue := fieldValueByIndex(val, field.index)
		if field.omitEmpty && isEmptyValue(childValue) {
			continue
		}
		child, err := normalize(childValue.Interface(), cfg)
		if err != nil {
			return Object{}, fmt.Errorf("toon: %s: %w", field.name, err)
		}
		fields = append(fields, Field{
			Key:   field.name,
			Value: child,
		})
	}
	return Object{Fields: fields}, nil
}

func normalizeObjectFields(fields []Field, cfg encoderOptions) (Object, error) {
	normalized := make([]Field, 0, len(fields))
	for _, field := range fields {
		child, err := normalize(field.Value, cfg)
		if err != nil {
			return Object{}, fmt.Errorf("toon: %s: %w", field.Key, err)
		}
		normalized = append(normalized, Field{
			Key:   field.Key,
			Value: child,
		})
	}
	return Object{Fields: normalized}, nil
}

func normalizeFloat(f float64) (normalizedValue, error) {
	switch {
	case math.IsNaN(f):
		return nil, nil
	case math.IsInf(f, 1), math.IsInf(f, -1):
		return nil, nil
	default:
		if f == math.Copysign(0, -1) {
			f = 0
		}
		return numberValue{literal: formatpkg.FormatNumber(f)}, nil
	}
}

func normalizeNumberString(s string) (normalizedValue, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// A token that float64 cannot parse stays a string.
		return s, nil
	}
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, nil
	}
	return numberValue{literal: formatpkg.FormatNumber(f)}, nil
}
