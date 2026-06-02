---
id: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001
title: "moai spec drift — canonical close/sync commit convention alignment"
version: "0.1.0"
status: draft
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/spec"
lifecycle: spec-anchored
tier: S
tags: "drift, lifecycle, commit-convention, false-positive, internal-spec"
---

# moai spec drift — canonical close/sync commit convention alignment

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring (Tier S). Root cause pre-verified by orchestrator. |

---

## A. Background and Problem Statement

### A.1 Symptom

`moai spec drift` reports **66 DRIFT entries**, of which approximately **43 are false-positives**. These false-positives are SPECs whose frontmatter declares `status: completed`, yet the git-implied status resolved by the drift walker comes out as `implemented`, `in-progress`, or `planned` — so the tool flags a mismatch (DRIFT) that does not exist.

The canonical lifecycle audit tool `moai spec audit` reports these same SPECs as **CLEAN (0 MUST-FIX)**. The two tools disagree, and the SPEC state on disk is correct (frontmatter `completed`, audit clean). Therefore the **drift output is wrong**, not the SPEC state.

Concrete evidence (verified this session, HEAD `a3dcc6b31`): `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` was just closed via the canonical `moai spec close` path. Its frontmatter is `status: completed` and `moai spec audit` is clean, yet `moai spec drift` reports its git-implied status as `in-progress`. The same symptom reproduces on `SPEC-DESIGN-001`, `SPEC-V3R2-ORC-003`, `SPEC-V3R3-CI-AUTONOMY-001`, and roughly 40 others.

### A.2 Root Cause (pre-verified by orchestrator — incorporated, not re-investigated)

The defect lives in two files in `internal/spec`:

- `internal/spec/drift.go` — `getGitImpliedStatus()` is a newest-first walker over `git log main --grep=<specID>`. It adopts the status of the **first** commit whose `ClassifyPRTitle()` returns a non-empty status. It applies a `shouldSkipCommitTitle()` filter that skips only the literal, colon-terminated prefixes `chore(spec):` and `chore(specs):`.
- `internal/spec/transitions.go` — `transitionRules` (prefix → status), evaluated by `ClassifyPRTitle()` via `strings.HasPrefix(lowerTitle, rule.prefix)`.

The relevant `transitionRules` entries and their failure contribution:

| Rule prefix | Mapped status | Failure contribution |
|-------------|---------------|----------------------|
| `chore(spec)` | `""` (skip-meta) | Intended to skip metadata sweeps. **Does NOT match** `chore(SPEC-V3R5-...)` because the rule has a literal `)` after `spec` while the SPEC-ID-scoped commit has `-` there. The over-broad concern is the **generic** `chore` rule below. |
| `docs(sync)` | `completed` | The **only** rule yielding `completed`. The canonical sync commit `docs(SPEC-{ID}): sync-phase artifacts` does NOT start with `docs(sync)`, so it never reaches this rule. |
| `docs` | `in-progress` | The canonical sync commit `docs(SPEC-{ID}): sync-phase artifacts` matches **here** → classified `in-progress`. |
| `chore` (generic) | `in-progress` | The canonical close commit `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close` AND the SHA-backfill commit `chore(SPEC-{ID}): backfill ...` both match **here** → classified `in-progress`. |

### A.3 The Convention Mismatch (the heart of the defect)

The drift classifier recognizes `completed` **only** via a `docs(sync):` prefix. But the canonical commit convention — the Status Transition Ownership Matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md` — uses:

- sync (`in-progress → implemented`): `docs(SPEC-{ID}): sync-phase artifacts` **or** `chore(SPEC-{ID}): sync-phase artifacts`
- close (`implemented → completed`): `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close`

Neither canonical subject matches `docs(sync):`. As a result, the drift walker can **never** resolve `completed` for any SPEC closed via the canonical `moai spec close` path. The newest commit (the close, which should mean `completed`) is instead swallowed by the generic `chore` rule and classified `in-progress`, producing a spurious DRIFT against the frontmatter `completed`.

### A.4 Walker order trace (worked example)

For `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`, `git log main --grep=<specID>` newest-first yields (abbreviated):

1. `chore(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): backfill §E.2/§E.5 commit SHA` → generic `chore` → `in-progress` → **walker returns here**.
2. `chore(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): Mx-phase audit-ready signal + 4-phase close` (never reached).
3. `docs(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): sync-phase artifacts (status→implemented, ...)` (never reached).
4. `feat(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): nested GitStrategyConfig ...` (never reached).

Frontmatter is `completed`; walker returns `in-progress` → spurious DRIFT.

This trace shows the fix must (a) classify the close/backfill commit as `completed` (or skip the SHA-backfill chore and classify the real close commit as `completed`), and (b) keep the metadata-sweep `chore(spec):`/`chore(specs):` skip intact.

---

## B. Out of Scope (What NOT to Build)

### Out of Scope — Implementation of the Go fix

This is a **plan-phase-only** SPEC. No Go code is written. The fix design lives in `plan.md`; the actual `internal/spec` edits and tests are deferred to the run phase.

### Out of Scope — Changing `moai spec audit` behavior

`moai spec audit` already reports these SPECs correctly (CLEAN). The audit engine, its era classification, and its grandfather clause are NOT touched. The fix targets only the drift classifier.

### Out of Scope — New commit-convention introduction

This SPEC aligns the drift classifier with the **existing** canonical commit convention (Status Transition Ownership Matrix). It does NOT introduce, rename, or deprecate any commit-subject pattern. The convention is the SSOT; the classifier is brought into agreement with it.

### Out of Scope — Word-boundary SPEC-ID filter changes

The LSGF-001 word-boundary filter (`commitMatchesSPECID` / `ExtractSPECIDs`) behavior is preserved unchanged. This SPEC does not modify substring-collision protection.

### Out of Scope — Reducing all 66 DRIFT entries to zero

Only the ~43 `completed`-status false-positives are in scope. The remaining ~23 entries may include genuine drift (e.g., true status lag) or unrelated classification gaps that are not addressed here.

---

## C. Requirements (GEARS)

### C.1 Functional Requirements

- **REQ-DCA-001** (Event-driven): **When** the drift walker classifies a commit whose subject matches the canonical close convention `chore(SPEC-{ID}): ... 4-phase close` (or carries the `Mx-phase audit-ready signal` infix), the classifier shall resolve the commit to status `completed`.

- **REQ-DCA-002** (Ubiquitous): The classifier shall distinguish a SPEC-ID-scoped chore commit (`chore(SPEC-XXX-NNN): ...`, which carries lifecycle meaning) from the metadata-sweep chore commit (`chore(spec): ...` / `chore(specs): ...`, which carries no lifecycle meaning and must be skipped).

- **REQ-DCA-003** (Unwanted behavior): The classifier shall not let the over-broad generic `chore` prefix swallow a `chore(SPEC-...)` lifecycle commit and misclassify it as `in-progress`.

- **REQ-DCA-004** (State-driven): **While** the drift walker iterates commits newest-first, the walker shall return `completed` for a canonically-closed SPEC before reaching the earlier sync `docs(SPEC-{ID}): sync-phase` commit that would otherwise resolve to `in-progress`.

- **REQ-DCA-005** (Capability gate): **Where** the run-phase analysis (plan.md §F) determines that classifying the close commit as `completed` alone is insufficient for robustness, the classifier shall additionally map the canonical sync commit `docs(SPEC-{ID}): sync-phase artifacts` (and its `chore(SPEC-{ID}): sync-phase artifacts` variant) to status `implemented`.

### C.2 Regression-Preservation Requirements

- **REQ-DCA-006** (Unwanted behavior): The classifier shall not break the existing metadata-sweep skip behavior — `chore(spec):` and `chore(specs):` commits shall continue to be skipped (preserving SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 AC-LSCSK-003).

- **REQ-DCA-007** (Ubiquitous): The classifier shall preserve the LSGF-001 word-boundary SPEC-ID filter behavior unchanged (no substring-collision regression).

### C.3 Verification Requirements

- **REQ-DCA-008** (Event-driven): **When** the fix is implemented and `moai spec drift --count` is run against this repository, the reported drift count shall drop from 66 toward approximately 23 (the ~43 `completed`-status false-positives resolve to aligned), within a tolerance band defined in §3 AC-DCA-001.

- **REQ-DCA-009** (Ubiquitous): The `moai spec audit` result shall remain clean (no new MUST-FIX finding introduced by the drift classifier change).

---

## 3. Acceptance Criteria (measurable, REQ↔AC bidirectional traceability)

> All AC are run-phase verifiable. This plan-phase SPEC defines them as the implementation oracle.

### AC-DCA-001 — Drift count drops into the target band — verifies REQ-DCA-001, REQ-DCA-004, REQ-DCA-008

- **Given** the repository at the post-fix state
- **When** `moai spec drift --count` is executed
- **Then** the reported count is **≤ 30 AND ≥ 18** (headline band: 66 → ~23; tolerance accounts for newly-landed SPECs and genuine residual drift)
- **And** at least the 4 named exemplars (`SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`, `SPEC-DESIGN-001`, `SPEC-V3R2-ORC-003`, `SPEC-V3R3-CI-AUTONOMY-001`) no longer appear as DRIFT in `moai spec drift --json` output.

### AC-DCA-002 — Canonical close commit classifies as `completed` — verifies REQ-DCA-001, REQ-DCA-003

- **Given** a commit subject `chore(SPEC-EXAMPLE-001): Mx-phase audit-ready signal + 4-phase close`
- **When** it is classified by the drift classifier (`ClassifyPRTitle` or the walker helper, per plan.md §F design)
- **Then** the resolved status is `completed` (NOT `in-progress`).

### AC-DCA-003 — SPEC-ID-scoped chore is disambiguated from metadata-sweep chore — verifies REQ-DCA-002, REQ-DCA-006

- **Given** two commit subjects: `chore(SPEC-EXAMPLE-001): backfill §E.2/§E.5 commit SHA` and `chore(spec): frontmatter status sweep`
- **When** both are classified
- **Then** the first is treated as a lifecycle-bearing SPEC-ID-scoped chore (not skipped, classified per the close/backfill rules) and the second is skipped (skip-meta, empty status), preserving AC-LSCSK-003.

### AC-DCA-004 — Newest-first walker returns `completed` before the sync `docs` commit — verifies REQ-DCA-004

- **Given** a synthetic git-log fixture (newest-first): `chore(SPEC-X-001): ... 4-phase close`, then `docs(SPEC-X-001): sync-phase artifacts`, then `feat(SPEC-X-001): ...`
- **When** the walker resolves the git-implied status
- **Then** it returns `completed` (from the close commit) and does not fall through to the `docs` → `in-progress` rule.

### AC-DCA-005 — Metadata-sweep skip regression guard remains green — verifies REQ-DCA-006

- **Given** the existing test `TestClassifyPRTitle_ChoreSpecUnchanged` (AC-LSCSK-003) and the chore-skip fixtures in `internal/spec`
- **When** `go test ./internal/spec/...` is run after the fix
- **Then** all chore-skip regression tests pass (no behavioral change to `chore(spec):`/`chore(specs):` skipping).

### AC-DCA-006 — Word-boundary filter unchanged + audit stays clean — verifies REQ-DCA-007, REQ-DCA-009

- **Given** the LSGF-001 word-boundary tests and `moai spec audit`
- **When** `go test ./internal/spec/...` and `moai spec audit` are run after the fix
- **Then** the word-boundary (LSGF-001) tests pass unchanged AND `moai spec audit` reports 0 new MUST-FIX findings.

### AC-DCA-007 — Full package test suite green — verifies all REQ (gate)

- **Given** the post-fix `internal/spec` package
- **When** `go test ./internal/spec/...` is run
- **Then** the suite is green with no skipped or failing tests attributable to this change, and coverage for the modified files does not regress below the package baseline (85%).

### REQ ↔ AC Traceability Matrix

| REQ | Covered by AC |
|-----|---------------|
| REQ-DCA-001 | AC-DCA-001, AC-DCA-002 |
| REQ-DCA-002 | AC-DCA-003 |
| REQ-DCA-003 | AC-DCA-002 |
| REQ-DCA-004 | AC-DCA-001, AC-DCA-004 |
| REQ-DCA-005 | AC-DCA-004 (robustness path; activated per plan.md §F decision) |
| REQ-DCA-006 | AC-DCA-003, AC-DCA-005 |
| REQ-DCA-007 | AC-DCA-006 |
| REQ-DCA-008 | AC-DCA-001 |
| REQ-DCA-009 | AC-DCA-006 |

Every REQ maps to at least one AC; every AC traces back to at least one REQ.
