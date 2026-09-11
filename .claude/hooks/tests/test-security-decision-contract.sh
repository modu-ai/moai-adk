#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
contract="$root/.claude/rules/moai/core/security-decision-contract.md"
auditor="$root/.claude/agents/moai/sync-auditor.md"
quality="$root/.claude/skills/moai/workflows/sync/quality-gates-quality.md"

grep -q '| Critical | BLOCK' "$contract"
grep -q '| High | BLOCK' "$contract"
grep -q 'security-decision-contract.md' "$auditor"
grep -q 'Critical and High findings block' "$quality"
grep -q 'approved exception' "$quality"
if grep -q 'Only CRITICAL findings block' "$quality"; then
  printf 'FAIL: stale CRITICAL-only policy remains\n' >&2
  exit 1
fi
printf 'PASS: one severity and exception contract is referenced by sync audit\n'
