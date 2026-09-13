#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
interview="$root/.claude/skills/moai/workflows/plan/clarity-interview.md"
assembly="$root/.claude/skills/moai/workflows/plan/spec-assembly.md"

grep -q 'provisional_tier' "$interview"
grep -q 'before Phase 6 research starts' "$interview"
grep -q 'first consumes the interview' "$assembly"
grep -q 'provisional_tier' "$assembly"
grep -q 'only when the tier was absent' "$assembly"
printf 'PASS: tier routing is resolved before optional research and finalized once\n'
