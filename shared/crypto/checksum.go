package crypto

import "crypto/sha256"

// ChecksumBytes is the length of a Checksum.
const ChecksumBytes = sha256.Size

// Checksum returns SHA-256(data), an integrity fingerprint of non-secret
// content, such as a migration file, to detect that it changed. It is not
// for secrets (HashSecret) or passwords (PasswordHasher), and compares with
// plain equality: the content is public.
func Checksum(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}
