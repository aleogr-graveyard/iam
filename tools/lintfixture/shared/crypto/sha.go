// Package crypto is a lint fixture: shared/crypto may import primitives, so no rule must fire here.
package crypto

import "crypto/sha256"

// Sum imports crypto/sha256 inside shared/crypto, which is exempt from the crypto rule.
func Sum(b []byte) [32]byte {
	return sha256.Sum256(b)
}
