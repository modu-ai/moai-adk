#!/usr/bin/env bash
#
# test-census.sh — print a per-test census from a `go test -json` event stream.
#
# WHY THIS EXISTS
#   CI runs `go test` without -v or -json, so an rc=0 cannot distinguish
#   "the test ran and passed" from "the test skipped" from "the selector
#   matched nothing". This script reads the -json stream (which the CI step
#   redirects to a file, keeping the console clean) and prints the short
#   census a human actually needs: what did not run, and what failed.
#
# THE SINGLE-PREDICATE PROPERTY (REQ-CTO-006)
#   `go test -json` expresses BOTH forms of "did not run" as one event type,
#   distinguished only by the presence of a field:
#
#     {"Action":"skip","Package":"…","Test":"TestFoo"}  -> a test skipped itself
#     {"Action":"skip","Package":"…"}                   -> nothing ran in the package
#
#   So ONE Action=="skip" pass catches both, and the label is chosen by
#   reading .Test. This is not two detections wearing one name; a census that
#   filters `Action=="skip" and .Test != null` silently loses every
#   zero-test-file package.
#
# BUILD FAILURES ARE IN THE STREAM, NOT ON stderr
#   Measured (2026-08-29, go test -json on a package with a syntax error):
#   the compiler diagnostic arrives as build-output/build-fail events on
#   STDOUT and stderr is EMPTY. So a step that only redirects stdout would
#   swallow the build error entirely unless the census prints it back. That
#   is why BUILD FAILED rows exist and are printed first.
#
# OUTPUT CONTRACT
#   === test census ===
#   BUILD FAILED  <import-path>
#     <verbatim build output>
#   FAILED        <package>  <test>
#     <verbatim captured output of that test>
#   FAILED PKG    <package>
#     <verbatim package-level output>            (only when no test in the
#                                                 package failed and it was
#                                                 not a build failure)
#   NOTHING RAN   <package>
#   SKIPPED TEST  <package>  <test>
#   === totals: packages=N passed=N skipped=N nothing-ran=N failed=N build-failed=N ===
#
#   Failures come first because a red run is the case the census exists for.
#   Rows within a class are sorted, so the output is deterministic.
#
# Usage: bash scripts/ci-census/test-census.sh <stream.json>
# Exit:  0 always on a readable stream — this is a REPORTER, not a gate. The
#        CI step re-raises the exit status of `go test` itself; a census that
#        also failed would double-count and could turn a green run red.
#        Exit 2 only when it cannot read the stream or jq is unavailable.

set -u

# Cap on how many failing tests get their captured output printed. Beyond
# this the names are still listed; a red run with hundreds of failures is
# already a "read the artifact" situation and dumping it all defeats the
# console-cleanliness this whole change buys.
MAX_FAILURE_BODIES="${CENSUS_MAX_FAILURE_BODIES:-50}"

stream="${1:-}"
if [ -z "$stream" ]; then
	echo "usage: test-census.sh <go-test-json-stream-file>" >&2
	exit 2
fi
if [ ! -f "$stream" ]; then
	echo "test-census: stream file not found: $stream" >&2
	exit 2
fi
if ! command -v jq > /dev/null 2>&1; then
	echo "test-census: jq not found on PATH" >&2
	exit 2
fi

echo "=== test census ==="

# --- build failures -------------------------------------------------------
# A build-fail event carries ImportPath (not Package) and no Test. Its
# diagnostic text is in the preceding build-output events for the same path.
build_failed="$(jq -r 'select(.Action=="build-fail") | .ImportPath' "$stream" | sort -u)"
if [ -n "$build_failed" ]; then
	while IFS= read -r path; do
		[ -z "$path" ] && continue
		echo "BUILD FAILED  $path"
		jq -j --arg p "$path" \
			'select(.Action=="build-output" and .ImportPath==$p) | .Output' "$stream" |
			sed 's/^/  /'
	done <<- EOF
		$build_failed
	EOF
fi

# --- failing tests --------------------------------------------------------
failed_tests="$(jq -r 'select(.Action=="fail" and (.Test // null) != null) | "\(.Package)\t\(.Test)"' "$stream" | sort -u)"
failed_count=0
if [ -n "$failed_tests" ]; then
	while IFS="$(printf '\t')" read -r pkg test; do
		[ -z "$pkg" ] && continue
		failed_count=$((failed_count + 1))
		echo "FAILED        $pkg  $test"
		if [ "$failed_count" -le "$MAX_FAILURE_BODIES" ]; then
			jq -j --arg p "$pkg" --arg t "$test" \
				'select(.Action=="output" and .Package==$p and .Test==$t) | .Output' "$stream" |
				sed 's/^/  /'
		fi
	done <<- EOF
		$failed_tests
	EOF
	if [ "$failed_count" -gt "$MAX_FAILURE_BODIES" ]; then
		echo "  (captured output omitted beyond the first $MAX_FAILURE_BODIES failures — read the uploaded stream artifact)"
	fi
fi

# --- package-level failures with no failing test --------------------------
# Excludes build failures (they carry FailedBuild and are reported above) and
# packages that already have a FAILED row, which would only repeat them.
pkgs_with_failed_tests="$(jq -r 'select(.Action=="fail" and (.Test // null) != null) | .Package' "$stream" | sort -u)"
pkg_failed="$(jq -r 'select(.Action=="fail" and (.Test // null) == null and (.FailedBuild // null) == null) | .Package' "$stream" | sort -u)"
if [ -n "$pkg_failed" ]; then
	while IFS= read -r pkg; do
		[ -z "$pkg" ] && continue
		if printf '%s\n' "$pkgs_with_failed_tests" | grep -Fxq "$pkg"; then
			continue
		fi
		echo "FAILED PKG    $pkg"
		jq -j --arg p "$pkg" \
			'select(.Action=="output" and .Package==$p and (.Test // null) == null) | .Output' "$stream" |
			sed 's/^/  /'
	done <<- EOF
		$pkg_failed
	EOF
fi

# --- the single Action=="skip" pass (REQ-CTO-006) -------------------------
# One filter, two labels, chosen by the presence of .Test. Do not split this
# into two selects: that is the mutant the fixture check exists to reject.
jq -r 'select(.Action=="skip")
       | if (.Test // null) == null
         then "NOTHING RAN   \(.Package)"
         else "SKIPPED TEST  \(.Package)  \(.Test)"
         end' "$stream" | sort -u

# --- totals ---------------------------------------------------------------
count() { jq -r "$1" "$stream" | sort -u | grep -c . || true; }

packages=$(count 'select(has("Package")) | .Package')
passed=$(count 'select(.Action=="pass" and (.Test // null) != null) | "\(.Package)\t\(.Test)"')
skipped=$(count 'select(.Action=="skip" and (.Test // null) != null) | "\(.Package)\t\(.Test)"')
nothing_ran=$(count 'select(.Action=="skip" and (.Test // null) == null) | .Package')
failed=$(count 'select(.Action=="fail" and (.Test // null) != null) | "\(.Package)\t\(.Test)"')
build_failed_total=$(count 'select(.Action=="build-fail") | .ImportPath')

echo "=== totals: packages=$packages passed=$passed skipped=$skipped nothing-ran=$nothing_ran failed=$failed build-failed=$build_failed_total ==="
