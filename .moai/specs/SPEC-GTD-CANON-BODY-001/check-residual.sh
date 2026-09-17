#!/usr/bin/env bash
# AC-GCB-006 residual-wording check (read-only).
# Usage:
#   bash check-residual.sh              # check the 19 scoped files (run from repo root)
#   bash check-residual.sh FILE...      # check the given files instead (positive control)
# Exit: 0 PASS, 1 FAIL (offending lines printed), 2 grep error (missing/unreadable file).
# Allowed survivor: a matching line that contains the literal phrase "compat alias".
# Must run under bash (arrays).
[ -n "${BASH_VERSION:-}" ] || { echo "USAGE_ERROR: run with bash (bash check-residual.sh ...)"; exit 2; }
set -u

if [ "$#" -gt 0 ]; then
  files=("$@")
else
  files=(
    .claude/agents/moai/manager-lead.md
    .claude/rules/moai/workflow/kanban-dispatch.md
    .claude/rules/moai/workflow/kanban-dispatch-detail.md
    .claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
    .claude/skills/moai-kanban-foreman/SKILL.md
    .claude/skills/moai/SKILL.md
    .claude/skills/moai/workflows/gtd.md
    .claude/skills/moai/workflows/project/doc-generation.md
    .moai/docs/todo-queue-storage.md
    internal/template/templates/.claude/agents/moai/manager-lead.md
    internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
    internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md
    internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
    internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md
    internal/template/templates/.claude/skills/moai/SKILL.md
    internal/template/templates/.claude/skills/moai/workflows/gtd.md
    internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md
    internal/template/templates/.moai/docs/todo-queue-storage.md
    internal/template/templates/.codex/agents/moai/manager-lead.toml
  )
fi

echo "files=${#files[@]}"
out=$(grep -nE 'moai todo|/moai:todo' -- "${files[@]}")
rc=$?
if [ "$rc" -gt 1 ]; then
  echo "GREP_ERROR rc=$rc"
  exit 2
fi
bad_lines=$(printf '%s\n' "$out" | grep -vE '(^|[^A-Za-z])compat alias' | grep -E '.')
bad=0
[ -n "$bad_lines" ] && bad=$(printf '%s\n' "$bad_lines" | wc -l | tr -d ' ')
surv=0
[ -n "$out" ] && surv=$(printf '%s\n' "$out" | grep -cE '(^|[^A-Za-z])compat alias')
echo "grep_rc=$rc offending=$bad survivors=$surv"
if [ "$bad" -eq 0 ]; then
  echo PASS
  exit 0
fi
printf '%s\n' "$bad_lines"
echo FAIL
exit 1
