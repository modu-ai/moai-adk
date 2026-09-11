#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
spec="$root/.claude/rules/moai/workflow/spec-workflow.md"
guard="$root/.claude/rules/moai/workflow/main-checkout-branch-guard.md"

grep -q 'launcher worktree' "$spec"
grep -q 'shared primary checkout is not an execution fallback' "$spec"
grep -q 'moai cc -w <name>' "$guard"
if grep -q '^git checkout main$\|^git reset --hard origin/main$' "$spec"; then
  printf 'FAIL: destructive shared-checkout cleanup remains\n' >&2
  exit 1
fi
printf 'PASS: plan/run/sync use launcher worktrees and non-destructive integration\n'
