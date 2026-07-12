---
id: SPEC-HARNESS-EVOLVE-003
title: "Curator production wiring — Tier-Surface mapping + validation gates + re-proposal suppression"
version: "0.1.0"
status: draft
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness/safety, internal/harness/curator, internal/harness, internal/config"
lifecycle: spec-anchored
tags: "harness-evolve-epic, curator-wiring, tier-surface, l2-canary, l3-contradiction, negative-evidence, frozen-guard, re-proposal-suppression, glm-observe-only"
era: V3R6
tier: L
depends_on: [SPEC-HARNESS-EVOLVE-002]
---

# SPEC-HARNESS-EVOLVE-003 — Implementation Plan

> Counterpart to `spec.md` (requirements SSOT), `acceptance.md` (AC matrix),
> `design.md` (architecture decisions), `research.md` (codebase investigation).
> This document owns the milestone decomposition, file map, pre-flight checks,
> constraints, and self-verification deliverables for the run-phase delegations.

## §A. Context

### A.1 Epic position

This SPEC is **M3 of the HARNESS-EVOLVE Epic** (5 SPECs + 2 horizons per design
SSOT §7). EVOLVE-001 (observation) and EVOLVE-002 (write layer) are CLOSED.
This SPEC activates the production wiring that makes the harness actually
self-evolve. EVOLVE-004 (console verbs) and EVOLVE-005 (Recall wiring + typed
parser + template deployment) follow.

### A.2 SPEC artifact paths

- `.moai/specs/SPEC-HARNESS-EVOLVE-003/spec.md` — requirements SSOT (35 REQs)
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/plan.md` — this file
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/acceptance.md` — AC matrix
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/design.md` — architecture decisions
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/research.md` — codebase investigation
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/progress.md` — §E skeleton (run-phase fills)

### A.3 Current branch + baseline

- Branch: `main` (Hybrid Trunk Route A — Tier L defaults to Route B PR route,
  but plan-phase commits land on `main` per the Late-branch precondition; the
  PR branch is created at sync time)
- EVOLVE-002 closed at origin `fa2a3086a` (sync backfill); depends_on fulfilled
- Baseline: the EVOLVE-002 curator package (20 files) + the safety pipeline
  (7 files) are the immediate predecessors

### A.4 PRESERVE targets (do NOT modify — EVOLVE-002 / EVOLVE-001 output)

- `internal/harness/curator/writer.go` — `WriteManagedBlock` signature/behavior
- `internal/harness/curator/tier_gate.go` — `TierGatedWrite` signature/behavior
- `internal/harness/curator/approval.go` — `WriteManagedBlockGated` /
  `ApprovalDecision` / `RejectionRecorder` signatures
- `internal/harness/curator/crud.go` — per-bullet CRUD signatures
- `internal/harness/curator/append.go` — `AppendLearnedLocal` signature
- `internal/harness/curator/budget.go` — budget/cap enforcement
- `internal/harness/routing/` (EVOLVE-001 output) — read-only consumption only
- `internal/harness/learner.go` — 4-tier ladder untouched
- `internal/harness/safety/pipeline.go:14-15` — L1→L2→L3→L4→L5 order immutable
- `internal/harness/safety/rate_limit.go` — L4 unchanged

### A.5 EXTEND targets (additive — this SPEC's write surface)

- `internal/harness/safety/frozen_guard.go` — expand `frozenPrefixes` (A1, REQ-HEV3-023)
- `internal/harness/safety/pipeline.go` — replace `l3ContradictionCheck` no-op body (REQ-HEV3-015)
- `internal/harness/safety/contradiction.go` — add Frozen-rules contradiction detector
- `internal/harness/safety/canary.go` — extend for Curator-path shadow-apply (if needed)
- `internal/config/types.go` — extend `AutoDetectionConfig` with value-range validation
- `internal/harness/applier.go` — wire the Curator dispatch call chain (new production callers)

### A.6 NEW files (this SPEC creates)

- `internal/harness/negative_evidence.go` — A7 registry data structure + writer + reader
- `internal/harness/negative_evidence_test.go` — registry tests
- `internal/harness/safety/frozen_rules.go` — Frozen-rule registry consulted by L3
- `internal/harness/safety/frozen_rules_test.go` — Frozen-rule registry tests
- `internal/harness/curator/dispatch.go` — the tier→surface dispatch + L5 round injection (the production wiring of TierGatedWrite / WriteManagedBlockGated)
- `internal/harness/curator/dispatch_test.go` — dispatch tests
- `.moai/state/negative-evidence.jsonl` — the registry data file (runtime-created, gitignored runtime state)

## §B. Known Issues (auto-injected per manager-develop-prompt-template §B)

- **B1 Cross-platform build tags**: the new files are pure Go (no syscall); standard build tags suffice. Verify `GOOS=windows GOARCH=amd64 go build ./...`.
- **B2 Cross-SPEC policy conflict**: check that EVOLVE-002's `curator` package API is not concurrently modified by a sibling SPEC. The `curator` package is PRESERVE for this SPEC.
- **B3 C-HRA-008 subagent boundary**: the new dispatch/registry/wiring code MUST NOT invoke `AskUserQuestion`. AC: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/safety/ internal/harness/negative_evidence.go internal/harness/curator/dispatch.go | grep -v _test.go | grep -v '//'` returns 0 matches.
- **B4 Frontmatter canonical schema**: use `created:`/`updated:`/`tags:` (already correct in this SPEC's frontmatter).
- **B5 CI 3-tier awareness**: spec-lint, golangci-lint, Test each fail independently.
- **B6 spec-lint heading convention**: `### Out of Scope — <topic>` H3 sub-headings present in spec.md §E (verified — 7 H3 sub-headings).
- **B7 observer.go capture path**: the A7 registry writer uses absolute paths under `.moai/state/`; resolve via the existing project-root pattern (no `os.Getwd()` leak).
- **B8 Working tree hygiene**: the registry `.moai/state/negative-evidence.jsonl` is runtime state (gitignored); do NOT commit populated registry data. Commit only the schema/validator mechanism.
- **B9 Git commit + push**: manager-develop commits + pushes directly per Hybrid Trunk Route A (Tier L may switch to Route B at sync — orchestrator decides). Conventional Commits `feat(SPEC-HARNESS-EVOLVE-003): M{N} <subject>`.
- **B10 Untouched paths PRESERVE**: do NOT touch EVOLVE-002 artifacts (closed), EVOLVE-001 routing/, other SPEC directories, runtime-managed `.moai/state/` (except the registry file the SPEC owns), `.moai/research/`.
- **B11 AskUserQuestion prohibited**: return blocker reports; the orchestrator runs the L5 round.
- **B12 CHANGELOG**: manager-docs owns it (sync-phase).

## §C. Pre-flight (run before any code change)

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Verify EVOLVE-002 API is INERT (baseline 0 production callers)
grep -rn 'TierGatedWrite\|WriteManagedBlockGated' internal/ cmd/ pkg/ \
  | grep -v _test.go | grep -v 'internal/harness/curator/'
# Expected: EMPTY (0 production callers — confirms the wiring target)

# 5. Verify L3 is a no-op (the replacement target)
sed -n '68,74p' internal/harness/safety/pipeline.go
# Expected: the l3ContradictionCheck returning empty ContradictionReport{}

# 6. Verify safety frozen-guard scope (A1 expansion target)
grep -A5 'var frozenPrefixes' internal/harness/safety/frozen_guard.go
# Expected: the current 4-prefix list (no permission surfaces)

# 7. Verify auto_detection already exists (A6 is registration, not creation)
grep -n 'auto_detection' .moai/config/sections/harness.yaml
# Expected: line 2 (the block already exists)
```

## §D. Constraints (DO NOT VIOLATE)

- **PRESERVE**: EVOLVE-002 `curator` package API signatures (§A.4) — the writer is the inner primitive; the gates + dispatch + registry are outer decorators.
- **PRESERVE**: EVOLVE-001 `routing/` package — read-only consumption only.
- **PRESERVE**: `safety/pipeline.go` L1→L2→L3→L4→L5 order (line 14-15 `[HARD]`).
- **PRESERVE**: the existing `learner.go` 4-tier ladder `[1,3,5,10]`.
- **FORBIDDEN**: invoking `AskUserQuestion` from any new code (subagent boundary, REQ-HEV3-031).
- **FORBIDDEN**: bypassing `WriteManagedBlockGated` (the L5 gate is invariant, REQ-HEV3-006).
- **FORBIDDEN**: permanently suppressing a pattern key in the A7 registry (REQ-HEV3-022).
- **FORBIDDEN**: shipping registry DATA to templates (REQ-HEV3-029).
- **FORBIDDEN**: `--no-verify`, `--amend`, force-push to main.
- **REQUIRED**: Conventional Commits `feat(SPEC-HARNESS-EVOLVE-003): M{N} <subject>` + `🗿 MoAI` trailer.
- **REQUIRED**: each gate's reachability verified by a behavior-verifiable AC (inject → assert), NOT a grep-count (feedback_ac_token_presence_not_reachability).

## §E. Self-Verification Deliverables (per manager-develop-prompt-template §E)

When manager-develop reports completion, it MUST include:

- **E1 AC Binary PASS/FAIL Matrix** — every AC-HEV3-001…NNN with status + command + output.
- **E2 Cross-Platform Build** — `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3 Coverage** — `go test -cover ./internal/harness/safety/... ./internal/harness/curator/... ./internal/harness/... ./internal/config/...` ≥ 90%.
- **E4 Subagent Boundary** — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/safety/ internal/harness/negative_evidence.go internal/harness/curator/dispatch.go | grep -v _test.go | grep -v '//'` returns 0.
- **E5 Lint** — `golangci-lint run --timeout=2m` — NEW issues only (baseline separated).
- **E6 Branch HEAD + Push** — commit SHAs + push result.
- **E7 Blocker Report** — if a wiring need reveals an EVOLVE-002 API gap, return a blocker (do NOT modify the EVOLVE-002 API in this SPEC's run-phase).
- **E8 Reachability evidence** — for each gate (L1 A1, L2, L3, A7), the behavior-verifiable AC output showing the gate FIRES (not a grep-count). This is the feedback_ac_token_presence_not_reachability binding.

## §F. Milestones (priority-ordered, no time estimates)

### M0 — A1 Permission-surface Frozen expansion + foundation

**Scope**: expand `safety/frozen_guard.go` `frozenPrefixes` (REQ-HEV3-023/024/025);
add the Frozen-rule registry skeleton (`safety/frozen_rules.go`) that L3 will
consult. Foundation milestone — A1 is the cheapest pillar (a list expansion +
tests) and unblocks the L3 Frozen-rules contradiction check (M3).

**Files**: `internal/harness/safety/frozen_guard.go` (extend),
`internal/harness/safety/frozen_rules.go` (new), tests.

**Exit**: AC-HEV3-023…025 (settings.json-targeted proposal → L1-blocked; guard
self-protection).

### M1 — A7 Negative-evidence registry (data structure + writer + reader)

**Scope**: the registry at `.moai/state/negative-evidence.jsonl`
(REQ-HEV3-018); the writer (register-on-reject REQ-HEV3-019,
register-on-rollback REQ-HEV3-020); the re-proposal block with cooldown + N-new-evidence
threshold (REQ-HEV3-021/022). Reuse the `canary_veto.go` cooldown primitive
where semantics overlap.

**Files**: `internal/harness/negative_evidence.go` (new),
`internal/harness/negative_evidence_test.go` (new).

**Exit**: AC-HEV3-018…022 (registry CRUD; same-key re-proposal blocked; rollback
auto-registers; permanent-suppression forbidden).

### M2 — Tier↔Surface mapping activation (A6) + value-range validation

**Scope**: register `auto_detection` as a Tier-4 Evolvable surface
(REQ-HEV3-001); add value-range validation to `AutoDetectionConfig`
(REQ-HEV3-002); activate tier→surface dispatch (REQ-HEV3-003/004).

**Files**: `internal/config/types.go` (extend AutoDetectionConfig),
`internal/harness/curator/dispatch.go` (new — the tier→surface dispatch).

**Exit**: AC-HEV3-001…004 (auto_detection registered; out-of-range threshold
rejected; Tier-3 writes CLAUDE.local.md NOT CLAUDE.md; Tier-4 the reverse).

### M3 — L3 Contradiction activation (replace no-op)

**Scope**: replace the `pipeline.go:70-73` no-op with a real Frozen-rules
contradiction check (REQ-HEV3-015/016/017) consulting the M0 Frozen-rule
registry.

**Files**: `internal/harness/safety/pipeline.go` (replace l3ContradictionCheck
body — NOT the L1→L5 order), `internal/harness/safety/contradiction.go` (add
Frozen-rules detector).

**Exit**: AC-HEV3-015…017, AC-HEV3-033 (inject a contradiction → assert
`RejectedBy == 3`; L3 reachability baseline 0 → ≥1).

### M4 — L2 Canary activation for the Curator path

**Scope**: wire L2 shadow-apply + regression gate for the Curator write path
(REQ-HEV3-012/013/014); wire the Canary-veto → A7 registry auto-register
on rollback (REQ-HEV3-014).

**Files**: `internal/harness/safety/canary.go` (extend for Curator path),
`internal/harness/curator/dispatch.go` (wire the L2 consult).

**Exit**: AC-HEV3-012…014, AC-HEV3-034 (inject a regression → assert
`RejectedBy == 2`; L2 reachability baseline 0 → ≥1).

### M5 — L5 round injection + production wiring (TierGatedWrite / WriteManagedBlockGated)

**Scope**: the production wiring that wires the dispatch → Pipeline.Evaluate
→ L5 `WriteManagedBlockGated` (REQ-HEV3-005/006/007/008/009/010/011); the GLM
observe-only gate (REQ-HEV3-026/027); the A7 registry consult at the
dispatch entry (REQ-HEV3-035).

**Files**: `internal/harness/curator/dispatch.go` (the full call chain),
`internal/harness/applier.go` (the production caller entry).

**Exit**: AC-HEV3-005…011, AC-HEV3-026…027, AC-HEV3-035 (TierGatedWrite has a
production caller baseline 0 → ≥1; WriteManagedBlockGated same; L5 round
injects ApprovalDecision; GLM observe-only).

### M6 — Integration + anti-fabrication + template neutrality + Go quality

**Scope**: integration tests (full Curator cycle: observe → aggregate →
tier-qualify → L1→L5 → write/rollback → A7 register); anti-fabrication
validation (REQ-HEV3-028); template-neutrality verification (REQ-HEV3-029);
coverage to ≥ 90% (REQ-HEV3-030); subagent boundary final grep (REQ-HEV3-031);
no-new-hook-surface check (REQ-HEV3-032).

**Files**: integration test files, template-neutrality test.

**Exit**: AC-HEV3-028…032; full AC matrix green; coverage ≥ 90%.

## §G. Anti-Patterns (this SPEC's shape)

- **AP-HEV3-001 — Inert wiring**: shipping the dispatch call site behind a
  permanently-false feature flag so the baseline-0 → ≥1 reachability AC passes
  on paper while the wiring never fires in production (REQ-HEV3-011 violation).
- **AP-HEV3-002 — grep-count gate AC**: an AC that asserts `grep -c
  'ContradictionReport' pipeline.go` ≥ 1 while the L3 evaluator remains a
  no-op (feedback_ac_token_presence_not_reachability — the named anti-pattern).
- **AP-HEV3-003 — Autonomous write path**: a Curator code path that calls
  `WriteManagedBlock` directly, bypassing `WriteManagedBlockGated`
  (REQ-HEV3-006 violation).
- **AP-HEV3-004 — Permanent suppression**: an A7 registry entry with
  `cooldown_until: null` or `9999-12-31` (REQ-HEV3-022 violation).
- **AP-HEV3-005 — Permission-surface Curator target**: a dispatch path that
  admits a proposal targeting `settings.json` past L1 (REQ-HEV3-024 violation).
- **AP-HEV3-006 — Cross-surface leak**: a Tier-3 proposal written to CLAUDE.md
  (REQ-HEV3-004 violation).
- **AP-HEV3-007 — Registry DATA in templates**: shipping a populated
  `negative-evidence.jsonl` or a populated `auto_detection.rules` threshold
  correction to `internal/template/templates/` (REQ-HEV3-029 violation).
- **AP-HEV3-008 — L5 round in subagent**: the dispatch code invoking
  `AskUserQuestion` directly (REQ-HEV3-031 violation).
- **AP-HEV3-009 — Pipeline reorder**: changing the L1→L2→L3→L4→L5 order
  (pipeline.go:14-15 HARD violation).

## §H. Cross-References

- `.moai/specs/SPEC-HARNESS-EVOLVE-003/spec.md` — requirements SSOT
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/acceptance.md` — AC matrix
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/design.md` — architecture decisions
- `.moai/specs/SPEC-HARNESS-EVOLVE-003/research.md` — codebase investigation
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — §A-E delegation template (Tier L mandatory)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — defect-claim verification (research.md citations)

## §I. NEEDS CLARIFICATION (orchestrator-mediated AskUserQuestion before Implementation Kickoff Approval)

These are the architecture-level open questions. plan-auditor will flag any
unresolved `[NEEDS CLARIFICATION]` as a clarification-gate finding.

### H-1 — A7 cooldown duration + N-new-evidence default

**Question**: the A7 re-proposal block requires "cooldown elapsed AND N new
evidences". What are the defaults? The existing `canary_veto.go` uses 48h
(line 27: `canaryVetoCooldown = 48 * time.Hour`). Should A7 reuse 48h, or use
a different duration (e.g. 7d for a stronger suppression)? And the N default —
the SSOT says "N NEW evidences" without a number; spec.md proposes N=3.

**Default (spec.md)**: cooldown = 48h (reuse canary_veto), N = 3 new evidences.
**Impact if different**: M1 registry tests parameterize the values; no
architectural change.

### H-2 — A6 auto_detection value-range bounds

**Question**: the value-range validation (REQ-HEV3-002) needs `[lower, upper]`
bounds per threshold field (`file_count`, `spec_priority`, `domain`, etc.).
What are the canonical bounds? The current `harness.yaml` uses
`file_count <= 3`, `file_count > 3`, `security_keywords`, `spec_priority ==
critical`. Should the bounds be hardcoded in a Go struct, or read from a
schema file?

**Default (spec.md)**: hardcoded in `AutoDetectionConfig` Go struct with
documented bounds (e.g. `file_count` ∈ [1, 1000]); the bounds ship as a
neutral empty schema to templates (REQ-HEV3-029).
**Impact if different**: M2 validator tests + the template-neutrality check.

### H-3 — L3 Frozen-rules contradiction: rule set scope

**Question**: the L3 contradiction check (REQ-HEV3-015) consults "the
registered Frozen-rule set". Which rules are registered? The design SSOT §4
Frozen-zone row lists `.claude/rules/moai/**`, `.claude/agents/moai/**`,
evaluators, templates, permission surfaces, L1-L5 pipeline code. Should the
Frozen-rule registry enumerate rule FILES (a glob), or rule IDENTIFIERS (a
typed registry)?

**Default (spec.md)**: M0 ships a typed registry of Frozen-rule identifiers
(rule path + a stable name), seeded from the safety/frozen_guard.go prefix
list; the registry is extensible without code change.
**Impact if different**: M0 scope (file glob vs typed registry).

### H-4 — Curator dispatch entry: where in applier.go does the call chain attach?

**Question**: the production wiring (M5) needs a production caller entry in
`applier.go`. The existing `Apply()` (applier.go:328) is the harness-applier
path (skill frontmatter edits). Should the Curator Learned-surface dispatch
attach as (a) a new method on Applier (`ApplyCurator()`), (b) a separate
type in `curator/dispatch.go`, or (c) an extension of the existing `Apply()`
with a `ProposalKind` switch?

**Default (spec.md)**: option (b) — a separate `curator.Dispatch` type, because
the Curator path has the A7 registry consult + GLM gate + L5 round that the
harness-applier path does not. The Applier is PRESERVE for the existing path;
the new dispatch is a sibling.
**Impact if different**: M5 file map (applier.go extend vs curator/dispatch.go new).
