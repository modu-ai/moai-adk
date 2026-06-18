---
id: SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001
title: "Align SPEC-authoring guidance to the OutOfScopeRule lint convention"
version: "0.1.0"
status: implemented
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/agents + .claude/skills"
lifecycle: spec-anchored
tags: "guidance-drift, spec-lint, exclusions, out-of-scope, ssot-alignment"
tier: S
era: V3R6
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-18 | manager-spec | Initial plan-phase authoring (Tier S LEAN 2-artifact, ACs inline in §3) |

## A. Context (WHY)

A split-SSOT drift exists between the enforced spec-lint rule and the SPEC-authoring
guidance for the "Out of Scope" / exclusions section. The enforcement layer is correct
and stable; three authoring-guidance surfaces contradict it, so authors are instructed
to produce a heading shape that the lint rule rejects.

### A.1 Enforced SSOT (correct — NOT in scope to change)

- `internal/spec/lint.go` `OutOfScopeRule` (`Code() == "MissingExclusions"`, Error severity)
  requires three conditions in a `spec.md` body:
  1. The literal substring "out of scope" is present (case-insensitive).
  2. An `###` (H3 or deeper) heading whose text contains "out of scope".
  3. At least one `-` bullet item under that heading before the next `##` (H2) heading.
  When any condition fails it emits a `MissingExclusions` **Error**.
- `internal/spec/CLAUDE.md` Heading convention already codifies the canonical form:
  the "Out of Scope" section uses H3 (`###`) or H4 (`####`) sub-headings with the
  `### Out of Scope — <topic>` infix, each carrying at least one `-` bullet.

These two surfaces are the single source of truth and remain unchanged.

### A.2 Drifted authoring guidance (the fix target)

Three authoring surfaces contradict the enforced rule and cause recurring
`MissingExclusions` Errors:

1. `.claude/agents/moai/manager-spec.md` line 84 — instructs an H2 `## Exclusions (What NOT to Build)`
   heading with NO "out of scope" text. A spec.md authored verbatim per this line FAILS
   `OutOfScopeRule` (no H3 "out of scope" heading).
2. `.claude/agents/moai/plan-auditor.md` — SC-6 (line ~259), plus lines ~52 / ~91 / ~147
   reference an "Exclusions (What NOT to Build)" section rather than an "Out of Scope" H3.
3. `.claude/skills/moai-workflow-spec/SKILL.md` — line ~263 `### Exclusion Rules`
   and the verification checklist item line ~324 "Non-goals section present".

### A.3 Evidence of systemic drift

A live grep at plan-phase shows **347 of 353** existing `spec.md` files already contain
"out of scope" — but that compliance was achieved by per-SPEC manual patches, not by
correct guidance. Representative patch trails:

- `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001` CON-5 ("avoid the MissingExclusions lint ERROR
  that P0/P3 hit").
- `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` HISTORY D1 ("§E→Out of Scope + h3 sub-section
  (MissingExclusions ERROR)").
- `SPEC-V3R6-LIFECYCLE-REDESIGN-001` is the latest instance: its `## §J. Exclusions
  (What NOT to Build)` heading exactly matches manager-spec.md line 84 and failed lint.

Because manager-spec keeps emitting `## Exclusions` per its line 84, the Error recurs and
is patched per-SPEC at the symptom rather than at the guidance source. This SPEC fixes the
source so new SPECs satisfy the rule on first authoring.

## B. Scope (WHAT)

Align the three drifted authoring surfaces (§A.2) to the lint rule and `internal/spec/CLAUDE.md`
heading convention (§A.1). Each surface must instruct authors to include an exclusions
section that contains at least one `### Out of Scope — <topic>` H3 sub-heading with `-`
bullet items, so the section satisfies `OutOfScopeRule`. The user-facing "What NOT to Build"
intent is preserved; only the heading vocabulary/level changes.

### B.1 Exclusions (What NOT to Build)

This section preserves the canonical "What NOT to Build" intent, expressed in the target
convention so this SPEC does not reproduce the defect it fixes.

### Out of Scope — Enforcement-layer changes

- Modifying `internal/spec/lint.go` `OutOfScopeRule` logic, severity, or detection conditions.
  The rule is the correct SSOT and stays unchanged. Go change for the core fix (M1–M4) is
  exactly 0; the optional M5 CI guard is the only Go that may be touched (REQ-OSG-006, SHOULD).
  This aligns with plan.md §D ("Go change = 0 for the core fix … The optional CI guard (M5) is
  the only Go that may be touched, and only if elected.").
- Modifying `internal/spec/CLAUDE.md` heading convention. It already codifies the canonical
  form; it is a reference target, not an edit target.

### Out of Scope — Retroactive SPEC remediation

- Editing any of the 347 already-compliant `spec.md` files, or any of the 6 currently
  non-compliant ones (including `SPEC-V3R6-LIFECYCLE-REDESIGN-001`). This SPEC fixes the
  guidance source only; downstream SPEC bodies are out of scope and are remediated by their
  own owners under normal authoring.

### Out of Scope — Other guidance surfaces

- Any guidance file beyond the three named in §A.2. No broad sweep of every doc that
  mentions "Exclusions" is in scope — only the three surfaces that mechanically mislead
  SPEC authoring are fixed.

## C. Requirements (GEARS)

GEARS notation. Subject is generalized per the GEARS migration policy.

### REQ-OSG-001 (Ubiquitous — manager-spec alignment)

The `manager-spec` agent guidance **shall** instruct authors that every `spec.md` MUST
include an exclusions section containing at least one `### Out of Scope — <topic>` H3
sub-heading with one or more `-` bullet items, replacing the prior `## Exclusions (What NOT
to Build)` H2-only instruction.

### REQ-OSG-002 (Ubiquitous — plan-auditor alignment)

The `plan-auditor` agent guidance **shall** describe the exclusions check (SC-6 and the
related references) in terms of an "Out of Scope" H3 sub-heading with at least one bullet
item, so the audit criterion matches the `OutOfScopeRule` it gates.

### REQ-OSG-003 (Ubiquitous — workflow-spec skill alignment)

The `moai-workflow-spec` SKILL.md guidance **shall** present the exclusions/non-goals
guidance in terms of the `### Out of Scope — <topic>` convention so the skill's authoring
instruction and verification checklist match the lint rule.

### REQ-OSG-004 (Ubiquitous — intent preservation)

Each aligned surface **shall** preserve the user-facing "What NOT to Build" intent; the
change **shall** be limited to heading vocabulary and heading level, not the section's
purpose or the requirement that at least one exclusion entry exists.

### REQ-OSG-005 (Ubiquitous — Template-First mirroring)

For each aligned `.claude/...` guidance file, the corresponding
`internal/template/templates/.claude/...` mirror **shall** receive the identical alignment,
and the embedded template **shall** be regenerated via `make build`.

### REQ-OSG-006 (Where — capability gate, optional CI guard)

**Where** the run-phase elects to add a regression guard, a CI test **shall** assert that
the manager-spec guidance string matches the lint `OutOfScopeRule` convention (presence of
the `### Out of Scope` form), to prevent future re-drift. This requirement is a SHOULD
(see plan.md §F.3), not a blocker for closing this SPEC.

### REQ-OSG-007 (When — recursive-lint avoidance, self-applied)

**When** `moai spec lint` is run against this SPEC's own `spec.md`, the linter **shall**
report 0 errors — this SPEC's exclusions section (§B.1) uses the target `### Out of Scope —
<topic>` convention and therefore must not reproduce the very defect it fixes.

## D. Acceptance Criteria (inline — Tier S)

Given-When-Then scenarios. AC-OSG-00N maps 1:1 to REQ-OSG-00N unless noted.

### AC-OSG-001 — manager-spec instructs the Out of Scope H3 form

- **Given** `.claude/agents/moai/manager-spec.md` after the run-phase edit,
- **When** a reader inspects the exclusions instruction (formerly line 84),
- **Then** it directs authors to include at least one `### Out of Scope — <topic>` H3
  sub-heading with `-` bullet item(s), and the bare-H2 *mandate* bullet
  (`- [HARD] Every spec.md MUST include \`## Exclusions (What NOT to Build)\``) is no
  longer present as a mandate.
- **Negative evidence (mandate-scoped, NOT a substring grep)**:
  `grep -n '^- \[HARD\] Every spec.md MUST include .## Exclusions' .claude/agents/moai/manager-spec.md`
  returns empty. This deliberately scopes to the *bullet mandate* line. A naive substring grep
  for `## Exclusions (What NOT to Build)` is REJECTED here (D1): the old token sits inside
  backticks within a bullet (verified at L84 — it is NOT an actual `##` heading), so a substring
  grep would false-fail if the run-phase keeps the old token as a contrast example.
- **Positive evidence (authoring-guidance form, NOT the bare uppercase token)**:
  `grep -c '### Out of Scope' .claude/agents/moai/manager-spec.md` ≥ 1 — confirms the new
  `### Out of Scope — <topic>` H3 instruction. This targets the H3 authoring-guidance form
  specifically, NOT a case-insensitive `out of scope` count (D2): the pre-existing delegation-scope
  line `OUT OF SCOPE:` at manager-spec.md L77 (Scope Boundaries section, unrelated to
  spec.md-authoring guidance) already satisfies a case-insensitive count on its own — verified
  `grep -ic "out of scope"` returns 1 pre-edit from L77 alone — so a bare token count would
  false-PASS. The `### ` heading-prefix anchor disambiguates from the L77 line.
- **Run-phase warning (carried into plan.md §F M1)**: the implementer MUST NOT leave the literal
  bullet-mandate string `- [HARD] Every spec.md MUST include \`## Exclusions (What NOT to Build)\``
  as a contrast example, or the mandate-scoped negative grep above false-fails. If a contrast
  example is desired, phrase it without the verbatim bullet-mandate prefix.

### AC-OSG-002 — plan-auditor SC-6 references the Out of Scope H3 form

- **Given** `.claude/agents/moai/plan-auditor.md` after the run-phase edit,
- **When** a reader inspects SC-6 and the related lines (~52 / ~91 / ~147),
- **Then** the exclusions check is described in terms of an "Out of Scope" H3 sub-heading
  with at least one bullet, consistent with `OutOfScopeRule`.
- **Evidence**: `grep -n "Out of Scope" .claude/agents/moai/plan-auditor.md` shows the
  aligned SC-6 wording.

### AC-OSG-003 — workflow-spec skill presents the Out of Scope convention

- **Given** `.claude/skills/moai-workflow-spec/SKILL.md` after the run-phase edit,
- **When** a reader inspects the exclusion-rules guidance (~263) and the verification
  checklist (~324),
- **Then** both reference the `### Out of Scope — <topic>` convention rather than a
  generic "Exclusions"/"Non-goals" heading with no "out of scope" text.
- **Evidence**: `grep -n "Out of Scope" .claude/skills/moai-workflow-spec/SKILL.md` shows
  the aligned guidance and checklist item.

### AC-OSG-004 — intent preserved across all three surfaces

- **Given** all three aligned surfaces,
- **When** the diff is reviewed,
- **Then** each still requires at least one exclusion entry and still conveys "what NOT to
  build"; no requirement is removed and no surface drops the exclusions concept.
- **Mechanical floor (companion to the diff review, D4)**: each of the 3 aligned surfaces
  retains an exclusion-requirement / "what NOT to build" token after the edit, asserted by a
  `grep -c` ≥ 1 per surface (the human diff review remains the qualitative gate; this floor is
  the mechanical minimum):
  - `grep -ci "what NOT to build\|at least one .*exclusion\|exclusion.*at least one" .claude/agents/moai/manager-spec.md` ≥ 1
  - `grep -ci "what NOT to build\|at least one specific entry\|exclusion" .claude/agents/moai/plan-auditor.md` ≥ 1
  - `grep -ci "exclusion\|non-goal\|what NOT to build\|scope creep" .claude/skills/moai-workflow-spec/SKILL.md` ≥ 1
- **Qualitative evidence**: review diff shows heading-level/vocabulary changes only; the "at
  least one entry" requirement persists in each surface.

### AC-OSG-005 — Template-First mirrors aligned and embedded regenerated

- **Given** the three source edits,
- **When** the run-phase mirrors each edit to its
  `internal/template/templates/.claude/...` counterpart and runs `make build`,
- **Then** the three mirror files match their source counterparts on the aligned text, and
  `internal/template/embedded.go` is regenerated (working tree shows it changed or is clean
  if already current).
- **Evidence**: per-file `diff` of source vs mirror on the changed region shows parity;
  `make build` exits 0; `git status` reflects the regenerated `embedded.go`.

### AC-OSG-006 — optional CI guard (SHOULD)

- **Given** the run-phase elects to add the REQ-OSG-006 guard,
- **When** the new test runs (`go test ./internal/template/...` or the chosen package),
- **Then** it asserts the manager-spec guidance contains the `### Out of Scope` form and
  passes.
- **Note**: SHOULD, not a close blocker. If the guard is deferred, AC-OSG-006 is recorded
  as PASS-WITH-DEBT with a follow-up note; it does not block AC-OSG-001..005.

### AC-OSG-007 — this SPEC self-passes the lint it fixes

- **Given** this SPEC's own `spec.md`,
- **When** `moai spec lint .moai/specs/SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001/spec.md`
  is run at plan-phase,
- **Then** it reports 0 errors (the §B.1 exclusions section uses the target `### Out of
  Scope — <topic>` convention).
- **Evidence**: verbatim lint output captured in the plan-phase self-verification report.

## E. §E Section Skeleton (audit-ready signal placeholders)

The five `§E` markers below are emitted at plan-phase as placeholder headings only, for the
era-classification engine (`internal/spec/era.go` `hasAnyProgressMarker`) and audit
readability. The `§E.2`–`§E.5` sections are populated by manager-develop (run) and
manager-docs (sync/Mx) per the Forbidden-modifications matrix — NOT by manager-spec.

## §E.1 Plan-phase Audit-Ready Signal

- Tier S LEAN 2-artifact (spec.md + plan.md); ACs inline in §D.
- SPEC ID self-check: `decomposition: SPEC ✓ | V3R6 ✓ | OUTOFSCOPE ✓ | GUIDANCE ✓ | ALIGN ✓ | 001 ✓ → PASS`.
- Frontmatter 12 canonical fields + `tier: S` + `era: V3R6` validated pre-write.
- `moai spec lint` self-pass confirmed in the plan-phase report (AC-OSG-007).

## §E.2 Run-phase Evidence

- M1-M5 implemented in commit `77b568d7f`. Three guidance source files aligned to the `### Out of Scope — <topic>` H3 convention: `manager-spec.md` (bullet mandate at L84 + completeness refs L124/126/132/243), `plan-auditor.md` (SC-6 + refs L52/91/147), `moai-workflow-spec/SKILL.md` (Exclusion Rules + checklist). Three template mirrors aligned + `make build` regenerated `catalog.yaml` (3 in-scope skill hashes). M5 CI re-drift guard `internal/template/outofscope_guidance_align_test.go` added.

## §E.3 Run-phase Audit-Ready Signal

- AC-OSG-001..007 all PASS (manager-develop self-verification + orchestrator independent verify). `go build ./...` exit 0; `go test ./internal/template/ -run TestOutOfScopeGuidanceAligned` PASS. Mirror parity confirmed (`diff` IDENTICAL ×3). run_commit_sha: `77b568d7f`.

## §E.4 Sync-phase Audit-Ready Signal

- CHANGELOG `[Unreleased]` entry added (root-cause split-SSOT alignment summary). Status transition `in-progress → implemented` (orchestrator-direct, isolated worktree on origin/main during a multi-session parallel race). sync_commit_sha: `<backfill §E.4>`.

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_
