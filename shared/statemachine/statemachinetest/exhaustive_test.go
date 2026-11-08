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
	table  = []statemachinetest.Expect[string, string]{{From: "a", Event: "go", To: "b"}, {From: "b", Event: "go", To: "c"}}
)

func TestExhaustive_PassesOnMatchingTable(t *testing.T) {
	statemachinetest.Exhaustive(t, def(t), states, events, c{ok: true}, table)
}

func run(t *testing.T, d *statemachine.Definition[string, string, c], exp []statemachinetest.Expect[string, string]) *recorder {
	t.Helper()
	r := &recorder{TB: t}
	statemachinetest.Exhaustive[string, string, c](r, d, states, events, c{ok: true}, exp)
	return r
}

func TestExhaustive_CatchesExtraTransition(t *testing.T) {
	d := def(t, statemachine.Transition[string, string, c]{From: "c", Event: "stop", To: "a"})
	if r := run(t, d, table); len(r.errs) != 1 {
		t.Fatalf("want exactly one failure, got %v", r.errs)
	}
}

func TestExhaustive_CatchesMissingAndWrongTarget(t *testing.T) {
	missing := []statemachinetest.Expect[string, string]{{From: "a", Event: "go", To: "b"}, {From: "b", Event: "go", To: "c"}, {From: "c", Event: "go", To: "a"}}
	if r := run(t, def(t), missing); len(r.errs) != 1 {
		t.Fatalf("missing: %v", r.errs)
	}
	wrong := []statemachinetest.Expect[string, string]{{From: "a", Event: "go", To: "c"}, {From: "b", Event: "go", To: "c"}}
	if r := run(t, def(t), wrong); len(r.errs) != 1 {
		t.Fatalf("wrong target: %v", r.errs)
	}
}

func TestExhaustive_RejectsMalformedTable(t *testing.T) {
	dup := append(append([]statemachinetest.Expect[string, string]{}, table...), table[0])
	if r := run(t, def(t), dup); len(r.errs) == 0 {
		t.Fatal("duplicate expectation must fail")
	}
	out := append(append([]statemachinetest.Expect[string, string]{}, table...), statemachinetest.Expect[string, string]{From: "zz", Event: "go", To: "a"})
	if r := run(t, def(t), out); len(r.errs) == 0 {
		t.Fatal("expectation outside states/events must fail")
	}
}
