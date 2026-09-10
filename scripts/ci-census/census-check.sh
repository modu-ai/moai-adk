#!/usr/bin/env bash
#
# census-check.sh — fixture-based check of test-census.sh.
#
# Runs the census against a committed `go test -json` fixture stream and
# diffs the result against the committed expected output. This makes the
# census logic verifiable WITHOUT a CI run: the census is the thing that
# turns a CI artifact into evidence, so it must itself carry evidence.
#
# The fixture deliberately contains all five shapes the census must
# distinguish:
#
#   1. a test that called t.Skip           -> SKIPPED TEST
#   2. a package with no test files        -> NOTHING RAN
#   3. a failing test with captured output -> FAILED
#   4. a package-level failure with no failing test (TestMain exit) -> FAILED PKG
#   5. a package that failed to compile    -> BUILD FAILED
#
# Shapes 1 and 2 are the pair REQ-CTO-006 requires be detected by a SINGLE
# Action=="skip" pass and labelled apart by the presence of the Test field.
# A census that filters `Action=="skip" and .Test != null` passes a naive
# "names the skipped test" check while reporting nothing for shape 2 — so
# this check asserts the WHOLE output, not just the presence of a name.
#
# Usage: bash scripts/ci-census/census-check.sh
# Exit:  0 = census output matches expected; 1 = mismatch; 2 = setup error.

set -u

here="$(cd "$(dirname "$0")" && pwd)"
census="$here/test-census.sh"
fixture="$here/testdata/fixture.jsonl"
expected="$here/testdata/expected.txt"

for f in "$census" "$fixture" "$expected"; do
	if [ ! -f "$f" ]; then
		echo "census-check: missing required file: $f" >&2
		exit 2
	fi
done

actual="$(bash "$census" "$fixture")" || {
	echo "census-check: census script exited non-zero" >&2
	exit 2
}

if printf '%s\n' "$actual" | diff -u "$expected" - > /dev/null; then
	echo "census-check: PASS (census output matches $expected)"
	exit 0
fi

echo "census-check: FAIL — census output differs from $expected" >&2
printf '%s\n' "$actual" | diff -u "$expected" - >&2
exit 1
