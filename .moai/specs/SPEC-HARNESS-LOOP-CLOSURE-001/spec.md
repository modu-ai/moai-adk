---
id: SPEC-HARNESS-LOOP-CLOSURE-001
title: "First Clean Human-Gated Harness Apply + Auditable Lineage"
version: "0.1.0"
status: implemented
created: 2026-06-14
updated: 2026-06-14
author: GOOS
priority: P1
phase: "v0.2.0"
module: "internal/harness"
lifecycle: spec-anchored
tags: "harness, lineage, apply, rollback, audit, proof-of-mechanism"
tier: S
---

# SPEC-HARNESS-LOOP-CLOSURE-001 — First Clean Human-Gated Harness Apply + Auditable Lineage

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | GOOS | Initial draft — plan-phase artifacts (Tier S). Port arXiv 2606.09498v1 M6 (Auditable Lineage Logging) + M7 (loop skeleton, adapted to MoAI human gate). Prove the harness apply/lineage/rollback mechanism end-to-end with ONE clean human-gated apply. |

## A. Context (Why)

The moai-adk harness learning subsystem (`internal/harness/`) is **fully wired but has never closed its loop**. Verified runtime artifacts (live tree, 2026-06-14):

- `.moai/harness/observations.yaml` — 258 observations recorded
- `.moai/harness/usage-log.jsonl` — 536 events captured
- `.moai/harness/learning-history/tier-promotions.jsonl` — 8 tier registrations (every entry has `from_tier:""`, i.e. all bootstrap promotions)
- ZERO applies have ever executed: `.moai/harness/snapshots/` directory is **absent**, `.moai/harness/seeds/` is empty (`.gitkeep` only)

The full pipeline exists and is reachable:
`Observe (observer.go:53) → Aggregate (learner.go:46) → ClassifyTier (learner.go:113) → Promote (learner.go:142 WritePromotion) → MapPromotions (proposalgen/mapper.go:52) → 5-Layer Safety (safety/pipeline.go:89) → Apply (applier.go:176) / Rollback (applier.go:272 RestoreSnapshot)`.

But because no apply has ever run:

- The snapshot/rollback path (`applier.go:218 createSnapshot` + `applier.go:272 RestoreSnapshot`) has **never executed against a real proposal flow**.
- There is **no per-transition manifest** recording which surface was changed, what the accept/reject decision was, and why. `tier-promotions.jsonl` records tier promotions, NOT apply transitions. When a proposal is rejected (e.g. by L1 Frozen Guard or L5 human gate), nothing durable records that the candidate was seen-and-rejected.

Every downstream harness improvement (failure clustering, held-out gates, scorer-loop wiring — all deferred to a future Phase 5 SPEC) would be built on this **unproven foundation**. This SPEC's sole job is to **prove the mechanism**: make the Go apply path complete + testable so that ONE clean, human-gated apply can flow end-to-end, and make every transition (accept AND reject) **auditable and reversible**.

This SPEC deliberately ships almost no "intelligence". It ships **proof-of-mechanism**.

### A.1 Paper basis (context only)

This SPEC ports exactly two mechanisms from arXiv 2606.09498v1 "Self-Harness":

- **M6 — Auditable Lineage Logging**: every transition is logged (changed surface, decision, accept/reject, reason). Per the paper: *"rejected candidates remain logged but do not change the active harness."*
- **M7 — Loop skeleton**, ADAPTED to moai-adk's constitutional human gate. moai-adk does NOT adopt the paper's autonomous closure; the L5 human gate (`auto_apply: false`, oversight.go) stays mandatory.

### A.2 Prior-art harness SPECs (continuity — do not duplicate/contradict)

This SPEC closes the loop that the following SPECs built. It does NOT re-implement any of them:

- `SPEC-V3R5-HARNESS-AUTONOMY-001` — observer/learner autonomy foundation
- `SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001` — ClassifyTier wiring
- `SPEC-V3R6-HARNESS-PROPOSAL-GEN-001` — proposalgen/mapper (Promotion → Proposal)
- `SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001` — activation/apply path wiring

The 5-Layer safety pipeline, the `Apply` entry point, the `createSnapshot`/`RestoreSnapshot` pair, and the `OversightProposal` boundary type are all complete and correct from these SPECs. This SPEC ADDS only the per-transition lineage manifest and ensures the apply/snapshot/rollback path is proven by tests.

## B. Requirements (GEARS)

### B.1 Lineage manifest — append-only per-transition record (M6)

- **REQ-HLC-001** (Ubiquitous): The harness shall maintain an append-only lineage manifest at `.moai/harness/learning-history/manifest.jsonl`, with exactly one JSON entry appended per apply transition (one entry on accept, one entry on reject).

- **REQ-HLC-002** (Ubiquitous): A lineage entry shall carry, at minimum, the fields `proposal_id`, `target_path`, `applied_surface`, `decision` (one of `"approved"` | `"rejected"`), `timestamp`, and `reason`. The Go representation shall be a `LineageEntry` struct whose optional fields use `omitempty`, so existing readers and future schema additions remain backward-compatible.

- **REQ-HLC-003** (Event-driven): **When** `Apply()` reaches the actual file-modification step (i.e. the safety pipeline returned `DecisionApproved` and the snapshot was created successfully), the harness shall append a lineage entry with `decision: "approved"`, `applied_surface` set to the modified frontmatter field key (e.g. `description`), and `reason` describing the approved transition.

- **REQ-HLC-004** (Event-driven): **When** the safety pipeline rejects a proposal (`DecisionRejected` — e.g. the L1 Frozen Guard denies a non-`my-harness-*` target, or any L2–L4 layer rejects), the harness shall append a lineage entry with `decision: "rejected"` and a `reason` derived from the rejecting layer, **and the active harness shall be left unchanged** (no SKILL.md write, no snapshot directory created for the rejected proposal).

- **REQ-HLC-005** (State-driven): **While** the safety pipeline returns `DecisionPendingApproval` (the `auto_apply: false` default → L5 human gate), `Apply()` shall return an `ApplyPendingError` carrying the `OversightProposal` payload and shall NOT itself write a lineage entry for the pending state (the pending state is not a transition; lineage is written only on the subsequent approved/rejected resolution).

### B.2 First clean apply — Go path completeness (M7 adapted)

- **REQ-HLC-006** (State-driven): **While** `auto_apply` is `false` (the shipped default at `.moai/config/sections/harness.yaml`), a proposal that passes L1–L4 shall NOT be applied directly; the harness shall surface the `OversightProposal` to the orchestrator via `ApplyPendingError` (the orchestrator, NOT the harness, presents it via `AskUserQuestion`).

- **REQ-HLC-007** (Event-driven): **When** the orchestrator-approved proposal is re-submitted for application (the path that reaches `DecisionApproved`), `Apply()` shall write ONLY to the proposal's target `description` (or `triggers`) frontmatter field of a `.claude/skills/my-harness-*/SKILL.md` file, preserving all other frontmatter fields and the body byte-for-byte.

- **REQ-HLC-008** (Capability gate): **Where** a proposal's `target_path` is outside the `my-harness-*` writable namespace (e.g. `.claude/agents/moai/...`, `.claude/skills/moai-...`, `.claude/rules/moai/...`, `.moai/project/brand/...`), the L1 Frozen Guard shall reject the proposal (`DecisionRejected`, `RejectedBy: 1`) and the harness shall write a `decision: "rejected"` lineage entry without modifying any file.

### B.3 Snapshot / rollback verification (M7 reversibility)

- **REQ-HLC-009** (Event-driven): **When** `Apply()` writes to a SKILL.md for the first time, the snapshots base directory shall be created (it is currently absent) and a byte-for-byte backup plus a `manifest.json` shall be written BEFORE the SKILL.md modification.

- **REQ-HLC-010** (Event-driven): **When** `RestoreSnapshot()` is invoked against a snapshot directory created by a prior apply, the target SKILL.md shall be restored byte-identically to its pre-apply content.

### B.4 Backward compatibility + boundary

- **REQ-HLC-011** (Ubiquitous): The existing runtime artifacts `usage-log.jsonl` and `tier-promotions.jsonl` shall remain untouched by this SPEC's changes; the lineage `manifest.jsonl` is a NEW, purely additive file and old data shall continue to load.

- **REQ-HLC-012** (Ubiquitous): No source file under `internal/harness/` (outside `_test.go`) shall contain an `AskUserQuestion` or `mcp__askuser` call-site reference; the C-HRA-008 subagent-boundary guard (`internal/harness/subagent_boundary_test.go`) shall stay green.

## C. HARD Doctrine Constraints

Every acceptance criterion respects these constraints (mirrored into `acceptance.md`):

- **C1 — Human gate mandatory**: L5 human gate stays mandatory; `auto_apply` default remains `false`. No autonomous apply. (REQ-HLC-005/006)
- **C2 — FROZEN zone untouched**: No writes to `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, `.moai/project/brand/`. The L1 Frozen Guard (`safety/frozen_guard.go:18-23` `frozenPrefixes` + `IsFrozen:35`, hardcoded prefixes — the bare `internal/harness/frozen_guard.go` holds the SEPARATE `allowedPrefixes`/`IsAllowedPath`, not the wired L1 deny list) MUST reject any non-`my-harness-*` target. (REQ-HLC-008)
- **C3 — Namespace**: Only the `my-harness-*` namespace is writable. (REQ-HLC-007)
- **C4 — Template neutrality**: All new Go types/logic live in `internal/harness/`; nothing leaks into `internal/template/templates/**`; no SPEC IDs / dates / SHAs embedded in shipped code or template. (CLAUDE.local.md §15 / §25)
- **C5 — Subagent boundary**: No `AskUserQuestion` in `internal/harness/`; C-HRA-008 sentinel stays green. (REQ-HLC-012)
- **C6 — No FROZEN-constant changes**: tier thresholds `{1,3,5,10}` (`tier/tier.go:49`), the 4 score dimensions `{Functionality, Security, Craft, Consistency}` (`scorer.go:6-24`), rubric anchors `{0.25, 0.50, 0.75, 1.00}` (`rubric.go:17-28`), and the Security must-pass floor are ALL untouched.
- **C7 — Production no-oracle honesty**: This SPEC adds NO quality verdict / gate (that is the deferred M2-lite SPEC). It only proves apply + lineage + rollback.

## D. Acceptance Criteria

Enumerated in `acceptance.md`. AC count: 8 (AC-HLC-001 .. AC-HLC-008).

## Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly OUT OF SCOPE for this SPEC and are recorded as known-deferred follow-up candidates (a future Phase 5 SPEC):

- **Held-out split validation** — No held-out evaluation set, no train/validate split. Deferred.
- **Failure-signature clustering** — No failure clustering, no signature extraction. Deferred.
- **LLM proposer / K-candidate diversity** — No LLM-generated proposals, no K-candidate generation or diversity scoring. The proposal flow reuses the existing deterministic `proposalgen/mapper.go` path verbatim. Deferred.
- **Scorer-loop wiring** — The `scorer.go` 4-dimension scorer is NOT wired into the apply loop. This SPEC adds NO quality verdict / gate. Deferred (the "M2-lite" SPEC).
- **M5 merge policy** — No merge-policy mechanism from the paper. Deferred.
- **Flipping `auto_apply` to `true`** — `.moai/config/sections/harness.yaml` `auto_apply: false` stays unchanged (the human gate is the point). This SPEC does NOT change the default.
- **Changes to the 5-Layer safety pipeline logic** — `safety/pipeline.go`, `safety/frozen_guard.go`, `safety/canary.go`, `safety/oversight.go`, the rate limiter, and the contradiction detector are complete and correct. This SPEC does NOT modify the safety layers; it only ADDS the lineage write at the `Apply()` accept and reject paths.
- **Changes to FROZEN constants** — tier thresholds, score dimensions, rubric anchors, and the Security must-pass floor are untouched (C6).
- **Authoring a real `my-harness-*` skill** — Creating/seeding a production `my-harness-*` SKILL.md so a live apply has a target is an OPERATIONAL step performed by the orchestrator AFTER this SPEC's Go path is complete; it is NOT part of this SPEC's code. The Go apply/lineage/rollback path is proven via test fixtures (`t.TempDir()` SKILL.md). The "drive ONE real live apply" act is the orchestrator's run-phase/post-merge operational follow-through, gated by the L5 human gate.

## Cross-References

- `internal/harness/applier.go` — `Apply`:176, `createSnapshot`:218, `RestoreSnapshot`:272, `ApplyPendingError`:133, `snapshotManifest`:146 (reject early-return at applier.go:186 currently discards `decision.Reason` — run-phase MUST capture it into the rejected lineage entry before returning) (EXTEND `Apply` only — add lineage writes at accept + reject paths; do NOT change snapshot logic)
- `internal/harness/types.go` — ADD `LineageEntry` struct (additive, `omitempty` fields)
- `internal/harness/lineage.go` — NEW file: `WriteLineageEntry()` + `LoadManifest()` (mirror the `learner.go:145 WritePromotion` append-JSONL idiom)
- `internal/harness/lineage_test.go` — NEW file: write / load / append-only / rollback-byte-identical tests
- `internal/harness/learner.go` — `WritePromotion`:145 (reference idiom for the new lineage writer; DO NOT MODIFY)
- `internal/harness/safety/pipeline.go` — `Evaluate`:89 (Decision producer; DO NOT MODIFY)
- `internal/harness/safety/frozen_guard.go` — `IsFrozen`:35, `frozenPrefixes`:18-23, `LogViolation`:83 (the WIRED L1 guard; DO NOT MODIFY). NOTE: distinct from the bare `internal/harness/frozen_guard.go` (`allowedPrefixes`:18, `IsAllowedPath`:60, `EnsureAllowed`:103) — do not conflate the two files.
- `internal/harness/subagent_boundary_test.go` — C-HRA-008 binary CI guard (MUST stay green)
- `.moai/config/sections/harness.yaml` — `auto_apply: false`:116 (stays — out-of-scope to flip)
- CLAUDE.local.md §15 (template language neutrality), §25 (template internal-content isolation)
