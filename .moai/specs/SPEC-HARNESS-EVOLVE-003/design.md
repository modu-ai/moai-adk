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

# SPEC-HARNESS-EVOLVE-003 — Design Decisions

> Counterpart to `spec.md` (requirements SSOT), `plan.md` (implementation plan),
> `acceptance.md` (AC matrix), `research.md` (codebase investigation). This
> document owns the architecture decisions, alternatives considered, and the
> rationale for the chosen path. It is the SSOT for "why this shape, not the
> other shape".

## §A. Architecture Context

### A.1 The 3-Loop harness (EVOLVE Epic structure)

The HARNESS-EVOLVE Epic implements Lilian Weng's "Harness Engineering for
Self-Improvement" thesis as a 3-Loop architecture (design SSOT §5):

- **Loop 0 (Generator / observation)** — SHIPPED: `SPEC-HARNESS-EVOLVE-001`.
- **Loop 2 (Curator / promotion write)** — WRITE LAYER SHIPPED (EVOLVE-002),
  PRODUCTION WIRING + GATES + REGISTRY = **THIS SPEC**.
- **Loop 1 (Reflector / aggregation)** — PARTIALLY SHIPPED: the existing
  `learner.go`. Full Loop 1 activation (with A7 negative-evidence cross-check)
  is folded into this SPEC's A7 registry consult.

This SPEC sits between EVOLVE-002 (provides the write-layer API this SPEC
wires) and EVOLVE-004 (provides the console verbs that drive the wiring this
SPEC ships). The activated Curator pipeline is:

```
Curator pipeline (applier.go Apply path, extended by this SPEC)
   │
   ├── 0. A7 registry consult (early block — ErrReProposalSuppressed)
   │      re-proposal suppression BEFORE any gate fires
   │
   ├── 1. tier.ClassifyStatus(observations) → tier N
   │      the existing 4-tier ladder [1,3,5,10]
   │
   ├── 2. GLM observe-only gate (REQ-HEV3-026)
   │      if GLM session → record in ledger, return (no write)
   │
   ├── 3. tier→surface dispatch (REQ-HEV3-003, §B below)
   │      Tier 3 → CLAUDE.local.md, Tier 4 → CLAUDE.md
   │
   ├── 4. safety.Pipeline.Evaluate (L1→L2→L3→L4 — the gate chain, §C below)
   │      L1 IsFrozen (A1 expanded — §D)
   │      L2 EvaluateCanary (shadow + regression — §E)
   │      L3 ContradictionCheck (real Frozen-rules check — §F, replaces no-op)
   │      L4 RateLimit (existing, unchanged)
   │
   ├── 5. L5 AskUserQuestion round (orchestrator-mediated, §G state machine)
   │      returns ApprovalDecision
   │
   ├── 6. L5 ApprovalDecision branch point (double-write-avoidance invariant —
   │      REQ-HEV3-008/009 sequencing: WriteManagedBlockGated and
   │      TierGatedWrite are called on DISJOINT branches, NEVER both on the
   │      same branch, so WriteManagedBlock is reached at most once per cycle.
   │      Both internally delegate to WriteManagedBlock — approval.go:50 and
   │      tier_gate.go:103 respectively — so calling both sequentially would
   │      double-write the block.)
   │
   ├── 7a. on approval → curator.TierGatedWrite(path, blockType, observations, content)
   │      → tier-validate + SOLE write + snapshot + lineage
   │      (WriteManagedBlockGated is NOT called on this branch — its internal
   │      WriteManagedBlock delegate would double-write)
   │
   ├── 7b. on rejection → curator.WriteManagedBlockGated(path, blockType, content,
   │      decision{Approved:false}, recorder) → RejectionRecorder → lineage +
   │      A7 registry append (no write; TierGatedWrite is NOT called here)
   └── 7c. on rollback  → RestoreSnapshot + A7 registry auto-append
```

### A.2 The 3-Zone edit-surface contract (design SSOT §4)

This SPEC operates across the **Frozen Zone** (A1 expansion) and the
**Evolvable Zone** (A6 auto_detection registration). The Learned Zone is
where the writes LAND (Tier 3 / Tier 4 surfaces EVOLVE-002 built); this SPEC
drives them.

| Zone | Surfaces | This SPEC's interaction |
|------|----------|-------------------------|
| **Frozen (A1 expansion)** | + settings.json / settings.local.json / hook registration / frozen-guard source | **THIS SPEC** — L1 block-list expansion |
| **Evolvable (A6 registration)** | + harness.yaml `auto_detection` block (value-range validated) | **THIS SPEC** — Tier-4 surface registration |
| Learned | CLAUDE.md block + CLAUDE.local.md append (EVOLVE-002 built) + ledger layer | **THIS SPEC** — production wiring drives these surfaces |

### A.3 The A7 negative-evidence registry (design delta A7)

The registry is the "실패한 수정도 기록" (failed modifications are also
recorded) principle mechanized (design SSOT §1 A7 / card 8). It is a NEW
ledger-layer surface (distinct from EVOLVE-001's routing-ledger and
EVOLVE-002's lineage manifest). The cross-layer linkage: the digest-layer
bullet's `ledger_key` (REQ-HEV2-010) MAY reference an A7 registry entry,
making a "this bullet was once rejected" history navigable from the digest.

## §B. Tier↔Surface Mapping Table (A6 activation)

### B.1 The mapping (activation target)

| Tier | Evidence threshold | Target surface | BlockType | Writer entry point |
|------|-------------------|----------------|-----------|-------------------|
| 1 | 1 observation | auto-memory (Tier 1-2, observe-only) | (no managed block) | `~/.claude/projects/<hash>/memory/` |
| 2 | 3 observations | auto-memory (Tier 1-2) | (no managed block) | same |
| 3 | 5 observations | CLAUDE.local.md append-only | `BlockTypeLearnedLocal` | `curator.TierGatedWrite` → `AppendLearnedLocal` |
| 4 | 10 observations | CLAUDE.md digest managed block | `BlockTypeLearnedWorkflow` | `curator.TierGatedWrite` → `WriteManagedBlock` |
| 4 (Evolvable surface) | 10 observations | harness.yaml `auto_detection` block | (config edit, not a managed block) | (separate config-editor path, value-range validated) |

The dispatch reads the pattern's observation count → `tier.ClassifyStatus` →
the tier → the surface → the `BlockType` → calls `TierGatedWrite`.

### B.2 auto_detection value-range validation (A6 safety fence)

The `auto_detection` block ships with a schema of per-field bounds:

```go
// AutoDetectionBounds are the value-range bounds for auto_detection threshold
// edits. A Curator proposal editing auto_detection.rules.<level>.conditions
// is validated against these bounds; out-of-range values are rejected with
// ErrAutoDetectionOutOfRange (REQ-HEV3-002).
var AutoDetectionBounds = map[string][2]int{
    "file_count":  {1, 1000},       // file_count thresholds ∈ [1, 1000]
    // spec_priority is an enum, not numeric — validated against the enum set
    // domain is an enum — validated against the registered domain set
}
```

The bounds are SHIPPED to templates as a neutral empty schema (the mechanism);
learned threshold CORRECTIONS (a Curator-proposed adjustment) NEVER ship
(REQ-HEV3-029).

### B.3 Decision: register auto_detection as a config edit, NOT a managed block

**Decision**: the `auto_detection` Tier-4 surface is edited in-place in
`harness.yaml` (a YAML config edit), NOT via a new managed-block type. The
config editor path is value-range validated; it does NOT introduce a new
`BlockType` (the EVOLVE-002 §D.2 marker contract is untouched).

**Rationale**: `auto_detection` is a YAML config block, not a markdown managed
section. Forcing it through a managed-block writer would require a new
`BlockType`, a new marker contract, and a merge-extension — all out of scope
per EVOLVE-003 §E. The config editor reuses the existing YAML read/write path
(`internal/config/`).

## §C. The 5-Layer Gate Chain (activation, not reorder)

### C.1 The immutable order

`safety/pipeline.go:14-15`: `[HARD] L1→L2→L3→L4→L5 order is immutable`. This
SPEC does NOT reorder. It activates L2 and L3 for the Curator write path and
expands L1's block-list.

### C.2 L1 Frozen Guard (A1 expansion)

`safety/frozen_guard.go:18` currently lists (3 entries — verified via
`sed -n '18,25p'`; note `.claude/skills/moai-` is the dash-form, there is NO
`.claude/skills/moai/` directory entry here, unlike the meta-harness
`internal/harness/frozen_guard.go` 4-entry list):
```
.claude/agents/moai/
.claude/skills/moai-
.claude/rules/moai/
```

The A1 expansion adds (design SSOT §4 Frozen-zone A1 row):
```
.claude/settings.json
.claude/settings.local.json
internal/harness/safety/frozen_guard.go    (self-protection)
internal/harness/frozen_guard.go           (self-protection — the meta-harness guard too)
```

The hook-registration axis (the `settings.json` `hooks` block) is covered by
the `.claude/settings.json` prefix match (the hooks block lives inside
settings.json). A separate "permission mode" entry is NOT needed because
permission mode is a field inside settings.json.

**Decision**: expand by path prefix (consistent with the existing pattern),
not by content introspection. Content introspection (parsing settings.json to
check whether the `hooks` block is edited) is fragile and out of scope; the
path-prefix block covers the whole file, which is the conservative choice.

### C.3 L2 Canary (Curator path activation)

The existing `safety.EvaluateCanary` + `CanaryVeto` machinery is wired to the
harness-applier path (skill frontmatter edits). This SPEC extends L2 to cover
the Curator Learned-surface write path:

- **Shadow-apply**: the proposed CLAUDE.md block / CLAUDE.local.md append is
  applied to a temp copy; `EvaluateCanary` runs against the temp copy.
- **Regression gate**: the held-out signal set includes (a) the 3K digest
  budget (REQ-HEV2-008), (b) marker integrity (start/end markers present), (c)
  frontmatter validity of the target file (CLAUDE.md / CLAUDE.local.md).
- **Canary veto → A7 agreement**: when `CanaryVeto.VetoAndRollback` fires, the
  A7 registry gains a `rolled-back` entry for the same pattern key
  (REQ-HEV3-014). The two surfaces agree.

### C.4 L3 Contradiction (replace no-op)

`safety/pipeline.go:70-73` is explicitly a no-op:
```go
l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
    // Always return empty report in Phase 3 (actual skill trigger loading in Phase 4)
    return harness.ContradictionReport{}
},
```

This SPEC replaces the function BODY (not the field signature, not the order)
with a real Frozen-rules consult:

```go
l3ContradictionCheck: func(proposal harness.Proposal) harness.ContradictionReport {
    return DetectFrozenRuleContradictions(proposal, FrozenRuleRegistry)
},
```

Where `FrozenRuleRegistry` is the typed registry seeded from the L1 prefix
list + the A1 permission surfaces (REQ-HEV3-017 — cited rule identifiers in
the rejection Reason).

**Note**: the existing `DetectOverlappingTriggers` / `DetectChainRuleContradictions`
(consumed by a different L3 path for skill triggers) are NOT removed — they
serve the harness-applier path. The new `DetectFrozenRuleContradictions` is a
sibling function for the Curator path.

## §D. A7 Negative-Evidence Registry — Data Structure + Cooldown

### D.1 Registry entry shape

```jsonc
{
  "pattern_key": "feature+plan+autopilot+success",
  "outcome": "rejected",                              // "rejected" | "rolled-back"
  "ts": "2026-07-12T14:03:00Z",
  "evidence_count_at_event": 7,
  "cooldown_until": "2026-07-14T14:03:00Z",           // 48h default
  "new_evidence_since_event": 0,                      // incremented by later observations
  "machine_signal_ref": "lineage:manifest.jsonl#ln=42",
  "gate_origin": "L3"                                 // "L1"|"L2"|"L3"|"L4"|"L5"|"rollback"
}
```

### D.2 Re-proposal block predicate

A pattern key `K` is **re-proposal-suppressed** iff ALL hold:
1. An entry `E` exists for `K` in the registry.
2. `E.cooldown_until > now` (cooldown not elapsed) OR
   `new_evidence_since_event < N` (N default 3 — three NEW post-rejection
   evidences required).

The block lifts when BOTH the cooldown elapses AND N new evidences accumulate
(REQ-HEV3-022 — no permanent suppression).

### D.3 Cooldown primitive reuse

The existing `canary_veto.go` `RecordCooldown` / `CheckCooldown` is a 48h
cooldown keyed by proposal. The A7 registry reuses the DURATION constant
(`canaryVetoCooldown = 48 * time.Hour`, canary_veto.go:27) but keys on the
pattern key (not the proposal) and adds the N-new-evidence accumulator. The
two are distinct: the canary-veto cooldown blocks a re-proposal of the same
PROPOSAL; the A7 registry blocks a re-proposal of the same PATTERN.

**Decision**: reuse the constant, do NOT reuse the `cooldownEntry` struct
(canary_veto.go:50). The A7 entry shape (§D.1) is richer (pattern_key +
outcome + new_evidence_since_event + gate_origin).

### D.4 Rollback auto-register

When `RestoreSnapshot` (applier.go) fires for a Curator promotion, the wiring
appends a `rolled-back` entry to the A7 registry for the rolled-back pattern
key. This is the "롤백은 자동으로 negative 등재" (rollback auto-registers
negative evidence) principle from design SSOT §1 A7.

## §E. L5 Orchestrator Approval Round — State Machine

The L5 round is the human gate. The state machine:

```
 ┌──────────────────────────────────────────────────────────┐
 │  STATE: PROPOSAL_READY                                   │
 │  (Curator emits a tier-qualified proposal after L1-L4)   │
 └──────────────────────┬───────────────────────────────────┘
                        │
                        ▼
 ┌──────────────────────────────────────────────────────────┐
 │  STATE: AWAITING_L5                                      │
 │  The dispatch returns the proposal to the orchestrator   │
 │  via a blocker-report-shaped artifact (NOT AskUserQuestion│
 │  — the dispatch is subagent-side). The orchestrator runs  │
 │  the AskUserQuestion round.                              │
 └──────────────────────┬───────────────────────────────────┘
                        │
                ┌───────┴───────┐
                │               │
                ▼               ▼
 ┌──────────────────────┐ ┌──────────────────────┐
 │  STATE: APPROVED     │ │  STATE: REJECTED     │
 │  ApprovalDecision{   │ │  ApprovalDecision{   │
 │   Approved: true     │ │   Approved: false,   │
 │  }                   │ │   Rationale: "..."   │
 │                      │ │  }                   │
 │  → WriteManagedBlock │ │  → RejectionRecorder │
 │    Gated writes      │ │    + A7 registry     │
 │  → lineage "applied" │ │    + lineage         │
 └──────────────────────┘ │      "rejected"      │
                          └──────────────────────┘
```

The dispatch does NOT call `AskUserQuestion` (subagent boundary, REQ-HEV3-031).
The dispatch returns a `ProposalArtifact` (the proposal + the target surface +
the L1-L4 verdict); the orchestrator runs the round and re-enters the dispatch
with the `ApprovalDecision`.

## §F. Alternatives Considered (and rejected)

### F.1 Alternative — A7 registry as a column in the existing routing-ledger

**Rejected**: the routing-ledger (EVOLVE-001) is an observation log (Loop 0).
The A7 registry is a Curator-loop artifact (Loop 2). Conflating them couples
the observation layer to the promotion layer (violates the 3-Loop separation).
A separate `.moai/state/negative-evidence.jsonl` keeps the concerns distinct.

### F.2 Alternative — L3 contradiction by parsing all Frozen rule files at runtime

**Rejected**: parsing every file under `.claude/rules/moai/**` on every L3
consult is expensive and fragile (rule files change shape). The typed
`FrozenRuleRegistry` (M0) is seeded from the L1 prefix list + A1 surfaces
and carries stable identifiers; runtime consult is a registry lookup, not a
file parse.

### F.3 Alternative — Autonomous write path for Tier 1-2 (no L5)

**Rejected**: the design SSOT §5 Loop 2 + "인간은 목표와 기준 제공" require L5
for every CLAUDE.md / CLAUDE.local.md write. Tier 1-2 lands in auto-memory
(which is NOT a CLAUDE.md write — it's a separate surface), so Tier 1-2 does
not need L5. But the moment a pattern reaches Tier 3+ (the CLAUDE.local.md /
CLAUDE.md surfaces), L5 is mandatory. REQ-HEV3-006 binds this.

### F.4 Alternative — Expand the meta-harness frozen_guard.go (not the safety one)

**Rejected**: the two frozen_guard.go files serve different pipelines. The
L1 safety pipeline (`safety.Pipeline.Evaluate`) consults
`safety/frozen_guard.go` (frozen_guard.go:34 `IsFrozen`). The meta-harness
path consults `harness/frozen_guard.go` (the `EnsureAllowed` check for
project-harness proposals). A Curator proposal flows through the SAFETY
pipeline (L1), so the safety frozen_guard.go is the correct expansion target.
The meta-harness guard is ALSO expanded for self-protection
(REQ-HEV3-025 — both guard source files are Frozen).

### F.5 Alternative — Canary veto cooldown reused as-is for A7

**Rejected**: the canary-veto cooldown is keyed by proposal (a single
promotion attempt); the A7 registry is keyed by pattern (a structural class
that may recur across many proposals). The semantics differ: the canary veto
blocks re-proposing the SAME promotion; the A7 registry blocks re-proposing
the same PATTERN with insufficient new evidence. Reusing the data structure
would conflate the two. The DURATION constant is reused; the data structure
is net-new (§D.3, §D.4).

### F.6 Alternative — GLM gate inside the curator writer

**Rejected**: the writer is the mechanical write primitive; the model-class
gate is a Curator-pipeline policy concern (REQ-HEV3-027). Putting the gate in
the writer couples the writer to the session model — the writer would need to
read `model_class`, which violates the "machine-signal-only at the writer
boundary" principle (the writer accepts typed `BlockContent`, not session
context). The gate lives in the dispatch layer upstream.

## §G. Forward-Looking Considerations

### G.1 EVOLVE-004 dependencies on this SPEC

EVOLVE-004 (`/moai harness evolve | promote | demote | freeze | unfreeze`
console verbs) will:
- Drive the dispatch (M5) via CLI verbs.
- Surface the A7 registry in `status` / `doctor`.
- Provide the `moai harness rollback` verb that triggers `RestoreSnapshot` +
  A7 auto-register.

The dispatch + registry shipped here expose enough surface for EVOLVE-004 to
drive without further wiring-side changes.

### G.2 EVOLVE-005 dependencies on this SPEC

EVOLVE-005 (Recall wiring + typed parser + template deployment) will:
- Wire Phase −1 consumption of the digest layer (read the CLAUDE.md block).
- Wire Phase −1 ledger-layer search INCLUDING the A7 registry (the
  cross-check "has this pattern been rejected before?").
- Verify the auto_detection empty-schema template deployment.

The A7 registry shipped here is a ledger-layer surface EVOLVE-005 will grep.

### G.3 The A8 horizon (evolutionary exploration)

Design SSOT §1 A8 / §7 horizon v6 explicitly defers evolutionary exploration
(multi-candidate mutation/selection) until single-candidate promotion (this
SPEC) is stable with 0 regression. This SPEC does not introduce multi-candidate
promotion; A8 remains deferred.

## §H. Open Architecture Questions (NEEDS CLARIFICATION inputs)

These are the architecture-level inputs to the plan.md §I NEEDS CLARIFICATION
items:

1. **H-1 A7 cooldown duration + N default** — architecturally neutral
   (parameter values); the choice affects M1 tests only.
2. **H-2 auto_detection bounds** — hardcoded Go struct vs schema file;
   architecturally minor (both are in-process validators).
3. **H-3 Frozen-rule registry scope** — typed registry vs file glob;
   architecturally consequential for M0 (a glob is cheaper to ship but a
   typed registry carries stable identifiers REQ-HEV3-017 requires).
4. **H-4 dispatch entry point** — `ApplyCurator()` on Applier vs separate
   `curator.Dispatch` type vs `Apply()` switch; architecturally consequential
   for M5 (the Applier is PRESERVE for the existing path; a separate type is
   the conservative default).

All four are orchestrator-mediated (AskUserQuestion before Implementation
Kickoff Approval). The defaults in spec.md / plan.md are the conservative path.
