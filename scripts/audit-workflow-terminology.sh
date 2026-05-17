#!/bin/bash
# Terminology audit for SPEC-WORKTREE-DOCS-001 (REQ-WTD-002)
# Usage: ./scripts/audit-workflow-terminology.sh <file1> [file2 ...]
# Exit 0 if all worktree references are properly disambiguated; exit 1 otherwise.
set -euo pipefail

if [ $# -eq 0 ]; then
  echo "Usage: $0 <file1> [file2 ...]" >&2
  exit 1
fi

FILES=("$@")
EXIT=0

for f in "${FILES[@]}"; do
  if [ ! -f "$f" ]; then
    echo "::warning file=$f::file not found, skipping"
    continue
  fi

  # Match bare "worktree" not preceded by allowed disambiguating prefixes.
  # Allowed prefixes: L1, L2, L3, git, SPEC, Plan, Claude Code Native
  # Also allowed: code/CLI literals in backticks, heading anchors, compound nouns
  bare=$(grep -nE '(^|[^a-zA-Z0-9\-_/.`])worktree' "$f" \
    | grep -vE '(L1 worktree|L2 worktree|L3 worktree|git worktree|SPEC worktree|Plan worktree|Claude Code Native|worktree-state-guard|worktree-integration|Terminology Glossary|SPEC-to-Worktree|WorktreeCreate|WorktreeRemove)' \
    | grep -vE '`[^`]*worktree[^`]*`' \
    | grep -vE '\[HARD\]|## Worktree|^##|^# ' \
    | grep -vE 'moai worktree|moai-adk|\.moai/worktrees|\.claude/worktrees|~/.moai/worktrees|--worktree|isolation: .worktree|isolation: "worktree"' \
    | grep -vE 'WorktreeCreate|WorktreeRemove|worktree done|worktree new|worktree list|worktree remove|worktree prune|worktree add|git worktree' \
    | grep -vE '\*\*L[123]\*\*' \
    | grep -vE '^\|.*worktree' \
    || true)

  if [ -n "$bare" ]; then
    echo "::warning file=$f::bare worktree references found (review for L1/L2/L3 disambiguation):"
    echo "$bare"
    EXIT=1
  fi
done

exit $EXIT
