package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// maxLine bounds one input line; a longer line is an error, never silently cut.
const maxLine = 64 * 1024

// validatePattern rejects patterns that could silently never match: empty or
// slash-edged segments (a trailing slash is not supported) and malformed globs.
func validatePattern(p string) error {
	for _, seg := range strings.Split(p, "/") {
		if seg == "" {
			return fmt.Errorf("pattern %q: empty segment (leading, trailing or doubled slash)", p)
		}
		if _, err := path.Match(seg, ""); err != nil {
			return fmt.Errorf("pattern %q: %w", p, err)
		}
	}
	return nil
}

// readPatterns returns the validated, non-blank, non-comment lines of r. An
// empty result is an error: a guard without patterns protects nothing.
func readPatterns(r io.Reader) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(r)
	sc.Buffer(nil, maxLine)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := validatePattern(line); err != nil {
			return nil, err
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("no patterns found")
	}
	return out, nil
}

// run prints the paths read from in that match the patterns file. Any error is
// returned so that the guard fails closed.
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
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "sensitivepaths:", err)
		os.Exit(1)
	}
}
