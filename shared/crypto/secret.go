package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
)

// MinSecretBytes is the shortest secret HashSecret accepts (128 bits). Client
// secrets are generated as 32 random bytes; the floor refuses accidental
// low-entropy input such as an empty or truncated value.
const MinSecretBytes = 16

// SecretHashBytes is the length of a hash produced by HashSecret.
const SecretHashBytes = sha256.Size

// ErrShortSecret reports a secret shorter than MinSecretBytes.
var ErrShortSecret = errors.New("crypto: secret too short")

// HashSecret returns SHA-256(secret), the stored form of a high-entropy,
// machine-generated secret such as an OAuth client secret (32 random bytes).
//
// A single unsalted fast hash is enough only because the input already carries
// at least 128 bits of entropy: guessing it is infeasible however fast the hash
// is, and a salt would add nothing. It must NEVER be used for passwords or any
// other human-chosen value; those go through PasswordHasher (Argon2id).
//
// Secrets shorter than MinSecretBytes are refused with ErrShortSecret.
func HashSecret(secret []byte) ([]byte, error) {
	if len(secret) < MinSecretBytes {
		return nil, fmt.Errorf("%w: %d bytes, minimum %d", ErrShortSecret, len(secret), MinSecretBytes)
	}
	sum := sha256.Sum256(secret)
	return sum[:], nil
}

// VerifySecret reports whether hash is HashSecret(secret). The comparison of
// the hashes is constant time. A secret shorter than MinSecretBytes (including
// an empty one) or a hash that is not SecretHashBytes long is never a match;
// these lengths are not treated as secret.
func VerifySecret(secret, hash []byte) bool {
	if len(secret) < MinSecretBytes || len(hash) != SecretHashBytes {
		return false
	}
	sum := sha256.Sum256(secret)
	return subtle.ConstantTimeCompare(sum[:], hash) == 1
}
