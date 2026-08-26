#!/usr/bin/env bash
# t229 retrospective sweep — find every recorded audit_multi per-backend verdict
# table and extract its codex row (verdict field vs body judgment).
#
# Scope is stated explicitly so an absence claim is bounded (feedback_absence_claim_needs_named_scope):
#   - the primary checkout's .moai/reports/
#   - every .claude/worktrees/*/.moai/reports/
#   - .moai/state/audit-multi/*.json (the DQ-1 persisted ConvergenceResult)
# Deduped by content hash, because worktrees carry copies of the same report.
#
# PR #1663 review hardening: ROOT defaults to the repository root derived from
# this script's own location (never a developer-specific absolute path); the
# hash tool is selected up front and the script exits when none is available
# or a hash fails — a silently empty hash would mark later files as duplicates.
set -euo pipefail

ROOT="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)}"
OUT="${2:-/dev/stdout}"

if command -v sha256sum >/dev/null 2>&1; then
  hash_one() { sha256sum "$1" | cut -c1-12; }
elif command -v shasum >/dev/null 2>&1; then
  hash_one() { shasum -a 256 "$1" | cut -c1-12; }
else
  echo "retro-sweep: neither sha256sum nor shasum is available" >&2
  exit 1
fi

{
  echo "# t229 retrospective sweep"
  echo
  echo "scanned-at-root: $ROOT"
  echo

  echo "## A. persisted ConvergenceResult state files (.moai/state/audit-multi/)"
  n_state=0
  while IFS= read -r f; do
    n_state=$((n_state + 1))
    echo "- $f"
  done < <(find "$ROOT/.moai/state/audit-multi" "$ROOT"/.claude/worktrees/*/.moai/state/audit-multi -name '*.json' -type f 2>/dev/null)
  echo "count: $n_state"
  echo

  echo "## B. report files carrying a per-backend verdict table"
  echo
  echo "| sha256(12) | path | codex row |"
  echo "|---|---|---|"
  seen=""
  while IFS= read -r f; do
    grep -qE '^\|[[:space:]]*codex[[:space:]]*\|' "$f" 2>/dev/null || continue
    if ! h=$(hash_one "$f"); then
      echo "retro-sweep: hash failed: $f" >&2
      exit 1
    fi
    [ -n "$h" ] || { echo "retro-sweep: empty hash: $f" >&2; exit 1; }
    case " $seen " in *" $h "*) continue ;; esac
    seen="$seen $h"
    row=$(grep -E '^\|[[:space:]]*codex[[:space:]]*\|' "$f" | head -1 | tr -d '\n')
    echo "| $h | ${f#"$ROOT"/} | ${row//|/\\|} |"
  done < <(grep -rl '백엔드별 판정' "$ROOT/.moai/reports" "$ROOT"/.claude/worktrees/*/.moai/reports 2>/dev/null)
} >"$OUT"
