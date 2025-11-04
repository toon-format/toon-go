package codec

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

// Unmarshal decodes the TOON document in data into v, which must be a non-nil
// pointer. Struct fields use `toon` struct tags for naming and omitempty
// semantics, mirroring Marshal behaviour.
func Unmarshal(data []byte, v any, opts ...DecoderOption) error {
	if v == nil {
		return errors.New("toon: Unmarshal nil target")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("toon: Unmarshal target must be a non-nil pointer")
	}
	decoded, err := Decode(data, opts...)
	if err != nil {
		return err
	}
	return assignValue(rv.Elem(), decoded)
}

// UnmarshalString decodes the TOON document in s into v.
func UnmarshalString(s string, v any, opts ...DecoderOption) error {
	return Unmarshal([]byte(s), v, opts...)
}

func assignValue(dst reflect.Value, src any) error {
	if !dst.CanSet() {
		return errors.New("toon: cannot set destination value")
	}

	switch dst.Kind() {
	case reflect.Interface:
		if src == nil {
			dst.SetZero()
			return nil
		}
		dst.Set(reflect.ValueOf(src))
		return nil
	case reflect.Pointer:
		if src == nil {
			dst.SetZero()
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignValue(dst.Elem(), src)
	case reflect.Struct:
		obj, ok := src.(map[string]any)
		if !ok {
			return fmt.Errorf("toon: expected object for struct, got %T", src)
		}
		meta := cachedStructMeta(dst.Type())
		for _, fieldMeta := range meta.fields {
			value, exists := obj[fieldMeta.name]
			if !exists {
				continue
			}
			fieldValue := dst.FieldByIndex(fieldMeta.index)
			if err := assignValue(fieldValue, value); err != nil {
				return fmt.Errorf("%s: %w", fieldMeta.name, err)
			}
		}
		return nil
	case reflect.Map:
		if dst.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("toon: map key type must be string, got %s", dst.Type().Key())
		}
		obj, ok := src.(map[string]any)
		if !ok {
			return fmt.Errorf("toon: expected object for map, got %T", src)
		}
		if dst.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}
		for key, value := range obj {
			elem := reflect.New(dst.Type().Elem()).Elem()
			if err := assignValue(elem, value); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			dst.SetMapIndex(reflect.ValueOf(key), elem)
		}
		return nil
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			if src == nil {
				dst.SetZero()
				return nil
			}
			if str, ok := src.(string); ok {
				dst.SetBytes([]byte(str))
				return nil
			}
		}
		arr, ok := src.([]any)
		if !ok {
			return fmt.Errorf("toon: expected array for slice, got %T", src)
		}
		slice := reflect.MakeSlice(dst.Type(), len(arr), len(arr))
		for i, item := range arr {
			if err := assignValue(slice.Index(i), item); err != nil {
				return fmt.Errorf("index %d: %w", i, err)
			}
		}
		dst.Set(slice)
		return nil
	case reflect.Array:
		arr, ok := src.([]any)
		if !ok {
			return fmt.Errorf("toon: expected array for fixed array, got %T", src)
		}
		if len(arr) != dst.Len() {
			return fmt.Errorf("toon: array length mismatch: expected %d, got %d", dst.Len(), len(arr))
		}
		for i := 0; i < dst.Len(); i++ {
			if err := assignValue(dst.Index(i), arr[i]); err != nil {
				return fmt.Errorf("index %d: %w", i, err)
			}
		}
		return nil
	case reflect.String:
		switch val := src.(type) {
		case string:
			dst.SetString(val)
			return nil
		default:
			return fmt.Errorf("toon: cannot assign %T to string", src)
		}
	case reflect.Bool:
		if b, ok := src.(bool); ok {
			dst.SetBool(b)
			return nil
		}
		return fmt.Errorf("toon: cannot assign %T to bool", src)
	case reflect.Float32, reflect.Float64:
		if num, ok := toFloat64(src); ok {
			dst.SetFloat(num)
			return nil
		}
		return fmt.Errorf("toon: cannot assign %T to float", src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := toFloat64(src); ok {
			if math.Trunc(num) != num {
				return fmt.Errorf("toon: cannot assign non-integer %v to %s", num, dst.Type())
			}
			intVal := int64(num)
			if dst.OverflowInt(intVal) {
				return fmt.Errorf("toon: integer %v overflows %s", num, dst.Type())
			}
			dst.SetInt(intVal)
			return nil
		}
		return fmt.Errorf("toon: cannot assign %T to int", src)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if num, ok := toFloat64(src); ok {
			if math.Trunc(num) != num {
				return fmt.Errorf("toon: cannot assign non-integer %v to %s", num, dst.Type())
			}
			if num < 0 {
				return fmt.Errorf("toon: cannot assign negative %v to %s", num, dst.Type())
			}
			uintVal := uint64(num)
			if dst.OverflowUint(uintVal) {
				return fmt.Errorf("toon: integer %v overflows %s", num, dst.Type())
			}
			dst.SetUint(uintVal)
			return nil
		}
		return fmt.Errorf("toon: cannot assign %T to uint", src)
	default:
		return fmt.Errorf("toon: unsupported destination kind %s", dst.Kind())
	}
}

func toFloat64(v any) (float64, bool) {
	switch num := v.(type) {
	case float64:
		return num, true
	case float32:
		return float64(num), true
	case int:
		return float64(num), true
	case int8:
		return float64(num), true
	case int16:
		return float64(num), true
	case int32:
		return float64(num), true
	case int64:
		return float64(num), true
	case uint:
		return float64(num), true
	case uint8:
		return float64(num), true
	case uint16:
		return float64(num), true
	case uint32:
		return float64(num), true
	case uint64:
		return float64(num), true
	default:
		return 0, false
	}
}
