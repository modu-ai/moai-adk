---
id: SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001
title: "V3R4 Harness Classifier Runtime Wiring"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Evolution Loop Closure"
module: ".claude/skills/moai/workflows/harness.md, internal/cli/hook.go (Option A new subcommand) or internal/hook/post_tool.go (Option B)"
lifecycle: spec-anchored
tags: "harness, classifier, wiring, runtime, v3r6, tier-s-minimal"
depends_on: [SPEC-V3R4-HARNESS-001, SPEC-V3R4-HARNESS-002, SPEC-V3R4-HARNESS-003]
breaking: false
bc_id: []
related_theme: "Self-Evolving Harness v2 — Runtime Loop Closure"
target_release: v3.0.0
---

# SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 — V3R4 Harness Classifier Runtime Wiring

## §A. Background (Brain Phase 1 Discovery Findings)

V3R4 self-evolving harness classifier wiring gap analysis (verbatim from Brain Phase 1):

**V3R4 SPEC Lineage Status**:
- `SPEC-V3R4-HARNESS-001` (foundation): status `completed`. CLI retired per `BC-V3R4-HARNESS-001-CLI-RETIREMENT`. Workflow body owns lifecycle file-IO only.
- `SPEC-V3R4-HARNESS-002` (Multi-Event Observer): status `completed`. PostToolUse + Stop + SubagentStop + UserPromptSubmit hooks shipped.
- `SPEC-V3R4-HARNESS-003` (Embedding-Cluster Classifier): status `completed`. Stage 1 exact-match + Stage 2 SimHash 64-bit clustering implemented.
- `SPEC-V3R6-HARNESS-LEARNER-FIX-001`: status `implemented`. Subagent boundary fix.

**Classifier Code Complete (Verified Present)**:
- `internal/harness/learner.go`: exports `AggregatePatterns`, `ClassifyTier`, `WritePromotion`
- `internal/harness/classifier_simhash.go` (NEW from V3R4-003)
- `internal/harness/classifier_cluster.go` (NEW from V3R4-003)

**Wiring Gap (Root Cause)**:
The classifier caller exists ONLY at `internal/cli/harness.go:142, 157`, but this file is the V3R4 deprecation marker NOT registered in the cobra tree (per `BC-V3R4-HARNESS-001-CLI-RETIREMENT`). The workflow body `.claude/skills/moai/workflows/harness.md` is "file-IO only" by contract and never invokes the Go classifier.

**Observable Evidence (on this project)**:
- 97 usage-log events in `.moai/harness/usage-log.jsonl`
- 4 unique patterns identifiable via classifier aggregation
- 0 entries in `.moai/harness/learning-history/tier-promotions.jsonl` (file may not even exist)

Net effect: the entire learning loop is dark on this project despite all classifier code being shipped and all observer events being captured. The Brain Phase 1 audit (2026-05-24) identified the missing wiring as the highest-leverage Sprint 8 candidate.

## §B. Goal

Wire the workflow body §2.1 `status` verb to invoke the Go classifier (`AggregatePatterns` + `ClassifyTier` + `WritePromotion`) so that `learning-history/tier-promotions.jsonl` auto-generates on each status invocation. This closes the V3R4 learning loop without restoring the retired `moai harness` cobra subcommand (preserves `BC-V3R4-HARNESS-001-CLI-RETIREMENT` contract).

The change is a **single mechanical wiring insertion**: workflow body §2.1 calls a new hook subcommand (Option A recommended in plan.md §3) that invokes the existing Go classifier API and writes promotions. Status output rendering then includes the tier distribution table derived from `tier-promotions.jsonl`.

## §C. Scope

- **Files edited**: 1-2 (`.claude/skills/moai/workflows/harness.md` body + either `internal/cli/hook.go` for Option A new subcommand OR `internal/hook/post_tool.go` for Option B in-process batching)
- **LOC envelope**: ≤300 LOC total (workflow body Bash insertion ~20 lines + hook subcommand ~80-150 Go LOC)
- **Milestones**: 1 (M1 only — Tier S minimal)
- **Cobra tree changes**: zero new public top-level subcommand (Option A adds `moai hook harness-classify` under the existing `hook` namespace, which is orthogonal to the retired `harness` namespace)
- **Backward compatibility**: non-breaking. No public API surface change. No SPEC frontmatter schema change.

## §D. EARS Requirements

| ID | Pattern | Statement |
|----|---------|-----------|
| **REQ-HCW-001** | Ubiquitous | The workflow body §2.1 `status` verb SHALL invoke the Go classifier (`AggregatePatterns` + `ClassifyTier` + `WritePromotion`) before rendering the tier distribution table. |
| **REQ-HCW-002** | Event-driven | WHEN the `status` verb is invoked, the classifier SHALL read `.moai/harness/usage-log.jsonl`, aggregate patterns, classify tiers, and write promotions to `.moai/harness/learning-history/tier-promotions.jsonl`. |
| **REQ-HCW-003** | Event-driven (fail-safe) | WHEN the classifier returns an error, the `status` verb SHALL render an error annotation in the status output but continue rendering remaining sections (fail-open; do NOT abort the status command). |
| **REQ-HCW-004** | Ubiquitous (config-gated) | WHEN `learning.enabled` is `false` in `.moai/harness/harness.yaml`, the classifier invocation SHALL be a no-op (preserves `REQ-HRN-FND-009` contract from V3R4-HARNESS-001). |
| **REQ-HCW-005** | Ubiquitous (Optional, MAY) | The classifier invocation cadence MAY be a single batch on each `status` call (no streaming/incremental) for the V3R6 baseline. Future SPECs MAY add `PostToolUse` incremental invocation. |

Notes:
- REQ-HCW-005 is **Optional MAY** without AC coverage per spec.md §C decision rule (cadence policy descriptive, future SPEC scope).
- REQ-HCW-001 through REQ-HCW-004 are **mandatory** with full AC coverage in `acceptance.md`.

## §E. Exclusions

### Out of Scope (Explicit Deferral List)

The following items are explicitly OUT of scope for this SPEC and tracked as candidates for separate SPECs:

- **Proposal JSON generator** (`Pattern → Tier-4 proposal artifact`) — Generator layer is downstream of classifier wiring; would inflate this SPEC envelope. Successor: `SPEC-V3R6-HARNESS-PROPOSAL-GEN-001` (planned).
- **Rate Limiter wiring** (5-Layer L4) — Safety layer; classifier wiring must precede so the rate limiter has signals to limit. Successor: `SPEC-V3R6-HARNESS-RATELIMIT-001` (planned).
- **Snapshot wiring** (5-Layer L5) — Persistence layer; depends on stable promotion stream. Successor: `SPEC-V3R6-HARNESS-SNAPSHOT-001` (planned).
- **`.claude/agents/harness/` 4 specialist agent namespace cleanup** — Already chore-removed `4f1135684`; further consolidation per `CLAUDE.local.md §24` namespace policy. Successor: `SPEC-V3R6-HARNESS-AGENTS-CLEANUP-001` (planned).
- **Layer 2 Canary activation** — Deferred per `V3R4-001 §1.3`; requires production traffic baseline. No successor candidate yet.
- **Layer 3 Contradiction Detector activation** — Deferred per `V3R4-001 §1.3`; requires Layer 2 baseline. No successor candidate yet.
- **PostToolUse incremental classifier invocation** — Future cadence enhancement; REQ-HCW-005 declares this MAY scope. Successor: `SPEC-V3R6-HARNESS-INCREMENTAL-001` (planned).

## §F. Decision Rule (Optional MAY without AC coverage)

Per the canonical decision rule (see `acceptance.md` §F-AC):
- All **mandatory** REQs (REQ-HCW-001..004) MUST have direct AC coverage in `acceptance.md`.
- **Optional MAY** REQs (REQ-HCW-005) MAY exist without dedicated AC. Their absence from AC is itself the contract — future SPECs may add ACs when the MAY clause is exercised.

## §G. Wiring Mechanism Decision (Deferred to plan.md)

This `spec.md` describes the **capability** only ("MUST invoke Go classifier"). The HOW — which of Options A/B/C from the Wiring Branch Trade-off Matrix — is documented in `plan.md §3`. The recommended option (A: `moai hook harness-classify`) is documented as a recommendation, not a lock-in. The run-phase manager-strategy makes the final binding decision based on plan-auditor verdict and any new evidence surfaced at run time.

## §H. HISTORY

- 2026-05-24 — manager-spec — Initial draft. Tier S minimal Section A-E template. EARS REQ-HCW-001..005 (4 mandatory + 1 Optional MAY). Scope: 1-2 files, ≤300 LOC, 1 milestone. Wiring mechanism decision deferred to plan.md §3 (3 options A/B/C with Option A recommended). Frontmatter 12-field canonical schema. Depends on V3R4-HARNESS-001/002/003 (all completed).
