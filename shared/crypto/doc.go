// Package crypto is the single cryptography entry point of the platform
// (contract C2): randomness, message authentication, constant-time comparison
// and password hashing. No other package imports crypto/* or golang.org/x/crypto.
//
// Errors never contain passwords, keys, tags or hashes.
package crypto
