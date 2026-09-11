#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
rule="$root/rules/moai/workflow/resource-budget-contract.md"
docs="$root/skills/moai/workflows/sync/doc-execution.md"
quality="$root/skills/moai/workflows/sync/quality-gates-quality.md"
grep -q 'max_concurrency' "$rule"
grep -q 'bounded queue' "$rule"
grep -q 'resource-budget-contract.md' "$docs"
grep -q 'resource-budget-contract.md' "$quality"
grep -q 'peak_concurrency' "$rule"
echo 'PASS: sync fanout uses one bounded resource budget'
