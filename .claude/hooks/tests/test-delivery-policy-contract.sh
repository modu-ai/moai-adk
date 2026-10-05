#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
policy="$root/.claude/rules/moai/workflow/delivery-policy.md"
spec="$root/.claude/rules/moai/workflow/spec-workflow.md"
git_agent="$root/.claude/agents/moai/manager-git.md"
delivery="$root/.claude/skills/moai/workflows/sync/delivery.md"

grep -q 'Phase agents create commits but do not push' "$policy"
grep -q 'manager-git' "$spec"
grep -q 'Phase agents never push directly' "$git_agent"
grep -q 'No phase agent pushes directly' "$delivery"
if grep -q 'Tier S/M defaults to the direct Route A' "$git_agent"; then
  printf 'FAIL: stale direct-push manager-git policy remains\n' >&2
  exit 1
fi
printf 'PASS: all tiers share manager-git delivery ownership\n'
