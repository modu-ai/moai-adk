#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/snapshot-consumer-contract.md"
HOOK="$ROOT/.claude/hooks/moai/sync-phase-quality-gate.sh"
FOUR_DIM="$ROOT/.claude/workflows/sync-audit-4dim.js"
AUDITOR="$ROOT/.claude/agents/moai/sync-auditor.md"

grep -Fq 'moai verify check --key-current' "$RULE"
grep -Fq 'moai verify check --key-current' "$HOOK"
grep -Fq 'SNAPSHOT_STATUS=$(consume_snapshot)' "$HOOK"
grep -Fq 'snapshot_status=' "$HOOK"
grep -Fq 'snapshot_evidence' "$FOUR_DIM"
grep -Fq 'moai verify check --key-current' "$FOUR_DIM"
grep -Fq 'moai verify check --key-current' "$AUDITOR"
grep -Fq 'evidence_gap' "$AUDITOR"

printf 'PASS: hook, 4-dimension context, and fallback auditor consume one keyed snapshot\n'
