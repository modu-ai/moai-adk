#!/usr/bin/env bash
# AC-CSN-012 judgment command + its positive control, run verbatim as the AC specifies.
set -u
RE='SPEC-[A-Z0-9]+-[A-Z0-9-]*[0-9]{3}|\b(REQ|AC)-[A-Z0-9]+-[0-9]{3}\b|\b[0-9a-f]{7,40}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}'

echo "### judgment — the 4 unguarded files this SPEC edits"
grep -cE "$RE" \
  AGENTS.md \
  internal/template/templates/AGENTS.md \
  internal/template/templates/.claude/rules/moai/development/skill-authoring.md \
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
echo
echo "### remaining matches in skill-authoring.md (must be the two frontmatter example dates)"
grep -nE "$RE" internal/template/templates/.claude/rules/moai/development/skill-authoring.md
echo
echo "### positive control — SAME command against this SPEC's own spec.md; pass = non-zero"
grep -cE "$RE" .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md
