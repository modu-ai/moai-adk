---
id: SPEC-CLI-LINT-COMPLETED-001
title: "Add 'completed' to terminalStatusEnum in spec lint"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/spec"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "spec-lint, terminal-status, drift, chore"
depends_on: []
related_specs: [SPEC-CLI-SUBPKG-SPLIT-001]
---

## §A. Requirement

### §A.1 Background

`internal/spec/lint.go` declares `terminalStatusEnum` (line 959) as the set of lifecycle terminal states for which a mismatch with the git-implied status is considered normal — not a drift false positive. The map currently contains three entries: `superseded`, `archived`, `rejected`.

`StatusGitConsistencyRule.Check` (entry at line 972) consumes the map at line 984 via the early-return gate `if terminalStatusEnum[fm.Status] { return nil }`. This is the SOLE consumption of the map (verified by file scan — no other read sites).

The map is annotated at lines 957-958 with `@MX:NOTE` and `@MX:REASON` citing origin `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` and explicitly inviting future extension: *"extend this map only when adding future states"*.

The `completed` status (the V3R6 3-phase lifecycle final state per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`, reached via the sync commit) is **MISSING** from this map. After completion, no further active-work commits land in git history, so `getGitImpliedStatus` and the frontmatter `status` naturally diverge — and `StatusGitConsistencyRule` flags the SPEC as a drift false positive.

### §A.2 GEARS Requirement

**When** a SPEC frontmatter `status` field equals `completed`, the `StatusGitConsistencyRule` **shall not** emit a `StatusGitConsistency` finding.

**Rationale**: `completed` is a terminal lifecycle state — its git-implied status naturally diverges from active-work history once the 3-phase close (plan→run→sync) lands on the sync commit. The existing `terminalStatusEnum` gate at `internal/spec/lint.go:984` already encodes this terminal-state principle for `superseded` / `archived` / `rejected`; `completed` belongs in the same set.

### §A.3 Supporting Requirement (Behavior Preservation)

The `StatusGitConsistencyRule` **shall** continue to skip the drift check for the existing terminal states `superseded`, `archived`, and `rejected` with identical semantics to the pre-change behavior.

### §A.4 Anti-Requirement (Scope Guard)

The `StatusGitConsistencyRule` drift-detection algorithm (lines 988-1010, the `fm.Status != gitStatus` branch and downstream severity escalation) **shall not** be modified by this SPEC. Only the `terminalStatusEnum` map literal is in scope.

## §B. Constraints

- Single-line production-code change: add `"completed": true,` to `terminalStatusEnum` in `internal/spec/lint.go`.
- Single characterization test added to `internal/spec/lint_test.go`.
- No new exported symbols.
- No changes to `StatusEnum` validation — `completed` is already an accepted status value per the canonical 8-value enum.
- No changes to severity or `--strict` escalation semantics.

### Out of Scope — Non-completed terminal states

- No other terminal states (e.g., future lifecycle values) are added to `terminalStatusEnum` in this SPEC.
- The map remains a closed set of four entries after this change: `superseded`, `archived`, `rejected`, `completed`.

### Out of Scope — StatusGitConsistencyRule control flow

- The rule's drift detection algorithm (lines 988-1010) is untouched.
- The `getGitImpliedStatus` helper is untouched.
- The `fm.ID == "" || fm.Status == ""` early-return guard at line 976 is untouched.
- Severity escalation under `--strict` is untouched.

### Out of Scope — Other lint rules

- No other lint rule (e.g., `FrontmatterSchemaRule`, `OwnershipTransitionRule`, `DuplicateSPECIDRule`) is modified.
- No changes to `internal/spec/audit.go` era classification or drift detection.
- No changes to `internal/spec/era.go`.

## §C. Assumptions

- The V3R6 lifecycle defines `completed` as a terminal state (per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`) — confirmed by `lifecycle-sync-gate.md` § Era Definitions and by `internal/spec/CLAUDE.md` § Status enum (8 values).
- The `@MX:REASON` annotation at `lint.go:958` is an explicit invitation to extend the map for future terminal states — `completed` qualifies.
- The existing 3-entry behavior is the preservation target; no characterization tests for those 3 states need to be added unless they do not yet exist (manager-develop confirms at M1).

## §D. History

- **2026-07-07**: Plan-phase artifact creation by `manager-spec`. Tier S minimal SPEC; single milestone M1. Origin: Phase 1B follow-up to `SPEC-CLI-SUBPKG-SPLIT-001` (separate, independent SPEC — not a sub-milestone of the parent). Bug discovered during orchestrator pre-inspection of `internal/spec/lint.go`.
