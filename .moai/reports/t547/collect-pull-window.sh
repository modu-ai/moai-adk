#!/usr/bin/env bash
# collect-pull-window.sh — t547 collection tool for the AC-JFM-018 pull window.
#
# Admission rules (verdict.md §9.3 — all must hold):
#   1. mode=="pull" and timestamp >= ANCHOR
#   2. the row's log lives in a NON-primary tree (.claude/worktrees/*, ~/.moai/worktrees/**)
#   3. that tree's .claude/rules/moai/core/askuser-protocol.md carries the pull branch
#      (grep -c recommendation_mode >= 1)
# Rows carry no tree identity, so the tree is inferred from the log file's location
# (verdict.md §9.5). Denominator: every admitted row — no question_type filter.
#
# Usage:
#   collect-pull-window.sh                   measure only (never writes an export below n>=20)
#   collect-pull-window.sh [EXPORT_PATH]     measure; write the export only when n >= MIN_ROWS
#                                            (default <primary>/.moai/reports/t401/pull-window.jsonl)
#   collect-pull-window.sh --selftest        positive/negative-control fixture check (no scans)
set -euo pipefail

MIN_ROWS=20                       # AC-JFM-018 floor: n >= 20, below is a gap, never a pass
ANCHOR="2026-09-17T16:15:27Z"     # verdict.md §9.2 — operator disposition (a)
RULES_REL=".claude/rules/moai/core/askuser-protocol.md"
LOG_REL=".moai/logs/askuser-observations.jsonl"

# admitted_rows <tree> <primary>: print the admitted rows of one tree's log.
admitted_rows() {
  local tree="${1%/}" primary="${2%/}"
  [ "$tree" = "$primary" ] && return 0                                  # rule 2
  [ -f "$tree/$LOG_REL" ] || return 0
  [ "$(grep -c recommendation_mode "$tree/$RULES_REL" 2>/dev/null || true)" -ge 1 ] 2>/dev/null || return 0  # rule 3
  jq -c --arg a "$ANCHOR" 'select(.mode=="pull" and .timestamp >= $a)' "$tree/$LOG_REL"  # rule 1
}

selftest() {
  local dir rows viol rc
  dir="$(mktemp -d)"
  mkrow() { printf '{"mode":"%s","label_present":%s,"session_id":"%s","timestamp":"%s","question_count":1,"option_count":3,"payload_parsed":true}\n' "$@"; }
  mktree() { mkdir -p "$dir/$1/.moai/logs" "$dir/$1/$(dirname "$RULES_REL")"; }
  # primary: pull-aware rules, post-anchor pull row -> excluded by rule 2
  mktree primary; echo "recommendation_mode" > "$dir/primary/$RULES_REL"
  mkrow pull true p1 2026-09-18T00:00:00Z > "$dir/primary/$LOG_REL"
  # wt-ok: pull-aware worktree
  mktree wt-ok; echo "recommendation_mode" > "$dir/wt-ok/$RULES_REL"
  { mkrow pull false s1 2026-09-18T00:00:00Z   # admitted
    mkrow pull true  s1 2026-09-18T00:01:00Z   # admitted, violation
    mkrow pull true  s1 2026-09-17T16:15:26Z   # before anchor -> excluded (rule 1)
    mkrow push true  s2 2026-09-18T00:02:00Z   # push -> excluded (rule 1)
  } > "$dir/wt-ok/$LOG_REL"
  # wt-old: rules without the pull branch -> excluded by rule 3
  mktree wt-old; echo "no pull branch here" > "$dir/wt-old/$RULES_REL"
  mkrow pull true o1 2026-09-18T00:00:00Z > "$dir/wt-old/$LOG_REL"

  : > "$dir/export.jsonl"
  for t in primary wt-ok wt-old; do admitted_rows "$dir/$t" "$dir/primary" >> "$dir/export.jsonl"; done
  rows="$(wc -l < "$dir/export.jsonl" | tr -d ' ')"
  viol="$(jq -s '[.[] | select(.label_present==true)] | length' "$dir/export.jsonl")"
  rm -rf "$dir"
  if [ "$rows" = "2" ] && [ "$viol" = "1" ]; then
    echo "selftest PASS: fixture 7 rows over 3 trees (primary / pull-aware wt / pull-unaware wt; pre-anchor + push decoys) -> admitted rows=2 violations=1"
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
rows_tmp="$(mktemp)"
trap 'rm -f "$list" "$rows_tmp"' EXIT

# One tree per candidate log; sort -u so a tree is never scanned twice. primary is not listed.
# Globs, not find: a deep find over hundreds of trees does not finish in bounded time.
for tree in "$primary"/.claude/worktrees/*/ "$HOME"/.moai/worktrees/*/ "$HOME"/.moai/worktrees/*/*/; do
  if [ -f "$tree$LOG_REL" ]; then printf '%s\n' "${tree%/}"; fi
done | sort -u > "$list"

: > "$rows_tmp"
while IFS= read -r tree; do
  admitted_rows "$tree" "$primary" >> "$rows_tmp" || true
done < "$list"

rows="$(wc -l < "$rows_tmp" | tr -d ' ')"
echo "anchor=$ANCHOR  scanned_trees=$(wc -l < "$list" | tr -d ' ')  admitted_rows=$rows"
if [ "$rows" -eq 0 ]; then
  echo "READING: gap — the window has no admitted rows yet (수집 대기). n=0 is unmeasured, never violations==0."
  exit 0
fi

viol="$(jq -s '[.[] | select(.label_present==true)] | length' "$rows_tmp")"
first="$(jq -s -r 'sort_by(.timestamp) | .[0].timestamp' "$rows_tmp")"
last="$(jq -s -r 'sort_by(.timestamp) | .[-1].timestamp' "$rows_tmp")"
echo "violations=$viol  window=[$first .. $last]"
echo "per-session split (rows; the input to the calls_issued four-way contrast):"
jq -s -r 'group_by(.session_id)[] | "\(length)\t\(.[0].session_id)"' "$rows_tmp"

if [ "$rows" -lt "$MIN_ROWS" ]; then
  echo "READING: gap — n=$rows < $MIN_ROWS floor. No export written (export is forbidden below the floor)."
  exit 0
fi

mkdir -p "$(dirname "$out")"
cp "$rows_tmp" "$out"
echo "export=$out"
if [ "$viol" -gt 0 ]; then
  echo "READING: FAIL — $viol admitted row(s) carry label_present:true under pull mode."
else
  echo "READING: rows and violations cleared. Judgment still requires: provenance md with matching"
  echo "row count, calls_issued contrast == equal, and AC-JFM-023 green beforehand (acceptance.md)."
fi
