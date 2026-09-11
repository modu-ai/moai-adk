#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
doc="$root/skills/moai/workflows/plan/context-discovery.md"
grep -q 'Five technical keywords alone are NOT a skip condition' "$doc"
grep -q 'scope, constraints/non-goals' "$doc"
grep -q 'acceptance or stopping condition' "$doc"
grep -q 'authorization/ownership' "$doc"
! grep -q 'Input contains 5 or more distinct technical keywords' "$doc"
echo 'PASS: interview skipping requires intent completeness, not keyword count'
