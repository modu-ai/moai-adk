#!/bin/bash
# t518 M-B1 — the control pair at the real CLI surface, after the repair.
#
# Built from source on purpose: the installed ~/go/bin/moai is far behind this
# tree and measuring with it produced a wrong baseline earlier in this card.
#
# Every exit code is read on its own line, with no pipe anywhere: a piped $?
# reports the last stage of the pipeline, not the command.
set -u
WT=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518
cd "$WT" || exit 1
OUT="$WT/.moai/reports/t518/mb1-control-pair.txt"
BIN=/tmp/t518-moai

go build -o "$BIN" ./cmd/moai
echo "build rc=$?" > "$OUT"
{
  echo "tree: $WT"
  echo "HEAD: $(git rev-parse HEAD)"
  echo
} >> "$OUT"

run_case() {
  local label="$1"
  shift
  local tmp=/tmp/t518-case.txt
  "$BIN" "$@" > "$tmp" 2>&1
  local rc=$?
  {
    echo "=============================================================="
    echo "[$label] moai $*"
    echo "rc=$rc   (read on its own line, no pipe)"
    echo "--- output (last 6 lines) ---"
    tail -6 "$tmp"
    echo
  } >> "$OUT"
  echo "[$label] rc=$rc"
}

S=SPEC-SPEC-LINT-ID-ARG-001
run_case "1 ID form"            spec lint "$S"
run_case "2 path form (CONTROL)" spec lint ".moai/specs/$S/spec.md"
run_case "3 directory form"      spec lint ".moai/specs/$S"
run_case "4 absent ID"           spec lint SPEC-NOPE-999
run_case "5 SPEC-A-1 (anchor)"   spec lint SPEC-A-1
run_case "6 missing path"        spec lint ./missing/spec.md
run_case "7 mixed"               spec lint "$S" ".moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/spec.md"

echo "DONE"
