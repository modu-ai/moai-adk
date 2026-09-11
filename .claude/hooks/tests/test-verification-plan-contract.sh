#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/verification-plan-contract.md"
CONTEXT="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-context.md"
QUALITY="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-quality.md"

grep -Fq '(tree_key, command, toolchain, environment, scope)' "$RULE"
grep -Fq 'one owner per key' "$RULE"
grep -Fq 'COMPLETE result' "$RULE"
grep -Fq 'rerun_reason' "$RULE"
grep -Fq 'exact verification key' "$CONTEXT"
grep -Fq 'one verification-plan owner' "$QUALITY"
grep -Fq 'duplicate commands with a different key' "$QUALITY"
grep -Fq 'reuse the exact COMPLETE test/coverage result' "$QUALITY"

printf 'PASS: verification commands have one attributable owner and exact-key reuse\n'
