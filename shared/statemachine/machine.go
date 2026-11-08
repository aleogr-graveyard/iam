// Package statemachine is a small explicit state-machine library (contract C5).
//
// A Definition is a closed table of transitions. Anything not declared in the
// table is refused, so a flow or lifecycle cannot reach a state by accident.
// The library is pure: it computes the next state and never stores it. Callers
// persist Fired.To with their own version check (optimistic concurrency) and
// emit the audit event for the transition.
//
// The package uses only the standard library.
package statemachine

import (
	"errors"
	"fmt"
)

// ErrUndeclaredTransition is wrapped by the error Fire returns when the
// machine declares no transition for the (state, event) pair.
var ErrUndeclaredTransition = errors.New("statemachine: undeclared transition")

// Transition declares that Event moves the machine from From to To.
//
// Guard, when not nil, is evaluated with the caller's context; a non-nil
// result refuses the transition. A nil Guard means the transition is always
// allowed. Guards must be pure and fast, and must not panic: Fire does not
// recover, so a panic propagates to the caller (the state is untouched because
// Fire never stores it).
type Transition[S, E comparable, C any] struct {
	From  S
	Event E
	To    S
	Guard func(C) error
}

// Fired is the result of an allowed transition.
type Fired[S, E comparable] struct {
	From  S
	Event E
	To    S
}

type key[S, E comparable] struct {
	from S
	ev   E
}

// Definition is an immutable state machine. It is safe for concurrent use.
type Definition[S, E comparable, C any] struct {
	name  string
	ts    []Transition[S, E, C] // declaration order, private copy
	index map[key[S, E]]int
}

// New builds a Definition. It rejects an empty name, an empty transition list
// and a duplicate (From, Event) pair, so each pair has exactly one outcome.
// The transitions are copied: later changes to the slice do not affect the
// Definition.
func New[S, E comparable, C any](name string, ts ...Transition[S, E, C]) (*Definition[S, E, C], error) {
	if name == "" {
		return nil, errors.New("statemachine: name must not be empty")
	}
	if len(ts) == 0 {
		return nil, fmt.Errorf("statemachine: machine %q declares no transitions", name)
	}
	d := &Definition[S, E, C]{
		name:  name,
		ts:    append([]Transition[S, E, C](nil), ts...),
		index: make(map[key[S, E]]int, len(ts)),
	}
	for i, t := range d.ts {
		k := key[S, E]{t.From, t.Event}
		if _, dup := d.index[k]; dup {
			return nil, fmt.Errorf("statemachine: machine %q: duplicate transition from %v on event %v", name, t.From, t.Event)
		}
		d.index[k] = i
	}
	return d, nil
}

// Name returns the machine name.
func (d *Definition[S, E, C]) Name() string { return d.name }

// Fire applies ev to the state from. For an undeclared pair it returns an
// error wrapping ErrUndeclaredTransition that names the machine and the pair.
// If the guard fails, its error is returned wrapped (errors.Is works). In both
// failure cases Fired is the zero value, meaning the state is unchanged.
func (d *Definition[S, E, C]) Fire(from S, ev E, c C) (Fired[S, E], error) {
	i, ok := d.index[key[S, E]{from, ev}]
	if !ok {
		return Fired[S, E]{}, fmt.Errorf("%w: machine %q has no transition from %v on event %v",
			ErrUndeclaredTransition, d.name, from, ev)
	}
	t := d.ts[i]
	if t.Guard != nil {
		if err := t.Guard(c); err != nil {
			return Fired[S, E]{}, fmt.Errorf("statemachine: machine %q: guard refused %v on event %v: %w",
				d.name, from, ev, err)
		}
	}
	return Fired[S, E]{From: from, Event: ev, To: t.To}, nil
}
