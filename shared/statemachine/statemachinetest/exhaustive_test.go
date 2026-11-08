package statemachinetest_test

import (
	"errors"
	"testing"

	"github.com/aleogr/iam/shared/statemachine"
	"github.com/aleogr/iam/shared/statemachine/statemachinetest"
)

type recorder struct {
	testing.TB
	errs []string
}

func (r *recorder) Helper() {}
func (r *recorder) Errorf(format string, args ...any) {
	r.errs = append(r.errs, format)
}

type c struct{ ok bool }

func def(t *testing.T, extra ...statemachine.Transition[string, string, c]) *statemachine.Definition[string, string, c] {
	t.Helper()
	ts := append([]statemachine.Transition[string, string, c]{
		{From: "a", Event: "go", To: "b"},
		{From: "b", Event: "go", To: "c", Guard: func(x c) error {
			if x.ok {
				return nil
			}
			return errors.New("no")
		}},
	}, extra...)
	d, err := statemachine.New("m", ts...)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

var (
	states = []string{"a", "b", "c"}
	events = []string{"go", "stop"}
	table  = []statemachine.Fired[string, string]{{From: "a", Event: "go", To: "b"}, {From: "b", Event: "go", To: "c"}}
)

func TestExhaustive_PassesOnMatchingTable(t *testing.T) {
	statemachinetest.Exhaustive(t, def(t), states, events, c{ok: true}, table)
}

func run(t *testing.T, d *statemachine.Definition[string, string, c], exp []statemachine.Fired[string, string]) *recorder {
	t.Helper()
	r := &recorder{TB: t}
	statemachinetest.Exhaustive[string, string, c](r, d, states, events, c{ok: true}, exp)
	return r
}

func TestExhaustive_CatchesExtraTransition(t *testing.T) {
	d := def(t, statemachine.Transition[string, string, c]{From: "c", Event: "stop", To: "a"})
	if r := run(t, d, table); len(r.errs) == 0 {
		t.Fatalf("want a failure, got %v", r.errs)
	}
}

func TestExhaustive_CatchesMissingAndWrongTarget(t *testing.T) {
	missing := []statemachine.Fired[string, string]{{From: "a", Event: "go", To: "b"}, {From: "b", Event: "go", To: "c"}, {From: "c", Event: "go", To: "a"}}
	if r := run(t, def(t), missing); len(r.errs) == 0 {
		t.Fatalf("missing: %v", r.errs)
	}
	wrong := []statemachine.Fired[string, string]{{From: "a", Event: "go", To: "c"}, {From: "b", Event: "go", To: "c"}}
	if r := run(t, def(t), wrong); len(r.errs) == 0 {
		t.Fatalf("wrong target: %v", r.errs)
	}
}

func TestExhaustive_RejectsMalformedTable(t *testing.T) {
	dup := append(append([]statemachine.Fired[string, string]{}, table...), table[0])
	if r := run(t, def(t), dup); len(r.errs) == 0 {
		t.Fatal("duplicate expectation must fail")
	}
	out := append(append([]statemachine.Fired[string, string]{}, table...), statemachine.Fired[string, string]{From: "zz", Event: "go", To: "a"})
	if r := run(t, def(t), out); len(r.errs) == 0 {
		t.Fatal("expectation outside states/events must fail")
	}
}

func runWith(t *testing.T, d *statemachine.Definition[string, string, c], st, ev []string, exp []statemachine.Fired[string, string]) *recorder {
	t.Helper()
	r := &recorder{TB: t}
	statemachinetest.Exhaustive[string, string, c](r, d, st, ev, c{ok: true}, exp)
	return r
}

func TestExhaustive_CatchesEventMissingFromEventsList(t *testing.T) {
	d := def(t, statemachine.Transition[string, string, c]{From: "a", Event: "bypass", To: "c"})
	exp := append(append([]statemachine.Fired[string, string]{}, table...), statemachine.Fired[string, string]{From: "a", Event: "bypass", To: "c"})
	if r := runWith(t, d, states, events, exp); len(r.errs) == 0 {
		t.Fatal("event outside events must fail")
	}
}

func TestExhaustive_CatchesStateMissingFromStatesList(t *testing.T) {
	if r := runWith(t, def(t), []string{"a", "b"}, events, table); len(r.errs) == 0 {
		t.Fatal("To outside states must fail")
	}
}

func TestExhaustive_CatchesDeclaredTransitionMissingFromExpected(t *testing.T) {
	// Declared transition whose pair lies outside states x events and is also unexpected.
	d := def(t, statemachine.Transition[string, string, c]{From: "x", Event: "y", To: "a"})
	if r := runWith(t, d, states, events, table); len(r.errs) == 0 {
		t.Fatal("unexpected declared transition must fail")
	}
}

func TestExhaustive_RejectsEmptyOrDuplicateLists(t *testing.T) {
	if r := runWith(t, def(t), nil, nil, nil); len(r.errs) == 0 {
		t.Fatal("empty lists must fail")
	}
	if r := runWith(t, def(t), states, nil, table); len(r.errs) == 0 {
		t.Fatal("empty events must fail")
	}
	if r := runWith(t, def(t), nil, events, table); len(r.errs) == 0 {
		t.Fatal("empty states must fail")
	}
	if r := runWith(t, def(t), []string{"a", "a", "b", "c"}, events, table); len(r.errs) == 0 {
		t.Fatal("duplicate state must fail")
	}
	if r := runWith(t, def(t), states, []string{"go", "go", "stop"}, table); len(r.errs) == 0 {
		t.Fatal("duplicate event must fail")
	}
}

func TestExhaustive_StaleExpectationNotMaskedByDuplicates(t *testing.T) {
	stale := append(append([]statemachine.Fired[string, string]{}, table...), statemachine.Fired[string, string]{From: "a", Event: "stop", To: "b"})
	if r := runWith(t, def(t), []string{"a", "a", "b", "c"}, events, stale); len(r.errs) == 0 {
		t.Fatal("must fail")
	}
}
