// Package jsonx decodes untrusted JSON strictly and within bounds: it rejects
// duplicate object keys, unknown fields, trailing data, oversized bodies and
// excessive nesting. Errors never echo input content.
package jsonx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Limits bounds the size and nesting depth of a decoded document.
type Limits struct {
	// MaxBytes is the largest accepted body, in bytes.
	MaxBytes int64
	// MaxDepth is the deepest accepted nesting of arrays and objects.
	MaxDepth int
}

// DefaultLimits is a conservative default for API bodies.
var DefaultLimits = Limits{MaxBytes: 1 << 20, MaxDepth: 32}

// Sentinel errors. They carry no input content.
var (
	ErrTooLarge     = errors.New("jsonx: document too large")
	ErrTooDeep      = errors.New("jsonx: document nested too deeply")
	ErrDuplicateKey = errors.New("jsonx: duplicate object key")
	ErrUnknownField = errors.New("jsonx: unknown field")
	ErrTrailingData = errors.New("jsonx: trailing data after document")
	ErrInvalidJSON  = errors.New("jsonx: invalid JSON")
)

// DecodeStrict reads one JSON document from r into v, which must be a non-nil
// pointer. v is overwritten only on success; on error it is left untouched.
func DecodeStrict(r io.Reader, v any, lim Limits) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%w: target must be a non-nil pointer", ErrInvalidJSON)
	}
	if lim.MaxBytes < 0 || lim.MaxDepth < 1 {
		return fmt.Errorf("%w: invalid limits", ErrInvalidJSON)
	}

	data, err := io.ReadAll(io.LimitReader(r, lim.MaxBytes+1))
	if err != nil {
		return fmt.Errorf("%w: read failed", ErrInvalidJSON)
	}
	if int64(len(data)) > lim.MaxBytes {
		return ErrTooLarge
	}
	if err := scan(data, lim.MaxDepth); err != nil {
		return err
	}

	tmp := reflect.New(rv.Elem().Type())
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(tmp.Interface()); err != nil {
		return classify(err)
	}
	rv.Elem().Set(tmp.Elem())
	return nil
}

// classify maps a decoder error to a sentinel without echoing its message.
func classify(err error) error {
	if strings.HasPrefix(err.Error(), "json: unknown field ") {
		return ErrUnknownField
	}
	return ErrInvalidJSON
}

// frame is one open container; keys is non-nil only for objects.
type frame struct {
	keys      map[string]struct{}
	expectKey bool
}

// scan validates structure, depth, duplicate keys and trailing data.
func scan(data []byte, maxDepth int) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	var stack []frame
	for {
		tok, err := dec.Token()
		if err != nil {
			return scanError(err)
		}
		stack, err = step(stack, tok, maxDepth)
		if err != nil {
			return err
		}
		if len(stack) == 0 {
			break
		}
	}
	if _, err := dec.Token(); err != io.EOF {
		return ErrTrailingData
	}
	return nil
}

func scanError(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return fmt.Errorf("%w: unexpected end", ErrInvalidJSON)
	}
	return ErrInvalidJSON
}

// step applies one token to the container stack.
func step(stack []frame, tok json.Token, maxDepth int) ([]frame, error) {
	if d, ok := tok.(json.Delim); ok {
		switch d {
		case '{', '[':
			if len(stack) >= maxDepth {
				return stack, ErrTooDeep
			}
			f := frame{}
			if d == '{' {
				f.keys, f.expectKey = map[string]struct{}{}, true
			}
			markValue(stack)
			return append(stack, f), nil
		default: // '}' or ']'
			return stack[:len(stack)-1], nil
		}
	}
	if len(stack) > 0 && stack[len(stack)-1].expectKey {
		top := &stack[len(stack)-1]
		key, _ := tok.(string)
		if _, dup := top.keys[key]; dup {
			return stack, ErrDuplicateKey
		}
		top.keys[key] = struct{}{}
		top.expectKey = false
		return stack, nil
	}
	markValue(stack)
	return stack, nil
}

// markValue records that a value was consumed in the enclosing container, so an
// object expects a key next.
func markValue(stack []frame) {
	if n := len(stack); n > 0 && stack[n-1].keys != nil {
		stack[n-1].expectKey = true
	}
}
