package crypto

import "testing"

func TestConstantTimeEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"", "", true},
		{"secret", "secret", true},
		{"secret", "secreT", false},
		{"secret", "secrets", false},
		{"", "x", false},
	}
	for _, c := range cases {
		if got := ConstantTimeEqual([]byte(c.a), []byte(c.b)); got != c.want {
			t.Errorf("ConstantTimeEqual(%q, %q) = %v", c.a, c.b, got)
		}
	}
	if !ConstantTimeEqual(nil, []byte{}) {
		t.Error("nil and empty must be equal")
	}
}
