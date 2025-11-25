#!/bin/bash
# Preserve file permissions (from Context7 official docs)
set -euo pipefail

HOOK_TYPE="${1:-pre}"
FILE="${2:-.}"
PERMS_FILE="/tmp/perms_${FILE//\//_}.txt"

if [[ "$HOOK_TYPE" == "pre" ]]; then
  stat -c "%a %u:%g" "$FILE" > "$PERMS_FILE" 2>/dev/null || true
  exit 0
elif [[ "$HOOK_TYPE" == "post" ]]; then
  if [[ -f "$PERMS_FILE" ]]; then
    SAVED=$(cat "$PERMS_FILE")
    chmod ${SAVED%% *} "$FILE" 2>/dev/null || true
    rm "$PERMS_FILE"
  fi
  exit 0
fi
