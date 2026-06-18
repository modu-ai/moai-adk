---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
plan_version: "0.2.0"
spec_version: "0.2.0"
status: draft
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Plan — SPEC-V3R6-LIFECYCLE-REDESIGN-001

## §A. Context

This plan operationalizes the two-axis redesign defined in `spec.md`:
- **Axis A (3-phase restoration)**: rewrite `era.go` H-4, merge `§E.5` into `§E.4`, fold the `completed` transition into the sync commit, sweep 6 drift-surface rules to "3-phase close".
- **Axis B (Epic-based naming)**: rewrite `sprint-round-naming.md` SSOT, migrate 102 naming-surface files in priority tiers, re-anchor anti-patterns.

The redesign touches Go code (`internal/spec/era.go`, `audit.go`), 14 drift-surface files, 102 naming-surface files, 4 agent definitions, and the orchestrator output style. It is a Tier L lifecycle change requiring careful migration sequencing.

## §B. Known Issues (Inherited from Current State)

- KI-1: `era.go:135` H-4 predicate requires `§E.5 + mx_commit_sha`, which is the drift to be removed.
- KI-2 (D2): `audit.go:224-300` `checkV3R6Drift` emits THREE §E.5-keyed findings — `Y_Y_Y_Y_StatusDrift` (251), `Y_Y_N_Y` (268), and `Y_N_N_Y` (284) — predicated on `§E.5`/`mx_commit_sha`. ALL three must be re-anchored or retired; `Y_N_N_Y` ("§E.2 present, §E.5 absent") is the critical one — the 4-section end-state trips it catalog-wide.
- KI-3 (D1 corrected): the V3R6 SPECs carry the legacy 5-section layout; a naive H-4 rewrite does NOT reclassify them to V3R5 (the H-3 empty-sync_sha condition does not match a populated sync_sha) — the genuine regression vector is H-6 unclassified, whose at-risk set is empty for the current catalog (H-5 catches all). The migration window is defense-in-depth.
- KI-7 (D4): `internal/spec/transitions.go:74` `closeInfix4Phase = "4-phase close"` is the drift walker's only positive `completed` signal; renaming "4-phase close" → "3-phase close" in docs WITHOUT updating this matcher silently breaks drift close-recognition for all future closes. The convention is owned by SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001.
- KI-4: `sprint-round-naming.md` is 127 lines with 4 anti-patterns (AP-SRN-001..004) anchored to Sprint/Round terminology.
- KI-5: 102 files carry Sprint/cohort/Round/Wave terminology (excluding worktrees, agent-memory, specs, reports, backups).
- KI-6: Memory filenames (`project_sprint*_*.md`) and MEMORY.md index entries carry the legacy token — out of scope per EX-4 but noted for awareness.

## §C. Pre-flight (Before Run Phase)

- PF-1 (D3): Capture the CURRENT V3R6 count `N` at M1 start via `moai spec audit --json` (a MOVING baseline — illustratively N≈53 as of 2026-06-19, but a parallel session is still authoring V3R6 SPECs, so DO NOT hardcode). All count-dependent ACs assert invariance (post-migration count == this captured `N`), never equality to a frozen literal.
- PF-1b (D1): Re-derive the genuine H-6 at-risk set at M1 start (V3R6 SPECs lacking §E.4, lacking legacy §E.5+mx_sha, AND lacking modern `phase:`/`created≥2026-04-01`). Plan-phase snapshot = empty; if M1 re-measurement finds a non-empty set, enumerate it and confirm the migration window (REQ-LR-006) covers each member. Use the reproduction command in research.md §D.4.
- PF-2: Confirm `era_test.go`, `audit_test.go`, and `transitions_test.go` pass on the pre-migration tree (regression baseline).
- PF-3: Confirm the 6 drift-surface rule files are committed at their current state (rollback baseline).
- PF-4: Confirm the SPEC directory contains all 5 plan-phase artifacts (spec.md, plan.md, acceptance.md, research.md, design.md).

## §D. Constraints (Reiteration)

- C1: Run-phase milestones MUST be sequential where they touch `era.go` (M1) and its tests (M2); parallel within a milestone is allowed for independent rule-file edits.
- C2: The migration window (REQ-LR-006) MUST be in place BEFORE the H-4 rewrite ships — otherwise the 48 SPECs misclassify mid-migration.
- C3: The naming migration (Axis B) MAY proceed in parallel with Axis A after M2 (era.go) lands, because the two axes touch disjoint file sets (Axis A = Go + 6 rules; Axis B = sprint-round-naming.md + 102 surface files).

## §E. Self-Verification (Plan-Phase)

This plan-phase artifact set carries:
- spec.md — 12 canonical frontmatter fields, 21 GEARS REQs (REQ-LR-001..021, incl. D4 REQ-LR-020/021), exclusions section.
- plan.md — this file; 9 milestones (M1-M9), risk register, cross-references.
- acceptance.md — 13 ACs (AC-LR-001..013) with Given-When-Then scenarios.
- research.md — drift surface counts (14 Axis A, 102 Axis B — verified EXACT), era migration impact (V3R6 moving baseline, re-derived H-6 at-risk set = empty), Spec Kit verbatim citations, corrected era-reclassification trace.
- design.md — H-4 reclassification strategy (corrected H-5 fall-through mechanism + auto-fold migration), all-three-findings drift update, close-infix reconciliation (D4), Epic taxonomy mapping table.

## §F. Milestones (Priority-Ordered, No Time Estimates)

### M1 — Axis A Foundation: era.go H-4 Reclassification + Migration Window
Scope: Rewrite `internal/spec/era.go` `ClassifyEra` H-4 predicate to require `§E.2 + §E.4 + sync_commit_sha` (NOT `§E.5 + mx_commit_sha`). Add the dual-predicate migration window (REQ-LR-006): V3R6 is detected when EITHER the new predicate OR the legacy predicate holds. **D5 scope**: ALSO rewrite the `ClassifyEra` **doc-comment heuristic table (lines ~86-101)** which hardcodes the old `H-4: §E.2 + §E.5 + mx_commit_sha` text — the doc-comment must describe the new §E.4 predicate + legacy fallback. **D1 step**: at M1 start, re-derive the genuine H-6 at-risk set (PF-1b) — the corrected fall-through is H-5, NOT H-3 (H-3 only fires on empty sync_sha); confirm the at-risk set is empty or covered by the window.
Files: `internal/spec/era.go` (doc-comment ~86-101 + body ~117-146 + the `hasMxSection`/`mxSHA` parse lines ~118/120).
Dependencies: PF-1, PF-1b, PF-2.
Acceptance: `moai spec audit --json` V3R6 count after M1 == the M1-captured baseline `N` (invariance, D3); the genuine H-6 at-risk set is empty or each member is covered by the window.

### M2 — Axis A Tests + Drift Re-anchor + Close-Infix (D2 + D4)
Scope: (a) Update `internal/spec/era_test.go` to cover the new H-4 predicate, the dual-predicate window, and the H-5 fall-through. (b) **D2 — retire/re-anchor ALL THREE §E.5-keyed findings** in `internal/spec/audit.go` `checkV3R6Drift` (lines 251-297): `Y_Y_Y_Y_StatusDrift` re-anchored to the 3-marker predicate (`SyncStatusDrift`), `Y_Y_N_Y` retired, **`Y_N_N_Y` retired** (this is the one the 4-section end-state actively triggers catalog-wide); update `FindingType` constants (lines 54-65); add `audit_test.go` fixtures for all three (incl. the `Y_N_N_Y`-must-not-fire and `SyncStatusDrift`-must-fire cases). (c) **D4 — extend the close-infix matcher** in `internal/spec/transitions.go` (lines 73-83): add `closeInfix3Phase = "3-phase close"` and OR it into `closeInfixMatch`, RETAINING `closeInfix4Phase = "4-phase close"` for legacy git-history commits; update `transitions_test.go` + `drift_combined_scope_test.go` to assert both infixes are recognized.
Files: `internal/spec/era_test.go`, `internal/spec/audit.go` (lines 54-65, 251-297), `internal/spec/audit_test.go`, `internal/spec/transitions.go` (lines 73-83), `internal/spec/transitions_test.go`, `internal/spec/drift_combined_scope_test.go`.
Dependencies: M1.
Acceptance: `go test ./internal/spec/...` passes; V3R6 count invariant; `moai spec audit --json | grep -c '"finding_type": "Y_N_N_Y"'` == 0; `closeInfixMatch` recognizes both "3-phase close" and "4-phase close" (AC-LR-012).

### M3 — Axis A Backfill: progress.md Auto-Fold Migration
Scope: Implement REQ-LR-007 — a one-time migration that folds `§E.5 Mx-phase` content into `§E.4 Sync-phase` for the 48 affected SPECs. Record migration in a migration log (`.moai/state/lifecycle-redesign-migration.json`).
Files: a new migration script (`internal/spec/migrate_3phase.go` or equivalent), the legacy-layout progress.md files (content fold — count re-measured at M3, ~11 legacy-layout of the N V3R6 SPECs as of plan-phase).
Dependencies: M1, M2.
Acceptance: After backfill, the V3R6 count == the M1-captured baseline `N` (invariance, D3) via the new predicate; the migration log records all folded entries.

### M4 — Axis A Rules Sweep: 6 Drift-Surface Rule Files → "3-phase close" (+ D4 reconciliation + D5 worked example)
Scope: Edit the 6 drift-surface rule files to replace "4-phase close" with "3-phase close (plan→run→sync)" and update the Status Transition Ownership Matrix to merge `completed` into the sync commit. **D4 reconciliation**: the close-subject mandates in `spec-frontmatter-schema.md` (lines ~71/~78) and `lifecycle-sync-gate.md` (lines ~228/~235) are OWNED by SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 — each edit MUST carry a reconciliation note crediting that SPEC as the convention owner and recording that LIFECYCLE-REDESIGN-001 amends the infix from "4-phase close" to "3-phase close" (legacy retained in the matcher per M2/REQ-LR-020/021). **D5**: `lifecycle-sync-gate.md` ALSO update the `## §E.5 Mx-phase Audit-Ready Signal` worked example (line ~303), the H-4 heuristic-table row (line ~43), and the era-definition row (line ~28) to the 4-section layout.
Files: `.claude/rules/moai/core/verification-claim-integrity.md`, `.claude/rules/moai/development/spec-frontmatter-schema.md`, `.claude/rules/moai/development/agent-patterns.md`, `.claude/rules/moai/workflow/lifecycle-sync-gate.md`, `.claude/rules/moai/workflow/archived-agent-rejection.md`, `.claude/rules/moai/workflow/spec-workflow.md`.
Dependencies: M1, M2 (rule text + close-infix matcher must both be canonical first; doc-only rename without the M2 matcher update is forbidden per C8/REQ-LR-020).
Acceptance: `grep -r '4-phase close\|Mx-phase Audit-Ready' .claude/rules/moai/` returns 0 matches in the 6 files EXCEPT where a reconciliation note legitimately cites the legacy infix; both close-subject mandate files contain a DRIFT-LEGACY-CONVENTION-001 reconciliation note (AC-LR-012); the lifecycle-sync-gate.md §E.5 worked example reflects the 4-section layout (AC-LR-013).

### M5 — Axis A Supporting Surface: Agents, Hooks, Output-Styles, Skills
Scope: Edit the remaining 8 Axis A surface files (4 agent definitions, 1 hook, 1 output-style, 2 skills) to reflect the 3-phase close and the §E.4-only progress.md structure.
Files: `.claude/agents/moai/manager-{spec,develop,docs,git}.md`, `.claude/agents/harness/workflow-specialist.md`, `.claude/hooks/moai/status-transition-ownership.sh`, `.claude/output-styles/moai/moai.md`, `.claude/skills/harness-moaiadk-patterns/SKILL.md`, `.claude/skills/moai/workflows/plan/spec-assembly.md`.
Dependencies: M4 (rule text must be canonical first).
Acceptance: `grep -rl '4-phase\|Mx-phase\|§E\.5' .claude/` (excluding worktrees/agent-memory/specs/reports) returns 0 matches.

### M6 — Axis B SSOT Rewrite: sprint-round-naming.md → Epic Taxonomy
Scope: Rewrite `.claude/rules/moai/development/sprint-round-naming.md` to define the Epic-based taxonomy (Epic / SPEC / Milestone / Constitution). Re-anchor AP-SRN-001..004 to the new vocabulary. Remove Sprint/cohort/Round/Wave as canonical terms (preserve as "legacy aliases" in a migration note).
Files: `.claude/rules/moai/development/sprint-round-naming.md`.
Dependencies: none (Axis B is independent of Axis A after M2).
Acceptance: The SSOT defines exactly 4 canonical terms; AP-SRN-001..004 are re-anchored; Sprint/cohort/Round/Wave appear only in a "Legacy Aliases" migration section.

### M7 — Axis B T1 Migration: T1 Canonical Rule Files (D6: 10 live, 9 after M6 excludes sprint-round-naming.md)
Scope: Migrate Sprint/cohort/Round/Wave references in the T1 `.claude/rules/moai/` files (live `grep -rl` returns **10** as of 2026-06-19; **9** after excluding sprint-round-naming.md itself, handled in M6) to the Epic taxonomy. The binary pass condition is "0 residual matches", NOT a fixed file count (catalog is moving — D6).
Files: `.claude/rules/moai/core/{askuser-protocol,zone-registry}.md`, `.claude/rules/moai/development/manager-develop-prompt-template.md`, `.claude/rules/moai/workflow/{archived-agent-rejection,ci-autofix-protocol,ci-watch-protocol,orchestration-mode-selection,session-handoff,worktree-state-guard}.md`, and others matched by the Axis B grep.
Dependencies: M6 (SSOT must be canonical first).
Acceptance: `grep -r 'Sprint [0-9]\|코호트\|cohort\|Round [0-9]\|Wave [0-9]\|스프린트\|라운드' .claude/rules/moai/` returns 0 matches (excluding the "Legacy Aliases" section in sprint-round-naming.md).

### M8 — Axis B T2-T4 Migration: Agents, Output-Styles, Skills, Docs
Scope: Migrate the remaining Axis B surface files (agents, output-styles, skills, `.moai/docs/`, `.moai/project/`) to the Epic taxonomy.
Files: `.claude/output-styles/moai/moai.md`, `.claude/skills/{moai-domain-database,moai-workflow-ci-loop}/SKILL.md`, `.claude/skills/moai/workflows/{plan/clarity-interview,project/*}.md`, `.moai/docs/*.md`, `.moai/project/codemaps/*.md`.
Dependencies: M7.
Acceptance: Axis B grep returns 0 matches in T1-T4 scope; T5 (archived/historical) is best-effort.

### M9 — Sync Phase
Scope: CHANGELOG entry, README update, docs-site sweep coordination (cross-ref deferred to follow-up SPEC per EX-4), frontmatter status transition `draft → planned` (post-plan-auditor) then `in-progress → implemented → completed` across the run/sync phases.
Dependencies: M1-M8.
Acceptance: 4-phase close (now 3-phase close) — sync commit carries the `completed` transition; progress.md §E.4 records the sync audit; no separate Mx commit.

## §G. Anti-Patterns (Plan-Phase)

- AP-LR-P-001: Attempting to rewrite `era.go` H-4 WITHOUT the dual-predicate migration window (REQ-LR-006) — causes 48 SPECs to misclassify mid-migration.
- AP-LR-P-002: Editing the 6 drift-surface rules (M4) BEFORE `era.go` (M1) — rule text drifts from the Go logic.
- AP-LR-P-003: Treating the naming migration (Axis B) as blocking the lifecycle restoration (Axis A) — the two axes are independent after M2 and should proceed in parallel.
- AP-LR-P-004: Retroactively rewriting the 270 grandfather-protected SPECs to the new 4-section layout — violates N4 and the grandfather clause.
- AP-LR-P-005: Renaming memory files or docs-site content during this SPEC — out of scope per EX-4; deferred to follow-up SPECs.

## §H. Cross-References

- `spec.md` — REQ-LR-001..021, AC-LR-001..013.
- `acceptance.md` — Given-When-Then scenarios per AC group.
- `research.md` — drift surface measurement, era migration impact, Spec Kit citations, corrected era trace (§D.3) + re-derived at-risk set (§D.4).
- `design.md` — H-4 reclassification strategy (corrected mechanism), all-three-findings drift update (§B.4), close-infix reconciliation (§B.6), Epic taxonomy mapping.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix + close-subject mandate (target of M4; D4 reconciliation with DRIFT-LEGACY-CONVENTION-001).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — H-1..H-6 heuristic table + §E.5 worked example + close-subject mandate (target of M4; D5).
- `.claude/rules/moai/development/sprint-round-naming.md` — SSOT (target of M6 rewrite).
- `internal/spec/era.go` — H-4 detection logic + doc-comment (target of M1; D5).
- `internal/spec/audit.go` — drift detection (3 findings, target of M2; D2).
- `internal/spec/transitions.go` — `closeInfix4Phase`/`closeInfixMatch` (target of M2; D4).
- **SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001** — cross-SPEC owner of the close-subject convention (reconciled by M4, not overridden).
