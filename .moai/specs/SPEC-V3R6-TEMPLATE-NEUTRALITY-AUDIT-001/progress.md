---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Progress Tracking"
version: "0.1.1"
status: in-progress
created: 2026-05-23
updated: 2026-05-30
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, progress"
tier: L
related_specs: [SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Progress Tracking

## Status

**Phase**: run-phase (in-progress v0.1.1) — M1 complete; M2-M5 run-phase entered after plan-auditor iter-2 PASS 0.88 + GATE-2 approval
**Created**: 2026-05-23
**Last update**: 2026-05-30 (run-phase M2 begin: status draft → in-progress + Mode Selection logged)

## §E — Phase 0.95 Mode Selection

### Input parameters

- **tier**: L (Large; > 15 files affected across M2-M5)
- **scope (file count)**: ~12-80 template markdown files (C2 dominates at 73 files; total kept-class union before dedup) + 2 new files (Go test + CI workflow)
- **domain count**: 1 primary (markdown/template content under `internal/template/templates/`) + 1 secondary (Go test infra + CI YAML at M5)
- **file language mix**: ~95% markdown/template content; ~5% Go + YAML (M5 only)
- **concurrency benefit**: LOW — per Finding A4 (coding/content-heavy caveat), template-content sweep is sequential grep+rewrite per-milestone with a shared working tree; parallel spawn would race on the same tree
- **Agent Teams prereqs status**: NOT evaluated (harness level not `thorough` + no multi-domain research benefit)

### Mode evaluation table

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | not selected | Multi-milestone semantic content changes across 80+ files; not a single-line fix |
| 2 background | not selected | Write/Edit operations required (CONST-V3R2-020 forbids background writes) |
| 3 agent-team | not selected | Agent Teams capability gate not met; content-sweep is not multi-domain research-heavy |
| 4 parallel | not selected | Finding A4 — content/coding-heavy work, LOW concurrency benefit; shared working tree race risk |
| 5 sub-agent | **selected** | Tier L markdown/template sweep, sequential per-milestone manager-develop (cycle_type=ddd), Section A-E delegation |

### Decision

Decision: sub-agent

### Justification

Mode 5 (sub-agent, sequential per milestone) is the correct choice for this SPEC. Per Anthropic Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"), template-content sanitization is content-heavy with LOW concurrency benefit — the work is a deterministic grep+rewrite over a shared working tree where parallel spawn would race on the same files (M3 alone touches 73 markdown files). The orchestration-mode-selection.md tie-breaker rule for "Tier L + markdown / shell-script-only scope" explicitly prescribes Mode 5 with the Tier L Section A-E delegation template. The milestones run sequentially (M2 → M3 → M4 → M5) with a checkpoint commit per milestone.

## Plan-Phase Summary

Tier L SPEC. 13 GEARS REQs (11 active + REQ-TNA-003/007 deferred to ISOLATION-001) + 11 active binary ACs (+ AC-TNA-003/007 deferred markers) + 6 milestones (M1–M6). M1 already executed (commit `367a84715`); run-phase enters at M2.

### Rescope (v0.1.1, plan-audit iter-1 — FAIL 0.71 remediation)

- **KEPT in NEUTRALITY** (unique, not covered by ISOLATION leak test): C1 macOS paths, C2 `V3R[0-9]` refs (the bulk, 73 files), C4 `feedback_`/`memory.md` refs, C5 `CLAUDE.local.md` refs, C6 `PR #N` refs, C8 `GOOS=` PRESERVE.
- **DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001** (`internal_content_leak_test.go` strict-tier owns these): C3 dates (`S1-internal-date`), C7 commit hash (`S2-short-sha-sentence-final`).
- **C4 kept despite SCOPE defer instruction**: the leak test does NOT enforce the `feedback_`/`memory.md` substring class (default OR strict) — its C5 enforces only memory *paths*, a disjoint pattern. Deferring C4 would silently drop enforcement, so C4 stays NEUTRALITY-owned (verified 2026-05-30).
- **D2 AC awk fix**: AC-TNA-002/004/005 rewritten with non-self-terminating awk `awk '/^### CN /{f=1;next} /^### C[0-9] /{f=0} f'` (the original range `awk '/^### CN /,/^### C[0-9]+/'` self-terminated → allow-list count = 0 → `actual <= 0` impossible).

### Deliverables (this phase)

- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md` (§1–§5)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/plan.md` (§1 Milestones + §2 Risks + §3 Dependencies + Out of Scope)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/acceptance.md` (§1 ACs + §2 Test plan + Out of Scope)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/progress.md` (this file)

### Baseline measurements (re-measured 2026-05-30 at HEAD `ecda4ef04`, absolute-path grep — point-in-time; run-phase M-stages re-measure before fixing)

| Category | Pattern | Files (2026-05-30) | Files (2026-05-23) | Partition |
|---|---|--:|--:|---|
| C1 macOS-bias | `/Users/` | 4 | 4 | KEEP (binary) |
| C2 V3R refs | `V3R[0-9]` | **73** | 70 | KEEP (advisory, allow-list 15) |
| C3 Dates | `2026-0[5-9]` | 39 | 32 | **DEFER → ISOLATION** (informational) |
| C4 feedback/memory | `feedback_\|memory\.md` | 9 | 9 | KEEP (advisory, allow-list 7) |
| C5 CLAUDE.local | `CLAUDE\.local\.md` | **3** | 10 | KEEP (binary, allow-list empty) |
| C6 PR #N | `PR #[0-9]+` | 3 | 3 | KEEP (binary) |
| C7 Commit hash | hex 7-8 (leak-test S2) | ~2 | ~2 | **DEFER → ISOLATION** (informational) |
| C8 False positive | `GOOS=(linux\|windows\|darwin\|...)` | 3 (4 hits) | 3 (4 hits) | KEEP (PRESERVE) |

KEPT-class scope (C1/C2/C4/C5/C6/C8). Drift since plan-phase: C2 +3, C3 +7 (deferred), C5 −7 (partial prior cleanup). Run-phase M-stages re-measure at run HEAD before applying fixes.

## Milestone Status

| Milestone | Status | Owner | Trigger |
|---|---|---|---|
| M1 — SPEC scope finalize + allow-list draft | **complete** (commit `367a84715`; migration-matrix.md 232L, 8/8 sections, AC-TNA-010 PASS) | orchestrator-direct | (done) |
| M2 — C1 macOS-bias fix (4 files) | **complete** (commit `1046c6a3c`; AC-TNA-001 PASS, C1=0) | manager-develop ddd Tier L | (done) |
| M3 — C2 bare-narrative classification + fix (v0.1.2 narrowed; 7→6 files) | **complete** (AC-TNA-002 PASS, `actual=6 ≤ allowlist=6`) | manager-develop ddd Tier L | (done) |
| M4 — memory + CLAUDE.local refs (C4+C5, ~12 files; C3 deferred) | in-progress | manager-develop ddd Tier L | After M3 |
| M5 — PR refs + audit script + CI guard (C6; C7 deferred) | pending | manager-develop ddd Tier L | After M4 |
| M6 — Migration matrix finalize + guideline + chore implemented | pending | orchestrator-direct chore | After M5 |

## Next Steps

After this plan-audit iter-1 remediation (v0.1.1), expected sequence:

1. **plan-auditor iter-2** verification of the 5 artifacts (spec.md / plan.md / acceptance.md / migration-matrix.md / progress.md) with Tier L threshold ≥ 0.85. Expected to clear ≥ 0.85: the rescope deletes D1 (ISOLATION overlap), the awk fix resolves D2, the C7 defer shrinks D3, M1-complete resolves D4, baselines refresh resolves D6.
2. If plan-auditor PASS → GATE-2 (plan-to-implement HUMAN GATE) → run-phase enters at **M2** (M1 already complete, commit `367a84715`).
3. M2 (C1) → M3 (C2) → M4 (C4+C5; C3 deferred) → M5 (C6 + audit-script + CI-guard; C7 deferred) — sequential manager-develop delegation (Tier L Section A-E MANDATORY).
4. M6 chore + status update to `implemented v0.2.0`.

## Evidence Tracker

M1 evidence populated (migration-matrix.md shipped `367a84715`). M2–M6 populated during run-phase.

| AC | M-source | Evidence | Verified |
|---|---|---|---|
| AC-TNA-001 | M2 | `grep -rln '/Users/' internal/template/templates/` = 0 (commit `1046c6a3c`, 4 files C1-sanitized) | ☑ |
| AC-TNA-002 | M3 | **PASS** — C2 bare-narrative `actual=6 ≤ allowlist=6` (v0.1.2 narrowed scope). `manager-develop-prompt-template.md` GENERALIZEd both template + `.claude/` sides (identical edit, mirror delta -53 preserved, no new drift); 6 PRESERVE files remain (zone-registry namespace + manager-spec decomposition example + 4 harness V3R4-doctrine citations) | ☑ |
| AC-TNA-003 | — | **DEFERRED → ISOLATION-001** (leak-test `S1-internal-date`) | n/a |
| AC-TNA-004 | M4 | corrected awk allow-list=7 computable (D2 fix verified 2026-05-30) | ☐ (run-phase) |
| AC-TNA-005 | M4 | binary; allow-list empty (corrected awk=0) | ☐ (run-phase) |
| AC-TNA-006 | M5 | (pending) | ☐ |
| AC-TNA-007 | — | **DEFERRED → ISOLATION-001** (leak-test `S2-short-sha-sentence-final`) | n/a |
| AC-TNA-008 | M5 | (pending) | ☐ |
| AC-TNA-009 | M5 | (pending) | ☐ |
| AC-TNA-010 | M1 | **PASS** — migration-matrix.md 8/8 sections, `grep -cE '^### C[1-8] '` = 8 (commit `367a84715`, re-verified 2026-05-30) | ☑ |
| AC-TNA-011 | M5 | C8 baseline = 3 files preserved (verified 2026-05-30) | ☐ (run-phase) |
| AC-TNA-012 | M6 | (pending) | ☐ |
| AC-TNA-013 | M6 | (pending) | ☐ |

## Run-phase M3 Blocker (2026-05-30) — **RESOLVED by manager-spec C2 narrow (commit `c7c7b4e32`, v0.1.2)**

M2 complete (`1046c6a3c`, AC-TNA-001 PASS). M3 was initially **blocked** by a scope/design conflict; **manager-spec narrowed C2 to bare-narrative only (v0.1.2, commit `c7c7b4e32`)** which resolves Finding 1. Findings 2 (package RED) and 3 (mirror drift) are acknowledged as OUT-OF-SCOPE in spec.md §3.4. M3 then completed: bare-narrative C2 `actual=6 ≤ allowlist=6` (AC-TNA-002 PASS), `manager-develop-prompt-template.md` GENERALIZEd on both template + `.claude/` sides identically (mirror delta -53 preserved). Original blocker analysis (historical record):

1. **C2 target unachievable within scope**: 73 files contain `V3R[0-9]`. After excluding `zone-registry.md` (128 hits, allow-listed PRESERVE) and `CONST-V3R5-NNN` registry-ID citations (allow-listed PRESERVE), the residual bulk of `V3R[0-9]` hits are inside `SPEC-V3R6-*` SPEC-ID literals (~50+ files). Only ~6 files carry genuinely bare narrative `V3R4`/`V3R6` tokens (manager-spec.md decomposition example line 149, harness.md/moai SKILL.md "V3R4 Self-Evolving Harness", manager-develop-prompt-template.md Korean prose). Reducing `V3R[0-9]` to ≤18 would require removing/generalizing the `SPEC-V3R6-*` IDs from ~50 files — which the partition table (spec.md §3.3) assigns to **ISOLATION-001's `C1-spec-id-prefix` leak-test class, NOT NEUTRALITY's C2 class**, and which Section B2/D constraints forbid (must keep C2 disjoint from the leak test).

2. **`internal_content_leak_test.go` already RED (30 pre-existing violations)**: ISOLATION-001's CI guard (a `completed` SPEC, same Go package) is failing on `main` @ `a9757f484` with 30 `C1-spec-id` / `C2-req-ac` / `C4-finding` / `C5-archive-path` violations — independent of this SPEC. M2 introduced 0 regressions (its 5 files are disjoint from all failures). The `internal/template` package has **13 pre-existing failing test functions** at the M2 parent commit, including `TestRuleTemplateMirrorDrift` (6 workflowOpt mirror files drifted) + `TestLateBranchTemplateMirror` (manager-git/manager-spec) + `TestTemplateAgentsStructure`.

3. **Mirror-parity coupling**: 4 of the M3/M4/M5 C2/C5/C6 target files (`manager-develop-prompt-template.md`, `manager-spec.md`, `spec-workflow.md`, `manager-git.md`) sit on the byte-parity mirror allow-list (`rule_template_mirror_test.go`). Editing the template side without the operational `.claude/` side (or vice-versa) trips `RULE_TEMPLATE_MIRROR_DRIFT`. The operational `.claude/` mirrors are ALREADY ahead (genericized per CLAUDE.local.md §25) while the template mirrors lag — a pre-existing drift this SPEC's scope does not cover.

**Decision required** (manager-spec scope-doc update + orchestrator GATE): the C2 ≤18 target conflicts with the C2/C1-spec-id partition. Resolution options surfaced in the blocker report.

## Notes

- **fc47f31a7 prior context**: 본 SPEC은 commit `fc47f31a7`의 Critical 4 violations fix 이후 audit에서 추가 발견된 138 file scope를 다룬다. fc47f31a7의 3 변경 (SKILL.md / spec-frontmatter-schema.md / sprint-round-naming.md 신규)은 본 SPEC scope **외** — 이미 머지 완료, 재발 방지는 본 SPEC의 audit script (REQ-TNA-009) + CI guard (REQ-TNA-010)가 담당.
- **Sprint 1 cross-Sprint caveat**: SKILL-GEARS-ALIGN-001 (plan-complete `81d47a445`, 머지 대기) + CODE-COMMENTS-EN-001 Wave 4+ (single-SPEC within-phase, 진행 중)와 disjoint scope. PRESERVE list로 충돌 차단. Sprint = multi-SPEC group; Wave/Round = single-SPEC internal phase (per `.claude/rules/moai/development/sprint-round-naming.md`).
- **GEARS notation self-dogfooding**: §4 13 REQs 모두 Where/When/While/If-Then GEARS keywords 사용. 본 SPEC은 SKILL-GEARS-ALIGN-001 doctrine를 verbatim 적용한다.
- **Frontmatter canonical 12-field SSOT**: 4개 파일 모두 `.claude/rules/moai/development/spec-frontmatter-schema.md`의 12 필드 + optional `tier` 필드 사용. snake_case alias (`created_at`, `updated_at`, `labels`) 미사용.
