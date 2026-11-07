package jsonx

import (
	"reflect"
	"strings"
	"sync"
)

var fieldCache sync.Map // reflect.Type -> map[string]reflect.Type

// structFields returns the exact JSON names a struct accepts, mapped to the
// field type (nil when several promoted fields share a name).
func structFields(t reflect.Type) map[string]reflect.Type {
	if m, ok := fieldCache.Load(t); ok {
		return m.(map[string]reflect.Type)
	}
	out := map[string]reflect.Type{}
	collect(t, out, map[reflect.Type]bool{})
	fieldCache.Store(t, out)
	return out
}

func collect(t reflect.Type, out map[string]reflect.Type, visiting map[reflect.Type]bool) {
	if visiting[t] {
		return
	}
	visiting[t] = true
	defer delete(visiting, t)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if f.Anonymous && name == "" {
			if et := derefType(f.Type); et.Kind() == reflect.Struct {
				collect(et, out, visiting)
				continue
			}
		}
		if !f.IsExported() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		if _, dup := out[name]; dup {
			out[name] = nil
		} else {
			out[name] = f.Type
		}
	}
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
