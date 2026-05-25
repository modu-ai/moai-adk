---
id: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
title: "Lifecycle Sync Gate — Atomic 4-Phase Close + Cross-File Status Audit + Pre-Commit Drift Detection"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.1 follow-up to AGENT-RESPONSIBILITY-REALIGN + LIFECYCLE doctrine"
module: "internal/cli, internal/spec, .claude/hooks/moai, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "lifecycle, sync-gate, atomic-close, audit, drift-detection, hook, lint, status-transition, ownership-matrix, era-classification"
tier: L
depends_on: [SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001, SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001]
related_specs: [SPEC-V3R6-AGENT-TEAM-REBUILD-001, SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

## HISTORY

### v0.1.0 (2026-05-25, manager-spec)
- Initial plan-phase artifact set authored after 154-SPEC era × sync-signal cross-tab audit
- Sprint 11 cohort 1/N candidate, Tier L (>1000 LOC est. across 15 files)
- 5-artifact set: spec.md + plan.md + acceptance.md + design.md + research.md
- Motivation: 9 modern-standard violations detected, 4 fixed orchestrator-direct this session, 5 remaining surfaced as M6 dogfood targets
- L67 NEW pattern (manager-docs scope-creep producing Y|Y|Y|Y with spec.md status drift) catalyzed proactive guardrail design

---

## A. Context

### A.1 Problem Statement

The MoAI lifecycle close protocol (4-phase: plan → run → sync → Mx) currently relies on **5 sequential commits per SPEC close**:

1. sync-phase body commit (manager-docs writes progress.md §E.2/§E.3 + CHANGELOG)
2. sync_commit_sha atomic backfill chore (L60 chicken-and-egg)
3. Mx-phase body commit (orchestrator emits §E.5 audit-ready signal)
4. mx_commit_sha atomic backfill chore (L60 chicken-and-egg)
5. spec.md frontmatter status bump (implemented → completed)

This **5-commit cadence** is operationally fragile. The L67 pattern catalogued 2026-05-25 demonstrates that sync-phase agent spawns auto-absorb downstream Mx + 4-phase close concepts without explicit delegation, producing commits whose **commit message claims diverge from git state** (claimed counts/SHAs mismatch, partial §A backfills, missing HISTORY entries, frontmatter status drift between progress.md and spec.md).

The 154-SPEC retrospective audit conducted in this orchestrator turn (2026-05-25) classified all `status: implemented` SPECs across 4 era × sync-signal cross-tab dimensions:

```
80 N|N|N|N  — progress.md 자체 없음 (V2.x / V3R2-R4 era-final)
42 Y|N|N|N  — progress.md 있으나 sync 흔적 없음 (구 era 표준)
23 Y|N|N|Y  — sync section 있고 sync_commit_sha 없음 (V3R5 / 초기 V3R6)
 6 Y|Y|N|Y  — sync done, Mx 누락 (V3R6 모던 표준 위반) ← 본 SPEC 직접 대상
 3 Y|Y|Y|Y  — 4-phase 완전 종료지만 spec.md status drift (L67 매니저-docs scope-creep)
```

4 violations (the 3 Y|Y|Y|Y + LOCAL-NAMESPACE-CONSOLIDATION-001 status drift) were resolved this session via 4 atomic chore commits (`baaa1693e`, `d9ae06020`, `b8be7e44a`, `d74095e75`). **5 modern-era violations remain**:

- SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (sync `11abb9a30`, mx missing)
- SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 (sync `a853f2954`, mx missing)
- SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (sync `2d9871208b09e1ce647a4cc134b24267b713b42f`, mx=null literal)
- SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 (sync `009e68c5d`, mx missing)
- SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 (sync + mx both missing — broken state)

### A.2 Root Cause Analysis

The L60 atomic backfill design (chicken-and-egg: sync_commit_sha references its own future SHA) is **doubly susceptible to omission**:

1. **Complexity-driven omission**: 5 commits per close exceeds working memory; closers (orchestrator OR manager-docs spawn) skip steps 3-5 after step 2 success
2. **Canonical commit subject ambiguity**: `chore(SPEC-XXX): Mx-phase audit-ready signal + 4-phase close` does NOT enumerate the spec.md frontmatter bump obligation — it is implicit and frequently missed
3. **Scope-creep absorption (L67)**: When orchestrator delegates `/moai sync` to manager-docs, the spawn auto-absorbs Mx Step C + 4-phase close marker emission without explicit delegation, producing factually incorrect commit messages (claimed mx_commit_sha + claimed status: completed transitions that do not match actual git state)
4. **Cross-file invariant unguarded**: spec.md `status` field is the SPEC's lifecycle SSOT, but progress.md `§E.3 status` field is the run-phase witness. No mechanism enforces that these two values agree at any phase boundary.

### A.3 Proposed Architecture

5-deliverable single SPEC reduces 5 commits → 1 commit via atomic close + adds 3 defensive layers (audit, hook, lint):

1. **`moai spec close SPEC-XXX` CLI subcommand** — atomic single-commit 4-phase close. Reads progress.md §E.2/§E.3/§E.5 (sync + Mx body), validates 4-phase precondition matrix, writes spec.md `status: completed` + sync_commit_sha + mx_commit_sha + progress.md status in one staging cycle, emits single commit `chore(SPEC-XXX): 4-phase close — atomic` with deterministic message structure.

2. **`moai spec audit` CLI subcommand** — era-aware cross-file status drift detection. Iterates all `.moai/specs/SPEC-*/` directories, classifies each by era (V2.x / V3R2-R4 / V3R5 / V3R6) per heuristics (file presence + content patterns), grandfather-clauses 145 historical SPECs to "era-final" status, surfaces drift only for V3R6 modern-era SPECs. JSON output for CI integration.

3. **Pre-commit hook** (`handle-pre-commit-spec-status.sh`) — when a commit touches spec.md `status` field, verifies progress.md `§E.3 status` field matches; when commit message contains the canonical 4-phase close subject, enforces spec.md `status: completed`; on mismatch exit 2 with structured JSON, orchestrator translates to `AskUserQuestion` per agent-common-protocol §Hook Invocation Surface.

4. **spec-lint `OwnershipTransitionRule` extension** — detect direct `* → implemented` transitions in commit subjects authored by `manager-develop` (canonical owner of `in-progress → implemented` per Status Transition Ownership Matrix), emit `OwnershipTransitionInvalid` finding. Run-phase commits should leave SPEC at `status: in-progress`; only sync-phase (manager-docs) transitions to `implemented`.

5. **Rule file** (`.claude/rules/moai/workflow/lifecycle-sync-gate.md`) — protocol SSOT documenting era classification rules + grandfather clause policy + frontmatter optional `era:` field semantics + cross-reference to spec-frontmatter-schema.md Status Transition Ownership Matrix.

### A.4 Scope Decisions

- **In-scope**: All 5 deliverables above, applied to V3R6 modern-era SPECs only (grandfather clause for pre-V3R6 SPECs)
- **In-scope**: Frontmatter optional `era:` field added to spec-frontmatter-schema.md (auto-detected by audit command when absent)
- **In-scope**: M6 run-phase dogfood of `moai spec close` on the 5 known violations (acceptance verification)
- **Out-of-scope** (deferred to follow-up SPEC if needed): retroactive normalization of 145 pre-V3R6 SPECs (covered by grandfather clause; non-blocking)
- **Out-of-scope**: Modification of L60 atomic backfill design itself — `moai spec close` is additive, L60 remains backward-compatible for SPECs that prefer the legacy cadence

### A.5 Out of Scope

#### A.5.1 Out of Scope — Pre-V3R6 SPEC Retroactive Normalization

The 145 historical SPECs (80 N|N|N|N + 42 Y|N|N|N + 23 Y|N|N|Y) MUST NOT be touched by this SPEC. They are grandfather-clause-protected as era-final. `moai spec audit` MUST classify them as PASS-era-final, not surface them as drift. Any retroactive normalization is a separate follow-up SPEC with explicit user opt-in.

#### A.5.2 Out of Scope — L60 Atomic Backfill Redesign

The L60 chicken-and-egg pattern remains valid as a backward-compatible cadence. SPECs that choose the legacy 5-commit close continue to function. `moai spec close` is an additive atomic alternative, not a replacement.

#### A.5.3 Out of Scope — CHANGELOG.md Modification

CHANGELOG.md updates are sync-phase responsibility (owned by manager-docs). This SPEC's plan-phase artifacts MUST NOT modify CHANGELOG.md. Run-phase implementation may add a CHANGELOG entry under v3.0.1 unreleased section during sync milestone.

#### A.5.4 Out of Scope — Template Directory Modifications

`internal/template/templates/**` MUST NOT be modified by this SPEC's plan-phase artifacts. Per CLAUDE.local.md §25 + SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001, dev-internal tokens (SPEC IDs, REQ-LSG-*, audit citations) MUST NOT leak into templates. The 5 deliverables target user-project surfaces only.

---

## B. Requirements (GEARS Notation)

All requirements use GEARS notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format. Subject MAY be any noun (system, CLI, hook, lint engine, audit tool, agent, file).

### B.1 Ubiquitous Requirements

**REQ-LSG-001** — The `moai spec close` subcommand SHALL execute a single atomic commit that transitions spec.md `status: implemented → completed` AND backfills sync_commit_sha AND backfills mx_commit_sha AND updates progress.md `§E.3 status: completed` in one staging cycle.

**REQ-LSG-002** — The `moai spec audit` subcommand SHALL classify every `.moai/specs/SPEC-*/` directory into one of five era buckets (V2.x, V3R2-R4, V3R5, V3R6, unclassified) based on file presence and content heuristics specified in design.md §C.

**REQ-LSG-003** — The pre-commit hook `handle-pre-commit-spec-status.sh` SHALL verify that spec.md `status` field and progress.md `§E.3 status` field agree whenever either file is staged with a status field change.

**REQ-LSG-004** — The spec-lint `OwnershipTransitionRule` extension SHALL emit an `OwnershipTransitionInvalid` finding when a commit subject signals `manager-develop` authorship AND the diff contains a spec.md `status: * → implemented` transition.

**REQ-LSG-005** — The rule file `.claude/rules/moai/workflow/lifecycle-sync-gate.md` SHALL document the era classification heuristic table AND the grandfather clause policy AND the frontmatter optional `era:` field semantics AND cross-reference the Status Transition Ownership Matrix.

### B.2 Event-driven Requirements

**REQ-LSG-006** — **When** the user invokes `moai spec close SPEC-XXX`, the CLI SHALL validate the 4-phase precondition matrix (sync section present, mx section present, all AC PASS, no pending PASS-WITH-DEBT) before staging any change.

**REQ-LSG-007** — **When** the user invokes `moai spec audit --json`, the audit tool SHALL emit a JSON object containing per-SPEC era classification AND drift findings AND grandfather-clause exclusions.

**REQ-LSG-008** — **When** a commit subject matches the canonical pattern `^chore\(SPEC-[A-Z0-9-]+\): 4-phase close — atomic$`, the pre-commit hook SHALL enforce that spec.md `status: completed` is included in the staged diff.

**REQ-LSG-009** — **When** the audit tool detects a Y|Y|Y|Y drift case (4-phase complete but spec.md status not completed), the tool SHALL emit a `Y_Y_Y_Y_StatusDrift` finding with severity `MUST-FIX` and propose the canonical fix command `moai spec close SPEC-XXX --backfill-only`.

### B.3 State-driven Requirements

**REQ-LSG-010** — **While** the `moai spec close` precondition matrix validation is in progress, the CLI SHALL acquire a flock-based file lock on `.moai/state/spec-close.lock` to prevent concurrent close operations on the same SPEC.

**REQ-LSG-011** — **While** the pre-commit hook is executing, the hook SHALL NOT invoke `AskUserQuestion` directly; it SHALL emit exit code 2 with structured JSON output for the orchestrator to translate per agent-common-protocol §Hook Invocation Surface.

### B.4 Capability-gate (Where) Requirements

**REQ-LSG-012** — **Where** the `lint.skip: [OwnershipTransitionInvalid]` opt-out is present in a SPEC's frontmatter, the spec-lint extension SHALL skip the ownership transition check for that SPEC and emit an `OwnershipTransitionSkipped` informational finding.

**REQ-LSG-013** — **Where** the frontmatter `era:` field is absent, the audit tool SHALL auto-detect the era using the heuristic table (design.md §C.2) and emit an `EraAutoDetected` informational finding.

### B.5 Event-detected (Unwanted Behavior) Requirements

**REQ-LSG-014** — **When** the `moai spec close` precondition matrix validation fails (any required artifact missing), the CLI SHALL abort with exit code 1, emit a structured error message naming the missing artifact, AND NOT stage any change.

**REQ-LSG-015** — **When** the pre-commit hook detects a spec.md/progress.md status field mismatch, the hook SHALL emit exit code 2 with structured JSON output identifying the mismatched values AND the canonical resolution path.

---

## C. Non-Functional Requirements

**NFR-LSG-001** — **Performance**: `moai spec audit` SHALL complete within 5 seconds on a project with 200 SPECs (current count: 154 + headroom).

**NFR-LSG-002** — **Backward Compatibility**: SPECs choosing the legacy 5-commit close cadence SHALL continue to function without modification. `moai spec close` is additive.

**NFR-LSG-003** — **Cross-Platform**: All 5 deliverables SHALL function on macOS (Darwin) AND Linux. Windows support is best-effort (pre-commit hook uses POSIX bash; alternative PowerShell wrapper deferred).

**NFR-LSG-004** — **Observability**: `moai spec close` SHALL log every state transition to `.moai/logs/lifecycle-close.log` for audit trail.

**NFR-LSG-005** — **Concurrency Safety**: Multi-session pre-commit hook execution SHALL NOT corrupt progress.md or spec.md frontmatter via concurrent writes (file lock per REQ-LSG-010).

---

## D. Constraints

### D.1 HARD Constraints

1. **[HARD]** The 145 pre-V3R6 SPECs MUST be classified as era-final by the audit tool and MUST NOT be surfaced as drift findings (grandfather clause).
2. **[HARD]** The pre-commit hook MUST NOT invoke `AskUserQuestion` — exit code 2 + JSON output ONLY (REQ-LSG-011).
3. **[HARD]** `moai spec close` MUST be atomic: either all 5 state transitions occur in a single commit OR no transition is staged.
4. **[HARD]** Template directory (`internal/template/templates/**`) MUST NOT be modified by this SPEC's plan-phase artifacts (A.5.4 Out of Scope).
5. **[HARD]** The L60 atomic backfill design MUST remain backward-compatible — `moai spec close` is additive, not a replacement.

### D.2 SHOULD Constraints

1. **[SHOULD]** `moai spec audit` SHOULD support `--filter-era=V3R6` flag for narrowing scope to modern-era SPECs only.
2. **[SHOULD]** The rule file SHOULD include a worked example demonstrating the era auto-detection heuristic in action.
3. **[SHOULD]** The spec-lint extension SHOULD provide a clear remediation hint when emitting `OwnershipTransitionInvalid` (suggest sync-phase delegation).

---

## E. Acceptance Criteria Reference

Acceptance criteria enumerated in `acceptance.md`. Summary: 15 AC-LSG items binding 1:1 to REQ-LSG-001..015, plus 3 NFR verification AC items, plus 1 M6 dogfood AC verifying the 5 remaining violations are resolved via `moai spec close --backfill-only` after implementation.

---

## F. Risks

### F.1 R-LSG-001: Era Classification Heuristic False Positives

**Description**: The era heuristic table (design.md §C.2) relies on file presence and content patterns that may produce false positives for SPECs with non-standard structure.

**Mitigation**: Auto-detection emits informational finding (`EraAutoDetected`) not blocking error; explicit `era:` frontmatter field provides override mechanism.

### F.2 R-LSG-002: Pre-Commit Hook Performance Regression

**Description**: Adding a pre-commit hook that reads spec.md + progress.md may slow `git commit` on SPECs with large progress.md files.

**Mitigation**: Hook only triggers when spec.md or progress.md is staged (early-exit otherwise); NFR-LSG-001 bounds total tool execution at 5s.

### F.3 R-LSG-003: Spec-Lint Extension Breaks Existing Workflows

**Description**: `OwnershipTransitionRule` may emit findings on legacy commits (pre-V3R6) when audit tool is invoked on full repo history.

**Mitigation**: Extension only checks staged diff (not history); grandfather clause applies for era-final SPECs; opt-out via `lint.skip` available.

### F.4 R-LSG-004: Atomic Close Locking Contention

**Description**: File lock on `.moai/state/spec-close.lock` (REQ-LSG-010) may serialize parallel SPEC closes in multi-session workflows.

**Mitigation**: Per-SPEC lock filename (`spec-close-{SPEC-ID}.lock`) eliminates cross-SPEC contention; documented in design.md §D.

### F.5 R-LSG-005: M6 Dogfood Reveals Latent Bugs in Closer

**Description**: Run-phase M6 (dogfood on 5 known violations) may reveal closer bugs after M1-M5 implementation is complete.

**Mitigation**: M6 is final milestone; bugs surface before sync; manager-develop empowered to invoke DIAGNOSE-PATCH-VERIFY inline per L67 mitigation pattern.

---

## G. Related SPECs

- **Predecessor**: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (Status Transition Ownership Matrix originator) — establishes the manager-develop / manager-docs boundary this SPEC enforces mechanically
- **Predecessor**: SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 (GEARS notation canonicalization) — this SPEC uses GEARS 100%
- **Related**: SPEC-V3R6-AGENT-TEAM-REBUILD-001 (17→8 agent catalog) — manager-spec / manager-develop / manager-docs / manager-git ownership matrix consumed by REQ-LSG-004
- **Related**: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (dev-internal token isolation) — D.1.4 HARD constraint enforces this isolation in plan-phase

---

## H. Cross-References

- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — SSOT consumed by REQ-LSG-004 / REQ-LSG-005
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier L threshold (>1000 LOC OR >15 files) satisfied
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface — pre-commit hook integration pattern (exit 2 + JSON, orchestrator translates)
- `CLAUDE.local.md` § 23.9 Tier-based PR Routing — Tier L OR `--pr` flag triggers manager-git
- Memory L60 (atomic backfill chicken-and-egg) — design rationale for additive atomic close
- Memory L67 (manager-docs scope-creep) — pre-commit hook + spec-lint extension defensive layers respond to this pattern
