---
id: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001
title: "moai spec drift — canonical close/sync commit convention alignment"
version: "0.1.1"
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
| 0.1.1 | 2026-06-03 | manager-spec | plan-audit iter-1 remediation (FAIL 0.62 → rescope). D1: restrict in-scope to canonically-closed sub-class only; heterogeneity of the 43 acknowledged. D2: replaced 3 unverified exemplars (DESIGN-001, V3R2-ORC-003, V3R3-CI-AUTONOMY-001 lack the close-infix) with 4 git-log-verified close-infix SPECs. D3: baseline corrected 66 → 67. D4: backfill-no-regression AC added. AC-DCA-001 rewritten to named-exemplar form. §B carve-out for legacy + genuine-incomplete sub-classes. |

---

## A. Background and Problem Statement

### A.1 Symptom

`moai spec drift` reports **67 DRIFT entries** (verified baseline at HEAD `e5574f1ab`; the count is 67 not 66 because the plan-phase commit of this very SPEC added one `draft` SPEC that itself drifts — D3 correction from iter-1). Of the 67, approximately **43 are `completed`-status false-positives**: SPECs whose frontmatter declares `status: completed`, yet the git-implied status resolved by the drift walker comes out as `implemented`, `in-progress`, or `planned` — so the tool flags a mismatch (DRIFT) that does not exist.

The canonical lifecycle audit tool `moai spec audit` reports these same SPECs as **CLEAN (0 MUST-FIX)**. The two tools disagree, and the SPEC state on disk is correct (frontmatter `completed`, audit clean). Therefore the **drift output is wrong**, not the SPEC state.

### A.1.1 The 43 are HETEROGENEOUS (iter-1 D1 finding)

The plan-audit iter-1 (FAIL 0.62) revealed that the 43 `completed`-status false-positives are **not a single class**. They split into three distinct sub-classes by the commit convention that closed them:

| Sub-class | Population (approx.) | Closing convention | Reachable by the canonically-closed fix? |
|-----------|----------------------|--------------------|------------------------------------------|
| **(1) Canonically-closed** | ~20 | Newest classifiable exact-token commit is the `4-phase close` / `Mx-phase audit-ready` infix (optionally with a `backfill` chore immediately newer) | **YES — in scope** |
| **(2) Legacy close conventions** | remainder | `sync(specs): ...→ completed`, `docs(spec): ...closure`, or other pre-canonical subjects that the close-infix fix does not recognize | **NO — deferred (§B Out of Scope)** |
| **(3) Genuine incomplete-close** | remainder | NO close commit at all — the newest exact-token commit is a `feat`/`fix` (e.g., `SPEC-DESIGN-001`, whose newest exact-token commit is a `feat`). These are REAL drift: the SPEC is marked `completed` in frontmatter but was never closed via a close commit. | **NO — must NOT be silently cleared (§B Out of Scope)** |

This SPEC (rescoped per user-approved Tier S minimalism) targets **only sub-class (1)**. Sub-classes (2) and (3) are explicitly carved out in §B.

### A.1.2 Verified canonically-closed exemplars (D2 fix)

iter-1 used 4 named exemplars, 3 of which (`SPEC-DESIGN-001`, `SPEC-V3R2-ORC-003`, `SPEC-V3R3-CI-AUTONOMY-001`) lack the close-infix and would STILL drift after the fix → AC-DCA-001 would fail at run-phase (D2 BLOCKING). They are replaced with 4 git-log-verified close-infix exemplars (full evidence in §3 AC-DCA-001). Concrete primary exemplar (verified this session, HEAD `e5574f1ab`): `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` was closed via the canonical `moai spec close` path — its newest exact-token commits are `a3dcc6b31 chore(...): backfill §E.2/§E.5 commit SHA` (a `backfill` chore) immediately over `2d64f3052 chore(...): Mx-phase audit-ready signal + 4-phase close`. Frontmatter is `completed`, `moai spec audit` is clean, yet `moai spec drift` reports `in-progress`. Post-fix the walker skips the `backfill` chore and classifies the close-infix as `completed`.

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

### Out of Scope — Legacy close conventions (sub-class 2)

`completed`-status SPECs whose closing commit uses a **legacy convention** the canonically-closed fix does not recognize — e.g., `sync(specs): ...→ completed`, `docs(spec): ...closure`, or other pre-canonical close subjects — are explicitly OUT of scope. The close-infix (`4-phase close` / `Mx-phase audit-ready`) fix in this SPEC will NOT resolve these; they will remain DRIFT after the fix and that is acceptable. They are deferred to a NAMED follow-up SPEC: **SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001** (to be authored separately; suggested Tier S). Attempting to fold legacy-convention recognition into this SPEC would breach Tier S minimalism and re-introduce the heterogeneity that caused iter-1 FAIL.

### Out of Scope — Genuine incomplete-close SPECs (sub-class 3)

`completed`-status SPECs that have **NO close commit at all** — the newest exact-token commit is a `feat`/`fix` (e.g., `SPEC-DESIGN-001`) — are **REAL drift**, not false-positives. The frontmatter claims `completed` but the SPEC was never closed via a close commit. The drift tool **correctly** flags these. This SPEC MUST NOT silently clear them by changing the classifier. The correct remediation is operational, NOT classifier-side: actually run `moai spec close` on each so a genuine close commit exists. Any fix that makes the drift walker report `completed` for a SPEC with no close commit would mask a real lifecycle gap and is an explicit anti-goal (see plan.md §G AP-2).

### Out of Scope — Reducing all 67 DRIFT entries to zero

Only sub-class (1) (the ~20 canonically-closed false-positives) is in scope. The baseline is 67 (D3-corrected). The remaining ~47 entries include sub-classes (2) and (3) above plus unrelated classification gaps. A strict-decrease (not zero) is the directional target — see §3 AC-DCA-001.

---

## C. Requirements (GEARS)

### C.1 Functional Requirements

- **REQ-DCA-001** (Event-driven): **When** the drift walker classifies a commit whose subject matches the canonical close convention `chore(SPEC-{ID}): ... 4-phase close` (or carries the `Mx-phase audit-ready signal` infix), the classifier shall resolve the commit to status `completed`.

- **REQ-DCA-002** (Ubiquitous): The classifier shall distinguish a SPEC-ID-scoped chore commit (`chore(SPEC-XXX-NNN): ...`, which carries lifecycle meaning) from the metadata-sweep chore commit (`chore(spec): ...` / `chore(specs): ...`, which carries no lifecycle meaning and must be skipped).

- **REQ-DCA-003** (Unwanted behavior): The classifier shall not let the over-broad generic `chore` prefix swallow a `chore(SPEC-...)` lifecycle commit and misclassify it as `in-progress`.

- **REQ-DCA-004** (State-driven): **While** the drift walker iterates commits newest-first, the walker shall return `completed` for a canonically-closed SPEC before reaching the earlier sync `docs(SPEC-{ID}): sync-phase` commit that would otherwise resolve to `in-progress`.

- **REQ-DCA-005** (Unwanted behavior): The classifier shall not resolve any SPEC to `completed` when that SPEC has **no** canonical close commit (no `4-phase close` / `Mx-phase audit-ready` infix in its exact-token git history). Genuine incomplete-close SPECs (sub-class 3, §B) shall continue to be reported as drift — the fix recognizes the close-infix as a positive signal only; it does NOT infer `completed` from sync/feat commits in the absence of a close commit.

### C.2 Regression-Preservation Requirements

- **REQ-DCA-006** (Unwanted behavior): The classifier shall not break the existing metadata-sweep skip behavior — `chore(spec):` and `chore(specs):` commits shall continue to be skipped (preserving SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 AC-LSCSK-003).

- **REQ-DCA-007** (Ubiquitous): The classifier shall preserve the LSGF-001 word-boundary SPEC-ID filter behavior unchanged (no substring-collision regression).

### C.3 Verification Requirements

- **REQ-DCA-008** (Event-driven): **When** the fix is implemented and `moai spec drift --count` is run against this repository, the reported drift count shall **strictly decrease** from the verified baseline of 67 (the ~20 canonically-closed false-positives of sub-class 1 resolve to aligned). The primary success signal is the named-exemplar transition (AC-DCA-001), not a raw count band.

- **REQ-DCA-009** (Ubiquitous): The `moai spec audit` result shall remain clean (no new MUST-FIX finding introduced by the drift classifier change).

- **REQ-DCA-010** (State-driven): **While** a SPEC's frontmatter is `implemented` (aligned, NOT `completed`) and its newest exact-token commit is a SPEC-ID-scoped `backfill` chore over a sync `docs` commit, the narrow backfill-skip shall not flip the SPEC's git-implied status to `completed` — the SPEC shall remain aligned (`implemented`) after the fix.

---

## 3. Acceptance Criteria (measurable, REQ↔AC bidirectional traceability)

> All AC are run-phase verifiable. This plan-phase SPEC defines them as the implementation oracle.

### AC-DCA-001 — Named verified-close-infix exemplars transition DRIFT → aligned (PRIMARY) — verifies REQ-DCA-001, REQ-DCA-004, REQ-DCA-008

This is the **headline AC**, in named-exemplar form (objectively verifiable, immune to count fluctuation). Each of the following 4 exemplars was git-log-verified at HEAD `e5574f1ab` to (a) carry frontmatter `status: completed`, (b) currently appear as DRIFT, and (c) have a newest classifiable exact-token commit that the canonically-closed fix resolves to `completed` (close-infix, optionally behind a `backfill` chore that the narrow backfill-skip steps over).

- **Given** the repository at the post-fix state
- **When** `moai spec drift --json` is executed
- **Then** ALL 4 named exemplars below MUST be absent from the DRIFT set (i.e., `Drifted == false`, git-implied status resolves to `completed`):

| # | Exemplar SPEC-ID | git-log evidence (newest exact-token commits, oldest→newest reading right-to-left) | Why the walker reaches `completed` post-fix |
|---|------------------|-----------------------------------------------------------------------------------|---------------------------------------------|
| 1 | `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` | `a3dcc6b31 chore(...): backfill §E.2/§E.5 commit SHA` → `2d64f3052 chore(...): Mx-phase audit-ready signal + 4-phase close` | newest is a `backfill` chore (skipped) → next is `4-phase close` infix (→completed) |
| 2 | `SPEC-V3R6-CI-FLAKY-STABILIZE-002` | `68d7ea612 chore(...): L60 mx_commit_sha backfill — 41868a664` → `41868a664 chore(...): Mx-phase audit-ready signal + 4-phase close` | newest is a `backfill` chore (skipped) → next is `4-phase close` infix (→completed) |
| 3 | `SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001` | `50f44fc9d chore(...): Mx-phase audit-ready signal + 4-phase close` (newest exact-token, no backfill above it) | newest IS the `4-phase close` infix directly (→completed) |
| 4 | `SPEC-V3R6-CI-FLAKY-STABILIZE-001` | `efcceffcc chore(...): L60 mx_commit_sha backfill — 7ab847b3a` → `7ab847b3a chore(...): Mx-phase audit-ready signal + 4-phase close` | newest is a `backfill` chore (skipped) → next is `4-phase close` infix (→completed) |

Backup exemplar (held in reserve, also verified to reach `completed`): `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001` (newest exact-token `e7b119924 ...HISTORY...backfill` + `7c8178972 ...mx_commit_sha atomic backfill` both skipped → `014b0344d ...Mx-phase audit-ready signal + 4-phase close marker` →completed). If any of the 4 primary exemplars is closed/re-touched before run-phase such that its newest exact-token commit changes, substitute this backup.

> Rejected-in-iter-2 exemplars (D2 evidence — do NOT use): `SPEC-V3R6-LIFECYCLE-SYNC-GATE-001` (newest exact-token is `38d07a6a0 fix(...): closer.go ...` — a `fix` NEWER than the close commit → walker returns `implemented`, never reaches `completed`) and `SPEC-V3R6-AGENT-TEAM-REBUILD-001` (newest exact-token is `40dc43f5b chore(...): template internal-content leak cleanup pass 2` — generic `chore`, not `backfill`, not close-infix → walker returns `in-progress`). Both would FAIL at run-phase if used as exemplars; this is exactly the D2 trap.

### AC-DCA-001b — Count strictly decreases (SECONDARY / directional) — verifies REQ-DCA-008

- **Given** the verified baseline `moai spec drift --count == 67` (at HEAD `e5574f1ab`)
- **When** `moai spec drift --count` is executed at the post-fix state
- **Then** the count is **strictly less than 67** (`< 67`). No exact band is asserted — the named-exemplar AC-DCA-001 above is the binding success signal. (Directional rationale: the ~20 canonically-closed sub-class-1 SPECs resolve to aligned; expected post-fix count ≈ 67 − [≈20] ≈ 47, but newly-landed SPECs between now and run-phase make any exact band brittle, so only strict decrease is asserted.)

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

### AC-DCA-008 — Backfill-no-regression: an aligned `implemented` SPEC stays aligned (D4) — verifies REQ-DCA-010, REQ-DCA-005

This guards against the narrow backfill-skip accidentally flipping an already-aligned `implemented` SPEC to `completed`. The fixture is **synthetic** (no real-world SPEC in this repo currently has the `implemented`+aligned+newest-backfill shape — all real backfill chores sit atop `completed` SPECs — so a unit-test fixture is the correct vehicle).

- **Given** a synthetic git-log fixture (newest-first): `chore(SPEC-Y-001): backfill §E.2 sync_commit_sha — abc1234` (SPEC-ID-scoped backfill chore, NO close-infix), then `docs(SPEC-Y-001): sync-phase artifacts`, then `feat(SPEC-Y-001): M1 ...`, with the SPEC's frontmatter `status: implemented`
- **When** the walker resolves the git-implied status (with the narrow backfill-skip active)
- **Then** the resolved git-implied status is `implemented` (NOT `completed`) — i.e., skipping the backfill chore exposes the sync `docs` commit (→`implemented`), and since there is no close-infix anywhere in the fixture, the walker MUST NOT infer `completed`. The SPEC remains aligned (frontmatter `implemented` == git-implied `implemented`, `Drifted == false`).

### REQ ↔ AC Traceability Matrix

| REQ | Covered by AC |
|-----|---------------|
| REQ-DCA-001 | AC-DCA-001, AC-DCA-002 |
| REQ-DCA-002 | AC-DCA-003 |
| REQ-DCA-003 | AC-DCA-002 |
| REQ-DCA-004 | AC-DCA-001, AC-DCA-004 |
| REQ-DCA-005 | AC-DCA-008 (no-completed-without-close-commit) |
| REQ-DCA-006 | AC-DCA-003, AC-DCA-005 |
| REQ-DCA-007 | AC-DCA-006 |
| REQ-DCA-008 | AC-DCA-001, AC-DCA-001b |
| REQ-DCA-009 | AC-DCA-006 |
| REQ-DCA-010 | AC-DCA-008 |

Every REQ maps to at least one AC; every AC traces back to at least one REQ. (AC-DCA-001 / AC-DCA-001b / AC-DCA-002..008 = 9 AC total across 10 REQ.)
