package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// maxLine bounds one input line (patterns and paths are short).
const maxLine = 4096

// readPatterns returns the non-blank, non-comment lines of r.
func readPatterns(r io.Reader) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(r)
	sc.Buffer(nil, maxLine)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, sc.Err()
}

// run prints the paths read from in that match the patterns file. Diagnostics go
// to stderr; the exit code is always 0 because the guard must never fail a PR.
func run(patternsFile string, in io.Reader, out io.Writer) error {
	f, err := os.Open(patternsFile) //nolint:gosec // path comes from the workflow, not from user input
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	patterns, err := readPatterns(f)
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(in)
	sc.Buffer(nil, maxLine)
	for sc.Scan() {
		p := strings.TrimSpace(sc.Text())
		if p != "" && Match(patterns, p) {
			if _, err := fmt.Fprintln(out, p); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: sensitivepaths <patterns-file> < changed-paths")
		return
	}
	if err := run(os.Args[1], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "sensitivepaths:", err)
	}
}
