---
id: SPEC-HARNESS-EVOLVE-003
title: "Curator production wiring — Tier-Surface mapping + validation gates + re-proposal suppression"
version: "0.1.1"
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

# SPEC-HARNESS-EVOLVE-003 — Curator Production Wiring (Loop 2 activation)

## HISTORY

| Date | Version | Change | Author |
|--------|---------|--------|--------|
| 2026-07-12 | 0.1.0 | Initial plan-phase draft (Tier L, 5 content artifacts: spec.md + plan.md + acceptance.md + design.md + research.md, plus progress.md §E skeleton). M3 of the HARNESS-EVOLVE Epic per the approved design SSOT `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (§4 3-Zone edit-surface contract — Evolvable-zone A6 row; §5 Loop 2 gate chain L2→L3→L4→L5; §7 M3 milestone + verification matrix; §1 A1/A6/A7 deltas). Activates the PRODUCTION wiring that EVOLVE-002 deliberately left INERT (the write-layer API surface exists and is unit-tested against machine-signal inputs, but no live Curator write path drives it — verified: `grep -rn TierGatedWrite\|WriteManagedBlockGated internal/ cmd/ pkg/ | grep -v _test.go | grep -v curator/` returns empty). 7 pillars sourced from SSOT §1 (A1, A6, A7) + EVOLVE-002 spec.md §Out of Scope (Tier↔surface mapping, Gate activation, Re-proposal suppression). 4 NEEDS-CLARIFICATION items tracked in plan.md §H. | manager-spec |
| 2026-07-12 | 0.1.1 | plan-audit amendments D1-D3 + G1/G2 (PASS-WITH-DEBT 0.89 → re-audit ready). **D1** (sync-phase trap): AC-HEV3-023a compound-grep `≥3` trap fixed to per-token `≥1` × 2 tokens — design.md §C.2 decides the hooks axis is covered by the `.claude/settings.json` prefix match, so only 2 entries are added (not 3); the old compound count would mechanically max at 2 and never reach ≥3. **D2** (verification-claim-integrity §1.1 surface 3): research.md §C.2 + design.md §C.2 frozenPrefixes list corrected from 4 entries to the actual 3 (`sed -n '18,25p' internal/harness/safety/frozen_guard.go` — `.claude/skills/moai/` is ABSENT; it lives only in the meta-harness `internal/harness/frozen_guard.go` 4-entry list — the two files must not be conflated). **D3** (double-write hazard): REQ-HEV3-008/009 amended with a sequencing clause — TierGatedWrite (approval path, sole write via WriteManagedBlock) and WriteManagedBlockGated (rejection path, RejectionRecorder) are called on DISJOINT branches, never both on the same branch; design.md §A.1 call chain step 6→7a/7b updated to reflect the branch-split. **G1**: AC-HEV3-005 compound-grep `≥3 (one per token)` → per-token verification (3 separate greps, each ≥1) so a single token repeated 3× cannot mechanically satisfy it. **G2**: research.md §B.1 "20 files" → "18 Go source files" (`ls -1 internal/harness/curator/*.go \| wc -l` = 18 — the "20" counted `.`/`..`). spec-lint stays 0 findings (normal + strict). | manager-spec |

## §A. Context and Intent

`SPEC-HARNESS-EVOLVE-002` (completed, sync `29a6f53d0` + backfill `fa2a3086a`) built
the **write layer** of the self-evolving harness: the typed managed-block writer
package `internal/harness/curator/` (`WriteManagedBlock`, `TierGatedWrite`,
`WriteManagedBlockGated`, per-bullet CRUD, `AppendLearnedLocal`, budget/cap
enforcement, anti-fabrication input validation, `ApprovalDecision` +
`RejectionRecorder` L5 contract). The API surface is complete and unit-tested
against machine-signal inputs. **But it is INERT** — no production code path
drives it. The tier↔surface dispatch does not read harness config; the L5
approval round does not inject `ApprovalDecision` + `RejectionRecorder`; the
5-layer safety pipeline does not wrap the Curator write path; the
negative-evidence registry does not exist. Design gaps G4 ("L2 Canary / L3
Contradiction no-op"), G7 ("기각·롤백 이력이 재제안을 막지 못함"), and the A1
permission-surface gap remain open (design SSOT §3).

This SPEC activates **Loop 2 production wiring** of the 3-Loop self-evolving
harness (design SSOT §5, grounded in Lilian Weng's "Harness Engineering for
Self-Improvement"). It wires the completed write-layer API into the live
Curator pipeline and activates the gates that guard it. It is the SPEC that
makes the harness *actually self-evolve* — EVOLVE-001 observed, EVOLVE-002
built the pen, this SPEC signs the writes.

**Boundary principle.** This SPEC authors ONLY the production wiring + gate
activation + the negative-evidence registry + the permission-surface Frozen
expansion:
- the tier→surface dispatch that reads harness config and calls
  `TierGatedWrite` (the completed API),
- the L5 approval round injection (`WriteManagedBlockGated` with a live
  `ApprovalDecision` + `RejectionRecorder`),
- the L2 Canary and L3 Contradiction gates wrapping the Curator write path
  (currently no-op at pipeline.go:70-73 for L3; not wired to the Curator path
  for L2),
- the negative-evidence registry (data structure + writer + cooldown logic +
  rollback auto-register),
- the permission-surface Frozen expansion (settings.json / permission mode /
  hook registration / frozen-guard itself),
- the GLM observe-only model gate.

It performs NO console-verb addition, NO Recall-wiring consumption, NO typed
harness-spec.yaml parser, NO new managed-block type — those are EVOLVE-004 /
EVOLVE-005 territory (see §E Exclusions).

Three anti-fabrication principles bind the design (inherited from EVOLVE-001 §A
+ EVOLVE-002 §A + `.claude/rules/moai/core/verification-claim-integrity.md`
§1.1):

1. **Machine signals only** — Curator write authority, gate triggers, and the
   negative-evidence registry entries derive from mechanical state (tier
   counts, gate results, rollback events, manual verbs), never from model
   self-report. The wiring is invoked by the existing Curator pipeline, not by
   free-form model prose.
2. **Privacy / template neutrality** — the auto_detection Tier-4 surface edits
   and the negative-evidence registry entries carry distilled *generic*
   workflow knowledge and machine-signal references; NEVER verbatim user text
   or internal SPEC IDs / REQ tokens / dates / commit SHAs. The mechanism ships
   to `internal/template/templates/` only where it is a neutral empty scaffold
   (the auto_detection schema validator); registry DATA never ships.
3. **Evidence-or-null** — negative-evidence registry entries carry a
   machine-signal reference (rollback SHA, gate-reject event, lineage key);
   where evidence is absent, the reference is `null`, never an inferred value.

**L5 approval invariant (activation).** Every CLAUDE.md / CLAUDE.local.md
Curator write is gated by L5 `AskUserQuestion` approval routed through the
orchestrator (no autonomous write path exists, no exception — design SSOT §5
Loop 2 + "인간은 목표와 기준 제공"). EVOLVE-002 authored the writer-side
contract (`WriteManagedBlockGated` + `ApprovalDecision`); this SPEC wires the
orchestrator-side round that feeds it.

## §B. Scope Summary

**In scope (7 pillars — sourced from SSOT §1 A1/A6/A7 + EVOLVE-002 §Out of Scope)**:

1. **Tier↔Surface mapping activation (design delta A6)** — register the
   `harness.yaml` `auto_detection` block (which ALREADY exists at
   `.moai/config/sections/harness.yaml:2` — verified) as a Tier 4 Evolvable-zone
   editable surface with value-range validation (threshold upper/lower bounds).
   Activate tier→surface routing: Tier 3 → CLAUDE.local.md, Tier 4 → CLAUDE.md.
   This is the activation that EVOLVE-002 spec.md `Out of Scope — Tier↔surface
   mapping activation` explicitly defers here.
2. **L5 orchestrator approval round integration** — every CLAUDE.md /
   CLAUDE.local.md Curator write routes through an AskUserQuestion approval
   round (the orchestrator is the sole user-channel per REQ-HEV2-032 /
   archived-agent-rejection.md §C). `ApprovalDecision` + `RejectionRecorder`
   feed back into the write pipeline. The Curator returns proposals; the
   orchestrator runs the round.
3. **EVOLVE-002 residual — PRODUCTION wiring of TierGatedWrite /
   WriteManagedBlockGated** — EVOLVE-002 progress.md M7 residual risk names
   these as the completed write-layer API exercised only by tests; wire them
   into the live Curator pipeline (baseline 0 production callers → ≥1).
4. **L2 Canary activation** — shadow-apply + doctor + regression gate. Held-out
   validation: a proposed change must not regress a held-out signal. Activation
   of the L2 evaluator for the Curator write path (the existing
   `safety.EvaluateCanary` + `CanaryVeto` machinery is wired to the harness
   applier path, not yet to the Curator Learned-surface write path).
5. **L3 Contradiction activation** — Frozen-rules contradiction check: a
   proposal that contradicts a registered Frozen rule is rejected. Activation
   replaces the explicit no-op at `safety/pipeline.go:70-73`
   (`l3ContradictionCheck` currently returns `ContradictionReport{}` always).
6. **Design delta A7 — Re-proposal suppression** — negative-evidence registry.
   The pattern key of a rejected OR rolled-back promotion is registered as
   negative evidence. Same-key re-proposal is blocked until N NEW evidences
   accumulate AND a cooldown elapses. Rollback auto-registers the negative
   entry. EVOLVE-002 authored the registry KEY/token only (REQ-HEV2-010
   `ledger_key`); this SPEC ships the registry itself.
7. **Design delta A1 — Permission surface Frozen registration** — extend the
   `safety/frozen_guard.go` `frozenPrefixes` block-list (line 18 — verified)
   to cover the permission axis elevated to Frozen Zone in v5.1:
   settings.json / settings.local.json, permission mode, hook registration. A
   Curator proposal targeting these surfaces is L1-blocked. Include a GLM
   observe-only guard.

**Preserve**:
- `internal/harness/curator/` package (EVOLVE-002 output) — the typed writer
  API is NOT modified; this SPEC wires it into production. The writer is the
  inner write primitive; the gates + dispatch + registry are outer decorators
  (EVOLVE-002 design.md §F.1 forward-compatible contract).
- `internal/harness/routing/` (EVOLVE-001 output) — read-only consumption of
  the routing ledger by the negative-evidence registry's evidence-counter is
  fine; the ledger writer is untouched.
- `internal/harness/safety/pipeline.go` Pipeline.Evaluate L1→L2→L3→L4→L5 order
  — IMMUTABLE (pipeline.go:14-15 `[HARD] L1→L2→L3→L4→L5 order is immutable`).
  This SPEC replaces the L3 no-op function body and wires the Curator path
  through the existing pipeline; it does NOT reorder layers.
- `internal/harness/safety/canary_veto.go` CanaryVeto 48h cooldown — reused as
  the cooldown-enforcement primitive for A7 where the semantics overlap; the
  net-new A7 work is the negative-evidence registry data structure + the
  "N new evidences" accumulator.
- `internal/harness/learner.go` — the existing 4-tier learning ladder
  (`[1,3,5,10]`) and pattern aggregation are untouched; this SPEC consumes
  them read-only via the existing `tier.ClassifyStatus`.
- `internal/config/types.go AutoDetectionConfig` (line 740) — the existing
  config struct is EXTENDED additively with value-range validation, not
  rewritten.
- Template neutrality (CLAUDE.local.md §25): only the auto_detection schema
  validator mechanism ships to templates (a neutral schema, no learned DATA);
  registry entries and gate verdicts never ship.

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 Tier↔Surface mapping activation (design delta A6)

- **REQ-HEV3-001** (Ubiquitous — surface registration): The harness
  `auto_detection` block (at `.moai/config/sections/harness.yaml:2` — already
  present, verified) SHALL be registered as a Tier 4 Evolvable-zone editable
  surface in the Curator's surface registry, so the Curator MAY propose
  bounded edits to the complexity-estimator keywords and thresholds when a
  pattern's evidence count reaches the Tier 4 threshold (default 10). This is
  the activation that EVOLVE-002 spec.md `Out of Scope — Tier↔surface mapping
  activation` explicitly defers to this SPEC.
- **REQ-HEV3-002** (Compound — value-range validation): **While** the Curator
  proposes an edit to the `auto_detection.rules.<level>.conditions` threshold
  values, **When** the writer evaluates the proposed write, the writer SHALL
  validate that each numeric threshold falls within its registered
  `[lower, upper]` bound and reject with a typed `ErrAutoDetectionOutOfRange`
  error WITHOUT touching the file when the proposed value is out of range.
  Value-range validation is the safety fence that admits `auto_detection` to
  the Evolvable zone without permitting pathological thresholds (e.g. a
  `file_count <= 0` rule that would classify every SPEC as minimal).
- **REQ-HEV3-003** (Event-driven — tier→surface dispatch): **When** the
  Curator emits a tier-qualified proposal, the dispatch layer SHALL select the
  target surface by tier — Tier 3 → `CLAUDE.local.md` (the append-only
  Learned surface), Tier 4 → `CLAUDE.md` (the digest-layer managed block) —
  and call `curator.TierGatedWrite` with the matching `BlockType`
  (`BlockTypeLearnedLocal` / `BlockTypeLearnedWorkflow`).
- **REQ-HEV3-004** (Unwanted behavior — no cross-surface leak): The dispatch
  layer shall not write a Tier-3 proposal to `CLAUDE.md` nor a Tier-4 proposal
  to `CLAUDE.local.md`. Cross-surface leak is a reward-hacking shape
  (REQ-HEV2-027 self-tier-escalation binding, extended to the dispatch layer).

### C.2 L5 orchestrator approval round integration

- **REQ-HEV3-005** (Compound — L5 round injection): **While** the Curator
  proposes a CLAUDE.md or CLAUDE.local.md write, the production wiring
  **shall** route the proposal through an L5 `AskUserQuestion` approval round
  via the orchestrator's user-interaction channel (the writer is a
  subagent-side function and cannot call `AskUserQuestion` directly per
  REQ-HEV2-035 / C-HRA-008). **When** the orchestrator returns an
  `ApprovalDecision`, the wiring **shall** pass it — together with a
  `RejectionRecorder` bound to the lineage writer — into
  `curator.WriteManagedBlockGated`.
- **REQ-HEV3-006** (Unwanted behavior — no autonomous write path): The
  production wiring shall NOT call `WriteManagedBlock` /
  `TierGatedWrite` / `AppendLearnedLocal` directly without first passing
  through `WriteManagedBlockGated` with a live `ApprovalDecision`. An
  autonomous write path bypassing L5 is the reward-hacking surface the design
  SSOT §5 Loop 2 + "인간은 목표와 기준 제공" forbids; the gate is the
  invariant, not the writer.
- **REQ-HEV3-007** (Event-driven — rejection recording): **When** the L5
  `ApprovalDecision.Approved` is false, the wiring SHALL invoke the
  `RejectionRecorder` with `("rejected", rationale)` so a `LineageEntry`
  records the rejection, and SHALL NOT touch the target file (the
  `WriteManagedBlockGated` contract at `approval.go:41-51` enforces this; the
  wiring must thread the recorder through).

### C.3 EVOLVE-002 residual — PRODUCTION wiring (TierGatedWrite / WriteManagedBlockGated)

- **REQ-HEV3-008** (Compound — production caller for TierGatedWrite + sequencing):
  **When** the L5 approval decision is `Approved: true`, the dispatch SHALL call
  `curator.TierGatedWrite` (tier_gate.go:71) as the SOLE write entry point on
  the approval path (the tier-validation AND the write happen here). The
  baseline of 0 production callers (verified: `grep -rn TierGatedWrite
  internal/ cmd/ pkg/ | grep -v _test.go | grep -v curator/` returns empty)
  MUST rise to ≥1 production call site. The dispatch SHALL NOT call
  `WriteManagedBlockGated` on the approval path — its internal
  `WriteManagedBlock` delegate (approval.go:50) would double-write the block
  that `TierGatedWrite` already persisted (the double-write hazard; design.md
  §A.1 step 6→7a).
- **REQ-HEV3-009** (Compound — production caller for WriteManagedBlockGated +
  sequencing): **When** the L5 approval decision is `Approved: false`, the
  dispatch SHALL call `curator.WriteManagedBlockGated` (approval.go:41) with
  the rejection to exercise its `RejectionRecorder` contract (the rejection
  path — no write occurs; the gate returns `ErrApprovalRejected` at
  approval.go:48). The dispatch SHALL call it with a live `ApprovalDecision`
  and a non-nil `RejectionRecorder` bound to the lineage writer. The baseline
  of 0 production callers MUST rise to ≥1 production call site. The two
  functions are called on DISJOINT branches (approval → `TierGatedWrite`;
  rejection → `WriteManagedBlockGated`), ensuring `WriteManagedBlock` is
  reached at most once per Curator cycle (the double-write-avoidance
  invariant; design.md §A.1 step 6→7a/7b).
- **REQ-HEV3-010** (Ubiquitous — tier→surface config read): The dispatch
  layer SHALL read the pattern's tier from the existing learner aggregation
  (`tier.ClassifyStatus`) and the target surface from the tier→surface map
  (§C.1 REQ-HEV3-003), then construct the `BlockContent` with the explicit
  `Tier` field set — so the writer's `ErrTierNotQualified` self-tier-escalation
  guard (REQ-HEV2-027) is exercised on every Curator cycle, not just in tests.
- **REQ-HEV3-011** (Unwanted behavior — no inert wiring): The production
  wiring call sites SHALL NOT be dead code (unreachable branches guarded by
  permanently-false predicates) and SHALL NOT be feature-flagged off by
  default. A wiring that ships disabled-by-default is the named inert-wiring
  anti-pattern — the gates must fire on every Curator cycle.

### C.4 L2 Canary activation

- **REQ-HEV3-012** (Compound — shadow-apply): **While** the Curator proposes a
  Tier-4 CLAUDE.md managed-block write, the L2 Canary layer SHALL shadow-apply
  the proposed block to a temporary copy of the target file and run the
  existing `safety.EvaluateCanary` against the shadow-applied state, before
  the real write reaches the file.
- **REQ-HEV3-013** (Compound — regression gate / held-out): **While** the
  shadow-applied proposal regresses a held-out signal (a metric in the
  existing regression-gate suite — e.g. managed-block byte budget, marker
  integrity, frontmatter validity), **When** the L2 evaluator returns
  `CanaryResult.Rejected == true`, the wiring SHALL reject the proposal with
  `Decision.RejectedBy == 2` and SHALL NOT touch the target file.
- **REQ-HEV3-014** (Event-driven — Canary veto rollback auto-register):
  **When** the L2 Canary vetoes a proposal post-apply (the `CanaryVeto`
  `VetoAndRollback` path at canary_veto.go:106), the wiring SHALL
  auto-register the vetoed pattern key as negative evidence in the A7 registry
  (§C.6) — the Canary veto and the A7 registry are the two surfaces that
  record "this promotion failed"; they MUST agree.

### C.5 L3 Contradiction activation

- **REQ-HEV3-015** (Ubiquitous — replace no-op): The
  `safety.Pipeline.l3ContradictionCheck` function field (pipeline.go:70-73,
  currently `return harness.ContradictionReport{}` always — verified) SHALL
  be replaced with a real Frozen-rules contradiction check that consults the
  registered Frozen-rule set and returns a `ContradictionReport` whose
  `HasContradiction()` is true when the proposal contradicts a registered
  Frozen rule.
- **REQ-HEV3-016** (Compound — Frozen-rules contradiction rejection):
  **While** a Curator proposal targets a surface whose content contradicts a
  registered Frozen rule (a rule in `.claude/rules/moai/**`, a Frozen
  evaluator policy, or the A1 permission-surface Frozen set), **When** the L3
  evaluator returns `HasContradiction() == true`, the wiring SHALL reject the
  proposal with `Decision.RejectedBy == 3` and SHALL NOT touch the target
  file.
- **REQ-HEV3-017** (Capability gate — Frozen-rule registry): **Where** a
  Frozen rule carries a stable registered identifier, the L3 evaluator SHALL
  cite the rule identifier in the rejection `Reason` so the L5 reviewer (and
  the lineage entry) can trace which Frozen rule was contradicted. Opaque
  rejections without a rule reference are not acceptable (audit-trail
  discipline).

### C.6 Design delta A7 — Re-proposal suppression (negative-evidence registry)

- **REQ-HEV3-018** (Ubiquitous — registry data structure): The harness SHALL
  provide a negative-evidence registry at
  `.moai/state/negative-evidence.jsonl` (append-only jsonl, same family as
  the routing-ledger) recording for each entry: the pattern key, the outcome
  (`rejected` | `rolled-back`), the timestamp, the evidence count at
  rejection, the cooldown-until timestamp, and the machine-signal reference
  (lineage key / rollback SHA / gate-reject event). EVOLVE-002 authored the
  registry KEY/token only (REQ-HEV2-010 `ledger_key`); this REQ ships the
  registry itself.
- **REQ-HEV3-019** (Event-driven — register on reject): **When** a Curator
  proposal is rejected (by L2 Canary, L3 Contradiction, or L5 orchestrator
  disapproval), the wiring SHALL append a negative-evidence entry to the
  registry keyed by the proposal's pattern key, with the outcome `rejected`
  and the cooldown-until computed from the configured cooldown duration.
- **REQ-HEV3-020** (Event-driven — register on rollback): **When** a promotion
  is rolled back (via the existing `RestoreSnapshot` path, the `CanaryVeto`
  `VetoAndRollback` path, or a future `moai harness rollback` verb), the
  wiring SHALL auto-append a negative-evidence entry keyed by the rolled-back
  pattern key, with the outcome `rolled-back`. Rollback auto-registration is
  the "실패한 수정도 기록" (failed modifications are also recorded) principle
  from design SSOT §1 A7 / card 8.
- **REQ-HEV3-021** (Compound — re-proposal block): **While** a pattern key
  has a negative-evidence entry whose cooldown-until has not elapsed AND whose
  new-evidence-since-rejection count is below the configured threshold N
  (default N=3 — three NEW post-rejection evidences are required to re-propose
  the same key), **When** the Curator attempts to re-propose the same pattern
  key, the wiring SHALL block the proposal early (before L2/L3/L5) with a
  typed `ErrReProposalSuppressed` citing the registry entry. Same-key
  re-proposal is the exact anti-pattern A7 exists to prevent.
- **REQ-HEV3-022** (Unwanted behavior — no permanent suppression): The
  registry SHALL NOT permanently suppress a pattern key. Once the cooldown
  elapses AND N new evidences accumulate, the pattern key is re-eligible.
  Permanent suppression would freeze the harness against corrective
  re-learning; the cooldown + new-evidence threshold is the bounded gate.

### C.7 Design delta A1 — Permission surface Frozen registration

- **REQ-HEV3-023** (Ubiquitous — block-list expansion): The
  `internal/harness/safety/frozen_guard.go` `frozenPrefixes` list (line 18 —
  verified) SHALL be expanded to include the permission surfaces elevated to
  Frozen Zone in v5.1: `.claude/settings.json`, `.claude/settings.local.json`,
  the hook-registration axis (the `settings.json` `hooks` block), and the
  frozen-guard source files themselves (self-protection — the guard cannot be
  edited away by a proposal it permits).
- **REQ-HEV3-024** (Compound — L1 block on permission surface): **While** a
  Curator proposal targets a path in the expanded permission-surface Frozen
  set, **When** the L1 Frozen guard (`safety.IsFrozen` at frozen_guard.go:34)
  evaluates the proposal, the guard SHALL return true and the wiring SHALL
  reject the proposal with `Decision.RejectedBy == 1`. Permission-surface
  targeting is the "학습 루프의 권한 침범" (learning loop permission intrusion)
  risk the design SSOT §7 risk grid MID row names; L1 is the mechanical block.
- **REQ-HEV3-025** (Unwanted behavior — guard self-protection): The
  `frozen_guard.go` source files (both `internal/harness/safety/frozen_guard.go`
  AND `internal/harness/frozen_guard.go`) SHALL themselves be in the Frozen
  set — a Curator proposal that targets the guard source is L1-blocked. The
  guard cannot be the surface that permits its own modification.

### C.8 GLM observe-only guard

- **REQ-HEV3-026** (Capability gate — model-class gate): **Where** the
  Curator pipeline detects the session model class is GLM (per the
  `routing-ledger.jsonl` `model_class` field — EVOLVE-001 output), the
  Curator SHALL generate Tier 3+ promotion proposals as observe-only — the
  proposal is recorded in the ledger for future Opus/Fable sessions to
  promote, but is NOT written to any Learned surface by a GLM session. This
  binds the design SSOT §2 + STOP lesson: "GLM 세션은 관찰만" (GLM sessions
  observe only).
- **REQ-HEV3-027** (Ubiquitous — gate location): The model-class gate SHALL
  live in the Curator pipeline upstream of the writer (the dispatch layer
  that calls `TierGatedWrite`), NOT in the writer itself. The writer is the
  mechanical write primitive; the model-class gate is a Curator-pipeline
  policy concern (REQ-HEV2-031 machine-signal-only binding, extended to the
  model-class axis).

### C.9 Anti-fabrication + template neutrality (inherited, extended)

- **REQ-HEV3-028** (Unwanted behavior — registry anti-fabrication): The
  negative-evidence registry entries SHALL NOT carry verbatim user request
  text, internal SPEC IDs, REQ/AC tokens, internal session dates, or commit
  SHAs in any human-readable summary field. Entries carry machine-signal
  references (lineage key, rollback SHA) in typed reference fields only; the
  `pattern_key` is a structural token (`request_class + subcommand + mode +
  outcome` per design SSOT §5 Loop 1), not free-form prose.
- **REQ-HEV3-029** (Unwanted behavior — auto_detection template neutrality):
  The auto_detection schema validator mechanism (the value-range bound
  definitions — REQ-HEV3-002) MAY ship to the template tree
  (`internal/template/templates/.moai/config/sections/harness.yaml`) as a
  neutral empty schema, but learned threshold DATA (a Curator-proposed
  threshold correction) SHALL NEVER ship to templates — it is per-project
  learned state, not template-distributable (CLAUDE.local.md §25
  template-internal-content isolation).

### C.10 Go quality + subagent boundary + reachability guarantees

- **REQ-HEV3-030** (Ubiquitous — coverage and quality): The new
  `internal/harness/safety/` files (the L3 contradiction activation, the A7
  registry), the new `internal/harness/curator/` dispatch wiring, and the
  extended `internal/config/types.go` auto_detection validator SHALL reach
  ≥ 90% statement coverage, use table-driven tests with `t.TempDir()`
  isolation, wrap errors with `%w`, and set no OTEL environment variables via
  `t.Setenv` in parallel tests.
- **REQ-HEV3-031** (Unwanted behavior — subagent boundary): The new wiring in
  `internal/harness/safety/`, `internal/harness/curator/`, and the Curator
  dispatch layer SHALL NOT invoke `AskUserQuestion` (subagent boundary per
  C-HRA-008 / archived-agent-rejection.md §C). L5 approval is routed through
  the orchestrator; the Curator returns a proposal artifact + the
  `ApprovalDecision` channel contract, and the orchestrator runs the
  `AskUserQuestion` round.
- **REQ-HEV3-032** (Unwanted behavior — no new hook surface): This SPEC shall
  add NO new hook wrapper script, NO `settings.json` / `settings.json.tmpl`
  hook registration change, and NO new gate beyond the activation of L2/L3
  and the A7 registry consult. The activation reuses the existing
  `safety.Pipeline.Evaluate` L1→L5 chain.
- **REQ-HEV3-033** (Ubiquitous — L3 reachability): The L3 contradiction check
  SHALL fire on every Curator proposal cycle (not be a dead no-op). The
  baseline of 0 real L3 consultations (the no-op at pipeline.go:70-73 never
  consults the Frozen-rule set — verified) MUST rise to ≥1 consultation per
  Curator cycle. A grep-count AC that passes while the L3 evaluator is inert
  is the named `feedback_ac_token_presence_not_reachability` anti-pattern;
  the AC matrix binds this with a behavior-verifiable assertion (inject a
  contradiction → assert `RejectedBy == 3`).
- **REQ-HEV3-034** (Ubiquitous — L2 reachability): The L2 Canary check SHALL
  fire for every Tier-3+ Curator proposal (shadow-apply + regression gate).
  The baseline of 0 L2 consultations for the Curator write path MUST rise to
  ≥1 consultation per Tier-3+ Curator proposal.
- **REQ-HEV3-035** (Ubiquitous — A7 registry reachability): The A7
  negative-evidence registry SHALL be consulted before every Curator proposal
  (early-block check for same-key re-proposal). The baseline of 0 registry
  consultations MUST rise to ≥1 consultation per Curator cycle.

## §D. Reference — wiring contract (SSOT)

### D.1 Production wiring call chain (target)

```
Curator pipeline (applier.go Apply path, extended)
  │
  ├── 0. A7 registry consult (early block — ErrReProposalSuppressed)
  │      └─ .moai/state/negative-evidence.jsonl read
  │
  ├── 1. tier.ClassifyStatus(observations) → tier N
  │
  ├── 2. GLM observe-only gate (REQ-HEV3-026)
  │      └─ if GLM → record in ledger, return (no write)
  │
  ├── 3. tier→surface dispatch (REQ-HEV3-003)
  │      └─ Tier 3 → CLAUDE.local.md, Tier 4 → CLAUDE.md
  │
  ├── 4. safety.Pipeline.Evaluate (L1→L2→L3→L4) — the gate chain
  │      ├── L1 IsFrozen (A1 expanded frozenPrefixes — REQ-HEV3-023)
  │      ├── L2 EvaluateCanary (shadow-apply + regression — REQ-HEV3-012/013)
  │      ├── L3 ContradictionCheck (real Frozen-rules check — REQ-HEV3-015)
  │      └── L4 RateLimit (existing)
  │
  ├── 5. L5 AskUserQuestion round (orchestrator-mediated — REQ-HEV3-005)
  │      └─ orchestrator returns ApprovalDecision
  │
  ├── 6. curator.WriteManagedBlockGated(path, blockType, content, decision, recorder)
  │      └─ the EVOLVE-002 API (approval.go:41), now with a production caller
  │
  ├── 7a. on approval  → curator.TierGatedWrite dispatch → write + snapshot + lineage
  ├── 7b. on rejection → RejectionRecorder("rejected", rationale) → lineage + A7 registry
  └── 7c. on rollback  → RestoreSnapshot + A7 registry auto-append (REQ-HEV3-020)
```

### D.2 A7 negative-evidence registry entry shape (sketch)

```jsonc
{
  "pattern_key": "feature+plan+autopilot+success",   // structural token, NOT prose
  "outcome": "rejected",                              // "rejected" | "rolled-back"
  "ts": "2026-07-12T14:03:00Z",
  "evidence_count_at_event": 7,
  "cooldown_until": "2026-07-14T14:03:00Z",           // 48h default (reuse canary_veto cooldown)
  "new_evidence_since_event": 0,                      // incremented by later observations of same key
  "machine_signal_ref": "lineage:manifest.jsonl#ln=42",  // evidence-or-null
  "gate_origin": "L3"                                 // "L1"|"L2"|"L3"|"L4"|"L5"|"rollback"
}
```

The full machine-verifiable AC matrix (AC-HEV3-001 … AC-HEV3-NNN) lives in
`acceptance.md` (SSOT). Every REQ maps to at least one AC; the gates (L1/L2/L3/A7)
each carry a baseline-0 → ≥1 behavior assertion (inject the gate-triggering
input → assert the gate fires) per the `feedback_ac_token_presence_not_reachability`
discipline inherited from EVOLVE-001 / EVOLVE-002.

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — Console verbs (EVOLVE-004)

- NO `/moai harness evolve | promote | demote | freeze | unfreeze` verbs, NO
  `status` / `doctor` extension for the A7 registry view, NO
  `moai harness rollback` verb (the rollback *machinery* — `RestoreSnapshot` —
  is reused here, but the human-facing console verb that drives it is
  `SPEC-HARNESS-EVOLVE-004`). This SPEC authors the wiring + gates + registry;
  the human-facing console surface that drives it is EVOLVE-004.

### Out of Scope — Recall wiring + typed parser + template deployment (EVOLVE-005)

- NO Phase −1 Harness Recall wiring (the *consumption* of the digest layer by
  the router), NO Phase Ω routing-bias consumption, NO `harness-spec.yaml`
  typed Go parser (G5), NO template-deployment verification of the empty
  marker beyond the marker-contract checks EVOLVE-002 already shipped
  (`SPEC-HARNESS-EVOLVE-005`). This SPEC *activates* the write path; the
  read-side Recall wiring is EVOLVE-005.

### Out of Scope — EVOLVE-002 write-layer API changes

- NO modification to the `curator.WriteManagedBlock` / `TierGatedWrite` /
  `WriteManagedBlockGated` / per-bullet CRUD / `AppendLearnedLocal` signatures
  or behavior. EVOLVE-002 shipped the write-layer API; this SPEC wires it into
  production. If a wiring need reveals an API gap, return a blocker and
  re-delegate to manager-spec — do NOT modify the EVOLVE-002 API in this
  SPEC's run-phase.

### Out of Scope — Loop 0 / Loop 1 changes (EVOLVE-001 boundary)

- NO changes to the EVOLVE-001 observation layer. The routing-ledger writer
  (`internal/harness/routing/`) is untouched; the A7 registry's
  evidence-counter consumes the ledger read-only via the existing
  `routing/reader.go`.

### Out of Scope — Safety pipeline reorder

- NO change to the L1→L2→L3→L4→L5 order (pipeline.go:14-15 `[HARD]`). This
  SPEC replaces the L3 no-op function BODY and wires the Curator path through
  the existing pipeline; the order is immutable.

### Out of Scope — New managed-block type

- NO new `BlockType` beyond `BlockTypeLearnedWorkflow` /
  `BlockTypeLearnedLocal` / `BlockTypeHarnessGenerated` (EVOLVE-002 §D.2).
  The auto_detection Tier-4 surface is edited in-place in `harness.yaml` (a
  config edit, not a managed-block write); it does not introduce a new block
  type.

### Out of Scope — New hook surface

- NO new hook wrapper script, NO `settings.json` / `settings.json.tmpl` hook
  registration change (REQ-HEV3-032). The activation reuses the existing
  `safety.Pipeline.Evaluate` chain invoked from the Curator pipeline.

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site
  4-locale updates are a follow-up sync/docs concern.

## §F. Cross-References

- `.moai/reports/harness-self-evolving-redesign-final-20260712.html` — design
  SSOT (§1 A1/A6/A7 deltas; §3 gaps G4/G7; §4 3-Zone edit-surface contract
  Evolvable-zone A6 row + Frozen-zone A1 row; §5 Loop 2 gate chain
  L2→L3→L4→L5; §7 M3 milestone + verification matrix + risk grid MID rows).
- `.moai/specs/SPEC-HARNESS-EVOLVE-002/spec.md` — upstream Epic predecessor
  (Loop 2 write layer; `depends_on` target). Its `§Out of Scope — Tier↔surface
  mapping activation`, `§Out of Scope — Gate activation`, and `§Out of Scope —
  Re-proposal suppression` sections DEFINE this SPEC's scope.
- `internal/harness/curator/tier_gate.go` (line 71) — `TierGatedWrite`, the
  tier-gated dispatch entry point this SPEC wires into production (REQ-HEV3-008).
- `internal/harness/curator/approval.go` (line 41) — `WriteManagedBlockGated` +
  `ApprovalDecision` + `RejectionRecorder`, the L5-gated write contract this
  SPEC wires into production (REQ-HEV3-009).
- `internal/harness/safety/pipeline.go` (line 70-73) — the L3 no-op this SPEC
  replaces (REQ-HEV3-015). Line 14-15: the immutable L1→L2→L3→L4→L5 order.
- `internal/harness/safety/frozen_guard.go` (line 18) — `frozenPrefixes`, the
  L1 block-list this SPEC expands with permission surfaces (REQ-HEV3-023, A1).
- `internal/harness/safety/canary_veto.go` — `CanaryVeto` 48h cooldown
  machinery, reused as the A7 cooldown-enforcement primitive where semantics
  overlap (REQ-HEV3-014 / REQ-HEV3-021).
- `internal/config/types.go` (line 740) — `AutoDetectionConfig`, the existing
  config struct extended additively with value-range validation (REQ-HEV3-002).
- `.moai/config/sections/harness.yaml` (line 2) — the live `auto_detection`
  block (already present — verified), the Tier-4 Evolvable surface this SPEC
  registers (REQ-HEV3-001).
- `internal/harness/routing/reader.go` (EVOLVE-001 output) — read-only
  consumption by the A7 registry's evidence-counter (REQ-HEV3-021).
- CLAUDE.local.md §2 (Template-First) + §25 (Template Internal-Content
  Isolation) — the auto_detection schema validator ships to templates;
  registry DATA never does (REQ-HEV3-029).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 —
  the no-unobserved-defect-claim invariant; the baseline-0 → ≥1 ACs are its
  operational mechanism (every gate's reachability is mechanically verified).
- `.claude/rules/moai/workflow/archived-agent-rejection.md` §C — subagent
  boundary (REQ-HEV3-031); the Curator returns proposals + the
  `ApprovalDecision` contract, the orchestrator runs `AskUserQuestion`
  (REQ-HEV3-005).
- `SPEC-HARNESS-EVOLVE-004..005` (unauthored) — Epic successors consuming
  these wiring surfaces.
- `plan.md` / `acceptance.md` / `design.md` / `research.md` — implementation
  plan + AC matrix + architecture decisions + codebase investigation (SSOT).
