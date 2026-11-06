#!/usr/bin/env bash
# Fails unless every go.mod in the repository is either a module listed in
# go.work (and so tested and linted by make test / make lint) or one of the
# explicit exclusions below. A new module left out of go.work would otherwise
# escape lint and tests without anyone noticing.
#
# Exclusions:
#   tools/go.mod          the repo's own tool code; make test and make lint
#                         handle it with GOWORK=off.
#   tools/lintfixture     the lint fixture; it must fail lint on purpose and is
#                         run by make lint-rules.
#   tools/<tool>/go.mod   one modfile per external tool. They must hold no Go
#                         code (only go.mod and go.sum), so nothing hides there.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

# Captured first so that a broken go.work fails the script (set -e).
listed="$(go list -m -f '{{.Dir}}')"
if [ -z "$listed" ]; then
	echo "FAIL: go.work lists no modules" >&2
	exit 1
fi
mods="$(find . -path ./.git -prune -o -name go.mod -type f -print | sed 's|^\./||' | sort)"

bad=0
while IFS= read -r mod; do
	dir="$(dirname "$mod")"
	case "$dir" in
	tools | tools/lintfixture) continue ;;
	tools/*/*) ;; # deeper than tools/<tool>: not an exclusion
	tools/*)
		extra="$(find "$dir" -type f ! -name go.mod ! -name go.sum)"
		if [ -n "$extra" ]; then
			printf 'FAIL: tool modfile directory %s holds more than go.mod/go.sum:\n%s\n' "$dir" "$extra" >&2
			bad=1
		fi
		continue
		;;
	esac
	abs="$root"
	[ "$dir" = . ] || abs="$root/$dir"
	if ! grep -qxF -- "$abs" <<<"$listed"; then
		echo "FAIL: $mod is not listed in go.work (add it, or exclude it in $0)" >&2
		bad=1
	fi
done <<<"$mods"

if [ "$bad" -ne 0 ]; then
	exit 1
fi
echo "every module is in go.work or explicitly excluded"
