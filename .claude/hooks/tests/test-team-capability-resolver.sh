#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/team-capability-resolver.md"
MODE="$ROOT/.claude/skills/moai/workflows/run/mode-orchestration.md"
ORCH="$ROOT/.claude/rules/moai/workflow/orchestration-mode-selection.md"
SPEC="$ROOT/.claude/rules/moai/workflow/spec-workflow.md"

for token in team_requested feature_flag runtime_probe TEAM_AVAILABLE MODE_TEAM_UNAVAILABLE TEAM_NOT_REQUESTED; do
  grep -Fq "$token" "$RULE"
done
grep -Fq 'team-capability-resolver.md' "$MODE"
grep -Fq 'current-runtime probe' "$MODE"
grep -Fq 'resolver-gated' "$ORCH"
grep -Fq 'team-capability-resolver.md' "$SPEC"
grep -Fq 'historical' "$RULE"

printf 'PASS: team mode separates current capability probing from retired genealogy\n'
