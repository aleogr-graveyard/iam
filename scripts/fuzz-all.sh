#!/usr/bin/env bash
# Runs every Fuzz* target in every go.work module (except e2e, which needs a
# database and has no targets). FUZZTIME defaults to 5m. On failure the crashing
# inputs (testdata/fuzz) are copied to dist/fuzz-failures/ and the script exits 1.
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"
fuzztime=${FUZZTIME:-5m}
failures="$root/dist/fuzz-failures"
found=0
failed=0

# fuzz_dir <module dir>: `go test -list` prints the matching names of a package
# followed by its "ok <pkg>" line, which is how targets are tied to a package.
fuzz_dir() {
	local dir=$1 line names=() name pkg pkgdir rel listing
	# Captured first so that a build failure fails the script (set -e).
	listing=$(cd "$dir" && go test -list '^Fuzz' ./... </dev/null)
	while IFS= read -r line; do
		case "$line" in
		Fuzz*) names+=("$line") ;;
		ok*)
			pkg=$(awk '{print $2}' <<<"$line")
			for name in "${names[@]+"${names[@]}"}"; do
				found=1
				echo "==> fuzzing $name in $pkg ($fuzztime)"
				if ! (cd "$dir" && go test -run='^$' -fuzz="^$name\$" -fuzztime="$fuzztime" "$pkg" </dev/null); then
					failed=1
					pkgdir=$(cd "$dir" && go list -f '{{.Dir}}' "$pkg" </dev/null)
					if [ -d "$pkgdir/testdata/fuzz" ]; then
						rel=${pkgdir#"$root"/}
						mkdir -p "$failures/$rel"
						cp -R "$pkgdir/testdata/fuzz" "$failures/$rel/"
					fi
				fi
			done
			names=()
			;;
		esac
	done <<<"$listing"
}

# Collect the module list first so no command below can consume its stdin. A
# plain assignment, not process substitution, so a broken go.work fails (set -e).
mods=$(go list -m -f '{{.Dir}}' </dev/null)
if [ -z "$mods" ]; then
	echo "go.work lists no modules" >&2
	exit 1
fi
mapfile -t dirs <<<"$mods"
for dir in "${dirs[@]}"; do
	[ "$dir" = "$root/e2e" ] && continue
	fuzz_dir "$dir"
done

if [ "$found" -eq 0 ]; then
	echo "no fuzz targets"
fi
exit "$failed"
