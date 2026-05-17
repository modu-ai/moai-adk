---
id: SPEC-WORKTREE-DOCS-001
title: "Worktree Workflow Documentation Harmonization (L1/L2/L3 Opt-In Policy)"
version: "0.1.0"
status: draft
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, documentation, harmonization, override, opt-in, terminology"
---

# Design — SPEC-WORKTREE-DOCS-001

## 1. Architectural Decisions

### AD-001 — Terminology Scheme: Numeric Prefix (L1/L2/L3) over Descriptive

**Decision**: Adopt `L1` / `L2` / `L3` numeric-layer prefixes as the canonical
disambiguation vocabulary for worktree mentions.

**Alternatives considered**:

- (a) **Descriptive prefixes** ("Claude Code Native worktree", "SPEC worktree", "Plan
  worktree"): More readable in isolation but each mention adds 2-4 tokens. Inconsistent
  abbreviation across files (e.g., "Claude Code worktree" vs "Claude Native worktree"
  vs "ephemeral worktree") creates re-conflation risk.
- (b) **Numeric + descriptive compound** ("L1 (Claude Code Native)", "L2 (SPEC
  worktree)"): Verbose; first-occurrence helpful but later occurrences are noisy.
- (c) **Numeric only** (chosen): Single canonical token. First occurrence in a rule
  file pairs with glossary definition (e.g., "L1 worktree (see § Terminology
  Glossary)"); subsequent occurrences use the bare token. Lowest token cost +
  unambiguous mapping.

**Rationale**: The forensic audit § H6 (Naming Overload) showed that the cognitive
load of disambiguating "worktree" across 3 layers is high. Numeric layers force the
reader to consult the glossary once per session, then operate on a single mnemonic
("L1 = Claude Code, L2 = moai CLI, L3 = workflow flag"). This matches the audit's
recommended terminology in § H6 Verdict.

**Consequence**: Initial reading curve (~30 seconds to internalize the glossary) is
amortized across all future worktree references. Rule file diffs become consistent
across all 5 files.

### AD-002 — Override Propagation: Global Rules (Option A) over Project-Overrides Directory (Option B)

**Decision**: Update the global rules in-place (Option A). Do NOT create
`.claude/rules/moai/project-overrides/`.

**Alternatives considered**:

- (a) **Option A — Global rules** (chosen): Edit `spec-workflow.md`, `CLAUDE.md` §14,
  etc. directly with the new opt-in semantics. Pros: single source of truth; no
  precedence ambiguity; cheapest to maintain. Cons: this is a doctrine-level change
  applicable to all moai-adk-go users (i.e., the framework itself, not just this
  project), so it must be doctrine-level.
- (b) **Option B — Project overrides directory**: Create
  `.claude/rules/moai/project-overrides/worktree-disabled.md` with paths-restricted
  frontmatter so it loads only in this project. Pros: future projects could keep the
  "MUST" mandates. Cons: introduces precedence rules; needs a SPEC for how overrides
  interact with global rules; this project IS the framework, not a downstream consumer.
- (c) **Option C — Pre-flight check in orchestrator**: Read `feedback_worktree_*.md`
  at session start and dynamically adjust behavior. Pros: keeps rules pristine.
  Cons: requires Go-code changes (out of plan-phase scope); adds runtime check
  overhead; doctrine should be in rules, not auto-memory.

**Rationale**: moai-adk-go IS the framework. Any "project-specific" override in this
repo is by definition a framework-level doctrine. The audit § Coverage Gap 2 noted
Option A as the simplest path; Option B was offered as a hedge for projects DIFFERENT
from this one. Since we are editing the framework itself, Option A is correct.

**Consequence**: `.claude/rules/moai/project-overrides/` directory mechanism remains
unused. If a downstream user (different project) ever needs the old "MUST worktree"
semantics, they would author their own project-overrides — a future SPEC could
formalize that mechanism then. For now, no overhead.

### AD-003 — Backward Compat: Preserve Wave 5 Primitive Documentation + Code

**Decision**: Soften the rules around the Wave 5 primitive but PRESERVE both the
documentation and the Go CLI code.

**Alternatives considered**:

- (a) **Delete worktree-state-guard.md + delete `internal/cli/worktree/guard.go`**:
  Removes "operationally dead code" per audit § Coverage Gap 3. Pros: smaller
  surface area. Cons: irreversible; loses the diagnostic primitive that
  power-users could invoke manually; SPEC-V3R3-CI-AUTONOMY-001 Wave 5's investment
  becomes wasted; reactivation in a future Wave 6 would require a new SPEC + reimpl.
- (b) **Preserve docs + code, add dormancy banner** (chosen): The forensic audit
  recommended a warning banner ("Primitives exist but orchestrator wiring is
  deferred to Wave 6"); we adopt this verbatim in T5. Pros: zero-cost preservation;
  Wave 6 can re-activate; power users have a documented manual path; supports
  REQ-WTD-005.
- (c) **Move docs from `rules/` to `references/`** (audit recommendation):
  Reclassifies the file as informational. Pros: signals lower normative weight.
  Cons: changes paths frontmatter, requires updating cross-references in 3+ other
  files; orthogonal to the doctrine question.

**Rationale**: REQ-WTD-005 requires backward compatibility. The primitive's existence
is a feature, not a bug — it gives power users an escape hatch when Claude Code's L1
isolation produces empty `worktreePath`. Deletion would close a future option for no
present gain.

**Consequence**: `worktree-state-guard.md` gains a 2-line dormancy note in § Overview.
The Go CLI code is untouched. CI tests for the Go code (if any in
`internal/cli/worktree/guard_test.go`) continue to pass without modification.

### AD-004 — CLAUDE.md §14 Bullet Preservation: Soften, Don't Remove

**Decision**: Keep all 4 bullets in `CLAUDE.md` §14 but convert HARD to SHOULD/advisory.
Do NOT delete the bullets.

**Alternatives considered**:

- (a) **Delete §14 entirely**: Cleanest cosmetically but removes useful guidance
  about when isolation is beneficial (e.g., parallel team-mode implementations).
- (b) **Replace 4 bullets with 1-line cross-reference** to worktree-integration.md:
  Reduces CLAUDE.md size by ~6 LOC. Cons: agents that read only CLAUDE.md miss the
  guidance.
- (c) **Soften, preserve structure** (chosen): Each bullet's intent (which agent
  types benefit from isolation) remains; the enforcement tier shifts. Pros:
  zero structural change for downstream readers; the guidance is still valuable as
  a hint to Claude Code runtime even though moai no longer mandates the choice.

**Rationale**: CLAUDE.md is the orchestrator's primary execution directive. Removing
§14 would silently delete operational knowledge. Softening preserves the knowledge
while making explicit that the decision authority is Claude Code runtime, not moai.

**Consequence**: §14 stays at ~6 LOC. Bullet 1 (write teammates) and Bullet 4 (GitHub
fixers) shift from MUST to SHOULD. Bullets 2 (read-only MUST NOT) and 3 (one-shot
SHOULD) keep their existing tier (already advisory).

## 2. Failure Modes & Recovery

### FM-1 — Run-PR T2 Audit Reports Bare "worktree" Occurrences

**Scenario**: After T1+T3-T6 edits, `scripts/audit-workflow-terminology.sh`
(implemented in T0) reports residual bare "worktree" mentions in the affected files.

**Recovery**: T2 task explicitly handles each finding by either prefixing with
L1/L2/L3 or adding a parenthetical clarification. If the audit reports cross-file
findings outside the 6 affected files (R1), those findings are recorded in
acceptance.md AC-WTD-007 and become follow-up SPEC candidates — NOT in-scope for
this SPEC.

### FM-2 — Run-PR Acceptance Gate Detects Lingering [HARD] Worktree Mandate

**Scenario**: AC-WTD-001 grep audit detects a `[HARD]` line in
`spec-workflow.md`, `CLAUDE.md`, or another affected file that still forces
worktree usage.

**Recovery**: The run-phase agent re-applies T3 or T4 to the missed location.
Acceptance gate re-runs; binary PASS/FAIL.

### FM-3 — Memory File T7/T8 Race Condition (Multi-Session)

**Scenario**: User has two Claude Code sessions open. Session A is running this SPEC;
Session B reads `MEMORY.md` after T7 completes but before T8 commits.

**Recovery**: Session B reads either pre-T7 (no new memory) or post-T7 (new memory
present, no index entry). Both states are recoverable: Session B's next MEMORY.md
read picks up T8 once committed. No data corruption because shell-level file writes
are atomic at this size.

### FM-4 — Plan-Auditor Rejects spec.md After Plan-PR Submission

**Scenario**: Plan-auditor finds a gap in spec.md (e.g., EARS REQ ambiguity, missing
Out-of-Scope item).

**Recovery**: Re-submit a revised spec.md / plan.md in the plan-PR branch. No
downstream impact — plan-PR has no run-phase deliverables.

### FM-5 — Spec-Status-Sync Fails Post-Plan-PR-Merge

**Scenario**: After plan-PR merges, the CI workflow that transitions `status: draft`
→ `status: planned` in spec.md fails (e.g., HTTP 403 push permission, as previously
seen in SPEC-V3R4-NAMESPACE-001).

**Recovery**: Manually update `status: draft` → `status: planned` in a follow-up
chore commit on main. Run-PR proceeds independently of frontmatter state because
the run-phase agent reads the SPEC ID, not the status field.

## 3. Cross-Cutting Concerns

### CC-1 — Token Budget Impact

The 5 rule files + CLAUDE.md combined post-edit size is estimated at ~95 LOC larger
than pre-edit (terminology glossary adds ~25 LOC; 8 task touches add ~70 LOC; net
some softening reduces ~10 LOC). The progressive-disclosure system loads these rules
only when their `paths` frontmatter matches, so the per-session token cost is
bounded by the rules that load. Net incremental token cost per session: ~300-500
tokens depending on active phase. Acceptable.

### CC-2 — Lessons-Protocol Compliance

The supersede convention `[SUPERSEDED by feedback_worktree_autonomous]` follows MoAI
Lessons Protocol per `.claude/rules/moai/core/moai-constitution.md` § Lessons
Protocol. T8 verification gate checks the marker is present in both the old memory
file body AND the MEMORY.md index entry (when present).

### CC-3 — No Code Path Changes

This SPEC strictly edits documentation and auto-memory. No Go code is modified. No
test files are added or removed. No CI configurations change. Run-PR's `go test ./...`
gate is a no-op for this SPEC (must still pass to confirm no accidental code-touch).

### CC-4 — Cross-SPEC Coupling

This SPEC is operationally independent of SPEC-V3R4-WORKFLOW-SPLIT-001 (Bundle F
workflow phase-split, currently in-progress on `chore/SPEC-V3R4-WORKFLOW-SPLIT-001-wave-0`).
The two SPECs touch overlapping prose in `spec-workflow.md` only in the § Phase
Overview / § SPEC Phase Discipline sections. Run-PR for this SPEC SHOULD wait for
Bundle F Wave 1 to merge OR explicitly rebase against the latest main HEAD before
run-phase begins. If a textual conflict arises in `spec-workflow.md` during
run-phase, the manager-develop agent will surface it for resolution.

## 4. References

- Forensic Audit: `.moai/research/worktree-forensic-audit-2026-05-17.md` §§ H6 (Naming
  Overload), Coverage Gap 2 (Override Propagation), Coverage Gap 3 (Wave 5 Primitive),
  Category 8 (Testing & CI implications)
- Lessons Protocol: `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol
- SPEC Frontmatter Schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Worktree-state-guard primitive: `internal/cli/worktree/guard.go` + tests in
  `internal/cli/worktree/guard_test.go` (preserved by AD-003)
