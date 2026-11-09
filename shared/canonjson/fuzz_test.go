package canonjson_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/aleogr/iam/shared/canonjson"
)

// FuzzCanonJSON_RoundTrip decodes arbitrary JSON, canonicalizes it, decodes
// the result and canonicalizes again: the two outputs must be identical
// (canonicalization is a fixed point) and must decode to the same value as the
// input. Inputs canonjson refuses (floats, unsafe integers) must fail with
// ErrUnsupported, never panic.
func FuzzCanonJSON_RoundTrip(f *testing.F) {
	for _, s := range []string{
		`{"b":1,"a":[true,false,null,"x"]}`, `{"€":1,"\r":2,"😀":3,"דּ":4}`,
		`"\u0000\u001f\u007f "`, `[1.5]`, `9007199254740993`, `-9007199254740991`, `{"a":{"a":{"a":[]}}}`,
		`{"a":1,"a":2}`, `"\ud800"`, `[]`, `{}`, `-0`, `1e3`,
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.UseNumber()
		var in any
		if err := dec.Decode(&in); err != nil {
			return
		}
		v, ok := integers(in)
		first, err := canonjson.Marshal(v)
		if err != nil {
			if !errors.Is(err, canonjson.ErrUnsupported) {
				t.Fatalf("error %v is not ErrUnsupported", err)
			}
			return
		}
		if !ok {
			t.Fatalf("accepted a value with a non-integer or unsafe number: %s", first)
		}
		var back any
		dec = json.NewDecoder(bytes.NewReader(first))
		dec.UseNumber()
		if err := dec.Decode(&back); err != nil {
			t.Fatalf("output %q is not JSON: %v", first, err)
		}
		back, _ = integers(back)
		if !reflect.DeepEqual(back, v) {
			t.Fatalf("round trip changed the value:\n in  %#v\n out %#v", v, back)
		}
		second, err := canonjson.Marshal(back)
		if err != nil || !bytes.Equal(first, second) {
			t.Fatalf("not a fixed point: %q then %q (%v)", first, second, err)
		}
	})
}

// integers converts every json.Number into an int64 when it is an integer
// canonjson accepts, and into a float64 otherwise (which canonjson refuses);
// ok reports whether no float64 was produced.
func integers(v any) (out any, ok bool) {
	switch x := v.(type) {
	case json.Number:
		if n, err := strconv.ParseInt(string(x), 10, 64); err == nil && n >= -canonjson.MaxSafeInteger && n <= canonjson.MaxSafeInteger {
			return n, true
		}
		return 1.5, false
	case []any:
		ok = true
		for i, e := range x {
			var eok bool
			x[i], eok = integers(e)
			ok = ok && eok
		}
		return x, ok
	case map[string]any:
		return x, integersInMap(x)
	}
	return v, true
}

func integersInMap(m map[string]any) bool {
	ok := true
	for k, e := range m {
		var eok bool
		m[k], eok = integers(e)
		ok = ok && eok
	}
	return ok
}
