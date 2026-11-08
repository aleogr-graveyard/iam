package ids_test

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
)

var t0 = time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)

func gen(t time.Time) ids.Generator { return ids.Generator{Clock: &clock.Fixed{T: t}} }

func TestNew_IsVersion7AndVariant(t *testing.T) {
	for range 200 {
		id, err := gen(t0).New()
		if err != nil {
			t.Fatal(err)
		}
		if id[6]>>4 != 7 {
			t.Fatalf("version nibble = %x, want 7", id[6]>>4)
		}
		if id[8]>>6 != 0b10 {
			t.Fatalf("variant bits = %b, want 10", id[8]>>6)
		}
		s := id.String()
		if s[14] != '7' || !strings.ContainsRune("89ab", rune(s[19])) {
			t.Fatalf("bad string form %q", s)
		}
	}
}

func TestNew_EncodesTimestampBigEndian(t *testing.T) {
	id, err := gen(t0).New()
	if err != nil {
		t.Fatal(err)
	}
	var ms uint64
	for _, b := range id[:6] {
		ms = ms<<8 | uint64(b)
	}
	if ms != uint64(t0.UnixMilli()) {
		t.Fatalf("timestamp = %d, want %d", ms, t0.UnixMilli())
	}
}

func TestNew_RandomBitsDiffer(t *testing.T) {
	seen := map[ids.ID]bool{}
	for range 100 {
		id, err := gen(t0).New()
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
}

func TestNew_SortsByTime(t *testing.T) {
	c := &clock.Fixed{T: t0}
	g := ids.Generator{Clock: c}
	prev, err := g.New()
	if err != nil {
		t.Fatal(err)
	}
	for range 50 {
		c.Advance(time.Millisecond)
		next, err := g.New()
		if err != nil {
			t.Fatal(err)
		}
		if next.String() <= prev.String() {
			t.Fatalf("%s did not sort after %s", next, prev)
		}
		prev = next
	}
}

func TestNew_RejectsOutOfRangeTime(t *testing.T) {
	maxMs := int64(1)<<48 - 1
	cases := map[string]time.Time{
		"before epoch": time.UnixMilli(-1),
		"after 48bit":  time.UnixMilli(maxMs + 1),
		"far future":   time.UnixMilli(math.MaxInt64),
	}
	for name, tm := range cases {
		if _, err := gen(tm).New(); !errors.Is(err, ids.ErrTimeOutOfRange) {
			t.Errorf("%s: err = %v, want ErrTimeOutOfRange", name, err)
		}
	}
	for name, tm := range map[string]time.Time{"epoch": time.UnixMilli(0), "max": time.UnixMilli(maxMs)} {
		if _, err := gen(tm).New(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestNew_NilClock(t *testing.T) {
	if _, err := (ids.Generator{}).New(); err == nil {
		t.Fatal("expected error for nil clock")
	}
}

func TestParse_RoundTrip(t *testing.T) {
	id, err := gen(t0).New()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ids.Parse(id.String())
	if err != nil || got != id {
		t.Fatalf("Parse(%q) = %v, %v", id, got, err)
	}
}

func TestParse_RejectsMalformed(t *testing.T) {
	valid := "0192f3a0-7b3c-7abc-8def-0123456789ab"
	if _, err := ids.Parse(valid); err != nil {
		t.Fatalf("valid rejected: %v", err)
	}
	cases := map[string]string{
		"empty":         "",
		"uppercase":     strings.ToUpper(valid),
		"braces":        "{" + valid + "}",
		"urn":           "urn:uuid:" + valid,
		"too short":     valid[:35],
		"too long":      valid + "0",
		"no dashes":     strings.ReplaceAll(valid, "-", ""),
		"misplaced":     "0192f3a07-b3c-7abc-8def-0123456789ab",
		"non hex":       "0192f3a0-7b3c-7abc-8def-0123456789ag",
		"space":         " " + valid[1:],
		"version 4":     "0192f3a0-7b3c-4abc-8def-0123456789ab",
		"variant ncs":   "0192f3a0-7b3c-7abc-0def-0123456789ab",
		"variant ms":    "0192f3a0-7b3c-7abc-cdef-0123456789ab",
		"nul":           valid[:35] + "\x00",
		"multibyte":     "0192f3a0-7b3c-7abc-8def-0123456789é",
		"plus sign hex": "+192f3a0-7b3c-7abc-8def-0123456789ab",
	}
	for name, in := range cases {
		if _, err := ids.Parse(in); !errors.Is(err, ids.ErrInvalid) {
			t.Errorf("%s: Parse(%q) err = %v, want ErrInvalid", name, in, err)
		}
	}
}

func TestZero(t *testing.T) {
	var z ids.ID
	if !z.IsZero() {
		t.Fatal("zero ID must be zero")
	}
	if z.String() != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("zero string = %q", z)
	}
	got, err := ids.Parse(z.String())
	if err != nil || !got.IsZero() {
		t.Fatalf("zero must round-trip: %v, %v", got, err)
	}
	id, _ := gen(t0).New()
	if id.IsZero() {
		t.Fatal("generated ID reported zero")
	}
}

func TestUnmarshalText_Strict(t *testing.T) {
	var id ids.ID
	if err := id.UnmarshalText([]byte("0192F3A0-7B3C-7ABC-8DEF-0123456789AB")); !errors.Is(err, ids.ErrInvalid) {
		t.Fatalf("uppercase accepted: %v", err)
	}
	if !id.IsZero() {
		t.Fatal("failed unmarshal must leave the ID untouched")
	}
	if err := id.UnmarshalText([]byte("0192f3a0-7b3c-7abc-8def-0123456789ab")); err != nil {
		t.Fatal(err)
	}
	b, err := id.MarshalText()
	if err != nil || string(b) != "0192f3a0-7b3c-7abc-8def-0123456789ab" {
		t.Fatalf("MarshalText = %q, %v", b, err)
	}
}

func TestTypedIDs_DoNotConvert(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeFor[ids.TenantID](), reflect.TypeFor[ids.OrganizationID](),
		reflect.TypeFor[ids.UserID](), reflect.TypeFor[ids.ApplicationID](),
		reflect.TypeFor[ids.EventID](),
	}
	for i := range types {
		for j := range types {
			if i != j && (types[i] == types[j] || types[i].AssignableTo(types[j])) {
				t.Errorf("%v and %v must be distinct and not assignable", types[i], types[j])
			}
		}
	}
}

func TestTypedIDs_JSONRoundTrip(t *testing.T) {
	const u = "0192f3a0-7b3c-7abc-8def-0123456789ab"
	type doc struct {
		T ids.TenantID       `json:"t"`
		O ids.OrganizationID `json:"o"`
		U ids.UserID         `json:"u"`
		A ids.ApplicationID  `json:"a"`
		E ids.EventID        `json:"e"`
	}
	in := `{"t":"` + u + `","o":"` + u + `","u":"` + u + `","a":"` + u + `","e":"` + u + `"}`
	var d doc
	if err := json.Unmarshal([]byte(in), &d); err != nil {
		t.Fatal(err)
	}
	if d.T.String() != u || d.E.String() != u || d.T.IsZero() {
		t.Fatalf("decoded wrong: %+v", d)
	}
	out, err := json.Marshal(d)
	if err != nil || string(out) != in {
		t.Fatalf("Marshal = %s, %v", out, err)
	}
	for _, bad := range []string{`{"t":"X"}`, `{"t":1}`, `{"t":null}x`, `{"t":"` + strings.ToUpper(u) + `"}`} {
		var d2 doc
		if err := json.Unmarshal([]byte(bad), &d2); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
}

func TestTypedIDs_MapKeyMarshal(t *testing.T) {
	m := map[ids.TenantID]int{}
	var k ids.TenantID
	m[k] = 1
	if b, err := json.Marshal(m); err != nil || !strings.Contains(string(b), "00000000-") {
		t.Fatalf("map key marshal: %s %v", b, err)
	}
}

func FuzzParse(f *testing.F) {
	f.Add("0192f3a0-7b3c-7abc-8def-0123456789ab")
	f.Add("00000000-0000-0000-0000-000000000000")
	f.Add("0192F3A0-7B3C-7ABC-8DEF-0123456789AB")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		id, err := ids.Parse(s)
		if err != nil {
			return
		}
		if id.String() != s {
			t.Fatalf("round trip: Parse(%q).String() = %q", s, id.String())
		}
		var u ids.ID
		if err := u.UnmarshalText([]byte(s)); err != nil || u != id {
			t.Fatalf("UnmarshalText disagrees with Parse for %q", s)
		}
	})
}
