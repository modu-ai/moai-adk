#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
docs="$root/.claude/skills/moai/workflows/sync/delivery.md"
if grep -q '^wait$' "$docs"; then
  printf 'FAIL: unqualified wait remains in cross-build example\n' >&2
  exit 1
fi

run_matrix() {
  false & p1=$!
  true & p2=$!
  failed=0
  for pid in "$p1" "$p2"; do
    if wait "$pid"; then
      :
    else
      failed=1
    fi
  done
  (( failed == 0 ))
}

set +e
run_matrix
status=$?
set -e
[[ "$status" -eq 1 ]]
printf 'PASS: one failed cross-build target makes the matrix fail\n'
