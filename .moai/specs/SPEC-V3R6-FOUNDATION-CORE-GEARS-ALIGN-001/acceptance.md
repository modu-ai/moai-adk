---
id: SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001
title: "moai-foundation-core SKILL bundle을 GEARS 우선 가이드로 정렬 — Acceptance Criteria"
version: "0.1.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai-foundation-core, internal/template/templates/.claude/skills/moai-foundation-core"
lifecycle: spec-anchored
tags: "gears, ears, skill, foundation, core, acceptance, sprint-10, v3.0.0"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001]
sync_commit_sha: "a853f2954"
---

# SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 — Acceptance Criteria

## §A Definition of Done

This SPEC is DONE when ALL 9 mandatory ACs PASS, template mirror parity is verified zero-diff, `make build` regenerates `embedded.go` cleanly, and run-phase introduces zero new `LegacyEARSKeyword` self-lint warnings.

## §B Tier M Mandatory AC Count

9 mandatory ACs (AC-FCG-001..009) covering: GEARS notation re-label at SKILL bundle entry points, generalized-subject example demonstration, capability-gate (`Where`) example demonstration, IF/THEN deprecation enforcement on edited content, template mirror parity, `make build` regeneration, predecessor SSOT protection, `modules/spec-ears-format.md` body preservation, and GEARS notation count threshold.

## §C Given-When-Then Scenarios

### AC-FCG-001 — SKILL.md presents GEARS as primary notation

**Given**: A downstream agent or skill author loads `.claude/skills/moai-foundation-core/SKILL.md` after the run-phase completes
**When**: The agent reads § "SPEC-First DDD - Development Workflow" (current lines 110-130 area, approximately)
**Then**:
- The section MUST mention "GEARS" at least once before any "EARS" mention
- The five GEARS patterns (Ubiquitous, `When <event>`, `While <state>`, `Where <capability>`, `When <event-detected>`) MUST be listed
- Any retained EARS reference MUST be labeled "(legacy reference, 6-month backward-compat)"
- `grep -c 'GEARS' .claude/skills/moai-foundation-core/SKILL.md` MUST return ≥3

**Verification**: `grep -n 'GEARS\|EARS' .claude/skills/moai-foundation-core/SKILL.md` shows GEARS occurrences ordered before EARS legacy labels.

### AC-FCG-002 — references/examples.md demonstrates GEARS with generalized subject

**Given**: A downstream consumer reads `.claude/skills/moai-foundation-core/references/examples.md`
**When**: The consumer reaches the format examples block (current lines 60-80 area)
**Then**:
- The block MUST be labeled "GEARS Format (current):" as the primary heading
- At least one example MUST use a generalized subject (not "The system"; e.g., "The auth service shall …", "The agent shall …", "The component shall …")
- At least one example MUST use `Where <capability>` capability gate pattern
- An "EARS Format (legacy reference, 6-month backward-compat):" sub-block MUST be retained underneath with the original 4-5 EARS examples preserved verbatim

**Verification**: `grep -B1 -A8 'GEARS Format' .claude/skills/moai-foundation-core/references/examples.md` shows the block + at least one generalized subject + at least one `Where` example.

### AC-FCG-003 — references/reference.md corrects EARS expansion and labels notation

**Given**: A downstream consumer reads `.claude/skills/moai-foundation-core/references/reference.md`
**When**: The consumer reaches line 20 (current text "Format: EARS (Event-Action-Response-State) specifications")
**Then**:
- Line 20 MUST read "Format: GEARS (Generalized EARS) — primary notation; EARS retained as legacy reference for the 6-month backward-compatibility window"
- Wherever EARS is referenced as legacy, the expansion MUST be corrected to "Easy Approach to Requirements Syntax" (not the incorrect "Event-Action-Response-State")
- Lines 198, 377, 410, 447 EARS references MUST carry a "(legacy)" label OR be reframed to point at GEARS as primary

**Verification**: `grep -n 'EARS' .claude/skills/moai-foundation-core/references/reference.md` shows no occurrence of "Event-Action-Response-State"; `grep -c 'GEARS' references/reference.md` ≥1.

### AC-FCG-004 — No new IF/THEN modality introduced in edited content

**Given**: The run-phase completes M2-M5 file edits
**When**: Orchestrator runs `grep -rn 'IF.*THEN\|IF .* THEN' .claude/skills/moai-foundation-core/ | grep -v spec-ears-format.md | grep -v "_legacy"`
**Then**: Output MUST be EMPTY (zero matches outside the protected legacy reference file or sub-blocks explicitly tagged "(legacy reference)")
- Pre-existing `IF/THEN` text on protected files (`modules/spec-ears-format.md`) is acceptable per EXC-FCG-008
- Newly authored content in this SPEC's edit set MUST use `When <event-detected>` GEARS form per REQ-FCG-007 / REQ-FCG-012

**Verification**: `grep -rn 'IF.*THEN' .claude/skills/moai-foundation-core/ | grep -v 'spec-ears-format.md'` returns empty.

### AC-FCG-005 — Template mirror parity verified zero-diff

**Given**: The run-phase completes all edits to `.claude/skills/moai-foundation-core/`
**When**: Orchestrator runs `diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/`
**Then**: Output MUST be EMPTY (zero diff, full mirror parity per REQ-FCG-010)
- All ≤8 local file edits MUST have corresponding template mirror updates
- No file count difference; no content drift

**Verification**: `diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/ ; echo "exit=$?"` returns `exit=0`.

### AC-FCG-006 — `make build` regenerates embedded.go cleanly

**Given**: Template mirror updates are committed
**When**: Orchestrator runs `make build`
**Then**:
- `make build` MUST succeed (exit code 0)
- `git diff internal/template/embedded.go` MUST show non-empty diff (regenerated content reflects template edits)
- `go test ./internal/template/...` MUST pass (no regression in template loading)

**Verification**: `make build && go test ./internal/template/...` exits 0; `git status internal/template/embedded.go` shows modified.

### AC-FCG-007 — modules/spec-ears-format.md body preserved verbatim

**Given**: The run-phase touches sibling files referencing `modules/spec-ears-format.md`
**When**: Orchestrator runs `git diff HEAD~6..HEAD -- .claude/skills/moai-foundation-core/modules/spec-ears-format.md` after M6 completes
**Then**: Output MUST be EMPTY (the file body is NOT modified per EXC-FCG-008 + REQ-FCG-005 + REQ-FCG-011)
- The v3.0.0 DEPRECATED banner (current lines 9-15) MUST remain verbatim
- File content MUST be byte-identical pre-vs-post run-phase
- Template mirror of this file MUST also remain unchanged

**Verification**: `diff .claude/skills/moai-foundation-core/modules/spec-ears-format.md internal/template/templates/.claude/skills/moai-foundation-core/modules/spec-ears-format.md` returns empty; `git log --oneline -- .claude/skills/moai-foundation-core/modules/spec-ears-format.md` shows no new commits attributed to this SPEC.

### AC-FCG-008 — Predecessor SPEC bodies not modified

**Given**: The run-phase commits land
**When**: Orchestrator runs `git log --oneline -- .moai/specs/SPEC-V3R6-{GEARS-MIGRATION-001,SKILL-GEARS-ALIGN-001,PLAN-AUDITOR-GEARS-ALIGN-001}/`
**Then**: No new commits from this SPEC's run-phase modify any spec.md / plan.md / acceptance.md / progress.md body content in the 3 predecessor SPEC directories
- Only frontmatter `status` field of THIS SPEC's spec.md is allowed to transition `draft → in-progress` (Status Transition Ownership Matrix — manager-develop owns)
- Predecessor SPEC SSOT preservation per EXC-FCG-009

**Verification**: `git diff HEAD~6..HEAD -- .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/ .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/ .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/` returns empty.

### AC-FCG-009 — Self-lint zero new LegacyEARSKeyword warnings

**Given**: M6 completes and all edits land
**When**: Orchestrator runs `moai spec lint .moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md` AND `moai spec lint --strict` against the full repo
**Then**:
- This SPEC's own spec.md MUST emit zero `LegacyEARSKeyword` warnings (12/12 REQs in pure GEARS notation per REQ-FCG-007 / REQ-FCG-012 self-dogfood)
- The `--strict` repo-wide lint MUST NOT regress (any pre-existing warnings on legacy SPECs or `modules/spec-ears-format.md` body are acceptable per EXC-FCG-008; warning count delta ≤ 0 introduced by this SPEC's edits)

**Verification**: `moai spec lint .moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md` exits 0 with zero `LegacyEARSKeyword` findings; pre-vs-post `moai spec lint --strict` warning count delta = 0.

## §D Edge Cases

### Edge case E1 — Discovered file count exceeds preliminary estimate

If M1 discovery reveals additional files needing edits beyond the 8-file preliminary edit set, the run-phase MUST:
1. Update plan.md §C.1 with the corrected count (L46 attribution discipline)
2. Re-confirm Phase 0.5 plan-auditor (re-spawn if scope materially expanded)
3. NOT silently absorb additional files into M2-M5 without spec.md REQ coverage

### Edge case E2 — Pre-existing GEARS reference in foundation files

If M1 discovery reveals files already partially GEARS-aligned (beyond `modules/spec-ears-format.md`), the run-phase MUST:
1. Document the pre-existing alignment in M1 verification
2. Verify those files do NOT need re-editing (avoid double-modification)
3. Confirm `grep -c 'GEARS'` count baseline pre-vs-post matches expectation

### Edge case E3 — Template mirror has pre-existing drift

If M1 baseline `diff -r` reveals pre-existing drift between local and template (NOT introduced by this SPEC), the run-phase MUST:
1. Surface the drift in plan.md §C.1 + M1 verification log
2. Decide: include drift fix in this SPEC's scope OR escalate to orchestrator for separate SPEC
3. Default decision: include in this SPEC if drift is in the touched files; escalate if in untouched files

### Edge case E4 — `make build` fails due to upstream Go module issue

If M5 `make build` fails for non-content reasons (e.g., Go module resolution failure unrelated to this SPEC), the run-phase MUST:
1. Document the upstream issue
2. Verify template files are byte-correct (mirror parity zero-diff)
3. Escalate `make build` failure to orchestrator before declaring M5 complete

## §E Quality Gate Criteria

| Gate | Threshold | Verification |
|------|-----------|--------------|
| Plan-auditor score | ≥0.85 (Tier M PASS) | Phase 0.5 auditor signal |
| Plan-auditor skip-eligibility | ≥0.90 (ideal for 1-pass) | Phase 0.5 auditor signal |
| Template mirror parity | zero diff | AC-FCG-005 |
| `make build` regeneration | exit 0 + non-empty diff | AC-FCG-006 |
| IF/THEN re-introduction | zero in edited content | AC-FCG-004 |
| GEARS notation surface count (SKILL.md) | ≥3 | AC-FCG-001 |
| Self-lint LegacyEARSKeyword delta | 0 (no new warnings) | AC-FCG-009 |
| Predecessor SPEC bodies modified | 0 commits | AC-FCG-008 |
| `modules/spec-ears-format.md` body modified | 0 commits | AC-FCG-007 |

## §F Out-of-scope Acceptance Boundaries

The following are explicitly NOT acceptance criteria for this SPEC:

- ❌ `internal/spec/lint.go` behavior verification (covered by SPEC-V3R6-GEARS-MIGRATION-001 AC suite)
- ❌ 4-locale docs-site rendering verification (covered by SPEC-V3R6-GEARS-MIGRATION-001 docs-site ACs)
- ❌ 88 legacy SPEC migration (deferred to SPEC-V3R6-GEARS-SWEEP-001)
- ❌ Downstream Sprint 10 cohort SPEC verification (each owned by separate SPEC)
- ❌ `moai-workflow-spec` SKILL bundle verification (covered by SPEC-V3R6-SKILL-GEARS-ALIGN-001 AC suite)
- ❌ `plan-auditor.md` agent body verification (covered by SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 AC suite)

## §G Traceability Matrix

| AC ID | Covers REQ IDs | Verification Type |
|-------|----------------|-------------------|
| AC-FCG-001 | REQ-FCG-001, REQ-FCG-002 | grep + manual read |
| AC-FCG-002 | REQ-FCG-004, REQ-FCG-006 | grep + manual read |
| AC-FCG-003 | REQ-FCG-003, REQ-FCG-008 | grep |
| AC-FCG-004 | REQ-FCG-007, REQ-FCG-012 | grep (zero-match assertion) |
| AC-FCG-005 | REQ-FCG-010 | `diff -r` zero-diff |
| AC-FCG-006 | REQ-FCG-010 (downstream effect) | `make build` + `go test` |
| AC-FCG-007 | REQ-FCG-005, REQ-FCG-011 | `git diff` + `diff` zero-output |
| AC-FCG-008 | EXC-FCG-009 (out-of-scope assertion) | `git diff` zero-output |
| AC-FCG-009 | REQ-FCG-007, REQ-FCG-009, REQ-FCG-012 | `moai spec lint` zero-delta |

All 12 REQs covered by at least one AC; all 10 EXCs honored via AC negative assertions or explicit out-of-scope boundary (§F).

## §H Cross-References

- spec.md REQ definitions: `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md` §3
- plan.md milestone decomposition: `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/plan.md` §D
- Predecessor AC patterns: `.moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/acceptance.md` (13 ACs, all PASS), `.moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/acceptance.md` (8 ACs, all PASS)
- Canonical GEARS notation: `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format"
- Lint behavior reference: `internal/spec/lint.go` `LegacyEARSKeyword` rule
