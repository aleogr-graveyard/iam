package jsonx_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aleogr/iam/shared/jsonx"
)

type inner struct {
	N int `json:"n"`
}

type doc struct {
	A int    `json:"a"`
	S string `json:"s"`
	I inner  `json:"i"`
}

func TestDecodeStrict(t *testing.T) {
	lim := jsonx.Limits{MaxBytes: 64, MaxDepth: 32}

	tests := []struct {
		name string
		in   string
		lim  jsonx.Limits
		want error
	}{
		{"valid", `{"a":1,"s":"x","i":{"n":2}}`, lim, nil},
		{"valid with whitespace", " {\"a\":1}\n", lim, nil},
		{"duplicate top level", `{"a":1,"a":2}`, lim, jsonx.ErrDuplicateKey},
		{"duplicate nested", `{"i":{"n":1,"n":2}}`, lim, jsonx.ErrDuplicateKey},
		{"duplicate in array element", `{"i":[{"n":1,"n":2}]}`, lim, jsonx.ErrDuplicateKey},
		{"unknown field", `{"zzz":1}`, lim, jsonx.ErrUnknownField},
		{"too large", `{"s":"` + strings.Repeat("x", 64) + `"}`, lim, jsonx.ErrTooLarge},
		{"exactly max bytes", `{"s":"` + strings.Repeat("x", 64-8) + `"}`, lim, nil},
		{"two values", `{"a":1}{"a":2}`, lim, jsonx.ErrTrailingData},
		{"trailing garbage", `{"a":1} x`, lim, jsonx.ErrTrailingData},
		{"syntax", `{"a":`, lim, jsonx.ErrInvalidJSON},
		{"empty", ``, lim, jsonx.ErrInvalidJSON},
		{"type mismatch", `{"a":"str"}`, lim, jsonx.ErrInvalidJSON},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got doc
			err := jsonx.DecodeStrict(strings.NewReader(tc.in), &got, tc.lim)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if err != nil && got != (doc{}) {
				t.Fatalf("value not zero after error: %+v", got)
			}
		})
	}
}

func TestDecodeStrictValue(t *testing.T) {
	var got doc
	if err := jsonx.DecodeStrict(strings.NewReader(`{"a":1,"s":"x","i":{"n":2}}`), &got, jsonx.DefaultLimits); err != nil {
		t.Fatal(err)
	}
	if got != (doc{A: 1, S: "x", I: inner{N: 2}}) {
		t.Fatalf("unexpected value %+v", got)
	}
}

func TestDecodeStrictLeavesValueZeroOnLateError(t *testing.T) {
	// The first field decodes fine; the unknown one fails afterwards.
	var got doc
	err := jsonx.DecodeStrict(strings.NewReader(`{"a":1,"zzz":2}`), &got, jsonx.DefaultLimits)
	if !errors.Is(err, jsonx.ErrUnknownField) {
		t.Fatalf("want ErrUnknownField, got %v", err)
	}
	if got != (doc{}) {
		t.Fatalf("value not zero: %+v", got)
	}
}

func TestDecodeStrictRejectsBadTarget(t *testing.T) {
	var d doc
	for _, v := range []any{nil, d, (*doc)(nil)} {
		if err := jsonx.DecodeStrict(strings.NewReader(`{}`), v, jsonx.DefaultLimits); err == nil {
			t.Fatalf("want error for target %T", v)
		}
	}
}

func TestErrorsDoNotEchoInput(t *testing.T) {
	const secret = "s3cr3t-token-value"
	inputs := []string{
		`{"` + secret + `":1}`,
		`{"a":1,"a":2,"` + secret + `":1}`,
		`{"a":"` + secret + `"}`,
		`{"a":1} ` + secret,
		`{"s":` + secret + `}`,
	}
	for _, in := range inputs {
		var d doc
		err := jsonx.DecodeStrict(strings.NewReader(in), &d, jsonx.DefaultLimits)
		if err == nil {
			t.Fatalf("want error for %q", in)
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error echoes input: %v", err)
		}
	}
}

func TestDecodeStrictDepthAndSiblings(t *testing.T) {
	nest := func(n int) string { return strings.Repeat("[", n) + strings.Repeat("]", n) }
	lim := jsonx.Limits{MaxBytes: 1 << 10, MaxDepth: 32}
	tests := []struct {
		name string
		in   string
		want error
	}{
		{"depth 32", nest(32), nil},
		{"depth 33", nest(33), jsonx.ErrTooDeep},
		{"depth 33 in object", strings.Repeat(`{"k":`, 33) + "1" + strings.Repeat("}", 33), jsonx.ErrTooDeep},
		{"same key in sibling objects", `[{"a":1},{"a":2}]`, nil},
		{"same key in parent and child", `{"a":{"a":1}}`, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v any
			if err := jsonx.DecodeStrict(strings.NewReader(tc.in), &v, lim); !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}
