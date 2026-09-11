#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
auditor="$root/.claude/agents/moai/plan-auditor.md"
grep -q 'The script never emits BLOCKING itself' "$auditor"
grep -q 'auditor decides BLOCKING' "$auditor"

fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT
mkdir -p "$fixture/.moai/specs/SPEC-OLD-001"
printf '%s\n' 'status: retired' > "$fixture/.moai/specs/SPEC-OLD-001/spec.md"
printf '%s\n' 'This change superseded SPEC-OLD-001 and absorbs its behavior.' > "$fixture/new-spec.md"

(
cd "$fixture"
grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' new-spec.md | sort -u | while read -r sid; do
  status=$(grep '^status:' ".moai/specs/$sid/spec.md" | head -1 | cut -d: -f2 | tr -d ' ')
  case "$status" in
    retired|superseded|archived)
      echo "REVIEW: $sid has status=$status"
      awk -v sid="$sid" 'BEGIN { RS = "" } index($0, sid) && tolower($0) ~ /revers|supersed|absorb|carve-out/ { gsub(/\n/, " "); print "  reconciliation candidate: " $0 }' new-spec.md
      ;;
  esac
done
) > "$fixture/output"
output=$(cat "$fixture/output")
grep -q 'REVIEW: SPEC-OLD-001' <<< "$output"
grep -q 'reconciliation candidate' <<< "$output"
if grep -q 'BLOCKING:' <<< "$output"; then
  printf 'FAIL: discovery script emitted a verdict instead of a review candidate\n' >&2
  exit 1
fi
printf 'PASS: D7 emits review candidates and leaves blocking judgment to the auditor\n'
