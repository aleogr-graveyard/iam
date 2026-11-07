package jsonx

import (
	"reflect"
	"strings"
	"sync"
)

var fieldCache sync.Map // reflect.Type -> map[string]reflect.Type

// candidate is one field that may bind a JSON name.
type candidate struct {
	typ    reflect.Type
	depth  int
	tagged bool
}

// structFields returns the exact JSON names a struct accepts, mapped to the
// type of the field that binds the name under encoding/json's dominance rules.
// Ambiguous names are absent, as encoding/json drops them.
func structFields(t reflect.Type) map[string]reflect.Type {
	if m, ok := fieldCache.Load(t); ok {
		return m.(map[string]reflect.Type)
	}
	cands := map[string][]candidate{}
	collect(t, 0, cands, map[reflect.Type]bool{})
	out := make(map[string]reflect.Type, len(cands))
	for name, cs := range cands {
		if c, ok := dominant(cs); ok {
			out[name] = c.typ
		}
	}
	fieldCache.Store(t, out)
	return out
}

// dominant picks the shallowest candidate; at that depth a lone tagged one
// wins, otherwise a lone candidate wins; anything else is ambiguous.
func dominant(cs []candidate) (candidate, bool) {
	minDepth := cs[0].depth
	for _, c := range cs {
		minDepth = min(minDepth, c.depth)
	}
	var shallow, tagged []candidate
	for _, c := range cs {
		if c.depth != minDepth {
			continue
		}
		shallow = append(shallow, c)
		if c.tagged {
			tagged = append(tagged, c)
		}
	}
	if len(tagged) > 0 {
		shallow = tagged
	}
	if len(shallow) != 1 {
		return candidate{}, false
	}
	return shallow[0], true
}

func collect(t reflect.Type, depth int, out map[string][]candidate, visiting map[reflect.Type]bool) {
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
				collect(et, depth+1, out, visiting)
				continue
			}
		}
		if !f.IsExported() {
			continue
		}
		tagged := name != ""
		if !tagged {
			name = f.Name
		}
		out[name] = append(out[name], candidate{typ: f.Type, depth: depth, tagged: tagged})
	}
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
