---
id: SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001
title: "Local Agent Namespace Consolidation — Progress Tracking"
version: "0.1.2"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.7.0"
module: ".claude/agents/local + .claude/skills/moai/workflows + internal/template/templates + .moai/docs"
lifecycle: spec-anchored
tags: "local-namespace, dev-only, agent-migration, template-refactor, claude-local-externalization, sprint-10-lane-b, thin-command-pattern"
tier: M
depends_on: []
related_specs: []
plan_commit_sha: "651623dc1"
run_commit_sha: "<pending>"
sync_commit_sha: "<pending>"
mx_commit_sha: "<pending>"
---

# Progress — SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001

## A. Lifecycle Reflection

| Phase | Status | Owner | Commit SHA | Date |
|-------|--------|-------|------------|------|
| Plan | in-progress | manager-spec | `<pending>` | 2026-05-25 |
| Plan-Audit | pending | plan-auditor | n/a | — |
| Run | not-started | manager-develop | `<pending>` | — |
| Sync | not-started | manager-docs | `<pending>` | — |
| Mx | not-started | manager-docs or orchestrator | `<pending>` | — |

## B. Milestone Status

| Milestone | Description | Status | Files Modified | Commit SHA |
|-----------|-------------|--------|----------------|------------|
| M1 | Namespace contract documentation update (agent-authoring.md + skill-authoring.md local+template mirror; dev-only-isolation.md local-only per spec.md §E) | not-started | 5 expected | `<pending>` |
| M2 | Local agent body authoring (release-update-specialist + github-specialist, local-only, NO template mirror) | not-started | 2 expected | `<pending>` |
| M3 | Dev-only skill removal + thin command rewiring (97/98 wrappers + 2 skill file deletions) | not-started | 4 expected (2 modified + 2 deleted) | `<pending>` |
| M4 | Template generic refactor — 17 leak removal across 13 template files + 13 local mirror | not-started | 26 expected | `<pending>` |
| M5 | Generic patterns guide authoring (.moai/docs/generic-patterns-guide.md, local + template mirror) | not-started | 2 expected | `<pending>` |
| M6 | progress.md backfill + verification batch + handoff to manager-docs | not-started | 1 expected (this file) | `<pending>` |

Total expected file count across M1-M6: ~40 files (M1 5 + M2 2 + M3 4 + M4 26 + M5 2 + M6 1 = 40; some overlap possible if M4 mirrors M1 files; tracked precisely in run-phase commit attributions per L46 path-specific staging discipline). Updated from iter-1 estimate of ~41 per D7 dev-only-commands-isolation.md template mirror removal.

## C. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-05-25 | Sprint 10 lane B entry chosen as 3-scope consolidation (W3-arch + W4 + W5) instead of 3 separate micro-SPECs | Per AskUserQuestion + prior session investigation: the three scopes share a single failure mode (local-only doctrine bleeding into template surface) and a single architectural principle (user-vs-maintainer artifact separation via PRESERVE-list contract). Consolidating produces 1 commit cohort + 1 CHANGELOG entry + 1 verification batch vs 3 drifted micro-SPECs. Tier M (4-6 milestones) appropriate for the ~41 file scope. |
| 2026-05-25 | `.claude/agents/local/` chosen as new namespace name (vs alternatives `.claude/agents/dev/` or `.claude/agents/maintainer/`) | "local" matches CLAUDE.local.md naming family + parallels `.claude/agent-memory-local/` from agent-authoring.md (existing user-not-shared memory scope). "dev" risks confusion with "development-mode" config keys. "maintainer" is too project-specific. |
| 2026-05-25 | Skill body deleted (M3) rather than retained as fallback | Per CLAUDE.local.md §2 Local-Only Files + §21 Dev-Only Commands Isolation, the 2 dev-only skill bodies are already local-only and never template-distributed. Retaining them as fallback duplicates the migrated content with no observable benefit — the thin command wrapper routes to the new agent, not the old skill. Deletion enforces single-source-of-truth. |
| 2026-05-25 | 99-release.md NOT included in this migration | Differential treatment is intentional. The production release workflow has higher risk surface (PR merge + git tag + version injection) and merits its own SPEC if migration is desired. Out-of-scope clearly enumerated in spec.md §E. |

## D. References

- spec.md (this SPEC) — REQ-LNC-001 through REQ-LNC-013, B.1/B.2/B.3 scope decomposition, D HARD/SHOULD constraints
- plan.md (this SPEC) — M1-M6 milestone breakdown, §F TRUST 5 mapping
- acceptance.md (this SPEC) — AC-LNC-001 through AC-LNC-011, per-AC verification commands
- `CLAUDE.local.md` §2 (Template-First Rule), §21 (Dev-Only Commands Isolation), §22 (Dev Settings Intent), §23 (Local Git Workflows + Hook Setup), §24 (Harness Namespace 분리 정책 + moai update contract) — local doctrine being externalized
- `.moai/docs/dev-only-commands-isolation.md` — Dev-only contract that M1 + M3 update
- `.claude/rules/moai/development/coding-standards.md` § Thin Command Pattern (lines 56-77) — HARD doctrine that M3 must preserve
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention — namespace SSOT updated in M1
- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy — cross-reference updated in M1
- `.claude/skills/moai/workflows/release-update.md` — predecessor skill body migrated in M2 (deleted in M3)
- `.claude/skills/moai/workflows/github.md` — predecessor skill body migrated in M2 (deleted in M3)
- `.claude/commands/97-release-update.md` — thin command wrapper updated in M3
- `.claude/commands/98-github.md` — thin command wrapper updated in M3
- `.moai/docs/generic-patterns-guide.md` (NEW) — externalized patterns authored in M5
- `agent-common-protocol.md` § Parallel Execution + § Pre-Spawn Sync Check — verification batch + race mitigation discipline
- Status Transition Ownership Matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md` — `draft → in-progress` owned by manager-develop (M1 first commit), `in-progress → implemented` owned by manager-docs (sync commit)

## E. Phase-Specific Audit-Ready Signals

### E.1 Plan-phase Audit-Ready Signal

Authored by manager-spec at plan-phase completion. Populated upon Phase 0.5 plan-auditor verdict.

| Signal | Value | Status |
|--------|-------|--------|
| 4 artifacts created (spec/plan/acceptance/progress) | YES — all 4 written in single manager-spec session | `<pending verification>` |
| Frontmatter 12-canonical-field validation | PASSED at write time per pre-write checklist | `<pending plan-auditor confirmation>` |
| SPEC ID regex compliance | PASSED — decomposition: SPEC ✓ \| V3R6 ✓ \| LOCAL ✓ \| NAMESPACE ✓ \| CONSOLIDATION ✓ \| 001 ✓ → PASS | `<pending plan-auditor confirmation>` |
| 13 REQ-LNC GEARS notation compliance (iter-3) | 4 Ubiquitous + 3 Event-driven + 2 State-driven + 2 Where capability + 2 Unwanted = 13 total, zero IF/THEN. iter-3 D_new3: REQ-LNC-014 DELETED (redundant subset of REQ-LNC-011 second clause); REQ count 14 → 13; Where-capability count 3 → 2. | `<pending plan-auditor confirmation>` |
| 12 AC-LNC independent verifiability | 11 MUST-PASS AC (AC-LNC-001 through AC-LNC-011) have grep/test/file-existence commands per acceptance.md §B; AC-LNC-012 NEW SOFT (deferred) for REQ-LNC-009 traceability anchor — does NOT block Definition of Done. AC-LNC-006 binding reverted to REQ-LNC-011 only in iter-3 (REQ-LNC-014 deleted). AC count 11 → 12. | `<pending plan-auditor confirmation>` |
| HARD constraints documented | 5 HARD constraints (Thin Command Pattern, Template-First, Namespace contract update [iter-2: §24.4 dropped, 2 in-scope SSOTs only], Dev-only isolation, GEARS discipline) per spec.md §D.1 | `<pending plan-auditor confirmation>` |
| plan-auditor verdict | Tier M PASS threshold ≥ 0.80 | `<pending>` |
| plan-auditor 4-dimension scores | Functionality / Security / Craft / Consistency | `<pending>` |

### E.2 Run-phase Audit-Ready Signal

Populated by manager-develop at run-phase completion.

`<pending — owned by manager-develop per REQ-ARR-002>`

### E.3 Run-phase Audit-Ready Signal (M6 verification batch outputs)

Populated by manager-develop after executing the 7-command verification batch per plan.md §C.M6.

`<pending — owned by manager-develop per REQ-ARR-002>`

### E.4 Sync-phase Audit-Ready Signal

Populated by manager-docs at sync-phase completion.

`<pending — owned by manager-docs per REQ-ARR-003>`

### E.5 Mx-phase Audit-Ready Signal

Populated by manager-docs or orchestrator at Mx Step C judgment time.

`<pending — Mx Step C EVALUATE-SKIP expected for markdown-only changes; if M3 updates commands_audit_test.go .go file, EVALUATE-EXECUTE applies>`

## F. HISTORY

| Version | Date | Author | Iteration | Description |
|---------|------|--------|-----------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | iter-1 | Initial progress.md authoring — §A Lifecycle Reflection + §B Milestone Status (M1-M6) + §C Decision Log (4 entries) + §D References + §E Phase-Specific Audit-Ready Signals (E.1-E.5). All milestones not-started. |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 | Focused defect resolution per plan-auditor iter-1 0.73 FAIL — D7 M1 file count 6 → 5 (dev-only-commands-isolation.md template mirror dropped per spec.md §E), total file count ~41 → ~40 + arithmetic breakdown added (M1 5 + M2 2 + M3 4 + M4 26 + M5 2 + M6 1 = 40), D6 §E.1 REQ count 13 → 14 (REQ-LNC-014 NEW Where-capability) + GEARS notation breakdown updated (3 Where-capability instead of 2) + AC-LNC-006 binding broadens noted, HARD constraint #3 §24.4 dropped noted, D8 HISTORY section NEW. tier:M frontmatter added per D13. |
| 0.1.2 | 2026-05-25 | manager-spec | iter-3 | Narrow-scope surgical defect resolution per plan-auditor iter-2 0.74 PASS-WITH-DEBT (stagnation, LEAN STOP signal): §E.1 audit-ready signal table updated for new counts — REQ count 14 → 13 (D_new3 REQ-LNC-014 deletion), GEARS breakdown 3 Where-capability → 2 Where-capability, AC count 11 → 12 (D_new4 AC-LNC-012 NEW deferred-verification marker binds orphan REQ-LNC-009), AC-LNC-006 binding reverted to REQ-LNC-011 only. File-count breakdown unchanged (40 total) — D_new4 AC-LNC-012 addition is internal to acceptance.md (already counted as 1 of the 40 files). Other iter-3 defects (D_new1 REQ-LNC-002 5-field truth, D_new2 REQ-LNC-007 stdout-emptiness, D_new5 plan.md M2 §C.2 rewrite, D_new6 §E 8-phase→9-phase) are body-scope only, no progress.md signal update needed. |
