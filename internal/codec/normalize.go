package codec

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strconv"
	"time"
)

// normalize applies the data-model rules from Section 2 and Section 3 to a Go
// value, producing a structure ready for encoding. The returned value is one of:
//   - nil
//   - bool
//   - string
//   - float64
//   - Object
//   - []normalizedValue
//
// Big integers that exceed IEEE 754 precision are converted to decimal strings.
func normalize(v any, cfg encoderOptions) (normalizedValue, error) {
	if v == nil {
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
		return normalizeFloat(float64(val))
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
		if val == nil {
			return nil, nil
		}
		if val.IsInt64() {
			return normalize(val.Int64(), cfg)
		}
		return val.String(), nil
	case big.Int:
		return normalize(&val, cfg)
	case time.Time:
		return cfg.timeFormatter(val), nil
	case fmt.Stringer:
		return val.String(), nil
	case Object:
		return normalizeObjectFields(val.Fields, cfg)
	case Field:
		return normalizeObjectFields([]Field{val}, cfg)
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Pointer:
		if val.IsNil() {
			return nil, nil
		}
		return normalize(val.Elem().Interface(), cfg)
	case reflect.Slice, reflect.Array:
		length := val.Len()
		result := make([]normalizedValue, 0, length)
		for i := 0; i < length; i++ {
			item, err := normalize(val.Index(i).Interface(), cfg)
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}
		return result, nil
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("toon: unsupported map key type %s", val.Type().Key())
		}
		iter := val.MapRange()
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
		return normalizeStructValue(val, cfg)
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
		s := strconv.FormatFloat(f, 'f', -1, 64)
		return numberValue{literal: s}, nil
	}
}

func normalizeNumberString(s string) (normalizedValue, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Preserve as string literal; encoder will handle quoting.
		return s, nil
	}
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, nil
	}
	if f == 0 {
		f = 0
	}
	return numberValue{literal: strconv.FormatFloat(f, 'f', -1, 64)}, nil
}
