package storagehelpers

import (
	"testing"

	"github.com/aleogr/iam/shared/storage/storagetest"
)

// Test files may import storagetest; no rule must fire here.
func TestUse(t *testing.T) {
	storagetest.Run()
}
