#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
docs="$root/.claude/agents/moai/manager-docs.md"
grep -q '\[a-z\]?' "$docs"

fixture=$(mktemp)
trap 'rm -f "$fixture"' EXIT
printf '%s\n' '### AC-SYN-001a' '### AC-SYN-001b' > "$fixture"
count=$(rg -o 'AC-SYN-001[a-z]?' "$fixture" | sort -u | wc -l | tr -d ' ')
[[ "$count" -eq 2 ]]
printf 'PASS: AC suffixes a/b remain distinct criteria (count=%s)\n' "$count"
