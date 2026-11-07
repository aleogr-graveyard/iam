package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-jose/go-jose/v4"

	"github.com/aleogr/iam/shared/jsonx"
)

// Algorithm is a JWS signature algorithm name (RFC 7518, section 3.1).
type Algorithm string

// ES256 is ECDSA over P-256 with SHA-256, the only algorithm supported in
// milestone 1 (ADR 0010).
const ES256 Algorithm = "ES256"

// MaxTokenBytes is the longest compact JWS that Verify parses (8 KiB) and
// that Sign produces.
const MaxTokenBytes = 8 << 10

// maxClaimsDepth bounds the nesting of decoded claims.
const maxClaimsDepth = 16

// Errors returned by the JOSE functions. They never contain token content or
// key material.
var (
	ErrTokenTooLarge        = errors.New("crypto: token too large")
	ErrMalformedToken       = errors.New("crypto: malformed token")
	ErrUnsupportedAlgorithm = errors.New("crypto: unsupported algorithm")
	ErrUnknownKey           = errors.New("crypto: unknown key ID")
	ErrInvalidSignature     = errors.New("crypto: invalid signature")
	ErrInvalidKey           = errors.New("crypto: invalid key")
)

// PublicKey is a verification key. It is built only by this package (from a
// SigningKey or a KeySet), so callers never handle crypto/* types.
type PublicKey struct {
	ID  string
	Alg Algorithm
	key *ecdsa.PublicKey
}

// Verifier verifies compact JWS tokens against a fixed set of keys and an
// explicit algorithm allowlist. It is safe for concurrent use.
type Verifier struct {
	allowed map[Algorithm]bool
	algs    []jose.SignatureAlgorithm
	keys    map[string]PublicKey
}

// NewVerifier returns a Verifier for keys, accepting only the algorithms in
// allowed. allowed must be non-empty and hold only supported algorithms;
// keys must be non-empty, built by this package, with distinct IDs and an
// allowed algorithm each.
func NewVerifier(keys []PublicKey, allowed []Algorithm) (*Verifier, error) {
	v := &Verifier{allowed: map[Algorithm]bool{}, keys: map[string]PublicKey{}}
	if err := v.setAllowed(allowed); err != nil {
		return nil, err
	}
	if err := v.setKeys(keys); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Verifier) setAllowed(allowed []Algorithm) error {
	if len(allowed) == 0 {
		return fmt.Errorf("%w: empty allowlist", ErrUnsupportedAlgorithm)
	}
	for _, a := range allowed {
		if a != ES256 {
			return ErrUnsupportedAlgorithm
		}
		if !v.allowed[a] {
			v.allowed[a] = true
			v.algs = append(v.algs, jose.SignatureAlgorithm(a))
		}
	}
	return nil
}

func (v *Verifier) setKeys(keys []PublicKey) error {
	if len(keys) == 0 {
		return fmt.Errorf("%w: no verification keys", ErrInvalidKey)
	}
	if err := checkUniqueIDs(keys); err != nil {
		return err
	}
	for _, k := range keys {
		if k.key == nil || k.ID == "" {
			return fmt.Errorf("%w: incomplete public key", ErrInvalidKey)
		}
		if !v.allowed[k.Alg] {
			return fmt.Errorf("%w: key algorithm not allowed", ErrUnsupportedAlgorithm)
		}
		v.keys[k.ID] = k
	}
	return nil
}

// header is the only protected header Verify accepts: members other than
// these (jku, jwk, x5u, x5c, crit, b64, ...) are refused, so no key named or
// embedded in a token is ever fetched or used.
type header struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// Verify checks token and decodes its claims into dst, which is written only
// on success. In order: length at most MaxTokenBytes (before any parsing);
// three segments of canonical unpadded base64url; a header of only alg, kid
// and typ; alg in the allowlist; kid naming a known key of that algorithm;
// the signature; claims decoded strictly (duplicate or unknown members
// refused, see jsonx). Keys in the header are never used.
//
// ECDSA signatures are malleable: (r, n-s) verifies like (r, s). Identify
// tokens by a claim (jti), never by their bytes.
func (v *Verifier) Verify(token string, dst any) error {
	if len(token) > MaxTokenBytes {
		return fmt.Errorf("%w: maximum %d bytes", ErrTokenTooLarge, MaxTokenBytes)
	}
	h, err := parseCompact(token)
	if err != nil {
		return err
	}
	if !v.allowed[Algorithm(h.Alg)] {
		return ErrUnsupportedAlgorithm
	}
	key, ok := v.keys[h.Kid]
	if !ok || key.Alg != Algorithm(h.Alg) {
		return ErrUnknownKey
	}
	jws, err := jose.ParseSignedCompact(token, v.algs)
	if err != nil {
		return fmt.Errorf("%w: JWS", ErrMalformedToken)
	}
	payload, err := jws.Verify(key.key)
	if err != nil {
		return ErrInvalidSignature
	}
	lim := jsonx.Limits{MaxBytes: MaxTokenBytes, MaxDepth: maxClaimsDepth}
	if err := jsonx.DecodeStrict(bytes.NewReader(payload), dst, lim); err != nil {
		return fmt.Errorf("%w: claims", ErrMalformedToken)
	}
	return nil
}

// parseCompact checks the compact serialization (RFC 7515, section 7.1) and
// returns its protected header. Every segment must be canonical unpadded
// base64url, so that no two strings verify as the same token through
// whitespace, padding bits or another alphabet.
func parseCompact(token string) (header, error) {
	segs := strings.Split(token, ".")
	if len(segs) != 3 {
		return header{}, fmt.Errorf("%w: want three segments", ErrMalformedToken)
	}
	var raw [3][]byte
	for i, s := range segs {
		b, err := decodeSegment(s)
		if err != nil {
			return header{}, err
		}
		raw[i] = b
	}
	var h header
	lim := jsonx.Limits{MaxBytes: MaxTokenBytes, MaxDepth: 1}
	if err := jsonx.DecodeStrict(bytes.NewReader(raw[0]), &h, lim); err != nil {
		return header{}, fmt.Errorf("%w: header", ErrMalformedToken)
	}
	return h, nil
}

var b64urlStrict = base64.RawURLEncoding.Strict()

func decodeSegment(s string) ([]byte, error) {
	if strings.Trim(s, b64urlAlphabet) != "" {
		return nil, fmt.Errorf("%w: segment encoding", ErrMalformedToken)
	}
	b, err := b64urlStrict.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: segment encoding", ErrMalformedToken)
	}
	return b, nil
}

const b64urlAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
