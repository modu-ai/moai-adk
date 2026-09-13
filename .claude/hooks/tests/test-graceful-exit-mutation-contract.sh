#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/graceful-exit-mutation-contract.md"
DOC="$ROOT/.claude/skills/moai/workflows/sync/doc-execution.md"
DELIVERY="$ROOT/.claude/skills/moai/workflows/sync/delivery.md"

grep -Fq 'before_tree_key' "$RULE"
grep -Fq 'after_tree_key' "$RULE"
grep -Fq 'partial changes remain' "$RULE"
grep -Fq 'Verified rollback' "$RULE"
grep -Fq 'mutation-aware report' "$DOC"
grep -Fq 'applied/unapplied operations' "$DELIVERY"
grep -Fq 'rollback status' "$DELIVERY"
grep -Fq 'evidence gaps' "$DELIVERY"

printf 'PASS: graceful exits distinguish no mutation from partial or rolled-back changes\n'
