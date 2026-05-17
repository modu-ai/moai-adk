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

# Plan — SPEC-WORKTREE-DOCS-001

## 1. Plan Summary

### 1.1 Strategy

Single-Wave documentation-only edit. The 6 file edits + 2 memory-file edits are tightly
coupled (all share the same L1/L2/L3 vocabulary + soft-mandate conversion logic) and
do NOT span Go code, tests, or build outputs. A single run-PR is the most defensible
unit because the conversion is mechanical (HARD → SHOULD) and the terminology
glossary is the single source of truth that every edit references.

### 1.2 Base branch

- `main` (HEAD at SPEC-PR creation time, currently `41b6f37dc` per `git log -1` on main).
- Plan-PR branch: `plan/SPEC-WORKTREE-DOCS-001` (current branch when this plan is being
  authored).
- Run-PR branch: `feat/SPEC-WORKTREE-DOCS-001-doctrine-rewrite` (created after plan-PR
  merge per CLAUDE.local.md §18.2 naming convention).
- Sync-PR branch: `sync/SPEC-WORKTREE-DOCS-001` (created after run-PR merge).

### 1.3 Working location

- Main checkout (`/Users/goos/MoAI/moai-adk-go`). Per the user's 2026-05-17 re-defined
  policy, neither `moai worktree new` nor `--worktree` flag is used for this SPEC.
- All three phases (plan / run / sync) execute on `feat/SPEC-WORKTREE-DOCS-001-*` and
  `sync/SPEC-WORKTREE-DOCS-001` feature branches in main checkout.

### 1.4 PR sequence

1. **Plan-PR** (this artifact): branch `plan/SPEC-WORKTREE-DOCS-001` → squash merge into main.
2. **Run-PR**: branch `feat/SPEC-WORKTREE-DOCS-001-doctrine-rewrite` → squash merge into main.
3. **Sync-PR**: branch `sync/SPEC-WORKTREE-DOCS-001` → squash merge into main.

## 2. Wave Decomposition

This is a single-Wave SPEC. No Wave-1 / Wave-2 split needed because:

- 6 files edited are documentation-only (no Go code, no test churn).
- Net LOC delta < 100 lines across all files.
- Edits are tightly coupled by the shared terminology glossary; splitting them would
  cause cross-PR readability breakage.
- All edits land in a single run-PR for atomic doctrine rollout.

## 3. Tasks

The run-phase implements the 5 EARS requirements through 8 sequential tasks.

### T1 — Author L1/L2/L3 Terminology Glossary in worktree-integration.md

**Deliverable**: A new § Terminology Glossary block inserted into
`worktree-integration.md` § Overview (after the Claude Code Native / MoAI Worktree
descriptive bullet pairs, before the § Comparison Table). The glossary table contains
4 rows: L1, L2, L3, git worktree — each with "what it does", "who owns", "when used"
columns.

**Verification**: `grep -n "Terminology Glossary" .claude/rules/moai/workflow/worktree-integration.md`
returns 1 match. `grep -n "^| L1\|^| L2\|^| L3\|^| git worktree" worktree-integration.md`
returns 4 matches.

**Implements**: REQ-WTD-002.

### T2 — Apply Terminology Prefixes Across Worktree Prose

**Deliverable**: In each of the 5 affected rule files + `CLAUDE.md` §14, replace
ambiguous "worktree" mentions with L1/L2/L3 disambiguation either by prefix
(e.g., "L2 SPEC worktree") or parenthetical (e.g., "worktree (L1)") per REQ-WTD-002
clauses (a)/(b). Where the term is the underlying git mechanism, use "git worktree"
or keep unprefixed only if context is unambiguous.

**Verification**: Run `scripts/audit-workflow-terminology.sh` (NEW, T0 prerequisite, see
T0 below) — exits 0 if every "worktree" occurrence in the 6 affected files is either
within the glossary (T1), in code/CLI/config strings, or properly prefixed. Returns
non-zero on bare ambiguous occurrences.

**Implements**: REQ-WTD-002.

### T3 — Convert spec-workflow.md Step 2/3 [HARD] → SHOULD

**Deliverable**: Edit `spec-workflow.md` § SPEC Phase Discipline table rows for Step 2
and Step 3; soften the prose "[HARD] Step ordering rules" bullets covering plan-vs-run
and run-vs-sync; soften the prose "[HARD] Anti-patterns" bullets; add a new note
"[2026-05-17 user policy] Worktree usage is opt-in per `feedback_worktree_autonomous`.
Default flow uses feature branch on main checkout." Edit the `## Run Phase` section's
opening `[HARD]` line and the `## Sync Phase` section's opening `[HARD]` line.

**Verification**:
- `grep -c '^\[HARD\]' .claude/rules/moai/workflow/spec-workflow.md` returns the
  pre-edit baseline minus 5 (3 removed in Phase Discipline + 1 in Run Phase intro +
  1 in Sync Phase intro). Snapshot pre-edit count: `git show
  origin/main:.claude/rules/moai/workflow/spec-workflow.md | grep -c '^\[HARD\]'`.
- `grep -c "user opted in\|opt-in\|opt in" .claude/rules/moai/workflow/spec-workflow.md`
  returns at least 2 (post-edit).
- The phrase "feedback_worktree_autonomous" appears at least once.

**Implements**: REQ-WTD-001, REQ-WTD-003.

### T4 — Update CLAUDE.md §14 Worktree Isolation Rules

**Deliverable**: Edit `CLAUDE.md` lines 438-446 (§14 § Worktree Isolation Rules
[HARD] block). Soften 4 `[HARD]` bullets per REQ-WTD-003. Add 1 explanatory line:
"When `Agent(isolation: "worktree")` is set, Claude Code runtime decides whether to
isolate. MoAI orchestrator does not mandate isolation. Per user policy 2026-05-17,
L2/L3 worktree usage is user opt-in." Preserve the existing cross-reference to
`worktree-integration.md`.

**Verification**:
- `grep -c '^- \[HARD\].*worktree' CLAUDE.md` for §14 block returns 0 (previously 4).
- `grep -n "Per user policy 2026-05-17" CLAUDE.md` returns 1 match.

**Implements**: REQ-WTD-001, REQ-WTD-003.

### T5 — Update worktree-state-guard.md (Dormant Note + Terminology)

**Deliverable**: Add a 2-line note to `worktree-state-guard.md` § Overview (after the
first paragraph, before "This rule defines when and how..."): "**Operational status
(2026-05-17)**: Primitive is dormant by default. It activates only when Claude Code
runtime opts into L1 isolation OR the user manually invokes `moai worktree
{snapshot,verify,restore}` from an agent prompt. Wave 5 orchestrator wiring remains
out-of-scope (forensic-audit items 1-6, deferred)." Apply L1/L2 prefixes to remaining
"worktree" mentions in the body where ambiguous.

**Verification**:
- `grep -n "Operational status (2026-05-17)" worktree-state-guard.md` returns 1 match.
- `grep -n "dormant" worktree-state-guard.md` returns at least 1 match.

**Implements**: REQ-WTD-001, REQ-WTD-002, REQ-WTD-005.

### T6 — Update session-handoff.md Block 0 Conditional Wording

**Deliverable**: Edit `session-handoff.md` lines 152 + 199-201 ("[HARD] When the
SPEC was initialized via `/moai plan --worktree`..." and the single-session vs
multi-session decision paragraph). Clarify that Block 0 is required ONLY when the
user explicitly opted into L3 (`--worktree` flag). For default `--branch` or no-flag
flow (which is now the default per 2026-05-17 policy), the standard 6-block structure
suffices. Add cross-reference to `feedback_worktree_autonomous` memory.

**Verification**:
- `grep -n "Block 0 is REQUIRED only when SPEC was initialized with .--worktree." session-handoff.md`
  returns 1 match (existing text, preserved).
- `grep -n "feedback_worktree_autonomous\|2026-05-17 user policy" session-handoff.md`
  returns at least 1 match (new cross-reference).

**Implements**: REQ-WTD-001, REQ-WTD-005.

### T7 — Create feedback_worktree_autonomous.md Memory File

**Deliverable**: Create
`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_autonomous.md`
with frontmatter:

```yaml
---
name: feedback-worktree-autonomous
description: Worktree opt-in policy (2026-05-17). L1 Claude Code autonomous, L2/L3 user opt-in. Supersedes feedback-worktree-never-use.
metadata:
  type: feedback
---
```

Body MUST include:
- Rule: L1/L2/L3 layer matrix with opt-in/autonomous designation.
- **Why:** The 2026-05-15 "worktree 영구 미사용" directive proved too restrictive when
  Claude Code runtime needs to make per-call isolation decisions (e.g., parallel
  team-mode read-only tasks where ephemeral L1 worktrees are still beneficial).
  User wants moai to stop *mandating* worktree usage while preserving Claude Code
  runtime's autonomy and the user's opt-in path.
- **How to apply:** For new SPECs, default to main-checkout + feature-branch flow.
  Do not use `moai worktree new` or `--worktree` unless user explicitly requests.
  When spawning `Agent(isolation: "worktree")`, accept whatever Claude Code runtime
  decides — do not enforce the isolation flag.
- **Effective from**: 2026-05-17.
- **Supersedes**: `feedback_worktree_never_use.md` (2026-05-15) — that file's
  blanket prohibition is rescinded.
- **Related memories**: `[[feedback-worktree-never-use]]` (superseded),
  `[[feedback-worktree-flag]]` (historical).

**Verification**:
- File exists at the specified absolute path.
- `head -8 file | grep -c "^name:\|^description:\|^metadata:\|^  type:"` returns 4.
- File contains the substrings "Why:", "How to apply:", "Effective from", "Supersedes".

**Implements**: REQ-WTD-004.

### T8 — Update MEMORY.md Index + Patch Superseded Entry

**Deliverable**: Edit
`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md`:

1. Add a new one-line index entry under "## Pending Design Proposals" (or the closest
   appropriate section if a "## Feedback" section exists in the truncated full file):
   `- [Worktree Opt-In Policy](./feedback_worktree_autonomous.md) - 2026-05-17: L1 autonomous, L2/L3 user opt-in, supersedes feedback_worktree_never_use.`
2. Locate any existing index entry referencing `feedback_worktree_never_use.md` and
   prepend `[SUPERSEDED by feedback_worktree_autonomous]` to its description line.
   If no such entry exists in the current MEMORY.md, no patching is needed — only
   the new entry is added.

Also: edit the first heading of `feedback_worktree_never_use.md` body (line 10 of
that file) to prepend `[SUPERSEDED by feedback_worktree_autonomous]` per Lessons
Protocol.

**Verification**:
- `grep -c "feedback_worktree_autonomous" MEMORY.md` returns at least 1.
- `head -15 feedback_worktree_never_use.md | grep -c "SUPERSEDED"` returns at least 1.

**Implements**: REQ-WTD-004.

### T0 (Prerequisite) — Author Terminology Audit Script

Run-phase will also create `scripts/audit-workflow-terminology.sh` to support T2
verification. Script greps the 6 affected files for bare "worktree" occurrences not
within a prefixed compound or glossary block; exits non-zero on findings. This is a
small bash script (~30 LOC) added under `scripts/` (not under `.moai/`).

NOTE: T0 is implementation-phase scaffolding; the script is not required by the EARS
contract — it merely supports AC-WTD-006 (terminology audit gate). The script is
SHOULD-have, not MUST-have. If time-constrained, replace with manual `grep -rn`
verification.

## 4. Risk Analysis

### R1 — Stale Cross-References After [HARD] Removal

**Risk**: Other rule files (e.g., `.claude/skills/moai/workflows/*.md`,
`.claude/skills/moai/team/*.md`) may cross-reference the soon-to-be-softened [HARD]
rules in `spec-workflow.md`. If they cite "[HARD] Step 2 MUST create worktree" verbatim,
they will become stale.

**Mitigation**: T2 verification script (`scripts/audit-workflow-terminology.sh`) will
flag any file outside the 6 affected files that uses the exact phrase "MUST create
worktree" or "MUST use isolation: worktree". If found, those files become candidates
for a follow-up SPEC (out-of-scope for this SPEC). A grep audit is documented in
acceptance.md AC-WTD-007.

### R2 — Wave 5 Primitive Activation Regression

**Risk**: Softening worktree-state-guard.md mandates could be misread as deprecating
the primitive itself. Future agents might delete the Go CLI code.

**Mitigation**: T5 deliverable explicitly states "Primitive is dormant by default"
and preserves the documentation. Acceptance test AC-WTD-005 verifies the Go CLI
binary still exists post-edit. design.md AD-002 documents the explicit decision to
preserve the primitive.

### R3 — Memory Hierarchy Race Condition

**Risk**: Multiple plan/run/sync sessions reading MEMORY.md simultaneously while T8
patches the file could read inconsistent state.

**Mitigation**: T8 is a single-file atomic write. MEMORY.md is auto-loaded but its
updates are read-eventually-consistent across sessions. If a session starts in the
middle of T8, it reads either pre-T8 or post-T8 — no torn state because shell-level
file writes are atomic for this size.

### R4 — Backward Compat Drift in CI

**Risk**: CI tests in `internal/cli/launcher_*` or `internal/template/agentless_audit_test.go`
might encode the [HARD] worktree rules as text-scan assertions and would fail after
T3/T4 softening.

**Mitigation**: Run `go test ./...` as part of acceptance gate (AC-WTD-009). If any
test fails on the softened rule text, the SPEC's scope expands to include test
updates in run-PR. design.md AD-003 documents the fail-fast assumption.

### R5 — User Policy Re-reversal

**Risk**: User may change policy direction again (e.g., back to mandatory worktree).
The SPEC artifact would become stale.

**Mitigation**: Documentation includes a 1-line "Effective from: 2026-05-17" marker
on every soft-mandate line. Reversal would be a new SPEC, not an edit. Memory file
`feedback_worktree_autonomous.md` carries the date prominently so re-reversal is
detectable via memory-hierarchy diff.

## 5. Mitigation Summary

All 5 risks have explicit mitigations documented in acceptance.md and design.md. No
risk has a probability/impact pairing high enough to warrant deferring the SPEC.
R1 (stale cross-refs) and R4 (CI text-scan) are the highest-impact risks; both are
caught by the acceptance gate before run-PR merge.

## 6. Implementation Sequence

```
Plan-PR (this artifact):
  branch: plan/SPEC-WORKTREE-DOCS-001
  files: 5 markdown artifacts in .moai/specs/SPEC-WORKTREE-DOCS-001/
  merge: squash → main
  status: this SPEC stays "draft" in frontmatter until plan-PR merge
          (status transitions to "planned" via spec-status-sync workflow post-merge)

Run-PR:
  branch: feat/SPEC-WORKTREE-DOCS-001-doctrine-rewrite
  files: 6 file edits (5 rules + CLAUDE.md) + 2 memory edits + 1 script
  task order: T0 (script) → T1 → T2 (audit) → T3 → T4 → T5 → T6 → T7 → T8
  gates: acceptance.md AC-WTD-001..009 all PASS
  merge: squash → main

Sync-PR (optional, minimal scope):
  branch: sync/SPEC-WORKTREE-DOCS-001
  files: status frontmatter transitions, lessons.md entry capture, CHANGELOG.md row
  merge: squash → main
```

## 7. Out of Scope

Mirrored from spec.md §3 Non-Goals. Explicit list:

- VERIFY-001 (Wave 5 orchestrator wiring) — DROPPED.
- Forensic audit items 13-27 (deferred).
- L1 Claude Code runtime field semantics investigation (outside moai scope).
- `.claude/rules/moai/project-overrides/` directory mechanism (deferred per AD-002).
- Recovery procedure for "user opted into `--worktree` but wants to abandon mid-flight" (deferred).
