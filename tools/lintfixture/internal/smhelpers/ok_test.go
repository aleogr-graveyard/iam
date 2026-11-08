package smhelpers

import (
	"testing"

	"github.com/aleogr/iam/shared/statemachine/statemachinetest"
)

// Test files may import statemachinetest; no rule must fire here.
func TestUse(t *testing.T) {
	statemachinetest.Check()
}
