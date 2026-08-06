---
id: SPEC-SYNC-PARALLEL-DOCS-001
title: "Sync-phase parallelization: docs∥audit (A5), MX-early+parallel (A7), §E+7-batch attributable diff-check (A9), plan-auditor Tier retry ceilings (A6)"
version: 0.1.0
status: draft
created: 2026-08-07
updated: 2026-08-07
author: manager-spec
priority: P2
phase: "v3.x target"
module: "workflow-sync"
lifecycle: spec-anchored
tags: "sync-phase, parallelization, docs-audit, mx-tag, section-e, plan-auditor, autonomy-epic, verification-claim-integrity"
tier: M
related_specs: [SPEC-AUDIT-SNAPSHOT-001, SPEC-SYNC-AUDIT-FALSIFICATION-001, SPEC-AUDIT-GATE-INTEGRITY-001]
depends_on: [SPEC-AUDIT-SNAPSHOT-001]
---

# SPEC-SYNC-PARALLEL-DOCS-001 — Sync-phase parallelization (A5/A7/A9/A6)

## HISTORY

- 2026-08-07 — Initial draft. Codifies the REMAINING §3.5 items (A5, A7, A9, A6) of the autonomy-workflow redesign report (`.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 rows A5/A6/A7/A9). Sibling to `SPEC-AUDIT-SNAPSHOT-001` (completed; absorbed A1-A4). No prior art in the SPEC catalog; closest neighbors are `SPEC-AUDIT-SNAPSHOT-001` (shared snapshot infrastructure reused by A9) and `SPEC-SYNC-AUDIT-FALSIFICATION-001` (sync-auditor obligation surface — distinct from execution-order parallelization here).

## §A. User Story

**As a** MoAI maintainer driving multi-SPEC epics through plan→run→sync,
**I want** the sync phase to (1) draft documentation concurrently with the read-only quality audit, (2) scan MX tags concurrently with the audit and halt on P1/P2 violations before coverage runs, (3) substitute orchestrator re-execution with an attributable diff-check against manager-develop §E evidence, and (4) bound plan-auditor retry spawns by SPEC Tier,
**so that** sync-phase latency stops being dominated by serial doc-audit-MX-coverage chains, by a "30-min coverage then 1 missing @MX:ANCHOR aborts all" worst case, by duplicate test/lint re-execution the §E evidence already captured, and by a flat 3-iter plan-auditor tail on simple SPECs.

**Outcome hypotheses (from §3.5 design report):**

- **A5**: docs drafting overlaps the audit fan-out; critical-path loses 1 serial phase. Combined with the A4 snapshot de-dup, sync time roughly halves on 4-locale docs-site updates.
- **A7**: the MX scan overlaps the audit AND its P1/P2 gate fires before coverage execution — eliminates the coverage-then-abort worst case and removes a serial phase.
- **A9**: completion-phase verification time roughly halves — re-execution is replaced by attributable diff-check against the §E artifact + the shared snapshot (per `verification-claim-integrity.md` §2 attribution; NOT silent removal).
- **A6**: the 3-iter plan-auditor tail is reserved for Tier L SPECs; most (S/M) get a single spawn, removing the worst-case 3× cold-spawn cost on simple SPECs.

## §B. Scope

**In scope — exactly A5, A7, A9, A6 from §3.5:**

- **A5 (docs ∥ audit parallelization)** — `manager-docs` docs drafter fan-out launches CONCURRENTLY with the Phase 7-10 read-only audit. CHANGELOG/README/docs-site derive from SPEC + git diff (input-independent of the audit). gate-sync-2 merges the docs draft + audit verdict. Removes 1 phase from the critical path.
- **A7 (MX Tag early + parallel)** — Phase 9 MX Tag scan launches CONCURRENTLY with the Phase 7 audit fan-out (not serially after Phase 8). P1/P2 violations (missing `@MX:ANCHOR` on fan_in≥3, missing `@MX:WARN` on goroutine) abort BEFORE Phase 10 coverage runs.
- **A9 (§E + 7-batch integration)** — manager-develop §E evidence promoted to a formal attributable artifact (command + observed output + baseline). Orchestrator trust-but-verify switches from re-execution → attributable diff-check against §E + the shared diagnostic snapshot (from AUDIT-SNAPSHOT-001 A4). Per `verification-claim-integrity.md` §2 — NOT silent removal; attributable diff-check substitution.
- **A6 (plan-auditor retry Tier ceilings)** — `harness.yaml` plan-auditor `max_iterations` becomes Tier-aware: S=1, M=2, L=3. The 3-iter tail is reserved for Tier L; most S/M SPECs get a single spawn.

### Out of Scope — Redesign items beyond A5/A7/A9/A6

- A8 (per-edit hook integration), A10 (Stop-chain shell shortening), A11 (mode-aware hooks) — separate sibling epic SPECs per AUDIT-SNAPSHOT-001 §B.
- `MOAI_AUTONOMY_TIER` mode-token introduction — sibling SPEC `SPEC-STOPCHAIN-TRIM-001` (planned).
- Goal-evaluator HTML dashboard and `moai_goal_render` surface — epic-level P1 work.
- MCP tool surface (`moai_verify_snapshot`, `moai_goal_status`, etc.) — epic-level P1 work.

### Out of Scope — Audit semantics changes

- Changing WHAT the plan-auditor or sync-auditor evaluates (AC content, scoring rubric, severity definitions, 4-dimension weights). A5/A7/A9/A6 change WHEN and HOW OFTEN these run, plus HOW the orchestrator consumes their output, NOT WHAT they measure.
- Lowering the PASS thresholds (Tier S 0.75 / M 0.80 / L 0.85) or the skip-eligible alignment (A2, already shipped).
- Removing the cold sync-auditor agent — it remains the fallback path per AUDIT-SNAPSHOT-001 REQ-AUDIT-SNAPSHOT-003 (A3 binding promotion fallback).

### Out of Scope — Run-phase §E evidence schema redesign

- Redefining the E1-E8 §E item structure (`.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E). A9 promotes §E to a formal attributable artifact and binds the orchestrator's diff-check to it; the E1-E8 item content itself is unchanged.

### Out of Scope — Parallelization beyond the sync phase

- Plan-phase and run-phase parallelization (plan-auditor concurrency, manager-develop milestone fan-out). A5/A7 bind the sync-phase quality-verification chain only.

## §C. Requirements (GEARS)

### A5 — docs ∥ audit parallelization

#### REQ-SPD-001 — Docs drafter fan-out concurrent with Phase 7-10 audit

**Where** the sync scope spans several independent document families, **When** the orchestrator enters Phase 7 (Quality Verification), the orchestrator shall launch the docs drafter fan-out CONCURRENTLY with the Phase 7-10 read-only quality audit in the same turn, rather than serially after the audit completes.

The docs drafter fan-out reuses the existing `FO-SYNC-4` five-drafter structure (`.claude/skills/moai/workflows/sync/doc-execution.md` L126-140: D1 CHANGELOG / D2 README+docs-site / D3 project docs / D4 SPEC artifacts / D5 codemaps) — the drafter set is unchanged; only the SCHEDULING changes from serial-after-audit to concurrent-with-audit.

#### REQ-SPD-002 — Docs drafter input independence

**While** the docs drafter fan-out is running concurrently with the Phase 7-10 audit, the docs drafter input shall derive from SPEC artifacts + git diff + the divergence report from Phase 11 Step 1.5, NOT from the concurrent audit's quality report or verdict.

Input-independence is the precondition that makes concurrency safe: the docs draft does not depend on audit output, so the audit may finish in any order relative to the draft without blocking or invalidating either side. The input source already matches `manager-docs.md` L40-49 (source-analysis → architecture → content-generation reads the source tree + SPEC, not the audit report).

#### REQ-SPD-003 — gate-sync-2 merge with single-writer applier

**When** both the docs draft and the audit verdict are produced, the orchestrator shall merge them at `gate-sync-2` (HUMAN GATE 2: Documentation Scope, `.claude/skills/moai/workflows/sync/doc-execution.md` L87-96) by applying the existing single-writer applier pattern: `manager-docs` applies the five drafts sequentially as the sole write-capable agent, and the audit verdict is surfaced to the user at the same gate.

The concurrency guard (`[HARD]` no two write-capable agents run concurrently, `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution) is preserved: the concurrent fan-out is entirely read-only drafters + read-only auditors; the single-writer applier pass runs at gate-sync-2 after both fan-outs return.

### A7 — MX Tag early + parallel

#### REQ-SPD-004 — MX Tag scan concurrent with Phase 7 audit

**When** the orchestrator enters Phase 7 (Quality Verification), the orchestrator shall launch the Phase 9 MX Tag scan CONCURRENTLY with the Phase 7 audit fan-out in the same turn, rather than serially after Phase 8 (Security).

The MX scan reuses the existing `FO-SYNC-2` sharded-scan structure (`.claude/skills/moai/workflows/sync/quality-gates-quality.md` L186: one read-only `Agent()` per language/package shard, 3-5 concurrent per the Mode 4 ceiling) — the sharding is unchanged; only the SCHEDULING changes from serial-after-Phase-8 to concurrent-with-Phase-7.

#### REQ-SPD-005 — P1/P2 gate fires before coverage execution

**When** P1 (missing `@MX:ANCHOR` on fan_in≥3 exported function) or P2 (missing `@MX:WARN` on goroutine/async pattern) violations are detected by the concurrent MX scan, the orchestrator shall halt sync BEFORE Phase 10 (Coverage Analysis) executes, eliminating the "30-min coverage then 1 missing tag aborts all" worst case.

The P1/P2 blocking semantics (`.claude/skills/moai/workflows/sync/quality-gates-quality.md` L143-153) are unchanged — P1/P2 still block sync; A7 only changes the ORDERING so the block fires before coverage cost is paid. P3 (advisory) and P4 (advisory) findings remain non-blocking.

#### REQ-SPD-006 — MX scan input is git-diff modified-files set

**While** the MX scan is running concurrently with the audit, the MX scan input shall be the modified-files set (git diff since last sync), independent of the concurrent audit verdict and independent of the coverage measurement.

This input-independence is the same precondition as REQ-SPD-002: the MX scan reads `git diff` + source files, not audit output, so concurrency is safe.

### A9 — §E + 7-batch integration (attributable diff-check)

#### REQ-SPD-007 — §E evidence as formal attributable artifact

**Where** `manager-develop` reports run-phase completion, the `§E` self-verification matrix (`.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E, items E1-E8) shall be a formal attributable artifact per `.claude/rules/moai/core/verification-claim-integrity.md` §2: each §E item names the command that was run, the observed verbatim output, and the baseline-attribution (this run, this tree).

The §E item structure (E1 AC matrix / E2 cross-platform build / E3 coverage / E4 boundary grep / E5 lint / E6 push / E7 blocker / E8 RED output) is unchanged; A9 binds the attribution discipline (command + verbatim output + baseline) rather than adding new §E items.

#### REQ-SPD-008 — Trust-but-verify switches to attributable diff-check

**Where** the orchestrator trust-but-verify batch (`agent-common-protocol.md` § Parallel Execution; the canonical 7-command batch) previously re-executed §E-cited commands (test/lint/vet/cover), the orchestrator shall switch to an attributable diff-check: verify that the §E evidence is attributable to the current tree-state by matching (a) the shared diagnostic snapshot key (HEAD SHA, from AUDIT-SNAPSHOT-001 REQ-AUDIT-SNAPSHOT-004) to the current HEAD, AND (b) the §E-cited command to the snapshot's recorded command, AND (c) the §E-cited observed output to the snapshot's recorded output.

On all three matching, the orchestrator consumes the §E evidence as attributable verification (no re-execution). This substitution is the source of the A9 time saving: test/lint/vet/cover run once at §E (manager-develop) and the snapshot records them once; the orchestrator's verification batch reads the recorded result instead of re-running.

#### REQ-SPD-009 — Diff-check fallback to re-execution (not silent removal)

**When** any of the three diff-check matches fails — (a) snapshot key mismatch (HEAD SHA changed since §E was recorded), OR (b) §E-cited command does not match the snapshot's recorded command, OR (c) §E evidence is missing or cites no observable output — the orchestrator shall fall back to re-execution of the affected verification dimension, preserving the `verification-claim-integrity.md` §1.1 invariant (no unobserved-claim).

A9 is NOT silent removal of verification: the diff-check substitutes for re-execution only when the attribution chain is intact; on any break, re-execution is restored. This requirement is the safety boundary that distinguishes A9 from a verification bypass.

### A6 — plan-auditor retry Tier ceilings

#### REQ-SPD-010 — Tier-aware plan-auditor max_iterations in harness.yaml

**Where** `harness.yaml` configures the plan-auditor retry contract (`.moai/config/sections/harness.yaml` `levels.{minimal,standard,thorough}.plan_audit.max_iterations`), the effective retry ceiling SHALL be Tier-dependent: Tier S = 1, Tier M = 2, Tier L = 3, overriding the flat per-level `max_iterations` for any SPEC whose `tier:` frontmatter field is set.

The Tier-aware ceiling is resolved by consulting the SPEC's `tier:` frontmatter field (S/M/L per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields). Where `tier:` is absent (treated as Tier L for backward compat per § SPEC Complexity Tier), the ceiling remains 3 — preserving the current behavior for pre-LEAN SPECs.

#### REQ-SPD-011 — plan-auditor Retry Loop Contract consults Tier ceiling

**Where** the plan-auditor Retry Loop Contract (`.claude/agents/moai/plan-auditor.md` § Retry Loop Contract, L386-418) bounds the iteration count, the contract SHALL consult the Tier-aware ceiling from REQ-SPD-010 in place of the flat `max_iterations: 3` constant.

On ceiling exhaustion, the orchestrator escalation path is unchanged: AskUserQuestion with three options (continue with current approach / revise SPEC / try alternative approach), per the existing spec-workflow § Re-planning Gate. A6 only changes the CEILING value, not the escalation semantics.

### Cross-cutting

#### REQ-SPD-012 — Write-capable concurrency guard preserved

**While** the A5 docs drafter fan-out and the A7 MX scan run concurrently with the Phase 7 audit, all concurrent agents SHALL be read-only (drafters return draft text; MX shards return findings; auditors read tree state). The single write-capable pass (`manager-docs` applying the docs drafts at gate-sync-2) runs after both fan-outs return, preserving the `[HARD]` rule that no two write-capable agents run concurrently.

## §D. Constraints

1. **Verification-claim-integrity invariant is inviolable** (from `verification-claim-integrity.md` §1.1 + §2): A9's diff-check substitution binds the orchestrator to the attribution chain; on any mismatch, re-execution is restored (REQ-SPD-009). Silent removal of verification is prohibited.
2. **Concurrency guard is inviolable** (from `agent-common-protocol.md` § Background Agent Execution): A5/A7 parallelism is bought entirely by making the concurrent agents read-only; the single-writer applier pass at gate-sync-2 is unchanged. Two write-capable agents NEVER run concurrently.
3. **A5/A7 input-independence is the concurrency precondition**: docs drafters derive input from SPEC + git diff (REQ-SPD-002); MX scan derives input from git diff (REQ-SPD-006). Neither reads the concurrent audit's output. Violating input-independence creates a hidden serial dependency and breaks the concurrency contract.
4. **Audit semantics unchanged**: A5/A7/A9/A6 change WHEN/HOW OFTEN audits run and HOW the orchestrator consumes their output, NOT WHAT they measure. PASS thresholds (S 0.75 / M 0.80 / L 0.85), 4-dimension weights, severity definitions, and AC content are immutable.
5. **A9 builds on the A4 shared snapshot** (AUDIT-SNAPSHOT-001 REQ-AUDIT-SNAPSHOT-004): the attributable diff-check consults the snapshot keyed by HEAD SHA. A9 does NOT invent a parallel snapshot store.
6. **Backward compatibility**: minimal-harness users (`levels.minimal.evaluator: false`) and pre-A6 SPECs (no `tier:` field → treated as Tier L → ceiling 3) MUST keep working. A6 is additive (Tier-aware ceiling) where `tier:` is set; legacy fallback is preserved.
7. **No new user-facing CLI surface in this SPEC**: the diff-check and the Tier-ceiling resolution are orchestrator-internal; MCP/CLI exposure is deferred to epic-level P1 work.

## §E. Assumptions

1. The `AUDIT-SNAPSHOT-001` shared diagnostic snapshot infrastructure (`moai verify check --key-current`, keyed by HEAD SHA, per `quality-gates-quality.md` Step 0.5.2) is reachable from the orchestrator trust-but-verify batch. A9 wires the orchestrator batch to consume this snapshot; it does not modify the snapshot store.
2. The `manager-docs.md` source-analysis workflow (L40-49) already reads the source tree + SPEC artifacts, not the audit report — so input-independence (REQ-SPD-002) is a codification of existing behavior, not a new constraint.
3. The `FO-SYNC-4` five-drafter structure and the `FO-SYNC-2` sharded-scan structure already exist; A5/A7 change their SCHEDULING (serial-after → concurrent-with), not their internal structure.
4. The `tier:` frontmatter field is already defined (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields); A6 consults it where present and falls back to 3 (Tier L) where absent.
5. The `gate-sync-2` HUMAN GATE 2 (Documentation Scope, `doc-execution.md` L87-96) is the natural merge point: it already requires user review of the divergence report before doc regeneration — adding the concurrent audit verdict to the same gate adds no extra human round-trip.

## §F. Open Questions (for plan-auditor)

- **OQ-1** (open): Does the orchestrator trust-but-verify batch today carry a structured "I am about to re-run command X" preamble that the A9 diff-check can intercept, or does the re-execution happen implicitly as a Bash call? (Determines whether A9 wires a literal diff-check hook at the batch entry, or whether it is a doctrinal switch the orchestrator applies when composing the batch.)
- **OQ-2** (open): Should the A6 Tier-aware ceiling live in `harness.yaml` (a new per-Tier `max_iterations` map alongside the per-level one) or in the plan-auditor agent body (a Tier→ceiling table the agent consults at spawn)? The former keeps config authority in `harness.yaml`; the latter keeps it co-located with the Retry Loop Contract. Plan-auditor's preference is solicited.

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 rows A5/A6/A7/A9, §1.3 (sync-phase bottlenecks).
- Sibling SPEC (completed): `.moai/specs/SPEC-AUDIT-SNAPSHOT-001/` — A1-A4; A4 shared snapshot is the infrastructure A9 builds on.
- Sync skill router: `.claude/skills/moai/workflows/sync.md` (FO-SYNC-1 4-dim binding, Phase Routing Table L47-53, HUMAN GATE Map L73-80).
- Quality-verification sub-skill: `.claude/skills/moai/workflows/sync/quality-gates-quality.md` (Step 0.5.2 snapshot consumption L41-43; Phase 9 MX validation L139-244; FO-SYNC-2 sharding L186; Phase 10 coverage L246-303).
- Doc-execution sub-skill: `.claude/skills/moai/workflows/sync/doc-execution.md` (Phase 11 divergence analysis L54-83; gate-sync-2 L87-96; FO-SYNC-4 five-drafter L126-140).
- manager-docs agent: `.claude/agents/moai/manager-docs.md` L40-49 (source-analysis → architecture → content-generation).
- manager-develop §E template: `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E (E1-E8).
- plan-auditor Retry Loop Contract: `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract L386-418.
- harness.yaml: `.moai/config/sections/harness.yaml` `levels.{minimal,standard,thorough}.plan_audit.max_iterations` L71-105.
- Verification-claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (no unobserved-claim) + §2 (baseline-attribution).
- Concurrency guard: `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution.
- Tier schema: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields (`tier: S|M|L`).

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-SPD-001 (A5 concurrent launch): docs drafter fan-out launches in the same turn as Phase 7 audit (not serially after).
- AC-SPD-002 (A5 input independence): docs drafter reads SPEC + git diff, NOT the concurrent audit result.
- AC-SPD-003 (A5 gate-sync-2 merge): manager-docs applies drafts as single writer at gate-sync-2; no concurrent write-capable agents.
- AC-SPD-004 (A7 MX concurrent): MX Tag scan launches concurrently with Phase 7 audit (same turn).
- AC-SPD-005 (A7 P1/P2 pre-coverage): P1/P2 violations halt sync before Phase 10 coverage executes.
- AC-SPD-006 (A7 no false abort): no P1/P2 violations → Phase 10 coverage proceeds (no regression).
- AC-SPD-007 (A9 §E attributable): §E evidence carries command + observed output + baseline per VCI §2.
- AC-SPD-008 (A9 diff-check): all-three matches → orchestrator consumes attributable evidence, no re-execution.
- AC-SPD-009 (A9 fallback): any match fails → orchestrator falls back to re-execution (not silent skip).
- AC-SPD-010 (A6 Tier S ceiling = 1): Tier S SPEC → plan-auditor retry ceiling 1.
- AC-SPD-011 (A6 Tier M ceiling = 2): Tier M SPEC → plan-auditor retry ceiling 2.
- AC-SPD-012 (A6 Tier L ceiling = 3): Tier L SPEC → plan-auditor retry ceiling 3 (legacy fallback).
- AC-SPD-013 (concurrency guard): A5/A7 concurrent agents are all read-only; single-writer applier at gate-sync-2.
- AC-SPD-014 (audit semantics unchanged): A5/A7/A9/A6 change WHEN/HOW OFTEN, not WHAT — PASS thresholds + 4-dim weights + severity immutable.
