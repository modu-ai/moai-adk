#!/bin/bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
doc="$root/skills/moai/workflows/project/mode-detection.md"
grep -q 'one batched AskUserQuestion' "$doc"
grep -q 'Do not issue four sequential question calls' "$doc"
grep -q 'batch response count as one user round' "$doc"
! grep -q 'Present each as a separate AskUserQuestion' "$doc"
echo 'PASS: new-project extended axes are collected in one question batch'
