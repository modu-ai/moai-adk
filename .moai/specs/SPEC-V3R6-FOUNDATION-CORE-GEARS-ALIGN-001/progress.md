---
id: SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001
title: "moai-foundation-core SKILL bundle을 GEARS 우선 가이드로 정렬 — Progress"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai-foundation-core, internal/template/templates/.claude/skills/moai-foundation-core"
lifecycle: spec-anchored
tags: "gears, ears, skill, foundation, core, progress, sprint-10, v3.0.0"
tier: M
issue_number: null
plan_commit_sha: "<pending>"
run_commit_sha: "<pending>"
sync_commit_sha: "<pending>"
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001]
---

# SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 — Progress

## §A Lifecycle

| Phase | Status | Owner | Commit SHA | Date | Notes |
|-------|--------|-------|------------|------|-------|
| plan-phase | in-progress | manager-spec | (pending) | 2026-05-25 | 4 artifacts authored: spec.md (12 GEARS REQs + 10 EXCs + 6 risks), plan.md (6 milestones M1-M6 + verification strategy), acceptance.md (9 mandatory ACs + 4 edge cases + traceability matrix), progress.md (this file). Plan-phase audit-ready signal § E.1 below. |
| run-phase | pending | manager-develop | — | — | Awaiting plan-phase commit + Phase 0.5 plan-auditor PASS. |
| sync-phase | pending | manager-docs | — | — | Awaiting run-phase completion. |
| Mx-phase | pending | orchestrator | — | — | Awaiting sync-phase completion; markdown-only run-phase suggests Mx Step C EVALUATE-SKIP likely (no .go files modified, no goroutines, no fan_in changes). |

## §B Known Issues + Conflicts

### B.1 Discovered file count exceeds initial estimate

**Issue**: User prompt provided "10 files" estimate; actual `find` discovery returned 40 files (20 local + 20 template mirror, 1:1 parity).
**Resolution**: Documented honestly in plan.md §C.1 per L46 attribution discipline. Actual edit set scope is ≤8 local files + ≤8 template mirror = ≤16 files (subset of 40-file discovery scope). Files intentionally NOT edited (12 local + 12 mirror) include modules with no SPEC notation content (trust-*, delegation-*, token-optimization, modular-system, agents-reference, execution-rules, patterns).

### B.2 `modules/spec-ears-format.md` already partially aligned

**Issue**: Predecessor SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0 already updated this file's banner (lines 9-15 carry the v3.0.0 DEPRECATED notice pointing at canonical GEARS guide).
**Resolution**: EXC-FCG-008 + REQ-FCG-005 + REQ-FCG-011 + AC-FCG-007 collectively protect this file's body from any further modification. Only sibling cross-references to it MAY be re-labeled "(legacy reference, deprecated)" without touching the target file.

### B.3 Sprint 10 cohort coordination

**Issue**: 5 downstream Sprint 10 cohort SPECs (WORKFLOW-PLAN, DOCS-SITE-FULL, WORKFLOW-SPEC-EXTRAS, MISC-DOCS, RULES-GO-DOCS) are not yet authored but reference `moai-foundation-core` for foundational patterns.
**Resolution**: This SPEC (cohort entry #3 of 8) lands before downstream cohort, providing the GEARS-aligned foundation. Cohort sequencing documented in MEMORY.md entry "Sprint 10 GEARS sweep cohort 2/8 close + 6 SPECs paste-ready" (2026-05-25).

## §C Pre-flight Requirements (orchestrator-owned before manager-develop spawn)

| Check | Status | Notes |
|-------|--------|-------|
| spec.md frontmatter 12 canonical fields present | ✓ | All present + optional depends_on/related_specs/tier populated |
| spec.md frontmatter `status: draft` | ✓ | manager-spec owns initial draft per Status Transition Ownership Matrix |
| 4 artifacts present (spec.md + plan.md + acceptance.md + progress.md) | ✓ | All authored 2026-05-25 |
| Predecessor SPEC IDs verified | ✓ | `ls .moai/specs/ \| grep GEARS` confirmed 3 predecessors |
| Discovered file inventory documented in plan.md §C.1 | ✓ | 40-file scope, ≤16-file edit set, intentional-skip list itemized |
| GEARS self-dogfood verified (≥80% REQs in GEARS) | ✓ | 12/12 = 100% GEARS notation (Ubiquitous, Event-driven `When`, State-driven `While`, Where capability, generalized subject) |
| Phase 0.5 plan-auditor PASS (≥0.85 Tier M) | pending | Estimate ~0.87-0.90 per plan.md §E.3 |

## §D Constraints

- [HARD] manager-spec MUST NOT invoke AskUserQuestion (subagent boundary; canonical reference `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator–Subagent Boundary)
- [HARD] Predecessor SPEC bodies MUST NOT be modified (EXC-FCG-009 + AC-FCG-008)
- [HARD] `modules/spec-ears-format.md` body MUST NOT be modified (EXC-FCG-008 + REQ-FCG-005 + REQ-FCG-011 + AC-FCG-007)
- [HARD] Template mirror parity MUST be verified zero-diff post-edits (REQ-FCG-010 + AC-FCG-005)
- [HARD] `make build` MUST regenerate `embedded.go` after template edits (CLAUDE.local.md §2 Template-First Rule + AC-FCG-006)
- [HARD] Path-specific `git add` MUST be used (L46 attribution discipline; `git add .` PROHIBITED)
- [HARD] No new `IF/THEN` modality in edited content (REQ-FCG-007 + REQ-FCG-012 + AC-FCG-004)

## §E Self-Verification Signals

### §E.1 Plan-phase Audit-Ready Signal

**Plan-phase complete (2026-05-25)** — 4 artifacts authored, ready for Phase 0.5 plan-auditor review.

**Self-audit estimate**:
- MP-1 Goal clarity: 0.92 (single objective: foundation-tier GEARS alignment; well-bounded)
- MP-2 EARS/GEARS notation discipline: 0.92 (12/12 REQs in GEARS notation; self-dogfood verified; predecessor pattern fidelity)
- MP-3 Scope discipline: 0.85 (40-file discovery documented honestly; ≤16-file edit set bounded; 10 EXCs cover all out-of-scope boundaries)
- MP-4 Risk identification: 0.88 (6 risks R1-R6 covering self-lint, mirror drift, scope creep, file count variance, cohort coordination, partial pre-alignment)
- MP-5 Traceability: 0.90 (9 ACs cover all 12 REQs + 10 EXCs via §G traceability matrix in acceptance.md)
- **Aggregate weighted estimate**: ~0.875-0.90

**Skip-eligibility (Tier M ≥0.90)**: MARGINAL — orchestrator MAY skip if aggregate ≥0.90, otherwise re-spawn plan-auditor for 1 iteration; either path expected to enable 1-pass run-phase per predecessor precedent.

**Predecessor pattern fidelity**:
- SKILL-GEARS-ALIGN-001 (Tier M): 0.892 → 1-pass run-phase success → 4-phase CLOSED
- PLAN-AUDITOR-GEARS-ALIGN-001 (Tier S): 0.913 skip-eligible → 1-pass run-phase success → 4-phase CLOSED
- This SPEC (Tier M): ~0.875-0.90 estimate → 1-pass run-phase target

**Ready for**:
1. plan.md + spec.md + acceptance.md + progress.md commit attribution (orchestrator-owned)
2. Phase 0.5 plan-auditor spawn (orchestrator-owned)
3. manager-develop run-phase spawn (orchestrator-owned, post plan-auditor PASS/SKIP)

### §E.2 Run-phase Evidence

**M1 — Pre-flight Verification + Edit Set Finalization (2026-05-25)**

M1 verification log:

1. **Pre-spawn fetch**: `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0` (clean, no divergence)
2. **Comprehensive notation grep across all 20 local files** (`grep -rn 'EARS\|GEARS\|shall\|WHEN \|WHILE \|WHERE \|IF.*THEN' .claude/skills/moai-foundation-core/`):
   - 30 total matches across 8 distinct files (plan.md §5.1 baseline confirmed)
   - **D1 finding (orchestrator pre-flight)**: 2 additional files contain EARS references beyond initial 7-file estimate:
     - `modules/agents-reference.md:63` "EARS SPEC generation" (workflow-spec row)
     - `modules/token-optimization.md:44` "EARS format, acceptance criteria" (Phase 1 row)
   - These were classified as "no SPEC notation content" in plan.md §5.1 initial discovery — **corrected** per L46 attribution discipline
3. **Template mirror parity baseline**: `diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/` → EMPTY (zero pre-existing drift; 100% parity baseline established)
4. **DEPRECATED banner confirmation**: `head -20 .claude/skills/moai-foundation-core/modules/spec-ears-format.md` confirmed v3.0.0 DEPRECATED banner present verbatim at lines 9-15 (REQ-FCG-005 + REQ-FCG-011 baseline)
5. **Intentional-skip files verified clean** (no SPEC notation): trust-5-{framework,implementation,validation}.md (note: trust-5-implementation.md:178 is SQL `WHERE`, not requirements syntax), delegation-{patterns,implementation,advanced}.md, modular-system.md, execution-rules.md, patterns.md
6. **Final edit set FINALIZED**: 10 local files + 10 template mirrors = 20 files (vs initial estimate 8+8=16; +2 files per D1 surface)
7. **Plan.md §C.1 + §B.3 updated** in this M1 commit reflecting actual scope per L46 attribution discipline (orchestrator pre-authorized scope-doc update for D1 fix)

**M1 commit attribution** (this commit; SHA to be assigned by `git commit`): `feat(SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001): M1 pre-flight audit + scope finalization (D1+D3 plan fixes)`

Expected content from M2-M6 (pending):
- M2-M5 commit SHAs (one per milestone)
- M6 frontmatter transition log (`status: draft → in-progress`) + progress.md §E.2 final self-population
- 9 AC verification results (PASS/FAIL/PASS-with-note)
- 7-item Trust-but-verify batch results (per plan.md §E.2)

### §E.3 Run-phase Audit-Ready Signal

**(pending — manager-develop will populate after M6 complete)**

### §E.4 Sync-phase Audit-Ready Signal

**(pending — manager-docs will populate after sync-phase complete)**

### §E.5 Mx-phase Audit-Ready Signal

**(pending — orchestrator will populate after Mx Step C judgment: EVALUATE-PASS / EVALUATE-SKIP based on .go file delta = 0 expected per Tier M markdown-only nature)**

## §F Milestones (cross-reference plan.md §D)

| ID | Description | Owner | Status | Commit SHA |
|----|-------------|-------|--------|------------|
| M1 | Pre-flight verification + edit set finalization | manager-develop | pending | — |
| M2 | SKILL.md + INDEX.md re-label | manager-develop | pending | — |
| M3 | spec-first-ddd.md + spec-ddd-implementation.md + progressive-disclosure.md re-label | manager-develop | pending | — |
| M4 | commands-reference.md re-label | manager-develop | pending | — |
| M5 | references/examples.md + references/reference.md re-label + template parity + `make build` | manager-develop | pending | — |
| M6 | progress.md run-phase audit-ready signal + frontmatter `status: draft → in-progress` | manager-develop | pending | — |

## §G Anti-Patterns to Avoid (cross-reference plan.md §F)

(See plan.md §F for the full 8-item anti-pattern list. Critical reminders: no `internal/spec/lint.go` modification; no `modules/spec-ears-format.md` body modification; no predecessor SPEC body modification; no `git add .`; no skipping `make build`; no `IF/THEN` modality re-introduction; no early backward-compat termination; no scope creep into downstream cohort SPECs.)

## §H Cross-References

- spec.md: `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md`
- plan.md: `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/plan.md`
- acceptance.md: `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/acceptance.md`
- Predecessor 4-artifact reference: `.moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/` (CLOSED `ebe492670`), `.moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/` (CLOSED `ebe492670`)
- Canonical GEARS guide: `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format"
- Status Transition Ownership Matrix: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Agent-common-protocol Pre-Spawn Sync Check: `.claude/rules/moai/core/agent-common-protocol.md`
- Template-First Rule: `CLAUDE.local.md` §2
- Sprint 10 cohort memo: MEMORY.md "Sprint 10 GEARS sweep cohort 2/8 close + 6 SPECs paste-ready" (2026-05-25)
