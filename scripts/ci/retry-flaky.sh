#!/usr/bin/env bash
#
# retry-flaky.sh — run a command, and retry it exactly once when every failing test
# it names is a registry-listed known-flaky test (REQ-CCI-013, design D8).
#
# WHY THIS EXISTS
#   A blanket retry hides regressions. The CI race and test jobs wrap their test
#   command here, so a known flaky test gets one recorded second attempt and no
#   other failure gets a second attempt at all.
#
# RULES
#   - Retry only when the run failed, every failing top-level test name is listed
#     in the registry, and no package failed without naming a failing test. Two
#     stream forms are read:
#       go test -json (the CI form): a top-level test is an event with
#         "Action":"fail" and a "Test" field that contains no "/". A package whose
#         failure names no failing test of its own blocks the retry.
#       go test -v (text form): a top-level `--- FAIL: <name>` line. Indented
#         subtest lines do not count.
#     A failure that is not in the registry, or that names no test, never retries.
#   - Retry exactly once. The retry's status is the run's status, so a registry
#     test that fails twice fails the run.
#   - Every registry entry cites evidence: the name first, then the citation. An
#     entry without a citation refuses the whole run (exit 2) before the command runs.
#   - Streams. The wrapper's stdout is the command's stdout from the FINAL attempt
#     only, so it stays one pure test stream. Every notice the wrapper writes goes
#     to stderr, and the command's own stderr passes through to stderr. The first
#     attempt is buffered, because the wrapper learns whether it is final only when
#     it ends; a non-final attempt never reaches stdout. The final attempt streams
#     straight through. When a retry runs, the first attempt's stream is kept as
#     <log-stem>.attempt1.json beside the retry log.
#   - Reading a stream needs jq, as scripts/ci-census/test-census.sh does. Without
#     jq no retry can be established, so the wrapper does not retry.
#
# Usage: bash scripts/ci/retry-flaky.sh [--registry FILE] -- <command> [args...]
# Env:   FLAKY_RETRY_LOG  retry log file (default: .moai/logs/ci-flaky-retry.log)
# Exit:  0 when the run passes (first attempt or retry); otherwise the failing
#        attempt's status (the retry's status when a retry ran); 2 on a usage or
#        registry error.

set -u

script_dir="$(cd "$(dirname "$0")" && pwd)"
registry="$script_dir/flaky-registry.txt"

while [ $# -gt 0 ]; do
	case "$1" in
		--registry)
			if [ $# -lt 2 ]; then
				echo "retry-flaky: --registry needs a file argument" >&2
				exit 2
			fi
			registry="$2"
			shift 2
			;;
		--)
			shift
			break
			;;
		*)
			echo "retry-flaky: unknown argument: $1" >&2
			exit 2
			;;
	esac
done

if [ $# -eq 0 ]; then
	echo "usage: retry-flaky.sh [--registry FILE] -- <command> [args...]" >&2
	exit 2
fi

if [ ! -f "$registry" ]; then
	echo "retry-flaky: registry not found: $registry" >&2
	exit 2
fi

# A non-comment, non-blank registry line with no citation after the name refuses the run.
uncited="$(awk '/^[[:space:]]*(#|$)/ { next } NF < 2 { print FNR ": " $0; exit }' "$registry")"
if [ -n "$uncited" ]; then
	echo "retry-flaky: registry entry without an evidence citation (line $uncited)" >&2
	exit 2
fi
registry_names="$(awk '/^[[:space:]]*(#|$)/ { next } { print $1 }' "$registry")"

is_registered() {
	printf '%s\n' "$registry_names" | grep -Fxq -- "$1"
}

# failing_names <stream>: the failing top-level test names, one per line, sorted and
# unique. JSON events and text `--- FAIL:` lines are both read. A JSON line is never a
# text line, and a text line is never a JSON event, so no failure is counted twice.
failing_names() {
	{
		jq -R -r 'fromjson? | objects | select(.Action == "fail" and (.Test | type) == "string" and (.Test | contains("/") | not)) | .Test' "$1"
		grep -E '^--- FAIL: ' "$1" | sed -e 's/^--- FAIL: //' -e 's/ .*$//'
	} | sort -u
}

# unnamed_packages <stream>: the packages that failed at package level (an event with no
# "Test") and have no failing top-level test of their own, JSON form only. Such a failure
# names no test, so it blocks the retry even when every named test is registered.
unnamed_packages() {
	local named
	named="$(jq -R -r 'fromjson? | objects | select(.Action == "fail" and (.Test | type) == "string" and (.Test | contains("/") | not)) | .Package' "$1" | sort -u)"
	jq -R -r 'fromjson? | objects | select(.Action == "fail" and (.Test | type) != "string") | .Package' "$1" | sort -u |
	while IFS= read -r pkg; do
		if ! printf '%s\n' "$named" | grep -Fxq -- "$pkg"; then
			echo "$pkg"
		fi
	done
}

work="$(mktemp -d "${TMPDIR:-/tmp}/retry-flaky.XXXXXX")"
trap 'rm -rf "$work"' EXIT

# Attempt 1 is buffered: until it ends, the wrapper cannot tell whether it is the final attempt.
"$@" > "$work/attempt1.json"
status=$?
if [ "$status" -eq 0 ]; then
	cat "$work/attempt1.json"
	exit 0
fi

# Without jq the stream cannot be read, so no retry can be established.
if ! command -v jq > /dev/null 2>&1; then
	echo "retry-flaky: jq not found; the stream cannot be read, so no retry" >&2
	cat "$work/attempt1.json"
	exit "$status"
fi

failing="$(failing_names "$work/attempt1.json")"
if [ -z "$failing" ]; then
	cat "$work/attempt1.json"
	exit "$status"
fi

unnamed="$(unnamed_packages "$work/attempt1.json")"
if [ -n "$unnamed" ]; then
	unnamed_list="$(printf '%s' "$unnamed" | tr '\n' ' ')"
	echo "retry-flaky: $unnamed_list failed without naming a failing test; no retry" >&2
	cat "$work/attempt1.json"
	exit "$status"
fi

names_csv=""
for name in $failing; do
	if ! is_registered "$name"; then
		echo "retry-flaky: $name is not in the registry; no retry" >&2
		cat "$work/attempt1.json"
		exit "$status"
	fi
	names_csv="${names_csv:+$names_csv,}$name"
done

echo "retry-flaky: $names_csv failed on attempt 1 and is registry-listed; re-running once" >&2
log_file="${FLAKY_RETRY_LOG:-.moai/logs/ci-flaky-retry.log}"
mkdir -p "$(dirname "$log_file")"
cp "$work/attempt1.json" "${log_file%.log}.attempt1.json"

# The final attempt streams straight through, so a run cut short still leaves its partial stream.
"$@"
retry_status=$?
outcome=fail
if [ "$retry_status" -eq 0 ]; then
	outcome=pass
fi
record="retried: $names_csv, attempt 2, outcome $outcome"
echo "retry-flaky: $record" >&2
printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$record" >> "$log_file"
exit "$retry_status"
