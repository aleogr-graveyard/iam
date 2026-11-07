package crypto

import (
	"bytes"
	"errors"
	"regexp"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	a, err := RandomBytes(32)
	if err != nil || len(a) != 32 {
		t.Fatalf("RandomBytes(32) = %d bytes, %v", len(a), err)
	}
	b, err := RandomBytes(32)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, b) {
		t.Fatal("two random draws are equal")
	}
	for _, n := range []int{0, -1} {
		if _, err := RandomBytes(n); !errors.Is(err, ErrInvalidLength) {
			t.Errorf("RandomBytes(%d) err = %v, want ErrInvalidLength", n, err)
		}
	}
}

func TestRandomToken_LengthAndAlphabet(t *testing.T) {
	alphabet := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	for _, tc := range []struct{ n, chars int }{{16, 22}, {32, 43}, {33, 44}} {
		tok, err := RandomToken(tc.n)
		if err != nil {
			t.Fatalf("RandomToken(%d): %v", tc.n, err)
		}
		if len(tok) != tc.chars {
			t.Errorf("RandomToken(%d) has %d chars, want %d", tc.n, len(tok), tc.chars)
		}
		if !alphabet.MatchString(tok) {
			t.Errorf("RandomToken(%d) = %q is not unpadded base64url", tc.n, tok)
		}
	}
	a, _ := RandomToken(32)
	b, _ := RandomToken(32)
	if a == b {
		t.Fatal("two tokens are equal")
	}
}

func TestRandomToken_RejectsWeakLength(t *testing.T) {
	for _, n := range []int{-1, 0, 15} {
		if _, err := RandomToken(n); !errors.Is(err, ErrInvalidLength) {
			t.Errorf("RandomToken(%d) err = %v, want ErrInvalidLength", n, err)
		}
	}
}
