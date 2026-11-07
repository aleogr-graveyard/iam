package crypto

import "crypto/subtle"

// ConstantTimeEqual reports whether a and b are equal, in time that depends
// only on their lengths, never on their contents. Inputs of different lengths
// are unequal; their lengths are not treated as secret.
func ConstantTimeEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
