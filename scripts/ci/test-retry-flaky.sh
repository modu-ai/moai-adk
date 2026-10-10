#!/usr/bin/env bash
#
# test-retry-flaky.sh — fixture test for scripts/ci/retry-flaky.sh.
#
# WHY THIS EXISTS
#   The CI race and test jobs run their test command through the flaky-retry
#   wrapper. A wrapper that retries the wrong failure hides a real regression,
#   so each branch of its retry decision is checked here against a fixture
#   command. The fixture counts its own invocations, so "re-run exactly once"
#   is observed, not inferred from the wrapper's own output.
#
# CONTRACT UNDER TEST (REQ-CCI-013, AC-CCI-013-1, design D8)
#   1. A failing registry test is re-run exactly once; the retry's status is the run's.
#   2. The retry and its outcome are recorded on stdout and in the retry log.
#   3. A registry test that also fails on the retry fails the run.
#   4. A failure outside the registry never retries, even beside a registry test.
#   5. A registry entry without an evidence citation refuses the run before any command runs.
#
# Usage: bash scripts/ci/test-retry-flaky.sh
# Exit:  0 when every case passes; 1 when any case fails.

set -u

script_dir="$(cd "$(dirname "$0")" && pwd)"
wrapper="$script_dir/retry-flaky.sh"
tmp="$(mktemp -d "${TMPDIR:-/tmp}/retry-flaky-test.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT

failures=0

expect_eq() { # <label> <want> <got>
	if [ "$2" = "$3" ]; then
		echo "PASS  $1"
	else
		echo "FAIL  $1: want '$2', got '$3'"
		failures=$((failures + 1))
	fi
}

expect_has() { # <label> <needle> <haystack>
	if printf '%s\n' "$3" | grep -Fq -- "$2"; then
		echo "PASS  $1"
	else
		echo "FAIL  $1: '$2' not found"
		failures=$((failures + 1))
	fi
}

expect_lacks() { # <label> <needle> <haystack>
	if printf '%s\n' "$3" | grep -Fq -- "$2"; then
		echo "FAIL  $1: '$2' unexpectedly found"
		failures=$((failures + 1))
	else
		echo "PASS  $1"
	fi
}

# Fixture command under test: prints go-test-shaped output and counts how many
# times it ran, so the number of runs is observed rather than inferred.
cat > "$tmp/fake-test.sh" <<'FAKE'
#!/usr/bin/env bash
# fake-test.sh <scenario> <counter-file>
scenario="$1"
counter="$2"
n=$(( $(cat "$counter" 2>/dev/null || echo 0) + 1 ))
echo "$n" > "$counter"
case "$scenario" in
	flaky-once)
		echo "=== RUN   TestFlakyRegistered"
		if [ "$n" -eq 1 ]; then
			echo "--- FAIL: TestFlakyRegistered (0.00s)"
			echo "FAIL"
			exit 1
		fi
		echo "--- PASS: TestFlakyRegistered (0.00s)"
		echo "PASS"
		exit 0
		;;
	flaky-always)
		echo "=== RUN   TestFlakyRegistered"
		echo "--- FAIL: TestFlakyRegistered (0.00s)"
		echo "FAIL"
		exit 1
		;;
	solid)
		echo "=== RUN   TestSolid"
		echo "--- FAIL: TestSolid (0.00s)"
		echo "FAIL"
		exit 1
		;;
	mixed)
		echo "=== RUN   TestFlakyRegistered"
		echo "=== RUN   TestSolid"
		echo "--- FAIL: TestFlakyRegistered (0.00s)"
		echo "--- FAIL: TestSolid (0.00s)"
		echo "FAIL"
		exit 1
		;;
	pass)
		echo "=== RUN   TestSolid"
		echo "--- PASS: TestSolid (0.00s)"
		echo "PASS"
		exit 0
		;;
esac
FAKE
chmod +x "$tmp/fake-test.sh"

registry="$tmp/registry.txt"
cat > "$registry" <<'REG'
# fixture registry: one entry per line, the name first, then the evidence citation.
TestFlakyRegistered  fixture-evidence: SPEC-CI-FLAKY-STABILIZE-001 precedent
REG

uncited="$tmp/registry-uncited.txt"
printf 'TestFlakyRegistered\n' > "$uncited"

# run_case <label> <scenario> <registry> — runs the wrapper around the fixture.
# Sets rc (wrapper exit status), out (stdout and stderr), count (fixture runs),
# and log (contents of the retry log).
run_case() {
	counter="$tmp/count-$1"
	FLAKY_RETRY_LOG="$tmp/retry-$1.log"
	export FLAKY_RETRY_LOG
	out="$(bash "$wrapper" --registry "$3" -- "$tmp/fake-test.sh" "$2" "$counter" 2>&1)"
	rc=$?
	count="$(cat "$counter" 2>/dev/null || echo 0)"
	log="$(cat "$FLAKY_RETRY_LOG" 2>/dev/null || true)"
}

echo "=== retry-flaky fixture cases ==="

expect_eq "wrapper present" "yes" "$([ -f "$wrapper" ] && echo yes || echo no)"

# Case 1: a registry test fails once and passes on the retry.
run_case flaky-once flaky-once "$registry"
expect_eq   "flaky-once: exit status is the retry's (0)" "0" "$rc"
expect_eq   "flaky-once: command ran exactly twice" "2" "$count"
expect_has  "flaky-once: retry recorded on stdout" "retried: TestFlakyRegistered, attempt 2, outcome pass" "$out"
expect_has  "flaky-once: retry recorded in the log" "retried: TestFlakyRegistered, attempt 2, outcome pass" "$log"

# Case 2: a registry test fails on both attempts; exactly one retry, and the run fails.
run_case flaky-always flaky-always "$registry"
expect_eq   "flaky-always: run fails (non-zero)" "1" "$rc"
expect_eq   "flaky-always: exactly one retry (two runs in total)" "2" "$count"
expect_has  "flaky-always: failed retry recorded" "retried: TestFlakyRegistered, attempt 2, outcome fail" "$out"

# Case 3: a non-registry failure never retries.
run_case solid solid "$registry"
expect_eq    "solid: run fails (non-zero)" "1" "$rc"
expect_eq    "solid: zero retries (one run in total)" "1" "$count"
expect_lacks "solid: no retry recorded" "retried:" "$out"

# Case 4: a registry failure beside a non-registry failure does not retry.
run_case mixed mixed "$registry"
expect_eq    "mixed: run fails (non-zero)" "1" "$rc"
expect_eq    "mixed: zero retries (one run in total)" "1" "$count"
expect_lacks "mixed: no retry recorded" "retried:" "$out"

# Case 5: a passing run is not re-run.
run_case pass pass "$registry"
expect_eq    "pass: exit status 0" "0" "$rc"
expect_eq    "pass: one run in total" "1" "$count"
expect_lacks "pass: no retry recorded" "retried:" "$out"

# Case 6: a registry entry without an evidence citation refuses the run before any command runs.
run_case uncited flaky-once "$uncited"
expect_eq "uncited: refused with usage status 2" "2" "$rc"
expect_eq "uncited: command never ran" "0" "$count"

echo "=== retry-flaky fixture: failures=$failures ==="
if [ "$failures" -eq 0 ]; then
	exit 0
fi
exit 1
