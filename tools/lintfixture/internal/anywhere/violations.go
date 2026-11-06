// Package anywhere is a lint fixture: it breaks the cryptography and randomness rules on purpose.
package anywhere

import (
	"crypto/sha256"
	"math/rand/v2"
)

// Violations uses a crypto primitive and math/rand outside shared/crypto.
func Violations() ([32]byte, int) {
	return sha256.Sum256(nil), rand.IntN(10)
}
