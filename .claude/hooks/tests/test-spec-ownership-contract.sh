#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
docs="$root/.claude/skills/moai/workflows/sync/doc-execution.md"
owner="$root/.claude/agents/moai/manager-docs.md"

grep -q 'manager-docs.*MUST NOT edit' "$docs"
grep -q 'manager-spec' "$docs"
grep -q 'MUST NOT modify spec.md / plan.md / acceptance.md body content' "$owner"
grep -q 'structured blocker report' "$owner"
printf 'PASS: SPEC body ownership routes divergence to manager-spec\n'
