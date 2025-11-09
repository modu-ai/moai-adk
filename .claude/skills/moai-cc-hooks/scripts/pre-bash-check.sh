#!/bin/bash
# Pre-Bash Command Validator (from Context7 official docs)
set -euo pipefail

FORBIDDEN=(
  "rm -rf /"
  "sudo rm"
  "chmod 777"
  "eval.*curl"
)

COMMAND="${1:-}"
for pattern in "${FORBIDDEN[@]}"; do
  if [[ "$COMMAND" =~ $pattern ]]; then
    echo "🔴 BLOCKED: $pattern" >&2
    exit 2
  fi
done
exit 0
