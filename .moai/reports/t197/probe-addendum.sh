#!/bin/bash
# t197 measurement probe — ADDENDUM.
#
#   bash .moai/reports/t197/probe-addendum.sh >> .moai/reports/t197/probe-output.txt
#
# Why an addendum rather than a re-run: 40 live citations name line ranges in
# probe-output.txt. Re-running the whole probe renumbers every one of them.
# Appending leaves lines 1-278 byte-identical and adds new ranges at the end.
#
# It measures exactly the three claims an audit found unsupported by their
# cited ranges (iteration 7, F1): the branch name, the generator's write path,
# and the ChatGPT credential key set the auth ladder recognizes.
#
# Read-only: every command is a read. No build, no external binary invocation.
# Run it from the repository root.

set -uo pipefail

FAILURES=0

run_rc() {
  local expected="$1"; shift
  echo "\$ $1"
  eval "$1"
  local rc=$?
  echo "rc=$rc"
  if [ "$rc" != "$expected" ]; then
    echo "UNEXPECTED: rc=$rc, expected $expected"
    FAILURES=$((FAILURES + 1))
  fi
  echo
}
run() { run_rc 0 "$1"; }

echo "### ADDENDUM A — claims whose cited range did not show them (iter7 F1)"
echo
echo "## A-1 branch name of the measured tree"
run 'git rev-parse --abbrev-ref HEAD'

echo "## A-2 the generator writes both wiring files"
run 'grep -n "func Wire(" internal/codexwiring/wire.go'
run 'grep -nE "(hooksPath|cfgPath) := filepath.Join\(projectRoot, (Hooks|Config)RelPath\)" internal/codexwiring/wire.go'
run 'grep -nE "writeAtomic\((hooksPath|cfgPath), " internal/codexwiring/wire.go'
run 'grep -n "codexwiring.Wire(projectRoot" internal/cli/init.go'

echo "## A-3 ChatGPT credential keys present in a real auth.json (shape only, no values)"
run 'sed -n "150,166p" .moai/reports/t197/probe-output.txt | grep -E "id_token|access_token|refresh_token|account_id"'

echo "### addendum result"
if [ "$FAILURES" = 0 ]; then
  echo "ADDENDUM OK: every command exited as expected"
else
  echo "ADDENDUM FAILED: $FAILURES command(s) exited unexpectedly"
fi
exit "$FAILURES"
