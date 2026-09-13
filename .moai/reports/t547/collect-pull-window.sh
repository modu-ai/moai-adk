#!/usr/bin/env bash
# collect-pull-window.sh — t547 collection tool for the AC-JFM-018 pull window.
#
# Scans every tree that may hold an askuser-observations.jsonl (primary checkout,
# card worktrees under .claude/worktrees/, L2 worktrees under ~/.moai/worktrees/),
# exports every row with mode=="pull", and prints the numbers the judgment needs:
# total rows, label_present violations, per-session split (input to the calls_issued
# four-way contrast), and the collection interval.
#
# Denominator rule (design.md §6.3): every recorded row with mode=="pull".
# No question_type filter. The pre-window push rows are excluded structurally —
# the observer stamps each row with the asking session's mode at question time,
# so no row recorded before the 2026-09-13 config flip can carry mode=="pull".
#
# Usage:
#   collect-pull-window.sh [EXPORT_PATH]     export + measure (default export:
#                                            <primary>/.moai/reports/t401/pull-window.jsonl)
#   collect-pull-window.sh --selftest        positive-control fixture check (no scans)
set -euo pipefail

MIN_ROWS=20   # AC-JFM-018 floor: n >= 20, below is a gap, never a pass

selftest() {
  local dir
  dir="$(mktemp -d)"
  printf '%s\n' \
    '{"mode":"pull","label_present":false,"session_id":"s1","timestamp":"2026-09-13T23:30:00Z","question_count":1,"option_count":3,"payload_parsed":true}' \
    '{"mode":"pull","label_present":true,"session_id":"s1","timestamp":"2026-09-13T23:31:00Z","question_count":1,"option_count":3,"payload_parsed":true}' \
    '{"mode":"push","label_present":true,"session_id":"s2","timestamp":"2026-09-10T09:00:00Z","question_count":1,"option_count":3,"payload_parsed":true}' \
    > "$dir/log.jsonl"
  jq -c 'select(.mode=="pull")' "$dir/log.jsonl" > "$dir/export.jsonl"
  local rows viol rc
  rows="$(wc -l < "$dir/export.jsonl" | tr -d ' ')"
  viol="$(jq -s '[.[] | select(.label_present==true)] | length' "$dir/export.jsonl")"
  rm -rf "$dir"
  if [ "$rows" = "2" ] && [ "$viol" = "1" ]; then
    echo "selftest PASS: fixture 3 rows (2 pull, 1 push) -> export rows=2 violations=1"
    rc=0
  else
    echo "selftest FAIL: rows=$rows violations=$viol (expected 2 / 1) — filter is broken, do not trust its output"
    rc=1
  fi
  return "$rc"
}

if [ "${1:-}" = "--selftest" ]; then
  selftest
  exit $?
fi

common_git="$(git rev-parse --path-format=absolute --git-common-dir)"
primary="$(dirname "$common_git")"
out="${1:-$primary/.moai/reports/t401/pull-window.jsonl}"

list="$(mktemp)"
export_tmp="$(mktemp)"
trap 'rm -f "$list" "$export_tmp"' EXIT

# One entry per candidate log file; sort -u so a tree is never scanned twice.
{
  printf '%s\n' "$primary/.moai/logs/askuser-observations.jsonl"
  find "$primary/.claude/worktrees" "$HOME/.moai/worktrees" \
    -maxdepth 5 -name askuser-observations.jsonl 2>/dev/null
} | sort -u > "$list"

: > "$export_tmp"
while IFS= read -r f; do
  [ -f "$f" ] || continue
  jq -c 'select(.mode=="pull")' "$f" >> "$export_tmp" 2>/dev/null || true
done < "$list"

mkdir -p "$(dirname "$out")"
cp "$export_tmp" "$out"

rows="$(wc -l < "$out" | tr -d ' ')"
if [ "$rows" -eq 0 ]; then
  echo "export=$out"
  echo "scanned_logs=$(wc -l < "$list" | tr -d ' ')  pull_rows=0"
  echo "READING: gap — the window has no rows yet (수집 대기). n=0 is unmeasured, never violations==0."
  exit 0
fi

viol="$(jq -s '[.[] | select(.label_present==true)] | length' "$out")"
first="$(head -1 "$out" | jq -r '.timestamp')"
last="$(tail -1 "$out" | jq -r '.timestamp')"

echo "export=$out"
echo "scanned_logs=$(wc -l < "$list" | tr -d ' ')"
echo "rows_recorded=$rows  violations=$viol  window=[$first .. $last]"
echo "per-session split (rows; the input to the calls_issued four-way contrast):"
jq -s -r 'group_by(.session_id)[] | "\(length)\t\(.[0].session_id)"' "$out"
if [ "$rows" -lt "$MIN_ROWS" ]; then
  echo "READING: gap — n=$rows < $MIN_ROWS floor. A run below the floor is a gap, never a pass."
elif [ "$viol" -gt 0 ]; then
  echo "READING: FAIL — $viol row(s) carry label_present:true under pull mode."
else
  echo "READING: rows and violations cleared. Judgment still requires: provenance md with matching"
  echo "row count, calls_issued contrast == equal, and AC-JFM-023 green beforehand (acceptance.md)."
fi
