// Package smhelpers is a lint fixture: production code must not import the test-only state-machine helpers.
package smhelpers

import "github.com/aleogr/iam/shared/statemachine/statemachinetest"

// Use imports statemachinetest outside a test file.
func Use() { statemachinetest.Check() }
