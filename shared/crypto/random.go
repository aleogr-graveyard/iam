package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// MinTokenBytes is the smallest entropy, in bytes (128 bits), that RandomToken
// accepts.
const MinTokenBytes = 16

// ErrInvalidLength reports a requested length that is not allowed.
var ErrInvalidLength = errors.New("crypto: invalid length")

// RandomBytes returns n bytes from the operating system's CSPRNG.
func RandomBytes(n int) ([]byte, error) {
	if n < 1 {
		return nil, fmt.Errorf("%w: %d random bytes", ErrInvalidLength, n)
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("crypto: reading randomness: %w", err)
	}
	return b, nil
}

// RandomToken returns nBytes of randomness encoded as unpadded base64url.
// nBytes must be at least MinTokenBytes.
func RandomToken(nBytes int) (string, error) {
	if nBytes < MinTokenBytes {
		return "", fmt.Errorf("%w: token of %d bytes, minimum %d", ErrInvalidLength, nBytes, MinTokenBytes)
	}
	b, err := RandomBytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
