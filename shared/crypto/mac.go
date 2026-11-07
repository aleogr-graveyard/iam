package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// MinMACKeyBytes is the shortest accepted MAC key (256 bits).
const MinMACKeyBytes = 32

// ErrShortKey reports a key shorter than the required minimum.
var ErrShortKey = errors.New("crypto: key too short")

// MAC computes domain-separated HMAC-SHA-256 tags. It is safe for concurrent use.
type MAC struct {
	key []byte
}

// NewMAC returns a MAC keyed with a copy of key, which must be at least
// MinMACKeyBytes long.
func NewMAC(key []byte) (*MAC, error) {
	if len(key) < MinMACKeyBytes {
		return nil, fmt.Errorf("%w: %d bytes, minimum %d", ErrShortKey, len(key), MinMACKeyBytes)
	}
	return &MAC{key: bytes.Clone(key)}, nil
}

// Sum returns HMAC-SHA-256(key, be64(len(domain)) || domain || msg). The length
// prefix keeps tags of different domains apart even when domain||msg collide.
func (m *MAC) Sum(domain string, msg []byte) []byte {
	h := hmac.New(sha256.New, m.key)
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(domain)))
	h.Write(n[:])
	h.Write([]byte(domain))
	h.Write(msg)
	return h.Sum(nil)
}

// Verify reports, in constant time, whether tag is the tag of msg in domain.
func (m *MAC) Verify(domain string, msg, tag []byte) bool {
	return hmac.Equal(m.Sum(domain, msg), tag)
}
