#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/sync-coverage-scope-contract.md"
CONTEXT="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-context.md"
QUALITY="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-quality.md"

for mode in auto force project status; do
  grep -Fq "| \`$mode\` |" "$RULE"
done
grep -Fq 'sync_scope' "$RULE"
grep -Fq 'coverage_scope' "$RULE"
grep -Fq 'pre-existing coverage gap' "$RULE"
grep -Fq 'changed package/import closure' "$CONTEXT"
grep -Fq 'full-repository diagnostics/coverage' "$CONTEXT"
grep -Fq 'mode-specific `coverage_scope`' "$QUALITY"
grep -Fq 'pre-existing coverage debt' "$QUALITY"

printf 'PASS: sync diagnostic and coverage scope is explicit for every mode\n'
