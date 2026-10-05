#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
assembly="$root/.claude/skills/moai/workflows/plan/spec-assembly.md"
harness="$root/.moai/config/sections/harness.yaml"
grep -q 'tier-resolved ceiling' "$assembly"
grep -q 'remaining_attempts' "$assembly"
grep -q 'S: 1' "$harness"
grep -q 'M: 2' "$harness"
grep -q 'L: 3' "$harness"
if grep -q 'standard.*max_iterations: 3' "$assembly"; then
  printf 'FAIL: workflow still overrides Tier S/M with a fixed three-iteration loop\n' >&2
  exit 1
fi
printf 'PASS: one tier-resolved audit ceiling owns all review attempts\n'
