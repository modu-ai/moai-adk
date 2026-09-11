#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
project="$root/.claude/skills/moai/workflows/project/doc-generation.md"
run="$root/.claude/skills/moai/workflows/run/phase-execution.md"
sync="$root/.claude/agents/moai/sync-auditor.md"

grep -q 'sync-auditor.*post-implementation only' "$project"
grep -q 'Invoke `plan-auditor` to review the pre-implementation plan' "$run"
grep -q 'Agent(subagent_type="plan-auditor")' "$run"
grep -q 'post-implementation only' "$sync"
if grep -q 'Invoke sync-auditor to review the plan' "$run"; then
  printf 'FAIL: pre-implementation sync-auditor route remains\n' >&2
  exit 1
fi
printf 'PASS: pre-implementation review uses plan-auditor only\n'
