package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Regression is one benchmark/unit pair that got significantly worse.
type Regression struct {
	Benchmark    string
	Unit         string
	DeltaPercent float64
	P            float64
}

// gatedUnits are the units whose growth counts as a regression.
var gatedUnits = map[string]bool{"sec/op": true, "B/op": true, "allocs/op": true}

// benchstat -format csv layout: a table header ",<unit>,CI,<unit>,CI,vs base,P"
// followed by one row per benchmark and a trailing geomean row.
const (
	columns  = 7
	colName  = 0
	colUnit  = 1
	colDelta = 5
	colP     = 6

	// benchstat leaves one-sided rows unpadded: name plus base value and CI, or
	// name, two empty fields, then head value and CI.
	oneSidedBase = 3
	oneSidedHead = 5
)

// Regressions parses `benchstat -format csv` output and returns every row of a
// gated unit whose delta exceeds maxDelta percent with p < alpha. Any input it
// cannot fully understand is an error, and so is input where no gated row was
// actually compared: the gate fails closed.
func Regressions(in io.Reader, maxDelta, alpha float64) ([]Regression, error) {
	r := csv.NewReader(in)
	r.FieldsPerRecord = -1
	var (
		out      []Regression
		unit     string
		tables   int
		compared int
	)
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("benchgate: reading csv: %w", err)
		}
		switch {
		case len(rec) == 1:
			continue // config line between tables, e.g. "pkg: x" or "goos: linux"
		case isHeader(rec):
			unit = rec[colUnit]
			tables++
			continue
		case len(rec) == columns && rec[colName] == "":
			continue // the row naming the base and head input files
		case tables == 0:
			return nil, fmt.Errorf("benchgate: unexpected record %q before any benchstat table", rec)
		case len(rec) == oneSidedBase || len(rec) == oneSidedHead:
			continue // benchmark present on one side only: nothing to compare
		}
		reg, found, counted, err := parseRow(rec, unit, maxDelta, alpha)
		if err != nil {
			return nil, err
		}
		if counted {
			compared++
		}
		if found {
			out = append(out, reg)
		}
	}
	if tables == 0 {
		return nil, errors.New("benchgate: no benchstat table found in input")
	}
	if compared == 0 {
		return nil, errors.New("benchgate: no benchmark was compared (nothing on both sides)")
	}
	return out, nil
}

func isHeader(rec []string) bool {
	return len(rec) == columns && rec[colName] == "" && rec[colP] == "P"
}

// parseRow returns the regression (if any) and whether the row was a real
// comparison of a gated unit. The "~" benchstat prints for insignificant rows
// already applies its own alpha of 0.05; p is still checked against ours.
func parseRow(rec []string, unit string, maxDelta, alpha float64) (Regression, bool, bool, error) {
	if len(rec) != columns {
		return Regression{}, false, false, fmt.Errorf("benchgate: row %q has %d fields, want %d", rec, len(rec), columns)
	}
	name, delta, pcol := rec[colName], rec[colDelta], rec[colP]
	if name == "geomean" || !gatedUnits[unit] {
		return Regression{}, false, false, nil
	}
	switch delta {
	case "?":
		return Regression{}, false, false, nil // benchstat could not compare the row
	case "~":
		return Regression{}, false, true, nil
	}
	d, p, err := parseStats(delta, pcol)
	if err != nil {
		return Regression{}, false, false, fmt.Errorf("benchgate: %s: %w", name, err)
	}
	if d > maxDelta && p < alpha {
		return Regression{Benchmark: name, Unit: unit, DeltaPercent: d, P: p}, true, true, nil
	}
	return Regression{}, false, true, nil
}

func parseStats(delta, pcol string) (float64, float64, error) {
	v, ok := strings.CutSuffix(delta, "%")
	if !ok {
		return 0, 0, fmt.Errorf("bad delta %q", delta)
	}
	d, err := strconv.ParseFloat(v, 64)
	if err != nil || math.IsNaN(d) {
		return 0, 0, fmt.Errorf("bad delta %q", delta)
	}
	p, err := parseP(pcol)
	return d, p, err
}

// parseP reads "p=0.001 n=10".
func parseP(s string) (float64, error) {
	field, _, _ := strings.Cut(s, " ")
	v, ok := strings.CutPrefix(field, "p=")
	if !ok {
		return 0, fmt.Errorf("bad p column %q", s)
	}
	p, err := strconv.ParseFloat(v, 64)
	if err != nil || math.IsNaN(p) {
		return 0, fmt.Errorf("bad p column %q", s)
	}
	return p, nil
}
