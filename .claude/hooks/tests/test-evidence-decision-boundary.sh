#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
run_doc="$root/skills/moai/workflows/run/task-decomposition.md"
sync_doc="$root/skills/moai/workflows/sync/quality-gates-quality.md"
for token in tree_key 'COMPLETE' 'INCOMPLETE' 'CONTESTED' 'AC set' 'rubric'; do
  grep -q "$token" "$run_doc"
done
grep -q 'Run-evidence reuse boundary' "$sync_doc"
grep -q 'mismatch triggers' "$sync_doc"
grep -q 're-execution reason' "$sync_doc"
echo 'PASS: run evidence reuse is separated from sync verdict ownership'
