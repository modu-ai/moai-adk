#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
doc="$root/skills/moai/workflows/run/task-decomposition.md"
grep -q 'plan_artifact_hash' "$doc"
grep -q 'task_graph_id' "$doc"
grep -q 'do not' "$doc"
grep -q 're-invoke.*manager-spec' "$doc"
grep -q 'hash, scope, dependency set, or risk classification changes' "$doc"
grep -q 're-approval' "$doc"
echo 'PASS: run reuses an approved plan identity and replans only on drift'
