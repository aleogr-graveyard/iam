package jsonx_test

import (
	"bytes"
	"testing"

	"github.com/aleogr/iam/shared/jsonx"
)

func FuzzDecodeStrict(f *testing.F) {
	for _, s := range []string{
		`{"a":1,"s":"x","i":{"n":2}}`, `{"a":1,"a":2}`, `{"zzz":1}`, `[[[[`, `{"a":1}{"a":2}`, ``, `null`,
	} {
		f.Add([]byte(s))
	}
	lim := jsonx.Limits{MaxBytes: 256, MaxDepth: 8}
	f.Fuzz(func(t *testing.T, data []byte) {
		var got doc
		if err := jsonx.DecodeStrict(bytes.NewReader(data), &got, lim); err != nil && got != (doc{}) {
			t.Fatalf("value not zero after error %v: %+v", err, got)
		}
	})
}
