package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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
)

// Regressions parses `benchstat -format csv` output and returns every row of a
// gated unit whose delta exceeds maxDelta percent with p < alpha. Any input it
// cannot fully understand is an error: the gate fails closed.
func Regressions(in io.Reader, maxDelta, alpha float64) ([]Regression, error) {
	r := csv.NewReader(in)
	r.FieldsPerRecord = -1
	var (
		out    []Regression
		unit   string
		tables int
	)
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("benchgate: reading csv: %w", err)
		}
		if tables == 0 && !isHeader(rec) {
			continue // preamble such as "goos: linux"
		}
		if isHeader(rec) {
			unit = rec[colUnit]
			tables++
			continue
		}
		reg, found, err := parseRow(rec, unit, maxDelta, alpha)
		if err != nil {
			return nil, err
		}
		if found {
			out = append(out, reg)
		}
	}
	if tables == 0 {
		return nil, errors.New("benchgate: no benchstat table found in input")
	}
	return out, nil
}

func isHeader(rec []string) bool {
	return len(rec) == columns && rec[colName] == "" && rec[colP] == "P"
}

func parseRow(rec []string, unit string, maxDelta, alpha float64) (Regression, bool, error) {
	if len(rec) != columns {
		return Regression{}, false, fmt.Errorf("benchgate: row %q has %d fields, want %d", rec, len(rec), columns)
	}
	name, delta, pcol := rec[colName], rec[colDelta], rec[colP]
	if skipRow(name, unit, delta, pcol) {
		return Regression{}, false, nil
	}
	d, p, err := parseStats(delta, pcol)
	if err != nil {
		return Regression{}, false, fmt.Errorf("benchgate: %s: %w", name, err)
	}
	if d > maxDelta && p < alpha {
		return Regression{Benchmark: name, Unit: unit, DeltaPercent: d, P: p}, true, nil
	}
	return Regression{}, false, nil
}

// skipRow reports rows that cannot be a regression: the geomean summary, units we
// do not gate, "~" (not significant) and benchmarks present on one side only.
func skipRow(name, unit, delta, pcol string) bool {
	return name == "geomean" || !gatedUnits[unit] || delta == "~" || (delta == "" && pcol == "")
}

func parseStats(delta, pcol string) (float64, float64, error) {
	v, ok := strings.CutSuffix(delta, "%")
	if !ok {
		return 0, 0, fmt.Errorf("bad delta %q", delta)
	}
	d, err := strconv.ParseFloat(v, 64)
	if err != nil {
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
	if err != nil {
		return 0, fmt.Errorf("bad p column %q", s)
	}
	return p, nil
}
