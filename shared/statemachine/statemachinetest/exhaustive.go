// Package statemachinetest holds test helpers for machines built with the
// statemachine package.
package statemachinetest

import (
	"errors"
	"testing"

	"github.com/aleogr/iam/shared/statemachine"
)

// Expect is one transition the machine is expected to allow.
type Expect[S, E comparable] struct {
	From  S
	Event E
	To    S
}

// Exhaustive fires every (state, event) pair of states x events and compares
// the outcome with expected, which is the complete list of allowed
// transitions. A pair in expected must succeed with the stated To when ctx
// passes the guards; every other pair must fail with ErrUndeclaredTransition.
// So a transition added to the Definition without updating expected, or the
// reverse, fails the test. A malformed expected (duplicate pair, or a pair
// outside states x events) also fails. Every mismatch is reported with Errorf.
func Exhaustive[S, E comparable, C any](t testing.TB, d *statemachine.Definition[S, E, C], states []S, events []E, ctx C, expected []Expect[S, E]) {
	t.Helper()
	want := make(map[pair[S, E]]S, len(expected))
	for _, x := range expected {
		p := pair[S, E]{x.From, x.Event}
		if _, dup := want[p]; dup {
			t.Errorf("%s: expected table lists (%v, %v) twice", d.Name(), x.From, x.Event)
		}
		want[p] = x.To
	}
	seen := 0
	for _, s := range states {
		for _, ev := range events {
			to, declared := want[pair[S, E]{s, ev}]
			if declared {
				seen++
			}
			checkPair(t, d, s, ev, ctx, to, declared)
		}
	}
	if seen != len(want) {
		t.Errorf("%s: expected table has %d pair(s) outside states x events", d.Name(), len(want)-seen)
	}
}

type pair[S, E comparable] struct {
	from S
	ev   E
}

func checkPair[S, E comparable, C any](t testing.TB, d *statemachine.Definition[S, E, C], s S, ev E, ctx C, to S, declared bool) {
	t.Helper()
	got, err := d.Fire(s, ev, ctx)
	switch {
	case declared && err != nil:
		t.Errorf("%s: (%v, %v) should reach %v but failed: %v", d.Name(), s, ev, to, err)
	case declared && got.To != to:
		t.Errorf("%s: (%v, %v) reached %v, want %v", d.Name(), s, ev, got.To, to)
	case !declared && err == nil:
		t.Errorf("%s: (%v, %v) is allowed (to %v) but not in the expected table", d.Name(), s, ev, got.To)
	case !declared && !errors.Is(err, statemachine.ErrUndeclaredTransition):
		t.Errorf("%s: (%v, %v) failed with %v, want ErrUndeclaredTransition", d.Name(), s, ev, err)
	}
}
