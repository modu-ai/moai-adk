#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/read-only-status-contract.md"
DOC="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-context.md"
QUALITY="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-quality.md"

grep -Fq 'mode=status' "$RULE"
grep -Fq 'read_only=true' "$RULE"
grep -Fq 'writer_policy=deny' "$RULE"
grep -Fq 'before_tree_key' "$RULE"
grep -Fq 'after_tree_key' "$RULE"
grep -Fq 'git add' "$RULE"
grep -Fq 'auto-fix' "$RULE"
grep -Fq 'read_only=true' "$DOC"
grep -Fq 'writer_policy=deny' "$DOC"
grep -Fq 'before_tree_key' "$DOC"
grep -Fq 'Status mode early exit' "$QUALITY"
grep -Fq 'writer-attempt count' "$QUALITY"

printf 'PASS: sync status mode has an explicit writer deny boundary and mutation proof\n'
