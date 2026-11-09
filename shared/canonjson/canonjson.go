// Package canonjson encodes values as canonical JSON, the JSON Canonicalization
// Scheme of RFC 8785 restricted to the values a hash or a signature needs:
// null, booleans, strings, integers, arrays and objects with string keys.
// The same value always gives the same bytes, so the output can be hashed.
//
// The rules of RFC 8785 that apply here:
//
//   - no whitespace;
//   - object members sorted by their keys as arrays of UTF-16 code units;
//   - strings in UTF-8, escaping only '"', '\\' and the control characters
//     below U+0020 (\b \t \n \f \r in short form, the others as \u00xx in
//     lower-case hex); nothing else is escaped (not '/', '<', U+2028);
//   - integers in plain decimal, without sign for zero.
//
// As in encoding/json, a nil slice or map is null and an empty one is [] or
// {}: they are different values with different bytes, so a caller that means
// them to hash alike normalizes first (audit.Canonical turns nil Data into {}).
//
// It walks the value directly instead of re-reading the output of
// encoding/json: the value's Go type decides what is accepted, so a float that
// happens to be integral (1.0) is still refused, and no struct tag or
// MarshalJSON method can change the bytes. What it refuses, with
// ErrUnsupported:
//
//   - floating-point and complex numbers: RFC 8785 serializes them through
//     the ECMAScript number algorithm, which this package does not implement;
//   - integers outside ±MaxSafeInteger, which a JavaScript consumer could not
//     represent exactly;
//   - strings and keys that are not valid UTF-8, lone surrogates (U+D800 to
//     U+DFFF encoded in UTF-8) included;
//   - map keys whose kind is not string; structs, []byte, json.Number,
//     channels and functions (encode them explicitly first);
//   - nesting deeper than MaxDepth, which also stops cyclic values.
//
// Errors never echo the value.
package canonjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// MaxSafeInteger is the largest integer magnitude accepted: 2^53 - 1, the
// largest n for which every integer in [-n, n] is exact in an IEEE 754 double
// (JavaScript's Number.MAX_SAFE_INTEGER).
const MaxSafeInteger = 1<<53 - 1

// MaxDepth is the deepest accepted nesting of arrays, objects and pointers.
const MaxDepth = 32

// ErrUnsupported reports a value that has no canonical form here.
var ErrUnsupported = errors.New("canonjson: unsupported value")

// numberType is refused although its kind is string: it holds a number.
var numberType = reflect.TypeFor[json.Number]()

// Marshal returns the canonical JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	return appendValue(nil, reflect.ValueOf(v), 0)
}

func unsupported(what string) error { return fmt.Errorf("%w: %s", ErrUnsupported, what) }

func appendValue(b []byte, v reflect.Value, depth int) ([]byte, error) {
	if depth > MaxDepth {
		return nil, unsupported("nesting too deep")
	}
	if !v.IsValid() {
		return append(b, "null"...), nil
	}
	switch v.Kind() {
	case reflect.Bool:
		return strconv.AppendBool(b, v.Bool()), nil
	case reflect.String:
		if v.Type() == numberType {
			return nil, unsupported("json.Number")
		}
		return appendString(b, v.String())
	case reflect.Interface, reflect.Pointer:
		return appendIndirect(b, v, depth)
	case reflect.Slice, reflect.Array:
		return appendArray(b, v, depth)
	case reflect.Map:
		return appendObject(b, v, depth)
	}
	return appendInteger(b, v)
}

// appendIndirect writes what an interface or a pointer holds. An interface is
// not a level of nesting: it cannot hold itself without a pointer, slice or
// map in between, and each of those counts.
func appendIndirect(b []byte, v reflect.Value, depth int) ([]byte, error) {
	if v.IsNil() {
		return append(b, "null"...), nil
	}
	if v.Kind() == reflect.Pointer {
		depth++
	}
	return appendValue(b, v.Elem(), depth)
}

// appendInteger writes a signed or unsigned integer and refuses every other
// kind left.
func appendInteger(b []byte, v reflect.Value) ([]byte, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return appendInt(b, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.Uint() > MaxSafeInteger {
			return nil, unsupported("integer out of the safe range")
		}
		return strconv.AppendUint(b, v.Uint(), 10), nil
	}
	return nil, unsupported(v.Kind().String())
}

func appendInt(b []byte, n int64) ([]byte, error) {
	if n > MaxSafeInteger || n < -MaxSafeInteger {
		return nil, unsupported("integer out of the safe range")
	}
	return strconv.AppendInt(b, n, 10), nil
}

func appendArray(b []byte, v reflect.Value, depth int) ([]byte, error) {
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
		return nil, unsupported("byte slice")
	}
	if v.Kind() == reflect.Slice && v.IsNil() {
		return append(b, "null"...), nil
	}
	b = append(b, '[')
	for i := range v.Len() {
		if i > 0 {
			b = append(b, ',')
		}
		var err error
		if b, err = appendValue(b, v.Index(i), depth+1); err != nil {
			return nil, err
		}
	}
	return append(b, ']'), nil
}

func appendObject(b []byte, v reflect.Value, depth int) ([]byte, error) {
	if v.Type().Key().Kind() != reflect.String {
		return nil, unsupported("map key that is not a string")
	}
	if v.IsNil() {
		return append(b, "null"...), nil
	}
	type member struct {
		key   string
		units []uint16
		value reflect.Value
	}
	members := make([]member, 0, v.Len())
	for it := v.MapRange(); it.Next(); {
		k := it.Key().String()
		if !utf8.ValidString(k) {
			return nil, unsupported("key that is not valid UTF-8")
		}
		members = append(members, member{key: k, units: utf16.Encode([]rune(k)), value: it.Value()})
	}
	slices.SortFunc(members, func(x, y member) int { return slices.Compare(x.units, y.units) })
	b = append(b, '{')
	for i, m := range members {
		if i > 0 {
			b = append(b, ',')
		}
		b, _ = appendString(b, m.key) // validated above
		b = append(b, ':')
		var err error
		if b, err = appendValue(b, m.value, depth+1); err != nil {
			return nil, err
		}
	}
	return append(b, '}'), nil
}

const hexDigits = "0123456789abcdef"

// shortEscapes maps the control characters with a two-character escape.
var shortEscapes = [0x20]byte{'\b': 'b', '\t': 't', '\n': 'n', '\f': 'f', '\r': 'r'}

// appendString writes s as a JSON string with the escaping of RFC 8785
// (ECMAScript JSON.stringify). utf8.ValidString refuses encoded surrogates,
// so a lone surrogate never gets through.
func appendString(b []byte, s string) ([]byte, error) {
	if !utf8.ValidString(s) {
		return nil, unsupported("string that is not valid UTF-8")
	}
	b = append(b, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' || c == '\\':
			b = append(b, '\\', c)
		case c >= 0x20:
			b = append(b, c)
		case shortEscapes[c] != 0:
			b = append(b, '\\', shortEscapes[c])
		default:
			b = append(b, '\\', 'u', '0', '0', hexDigits[c>>4], hexDigits[c&0x0f])
		}
	}
	return append(b, '"'), nil
}
