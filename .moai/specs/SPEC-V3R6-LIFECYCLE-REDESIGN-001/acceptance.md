---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
acceptance_version: "0.2.0"
spec_version: "0.2.0"
status: draft
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Acceptance Criteria — SPEC-V3R6-LIFECYCLE-REDESIGN-001

## §A. Coverage Matrix

| AC ID | Requirement | Severity | Milestone |
|-------|-------------|----------|-----------|
| AC-LR-001 | REQ-LR-001, REQ-LR-002 | MUST-PASS | M4, M5 |
| AC-LR-002 | REQ-LR-003, REQ-LR-004 | MUST-PASS | M5 |
| AC-LR-003 | REQ-LR-005, REQ-LR-006, REQ-LR-007 | MUST-PASS | M1, M2, M3 |
| AC-LR-004 | REQ-LR-008, REQ-LR-009 | MUST-PASS | M4 |
| AC-LR-005 | REQ-LR-010 | MUST-PASS | M4 |
| AC-LR-006 | REQ-LR-011 | SHOULD-PASS | M5 |
| AC-LR-007 | REQ-LR-012, REQ-LR-013, REQ-LR-014 | MUST-PASS | M6, M7 |
| AC-LR-008 | REQ-LR-015, REQ-LR-016 | SHOULD-PASS | M7, M8 |
| AC-LR-009 | REQ-LR-017 | MUST-PASS | M6 |
| AC-LR-010 | REQ-LR-018 | SHOULD-PASS | M5, M6 |
| AC-LR-011 | REQ-LR-019 | MUST-PASS | M2 |
| AC-LR-012 | REQ-LR-020, REQ-LR-021 | MUST-PASS | M2, M4 |
| AC-LR-013 | REQ-LR-005 (D5 scope) | MUST-PASS | M1, M4 |

## §B. Severity Definitions

- **MUST-PASS**: Blocks SPEC completion. Failure requires root-cause analysis and fix before sync phase.
- **SHOULD-PASS**: Strong expectation. Failure requires documented debt acceptance in sync phase.
- **NICE-TO-HAVE**: Best-effort. Failure is noted but does not block.

## §C. Acceptance Criteria (Given-When-Then)

### AC-LR-001: 3-Phase Lifecycle Is Canonical

**Given** the MoAI rule corpus in `.claude/rules/moai/`
**When** a reader searches for the canonical lifecycle definition
**Then** exactly three phases are documented: `plan`, `run`, `sync`
**And** MX Tag is documented as a cross-cutting concern (NOT a fourth phase)
**And** no rule file references "Mx-phase" as a distinct phase

**Verification**: `grep -r 'four.phase\|4-phase\|Mx-phase' .claude/rules/moai/ | grep -v 'Legacy\|deprecated\|retired\|alias'` returns 0 matches.

### AC-LR-002: progress.md Is 4-Section

**Given** a newly authored SPEC's `progress.md`
**When** the manager-spec emits the progress.md skeleton
**Then** the file contains exactly four §E sections: §E.1 (Plan Audit-Ready), §E.2 (Run Evidence), §E.3 (Run Audit-Ready), §E.4 (Sync Audit-Ready)
**And** no `§E.5 Mx-phase Audit-Ready Signal` section is present

**Verification**: Read the progress.md template in `.claude/skills/moai/workflows/plan/spec-assembly.md`; assert `§E.5` count is 0 and `§E.4` count is 1.

### AC-LR-003: H-4 Rewrite Preserves the V3R6 Count (INVARIANCE, not a literal — D3)

**Given** the V3R6 SPEC population, whose count `N` is captured at **run-phase M1 start** via `moai spec audit --json` (a moving baseline — illustratively N≈53 as of plan-phase 2026-06-19, but NOT a frozen literal: a parallel session is authoring more V3R6 SPECs)
**When** `internal/spec/era.go` `ClassifyEra` is rewritten per REQ-LR-005 (drop §E.5 + mx_commit_sha requirement)
**Then** the V3R6 count remains exactly `N` (the M1-captured baseline) — every SPEC classified V3R6 before the rewrite is still V3R6 after (via the dual-predicate migration window REQ-LR-006 and/or the H-5 fall-through)
**And** after the M3 backfill folds §E.5 into §E.4, the count is still `N` (via the NEW predicate or H-5)
**And** the genuine H-6 at-risk set (V3R6 SPECs lacking §E.4, lacking legacy §E.5+mx_sha, AND lacking modern phase/created≥2026-04-01) re-measured at M1 is empty OR is explicitly enumerated and handled by the migration window

**Verification** (assert INVARIANCE, never equality to a frozen number):
```bash
# At M1 start — capture baseline N (do NOT hardcode):
N=$(moai spec audit --json | python3 -c 'import json,sys; d=json.load(sys.stdin); print(sum(1 for f in d["drift_findings"] if f.get("era")=="V3R6"))')
# After M1 and after M3 — re-run the same command; assert each result == $N.
# The literal 53 may appear in research.md only as "(e.g. N≈53 as of plan-phase)", never as the pass condition.
```
Run the count before M1 (baseline `N`), after M1 (must equal `N`), after M3 (must equal `N`). The PASS condition is the three counts being EQUAL TO EACH OTHER, not equal to any literal.

### AC-LR-004: completed Transition Merges Into Sync Commit

**Given** the Status Transition Ownership Matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md`
**When** the matrix is updated per REQ-LR-008/009
**Then** the `implemented → completed` transition is owned by `manager-docs` via the sync commit
**And** no separate "Mx chore commit" transition is listed
**And** the `.claude/hooks/moai/status-transition-ownership.sh` hook enforces the merged transition

**Verification**: Read the Status Transition Ownership Matrix; assert exactly one `* → completed` row owned by `manager-docs` via sync commit. Assert `status-transition-ownership.sh` accepts the merged transition.

### AC-LR-005: 6 Drift Rules Use "3-phase close"

**Given** the 6 drift-surface rule files (verification-claim-integrity.md, spec-frontmatter-schema.md, agent-patterns.md, lifecycle-sync-gate.md, archived-agent-rejection.md, spec-workflow.md)
**When** the M4 sweep is complete
**Then** each file uses "3-phase close (plan→run→sync)" terminology
**And** none uses "4-phase close"

**Verification**: `grep -l '4-phase close' .claude/rules/moai/{core,development,workflow}/{verification-claim-integrity,spec-frontmatter-schema,agent-patterns,lifecycle-sync-gate,archived-agent-rejection,spec-workflow}.md` returns 0 matches.

### AC-LR-006: MX Tag Validation Occurs During Sync

**Given** a SPEC entering the sync phase
**When** manager-docs performs the sync-phase quality gate
**Then** MX Tag validation is part of the sync gate (REQ-LR-011)
**And** no separate "Mx-phase" step is invoked

**Verification**: Read `.claude/agents/moai/manager-docs.md` sync-phase section; assert MX Tag validation is listed as a sync sub-step, NOT a separate phase.

### AC-LR-007: Epic Taxonomy Replaces Legacy Vocabulary

**Given** the rewritten `.claude/rules/moai/development/sprint-round-naming.md`
**When** a reader consults the SSOT
**Then** exactly four canonical terms are defined: `Epic` (multi-SPEC grouping), `SPEC` (single work unit), `Milestone` (within-SPEC step), `Constitution` (project governance)
**And** `Sprint`, `cohort`, `Round`, `Wave` appear only in a "Legacy Aliases" migration section
**And** AP-SRN-001..004 are re-anchored to the Epic taxonomy

**Verification**: Read sprint-round-naming.md; count canonical-term definitions (must be ≤ 4); assert AP-SRN-001..004 reference Epic/SPEC/Milestone (not Sprint/Round).

### AC-LR-008: Naming Migration Proceeds in Tiers (assert 0 RESIDUAL matches — D6)

**Given** the naming-surface files identified in research.md (count is a moving baseline; illustratively ~102 across all tiers, of which T1 = **10** `.claude/rules/moai/` files as of the live `grep -rl` on 2026-06-19, dropping to **9** after M6 excludes sprint-round-naming.md itself)
**When** the M7 (T1) and M8 (T2-T4) milestones complete
**Then** the T1, T2, T3, T4 scope contains **0 residual** Sprint/cohort/Round/Wave matches (the binary pass condition is "0 residual", NOT a file-count assertion)
**And** T5 (archived/historical) is best-effort with documented debt

> **Count correction (D6)**: an earlier draft claimed T1 = "11" files; the live `grep -rl` returns **10** (9 after M6 excludes sprint-round-naming.md). Since the catalog is moving, this AC asserts **0 residual matches** post-migration (its grep verification already does this), NOT a fixed file count. (Note: research.md §C.1 Axis A drift surface = 14 files and §C.2 Axis B total = 102 files are verified EXACT at plan-phase and unchanged here.)

**Verification**: `grep -rl 'Sprint [0-9]\|코호트\|cohort\|Round [0-9]\|Wave [0-9]\|스프린트\|라운드' .claude/rules/moai/ .claude/agents/ .claude/output-styles/ .claude/skills/ .moai/docs/ .moai/project/` returns **0 matches** (excluding sprint-round-naming.md Legacy Aliases section and archived T5 paths).

### AC-LR-009: Epic Preserves Sprint Semantics

**Given** the Epic term definition in the rewritten sprint-round-naming.md
**When** a multi-SPEC grouping is referenced
**Then** Epic denotes a time-unit or thematic container for one or more SPECs
**And** the semantics are identical to the pre-redesign Sprint (only the label changes)

**Verification**: Read the Epic definition; assert it describes "multi-SPEC grouping by schedule, release, or thematic focus" (the former Sprint definition).

### AC-LR-010: New SPECs Reference Epic (Not Sprint)

**Given** a SPEC authored under the redesigned lifecycle
**When** the manager-spec emits the progress.md skeleton and plan.md
**Then** the Epic (not Sprint) the SPEC belongs to is referenced in the plan.md context or frontmatter
**And** the progress.md uses the 4-section §E layout

**Verification**: Read plan.md §A Context of a post-redesign SPEC; assert "Epic" appears (not "Sprint"); assert progress.md §E section count is 4.

### AC-LR-011: Era Classification Emits None of the Three §E.5-Based Drift Findings (D2 — incl. `Y_N_N_Y`)

**Given** a SPEC authored under the redesigned 3-phase lifecycle (4-section progress.md: §E.2 present, §E.5 ABSENT), with status=in-progress
**When** `moai spec audit --json` evaluates it
**Then** NONE of the three §E.5-keyed findings is emitted: not `Y_Y_Y_Y_StatusDrift` (audit.go:251), not `Y_Y_N_Y` (audit.go:268), and **critically not `Y_N_N_Y`** (audit.go:284, "§E.2 sync section present but §E.5 Mx section absent")
**And** in particular the `Y_N_N_Y` finding does NOT fire — without this fix, the mandated 4-section end-state (§E.5 absent) would trip `Y_N_N_Y` MUST-FIX on EVERY non-completed V3R6 SPEC, a catalog-wide false-positive storm
**And** the SPEC classifies as V3R6 via the new H-4 predicate (§E.2 + §E.4 + sync_commit_sha)
**And** a 4-section SPEC with sync complete (§E.4 + sync_sha) but status != completed DOES emit the re-anchored `SyncStatusDrift` finding (the one surviving drift dimension)

**Verification**: Author a test SPEC (or use an `internal/spec/audit_test.go` fixture) with the 4-section layout, §E.5 absent, status=in-progress; run `moai spec audit --json`; assert `finding_type` ∈ the result does NOT include `Y_N_N_Y`, `Y_Y_N_Y`, or `Y_Y_Y_Y_StatusDrift`. Plus a catalog-wide assertion: after M2+M3, `moai spec audit --json | grep -c '"finding_type": "Y_N_N_Y"'` returns 0.

### AC-LR-012: Close-Infix Matcher Accepts Both Infixes; Drift Detector Unaffected (D4)

**Given** the "4-phase close" → "3-phase close" rename in the close-subject mandate prose
**When** `internal/spec/transitions.go` `closeInfixMatch` is updated (M2)
**Then** `closeInfixMatch("...3-phase close...")` returns `true` (new infix recognized)
**And** `closeInfixMatch("...4-phase close...")` STILL returns `true` (legacy infix retained — historical close commits in git history carry "4-phase close")
**And** the M4 doc-axis edits to spec-frontmatter-schema.md and lifecycle-sync-gate.md carry a reconciliation note crediting SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 as the close-subject convention owner
**And** `go test ./internal/spec/...` (including `drift_combined_scope_test.go` and `transitions_test.go`) passes after the change

**Verification**:
```bash
go test ./internal/spec/... -run 'TestCloseInfix|TestClassifyPRTitle|TestDriftCombinedScope' -v   # all PASS
grep -c '3-phase close' internal/spec/transitions.go   # ≥ 1 (new infix const added)
grep -c '4-phase close' internal/spec/transitions.go   # ≥ 1 (legacy infix retained, NOT deleted)
grep -l 'DRIFT-LEGACY-CONVENTION-001' .claude/rules/moai/development/spec-frontmatter-schema.md .claude/rules/moai/workflow/lifecycle-sync-gate.md   # both present (reconciliation note)
```

### AC-LR-013: era.go Doc-Comment + lifecycle-sync-gate.md §E.5 Worked Example Updated (D5)

**Given** the era.go `ClassifyEra` doc-comment block (lines ~86-101) and the lifecycle-sync-gate.md `## §E.5 Mx-phase Audit-Ready Signal` worked example (line ~303)
**When** M1 (era.go) and M4 (lifecycle-sync-gate.md) complete
**Then** the era.go doc-comment heuristic table describes the NEW H-4 predicate (`§E.2 + §E.4 + sync_commit_sha`) + the legacy fallback — NOT the old `§E.2 + §E.5 + mx_commit_sha` text
**And** the lifecycle-sync-gate.md worked example reflects the 4-section layout (no `## §E.5 Mx-phase` section in the NEW-SPEC example; the H-4 heuristic-table row at line ~43 and era-definition row at line ~28 updated)

**Verification**:
```bash
sed -n '86,101p' internal/spec/era.go | grep -c '§E.4'        # ≥ 1 (doc-comment mentions new §E.4 predicate)
sed -n '86,101p' internal/spec/era.go | grep -i 'new H-4\|§E.2 + §E.4'   # present
grep -c '§E.5 Mx-phase Audit-Ready' .claude/rules/moai/workflow/lifecycle-sync-gate.md   # reduced (worked example updated to 4-section)
```

## §D. Edge Cases

- EC-1: A SPEC authored during the migration window (between M1 and M3) with the OLD 5-section layout — must still classify as V3R6 via the dual-predicate window (REQ-LR-006).
- EC-2: A SPEC authored after M3 with the NEW 4-section layout but missing `sync_commit_sha` — classifies as V3R5 (H-3), correctly indicating incomplete sync.
- EC-3: A grandfather-protected SPEC (V2.x/V3R2-R4/V3R5) with stale `§E.5` content — remains grandfather-protected; the backfill migration (M3) skips it.
- EC-4: A SPEC with explicit `era: V3R6` frontmatter (39 SPECs) — H-override fires regardless of progress.md layout; no reclassification risk.
- EC-5: A SPEC referenced by an Epic in plan.md but the Epic has only 1 member — valid (Epic may contain a single SPEC; the grouping concept is preserved).

## §E. Quality Gate Criteria

- **Test gate**: `go test ./internal/spec/...` passes with ≥ 85% coverage on `era.go`, `audit.go`, and `transitions.go`.
- **Lint gate**: `golangci-lint run` reports 0 errors in `internal/spec/`.
- **Audit gate (INVARIANCE, D3)**: capture baseline `N` (V3R6 count) at M1 start; assert the post-M3 V3R6 count == `N` (no regression). The pass condition is invariance, NOT `≥` a frozen literal (the count is moving; e.g. N≈53 as of plan-phase). Additionally `moai spec audit --json | grep -c '"finding_type": "Y_N_N_Y"'` returns 0 after M2/M3 (D2 — no catalog-wide §E.5-absence drift storm).
- **Grep gate**: Axis A and Axis B grep commands (per AC-LR-005, AC-LR-008) return 0 residual matches in scope.
- **Drift-detector gate (D4)**: `go test ./internal/spec/...` (incl. `drift_combined_scope_test.go`) passes; `closeInfixMatch` recognizes both "3-phase close" and "4-phase close" (AC-LR-012).
- **Template gate**: progress.md template in spec-assembly.md produces a 4-section file (AC-LR-002).

## §F. Definition of Done

- [ ] All MUST-PASS ACs (AC-LR-001..005, AC-LR-007, AC-LR-009, AC-LR-011, AC-LR-012, AC-LR-013) verified PASS with mechanical evidence.
- [ ] All SHOULD-PASS ACs (AC-LR-006, AC-LR-008, AC-LR-010) verified PASS or carry documented debt.
- [ ] `moai spec audit --json` reports 0 new drift findings introduced by this SPEC; the V3R6 count is invariant across M1→M3 (== the M1-captured baseline `N`, D3); `Y_N_N_Y` finding count is 0 catalog-wide (D2).
- [ ] `go test ./internal/spec/...` passes (incl. `era_test.go`, `audit_test.go`, `transitions_test.go`, `drift_combined_scope_test.go`); `closeInfixMatch` recognizes both "3-phase close" and "4-phase close" (D4).
- [ ] progress.md §E.4 records the sync audit-ready signal with `sync_commit_sha`.
- [ ] The SPEC's own frontmatter transitions: `draft → planned` (plan PR merge) → `in-progress` (M1 commit) → `implemented` (sync commit, carries `completed` transition).
- [ ] No separate Mx chore commit is emitted for this SPEC (the completed transition rides the sync commit per REQ-LR-008).

## §G. Forward-Looking Checks

- FL-1: The Epic taxonomy is adopted by the next-authored SPEC (verify via AC-LR-010 on the subsequent SPEC).
- FL-2: The docs-site sweep (deferred per EX-4) is tracked as a follow-up SPEC candidate.
- FL-3: Memory file renaming (deferred per EX-4) is tracked as a follow-up chore.
- FL-4: The `Constitution` term (REQ-LR-012) is referenced for SDD alignment but no new slash command is introduced (EX-6); a follow-up SPEC MAY introduce `/moai constitution` if desired.
