#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/context-clear-policy.md"
CACHE="$ROOT/.claude/rules/moai/workflow/cache-aware-execution.md"
DOC="$ROOT/.claude/skills/moai/workflows/run/context-loading.md"

grep -Fq 'clear_reason' "$RULE"
grep -Fq 'context_snapshot_id' "$RULE"
grep -Fq 'plan_artifact_hash' "$RULE"
grep -Fq 'tree_key' "$RULE"
grep -Fq '| `warm` |' "$RULE"
grep -Fq '| `clear` |' "$RULE"
grep -Fq 'context-clear-policy.md' "$CACHE"
grep -Fq 'Warm/Clear Handoff Decision' "$DOC"
grep -Fq 'bare `/clear` does not prove continuity' "$DOC"

printf 'PASS: clear versus warm context decisions require an identity-bearing handoff\n'
