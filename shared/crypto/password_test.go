package crypto

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"
)

// Small parameters keep the tests fast; production values come from configuration.
var testParams = Argon2Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1}

func newTestHasher(t *testing.T) *PasswordHasher {
	t.Helper()
	return NewPasswordHasher(testParams, 2)
}

func mustHash(t *testing.T, h *PasswordHasher, pw string) string {
	t.Helper()
	phc, err := h.Hash(bounded(t), []byte(pw))
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	return phc
}

func TestPasswordHasher_RoundTrip(t *testing.T) {
	h := newTestHasher(t)
	phc := mustHash(t, h, "correct horse battery staple")
	if !strings.HasPrefix(phc, "$argon2id$v=19$m=64,t=1,p=1$") {
		t.Fatalf("unexpected PHC prefix: %s", phc)
	}
	ok, rehash, err := h.Verify(bounded(t), []byte("correct horse battery staple"), phc)
	if err != nil || !ok || rehash {
		t.Fatalf("Verify = %v, %v, %v", ok, rehash, err)
	}
	if again := mustHash(t, h, "correct horse battery staple"); again == phc {
		t.Fatal("two hashes of the same password share a salt")
	}
}

// Pins the PHC encoding against the reference Argon2id derivation.
func TestPasswordHasher_PHCFormat(t *testing.T) {
	phc := mustHash(t, newTestHasher(t), "pw")
	parts := strings.Split(phc, "$")
	if len(parts) != 6 {
		t.Fatalf("parts: %d", len(parts))
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		t.Fatalf("salt: %d bytes, %v", len(salt), err)
	}
	key, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) != 32 {
		t.Fatalf("key: %d bytes, %v", len(key), err)
	}
	if want := argon2.IDKey([]byte("pw"), salt, 1, 64, 1, 32); !bytes.Equal(key, want) {
		t.Fatal("key does not match the reference Argon2id derivation")
	}
}

func TestPasswordHasher_WrongPassword(t *testing.T) {
	h := newTestHasher(t)
	phc := mustHash(t, h, "right")
	for _, pw := range []string{"wrong", "Right", "right ", ""} {
		ok, rehash, err := h.Verify(bounded(t), []byte(pw), phc)
		if err != nil || ok || rehash {
			t.Errorf("Verify(%q) = %v, %v, %v", pw, ok, rehash, err)
		}
	}
}

func TestPasswordHasher_NeedsRehashWhenParamsChange(t *testing.T) {
	old := NewPasswordHasher(testParams, 1)
	phc := mustHash(t, old, "pw")
	for _, p := range []Argon2Params{
		{MemoryKiB: 128, Iterations: 1, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 2, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 1, Parallelism: 2},
	} {
		ok, rehash, err := NewPasswordHasher(p, 1).Verify(bounded(t), []byte("pw"), phc)
		if err != nil || !ok || !rehash {
			t.Errorf("params %+v: Verify = %v, %v, %v; want ok and rehash", p, ok, rehash, err)
		}
	}
}

// A stored hash may cost at most 4x the configured parameters (DoS through stored hashes).
func TestPasswordHasher_ParameterCeiling(t *testing.T) {
	h := newTestHasher(t)
	atLimit := mustHash(t, NewPasswordHasher(Argon2Params{MemoryKiB: 256, Iterations: 4, Parallelism: 4}, 1), "pw")
	if ok, rehash, err := h.Verify(bounded(t), []byte("pw"), atLimit); err != nil || !ok || !rehash {
		t.Fatalf("4x hash: %v, %v, %v", ok, rehash, err)
	}
	for _, p := range []Argon2Params{
		{MemoryKiB: 257, Iterations: 1, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 5, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 1, Parallelism: 5},
	} {
		phc := mustHash(t, NewPasswordHasher(p, 1), "pw")
		if _, _, err := h.Verify(bounded(t), []byte("pw"), phc); !errors.Is(err, ErrHashParamsTooHigh) {
			t.Errorf("params %+v: err = %v, want ErrHashParamsTooHigh", p, err)
		}
	}
}

func TestPasswordHasher_RejectsMalformedPHC(t *testing.T) {
	h := newTestHasher(t)
	valid := mustHash(t, h, "pw")
	parts := strings.Split(valid, "$")
	salt, key := parts[4], parts[5]
	with := func(head string) string { return head + "$" + salt + "$" + key }
	cases := map[string]string{
		"empty":            "",
		"argon2i":          with("$argon2i$v=19$m=64,t=1,p=1"),
		"argon2d":          with("$argon2d$v=19$m=64,t=1,p=1"),
		"bcrypt":           "$2b$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234",
		"no leading $":     strings.TrimPrefix(valid, "$"),
		"missing version":  "$argon2id$m=64,t=1,p=1$" + salt + "$" + key,
		"missing params":   "$argon2id$v=19$" + salt + "$" + key,
		"missing hash":     "$argon2id$v=19$m=64,t=1,p=1$" + salt,
		"empty hash":       "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$",
		"extra field":      valid + "$extra",
		"old version":      with("$argon2id$v=16$m=64,t=1,p=1"),
		"version leading0": with("$argon2id$v=019$m=64,t=1,p=1"),
		"missing p":        with("$argon2id$v=19$m=64,t=1"),
		"params reordered": with("$argon2id$v=19$t=1,m=64,p=1"),
		"extra param":      with("$argon2id$v=19$m=64,t=1,p=1,k=1"),
		"keyid param":      with("$argon2id$v=19$m=64,t=1,p=1,keyid=AAAA"),
		"huge m":           with("$argon2id$v=19$m=4294967295,t=1,p=1"),
		"overflow m":       with("$argon2id$v=19$m=99999999999999999999,t=1,p=1"),
		"negative t":       with("$argon2id$v=19$m=64,t=-1,p=1"),
		"plus sign":        with("$argon2id$v=19$m=+64,t=1,p=1"),
		"leading zero":     with("$argon2id$v=19$m=064,t=1,p=1"),
		"zero t":           with("$argon2id$v=19$m=64,t=0,p=1"),
		"zero p":           with("$argon2id$v=19$m=64,t=1,p=0"),
		"m below 8p":       with("$argon2id$v=19$m=8,t=1,p=2"),
		"space":            with("$argon2id$v=19$m=64, t=1,p=1"),
		"padded salt":      "$argon2id$v=19$m=64,t=1,p=1$" + salt + "==$" + key,
		"bad base64":       "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + key[:len(key)-1] + "!",
		"url alphabet":     "$argon2id$v=19$m=64,t=1,p=1$" + "____________________ww" + "$" + key,
		"short salt":       "$argon2id$v=19$m=64,t=1,p=1$" + "AAAAAAAAAA" + "$" + key,
		"short hash":       "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + "AAAAAAAAAAAAAAAAAAAAAA",
		"long hash":        "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + strings.Repeat("A", 128),
		"oversized":        valid + strings.Repeat("A", 4096),
	}
	for name, phc := range cases {
		t.Run(name, func(t *testing.T) {
			ok, rehash, err := h.Verify(bounded(t), []byte("pw"), phc)
			if ok || rehash || err == nil {
				t.Fatalf("Verify = %v, %v, %v; want rejection", ok, rehash, err)
			}
			if !errors.Is(err, ErrMalformedHash) && !errors.Is(err, ErrHashParamsTooHigh) {
				t.Fatalf("err = %v, want ErrMalformedHash or ErrHashParamsTooHigh", err)
			}
			if strings.Contains(err.Error(), salt) || strings.Contains(err.Error(), key) {
				t.Fatalf("error leaks hash material: %v", err)
			}
		})
	}
}

// bcrypt silently truncates at 72 bytes; Argon2id must never truncate.
func TestPasswordHasher_NoTruncation(t *testing.T) {
	h := newTestHasher(t)
	prefix := strings.Repeat("a", 72)
	phc := mustHash(t, h, prefix+"X")
	for _, pw := range []string{prefix + "Y", prefix, prefix + "XX"} {
		ok, _, err := h.Verify(bounded(t), []byte(pw), phc)
		if err != nil || ok {
			t.Errorf("password of %d bytes sharing the first 72 verified: %v, %v", len(pw), ok, err)
		}
	}
	long := strings.Repeat("b", MaxPasswordBytes)
	phcLong := mustHash(t, h, long)
	ok, _, err := h.Verify(bounded(t), []byte(long[:MaxPasswordBytes-1]+"c"), phcLong)
	if err != nil || ok {
		t.Fatalf("passwords differing in the last byte verified: %v, %v", ok, err)
	}
}

func TestPasswordHasher_TooLong(t *testing.T) {
	h := newTestHasher(t)
	pw := []byte(strings.Repeat("p", MaxPasswordBytes+1))
	if _, err := h.Hash(bounded(t), pw); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("Hash err = %v, want ErrPasswordTooLong", err)
	}
	phc := mustHash(t, h, "pw")
	ok, rehash, err := h.Verify(bounded(t), pw, phc)
	if ok || rehash || !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("Verify = %v, %v, %v; want ErrPasswordTooLong", ok, rehash, err)
	}
	if strings.Contains(err.Error(), "ppp") {
		t.Fatalf("error leaks the password: %v", err)
	}
}

func TestPasswordHasher_ConcurrencyLimit(t *testing.T) {
	h := NewPasswordHasher(testParams, 1)
	phc := mustHash(t, h, "pw")
	release, err := h.acquire(bounded(t))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := h.Hash(ctx, []byte("pw")); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Hash err = %v, want DeadlineExceeded", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()
	if ok, _, err := h.Verify(ctx2, []byte("pw"), phc); ok || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Verify = %v, %v; want DeadlineExceeded", ok, err)
	}
	release()
	// The slot is free again, and failed calls did not leak it.
	for i := range 3 {
		if ok, _, err := h.Verify(bounded(t), []byte("pw"), phc); err != nil || !ok {
			t.Fatalf("call %d after release: %v, %v", i, ok, err)
		}
	}
}

func TestPasswordHasher_CanceledContext(t *testing.T) {
	h := newTestHasher(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// The slot is free, so a select alone would pick at random; repeat to be sure
	// an ended context never takes it.
	for range 50 {
		if _, err := h.Hash(ctx, []byte("pw")); !errors.Is(err, context.Canceled) {
			t.Fatalf("Hash err = %v, want Canceled", err)
		}
	}
}

// bounded fails fast instead of hanging when a slot has leaked.
func bounded(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestPasswordHasher_ReleasesSlotOnEveryPath(t *testing.T) {
	h := NewPasswordHasher(testParams, 1)
	phc, err := h.Hash(bounded(t), []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]func() error{
		"right password": func() error { _, _, err := h.Verify(bounded(t), []byte("pw"), phc); return err },
		"wrong password": func() error { _, _, err := h.Verify(bounded(t), []byte("wrong"), phc); return err },
		"malformed hash": func() error { _, _, _ = h.Verify(bounded(t), []byte("pw"), "$argon2i$bad"); return nil },
		"too long":       func() error { _, _ = h.Hash(bounded(t), make([]byte, MaxPasswordBytes+1)); return nil },
		"hash":           func() error { _, err := h.Hash(bounded(t), []byte("pw")); return err },
	}
	for name, call := range paths {
		if err := call(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if _, err := h.Hash(bounded(t), []byte("pw")); err != nil {
			t.Fatalf("slot leaked after %s: %v", name, err)
		}
	}
}

func TestPasswordHasher_ReleasesSlotOnPanic(t *testing.T) {
	h := NewPasswordHasher(testParams, 1)
	func() {
		defer func() { _ = recover() }()
		_ = h.withSlot(bounded(t), func() error { panic("boom") })
	}()
	if _, err := h.Hash(bounded(t), []byte("pw")); err != nil {
		t.Fatalf("slot leaked after panic: %v", err)
	}
}

func TestPasswordHasher_InvalidParams(t *testing.T) {
	for _, p := range []Argon2Params{
		{MemoryKiB: 0, Iterations: 1, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 0, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 1, Parallelism: 0},
		{MemoryKiB: 8, Iterations: 1, Parallelism: 2},
	} {
		h := NewPasswordHasher(p, 1)
		if _, err := h.Hash(bounded(t), []byte("pw")); !errors.Is(err, ErrInvalidParams) {
			t.Errorf("params %+v: Hash err = %v, want ErrInvalidParams", p, err)
		}
		if _, _, err := h.Verify(bounded(t), []byte("pw"), "x"); !errors.Is(err, ErrInvalidParams) {
			t.Errorf("params %+v: Verify err = %v, want ErrInvalidParams", p, err)
		}
	}
}

func TestPasswordHasher_ConcurrentUse(t *testing.T) {
	h := newTestHasher(t)
	errs := make(chan error, 8)
	for i := range 8 {
		go func() {
			pw := []byte(fmt.Sprintf("pw-%d", i))
			phc, err := h.Hash(bounded(t), pw)
			if err == nil {
				var ok bool
				ok, _, err = h.Verify(bounded(t), pw, phc)
				if err == nil && !ok {
					err = errors.New("round trip failed")
				}
			}
			errs <- err
		}()
	}
	for range 8 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}
