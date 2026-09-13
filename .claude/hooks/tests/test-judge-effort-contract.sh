#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
script="$root/workflows/sync-audit-4dim.js"
rule="$root/rules/moai/workflow/verify-judge-effort-contract.md"
grep -q 'const JUDGE_EFFORT' "$script"
grep -q 'judge_effort' "$script"
grep -q "new Set(\['low', 'medium', 'high'\])" "$script"
! grep -q "effort: 'xhigh'" "$script"
grep -q 'xhigh' "$rule"
echo 'PASS: 4-dimension judges use the resolved supported effort profile'
