package main

import (
	"os"
	"strings"
	"testing"
)

func regressionsFrom(t *testing.T, file string) ([]Regression, error) {
	t.Helper()
	f, err := os.Open("testdata/" + file) //nolint:gosec // fixed fixture names from this file
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	return Regressions(f, 10, 0.05)
}

func TestRegressions_FlagsSignificantSlowdown(t *testing.T) {
	got, err := regressionsFrom(t, "regression.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d regressions, want 1: %+v", len(got), got)
	}
	r := got[0]
	if r.Benchmark != "A-4" || r.Unit != "sec/op" || r.DeltaPercent < 25 || r.DeltaPercent > 26 || r.P >= 0.05 {
		t.Fatalf("unexpected regression: %+v", r)
	}
}

func TestRegressions_IgnoresNoise(t *testing.T) {
	got, err := regressionsFrom(t, "noise.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("noise flagged: %+v", got)
	}
	// benchstat prints "~" for insignificant rows, but a laxer -alpha upstream
	// must not turn a large, insignificant delta into a failure either.
	const lax = ",sec/op,CI,sec/op,CI,vs base,P\nA-4,1,1%,1.15,1%,+15.00%,p=0.300 n=10\n"
	got, err = Regressions(strings.NewReader(lax), 10, 0.05)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestRegressions_IgnoresImprovement(t *testing.T) {
	got, err := regressionsFrom(t, "improvement.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("improvement flagged: %+v", got)
	}
}

func TestRegressions_MalformedCSV(t *testing.T) {
	for _, in := range []string{
		"",
		"garbage",
		"\"unterminated,quote\n",
		",sec/op,CI,sec/op,CI,vs base,P\nA-4,1,1%,2,1%,+nope%,p=0.001 n=10\n",
		",sec/op,CI,sec/op,CI,vs base,P\nA-4,1,1%,2,1%,+25.00%,garbage\n",
		",sec/op,CI,sec/op,CI,vs base,P\nA-4,1,2\n",
	} {
		if _, err := Regressions(strings.NewReader(in), 10, 0.05); err == nil {
			t.Errorf("input %q: expected an error", in)
		}
	}
}
