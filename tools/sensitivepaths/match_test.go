package main

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	patterns := []string{"authn/**", "authz/core/**", "shared/crypto/**", ".github/**", "docs/brief.md", "CLAUDE.md", "Makefile", "a/**/b.go"}
	cases := []struct {
		path string
		want bool
	}{
		{"authn/core/x.go", true},
		{"authz/core/a/b.go", true},
		{"authz/sdk/x.go", false},
		{"docs/brief.md", true},
		{"docs/roadmap.md", false},
		{".github/workflows/ci.yml", true},
		{"shared/cryptography/x.go", false},
		{"shared/crypto", false},
		{"CLAUDE.md", true},
		{"sub/CLAUDE.md", false},
		{"Makefile", true},
		{"Makefile2", false},
		{"a/b.go", true},
		{"a/x/y/b.go", true},
		{"a/x/c.go", false},
		{"", false},
	}
	for _, c := range cases {
		if got := Match(patterns, c.path); got != c.want {
			t.Errorf("Match(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestRunFiltersStdin(t *testing.T) {
	var out strings.Builder
	in := strings.NewReader("authn/core/x.go\ndocs/roadmap.md\n\n")
	if err := run("../../.github/sensitive-paths.txt", in, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "authn/core/x.go\n" {
		t.Fatalf("got %q", out.String())
	}
}

func fileLines(t *testing.T, name string, transform func(string) string) []string {
	t.Helper()
	b, err := os.ReadFile(name) //nolint:gosec // fixed test paths
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		out = append(out, transform(l))
	}
	slices.Sort(out)
	return out
}

// TestCodeownersInSync keeps CODEOWNERS and sensitive-paths.txt listing the same
// patterns, all owned by @aleogr.
func TestCodeownersInSync(t *testing.T) {
	want := fileLines(t, "../../.github/sensitive-paths.txt", func(s string) string { return s })
	got := fileLines(t, "../../.github/CODEOWNERS", func(s string) string {
		pat, owner, ok := strings.Cut(s, " ")
		if !ok || strings.TrimSpace(owner) != "@aleogr" {
			t.Errorf("CODEOWNERS line %q: want '<pattern> @aleogr'", s)
		}
		return strings.TrimPrefix(pat, "/")
	})
	if !slices.Equal(want, got) {
		t.Fatalf("CODEOWNERS patterns differ from sensitive-paths.txt:\n%v\n%v", got, want)
	}
}
