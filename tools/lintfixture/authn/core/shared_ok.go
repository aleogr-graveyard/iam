package core

import "github.com/aleogr/iam/shared/crypto"

// SharedOK imports a shared package, which the core allow-list permits.
func SharedOK() [32]byte { return crypto.Sum(nil) }
