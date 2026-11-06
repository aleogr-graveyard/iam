// Package main implements the auto-merge guard's path matcher: it reports which
// changed paths fall under the sensitive patterns of .github/sensitive-paths.txt.
package main

import (
	"path"
	"strings"
)

// Match reports whether name matches at least one pattern. Patterns are
// gitignore-style globs over slash-separated paths: "**" as a whole segment
// matches one or more segments when it ends the pattern and zero or more
// otherwise; other segments use path.Match. There is no prefix matching, so
// "shared/crypto/**" does not match "shared/cryptography/x.go".
func Match(patterns []string, name string) bool {
	segs := strings.Split(name, "/")
	for _, p := range patterns {
		if matchSegments(strings.Split(p, "/"), segs) {
			return true
		}
	}
	return false
}

func matchSegments(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			if len(rest) == 0 {
				return len(segs) > 0
			}
			for i := 0; i <= len(segs); i++ {
				if matchSegments(rest, segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		ok, err := path.Match(pat[0], segs[0])
		if err != nil || !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}
