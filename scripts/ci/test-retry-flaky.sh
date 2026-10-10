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
#   2. The retry and its outcome are recorded on stderr and in the retry log.
#   3. A registry test that also fails on the retry fails the run.
#   4. A failure outside the registry never retries, even beside a registry test.
#   5. A registry entry without an evidence citation refuses the run before any command runs.
#   6. A go test -json stream is read the same way: a failing registry test is re-run
#      exactly once, and the retry is recorded on stderr.
#   7. Stdout carries the final attempt's stream only, as valid JSON lines; notices go to stderr.
#   8. A retried first attempt is kept beside the retry log.
#   9. A failure that names no test (a package-level failure) never retries, even beside a registry test.
#  10. An interrupted first attempt's partial stream reaches stderr; stdout never carries it.
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
	json-flaky-once)
		if [ "$n" -eq 1 ]; then
			echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestFlakyRegistered"}'
			echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestFlakyRegistered","Output":"ATTEMPT-ONE\n"}'
			echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Test":"TestFlakyRegistered","Elapsed":0}'
			echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Elapsed":0}'
			exit 1
		fi
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestFlakyRegistered"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestFlakyRegistered","Output":"ATTEMPT-TWO\n"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"pass","Package":"example/fixture","Test":"TestFlakyRegistered","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"pass","Package":"example/fixture","Elapsed":0}'
		exit 0
		;;
	json-flaky-always)
		mark=ATTEMPT-ONE
		if [ "$n" -gt 1 ]; then
			mark=ATTEMPT-TWO
		fi
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestFlakyRegistered"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestFlakyRegistered","Output":"'"$mark"'\n"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Test":"TestFlakyRegistered","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Elapsed":0}'
		exit 1
		;;
	json-solid)
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestSolid"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestSolid","Output":"SOLID-MARK\n"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Test":"TestSolid","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Elapsed":0}'
		exit 1
		;;
	json-mixed)
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestFlakyRegistered"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestSolid"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestSolid","Output":"MIXED-MARK\n"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Test":"TestFlakyRegistered","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Test":"TestSolid","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture","Elapsed":0}'
		exit 1
		;;
	json-pass)
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture","Test":"TestSolid"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestSolid","Output":"PASS-MARK\n"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"pass","Package":"example/fixture","Test":"TestSolid","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"pass","Package":"example/fixture","Elapsed":0}'
		exit 0
		;;
	json-unnamed-pkg)
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"run","Package":"example/fixture-a","Test":"TestFlakyRegistered"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture-a","Test":"TestFlakyRegistered","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture-a","Elapsed":0}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture-b","Output":"UNNAMED-MARK\n"}'
		echo '{"Time":"2026-10-10T00:00:00Z","Action":"fail","Package":"example/fixture-b","Elapsed":0}'
		exit 1
		;;
		interrupt-partial)
			echo "$$" > "$counter.pid"
			echo '{"Time":"2026-10-10T00:00:00Z","Action":"output","Package":"example/fixture","Test":"TestSolid","Output":"INTERRUPT-MARK\n"}'
			exec sleep 30
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

# run_case <label> <scenario> <registry> — runs the wrapper around the fixture, with stdout
# and stderr captured apart. Sets rc (wrapper exit status), count (fixture runs), log (the
# retry log), sout (stdout), serr (stderr), out (both), and jq_rc (0 when stdout reads as
# JSON lines, which only the JSON cases check).
run_case() {
	counter="$tmp/count-$1"
	FLAKY_RETRY_LOG="$tmp/retry-$1.log"
	export FLAKY_RETRY_LOG
	bash "$wrapper" --registry "$3" -- "$tmp/fake-test.sh" "$2" "$counter" > "$tmp/stdout-$1.txt" 2> "$tmp/stderr-$1.txt"
	rc=$?
	count="$(cat "$counter" 2>/dev/null || echo 0)"
	log="$(cat "$FLAKY_RETRY_LOG" 2>/dev/null || true)"
	sout="$(cat "$tmp/stdout-$1.txt")"
	serr="$(cat "$tmp/stderr-$1.txt")"
	out="$(cat "$tmp/stdout-$1.txt" "$tmp/stderr-$1.txt")"
	jq -c . "$tmp/stdout-$1.txt" > /dev/null 2>&1
	jq_rc=$?
}

# make_registry <label> — one temporary registry per JSON case; sets reg_path.
make_registry() {
	reg_path="$tmp/registry-$1.txt"
	cat > "$reg_path" <<'REG'
# fixture registry: one entry per line, the name first, then the evidence citation.
TestFlakyRegistered  fixture-evidence: SPEC-CI-FLAKY-STABILIZE-001 precedent
REG
}

echo "=== retry-flaky fixture cases ==="

expect_eq "wrapper present" "yes" "$([ -f "$wrapper" ] && echo yes || echo no)"

# Case 1: a registry test fails once and passes on the retry.
run_case flaky-once flaky-once "$registry"
expect_eq   "flaky-once: exit status is the retry's (0)" "0" "$rc"
expect_eq   "flaky-once: command ran exactly twice" "2" "$count"
expect_has  "flaky-once: retry recorded on stderr" "retried: TestFlakyRegistered, attempt 2, outcome pass" "$serr"
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

echo "=== retry-flaky JSON-stream cases ==="

# JSON case 1: a go test -json registry test fails once and passes on the retry.
make_registry json-flaky-once
run_case json-flaky-once json-flaky-once "$reg_path"
expect_eq    "json flaky-once: exit status is the retry's (0)" "0" "$rc"
expect_eq    "json flaky-once: command ran exactly twice" "2" "$count"
expect_has   "json flaky-once: retry recorded on stderr" "retried: TestFlakyRegistered, attempt 2, outcome pass" "$serr"
expect_lacks "json flaky-once: no wrapper notice on stdout" "retry-flaky:" "$sout"
expect_eq    "json flaky-once: stdout is valid JSON lines" "0" "$jq_rc"
expect_has   "json flaky-once: stdout carries the final attempt" "ATTEMPT-TWO" "$sout"
expect_lacks "json flaky-once: stdout omits the first attempt" "ATTEMPT-ONE" "$sout"
expect_has   "json flaky-once: first attempt kept beside the retry log" "ATTEMPT-ONE" "$(cat "$tmp/retry-json-flaky-once.attempt1.json" 2>/dev/null || true)"
expect_lacks "json flaky-once: kept file holds only the first attempt" "ATTEMPT-TWO" "$(cat "$tmp/retry-json-flaky-once.attempt1.json" 2>/dev/null || true)"
expect_has   "json flaky-once: retry recorded in the log" "retried: TestFlakyRegistered, attempt 2, outcome pass" "$log"

# JSON case 2: a go test -json registry test fails on both attempts; one retry, and the run fails.
make_registry json-flaky-always
run_case json-flaky-always json-flaky-always "$reg_path"
expect_eq    "json flaky-always: run fails (non-zero)" "1" "$rc"
expect_eq    "json flaky-always: exactly one retry (two runs in total)" "2" "$count"
expect_has   "json flaky-always: failed retry recorded on stderr" "retried: TestFlakyRegistered, attempt 2, outcome fail" "$serr"
expect_eq    "json flaky-always: stdout is valid JSON lines" "0" "$jq_rc"
expect_has   "json flaky-always: stdout carries the second attempt" "ATTEMPT-TWO" "$sout"
expect_lacks "json flaky-always: stdout omits the first attempt" "ATTEMPT-ONE" "$sout"

# JSON case 3: a non-registry failure never retries.
make_registry json-solid
run_case json-solid json-solid "$reg_path"
expect_eq    "json solid: run fails (non-zero)" "1" "$rc"
expect_eq    "json solid: zero retries (one run in total)" "1" "$count"
expect_has   "json solid: refusal named on stderr" "TestSolid is not in the registry; no retry" "$serr"
expect_lacks "json solid: no retry recorded" "retried:" "$serr"
expect_eq    "json solid: stdout is valid JSON lines" "0" "$jq_rc"
expect_has   "json solid: stdout is the single attempt's stream" "SOLID-MARK" "$sout"

# JSON case 4: a registry failure beside a non-registry failure does not retry.
make_registry json-mixed
run_case json-mixed json-mixed "$reg_path"
expect_eq    "json mixed: run fails (non-zero)" "1" "$rc"
expect_eq    "json mixed: zero retries (one run in total)" "1" "$count"
expect_has   "json mixed: refusal names the unregistered test on stderr" "TestSolid is not in the registry; no retry" "$serr"
expect_lacks "json mixed: no retry recorded" "retried:" "$serr"
expect_eq    "json mixed: stdout is valid JSON lines" "0" "$jq_rc"

# JSON case 5: a passing run is not re-run.
make_registry json-pass
run_case json-pass json-pass "$reg_path"
expect_eq    "json pass: exit status 0" "0" "$rc"
expect_eq    "json pass: one run in total" "1" "$count"
expect_lacks "json pass: no retry recorded" "retried:" "$serr"
expect_eq    "json pass: stdout is valid JSON lines" "0" "$jq_rc"
expect_has   "json pass: stdout carries the stream" "\"Action\":\"pass\"" "$sout"

# JSON case 6: a registry test beside a package-level failure that names no test does not
# retry. The unnamed package is fixture-b; fixture-a's registry test alone would retry.
make_registry json-unnamed-pkg
run_case json-unnamed-pkg json-unnamed-pkg "$reg_path"
expect_eq    "json unnamed-pkg: run fails (non-zero)" "1" "$rc"
expect_eq    "json unnamed-pkg: zero retries (a failure that names no test blocks the retry)" "1" "$count"
expect_has   "json unnamed-pkg: refusal names the package on stderr" "example/fixture-b" "$serr"
expect_lacks "json unnamed-pkg: no retry recorded" "retried:" "$serr"

# JSON case 7: an uncited registry entry refuses the run before the command runs.
printf 'TestFlakyRegistered\n' > "$tmp/registry-json-uncited.txt"
run_case json-uncited json-flaky-once "$tmp/registry-json-uncited.txt"
expect_eq    "json uncited: refused with usage status 2" "2" "$rc"
expect_eq    "json uncited: command never ran" "0" "$count"
expect_has   "json uncited: refusal named on stderr" "without an evidence citation" "$serr"

echo "=== retry-flaky interrupt cases ==="

# Interrupt case 1: a TERM during the first attempt keeps its partial stream on stderr, and
# stdout never carries it. The fixture prints one marked JSON line, then sleeps 30 seconds, so
# the case waits (bounded, about 10 seconds) for the mark to reach the wrapper's attempt-1 file,
# signals the wrapper and then its child, and stays fast. TMPDIR points at a test-owned
# directory, so the wrapper's work directory is found by glob.
interrupt_tmp="$tmp/interrupt-tmp"
mkdir -p "$interrupt_tmp"
icount="$tmp/count-interrupt-partial"
FLAKY_RETRY_LOG="$tmp/retry-interrupt-partial.log"
export FLAKY_RETRY_LOG
TMPDIR="$interrupt_tmp" bash "$wrapper" --registry "$registry" -- "$tmp/fake-test.sh" interrupt-partial "$icount" > "$tmp/stdout-interrupt-partial.txt" 2> "$tmp/stderr-interrupt-partial.txt" &
wpid=$!
polls=0
until grep -qs 'INTERRUPT-MARK' "$interrupt_tmp"/retry-flaky.*/attempt1.json || [ "$polls" -ge 100 ]; do
	sleep 0.1
	polls=$((polls + 1))
done
kill -TERM "$wpid" 2> /dev/null
sleep 0.2
ipid="$(cat "$icount.pid" 2> /dev/null || true)"
if [ -n "$ipid" ]; then
	kill -TERM "$ipid" 2> /dev/null
fi
wait "$wpid"
irc=$?
ierr="$(cat "$tmp/stderr-interrupt-partial.txt")"
iout="$(cat "$tmp/stdout-interrupt-partial.txt")"
expect_eq    "interrupt: exit status is 128 plus SIGTERM (143)" "143" "$irc"
expect_has   "interrupt: the first attempt's partial stream reaches stderr" "INTERRUPT-MARK" "$ierr"
expect_lacks "interrupt: stdout does not carry the interrupted attempt" "INTERRUPT-MARK" "$iout"
expect_eq    "interrupt: the work directory is removed" "0" "$(ls -d "$interrupt_tmp"/retry-flaky.* 2> /dev/null | wc -l | tr -d ' ')"

echo "=== retry-flaky fixture: failures=$failures ==="
if [ "$failures" -eq 0 ]; then
	exit 0
fi
exit 1
