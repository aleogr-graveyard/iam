// Package statemachinetest holds test helpers for machines built with the
// statemachine package. Production code must not import it (depguard).
package statemachinetest

import (
	"errors"
	"testing"

	"github.com/aleogr/iam/shared/statemachine"
)

// Exhaustive checks a Definition against the complete list of transitions the
// caller expects (a statemachine.Fired per allowed transition). It fails when:
//
//   - states or events is empty or has duplicates;
//   - a declared transition uses a state or event outside states/events, or is
//     missing from expected (or differs in To);
//   - expected has a duplicate, or a pair outside states x events;
//   - firing any (state, event) pair of states x events gives a different result
//     than expected: an expected pair must succeed with the right To when ctx
//     passes the guards, every other pair must fail with ErrUndeclaredTransition.
//
// Every mismatch is reported with Errorf.
func Exhaustive[S, E statemachine.Enum, C any](t testing.TB, d *statemachine.Definition[S, E, C], states []S, events []E, ctx C, expected []statemachine.Fired[S, E]) {
	t.Helper()
	name := d.Name()
	if len(states) == 0 || len(events) == 0 {
		t.Errorf("%s: states and events must not be empty", name)
		return
	}
	stateSet := toSet(t, name, "state", states)
	eventSet := toSet(t, name, "event", events)
	want := expectedTable(t, name, expected)
	checkDeclared(t, d, stateSet, eventSet, want)

	seen := 0
	for _, s := range states {
		for _, ev := range events {
			k := statemachine.Fired[S, E]{From: s, Event: ev}
			if _, ok := want[k]; ok {
				seen++
			}
		}
	}
	// states and events are duplicate-free, so seen counts distinct pairs.
	if seen != len(want) {
		t.Errorf("%s: expected table has %d pair(s) outside states x events", name, len(want)-seen)
	}
	for _, s := range states {
		for _, ev := range events {
			k := statemachine.Fired[S, E]{From: s, Event: ev}
			to, declared := want[k]
			checkPair(t, d, k, ctx, to, declared)
		}
	}
}

func toSet[T comparable](t testing.TB, name, what string, xs []T) map[T]struct{} {
	t.Helper()
	set := make(map[T]struct{}, len(xs))
	for _, x := range xs {
		if _, dup := set[x]; dup {
			t.Errorf("%s: %s %v listed twice", name, what, x)
		}
		set[x] = struct{}{}
	}
	return set
}

func expectedTable[S, E statemachine.Enum](t testing.TB, name string, expected []statemachine.Fired[S, E]) map[statemachine.Fired[S, E]]S {
	t.Helper()
	want := make(map[statemachine.Fired[S, E]]S, len(expected))
	for _, x := range expected {
		k := statemachine.Fired[S, E]{From: x.From, Event: x.Event}
		if _, dup := want[k]; dup {
			t.Errorf("%s: expected table lists (%v, %v) twice", name, x.From, x.Event)
		}
		want[k] = x.To
	}
	return want
}

func checkDeclared[S, E statemachine.Enum, C any](t testing.TB, d *statemachine.Definition[S, E, C], states map[S]struct{}, events map[E]struct{}, want map[statemachine.Fired[S, E]]S) {
	t.Helper()
	for _, tr := range d.Transitions() {
		_, fromOK := states[tr.From]
		_, toOK := states[tr.To]
		_, evOK := events[tr.Event]
		if !fromOK || !toOK || !evOK {
			t.Errorf("%s: declared (%v, %v) -> %v uses a state or event missing from the lists", d.Name(), tr.From, tr.Event, tr.To)
		}
		to, ok := want[statemachine.Fired[S, E]{From: tr.From, Event: tr.Event}]
		if !ok || to != tr.To {
			t.Errorf("%s: declared (%v, %v) -> %v is missing from the expected table", d.Name(), tr.From, tr.Event, tr.To)
		}
	}
}

func checkPair[S, E statemachine.Enum, C any](t testing.TB, d *statemachine.Definition[S, E, C], k statemachine.Fired[S, E], ctx C, to S, declared bool) {
	t.Helper()
	got, err := d.Fire(k.From, k.Event, ctx)
	switch {
	case declared && err != nil:
		t.Errorf("%s: (%v, %v) should reach %v but failed: %v", d.Name(), k.From, k.Event, to, err)
	case declared && got.To != to:
		t.Errorf("%s: (%v, %v) reached %v, want %v", d.Name(), k.From, k.Event, got.To, to)
	case !declared && err == nil:
		t.Errorf("%s: (%v, %v) is allowed (to %v) but not in the expected table", d.Name(), k.From, k.Event, got.To)
	case !declared && !errors.Is(err, statemachine.ErrUndeclaredTransition):
		t.Errorf("%s: (%v, %v) failed with %v, want ErrUndeclaredTransition", d.Name(), k.From, k.Event, err)
	}
}
