package canonjson_test

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/aleogr/iam/shared/canonjson"
)

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := canonjson.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%#v): %v", v, err)
	}
	return string(b)
}

// TestCanonJSON_RFC8785Vectors uses the RFC 8785 examples that hold no
// floating-point numbers: property sorting by UTF-16 code units (section
// 3.2.3, including a supplementary-plane key that sorts below U+FB33 in
// UTF-16 but above it in code points), string escaping (3.2.2.2) and the
// literals of the main example (3.2.4).
func TestCanonJSON_RFC8785Vectors(t *testing.T) {
	sorting := map[string]any{
		"\u20ac":     "Euro Sign",
		"\r":         "Carriage Return",
		"\ufb33":     "Hebrew Letter Dalet With Dagesh",
		"1":          "One",
		"\U0001F600": "Emoji: Grinning Face",
		"\u0080":     "Control",
		"\u00f6":     "Latin Small Letter O With Diaeresis",
	}
	wantSorting := "{\"\\r\":\"Carriage Return\",\"1\":\"One\",\"\u0080\":\"Control\",\"ö\":\"Latin Small Letter O With Diaeresis\"," +
		"\"€\":\"Euro Sign\",\"\U0001F600\":\"Emoji: Grinning Face\",\"\ufb33\":\"Hebrew Letter Dalet With Dagesh\"}"
	if got := mustMarshal(t, sorting); got != wantSorting {
		t.Errorf("sorting:\n got  %s\n want %s", got, wantSorting)
	}

	main := map[string]any{
		"literals": []any{nil, true, false},
		"string":   "\u20ac$\u000F\u000aA'\u0042\u0022\u005c\\\"/",
	}
	wantMain := `{"literals":[null,true,false],"string":"€$\u000f\nA'B\"\\\\\"/"}`
	if got := mustMarshal(t, main); got != wantMain {
		t.Errorf("main example:\n got  %s\n want %s", got, wantMain)
	}
}

func TestCanonJSON_Escaping(t *testing.T) {
	for in, want := range map[string]string{
		"\b\t\n\f\r":      `"\b\t\n\f\r"`,
		"\x00\x01\x1f":    `"\u0000\u0001\u001f"`,
		"\x7f":            "\"\x7f\"",            // DEL is not escaped
		"<>&\u2028\u2029": "\"<>&\u2028\u2029\"", // no HTML or line-separator escaping
		`"\/`:             `"\"\\/"`,
		"":                `""`,
	} {
		if got := mustMarshal(t, in); got != want {
			t.Errorf("Marshal(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestCanonJSON_Values(t *testing.T) {
	type label string
	var nilMap map[string]string
	for _, tc := range []struct {
		in   any
		want string
	}{
		{nil, "null"},
		{true, "true"},
		{0, "0"},
		{-0, "0"},
		{int8(-7), "-7"},
		{uint16(65535), "65535"},
		{int64(canonjson.MaxSafeInteger), "9007199254740991"},
		{int64(-canonjson.MaxSafeInteger), "-9007199254740991"},
		{label("x"), `"x"`},
		{[]string{"b", "a"}, `["b","a"]`},
		{[0]int{}, "[]"},
		{[]any{}, "[]"},
		{map[string]string{}, "{}"},
		{nilMap, "null"},
		{map[label]int{"b": 1, "a": 2}, `{"a":2,"b":1}`},
		{map[string]any{"z": map[string]any{"y": []any{1, "x"}}, "a": nil}, `{"a":null,"z":{"y":[1,"x"]}}`},
		{ptr("p"), `"p"`},
	} {
		if got := mustMarshal(t, tc.in); got != tc.want {
			t.Errorf("Marshal(%#v) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func ptr[T any](v T) *T { return &v }

func TestCanonJSON_RejectsFloat(t *testing.T) {
	for _, v := range []any{1.0, float32(2), math.NaN(), math.Inf(1), []any{0.5}, map[string]any{"f": 1.5}, complex(1, 2)} {
		if _, err := canonjson.Marshal(v); !errors.Is(err, canonjson.ErrUnsupported) {
			t.Errorf("Marshal(%#v) error = %v, want ErrUnsupported", v, err)
		}
	}
}

func TestCanonJSON_RejectsUnsupported(t *testing.T) {
	type pair struct{ A int }
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	for name, v := range map[string]any{
		"struct":          pair{A: 1},
		"bytes":           []byte("x"),
		"int key":         map[int]string{1: "a"},
		"chan":            make(chan int),
		"func":            func() {},
		"json.Number":     json.Number("1"),
		"cycle (depth)":   cyclic,
		"too deep":        nest(canonjson.MaxDepth + 1),
		"unsafe int":      int64(canonjson.MaxSafeInteger + 1),
		"unsafe negative": int64(-canonjson.MaxSafeInteger - 1),
		"unsafe uint":     uint64(1) << 63,
		"invalid UTF-8":   "\xff",
		"lone surrogate":  "\xed\xa0\x80", // U+D800 encoded in UTF-8
		"invalid key":     map[string]int{"\xff": 1},
		"nested invalid":  []any{map[string]any{"a": "\xc3"}},
	} {
		_, err := canonjson.Marshal(v)
		if !errors.Is(err, canonjson.ErrUnsupported) {
			t.Errorf("%s: error = %v, want ErrUnsupported", name, err)
		}
		if err != nil && strings.Contains(err.Error(), "\xff") {
			t.Errorf("%s: error echoes the input", name)
		}
	}
	if _, err := canonjson.Marshal(nest(canonjson.MaxDepth)); err != nil {
		t.Errorf("depth MaxDepth: %v, want accepted", err)
	}
}

// nest returns n arrays nested in each other.
func nest(n int) any {
	var v any = "leaf"
	for range n {
		v = []any{v}
	}
	return v
}

func TestCanonJSON_OutputIsValidJSON(t *testing.T) {
	in := map[string]any{"k\u0001": []any{"\u00e9\U0001F600", int64(-12), true, nil}}
	b := mustMarshal(t, in)
	var back any
	if err := json.Unmarshal([]byte(b), &back); err != nil {
		t.Fatalf("output %s is not JSON: %v", b, err)
	}
}
