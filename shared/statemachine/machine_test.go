package statemachine_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/aleogr/iam/shared/statemachine"
)

type state string
type event string

type ctx struct{ ok bool }

var errDenied = errors.New("denied")

func guard(c ctx) error {
	if !c.ok {
		return errDenied
	}
	return nil
}

func tr(from state, ev event, to state) statemachine.Transition[state, event, ctx] {
	return statemachine.Transition[state, event, ctx]{From: from, Event: ev, To: to}
}

func newDef(t *testing.T) *statemachine.Definition[state, event, ctx] {
	t.Helper()
	d, err := statemachine.New("login",
		tr("start", "submit", "verify"),
		statemachine.Transition[state, event, ctx]{From: "verify", Event: "pass", To: "done", Guard: guard},
		tr("verify", "fail", "start"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestFire_DeclaredTransition(t *testing.T) {
	d := newDef(t)
	got, err := d.Fire("start", "submit", ctx{})
	if err != nil {
		t.Fatal(err)
	}
	want := statemachine.Fired[state, event]{From: "start", Event: "submit", To: "verify"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if got, err = d.Fire("verify", "pass", ctx{ok: true}); err != nil || got.To != "done" {
		t.Fatalf("guarded transition: %+v, %v", got, err)
	}
}

func TestFire_UndeclaredIsRefused(t *testing.T) {
	d := newDef(t)
	got, err := d.Fire("done", "submit", ctx{ok: true})
	if !errors.Is(err, statemachine.ErrUndeclaredTransition) {
		t.Fatalf("err = %v", err)
	}
	for _, part := range []string{"login", "done", "submit"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error %q does not name %q", err, part)
		}
	}
	if got != (statemachine.Fired[state, event]{}) {
		t.Fatalf("Fired must be zero, got %+v", got)
	}
}

func TestFire_GuardErrorIsReturnedAndStateUnchanged(t *testing.T) {
	d := newDef(t)
	got, err := d.Fire("verify", "pass", ctx{ok: false})
	if !errors.Is(err, errDenied) {
		t.Fatalf("err = %v", err)
	}
	if errors.Is(err, statemachine.ErrUndeclaredTransition) {
		t.Fatal("a guard error is not an undeclared transition")
	}
	if got != (statemachine.Fired[state, event]{}) {
		t.Fatalf("Fired must be zero, got %+v", got)
	}
}

func TestNew_Rejects(t *testing.T) {
	cases := map[string]struct {
		name string
		ts   []statemachine.Transition[state, event, ctx]
	}{
		"duplicate": {"m", []statemachine.Transition[state, event, ctx]{tr("a", "e", "b"), tr("a", "e", "c")}},
		"no name":   {"", []statemachine.Transition[state, event, ctx]{tr("a", "e", "b")}},
		"empty":     {"m", nil},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			d, err := statemachine.New(c.name, c.ts...)
			if err == nil || d != nil {
				t.Fatalf("want error and nil definition, got %v, %v", d, err)
			}
		})
	}
}

func TestNew_DuplicateTransitionNamesThePair(t *testing.T) {
	_, err := statemachine.New("m", tr("a", "e", "b"), tr("a", "e", "b"))
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("err = %v", err)
	}
}

func TestNew_CopiesTransitions(t *testing.T) {
	ts := []statemachine.Transition[state, event, ctx]{tr("a", "e", "b")}
	d, err := statemachine.New("m", ts...)
	if err != nil {
		t.Fatal(err)
	}
	ts[0].To = "evil"
	ts[0].Guard = guard
	got, err := d.Fire("a", "e", ctx{})
	if err != nil || got.To != "b" {
		t.Fatalf("definition changed after New: %+v, %v", got, err)
	}
}

func TestFire_NilGuardAlwaysAllows(t *testing.T) {
	d := newDef(t)
	if _, err := d.Fire("start", "submit", ctx{ok: false}); err != nil {
		t.Fatal(err)
	}
}

func TestFire_ConcurrentIsSafe(t *testing.T) {
	d := newDef(t)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if _, err := d.Fire("start", "submit", ctx{}); err != nil {
					t.Error(err)
				}
				_, _ = d.Fire("verify", "pass", ctx{})
				_ = d.Mermaid()
			}
		}()
	}
	wg.Wait()
}

func TestMermaid_IsDeterministic(t *testing.T) {
	want := `stateDiagram-v2
    state "start" as s0
    state "verify" as s1
    state "done" as s2
    s0 --> s1 : submit
    s1 --> s2 : pass
    s1 --> s0 : fail
`
	d := newDef(t)
	for i := 0; i < 20; i++ {
		if got := d.Mermaid(); got != want {
			t.Fatalf("got:\n%s\nwant:\n%s", got, want)
		}
	}
}

func TestMermaid_SanitizesLabels(t *testing.T) {
	hostile := state("a\nb : c --> \"d\"; e")
	d, err := statemachine.New("m", statemachine.Transition[state, event, ctx]{From: hostile, Event: "x\ny:z-->", To: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	out := d.Mermaid()
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 1+2+1 {
		t.Fatalf("hostile label broke the line structure:\n%s", out)
	}
	for _, l := range lines[1:] {
		if strings.Count(l, "-->") > 1 || strings.Count(l, `"`) > 2 || strings.Count(l, " : ") > 1 {
			t.Errorf("unsafe line %q", l)
		}
	}
	want := "    state \"a b #58; c --#gt; #quot;d#quot;#59; e\" as s0\n"
	if !strings.Contains(out, want) {
		t.Errorf("missing %q in\n%s", want, out)
	}
}

func TestTransitions_ReturnsCopyInDeclarationOrder(t *testing.T) {
	d := newDef(t)
	got := d.Transitions()
	want := []statemachine.Fired[state, event]{
		{From: "start", Event: "submit", To: "verify"},
		{From: "verify", Event: "pass", To: "done"},
		{From: "verify", Event: "fail", To: "start"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	got[0].To = "evil"
	if d.Transitions()[0].To != "verify" {
		t.Fatal("Transitions must return a copy")
	}
}

type code int

func TestNew_AcceptsIntEnums(t *testing.T) {
	d, err := statemachine.New("n", statemachine.Transition[code, code, struct{}]{From: 1, Event: 2, To: 3})
	if err != nil {
		t.Fatal(err)
	}
	if f, err := d.Fire(1, 2, struct{}{}); err != nil || f.To != 3 {
		t.Fatalf("%+v %v", f, err)
	}
}

func TestMermaid_EmptyLabelsAndAmpersand(t *testing.T) {
	d, err := statemachine.New("m", statemachine.Transition[state, event, ctx]{From: "", Event: "", To: "a&b"})
	if err != nil {
		t.Fatal(err)
	}
	out := d.Mermaid()
	for _, want := range []string{`state "(empty)" as s0`, `state "a#38;b" as s1`, `s0 --> s1 : (empty)`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if strings.Contains(out, `""`) {
		t.Errorf("empty quoted label in\n%s", out)
	}
}

func Example() {
	d, err := statemachine.New("door",
		statemachine.Transition[string, string, struct{}]{From: "closed", Event: "open", To: "opened"},
		statemachine.Transition[string, string, struct{}]{From: "opened", Event: "close", To: "closed"},
	)
	if err != nil {
		panic(err)
	}
	f, _ := d.Fire("closed", "open", struct{}{})
	fmt.Println(f.From, "->", f.To)
	_, err = d.Fire("closed", "close", struct{}{})
	fmt.Println(errors.Is(err, statemachine.ErrUndeclaredTransition))
	fmt.Print(d.Mermaid())
	// Output:
	// closed -> opened
	// true
	// stateDiagram-v2
	//     state "closed" as s0
	//     state "opened" as s1
	//     s0 --> s1 : open
	//     s1 --> s0 : close
}
