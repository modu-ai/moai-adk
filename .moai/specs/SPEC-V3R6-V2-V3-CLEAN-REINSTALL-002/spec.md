---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002
title: "Repair v2→v3 clean-reinstall regression (#1084 silent data loss loop, #1086 arbitrary-directory pollution)"
version: "0.1.0"
status: draft
created: 2026-07-16
updated: 2026-07-16
author: manager-spec
priority: P1
phase: "v3.0.0-rc-stabilization"
module: "internal/cli"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, regression-repair, fingerprint-convergence, idempotency"
tier: M
depends_on: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001]
related_specs: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001, SPEC-V3R3-UPDATE-CLEANUP-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001]
---

# SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002

## §A Motivation

### §A.1 Regression Context

`SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` (status: completed; the "parent SPEC") implemented a version-aware clean-reinstall path that detects v2 projects via a 3-signal fingerprint heuristic (Signal 1 = `system.yaml moai.version`; Signal 2 = `.agency/` legacy dir; Signal 3 = any of 43 DeprecatedPaths entries) and runs a 7-step backup-remove-reinstall cycle (REQ-VVCR-004). The parent's design intent is correct: v2 projects need a one-time migration to the v3 namespace layout, and user-asset preservation (PRESERVE inventory + SHA-256 hash diff per REQ-VVCR-005..008) MUST be honored throughout.

The implementation regressed in three ways that share the clean-reinstall code path:

**#1084 (High — silent data loss + infinite loop).** `moai update` re-triggers the clean-reinstall on every invocation, overwriting user customizations in `.moai/config/sections/language.yaml` (observed: ko→en) and `design.yaml` with template defaults. The fingerprint never converges to `IsV2: false` after the first clean-reinstall, so REQ-VVCR-027 (idempotency on v3 projects) is violated. The "Removed 10 deprecated paths" log is phantom — the actual filesystem diff is 0 paths removed on the second run, but the log still reports 10. Issue body explicitly cites the parent SPEC's `DeprecatedPaths 9→43 entries` expansion (§A.4: 9 pre-existing + 31 Category B v.2.x-era + 3 Category C rc1-stage = 43 total) as the regression trigger.

**#1086 (High — arbitrary-directory pollution).** `moai update` invoked in a directory that is NOT a moai project (e.g., `/tmp/some-random-dir`) falsely signals `signals: version=true` and installs the full template tree, creating `.moai/` and `.claude/` directories where none belonged. The parent's Option α broader-detection policy treats "system.yaml missing" as a POSITIVE Signal 1 (intended for v2 projects predating system.yaml), but this predicate also fires in arbitrary non-project directories — the fingerprint lacks a "is this actually a moai project?" precondition.

**Shared root cause.** Both issues trace to the same `runUpdate → detectV2Fingerprint → runCleanReinstall` path implemented by the parent SPEC. #1084 is the loop-control regression; #1086 is the predicate-precondition regression. The parent SPEC's user-asset preservation contract (HARD-1..HARD-6) is NOT the cause and MUST NOT be weakened to fix the loop.

### §A.2 User Story

As a moai user who customized `language.yaml` (conversation_language: ko) on a v3 project, when I run `moai update` to pull template refreshes, my customization MUST survive byte-identical across consecutive runs; the update MUST NOT silently reset it to the en template default. As a developer who runs `cd /tmp/some-dir && moai update` by accident (or in a fresh clone), the command MUST refuse rather than pollute the cwd with a full template tree.

### §A.3 Non-Goals

This SPEC is a **regression repair**, NOT a redesign of the parent's design intent. The parent's 7-step canonical order (REQ-VVCR-004), PRESERVE inventory semantics (REQ-VVCR-005..008), SHA-256 hash-diff detection of user-modified configs (REQ-VVCR-007), and backup-before-removal guarantee (REQ-VVCR-023/024) all remain authoritative. This SPEC tightens the fingerprint predicate, gates the REMOVE-phase log on actual removals, and decouples `.agency/` migration from clean-reinstall activation. It does NOT alter the PRESERVE inventory, does NOT weaken namespace protection, does NOT add new DeprecatedPaths entries, and does NOT redesign the backup directory scheme.

## §B Requirements

> **Notation:** GEARS (current canonical). `When` event-driven, `While` state-driven, `Where` capability-gate, `shall not` for unwanted behavior. REQ token prefix `REQ-CRR-NNN` (Clean-Reinstall-Repair) — distinct from the parent's `REQ-VVCR-NNN` to avoid collision with the parent's 35 REQ tokens.

### §B.1 Fingerprint Convergence (#1084 loop)

**REQ-CRR-001 (V3-version negative-override).** **When** Signal 1 reads a `moai.version` value starting with `v3.` (i.e., the `v3.*` prefix) from `.moai/config/sections/system.yaml`, the v2 fingerprint heuristic (**defined by parent REQ-VVCR-001**) **shall** return `IsV2: false` regardless of Signal 2 or Signal 3 state. This is a NEGATIVE-OVERRIDE that operationalizes parent REQ-VVCR-027 (idempotency on v3 projects): a v3 version field is proof of prior successful migration; any lingering deprecated paths on a v3 project are user-retained legacy (PRESERVE inventory), not v2-zombie evidence.

**REQ-CRR-002 (Loop termination invariant).** **When** a v3 project (per REQ-CRR-001) invokes `moai update`, the clean-reinstall code path (**defined by parent REQ-VVCR-002**) **shall not** activate. The update **shall** route to the existing file-level sync code path (the `IsV2: false` branch in `internal/cli/update.go` `runUpdate`), leaving the project's user-modified config sections byte-identical.

**REQ-CRR-003 (User-modified config preservation during file-level sync).** **While** the file-level sync code path (REQ-CRR-002) encounters user-modified `.moai/config/sections/*.yaml` files (detected via SHA-256 hash diff against the v3 template baseline per parent REQ-VVCR-007), the file-level sync **shall** preserve the user's modifications byte-identical and **shall not** overwrite with template defaults. This is the direct repair for the #1084 `language.yaml ko→en reset` symptom.

### §B.2 Non-Project Directory Rejection (#1086 pollution)

**REQ-CRR-004 (Positive project-marker precondition).** **Where** the v2 fingerprint heuristic is invoked, the heuristic **shall** require a positive moai-project marker — specifically the presence of `.moai/config/sections/system.yaml` as a regular file — BEFORE any Signal 1/2/3 evaluation. **When** the project marker is absent, the heuristic **shall** return `IsV2: false` without evaluating Signals, and `moai update` **shall** exit with a structured `not a moai project` error directing the user to `moai init`.

**REQ-CRR-005 (Non-project cwd rejection).** **When** `moai update` is invoked in a directory that fails the positive project-marker precondition (REQ-CRR-004), the update **shall not** install the template tree, **shall not** create `.moai/` or `.claude/` directories, and **shall not** invoke `runCleanReinstall`. The structured error message **shall** name the missing marker file and link the canonical remedy (`moai init`).

### §B.3 Phantom-Log Elimination

**REQ-CRR-006 (Actual-removal-count log gating).** **When** the REMOVE phase of `runCleanReinstall` completes, the log message reporting deprecated-path removal **shall** report the actual count of paths removed from the filesystem — computed as (paths existing pre-REMOVE) minus (paths existing post-REMOVE) — NOT the planned-list length (43 entries per parent §A.4). **When** the actual removed count is zero, the REMOVE phase **shall** omit the `Removed N deprecated paths` message entirely (a `no deprecated paths found to remove` informational line is permitted in its place).

### §B.4 `.agency/` Migration Decoupling

**REQ-CRR-007 (Independent `.agency/` migration trigger).** **Where** the `.agency/` legacy directory is present on a v3 project (per REQ-CRR-001), the `runMigrateAgency` invocation (**defined by parent REQ-VVCR-025**) **shall** fire independently of the v2 fingerprint verdict, gated solely on `.agency/` presence. **When** a v3 project still has `.agency/` directory present, `moai update` **shall** invoke `runMigrateAgency` to migrate `.agency/` content into `.moai/` WITHOUT activating the full clean-reinstall code path. This decoupling preserves the user-asset migration contract while breaking the loop.

### §B.5 Reproduction-First Acceptance (test shape — run-phase authored)

**REQ-CRR-008 (Reproduction test: fingerprint non-convergence).** The run-phase test suite **shall** include a failing-reproduction test (authored during run-phase per CLAUDE.md §7 Rule 4) that demonstrates the #1084 regression on the pre-fix implementation: a fixture v3 project (`system.yaml` with `moai.version: v3.0.0-rc2` + user-modified `language.yaml` with `conversation_language: ko`), `moai update` invoked twice in succession; the test asserts that the second invocation does NOT trigger the clean-reinstall code path (no backup directory created at `.moai/backups/v2-to-v3-*-{stamp}/`, no REMOVE-phase log emitted, `language.yaml` byte-identical across both runs). The test MUST fail on the pre-fix implementation and pass on the post-fix implementation.

**REQ-CRR-009 (Reproduction test: non-project directory pollution).** The run-phase test suite **shall** include a failing-reproduction test that demonstrates the #1086 regression on the pre-fix implementation: `moai update` invoked in a fixture directory with NO `.moai/` directory; the test asserts that no `.moai/` or `.claude/` directory is created, no template files are written, and a structured `not a moai project` error is emitted. The test MUST fail on the pre-fix implementation and pass on the post-fix implementation.

### §B.6 Non-Weakening of Namespace Protection

**REQ-CRR-010 (User-asset preservation guarantee — non-weakening).** The implementation of REQ-CRR-001..009 **shall not** weaken the parent SPEC's user-asset preservation contract (parent §C HARD-1 through HARD-6). Specifically: PRESERVE inventory integrity, user-modified config preservation, namespace-protected path protection, and backup-before-removal semantics **shall** all remain intact. The fix tightens the fingerprint predicate and gates the REMOVE-phase log; it does NOT alter the PRESERVE inventory enumeration, the SHA-256 hash-diff detection, the backup directory creation, or the MERGE-back path restoration.

### §B.7 Idempotency Verification

**REQ-CRR-011 (Three-run idempotency on v3 project).** **When** the fix is applied, three consecutive `moai update` invocations on a fixture v3 project (with user-modified `language.yaml` containing `conversation_language: ko`) **shall** produce: (a) no changes to `language.yaml` (byte-identical `conversation_language: ko` across all three runs), (b) no backup directory creation on runs 2 and 3, and (c) no REMOVE-phase log emission on runs 2 and 3. The first run MAY legitimately perform file-level sync if template content drifted; runs 2 and 3 are no-ops on user assets.

## §C Constraints

### HARD-1 (Parent SPEC authority preserved)
The parent SPEC `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` (status: completed) remains the authoritative design document for the clean-reinstall workflow. This SPEC repairs its implementation; it does NOT supersede its design intent.

### HARD-2 (No PRESERVE inventory weakening)
The fix implemented under this SPEC MUST NOT alter the PRESERVE inventory enumeration (parent REQ-VVCR-005/006), the SHA-256 hash-diff detection of user-modified configs (parent REQ-VVCR-007/008), the backup directory creation contract (parent REQ-VVCR-009/010), or the MERGE-back path restoration (parent REQ-VVCR-013..016). The PRESERVE inventory is sacrosanct.

### HARD-3 (No new DeprecatedPaths entries)
The 43-entry DeprecatedPaths list (parent §A.4: 9 + 31 + 3) is FROZEN. The fix MUST NOT add, remove, or reorder entries. The convergence repair is achieved via the v3-version negative-override (REQ-CRR-001), not by pruning the list.

### HARD-4 (Reproduction-First binding)
CLAUDE.md §7 Rule 4 binds: the acceptance criteria REQ-CRR-008 and REQ-CRR-009 define the failing-reproduction test shapes NOW (plan-phase); the tests themselves are authored during run-phase M4 and MUST fail on the pre-fix implementation before any production fix is applied.

### HARD-5 (No code changes in plan-phase)
This SPEC is plan-phase ONLY. NO code changes under `internal/`, `pkg/`, `cmd/`, or template source. NO commits. The run-phase SPEC (separate delegation) performs the implementation.

### SHOULD-1 (Cross-platform parity)
The fingerprint predicate change (REQ-CRR-001/004) MUST behave identically on macOS, Linux, and Windows. The file-existence check for `.moai/config/sections/system.yaml` uses `os.Stat` semantics; cross-platform path resolution MUST NOT introduce platform-specific divergence.

## §D Acceptance Index

> Acceptance criteria are enumerated in `acceptance.md`. Bidirectional REQ↔AC traceability:

| AC ID | Severity | REQ IDs | Synopsis |
|-------|----------|---------|----------|
| AC-CRR-001 | MUST | REQ-CRR-001 | V3-version negative-override fires |
| AC-CRR-002 | MUST | REQ-CRR-002, REQ-CRR-008 | Loop terminates on v3 project (reproduction test) |
| AC-CRR-003 | MUST | REQ-CRR-003, REQ-CRR-010 | `language.yaml` ko survives consecutive runs |
| AC-CRR-004 | MUST | REQ-CRR-004 | Positive project-marker precondition |
| AC-CRR-005 | MUST | REQ-CRR-005, REQ-CRR-009 | Non-project directory rejection (reproduction test) |
| AC-CRR-006 | MUST | REQ-CRR-006 | Actual-removal-count log (phantom eliminated) |
| AC-CRR-007 | MUST | REQ-CRR-007 | Independent `.agency/` migration on v3 project |
| AC-CRR-008 | MUST | REQ-CRR-010 | Non-weakening of PRESERVE / namespace protection |
| AC-CRR-009 | MUST | REQ-CRR-011 | Three-run idempotency on v3 project |
| AC-CRR-010 | SHOULD | REQ-CRR-001..011 | Cross-platform parity (macOS/Linux/Windows) |

## §E Out of Scope

### Out of Scope — Linux/bash dotfile pollution (#1081)
GitHub issue **#1081** (Linux/bash dotfile pollution) is OUT of this Epic. The dotfile pollution traces to an upstream Claude Code dependency (the dotfile creation happens in the Claude Code runtime, not in moai-adk's `moai update` code path). #1081 is tracked on a separate upstream-dependency track; this SPEC does NOT attempt to mitigate it.

### Out of Scope — doctor false signals (#1087, #1088)
GitHub issues **#1087** (Harness 5-Layer false-FAIL) and **#1088** (Skills Allowlist false-positive) are doctor-command false-signal defects unrelated to the clean-reinstall regression. They are scheduled for **SPEC-2** of this Epic (the second SPEC). This SPEC's scope is limited to the clean-reinstall code path (`internal/cli/v2_detection.go`, `internal/cli/update_clean_install.go`, `internal/cli/update.go`); the doctor command surface (`internal/cli/doctor*.go`, `internal/harness/`) is OUT.

### Out of Scope — zone-registry.md template packaging (#1090)
GitHub issue **#1090** (zone-registry.md template packaging gap) is a template-distribution defect unrelated to the clean-reinstall regression. It is scheduled for **SPEC-3** of this Epic (the third SPEC). This SPEC does NOT touch template packaging (`internal/template/templates/.claude/rules/moai/core/zone-registry.md`, catalog regeneration, `moai update` template-deployment surface).

### Out of Scope — DeprecatedPaths list pruning
The fix MUST NOT prune the 43-entry DeprecatedPaths list (HARD-3). Proposals to `fix the loop by removing Category B/C entries that the v3 template recreates` are OUT — that path weakens the v2-zombie detection surface and is not the chosen fix strategy. The chosen fix is the v3-version negative-override (REQ-CRR-001) which subsumes the loop without pruning.

### Out of Scope — Redesigning the clean-reinstall workflow
The fix MUST NOT redesign the parent SPEC's 7-step canonical order (REQ-VVCR-004), backup directory scheme (`.moai/backups/v2-to-v3-{ISO-8601-UTC}/`), or MERGE-back restoration semantics. The regression is a predicate + log-correctness bug, not a workflow-design bug.

### Out of Scope — `amendment_of` declaration
This SPEC is NOT an in-place amendment of `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`. The parent remains `status: completed` and authoritative for its design intent. This is a SEQUEL SPEC (`-002`) that repairs the parent's implementation regression without altering the parent's body or status.

## §F Cross-References

- **Parent SPEC**: `.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/spec.md` (the completed SPEC whose implementation regressed)
- **Dependency chain**: `SPEC-V3R3-UPDATE-CLEANUP-001` (DeprecatedPaths source), `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` (namespace-protect primitives)
- **GitHub issues**: #1084 (silent data loss loop), #1086 (arbitrary-directory pollution), #1081 (OUT — Linux dotfile, separate track), #1087/#1088 (OUT — doctor false signals, SPEC-2), #1090 (OUT — zone-registry template packaging, SPEC-3)
- **CLAUDE.md §7 Rule 4**: Reproduction-First Bug Fixing (binds REQ-CRR-008/009)
- **`.claude/rules/moai/development/spec-frontmatter-schema.md`**: 12-canonical-field frontmatter SSOT
- **`.claude/rules/moai/development/sprint-round-naming.md`**: Epic/SPEC/Milestone taxonomy (this SPEC = ENTRY SPEC of the `v3.0.0-rc Stabilization` Epic)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-16 | manager-spec | Initial plan-phase draft. Tier M regression-repair covering #1084 + #1086 via 11 REQ-CRR-NNN requirements + 10 AC-CRR-NNN criteria. Parent SPEC authority preserved; PRESERVE inventory non-weakened; Reproduction-First test shapes defined for run-phase M4 authoring. |
