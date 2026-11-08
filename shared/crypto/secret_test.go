package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func secret32(b byte) []byte { return bytes.Repeat([]byte{b}, 32) }

func mustHashSecret(t *testing.T, s []byte) []byte {
	t.Helper()
	h, err := HashSecret(s)
	if err != nil {
		t.Fatalf("HashSecret: %v", err)
	}
	return h
}

// Pins the construction: plain SHA-256 of the secret, 32 bytes.
func TestHashSecret_IsSHA256(t *testing.T) {
	h := mustHashSecret(t, []byte("0123456789abcdef"))
	ref := "9f9f5111f7b27a781f1f1ddde5ebc2dd2b796bfc7365c9c28b548e564176929f"
	if got := hex.EncodeToString(h); got != ref {
		t.Fatalf("HashSecret = %s, want %s", got, ref)
	}
	if len(h) != SecretHashBytes {
		t.Fatalf("len = %d, want %d", len(h), SecretHashBytes)
	}
}

func TestHashSecret_RefusesShortSecret(t *testing.T) {
	for _, n := range []int{0, 1, MinSecretBytes - 1} {
		if _, err := HashSecret(make([]byte, n)); !errors.Is(err, ErrShortSecret) {
			t.Errorf("len %d: err = %v, want ErrShortSecret", n, err)
		}
	}
	if _, err := HashSecret(nil); !errors.Is(err, ErrShortSecret) {
		t.Errorf("nil: err = %v, want ErrShortSecret", err)
	}
	mustHashSecret(t, make([]byte, MinSecretBytes))
}

func TestVerifySecret_MatchesOnlyExact(t *testing.T) {
	s := secret32(0x42)
	h := mustHashSecret(t, s)
	if !VerifySecret(s, h) {
		t.Fatal("VerifySecret refused the right secret")
	}
	other := secret32(0x42)
	other[31] ^= 1
	if VerifySecret(other, h) {
		t.Fatal("VerifySecret accepted a different secret")
	}
	flipped := bytes.Clone(h)
	flipped[0] ^= 1
	if VerifySecret(s, flipped) {
		t.Fatal("VerifySecret accepted a different hash")
	}
}

func TestHashSecret_VerifyRefusesMalformedInput(t *testing.T) {
	s := secret32(7)
	h := mustHashSecret(t, s)
	cases := map[string]struct{ secret, hash []byte }{
		"empty secret":       {nil, h},
		"short secret":       {s[:MinSecretBytes-1], mustSHA(s[:MinSecretBytes-1])},
		"nil hash":           {s, nil},
		"short hash":         {s, h[:31]},
		"long hash":          {s, append(bytes.Clone(h), 0)},
		"empty secret, hash": {[]byte{}, mustSHA(nil)},
	}
	for name, c := range cases {
		if VerifySecret(c.secret, c.hash) {
			t.Errorf("%s: VerifySecret = true", name)
		}
	}
}

// mustSHA returns the bare SHA-256 of b, bypassing the length floor, so the
// test proves VerifySecret refuses a short secret even with its matching hash.
func mustSHA(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

func TestHashSecret_DoesNotRetainInput(t *testing.T) {
	s := secret32(1)
	h := mustHashSecret(t, s)
	s[0] = 2
	if !VerifySecret(secret32(1), h) {
		t.Fatal("hash changed when the caller mutated the secret")
	}
}
