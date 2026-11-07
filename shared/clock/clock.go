// Package clock provides the time source injected into cores and shared domain
// packages, so that no core calls time.Now and tests control time exactly.
package clock

import (
	"sync"
	"time"
)

// Clock returns the current time.
type Clock interface {
	Now() time.Time
}

// System is the real clock. It is the only caller of time.Now outside adapters
// and tests.
type System struct{}

// Now returns the current local time, with its monotonic reading.
func (System) Now() time.Time { return time.Now() }

// Fixed is a clock for tests: it returns T until advanced. Use it through a
// pointer; it is safe for concurrent use.
type Fixed struct {
	T  time.Time
	mu sync.Mutex
}

// Now returns the fixed time.
func (f *Fixed) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.T
}

// Advance moves the clock forward by d (backward when d is negative).
func (f *Fixed) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.T = f.T.Add(d)
}
