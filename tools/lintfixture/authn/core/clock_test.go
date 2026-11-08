package core

import (
	"testing"
	"time"
)

// Core tests must use an injected clock too: the root-context exclusion for
// _test.go files must not hide this finding.
func TestClock(t *testing.T) {
	_ = time.Now()
}
