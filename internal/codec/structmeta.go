package codec

import (
	"reflect"
	"strings"
	"sync"
)

type structFieldMeta struct {
	name      string
	omitEmpty bool
	index     []int
}

type structMeta struct {
	fields []structFieldMeta
	lookup map[string]structFieldMeta
}

var structCache sync.Map // map[reflect.Type]structMeta

func cachedStructMeta(t reflect.Type) structMeta {
	if meta, ok := structCache.Load(t); ok {
		return meta.(structMeta)
	}
	meta := buildStructMeta(t)
	structCache.Store(t, meta)
	return meta
}

func buildStructMeta(t reflect.Type) structMeta {
	fields := make([]structFieldMeta, 0, t.NumField())
	lookup := make(map[string]structFieldMeta, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		tag := sf.Tag.Get("toon")
		if tag == "-" {
			continue
		}
		name, opts := parseStructTag(tag)
		if name == "" {
			name = sf.Name
		}
		meta := structFieldMeta{
			name:      name,
			omitEmpty: opts["omitempty"],
			index:     sf.Index,
		}
		fields = append(fields, meta)
		lookup[name] = meta
	}
	return structMeta{fields: fields, lookup: lookup}
}

func parseStructTag(tag string) (string, map[string]bool) {
	options := map[string]bool{}
	if tag == "" {
		return "", options
	}
	parts := strings.Split(tag, ",")
	name := parts[0]
	for _, opt := range parts[1:] {
		if opt == "" {
			continue
		}
		options[opt] = true
	}
	return name, options
}

func fieldValueByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Zero(v.Type().Elem())
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
	return false
}
