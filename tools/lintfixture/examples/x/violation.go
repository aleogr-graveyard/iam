// Package x is a lint fixture: an example client reaching an internal package.
package x

import "github.com/aleogr/iam/internal/database"

// Violation imports an internal package from an example client.
func Violation() string { return database.Name }
