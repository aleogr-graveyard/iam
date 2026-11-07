package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
)

func testKey() []byte { return bytes.Repeat([]byte{0x42}, 32) }

func TestMAC_RejectsShortKey(t *testing.T) {
	for _, n := range []int{0, 16, 31} {
		key := bytes.Repeat([]byte{0x01}, n)
		if _, err := NewMAC(key); !errors.Is(err, ErrShortKey) {
			t.Errorf("NewMAC(%d bytes) err = %v, want ErrShortKey", n, err)
		}
	}
	if _, err := NewMAC(bytes.Repeat([]byte{0x01}, 64)); err != nil {
		t.Errorf("NewMAC(64 bytes): %v", err)
	}
}

func TestMAC_DomainSeparation(t *testing.T) {
	m, err := NewMAC(testKey())
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("payload")
	a := m.Sum("session-id", msg)
	b := m.Sum("csrf", msg)
	if bytes.Equal(a, b) {
		t.Fatal("same message under different domains produced the same tag")
	}
	// The length prefix keeps the domain/message boundary unambiguous.
	if bytes.Equal(m.Sum("ab", []byte("c")), m.Sum("a", []byte("bc"))) {
		t.Fatal("domain/message boundary is ambiguous")
	}
	if m.Verify("csrf", msg, a) {
		t.Fatal("tag verified under another domain")
	}
}

func TestMAC_SumAndVerify(t *testing.T) {
	m, err := NewMAC(testKey())
	if err != nil {
		t.Fatal(err)
	}
	tag := m.Sum("d", []byte("m"))
	if len(tag) != sha256.Size {
		t.Fatalf("tag length %d", len(tag))
	}
	if !m.Verify("d", []byte("m"), tag) {
		t.Fatal("valid tag rejected")
	}
	bad := bytes.Clone(tag)
	bad[0] ^= 1
	for name, tg := range map[string][]byte{"flipped": bad, "truncated": tag[:16], "empty": nil} {
		if m.Verify("d", []byte("m"), tg) {
			t.Errorf("%s tag accepted", name)
		}
	}
	if m.Verify("d", []byte("n"), tag) {
		t.Fatal("tag accepted for another message")
	}
}

// Pins the construction: HMAC-SHA-256(key, be64(len(domain)) || domain || msg).
func TestMAC_KnownConstruction(t *testing.T) {
	key := testKey()
	m, err := NewMAC(key)
	if err != nil {
		t.Fatal(err)
	}
	h := hmac.New(sha256.New, key)
	h.Write([]byte{0, 0, 0, 0, 0, 0, 0, 3})
	h.Write([]byte("dom"))
	h.Write([]byte("msg"))
	if want := h.Sum(nil); !bytes.Equal(m.Sum("dom", []byte("msg")), want) {
		t.Fatal("MAC construction changed")
	}
}

func TestMAC_KeyIsCopied(t *testing.T) {
	key := testKey()
	m, err := NewMAC(key)
	if err != nil {
		t.Fatal(err)
	}
	before := m.Sum("d", nil)
	key[0] ^= 0xff
	if !bytes.Equal(before, m.Sum("d", nil)) {
		t.Fatal("MAC aliases the caller's key slice")
	}
}

func TestMAC_ErrorHasNoKey(t *testing.T) {
	_, err := NewMAC([]byte("short-secret-key"))
	if err == nil || strings.Contains(err.Error(), "short-secret-key") {
		t.Fatalf("err = %v", err)
	}
}
