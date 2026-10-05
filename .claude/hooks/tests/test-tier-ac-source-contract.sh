#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
docs="$root/.claude/agents/moai/manager-docs.md"
grep -q '`spec.md` when `tier: S`' "$docs"
grep -q 'use `acceptance.md` when' "$docs"
grep -q 'missing.*empty.*blocker' "$docs"

fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT
mkdir -p "$fixture/S" "$fixture/M"
printf '%s\n' '# inline AC' > "$fixture/S/spec.md"
printf '%s\n' '# external AC' > "$fixture/M/acceptance.md"

resolve() {
  local tier=$1 dir=$2 file
  case "$tier" in
    S) file="$dir/spec.md" ;;
    M|L) file="$dir/acceptance.md" ;;
    *) return 2 ;;
  esac
  test -s "$file"
}
resolve S "$fixture/S"
resolve M "$fixture/M"
set +e
resolve M "$fixture/S"
status=$?
set -e
[[ "$status" -ne 0 ]]
printf 'PASS: tier S resolves inline AC and M/L require acceptance.md\n'
