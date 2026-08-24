#!/usr/bin/env bash
# t229 retrospective sweep — find every recorded audit_multi per-backend verdict
# table and extract its codex row (verdict field vs body judgment).
#
# Scope is stated explicitly so an absence claim is bounded (feedback_absence_claim_needs_named_scope):
#   - the primary checkout's .moai/reports/
#   - every .claude/worktrees/*/.moai/reports/
#   - .moai/state/audit-multi/*.json (the DQ-1 persisted ConvergenceResult)
# Deduped by content hash, because worktrees carry copies of the same report.
set -uo pipefail

ROOT="${1:-/Users/goos/MoAI/moai-adk-go}"
OUT="${2:-/dev/stdout}"

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
  done < <(find "$ROOT" -path '*/.moai/state/audit-multi/*.json' -type f 2>/dev/null)
  echo "count: $n_state"
  echo

  echo "## B. report files carrying a per-backend verdict table"
  echo
  echo "| sha256(12) | path | codex row |"
  echo "|---|---|---|"
  seen=""
  while IFS= read -r f; do
    grep -qE '^\|[[:space:]]*codex[[:space:]]*\|' "$f" 2>/dev/null || continue
    h=$(shasum -a 256 "$f" | cut -c1-12)
    case " $seen " in *" $h "*) continue ;; esac
    seen="$seen $h"
    row=$(grep -E '^\|[[:space:]]*codex[[:space:]]*\|' "$f" | head -1 | tr -d '\n')
    echo "| $h | ${f#"$ROOT"/} | ${row//|/\\|} |"
  done < <(grep -rl '백엔드별 판정' "$ROOT/.moai/reports" "$ROOT"/.claude/worktrees/*/.moai/reports 2>/dev/null)
} >"$OUT"
