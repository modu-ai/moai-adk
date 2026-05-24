---
id: SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001
title: "moai-foundation-core SKILL bundle을 GEARS 우선 가이드로 정렬 — Plan"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai-foundation-core, internal/template/templates/.claude/skills/moai-foundation-core"
lifecycle: spec-anchored
tags: "gears, ears, skill, foundation, core, guide, alignment, plan, sprint-10, v3.0.0"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001]
---

# SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 — Implementation Plan

## §A Lifecycle

| Phase | Status | Owner | Commit SHA | Date |
|-------|--------|-------|------------|------|
| plan | draft | manager-spec | (pending) | 2026-05-25 |
| run M1 | pending | manager-develop | — | — |
| run M2 | pending | manager-develop | — | — |
| run M3 | pending | manager-develop | — | — |
| run M4 | pending | manager-develop | — | — |
| run M5 | pending | manager-develop | — | — |
| run M6 | pending | manager-develop | — | — |
| sync | pending | manager-docs | — | — |
| Mx | pending | orchestrator | — | — |

## §B Run-phase Strategy

### B.1 Pattern Selection

Mirror the **SKILL-GEARS-ALIGN-001 Tier M 1-pass pattern**:
- Single manager-develop spawn covering all 6 milestones M1-M6
- Phase 0.5 plan-auditor enforcement (Tier M not skip-eligible at 0.85 PASS, skip-eligible 0.90+)
- Path-specific `git add` per L46 attribution discipline (exact file list per milestone, never `git add .`)
- Markdown-only edits; no Go source; no test changes; no docs-site

### B.2 Self-dogfood Verification

Every REQ-FCG-XXX in spec.md MUST be in GEARS notation (Ubiquitous / `When` / `While` / `Where` / `When <event-detected>`). Run-phase MUST NOT introduce `IF <condition> THEN <action>` modality in any updated file content.

### B.3 Predecessor Pattern Fidelity

Follow the cohort pattern established by SPEC-V3R6-SKILL-GEARS-ALIGN-001 + SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001:

| Predecessor | Tier | Files | REQs | ACs | Plan-auditor | Phase | Outcome |
|-------------|------|-------|------|-----|--------------|-------|---------|
| SKILL-GEARS-ALIGN-001 | M | 5+5 | 12 GEARS | 13 PASS | 0.892 (PASS, not skip) | 1-pass run | CLOSED `ebe492670` |
| PLAN-AUDITOR-GEARS-ALIGN-001 | S | 1+1 | 9 GEARS | 8 PASS | 0.913 (skip-eligible) | 1-pass run | CLOSED `ebe492670` |
| **This SPEC (FOUNDATION-CORE-GEARS-ALIGN-001)** | **M** | **≤7+7 (14 of 40 scope)** | **12 GEARS** | **9 mandatory** | **target ≥0.85 PASS, ideal ≥0.90 skip-eligible** | **1-pass target** | **(pending)** |

## §C Scope

### §C.1 Discovered File Inventory (40-file scope; ≤14-file actual edit set)

[HARD] Honest count per L46 attribution discipline. User prompt estimated "10 files"; actual discovered scope = 40 files (20 local + 20 template mirror).

**Local files (20)**:
```
.claude/skills/moai-foundation-core/SKILL.md
.claude/skills/moai-foundation-core/modules/INDEX.md
.claude/skills/moai-foundation-core/modules/agents-reference.md
.claude/skills/moai-foundation-core/modules/commands-reference.md
.claude/skills/moai-foundation-core/modules/delegation-advanced.md
.claude/skills/moai-foundation-core/modules/delegation-implementation.md
.claude/skills/moai-foundation-core/modules/delegation-patterns.md
.claude/skills/moai-foundation-core/modules/execution-rules.md
.claude/skills/moai-foundation-core/modules/modular-system.md
.claude/skills/moai-foundation-core/modules/patterns.md
.claude/skills/moai-foundation-core/modules/progressive-disclosure.md
.claude/skills/moai-foundation-core/modules/spec-ddd-implementation.md
.claude/skills/moai-foundation-core/modules/spec-ears-format.md  [PROTECTED — EXC-FCG-008]
.claude/skills/moai-foundation-core/modules/spec-first-ddd.md
.claude/skills/moai-foundation-core/modules/token-optimization.md
.claude/skills/moai-foundation-core/modules/trust-5-framework.md
.claude/skills/moai-foundation-core/modules/trust-5-implementation.md
.claude/skills/moai-foundation-core/modules/trust-5-validation.md
.claude/skills/moai-foundation-core/references/examples.md
.claude/skills/moai-foundation-core/references/reference.md
```

**Template mirror files (20)**: 1:1 mirror at `internal/template/templates/.claude/skills/moai-foundation-core/` — same 20 paths. Verified via `find` 2026-05-25.

**Actual edit set (preliminary ≤7 local + 7 template mirror = 14 files)** — finalized at M1 boundary after re-confirmation grep:

| # | Path (local; mirror is parallel) | Edit type |
|---|----------------------------------|-----------|
| 1 | `SKILL.md` | § "SPEC-First DDD - Development Workflow" lines 116, 122 EARS→GEARS re-label + line 238 cross-reference label |
| 2 | `modules/INDEX.md` | Lines 32-46 (spec-first-ddd description), 188 cross-ref label |
| 3 | `modules/spec-first-ddd.md` | Lines 3, 14, 29-34 (EARS pattern table → GEARS), 37, 55, 63, 153, 161 cross-refs |
| 4 | `modules/spec-ddd-implementation.md` | Line 288 cross-ref label |
| 5 | `modules/progressive-disclosure.md` | Lines 184, 193, 201 EARS→GEARS re-label |
| 6 | `modules/commands-reference.md` | Lines 78, 90, 94-102 (EARS Format → GEARS Format with deprecated EARS sub-block) |
| 7 | `references/examples.md` | Lines 65-69 EARS block → GEARS block + EARS legacy sub-block |
| 8 | `references/reference.md` | Lines 20 (format declaration), 198, 377, 410, 447 EARS→GEARS with legacy cross-ref labels |
| 9 (PROTECTED) | `modules/spec-ears-format.md` | **NOT edited** per EXC-FCG-008 + REQ-FCG-005 + REQ-FCG-011 (banner already present) |
| 10 (No edit) | `modules/patterns.md` + 9 other modules | Verify in M2 grep that they contain no SPEC notation references |

Final edit set scope: ≤8 local files + ≤8 template mirrors = **≤16 files modified**. Plan.md §C.1 will be updated at M1 boundary if discovery reveals additional or fewer needed edits.

### §C.2 Out of Scope (mirrors spec.md §1.2)

- Go source code modification (EXC-FCG-001 / EXC-FCG-002)
- 4-locale docs-site modification (EXC-FCG-003)
- `.claude/skills/moai-workflow-spec/**` (EXC-FCG-004)
- `.claude/agents/meta/plan-auditor.md` (EXC-FCG-005)
- `.claude/agents/core/manager-spec.md` (EXC-FCG-006)
- 88 legacy SPEC rewrites (EXC-FCG-007)
- `modules/spec-ears-format.md` body (EXC-FCG-008 + REQ-FCG-005 + REQ-FCG-011)
- Predecessor SPEC bodies (EXC-FCG-009)
- Sprint 10 downstream cohort SPECs (EXC-FCG-010)

## §D Milestone Decomposition (M1-M6)

### M1 — Pre-flight Verification + Edit Set Finalization

**Owner**: manager-develop
**Scope**: Read-only verification + edit set lock-in
**Steps**:
1. `grep -rn 'shall\|WHEN\|WHILE\|WHERE\|IF.*THEN\|EARS\|GEARS' .claude/skills/moai-foundation-core/` — final inventory
2. `diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/` — confirm zero pre-existing drift (baseline established)
3. Confirm `modules/spec-ears-format.md` v3.0.0 DEPRECATED banner present at lines 9-15 (REQ-FCG-005 + REQ-FCG-011 compliance baseline)
4. Update plan.md §C.1 if discovery reveals additional or fewer needed edits than preliminary 8 files
5. Confirm Phase 0.5 plan-auditor PASS (≥0.85) before proceeding to M2

**Verification**: `git diff` empty (no edits in M1); plan.md §C.1 reflects final edit set
**Exit criteria**: plan-auditor signal received OR plan-auditor SKIP-eligible (≥0.90)

### M2 — SKILL.md + INDEX.md Re-label

**Owner**: manager-develop
**Scope**: 2 files local + 2 files template mirror = 4 files
**Steps**:
1. `SKILL.md` line 116: "Phase 1 SPEC (/moai:1-plan): workflow-spec generates **GEARS format** (primary; EARS retained as 6-month backward-compat legacy)."
2. `SKILL.md` line 122: Replace "EARS Format:" block with "GEARS Format (current):" block + 5-pattern table mirroring `moai-workflow-spec/SKILL.md` § "GEARS Format"; retain "EARS Format (legacy reference, 6-month backward-compat):" sub-block referencing 88 existing SPECs
3. `SKILL.md` line 238: Re-label cross-reference "workflow-spec for EARS format specification" → "workflow-spec for GEARS format specification (current) and EARS legacy reference"
4. `modules/INDEX.md` lines 32-46: Update spec-first-ddd description from "EARS format requirements" to "GEARS format requirements (current; EARS as legacy)"
5. `modules/INDEX.md` line 188: Add label "(legacy reference)" beside spec-ears-format.md
6. Mirror all 5 edits to `internal/template/templates/.claude/skills/moai-foundation-core/{SKILL.md, modules/INDEX.md}`

**Verification**: `diff -r` local vs template MUST show zero diff post-mirror; `grep -c 'GEARS' SKILL.md` ≥3 (was 0)
**Exit criteria**: 4 files updated, mirror parity confirmed

### M3 — spec-first-ddd.md + spec-ddd-implementation.md + progressive-disclosure.md Re-label

**Owner**: manager-develop
**Scope**: 3 files local + 3 files template mirror = 6 files
**Steps**:
1. `modules/spec-first-ddd.md` lines 3, 14, 29-34, 37, 55, 63, 153, 161: Replace "EARS format" / "EARS Patterns:" / "EARS Format Reference" with GEARS-primary phrasing; pattern table (lines 29-34) rewritten with five GEARS patterns including `Where <capability>` capability gate and `When <event-detected>` replacing `IF/THEN`
2. `modules/spec-ddd-implementation.md` line 288: Cross-reference label "(legacy reference)" added to `spec-ears-format.md` link
3. `modules/progressive-disclosure.md` lines 184, 193, 201: Re-label "EARS format" → "GEARS format (current)" with EARS legacy parenthetical
4. Mirror all 3 file edits to template

**Verification**: `diff -r` zero diff post-mirror; `grep -n 'IF.*THEN' modules/spec-first-ddd.md` returns 0 (REQ-FCG-007 self-check)
**Exit criteria**: 6 files updated, mirror parity confirmed

### M4 — commands-reference.md Re-label

**Owner**: manager-develop
**Scope**: 1 file local + 1 file template mirror = 2 files
**Steps**:
1. `modules/commands-reference.md` line 78: "Purpose: Generate SPEC document in **GEARS format (current)** with EARS legacy reference support"
2. Line 90: "Generates **GEARS format** SPEC document"
3. Lines 94-102: Replace "EARS Format (5 sections): WHEN/WHILE/WHERE/SHALL/SHALL NOT" block with "GEARS Format (5 patterns; current): Ubiquitous / `When <event>` / `While <state>` / `Where <capability>` / `When <event-detected>`" + "EARS Format (legacy reference, 6-month backward-compat):" sub-block
4. Line 102: Output description label "EARS document" → "GEARS document (or EARS legacy)"
5. Mirror to template

**Verification**: `diff -r` zero diff; `grep -c 'GEARS' commands-reference.md` ≥3
**Exit criteria**: 2 files updated, mirror parity confirmed

### M5 — references/examples.md + references/reference.md Re-label + Template Mirror Final Sync

**Owner**: manager-develop
**Scope**: 2 files local + 2 files template mirror = 4 files
**Steps**:
1. `references/examples.md` lines 65-69: Replace EARS Format block with GEARS Format (current) block including:
   - `[Ubiquitous] The system shall hash passwords using bcrypt.`
   - `[Event-driven] When user submits credentials, the system shall validate and return JWT.`
   - `[State-driven] While token is valid, the user shall access protected resources.`
   - `[Where, capability] Where SSO is enabled, the auth service shall federate identity to upstream IdP.` (generalized subject demonstration)
   - `[Unwanted] The system shall not store plain-text passwords.`
   - Underneath: "**EARS Format (legacy reference, 6-month backward-compat):**" sub-block with the original 4 examples preserved verbatim
2. `references/reference.md` line 20: "Format: GEARS (Generalized EARS) — primary notation; EARS retained as legacy reference for 6-month backward-compat window"; correct the EARS expansion from "Event-Action-Response-State" to "Easy Approach to Requirements Syntax" at lines 198, 377, 410, 447 wherever EARS is referenced
3. Mirror to template
4. **Final template parity check**: `diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/` MUST return zero diff (M2+M3+M4+M5 cumulative parity assertion)
5. **`make build` execution**: regenerate `internal/template/embedded.go` (CRITICAL — Template-First Rule per CLAUDE.local.md §2)

**Verification**: `diff -r` zero diff; `make build` succeeds; `git diff internal/template/embedded.go` non-empty (template content changed)
**Exit criteria**: 4 files updated + embedded.go regenerated, mirror parity confirmed

### M6 — Progress.md Run-phase Audit-Ready Signal + Frontmatter Status Transition

**Owner**: manager-develop
**Scope**: 1 file (progress.md run-phase section update + spec.md frontmatter `status: draft → in-progress` per Status Transition Ownership Matrix `manager-develop owns`)
**Steps**:
1. Update spec.md frontmatter: `status: draft → in-progress`, `updated: 2026-05-25` (Status Transition Ownership Matrix — manager-develop owns `draft → in-progress`)
2. Update progress.md §E.2 Run-phase Evidence with:
   - 6-milestone commit list (M1-M6 SHAs)
   - 9 AC verification results (PASS/FAIL/PASS-with-note)
   - `diff -r` zero-diff verification log
   - `make build` regenerated `embedded.go` verification
   - `grep -n 'IF.*THEN'` self-lint result (REQ-FCG-007 / REQ-FCG-012 confirmation)
3. Update progress.md §E.3 Run-phase Audit-Ready Signal: "all 6 milestones complete, 9 AC PASS, template mirror parity verified, `make build` regenerated, ready for sync-phase"

**Verification**: progress.md §E.2/§E.3 populated; spec.md frontmatter shows `status: in-progress`; commit attributed via path-specific `git add` per L46
**Exit criteria**: Run-phase complete, ready for manager-docs sync-phase

## §E Verification Strategy

### E.1 Pre-spawn Verification (orchestrator owns)

Before spawning manager-develop:
1. Phase 0.5 plan-auditor PASS (≥0.85 PASS threshold) OR SKIP-eligible (≥0.90 — Tier M margin) → enables 1-pass implementation
2. `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → MUST be `0 0` clean (Pre-Spawn Sync Check per agent-common-protocol.md)
3. `ls .moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/` → 4 artifacts present (spec.md + plan.md + acceptance.md + progress.md)

### E.2 Post-implementation Verification (orchestrator independent verification batch)

Trust-but-verify 7-item batch (parallel single-turn per agent-common-protocol.md §Parallel Execution):

1. `git log --oneline -10` — commits attributed
2. `git rev-list --count --left-right origin/main...HEAD` — `0 0` post-push
3. `grep -rn 'IF.*THEN\|IF .* THEN' .claude/skills/moai-foundation-core/ | grep -v spec-ears-format.md` — MUST return 0 matches (REQ-FCG-007 + REQ-FCG-012 enforcement; legacy reference file body protected per EXC-FCG-008)
4. `diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/` — MUST return zero diff (REQ-FCG-010 template mirror parity)
5. `head -20 .claude/skills/moai-foundation-core/modules/spec-ears-format.md` — confirm DEPRECATED banner lines 9-15 unchanged verbatim (REQ-FCG-005 + REQ-FCG-011 SSOT preservation)
6. `grep -c 'GEARS' .claude/skills/moai-foundation-core/SKILL.md` — MUST return ≥3 (was 0; primary notation transition evidenced)
7. `go test ./internal/template/...` — embedded.go regenerated correctly via `make build`

### E.3 Plan-auditor Self-audit Estimate

| Dimension | Predecessor SKILL-GEARS-ALIGN-001 | Predecessor PLAN-AUDITOR-GEARS-ALIGN-001 | This SPEC (estimate) |
|-----------|------------------------------------|------------------------------------------|----------------------|
| MP-1 Goal clarity | 0.93 | 0.95 | **0.92** |
| MP-2 EARS/GEARS notation discipline | 0.90 (self-dogfood) | 0.95 (self-dogfood) | **0.92** (self-dogfood) |
| MP-3 Scope discipline | 0.87 | 0.93 | **0.85** (40-file scope, ≤14-file edit set documented honestly) |
| MP-4 Risk identification | 0.85 | 0.88 | **0.88** (6 risks vs predecessor 5-6) |
| MP-5 Traceability | 0.91 | 0.92 | **0.90** |
| **Aggregate (weighted)** | **0.892** | **0.913** | **~0.875-0.90** |

**Estimated plan-auditor**: ~0.87-0.90 (PASS, marginally skip-eligible). Run-phase 1-pass success probability **HIGH** (Tier M precedent established by SKILL-GEARS-ALIGN-001 with the same pattern).

## §F Anti-Patterns to Avoid

- ❌ Modifying `internal/spec/lint.go` (out of scope; canonical migration done in SPEC-V3R6-GEARS-MIGRATION-001 M2)
- ❌ Modifying `modules/spec-ears-format.md` body (EXC-FCG-008 — file already aligned with deprecated banner)
- ❌ Modifying any of the 3 predecessor SPEC bodies (EXC-FCG-009 — SSOT preservation)
- ❌ Using `git add .` or `git add -A` (L46 attribution discipline violation; mandatory path-specific add)
- ❌ Skipping `make build` post-template-edit (CLAUDE.local.md §2 Template-First Rule violation)
- ❌ Adding `IF <condition> THEN <action>` modality in any updated content (REQ-FCG-007 / REQ-FCG-012 violation)
- ❌ Removing 88 legacy SPEC EARS examples or claiming early backward-compat expiry (REQ-FCG-011 + Sprint 10 cohort policy violation; 6-month window expires 2026-11-22)
- ❌ Expanding scope into downstream Sprint 10 cohort SPECs (EXC-FCG-010 — each owned by separate SPEC)

## §G Cross-References

- Predecessor SPECs (depends_on chain):
  - SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0 (lint engine + 4-locale docs-site, PR #1046)
  - SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0 (moai-workflow-spec authoring guide, commits `ebe492670`)
  - SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 v0.2.0 (plan-auditor agent, commits `ebe492670`)
- Status Transition Ownership Matrix: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- Agent-common-protocol Parallel Execution + Pre-Spawn Sync Check: `.claude/rules/moai/core/agent-common-protocol.md`
- Plan-auditor scoring rubric: `.claude/agents/meta/plan-auditor.md`
- Template-First Rule: `CLAUDE.local.md` §2 [HARD]
- Path-specific add discipline (L46): MEMORY.md historical entries 2026-05-23 onwards
- Sprint 10 cohort entry sequencing: MEMORY.md entry "Sprint 10 GEARS sweep cohort 2/8 close + 6 SPECs paste-ready" (2026-05-25)
