// Package storagehelpers is a lint fixture: production code must not import the storage contract suite.
package storagehelpers

import "github.com/aleogr/iam/shared/storage/storagetest"

// Use imports storagetest outside a test file.
func Use() { storagetest.Run() }
