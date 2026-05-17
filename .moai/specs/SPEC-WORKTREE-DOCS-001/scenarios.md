---
id: SPEC-WORKTREE-DOCS-001
title: "Worktree Workflow Documentation Harmonization (L1/L2/L3 Opt-In Policy)"
version: "0.1.0"
status: draft
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, documentation, harmonization, override, opt-in, terminology"
---

# Scenarios — SPEC-WORKTREE-DOCS-001

Wave-by-Wave test plans. This is a single-Wave SPEC; all scenarios apply to the
run-PR validation gate.

## 1. Smoke Test — No Force-Worktree Mandate Remains

**Scenario**: After the run-PR edits are applied to main, no rule file under
`.claude/rules/` (and no top-level rule like `CLAUDE.md` §14) contains a `[HARD]`
mandate that forces creation, reuse, or activation of a worktree.

**Given**:
- Run-PR `feat/SPEC-WORKTREE-DOCS-001-doctrine-rewrite` is merged into main.
- Working tree on main checkout.

**When**:
```bash
grep -rEn '\[HARD\][^.]*(MUST|must)[^.]*worktree' .claude/rules/ CLAUDE.md
```

**Then**:
- Command MUST return 0 matching lines.

**Failure mode**: If the grep returns non-zero matches, identify each match and
apply T3 or T4 again. Re-run gate.

**Cross-reference**: AC-WTD-001.

## 2. Terminology Audit — Bare "worktree" Mentions Are Disambiguated

**Scenario**: Every prose mention of "worktree" in the 6 affected files is either
prefixed with L1/L2/L3 / "git", parenthetically clarified, in a code/CLI literal, or
inside the terminology glossary block in `worktree-integration.md`.

**Given**:
- Run-PR is merged.
- `scripts/audit-workflow-terminology.sh` exists (T0).

**When**:
```bash
scripts/audit-workflow-terminology.sh \
  .claude/rules/moai/workflow/spec-workflow.md \
  .claude/rules/moai/workflow/worktree-integration.md \
  .claude/rules/moai/workflow/worktree-state-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  CLAUDE.md
```

**Then**:
- Script exits 0 (no ambiguous occurrences).
- Manual spot-check on 3 random matches confirms readability is preserved.

**Manual fallback**:
```bash
grep -rEn '(^|[^a-zA-Z0-9\-_/.])worktree' <files> \
  | grep -vE '(L1 worktree|L2 worktree|L3 worktree|git worktree|SPEC worktree|Plan worktree|Claude Code Native|Terminology Glossary)'
```
Remaining matches MUST all be in code/literal/heading contexts.

**Failure mode**: For each remaining bare mention, apply T2 (add a prefix or
parenthetical). Re-run audit.

**Cross-reference**: AC-WTD-002, AC-WTD-003.

## 3. User Policy Linkage Test — 2026-05-17 Reference Present

**Scenario**: At least 4 of the 5 affected rule files explicitly reference the
2026-05-17 user policy or the new `feedback_worktree_autonomous` memory file.

**Given**: Run-PR merged.

**When**:
```bash
grep -rln "2026-05-17\|feedback_worktree_autonomous" \
  .claude/rules/moai/workflow/spec-workflow.md \
  .claude/rules/moai/workflow/worktree-integration.md \
  .claude/rules/moai/workflow/worktree-state-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  CLAUDE.md
```

**Then**:
- Command returns at least 4 file paths.

**Failure mode**: For each missing file, add a cross-reference line per T3/T4/T5/T6.

**Cross-reference**: AC-WTD-004.

## 4. Memory Integrity Test — MEMORY.md + Supersede Marker

**Scenario**: The new `feedback_worktree_autonomous.md` memory file exists with the
correct frontmatter, MEMORY.md indexes it, and the superseded
`feedback_worktree_never_use.md` is marked.

**Given**: T7 + T8 of run-PR completed.

**When**:
```bash
F1=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_autonomous.md
F2=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_never_use.md
F3=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md

test -f "$F1" && \
  grep -q "^name: feedback-worktree-autonomous" "$F1" && \
  grep -q "^description:" "$F1" && \
  grep -q "^  type: feedback" "$F1" && \
  grep -q "Why:" "$F1" && \
  grep -q "How to apply:" "$F1" && \
  grep -q "Effective from" "$F1" && \
  head -15 "$F2" | grep -q "SUPERSEDED" && \
  grep -q "feedback_worktree_autonomous" "$F3"

echo "exit_status=$?"
```

**Then**:
- All 9 grep gates inside the `&&` chain return 0.
- Final `exit_status` is 0.

**Failure mode**: For each missing condition, return to T7 (memory file authorship)
or T8 (index patching) and complete the step. Re-verify.

**Cross-reference**: AC-WTD-006.

## 5. Wave 5 Primitive Preservation — Go Code Untouched

**Scenario**: This SPEC is documentation-only. The Wave 5 primitive's Go code is
unchanged. Tests pass.

**Given**: Run-PR merged.

**When**:
```bash
test -f internal/cli/worktree/guard.go
git log --oneline -5 -- internal/cli/worktree/guard.go internal/cli/worktree/guard_test.go
go test ./internal/cli/worktree/...
```

**Then**:
- `guard.go` file exists.
- `git log` shows no commits from `feat/SPEC-WORKTREE-DOCS-001-*` branch in the
  per-file history.
- `go test` exits 0 for the worktree CLI package.

**Failure mode**: If the SPEC accidentally modified Go code, revert that hunk in
the run-PR. Re-run test.

**Cross-reference**: AC-WTD-005, AC-WTD-009.

## 6. Spec Lint Clean

**Scenario**: The 5 plan-phase artifacts pass strict spec lint with zero findings.

**Given**: Plan-PR (this branch) artifacts in place.

**When**:
```bash
moai spec lint --strict
```

**Then**:
- Output contains `✓ No findings` (or equivalent zero-finding indicator) for
  `SPEC-WORKTREE-DOCS-001`.
- Exit code 0.

**Failure mode**: Read findings; correct frontmatter (likely created/updated/tags
canonical naming), EARS keyword usage, or other lint rules. Re-run.

**Cross-reference**: AC-WTD-008.

## 7. Cross-SPEC Coupling — No Conflict with Bundle F

**Scenario**: Bundle F SPEC-V3R4-WORKFLOW-SPLIT-001 is currently in-progress and may
edit `spec-workflow.md` during Wave 1. Our run-PR must rebase against latest main
before T3 begins.

**Given**: Bundle F Wave 1 may merge before our run-PR.

**When**: Before T3 of run-phase:
```bash
git fetch origin main
git rebase origin/main
```

**Then**:
- Either no conflict (Bundle F has not landed yet, or its edits are in
  non-overlapping sections), OR
- Conflict in `spec-workflow.md` is resolved by applying our T3 edits to the
  appropriate post-split sub-skill file location.

**Failure mode**: Surface conflict to user via the run-phase agent's blocker report
mechanism. Do NOT auto-resolve worktree-related rule conflicts.

**Cross-reference**: design.md § CC-4.

## 8. Out of Scope (mirror)

Items explicitly NOT tested by these scenarios (deferred):

- VERIFY-001 (orchestrator-side snapshot/verify/restore wiring) — DROPPED.
- L1 Claude Code runtime `worktreePath` field semantics — out of moai scope.
- `.claude/rules/moai/project-overrides/` directory mechanism — deferred per
  design.md AD-002.
- Recovery procedure for partial `--worktree` opt-in abandonment — deferred per
  spec.md § Non-Goals.
- Forensic audit items 13-27 — deferred to separate follow-up SPECs.
