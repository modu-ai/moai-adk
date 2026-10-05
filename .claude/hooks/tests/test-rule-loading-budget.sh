#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
budget="$root/rules/moai/workflow/rule-loading-budget.md"
grep -q 'paths:' "$budget"
grep -q 'InstructionsLoaded' "$budget"
grep -q 'reachability check' "$budget"
for f in \
  "$root/rules/moai/workflow/skill-routing.md" \
  "$root/rules/moai/workflow/phase-id-contract.md" \
  "$root/rules/moai/workflow/delivery-policy.md" \
  "$root/rules/moai/core/security-decision-contract.md"; do
  sed -n '1,8p' "$f" | grep -q '^paths:'
done
echo 'PASS: large conditional rules declare path-scoped loading and measurement gaps'
