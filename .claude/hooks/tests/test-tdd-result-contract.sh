#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/workflow/tdd-result-contract.md"
DOC="$ROOT/.claude/skills/moai/workflows/run/task-decomposition.md"

for result in EXPECTED_RED REGRESSION_FAILURE TOOL_FAILURE PASS; do
  grep -Fq "\`$result\`" "$RULE"
done
grep -Fq 'intended assertion' "$RULE"
grep -Fq 'verbatim output' "$RULE"
grep -Fq 'never an' "$RULE"
grep -Fq 'semantic result' "$DOC"
grep -Fq 'Compile errors' "$DOC"
grep -Fq 'regression set' "$DOC"

printf 'PASS: TDD distinguishes expected RED from regression and tool failures\n'
