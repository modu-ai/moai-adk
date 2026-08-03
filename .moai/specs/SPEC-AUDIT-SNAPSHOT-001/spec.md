---
id: SPEC-AUDIT-SNAPSHOT-001
title: "Audit fast-wins: sticky artifact-hash cache, per-tier skip alignment, 4-dim binding promotion, shared diagnostic snapshot"
version: 0.1.0
status: draft
created: 2026-08-03
updated: 2026-08-03
author: manager-spec
priority: P0
phase: "v3.x target"
module: workflow-audit
lifecycle: spec-anchored
tags: "audit, sync-phase, skip-policy, snapshot, plan-auditor, sync-auditor, autonomy-epic"
tier: M
related_specs: [SPEC-AUDIT-GATE-INTEGRITY-001, SPEC-AUTONOMY-RUN-GOAL-001]
---

# SPEC-AUDIT-SNAPSHOT-001 — Audit fast-wins (A1-A4)

## HISTORY

- 2026-08-03 — Initial draft. Codifies §3.5 items A1-A4 of the autonomy-workflow redesign report (`moai-autonomy-workflow-redesign-20260803.html`). First P0 SPEC of the autonomy-workflow-epic. No prior art in the SPEC catalog; closest neighbor is `SPEC-AUDIT-GATE-INTEGRITY-001` (skip-policy invariants), which this SPEC extends but does not supersede.

## §A. User Story

**As a** MoAI maintainer driving multi-SPEC epics through plan→run→sync,
**I want** the sync- and plan-audit pipeline to skip redundant re-audits when nothing changed, to bind the 4-dimension workflow verdict on the happy path, and to share a single diagnostic snapshot across all sync-phase consumers,
**so that** sync-phase latency stops being dominated by duplicate `go test` / `golangci-lint` / `go vet` / `go test -cover` runs and by a cold sync-auditor spawn that re-derives a verdict the 4-dim workflow already produced.

**Outcome hypotheses (from §3.5 design report):**

- A1+A2: plan→run double-spawn is eliminated for every SPEC whose plan artifacts are unchanged; "skip" flips from exception to default for legitimately-passed SPECs.
- A3: a clean sync emits a binding verdict WITHOUT spawning the cold sync-auditor subagent; the 4 parallel xhigh judges subsume the 1 serial judge.
- A4: per sync cycle, test/lint/vet/coverage execution count drops from ~4 to ~1 (the largest single time saving on medium Go SPECs — `golangci-lint` alone is 30-90s cold).

## §B. Scope

**In scope — exactly A1-A4 from §3.5 (existing knobs re-wired, no new mechanism):**

- **A1** — sticky artifact-hash cache (drop the 24h expiry condition; optionally extend the hash subject set for Tier L).
- **A2** — skip-eligible threshold aligned to per-tier PASS threshold (0.75 / 0.80 / 0.85).
- **A3** — 4-dimension workflow verdict promoted to binding on the happy path; cold sync-auditor retained as fallback for INCOMPLETE / must-pass-dim-0 / contested findings.
- **A4** — shared diagnostic snapshot keyed by tree-state (HEAD SHA), consumed by sync-auditor Evidence cells, the `sync-phase-quality-gate.sh` Stop hook, and the 4-dim workflow judges.

**Out of scope — A5-A11 and the broader autonomy epic:** parallel docs/audit (A5), per-edit unification (A8), Stop-chain shell shortening (A10), mode-aware hooks (A11), the `MOAI_AUTONOMY_TIER` mode token, the goal-evaluator HTML dashboard, and the MCP tool surface. These are tracked by sibling P0/P1 SPECs in the epic.

### Out of Scope — Redesign items beyond A1-A4

- A5 (docs ∥ audit parallelization), A8 (per-edit hook integration), A10 (Stop-chain shell shortening), A11 (mode-aware hooks) — separate epic SPECs.
- `MOAI_AUTONOMY_TIER` mode-token introduction — sibling P0 SPEC `SPEC-STOPCHAIN-TRIM-001` (planned).
- Goal-evaluator HTML dashboard and the `moai_goal_render` surface — epic-level P1 work.
- MCP tool surface (`moai_verify_snapshot`, `moai_goal_status`, etc.) — epic-level P1 work.

### Out of Scope — Audit semantics changes

- Changing what the plan-auditor or sync-auditor evaluates (AC content, scoring rubric, severity definitions). A1-A4 change WHEN and HOW OFTEN these run, not WHAT they measure.
- Lowering the PASS thresholds themselves (S 0.75 / M 0.80 / L 0.85 stay); A2 only aligns the SKIP bar to the existing PASS bar.
- Removing the cold sync-auditor agent (it remains the fallback path under A3).

## §C. Requirements (GEARS)

### REQ-AUDIT-SNAPSHOT-001 — Sticky artifact-hash cache (A1)

**Where** the plan-audit verdict cache is consulted, the `ComputeHash` function in `internal/runtime/audit_cache.go` SHALL be the sole determinant of cache validity: a cached verdict whose plan-artifact hash is unchanged SHALL remain valid regardless of the elapsed time since the verdict was recorded.

**When** the orchestrator evaluates the plan-audit skip policy condition set, the orchestrator SHALL NOT apply a "Within 24h" time-window condition; the time-based condition (condition 4 of the four-condition skip contract in `spec-workflow.md` § Phase Transitions / Plan Audit Gate skip policy) is REMOVED.

**Where** the SPEC tier is L, the artifact-hash subject set SHALL include `design.md` and `research.md` in addition to the legacy `{spec.md, plan.md, acceptance.md, tasks.md}` set, so that the sticky cache is both long-lived AND complete for Tier L.

### REQ-AUDIT-SNAPSHOT-002 — Skip-eligible threshold aligned to per-tier PASS (A2)

**Where** the orchestrator evaluates the plan-audit skip policy condition 2 (overall score threshold), the threshold SHALL equal the SPEC's per-tier PASS threshold — Tier S 0.75, Tier M 0.80, Tier L 0.85 — matching `spec-workflow.md` § SPEC Complexity Tier.

The flat `score >= 0.90` skip threshold is RETIRED. A SPEC whose plan-phase audit verdict legitimately PASSED is skip-eligible by default.

### REQ-AUDIT-SNAPSHOT-003 — 4-dimension workflow promoted to binding verdict (A3)

**When** the `.claude/workflows/sync-audit-4dim.js` workflow completes with all four dimensions scoring above their must-pass floor AND the workflow verdict is not `INCOMPLETE`, the orchestrator SHALL treat the workflow's harmonic-mean verdict as BINDING for the sync-phase quality decision on the happy path and SHALL NOT spawn the cold sync-auditor subagent.

**When** any of the following failure modes is observed — (a) the workflow verdict is `INCOMPLETE`, (b) any must-pass dimension scores 0, or (c) a **contested finding** is detected (defined mechanically as: any one of the 4 parallel judges reports a finding at `critical` severity, OR two or more judges return conflicting severity classifications for the same dimension — e.g. one judge marks Functionality `critical` while another marks it `minor`), the orchestrator SHALL spawn the cold sync-auditor subagent as the fallback binding-verdict owner. The two triggering predicates are machine-evaluable from the structured per-judge output the workflow already emits; no orchestrator judgment is required to detect a contested finding.

The cold sync-auditor's verdict domain (PASS/FAIL scoring on the same 4-dimension AC) is unchanged; A3 promotes the workflow verdict on the happy path only, replacing the serial spawn with an attributable parallel diff-check (same AC, 4 parallel judges vs 1 serial judge), not a deletion of the auditor role.

### REQ-AUDIT-SNAPSHOT-004 — Shared diagnostic snapshot keyed by tree-state (A4)

**Where** a sync-phase quality consumer (sync-auditor Evidence cell, `sync-phase-quality-gate.sh` Stop hook, or 4-dim workflow judge) requires a test / lint / vet / coverage result, the consumer SHALL consume a shared diagnostic snapshot recorded once per tree-state and keyed by the current HEAD SHA, rather than each independently re-executing `go test` / `golangci-lint` / `go vet` / `go test -cover`.

**When** the snapshot for the current HEAD SHA is absent (first consumer in a sync cycle) or invalid (SHA mismatch — a new commit has landed), the first consumer to request the snapshot SHALL trigger a single fresh recording; subsequent consumers within the same HEAD SHA SHALL read the recorded result without re-execution.

**When** the HEAD SHA changes between snapshot recording and a consumer read, the consumer SHALL NOT silently serve the stale snapshot; the consumer SHALL either re-trigger a fresh recording for the new SHA or surface an explicit error, preserving the baseline-attribution invariant (`verification-claim-integrity.md` §2).

**When** multiple consumers concurrently request a recording for the SAME HEAD SHA (e.g. the Stop hook firing while sync-auditor is mid-record, or 4-dim judges racing on the same SHA), the recording SHALL be serialized via a claim/lock (file lock or atomic claim-stamp on the snapshot store) so that exactly one consumer performs the recording and the rest read the recorded result. This requirement is mandatory because `internal/verify/store.go` `RecordCheck` (L100-120) is a read-modify-write over the per-key entry — last-writer-wins would silently drop the other consumers' recorded dimensions when two writers race on the same SHA with different command dimensions (`go test` vs `golangci-lint`). The `Save` path (L57-94) is already an atomic rename; the claim/lock brings the multi-dimension recording path to the same atomicity guarantee.

The snapshot infrastructure extended here is the existing `quality-gates-quality.md` Step 0.5.2 mechanism (`moai verify check --key-current` consuming a fresh recorded result). A4 is a wiring change — wire sync-auditor Evidence cells, the Stop hook, and 4-dim judges to consume the snapshot — NOT a new parallel snapshot store.

## §D. Constraints

1. **Tree-state binding is inviolable** (from `verification-claim-integrity.md` §2 + report §3.5 risk callout): the snapshot key MUST be the HEAD SHA; a new commit invalidates the prior snapshot. No consumer may silently serve a stale-SHA result.
2. **A3 binding promotion is an attributable diff-check, not a deletion**: the 4-dim workflow verdict subsumes the cold auditor only because it evaluates the same AC with 4 parallel xhigh judges vs 1 serial judge. The cold auditor remains the fallback for INCOMPLETE / must-pass-0 / contested.
3. **A4 is a wiring change, not new machinery**: the snapshot infrastructure already exists (Step 0.5.2). Wire consumers to it; do NOT invent a parallel snapshot store.
4. **Backward compatibility**: minimal-harness users (`levels.minimal.evaluator: false`) and the existing 4-condition skip-policy set (minus the retired time-window) MUST keep working. A1/A2 modify conditions; A3/A4 add a binding/snapshot layer on top.
5. **No new user-facing CLI surface in this SPEC**: the snapshot is consumed internally by sync-phase consumers; MCP/CLI exposure is deferred to epic-level P1 work.
6. **Audit semantics unchanged**: A1-A4 change WHEN/HOW OFTEN audits run, not WHAT they measure. PASS thresholds, AC content, and severity definitions are immutable in this SPEC.

## §E. Assumptions

1. The `.claude/workflows/sync-audit-4dim.js` workflow already emits a structured verdict (pass/fail per dimension + harmonic mean + INCOMPLETE flag) consumable by the orchestrator without modification to its output schema. (A3 adds a BINDING interpretation of that output; it does not change the output itself.)
2. The `quality-gates-quality.md` Step 0.5.2 snapshot mechanism is reachable from all three consumers (sync-auditor, Stop hook, 4-dim judges) via the existing `moai verify check --key-current` interface or an equivalent programmatic call.
3. The four-condition skip-policy contract in `spec-workflow.md` is the single authoritative skip surface; no parallel skip path exists in skill YAML or agent bodies that would silently bypass the A1/A2 changes.
4. Tier L's `design.md` and `research.md` are present-and-non-empty iff the SPEC is genuinely Tier L; absence of these files for a self-declared Tier L SPEC is an authoring error, not a condition this SPEC accommodates.

## §F. Open Questions (for plan-auditor)

- **OQ-1** (open): Does the 4-dim workflow's output schema already carry a "must-pass-dim-scored-0" signal distinct from the harmonic-mean verdict, or must A3 add such a signal? (Determines whether A3 is a pure consumer-side change or requires a workflow-output augmentation.)
- **OQ-2** (RESOLVED by plan-auditor iter 1 — `internal/verify/store.go` inspection): `Save` (L57-94) is atomic via rename, but `RecordCheck` (L100-120) is a read-modify-write — concurrent same-SHA writers racing on different command dimensions race last-writer-wins, dropping the other dimensions. Resolution folded into REQ-AUDIT-SNAPSHOT-004 paragraph 4 as a mandatory claim/lock SHALL (no longer conditional).

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 (items A1-A4), §1.3 (sync-phase bottlenecks), §3.5 risk callout (verification-claim integrity preservation).
- Skip-policy SSOT: `.claude/rules/moai/workflow/spec-workflow.md` § Phase Transitions / Plan Audit Gate skip policy (conditions 1-4) + § SPEC Complexity Tier (PASS thresholds).
- Hash source: `internal/runtime/audit_cache.go` `ComputeHash` (~line 90) + `planArtifactNames`.
- 4-dim workflow: `.claude/workflows/sync-audit-4dim.js` (header lines 5-9, 48-56).
- Sync skill surface: `.claude/skills/moai/workflows/sync.md` FO-SYNC-1 (lines 60-69).
- Snapshot mechanism: `.claude/skills/moai/workflows/sync/quality-gates-quality.md` Step 0.5.2.
- Auditor verification surface: `.claude/agents/moai/sync-auditor.md` § Per-Dimension Mechanical Verification.
- Stop hook: `.claude/hooks/moai/sync-phase-quality-gate.sh`.
- Integrity invariant: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 + §2.

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-AUDIT-SNAPSHOT-001 (A1): sticky cache — past-24h unchanged-hash skip still fires.
- AC-AUDIT-SNAPSHOT-002 (A2): per-tier skip threshold — a 0.78 Tier M SPEC is skip-eligible.
- AC-AUDIT-SNAPSHOT-003 (A3): clean sync emits binding verdict with no cold sync-auditor spawn.
- AC-AUDIT-SNAPSHOT-004 (A4): single test/lint/vet/cover run shared across 3 consumers; new SHA invalidates.
