---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
title: "MoAI Lifecycle Redesign — 3-Phase Restoration + Epic-Based Naming"
version: "0.2.0"
status: draft
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/spec + .claude/rules/moai"
lifecycle: spec-anchored
tags: "lifecycle, redesign, era, epic, sdd"
era: V3R6
tier: L
---

# SPEC-V3R6-LIFECYCLE-REDESIGN-001

## §A. Problem Statement

MoAI's lifecycle model drifted from its original 3-phase design (plan → run → sync, with MX Tag as a cross-cutting concern) into an accreted 4-phase model (plan → run → sync → Mx) during the V3R6 era. This drift is mechanical and pervasive: `internal/spec/era.go` H-4 detection requires `§E.2 + §E.5 + both commit_sha`, progress.md grew a 5th section (`§E.5 Mx-phase Audit-Ready Signal`), the Status Transition Ownership Matrix carries a separate Mx chore commit transition, and 14 live files in `.claude/` reference the "4-phase close" terminology.

Simultaneously, MoAI's work-unit vocabulary (`Sprint`, `cohort`, `Round`, `Wave`) is unaligned with the de-facto Spec-Driven Development (SDD) standard (GitHub Spec Kit). A web survey of the canonical SDD reference implementation (github/spec-kit) confirms that `Sprint`, `cohort`, `Wave`, `Epic`, and `Round` are **absent** from the dominant SDD vocabulary; Spec Kit uses `spec`, `plan`, `tasks`, `constitution`, `feature`, and `branch` exclusively.

This SPEC defines the redesign on two axes (Axis A: 3-phase restoration; Axis B: Epic-based naming) without modifying any production code or rule file — implementation is deferred to the run phase per the plan-run-sync discipline.

## §B. Context

### §B.1 Doctrine Drift (Axis A) — Verified Evidence

Original MoAI lifecycle (pre-V3R6 accretion):
- CLAUDE.md §5 "MX Tag Integration": plan (identify targets) / run (create tags) / sync (validate tags). MX Tag is **cross-cutting**, NOT a phase.
- `.claude/rules/moai/workflow/mx-tag-protocol.md`: MX Tag protocol is cross-cutting by design.

V3R6-era accretion (the drift):
- `internal/spec/era.go:135` (H-4): `if hasSyncSection && hasMxSection && syncSHA != "" && mxSHA != ""` — requires both `§E.5` marker and `mx_commit_sha` field. (Doc-comment lines ~86-101 also hardcode this — D5 scope.)
- `internal/spec/audit.go:224-300` `checkV3R6Drift`: emits THREE §E.5-keyed MUST-FIX findings — `Y_Y_Y_Y_StatusDrift` (251), `Y_Y_N_Y` (268), `Y_N_N_Y` (284) — on the §E.5/mx_commit_sha predicate (D2).
- `internal/spec/transitions.go:74` `closeInfix4Phase = "4-phase close"`: the drift walker's only positive `completed` signal is keyed on the "4-phase close" literal (D4).
- progress.md template grew `§E.5 Mx-phase Audit-Ready Signal` as a 5th section.
- Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md`): `implemented → completed` transition fires on a separate "Mx chore commit".

**Drift surface (measured 2026-06-18, `grep -rl '4-phase\|Mx-phase\|mx_commit_sha\|§E\.5\|Mx chore' .claude/` excluding worktrees/agent-memory/specs/reports)**: 14 files.

### §B.2 Industry Survey (Axis B) — Verbatim Citations

Primary source: GitHub Spec Kit (`https://github.com/github/spec-kit`), fetched 2026-06-18 via `mcp__web_reader__webReader`.

Spec Kit Core Commands (verbatim from the Core Commands table):
- `/speckit.constitution` — Create or update project governing principles and development guidelines
- `/speckit.specify` — Define what you want to build (requirements and user stories)
- `/speckit.plan` — Create technical implementation plans with your chosen tech stack
- `/speckit.tasks` — Generate actionable task lists for implementation
- `/speckit.implement` — Execute all tasks to build the feature according to the plan
- `/speckit.converge` — Assess the codebase against spec/plan/tasks and append remaining work

Work-unit vocabulary in Spec Kit: `spec`, `plan`, `tasks`, `constitution`, `feature`, `branch`. Branches are named by sequential numeric prefix (e.g. `001-create-taskify`), NOT by "sprint", "cohort", or "wave".

Vocabulary absent from Spec Kit README (verified by reading the full fetched content): `Sprint`, `Cohort`, `Wave`, `Round`, `Epic`.

Implication: MoAI's `SPEC` and `skill` naming already align with SDD. `Sprint`/`cohort`/`Round`/`Wave` are legacy Agile terms that conflict with the SDD-aligned vocabulary.

### §B.3 Era Migration Impact (MOVING baseline — D3; counts illustrative, re-measure at M1)

> **Moving-baseline note (D3)**: the V3R6 count is NOT a frozen literal. Plan-phase measurement (2026-06-18) = **48**; re-measurement (2026-06-19) = **53** (modern_era_clean 77, total 357), still rising because a parallel session is authoring more V3R6 SPECs. Every count below is an **illustrative snapshot**, NOT a pass condition. The authoritative baseline `N` is captured at run-phase M1 start; count-dependent ACs assert **invariance** (post == pre-migration N), never literal equality. See acceptance.md AC-LR-003 + §E quality gate.

| Metric | Value (e.g. 2026-06-18 → 2026-06-19) | Source |
|--------|-------|--------|
| Total SPECs in catalog | 351 → 357 | `moai spec audit --json total_specs` |
| Grandfather-protected (V2.x + V3R2-R4 + V3R5) | 270 → 271 | `grandfathered` |
| Modern-era clean (V3R6 H-4 satisfied) | 75 → 77 | `modern_era_clean` |
| Drift findings (all eras, all severities) | 319 | `len(drift_findings)` |
| V3R6-era SPECs (H-4 detected) | 48 → 53 | `drift_findings[era==V3R6]` |
| MUST-FIX drift findings | 6 | `severity==MUST-FIX` |
| progress.md files total | ~206 | `ls SPEC-*/progress.md` |
| progress.md with `§E.5` marker | ~83 | grep `§E.5` in progress.md |
| progress.md with `mx_commit_sha` field | ~84 | grep `mx_commit_sha:` in progress.md |
| SPECs with explicit `era: V3R6` frontmatter | ~39 | grep `^era:.*V3R6` |

**Migration-affected set (the N≈53 V3R6-era SPECs, re-measure at M1)**: These carry a V3R6 predicate. Of the current 53, **29** already carry the new layout (`§E.2 + §E.4 + sync_commit_sha`, caught directly by the NEW H-4) and **11** are legacy-layout (`§E.2 + §E.5 + sync_sha + mx_sha`, no `§E.4`). Correction (D1): when H-4 is rewritten to drop the `§E.5 + mx_commit_sha` requirement, the legacy-layout SPECs do **NOT** regress to V3R5 (the H-3 empty-sync_sha condition does not match a populated sync_sha); the genuine regression vector is H-6 unclassified, whose at-risk set is empty for the current catalog (research.md §D.4). The migration window + auto-fold backfill preserve V3R6 classification as defense-in-depth.

## §C. Goals / Non-Goals

### Goals
- G1: Restore the 3-phase lifecycle (plan → run → sync) as canonical, with MX Tag returned to cross-cutting status.
- G2: Redefine the progress.md §E structure from 5 sections to 4 sections, merging `§E.5 Mx-phase` into `§E.4 Sync-phase`.
- G3: Reclassify era.go H-4 detection to no longer require `§E.5 + mx_commit_sha`, without breaking the 48 existing V3R6 SPECs.
- G4: Merge the `completed` status transition into the sync commit; remove the separate Mx chore commit from the Status Transition Ownership Matrix.
- G5: Replace the Sprint/cohort/Round/Wave vocabulary with an Epic-based taxonomy aligned to SDD.
- G6: Sweep the 14 drift-surface files (Axis A) and the naming-surface files (Axis B) to reflect the new terminology.

### Non-Goals
- N1: This SPEC does NOT change the GEARS notation, TRUST 5 framework, or the 3-phase plan/run/sync **concept** itself — only the drift accretions are removed.
- N2: This SPEC does NOT rename the `SPEC-{DOMAIN}-{NUM}` identifier scheme (already SDD-aligned).
- N3: This SPEC does NOT rename the `skill` naming scheme (already SDD-aligned).
- N4: This SPEC does NOT retroactively rewrite the 270 grandfather-protected SPECs (V2.x/V3R2-R4/V3R5) — they remain era-protected per `lifecycle-sync-gate.md`.
- N5: This SPEC does NOT touch the agent catalog (8 retained agents) or the worktree system.

## §D. Requirements (GEARS)

### Axis A — 3-Phase Restoration

#### REQ-LR-001 (Ubiquitous)
The [MoAI lifecycle model] **shall** consist of exactly three canonical phases: `plan`, `run`, and `sync`.

#### REQ-LR-002 (Ubiquitous)
The [MX Tag system] **shall** be a cross-cutting concern active during `plan` (identify targets), `run` (create/update tags), and `sync` (validate tags), NOT a separate fourth phase.

#### REQ-LR-003 (Event-driven)
**When** a SPEC's `progress.md` is authored, the [manager-spec] **shall** populate exactly four §E sections: `§E.1 Plan-phase Audit-Ready Signal`, `§E.2 Run-phase Evidence`, `§E.3 Run-phase Audit-Ready Signal`, `§E.4 Sync-phase Audit-Ready Signal` (which absorbs the former `§E.5 Mx-phase` content).

#### REQ-LR-004 (Unwanted)
The [progress.md template] **shall not** contain a `§E.5 Mx-phase Audit-Ready Signal` section in new SPECs.

#### REQ-LR-005 (Event-driven)
**When** the `ClassifyEra` function in `internal/spec/era.go` evaluates a SPEC's signals, the [era classification engine] **shall** classify a SPEC as V3R6 when `§E.2` run-evidence marker is present AND `§E.4` sync marker is present AND `sync_commit_sha` field is non-empty — WITHOUT requiring `§E.5` or `mx_commit_sha`.

#### REQ-LR-006 (State-driven) — NARROWED (D1/D3): defense-in-depth, not regression-prevention
**While** the H-4 predicate is being rewritten, the [era classification engine] **shall** preserve V3R6 classification for every legacy-layout V3R6 SPEC (those carrying `§E.5 + mx_commit_sha` but lacking `§E.4`) by treating the presence of EITHER the new predicate (§E.2 + §E.4 + sync_sha) OR the legacy predicate (§E.2 + §E.5 + sync_sha + mx_sha) as a V3R6 signal during a migration window.

> **Scope correction (D1)**: This requirement was originally justified as preventing the existing V3R6 SPECs from regressing to **V3R5 (H-3)**. That justification was FALSE — H-3 (`era.go:130`) fires only when `sync_commit_sha` is empty, so a SPEC with a populated `sync_commit_sha` never matches H-3. The genuine regression vector is **H-6 unclassified**, and the re-derived H-6 at-risk set is **empty** for the current catalog (every current V3R6 SPEC is caught by the H-5 `created >= 2026-04-01` / modern-`phase:` heuristic even with no migration window — see research.md §D.4). REQ-LR-006 is therefore **NARROWED** to two surviving purposes: (a) **classification-rationale precision** — the legacy-layout SPECs classify as "V3R6 via legacy predicate" (explicit) rather than "V3R6 via H-5 date heuristic" (weaker), and (b) **future-proofing** — a not-yet-authored V3R6 SPEC with a pre-`modernEraThreshold` `created:` and no modern `phase:` COULD reach H-6, which the legacy fallback would still catch. It is KEPT (not removed) because the catalog is moving and an explicit predicate is a stronger signal than the date heuristic; it is NOT a misclassification-prevention necessity for the current population.

#### REQ-LR-007 (Capability gate)
**Where** a SPEC's `progress.md` carries the legacy 5-section layout (`§E.1..§E.5`) with populated commit SHAs, the [audit engine] **shall** auto-fold the `§E.5` content into `§E.4` during a one-time backfill migration and record the migration in a migration log.

#### REQ-LR-008 (Event-driven)
**When** the sync commit transitions a SPEC from `implemented` to `completed`, the [Status Transition Ownership Matrix] **shall** accept the sync commit (authored by `manager-docs`) as the canonical close commit, WITHOUT requiring a separate Mx chore commit.

#### REQ-LR-009 (Unwanted)
The [Status Transition Ownership Matrix] **shall not** list a separate `* → completed` transition owned by a standalone "Mx chore commit" — the completed transition is merged into the sync commit.

#### REQ-LR-010 (Ubiquitous)
The [6 drift-surface rule files] (verification-claim-integrity.md, spec-frontmatter-schema.md, agent-patterns.md, lifecycle-sync-gate.md, archived-agent-rejection.md, spec-workflow.md) **shall** use the phrase "3-phase close (plan→run→sync)" and **shall not** use "4-phase close".

#### REQ-LR-011 (State-driven)
**While** MX Tag validation occurs during sync, the [manager-docs] **shall** perform MX Tag validation as part of the sync-phase quality gate, NOT as a separate Mx-phase step.

### Axis B — Epic-Based Naming

#### REQ-LR-012 (Ubiquitous)
The [MoAI work-unit vocabulary] **shall** consist of at most four canonical terms: `Epic` (multi-SPEC grouping), `SPEC` (single work unit), `Milestone` (within-SPEC ordered step), and `Constitution` (project-level governance, aligned with Spec Kit).

#### REQ-LR-013 (Event-driven)
**When** the sprint-round-naming.md SSOT is rewritten, the [manager-spec] **shall** replace `Sprint` with `Epic`, remove `cohort` (folded into Epic-internal grouping), remove `Round` (folded into Milestone), remove `Wave` (legacy retired term), and retain `Milestone` as the industry-standard within-SPEC step.

#### REQ-LR-014 (Capability gate)
**Where** a multi-SPEC grouping is referenced, the [orchestrator output and documentation] **shall** use `Epic` (English, technical identifier) or `에픽` (Korean, user-facing) — **shall not** use `Sprint`, `스프린트`, `cohort`, `코호트`, `Round`, `라운드`, or `Wave`.

#### REQ-LR-015 (Event-driven)
**When** the naming-surface files (102+ files matched by `grep -rl 'Sprint|코호트|cohort|Round|Wave'`) are migrated, the [migration] **shall** proceed in priority tiers: (T1) the 11 canonical rule files in `.claude/rules/moai/`, (T2) agent definitions and output-styles, (T3) skill bodies, (T4) project docs (`.moai/docs/`, `.moai/project/`), (T5) archived/historical content (migrated only if actively referenced).

#### REQ-LR-016 (State-driven)
**While** the migration is in progress, the [anti-pattern catalogue] **shall** re-anchor AP-SRN-001..004 to the new Epic-based taxonomy (e.g. AP-SRN-001 becomes "Calling a multi-SPEC group 'Round' or 'Wave' instead of 'Epic'").

#### REQ-LR-017 (Ubiquitous)
The [Epic term] **shall** denote a time-unit or thematic container for one or more SPECs, equivalent to the pre-redesign `Sprint` semantics — the meaning is preserved, only the label changes.

### Cross-Axis

#### REQ-LR-018 (Event-driven)
**When** a SPEC is authored under the redesigned lifecycle, the [manager-spec] **shall** emit a 4-section `progress.md` skeleton (§E.1..§E.4) and reference the Epic (not Sprint) the SPEC belongs to in its `phase:` frontmatter or plan.md context.

#### REQ-LR-019 (Unwanted) — EXTENDED (D2): all THREE §E.5-keyed findings
The [audit engine `checkV3R6Drift`] **shall not** emit any of the **three** `§E.5`/`mx_commit_sha`-keyed findings — `Y_Y_Y_Y_StatusDrift` (audit.go:251), `Y_Y_N_Y` (audit.go:268), and `Y_N_N_Y` (audit.go:284) — predicated on the presence or absence of `§E.5` or `mx_commit_sha` for SPECs authored under the redesigned 3-phase lifecycle. In particular, the `Y_N_N_Y` finding ("§E.2 sync section present but §E.5 Mx section absent") **shall** be retired or re-anchored: under the mandated 4-section end-state (no `§E.5`), `Y_N_N_Y` would otherwise fire MUST-FIX on EVERY non-`completed` V3R6 SPEC, converting the clean end-state into a catalog-wide drift storm. The single surviving drift dimension **shall** be status drift on the 3-marker predicate (`§E.2 + §E.4 + sync_commit_sha` present but `status != completed`), re-anchored from `Y_Y_Y_Y_StatusDrift` to `SyncStatusDrift`.

#### REQ-LR-020 (Event-driven) — NEW (D4): close-subject convention reconciliation
**When** the "4-phase close" terminology is renamed to "3-phase close" in the close-subject mandate prose, the [redesign] **shall** reconcile with SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 (the cross-SPEC owner of the close-subject convention, per spec-frontmatter-schema.md:78 and lifecycle-sync-gate.md:235) AND **shall** update the drift detector's close-infix matcher in `internal/spec/transitions.go` (`closeInfix4Phase` const + `closeInfixMatch`, lines 73-83) to accept the new `"3-phase close"` infix while RETAINING the legacy `"4-phase close"` infix — because `closeInfixMatch` is the drift walker's only positive `completed` signal (transitions.go:69) and historical close commits in git history carry "4-phase close". A doc-only rename without the matcher update **shall not** be performed (it would silently break drift close-recognition for all future closes).

#### REQ-LR-021 (Unwanted) — NEW (D4): no silent override of the close-subject convention owner
The [redesign] **shall not** silently override the close-subject convention owned by SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 — the M4 doc-axis edits to spec-frontmatter-schema.md and lifecycle-sync-gate.md **shall** carry an explicit reconciliation note crediting DRIFT-LEGACY-CONVENTION-001 as the convention owner and recording that this SPEC amends the close infix from "4-phase close" to "3-phase close" (with backward-compat retention of the legacy infix).

## §E. Constraints

- C1: Plan-phase ONLY — this SPEC MUST NOT modify `era.go`, `audit.go`, any rule file, or any SSOT. Implementation is deferred to run-phase milestones.
- C2: GEARS notation required for all REQs (current notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format).
- C3: All 12 canonical frontmatter fields MUST be present (verified via self-check above).
- C4: Era migration MUST NOT lose V3R6 classification for the currently-V3R6 SPECs. The pass condition is **invariance** (D3): the V3R6 count after M3 == the baseline `N` captured at M1 start (`moai spec audit --json`), NOT equality to a frozen literal (the count is a moving baseline; e.g. N≈53 as of plan-phase).
- C5: Industry citations MUST be verifiable — Spec Kit citations carry the URL and fetch date.
- C6: No retroactive rewriting of 270 grandfather-protected SPECs (N4).
- C7: The Epic replacement for Sprint preserves semantics (multi-SPEC grouping); only the label changes.
- C8 (D4): The "4-phase close" → "3-phase close" rename MUST update `internal/spec/transitions.go` `closeInfix4Phase`/`closeInfixMatch` to accept BOTH infixes (legacy retained) in the same milestone as the doc-axis prose rename — a doc-only rename is forbidden (REQ-LR-020/021). The close-subject convention owner SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 MUST be credited, not silently overridden.

## §F. Assumptions

- A1: The 48 V3R6-era SPECs can be migrated via the auto-fold strategy (REQ-LR-007) without touching their `spec.md`/`plan.md`/`acceptance.md` bodies.
- A2: The `sprint-round-naming.md` rewrite (127 → ~120 lines) can preserve the anti-pattern structure while swapping the vocabulary.
- A3: The 14 drift-surface files (Axis A) can be edited in a single run-phase milestone without breaking downstream consumers.
- A4: The 102 naming-surface files (Axis B) can be migrated in priority tiers (T1-T5) without blocking the Axis A changes.
- A5: Spec Kit's vocabulary absence of `Sprint`/`cohort`/`Round`/`Epic` means Epic is MoAI's chosen replacement, NOT a direct Spec Kit borrowing — this is documented explicitly to prevent misattribution.

## §G. Dependencies

- `internal/spec/era.go` — doc-comment (lines ~86-101) + `ClassifyEra` body (lines ~117-146) rewritten (M1, D5 scope).
- `internal/spec/audit.go` `checkV3R6Drift` — all three §E.5-keyed findings (lines 251-297) + `FindingType` constants (lines 54-65) retired/re-anchored (M2, D2 scope).
- `internal/spec/transitions.go` — `closeInfix4Phase` const + `closeInfixMatch` (lines 73-83) extended to accept "3-phase close" while retaining "4-phase close" (M2, D4 scope).
- `.claude/rules/moai/development/sprint-round-naming.md` — SSOT rewrite (M6).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix + close-subject mandate (M4, D4 reconciliation).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — H-4 table + close-subject mandate + `## §E.5 Mx-phase` Worked Example (M4, D4/D5).
- **Cross-SPEC owner**: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 owns the close-subject convention (the "4-phase close" infix). This SPEC reconciles with it (REQ-LR-020/021); it does NOT silently override it. No blocking SPEC dependency, but the reconciliation note is mandatory.

## §H. Risks

- R1 (LOW, downgraded from HIGH per D1): Rewriting H-4 could in principle cause a legacy-layout V3R6 SPEC to fall to **H-6 unclassified** (NOT V3R5 — the original HIGH risk mis-stated the regression vector; H-3 only fires on empty `sync_commit_sha`). The genuine H-6 at-risk set is **empty** for the current catalog (all caught by H-5's `created >= 2026-04-01`/modern-`phase:` heuristic — research.md §D.4), so the residual risk is LOW and confined to a hypothetical future V3R6 SPEC with a pre-`modernEraThreshold` `created:` and no modern `phase:`. Mitigation: REQ-LR-006 dual-predicate window (defense-in-depth) + REQ-LR-007 auto-fold backfill + dedicated test fixtures + re-measure the H-6 at-risk set at M1 start (catalog is moving).
- R2 (MEDIUM): The naming migration (102 files) is large; incomplete migration leaves the vocabulary inconsistent. Mitigation: REQ-LR-015 priority tiers; T1 (rule files) ships first, T5 (archived) is best-effort.
- R3 (MEDIUM): The `completed` transition merge (REQ-LR-008/009) changes the commit-graph expectations of `status-transition-ownership.sh` hook. Mitigation: hook update is part of the same run-phase milestone.
- R4 (LOW): Memory files (`~/.claude/projects/*/memory/project_sprint*.md`) carry the legacy `sprint` token in filenames; renaming these breaks memory-index pointers. Mitigation: memory files are NOT in scope (N5 adjacent); `[SUPERSEDED]` markers added in new memory entries.
- R5 (LOW): External docs site (`adk.mo.ai.kr`) may carry Sprint/cohort terminology in published content. Mitigation: docs-site sweep is a follow-up SPEC, not this one.

## §I. Acceptance Criteria Summary

See `acceptance.md` for the full Given-When-Then matrix. Summary (all 13 ACs):
- AC-LR-001: 3-phase lifecycle is canonical (REQ-LR-001/002). MUST-PASS.
- AC-LR-002: progress.md is 4-section (REQ-LR-003/004). MUST-PASS.
- AC-LR-003: H-4 rewrite preserves the V3R6 count via invariance (REQ-LR-005/006/007). MUST-PASS.
- AC-LR-004: completed transition merges into sync commit (REQ-LR-008/009). MUST-PASS.
- AC-LR-005: 6 drift rules use "3-phase close" (REQ-LR-010). MUST-PASS.
- AC-LR-006: MX Tag validation occurs during sync, not as a separate phase (REQ-LR-011). SHOULD-PASS.
- AC-LR-007: Epic taxonomy replaces Sprint/cohort/Round/Wave (REQ-LR-012/013/014). MUST-PASS.
- AC-LR-008: naming migration proceeds in tiers (REQ-LR-015/016). SHOULD-PASS.
- AC-LR-009: Epic preserves Sprint semantics (REQ-LR-017). MUST-PASS.
- AC-LR-010: new SPECs reference Epic, not Sprint (REQ-LR-018). SHOULD-PASS.
- AC-LR-011: era classification emits none of the three §E.5-keyed drift findings, incl. `Y_N_N_Y` (REQ-LR-019). MUST-PASS.
- AC-LR-012: close-infix matcher accepts both "3-phase close" (new) and "4-phase close" (legacy); drift detector unaffected (REQ-LR-020/021). MUST-PASS.
- AC-LR-013: era.go doc-comment + lifecycle-sync-gate.md §E.5 worked example are updated to the 4-section layout (REQ-LR-005 / D5 scope). MUST-PASS.

## §J. Exclusions (What NOT to Build)

### Out of Scope — Deferred to follow-up SPECs

- EX-1: This SPEC does NOT implement any code change — it is the plan-phase artifact that DEFINES the redesign.
- EX-2: This SPEC does NOT modify the GEARS notation, TRUST 5 framework, or the agent catalog.
- EX-3: This SPEC does NOT retroactively rewrite the 270 grandfather-protected SPECs.
- EX-4: This SPEC does NOT rename memory files or docs-site published content (deferred to follow-up SPECs).
- EX-5: This SPEC does NOT change the `SPEC-{DOMAIN}-{NUM}` or `moai-*` naming schemes (already SDD-aligned).
- EX-6: This SPEC does NOT introduce a new "Constitution" slash command in MoAI — the `Constitution` term is referenced for SDD alignment but the command surface is unchanged.

## HISTORY

- 2026-06-18: plan-phase artifacts authored (spec.md, plan.md, acceptance.md, research.md, design.md) by manager-spec. Era migration impact measured via `moai spec audit --json` (48 V3R6 SPECs affected). Industry citations captured from GitHub Spec Kit (github/spec-kit, fetched 2026-06-18).
- 2026-06-19: plan-phase revision v0.2.0 after plan-audit iter-1 FAIL (0.71/0.85). Fixed 7 defects: D1 (corrected the FALSE H-3→V3R5 regression trace — actual fall-through is H-5; re-derived the true H-6 at-risk set = empty; NARROWED REQ-LR-006 to defense-in-depth); D2 (EXTENDED REQ-LR-019 to all three §E.5-keyed findings incl. `Y_N_N_Y`, which the 4-section end-state actively triggers catalog-wide); D3 (replaced literal-48 with moving-baseline invariance — re-measured V3R6=53 on 2026-06-19; ACs assert post==pre N captured at M1); D4 (NEW REQ-LR-020/021 + C8 reconciling the "4-phase close"→"3-phase close" rename with SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 + `internal/spec/transitions.go` `closeInfix4Phase`/`closeInfixMatch` dual-infix update); D5 (enumerated era.go doc-comment ~86-101 + lifecycle-sync-gate.md §E.5 worked example ~303 in scope); D6 (T1 file count corrected 11→10, AC-LR-008 converted to "0 residual matches"); D7 (extended §I to all 13 ACs). All ground-truth verified by direct source inspection of era.go/audit.go/transitions.go/drift.go.
