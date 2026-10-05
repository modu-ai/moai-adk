#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
contract="$root/.claude/rules/moai/workflow/phase-id-contract.md"
run="$root/.claude/skills/moai/workflows/run/phase-execution.md"
grep -q 'stable string IDs' "$contract"
grep -q 'canonical `phase_id`' "$run"
grep -q 'validates the phase DAG' "$run"
grep -q 'Deprecated display labels' "$contract"
printf 'PASS: skip/resume is defined over canonical phase IDs\n'
