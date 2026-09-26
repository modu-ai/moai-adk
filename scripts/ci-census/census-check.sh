#!/usr/bin/env bash
#
# census-check.sh — fixture-based check of test-census.sh.
#
# Runs the census against a committed `go test -json` fixture stream and
# diffs the result against the committed expected output. This makes the
# census logic verifiable WITHOUT a CI run: the census is the thing that
# turns a CI artifact into evidence, so it must itself carry evidence.
#
# The fixture deliberately contains all nine shapes the census must
# distinguish:
#
#   1. a test that called t.Skip           -> SKIPPED TEST
#   2. a package with no test files        -> NOTHING RAN
#   3. a failing test with captured output -> FAILED
#   4. a package-level failure with no failing test (TestMain exit) -> FAILED PKG
#   5. a package that failed to compile    -> BUILD FAILED
#   6. an AC snapshot absent report logged by a PASSING test (repeated, one
#      copy CRLF-terminated, one carrying '%') -> one ::notice per report
#   7. a passing package with a package-level pass event and a second
#      skipped test -> keeps the totals line honest: counting package-level
#      pass events inflates passed, and counting packages with skipped tests
#      as nothing-ran inflates nothing-ran, because this fixture has two such
#      packages but only one package with no test files
#   8. two tests with the same result in one package (two passes and two
#      skips in epsilon, two failures in alpha) -> the passed, skipped, and
#      failed totals count (package, test) pairs; counting packages instead
#      prints a smaller number
#   9. a second failing package (zeta) whose passing, skipped, and failing
#      tests reuse alpha's test names, its events interleaved between alpha's
#      as a parallel run emits them -> counting test names without the
#      package also prints a smaller number, and dropping the sort before the
#      FAILED and SKIPPED TEST rows prints them out of order
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
