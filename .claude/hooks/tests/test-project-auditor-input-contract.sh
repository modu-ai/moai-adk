#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
workflow="$root/.claude/skills/moai/workflows/project/doc-generation.md"
auditor="$root/.claude/agents/moai/plan-auditor.md"

grep -q 'input_type=project' "$workflow"
grep -q 'input_type=project' "$auditor"
grep -q 'product.md' "$auditor"
grep -q 'structure.md' "$auditor"
grep -q 'tech.md' "$auditor"
grep -q 'does not require `spec.md`' "$auditor"
grep -q 'PROJECT-review-' "$workflow"
printf 'PASS: project audit uses a typed project input contract\n'
