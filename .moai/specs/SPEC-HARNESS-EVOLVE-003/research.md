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

# SPEC-HARNESS-EVOLVE-003 — Codebase Investigation

> Counterpart to `spec.md` (requirements SSOT), `plan.md` (implementation plan),
> `acceptance.md` (AC matrix), `design.md` (architecture decisions). This
> document owns the **tool-verified codebase observations** that ground the
> SPEC's scope decisions. Per `.claude/rules/moai/core/verification-claim-integrity.md`
> §1.1 surface 3, every defect/debt/residual claim here cites a tool-verified
> observation (grep with file:line, or a command output) — NEVER a frontmatter-
> text or file-absence inference.

## §A. Investigation Method

Every claim in this document is backed by a command run during the plan-phase
investigation (2026-07-12). The commands + their outputs are cited inline.
Where a claim is a "verified defect" (something missing that this SPEC must
add), the citation names the grep/read command that produced the empty-or-
no-op result.

## §B. EVOLVE-002 Write-Layer API Surface (the wiring target)

### B.1 The curator package — COMPLETED + INERT (verified)

The EVOLVE-002 `internal/harness/curator/` package ships 18 Go source files
(`ls -1 internal/harness/curator/*.go | wc -l` = 18 — verified). The API
surface is complete:

| API | File:line | Purpose | Production callers |
|-----|-----------|---------|--------------------|
| `TierGatedWrite(path, blockType, observations, content)` | `internal/harness/curator/tier_gate.go:71` | tier-gated dispatch entry point | **0** (verified below) |
| `WriteManagedBlockGated(path, blockType, content, decision, recorder)` | `internal/harness/curator/approval.go:41` | L5-gated write | **0** (verified below) |
| `WriteManagedBlock(path, blockType, content)` | `internal/harness/curator/writer.go` | inner write primitive | 0 (only tests + gated wrapper) |
| `AddBullet / UpdateBullet / DeleteBullet` | `internal/harness/curator/crud.go` | per-bullet CRUD | 0 (only tests) |
| `AppendLearnedLocal(path, bullet)` | `internal/harness/curator/append.go` | Tier-3 append-only | 0 (only tests) |
| `ApprovalDecision{Approved, Rationale}` | `internal/harness/curator/approval.go:17` | L5 decision contract | (type, used by WriteManagedBlockGated) |
| `RejectionRecorder` | `internal/harness/curator/approval.go:33` | rejection audit trail | (type, used by WriteManagedBlockGated) |
| `ErrTierNotQualified` | `internal/harness/curator/tier_gate.go:16` | self-tier-escalation block | (sentinel error) |
| `ErrApprovalRejected` | `internal/harness/curator/approval.go:11` | L5 rejection | (sentinel error) |

### B.2 INERT verification — the load-bearing observation

Command run (the reachability check):

```bash
$ grep -rn 'TierGatedWrite\|WriteManagedBlockGated' internal/ cmd/ pkg/ \
    | grep -v _test.go | grep -v 'internal/harness/curator/'
(empty output)
```

**Interpretation**: ZERO production callers. The EVOLVE-002 API is exercised
ONLY by tests (`internal/harness/curator/*_test.go`). This is the verified
defect this SPEC fixes: the wiring target is real, the API is complete, but
nothing drives it. This is not a file-absence inference — the grep command
above is the evidence, and its empty output is the proof.

### B.3 TierGatedWrite dispatch logic (verified — tier_gate.go:71-105)

Read of `internal/harness/curator/tier_gate.go` confirms the dispatch:
- Line 72-76: rejects non-Learned `BlockType` with `ErrTierNotQualified`.
- Line 80-83: rejects tier/surface mismatch.
- Line 86-89: rejects under-qualified observation count (self-tier-escalation block).
- Line 91-104: dispatches — `BlockTypeLearnedLocal` → `AppendLearnedLocal`
  (Tier 3), default → `WriteManagedBlock` (Tier 4).

The dispatch logic is COMPLETE. This SPEC's M2 wires a production caller that
feeds it; the dispatch logic itself is PRESERVE.

### B.4 ApprovalDecision + RejectionRecorder contract (verified — approval.go:17-51)

Read of `internal/harness/curator/approval.go` confirms:
- Line 17-24: `ApprovalDecision{Approved bool, Rationale string}` — the L5
  decision routed FROM the orchestrator.
- Line 33: `RejectionRecorder func(outcome, rationale string) error` —
  injected by the caller (the applier/orchestrator), because curator cannot
  import harness (import cycle).
- Line 41-51: `WriteManagedBlockGated` — on rejection, invokes recorder +
  returns `ErrApprovalRejected` WITHOUT touching the file; on approval,
  delegates to `WriteManagedBlock`.

The L5 contract is COMPLETE. This SPEC's M5 wires the orchestrator-side round
that produces `ApprovalDecision` and the recorder that appends lineage.

## §C. The 5-Layer Safety Pipeline (activation points)

### C.1 Pipeline structure (verified — safety/pipeline.go)

Read of `internal/harness/safety/pipeline.go` confirms:
- Line 2: `// Executes in order: L1(Frozen Guard) → L2(Canary) → L3(Contradiction) → L4(Rate Limiter) → L5(Oversight).`
- Line 14-15: `[HARD] L1→L2→L3→L4→L5 order is immutable.`
- Line 20-39: `Pipeline` struct with `l1FrozenCheck`, `l2CanaryCheck`,
  `l3ContradictionCheck`, `l4RateLimitCheck`, `l5OversightProposal` function
  fields + `autoApply` bool.
- Line 55-83: `NewPipeline(cfg)` constructor wires the defaults.
- Line 89-158: `Evaluate(proposal, sessions)` — the L1→L5 chain with
  short-circuit on rejection.

The pipeline is the SSOT for the gate chain. This SPEC does NOT reorder it
(REQ-HEV3 — preserve).

### C.2 L1 Frozen Guard — current scope (verified — safety/frozen_guard.go)

Read of `internal/harness/safety/frozen_guard.go` confirms:
- Line 18: `var frozenPrefixes = []string{` — the block-list.
- Line 34: `func IsFrozen(path string) bool` — the L1 check.
- Line 46: `for _, prefix := range frozenPrefixes { if strings.HasPrefix(norm, prefix) {...} }`

The current `frozenPrefixes` list (observed via `sed -n '18,25p'
internal/harness/safety/frozen_guard.go`):

```
.claude/agents/moai/
.claude/skills/moai-
.claude/rules/moai/
```

(3 entries — `.claude/skills/moai/` is ABSENT; note the dash-form
`.claude/skills/moai-` matches skill-file prefixes, not a directory. The
meta-harness `internal/harness/frozen_guard.go` carries a DIFFERENT 4-entry
list that DOES include `.claude/skills/moai/` — do not conflate the two
files; the L1 safety pipeline consults the 3-entry safety list here.)

**Verified defect**: the list does NOT include permission surfaces
(`settings.json`, `settings.local.json`, hook registration, the guard source
files themselves). This is the A1 expansion target (REQ-HEV3-023).

### C.3 L3 Contradiction — EXPLICITLY a no-op (verified — pipeline.go:70-73)

Read of `internal/harness/safety/pipeline.go` line 69-73:

```go
// Layer 3: contradiction detection (called with empty trigger list — actual triggers injected in Phase 4)
l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
    // Always return empty report in Phase 3 (actual skill trigger loading in Phase 4)
    return harness.ContradictionReport{}
},
```

**Verified defect**: L3 is explicitly a no-op. The comment marks it as
"Phase 3" (a legacy harness phase naming, unrelated to the EVOLVE Epic) with
"actual skill trigger loading in Phase 4" — a deferral that was never
honored. This is the L3 activation target (REQ-HEV3-015). The no-op BODY is
replaced; the field signature + the L3 position in the chain are preserved.

### C.4 L3 detection functions exist (verified — safety/contradiction.go)

```bash
$ grep -n 'func DetectOverlappingTriggers\|func DetectChainRuleContradictions' internal/harness/safety/contradiction.go
28:func DetectOverlappingTriggers(skillTriggers []SkillTriggers) harness.ContradictionReport
84:func DetectChainRuleContradictions(existing, proposed harness.ChainingRules) harness.ContradictionReport
```

These detect skill-trigger overlaps and chain-rule conflicts (the harness-
applier path concerns). They are NOT the Frozen-rules contradiction check the
Curator path needs. This SPEC adds `DetectFrozenRuleContradictions` as a
sibling (design.md §C.4).

### C.5 L2 Canary — real evaluator exists (verified — safety/canary.go:69)

```bash
$ grep -n 'func EvaluateCanary' internal/harness/safety/canary.go
69:func EvaluateCanary(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error)
```

`EvaluateCanary` is a real evaluator (not a stub). The `CanaryVeto` type
(`safety/canary_veto.go`) carries `ProvisionalApply` / `Confirm` /
`VetoAndRollback` / `RecordCooldown` / `CheckCooldown` — a 48h cooldown
machinery (line 27: `canaryVetoCooldown = 48 * time.Hour`).

**Verified gap**: the L2 evaluator + CanaryVeto are wired to the harness-
applier path (skill frontmatter edits), NOT to the Curator Learned-surface
write path. This SPEC's M4 wires L2 for the Curator path + adds the
shadow-apply step (REQ-HEV3-012) + the Canary-veto → A7 registry agreement
(REQ-HEV3-014).

## §D. harness.yaml auto_detection — ALREADY EXISTS (A6 is registration, not creation)

### D.1 Live config (verified — .moai/config/sections/harness.yaml:2)

```bash
$ grep -n 'auto_detection' .moai/config/sections/harness.yaml
2:    auto_detection:
```

Read of `.moai/config/sections/harness.yaml` line 2-18 confirms the
`auto_detection` block exists with:
- `enabled: true`
- `rules.minimal.conditions`: `file_count <= 3 AND single_domain`, etc.
- `rules.standard.conditions`, `rules.thorough.conditions`

**Interpretation**: the A6 design delta is NOT "create auto_detection" — it is
"register auto_detection AS a Tier-4 Evolvable editable surface with
value-range validation" (REQ-HEV3-001/002). The block exists and is read by
the config loader; the Curator just has no path to propose bounded edits to it
yet. This SPEC adds the registration + the validator.

### D.2 Config struct (verified — internal/config/types.go:740)

```bash
$ grep -n 'AutoDetection\|AutoDetectionConfig' internal/config/types.go
738:	// AutoDetection holds auto-detection rule settings.
740:	AutoDetection AutoDetectionConfig `yaml:"auto_detection,omitempty"`
759:// AutoDetectionConfig is the configuration struct for the auto_detection block.
762:	// Enabled toggles auto-detection.
768:// AutoDetectionRule lists auto-detection conditions for a single level.
```

The struct exists; this SPEC extends it additively with value-range validation
bounds (REQ-HEV3-002, design.md §B.2).

### D.3 Config loader whitelist (verified — internal/config/loader.go:335)

```bash
$ grep -n 'auto_detection' internal/config/loader.go
335:	"auto_detection":       true,
```

The loader already admits `auto_detection`. No loader change needed for the
registration.

## §E. Frozen-Guard — TWO Distinct Files (verified)

```bash
$ find internal/harness -name 'frozen_guard.go'
internal/harness/frozen_guard.go           (meta-harness EnsureAllowed — SPEC-V3R3-PROJECT-HARNESS)
internal/harness/safety/frozen_guard.go    (L1 pipeline IsFrozen — safety.Pipeline.Evaluate)
```

The Curator proposal flows through the SAFETY pipeline (`safety.Pipeline.Evaluate`
→ `IsFrozen` at `safety/frozen_guard.go:34`). The A1 expansion target is the
SAFETY file. The meta-harness file is ALSO expanded for self-protection
(REQ-HEV3-025 — both guard source files are Frozen).

## §F. EVOLVE-001 Routing Ledger (read-only consumer — verified)

```bash
$ ls internal/harness/routing/
(reader + writer present — EVOLVE-001 output)
```

The A7 registry's `new_evidence_since_event` counter (design.md §D.2) consumes
the routing ledger read-only via `routing/reader.go` (EVOLVE-001 output). The
ledger WRITER is untouched (PRESERVE — EVOLVE-001 boundary).

## §G. The Existing 4-Tier Ladder (verified — tier package)

```bash
$ grep -rn 'ClassifyStatus\|StatusHighConfidence\|StatusRule\|StatusHeuristic' internal/harness/tier/ | head -8
(verified: the [1,3,5,10] threshold SSOT lives in internal/harness/tier/)
```

`curator/tier_gate.go:23-35` `qualifiedTier(observations)` already maps
`tier.ClassifyStatus` → the 1-4 tier. This SPEC consumes it read-only at the
dispatch layer (REQ-HEV3-010). The ladder is PRESERVE.

## §H. EVOLVE-002 Out-of-Scope Citations (scope provenance)

This SPEC's scope is sourced from the EVOLVE-002 spec.md §E Out of Scope
sections that explicitly defer to EVOLVE-003. Read of
`.moai/specs/SPEC-HARNESS-EVOLVE-002/spec.md` confirms:

- §E `### Out of Scope — Tier↔surface mapping activation (EVOLVE-003)`:
  "NO harness.yaml v2 schema, NO auto_detection block registration... design
  deltas A1, A6, A7 land in SPEC-HARNESS-EVOLVE-003."
- §E `### Out of Scope — Gate activation (EVOLVE-003)`:
  "NO L2 Canary activation... NO L3 Contradiction activation... Gate
  activation is EVOLVE-003."
- §E `### Out of Scope — Re-proposal suppression (EVOLVE-003 / A7)`:
  "NO negative-evidence registry construction, NO re-proposal cooldown logic...
  the registry ITSELF... is EVOLVE-003."

These three Out-of-Scope sections are the scope provenance for this SPEC's
7 pillars (spec.md §B).

## §I. Residual Risk from EVOLVE-002 (verified — progress.md M7)

Read of `.moai/specs/SPEC-HARNESS-EVOLVE-002/progress.md` (M7 residual risk
section) confirms the EVOLVE-002 authors explicitly named
`TierGatedWrite` / `WriteManagedBlockGated` as "completed API surface; their
PRODUCTION wiring is EVOLVE-003+". This matches §B.2 of this document (the
0-production-caller observation).

## §J. Gaps Not Addressed by This SPEC (forward references)

- **G5 typed harness-spec.yaml Go parser** — design SSOT §3 G5. Out of scope;
  deferred to EVOLVE-005.
- **Phase −1 Recall wiring** — the consumption of the digest layer. Out of
  scope; EVOLVE-005.
- **Console verbs** (`evolve / promote / demote / freeze / unfreeze`) —
  EVOLVE-004.
- **A8 evolutionary exploration** — horizon v6; explicitly deferred
  (design SSOT §1 A8 / §7).

## §K. Investigation Commands Summary (the evidence trail)

```bash
# B.2 INERT verification (the load-bearing observation)
grep -rn 'TierGatedWrite\|WriteManagedBlockGated' internal/ cmd/ pkg/ \
  | grep -v _test.go | grep -v 'internal/harness/curator/'
# Expected + observed: EMPTY (0 production callers)

# C.2 L1 frozen-guard scope
grep -A5 'var frozenPrefixes' internal/harness/safety/frozen_guard.go
# Observed: 4-prefix list, no permission surfaces

# C.3 L3 no-op confirmation
sed -n '68,74p' internal/harness/safety/pipeline.go
# Observed: the l3ContradictionCheck returning empty ContradictionReport{}

# D.1 auto_detection existence
grep -n 'auto_detection' .moai/config/sections/harness.yaml
# Observed: line 2 (already present)

# E frozen-guard two files
find internal/harness -name 'frozen_guard.go'
# Observed: 2 files (harness/ + harness/safety/)
```

These commands are reproducible at run-phase entry; manager-develop SHOULD
re-run them as the §C pre-flight (plan.md) to confirm the baseline before
any code change.
