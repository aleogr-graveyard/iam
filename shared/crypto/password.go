package crypto

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// MaxPasswordBytes is the longest accepted password. Longer input is refused,
// never truncated (NIST SP 800-63B; ADR 0027).
const MaxPasswordBytes = 1024

const (
	saltBytes = 16
	keyBytes  = uint32(32)

	// The only accepted algorithm and version.
	phcAlgorithm = "argon2id"
	phcVersion   = "v=19"

	// maxPHCBytes bounds the stored string before it is parsed; a well-formed
	// hash with the largest possible parameters is about 120 bytes.
	maxPHCBytes = 256

	// paramCeiling is how many times the configured cost a stored hash may
	// demand before Verify refuses it (DoS through stored hashes, TM-PWD-003).
	paramCeiling = 4
)

// Errors returned by PasswordHasher. They never contain the password or the hash.
var (
	ErrPasswordTooLong   = errors.New("crypto: password too long")
	ErrMalformedHash     = errors.New("crypto: malformed password hash")
	ErrHashParamsTooHigh = errors.New("crypto: password hash parameters above the accepted ceiling")
	ErrInvalidParams     = errors.New("crypto: invalid Argon2id parameters")
)

// Argon2Params are the Argon2id cost parameters.
type Argon2Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
}

// valid reports whether Argon2id can run with p (RFC 9106 section 3.1:
// memory at least 8 KiB per lane).
func (p Argon2Params) valid() bool {
	return p.Iterations >= 1 && p.Parallelism >= 1 && uint64(p.MemoryKiB) >= 8*uint64(p.Parallelism)
}

// exceeds reports whether any cost in p is above paramCeiling times limit.
func (p Argon2Params) exceeds(limit Argon2Params) bool {
	return uint64(p.MemoryKiB) > paramCeiling*uint64(limit.MemoryKiB) ||
		uint64(p.Iterations) > paramCeiling*uint64(limit.Iterations) ||
		uint64(p.Parallelism) > paramCeiling*uint64(limit.Parallelism)
}

// PasswordHasher hashes and verifies passwords with Argon2id, in PHC string
// format, and bounds how many derivations run at once. It is safe for
// concurrent use.
type PasswordHasher struct {
	params Argon2Params
	sem    chan struct{}
}

// NewPasswordHasher returns a hasher with the given parameters that runs at
// most maxConcurrent derivations at a time (values below 1 mean 1). Invalid
// parameters make every Hash and Verify fail with ErrInvalidParams.
func NewPasswordHasher(p Argon2Params, maxConcurrent int) *PasswordHasher {
	return &PasswordHasher{params: p, sem: make(chan struct{}, max(maxConcurrent, 1))}
}

// Hash returns the PHC string $argon2id$v=19$m=..,t=..,p=..$salt$hash of
// password with a fresh 16-byte salt. It waits for a free slot while ctx
// allows and returns ctx.Err() otherwise.
func (h *PasswordHasher) Hash(ctx context.Context, password []byte) (string, error) {
	if err := h.check(password); err != nil {
		return "", err
	}
	salt, err := RandomBytes(saltBytes)
	if err != nil {
		return "", err
	}
	var key []byte
	err = h.withSlot(ctx, func() error {
		key = derive(password, salt, h.params)
		return nil
	})
	if err != nil {
		return "", err
	}
	phc := encodePHC(h.params, salt, key)
	clear(key)
	return phc, nil
}

// Verify reports whether password matches the PHC string phc and, when it
// does, whether phc was made with parameters other than the configured ones
// and should be replaced. A wrong password and a right one take the same path.
// Hashes above 4x the configured cost are refused before any derivation.
func (h *PasswordHasher) Verify(ctx context.Context, password []byte, phc string) (ok, needsRehash bool, err error) {
	if err := h.check(password); err != nil {
		return false, false, err
	}
	stored, err := parsePHC(phc)
	if err != nil {
		return false, false, err
	}
	if stored.params.exceeds(h.params) {
		return false, false, ErrHashParamsTooHigh
	}
	err = h.withSlot(ctx, func() error {
		key := derive(password, stored.salt, stored.params)
		ok = ConstantTimeEqual(key, stored.key)
		clear(key)
		return nil
	})
	if err != nil {
		return false, false, err
	}
	return ok, ok && stored.params != h.params, nil
}

func (h *PasswordHasher) check(password []byte) error {
	if !h.params.valid() {
		return ErrInvalidParams
	}
	if len(password) > MaxPasswordBytes {
		return fmt.Errorf("%w: maximum %d bytes", ErrPasswordTooLong, MaxPasswordBytes)
	}
	return nil
}

// withSlot runs fn while holding a concurrency slot; the slot is released on
// every path, panics included.
func (h *PasswordHasher) withSlot(ctx context.Context, fn func() error) error {
	release, err := h.acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	return fn()
}

// acquire takes a slot, or returns ctx.Err() if ctx ends first. An already
// ended context never takes a slot.
func (h *PasswordHasher) acquire(ctx context.Context) (release func(), err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case h.sem <- struct{}{}:
		return func() { <-h.sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// derive returns the keyBytes-long Argon2id key; parsePHC accepts only stored
// keys of that length.
func derive(password, salt []byte, p Argon2Params) []byte {
	return argon2.IDKey(password, salt, p.Iterations, p.MemoryKiB, p.Parallelism, keyBytes)
}

var b64 = base64.RawStdEncoding.Strict()

const b64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func encodePHC(p Argon2Params, salt, key []byte) string {
	return fmt.Sprintf("$%s$%s$m=%d,t=%d,p=%d$%s$%s", phcAlgorithm, phcVersion,
		p.MemoryKiB, p.Iterations, p.Parallelism, b64.EncodeToString(salt), b64.EncodeToString(key))
}

type storedHash struct {
	params    Argon2Params
	salt, key []byte
}

// parsePHC accepts exactly the form encodePHC produces: argon2id, version 19,
// parameters m, t, p in that order as canonical decimals, a 16-byte salt and a
// 32-byte hash in unpadded standard base64.
func parsePHC(s string) (storedHash, error) {
	if len(s) > maxPHCBytes {
		return storedHash{}, fmt.Errorf("%w: too long", ErrMalformedHash)
	}
	f := strings.Split(s, "$")
	if len(f) != 6 || f[0] != "" {
		return storedHash{}, fmt.Errorf("%w: wrong number of fields", ErrMalformedHash)
	}
	if f[1] != phcAlgorithm || f[2] != phcVersion {
		return storedHash{}, fmt.Errorf("%w: unsupported algorithm or version", ErrMalformedHash)
	}
	p, err := parseParams(f[3])
	if err != nil {
		return storedHash{}, err
	}
	salt, err := decodeExact(f[4], saltBytes)
	if err != nil {
		return storedHash{}, err
	}
	key, err := decodeExact(f[5], int(keyBytes))
	if err != nil {
		return storedHash{}, err
	}
	return storedHash{params: p, salt: salt, key: key}, nil
}

func parseParams(s string) (Argon2Params, error) {
	kv := strings.Split(s, ",")
	if len(kv) != 3 {
		return Argon2Params{}, fmt.Errorf("%w: parameters", ErrMalformedHash)
	}
	m, errM := parseParam(kv[0], "m=")
	t, errT := parseParam(kv[1], "t=")
	p, errP := parseParam(kv[2], "p=")
	if errors.Join(errM, errT, errP) != nil || m > math.MaxUint32 || t > math.MaxUint32 || p > math.MaxUint8 {
		return Argon2Params{}, fmt.Errorf("%w: parameters", ErrMalformedHash)
	}
	params := Argon2Params{MemoryKiB: uint32(m), Iterations: uint32(t), Parallelism: uint8(p)}
	if !params.valid() {
		return Argon2Params{}, fmt.Errorf("%w: parameters out of range", ErrMalformedHash)
	}
	return params, nil
}

// parseParam reads "<prefix><canonical decimal>": digits only, no sign, no
// leading zero, at most 10 digits.
func parseParam(s, prefix string) (uint64, error) {
	v, ok := strings.CutPrefix(s, prefix)
	if !ok || v == "" || len(v) > 10 || (len(v) > 1 && v[0] == '0') || strings.Trim(v, "0123456789") != "" {
		return 0, ErrMalformedHash
	}
	return strconv.ParseUint(v, 10, 64)
}

// decodeExact decodes exactly n bytes of unpadded standard base64. The
// alphabet is checked first because the decoder silently skips CR and LF,
// which would otherwise let a shorter (even empty) salt or key through.
func decodeExact(s string, n int) ([]byte, error) {
	if len(s) != b64.EncodedLen(n) || strings.Trim(s, b64Alphabet) != "" {
		return nil, fmt.Errorf("%w: salt or hash encoding", ErrMalformedHash)
	}
	b, err := b64.DecodeString(s)
	if err != nil || len(b) != n {
		return nil, fmt.Errorf("%w: salt or hash encoding", ErrMalformedHash)
	}
	return b, nil
}
