// Command benchgate reads `benchstat -format csv` output and exits 1 when a
// benchmark regressed significantly. Any error also exits non-zero.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("benchgate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	maxDelta := fs.Float64("max-delta", 10, "largest tolerated slowdown, in percent")
	alpha := fs.Float64("alpha", 0.05, "significance level")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	in := stdin
	if fs.NArg() > 1 {
		_, _ = fmt.Fprintln(stderr, "usage: benchgate [flags] [file.csv]")
		return 2
	}
	if fs.NArg() == 1 {
		f, err := os.Open(fs.Arg(0))
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "benchgate:", err)
			return 2
		}
		defer func() { _ = f.Close() }()
		in = f
	}
	regs, err := Regressions(in, *maxDelta, *alpha)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	for _, r := range regs {
		_, _ = fmt.Fprintf(stderr, "REGRESSION %s %s %+.2f%% (p=%.3f)\n", r.Benchmark, r.Unit, r.DeltaPercent, r.P)
	}
	if len(regs) > 0 {
		return 1
	}
	_, _ = fmt.Fprintln(stdout, "benchgate: no significant regressions")
	return 0
}
