// Package x is a lint fixture: an end-to-end test reaching an internal package.
package x

import "github.com/aleogr/iam/internal/database"

// Violation imports an internal package from the e2e module.
func Violation() string { return database.Name }
