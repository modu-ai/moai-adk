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
#   - Retry only when the run failed, it printed at least one top-level
#     `--- FAIL: <name>` line (go test -v form; indented subtest lines do not
#     count), and every such <name> is listed in the registry. A failure that is
#     not in the registry, or that names no test, never retries.
#   - Retry exactly once. The retry's status is the run's status, so a registry
#     test that fails twice fails the run.
#   - Every registry entry cites evidence: the name first, then the citation. An
#     entry without a citation refuses the whole run (exit 2) before the command runs.
#   - Output is merged (2>&1) so failing test names reach the parser. Both attempts
#     stream to stdout as they run.
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

work="$(mktemp -d "${TMPDIR:-/tmp}/retry-flaky.XXXXXX")"
trap 'rm -rf "$work"' EXIT

"$@" 2>&1 | tee "$work/attempt1.txt"
status=${PIPESTATUS[0]}
if [ "$status" -eq 0 ]; then
	exit 0
fi

failing="$(grep -E '^--- FAIL: ' "$work/attempt1.txt" | sed -e 's/^--- FAIL: //' -e 's/ .*$//' | sort -u)"
if [ -z "$failing" ]; then
	exit "$status"
fi

names_csv=""
for name in $failing; do
	if ! is_registered "$name"; then
		echo "retry-flaky: $name is not in the registry; no retry" >&2
		exit "$status"
	fi
	names_csv="${names_csv:+$names_csv,}$name"
done

echo "retry-flaky: $names_csv failed on attempt 1 and is registry-listed; re-running once"
"$@" 2>&1 | tee "$work/attempt2.txt"
retry_status=${PIPESTATUS[0]}
outcome=fail
if [ "$retry_status" -eq 0 ]; then
	outcome=pass
fi
record="retried: $names_csv, attempt 2, outcome $outcome"
echo "retry-flaky: $record"

log_file="${FLAKY_RETRY_LOG:-.moai/logs/ci-flaky-retry.log}"
mkdir -p "$(dirname "$log_file")"
printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$record" >> "$log_file"

exit "$retry_status"
