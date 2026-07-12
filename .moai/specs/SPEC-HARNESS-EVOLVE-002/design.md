---
id: SPEC-HARNESS-EVOLVE-002
title: "Curator Editable Surfaces — Loop 2 (write layer) of the self-evolving harness"
version: "0.1.0"
status: draft
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness, internal/harness/curator, internal/template, internal/template/templates"
lifecycle: spec-anchored
tags: "harness-evolve-epic, curator, managed-block, learned-workflow, snapshot-rollback, lineage, 3-zone, recall-contract, template-first"
era: V3R6
tier: L
depends_on: [SPEC-HARNESS-EVOLVE-001]
---

# SPEC-HARNESS-EVOLVE-002 — Design Decisions

> Counterpart to `spec.md` (requirements SSOT), `plan.md` (implementation
> approach), `acceptance.md` (AC matrix). This document owns the architecture
> decisions, alternatives considered, and the rationale for the chosen path.
> It is the SSOT for "why this shape, not the other shape".

## §A. Architecture Context

### A.1 The 3-Loop harness (EVOLVE Epic structure)

The HARNESS-EVOLVE Epic implements Lilian Weng's "Harness Engineering for
Self-Improvement" thesis as a 3-Loop architecture (design SSOT §5):

- **Loop 0 (Generator / observation)** — every routing decision observed and
  recorded. **SHIPPED**: `SPEC-HARNESS-EVOLVE-001` (routing-ledger.jsonl +
  Go writer/reader + Stop-hook outcome capture).
- **Loop 2 (Curator / promotion write)** — distilled workflow knowledge
  written to permitted surfaces. **THIS SPEC**: the write-path machinery.
- **Loop 1 (Reflector / aggregation)** — pattern aggregation across
  observations, tier-count tracking. **Partially shipped**: the existing
  `internal/harness/learner.go` performs pattern aggregation; full Loop 1
  activation (with negative-evidence cross-check) is folded into
  EVOLVE-003 alongside the gates and re-proposal suppression.

EVOLVE-002 sits between EVOLVE-001 (provides the observation input the
Reflect/curate loops consume) and EVOLVE-003 (activates the gates that
*guard* the write surfaces this SPEC builds). The Curator pipeline is:

```
observation (EVOLVE-001 ledger)
   │
   ▼
aggregation (existing learner.go)
   │
   ▼
tier-qualified proposal (Tier 3 → CLAUDE.local.md, Tier 4 → CLAUDE.md)
   │
   ▼
L5 AskUserQuestion approval (orchestrator-mediated, REQ-HEV2-032)
   │
   ▼
typed writer (this SPEC: WriteManagedBlock + bullet CRUD)
   │
   ├──► snapshot (extended createSnapshot, REQ-HEV2-021)
   ├──► write (CLAUDE.md managed block OR CLAUDE.local.md append-only)
   ├──► lineage entry (extended WriteLineageEntry, REQ-HEV2-023)
   └──► ledger-key cross-link (digest layer ↔ ledger layer)
```

### A.2 The 3-Zone edit-surface contract (design SSOT §4)

This SPEC operates entirely in the **Learned Zone** (the third zone,
alongside Frozen and Evolvable):

| Zone | Surfaces | Writer | This SPEC's interaction |
|------|----------|--------|-------------------------|
| Frozen | rules, agents, evaluators, permission surfaces, hooks | human + SPEC only | NONE (Frozen-guard expansion is EVOLVE-003) |
| Evolvable | harness-* skill frontmatter, `.claude/agents/harness/`, `auto_detection` block | Curator (Tier 4) | NONE (Evolvable surface changes are EVOLVE-003) |
| **Learned** | digest layer (CLAUDE.md block + CLAUDE.local.md append) + ledger layer (routing-ledger.jsonl + lineage + neg-evidence registry + auto-memory) | Curator (tier-differentiated) | **THIS SPEC** — builds both digest surfaces + formalizes the ledger-layer contract |

The permission axis (design delta A1) — settings.json, permission mode, hook
registration, frozen-guard itself — is Frozen and is NOT touched by this
SPEC. Permission-surface Frozen registration is EVOLVE-003.

### A.3 The 2-layer Recall contract (design delta A3)

The Recall contract is the policy that prevents context-window exhaustion
from the harness's own learning. It splits the "what the harness knows"
into two layers with different loading semantics:

```
┌──────────────────────────────────────────────────────────────┐
│  DIGEST LAYER (always-loaded, ≤3K chars, ≤20 bullets)        │
│  Surface: CLAUDE.md MOAI:LEARNED-WORKFLOW block              │
│  Loaded: automatically (it's in CLAUDE.md)                   │
│  Content: distilled generic workflow summaries               │
│  Linkage: each bullet carries a ledger_key ↓                 │
└──────────────────────────┬───────────────────────────────────┘
                           │  (cross-layer linkage, REQ-HEV2-017)
                           ▼
┌──────────────────────────────────────────────────────────────┐
│  LEDGER LAYER (on-demand grep search)                        │
│  Surfaces:                                                   │
│   - .moai/state/routing-ledger.jsonl (EVOLVE-001)            │
│   - .moai/state/<learning-history>/manifest.jsonl (lineage)  │
│   - negative-evidence registry (EVOLVE-003 — registry itself)│
│   - ~/.claude/projects/<hash>/memory/ (auto-memory)          │
│  Loaded: only on explicit Phase −1 grep search               │
│  Content: full evidence, history, failure trajectories       │
└──────────────────────────────────────────────────────────────┘
```

The contract codifies "remember everything ✗, search when needed ○"
(design SSOT §1 / §5 A3). The digest is the *summary*; the ledger is the
*detail*. The cross-layer `ledger_key` linkage (REQ-HEV2-010, REQ-HEV2-017)
makes the contract navigable — a Phase −1 actor that reads a digest bullet
can grep the ledger for the underlying evidence via the key.

This contract is *formalized* (types + godoc) by this SPEC. The
*consumption wiring* (Phase −1 router-side digest read + conditional
ledger grep) is `SPEC-HARNESS-EVOLVE-005`.

## §B. Typed Managed-Block Writer — Design Decision

### B.1 Generalize, don't rewrite

The existing `internal/harness/layer3.go InjectMarker` is the SOLE
mechanical CLAUDE.md write path in the codebase. It has a single purpose:
inject the `## Project-Specific Configuration (Harness-Generated)` block.
It is well-tested (`layer3_test.go`), production-stable, and idempotent.

**Decision**: generalize `InjectMarker` into a typed writer API
(`WriteManagedBlock(path, blockType, content)`) rather than authoring a
parallel writer from scratch. The generalization:

1. Extracts the marker-block regex into a per-`BlockType` registry (each
   type maps to its heading + start/end markers — see spec.md §D.2).
2. Preserves `InjectMarker` as a thin wrapper over
   `WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)` for backward
   compatibility.
3. Adds `BlockTypeLearnedWorkflow` as the new type with its own marker
   contract.

**Rationale**:
- Reuse of the existing well-tested regex / atomic-replace logic minimizes
  risk (the Brownfield strategy per the DDD workflow).
- Backward compatibility preserves every existing `InjectMarker` caller
  byte-identical (no behavior change to the existing Harness-Generated
  block).
- A single typed writer makes future managed-block types (e.g. a future
  EVOLVE-003 `auto_detection` Tier-4 surface) a matter of adding an enum
  value + marker triplet, not authoring a new writer.

### B.2 Per-bullet CRUD as the default interface

The design SSOT §2 states: "Curator is forbidden from full rewrite, bullet
CRUD only" (context-collapse avoidance). This SPEC codifies the constraint:

- The default Curator interface is `AddBullet` / `UpdateBullet` /
  `DeleteBullet` (REQ-HEV2-007). Each locates the bullet by `ledger_key`
  and rewrites only the targeted line.
- A `ReplaceAllBullets` operation exists for the rare wholesale-rewrite
  path (e.g. a migration that re-keys every bullet), but it is NOT the
  default. The lint rule (EVOLVE-003 territory) will flag Curator pipelines
  that call `ReplaceAllBullets` more than once per N observations.

**Rationale**: a full block rewrite on every Curator run causes context-
collapse for the human reader (every bullet appears "new" even when only
one changed). Bullet-level CRUD preserves the stable bullets' identity and
makes the diff reviewable at L5 approval time.

### B.3 Marker contract — heading + start/end HTML comments

The marker contract is consistent with the existing `InjectMarker` pattern
(heading + HTML-comment start/end markers as an atomic match group). The
new types:

| Surface | Heading | Markers |
|---------|---------|---------|
| Digest (Tier 4) | `## MOAI:LEARNED-WORKFLOW` | `<!-- moai:learned-start -->` / `<!-- moai:learned-end -->` |
| Append-only (Tier 3) | `## MOAI:LEARNED-WORKFLOW-LOCAL` (NEEDS CLARIFICATION H-1) | `<!-- moai:learned-local-start -->` / `<!-- moai:learned-local-end -->` |
| Legacy (backward compat) | `## Project-Specific Configuration (Harness-Generated)` | `<!-- moai:harness-start ... -->` / `<!-- moai:harness-end -->` |

The HTML-comment form is chosen because:
- Markdown renderers ignore HTML comments (the marker is invisible in
  rendered output).
- Comments survive `moai update` `mergeSectionBased` as opaque text.
- The form is consistent with the existing `InjectMarker` (no new
  syntactic class to learn).

### B.4 Anti-fabrication input validation

The writer enforces the anti-fabrication principle (inherited from
EVOLVE-001 §A) at the input boundary. Bullet text matching the §25
forbidden-classes regex is rejected:

```go
var forbiddenPatterns = []*regexp.Regexp{
    regexp.MustCompile(`SPEC-[A-Z][A-Z0-9]+-[0-9]{3}`),  // internal SPEC IDs
    regexp.MustCompile(`(REQ|AC)-[A-Z][A-Z0-9]+-[0-9]{3}`),  // internal tokens
    regexp.MustCompile(`20[0-9]{2}-[0-1][0-9]-[0-3][0-9]`),  // ISO dates (internal)
    regexp.MustCompile(`\b[0-9a-f]{7,40}\b`),  // commit SHAs (7-40 hex chars)
}
```

(Refinement needed at run-phase: the SHA regex is conservative and may
false-positive on legitimate prose. Run-phase will tighten it to require
the SHA to appear in a "commit-like" context — e.g. preceded by "commit"
or "sha" — per the §25.3 D-007 inline-resolution precedent.)

**Rationale**: enforcing at the writer boundary prevents the Curator from
ever persisting a forbidden pattern, regardless of which model proposed
the bullet. The regex is the writer's invariant, not the model's
self-discipline.

## §C. Snapshot / Rollback / Lineage Extension — Design Decision

### C.1 Distinct restore units per surface

The existing `createSnapshot` records a single file per snapshot dir
(`proposal.TargetPath`). This SPEC extends it to record distinct restore
units per surface:

- A CLAUDE.md managed-block write snapshots the CLAUDE.md file (whole
  file — atomic restore).
- A CLAUDE.local.md append snapshots the CLAUDE.local.md file (whole file).
- The snapshot manifest gains entries like:
  ```json
  {
    "learned_surface": "claude.md.learned-workflow",
    "original_path": "CLAUDE.md",
    "byte_length_pre_write": 26432,
    "bullets_affected": ["X", "Y"]
  }
  ```

**Rationale**: distinct restore units allow a Tier-4 rollback without
reverting a concurrent Tier-3 append (they are independent surfaces).
Treating them as a single unit would force a coupled rollback that loses
unrelated writes.

### C.2 Byte-identical rollback via pre-write byte-length

The post-rollback integrity check uses byte-length comparison:

```go
if len(postRestoreBytes) != manifest.ByteLengthPreWrite {
    return ErrRollbackIntegrityFailed
}
```

This is a stronger check than "file exists" — it catches partial writes,
encoding drift, and concurrent-session interference.

### C.3 LineageEntry extension (additive)

The existing `LineageEntry` is extended additively (existing fields
preserved verbatim):

```go
type LineageEntry struct {
    // ...existing fields preserved verbatim...
    LearnedSurface  string   // new: "claude.md.learned-workflow" | ...
    BulletsChanged  []string // new: ledger_keys of bullets add/upd/del
    SnapshotDir     string   // new: pointer to the snapshot restore unit
}
```

The extension is additive — existing lineage readers (e.g.
`LoadManifest`) continue to parse legacy entries unchanged. The new fields
default to zero values (`""` / `nil`) for legacy entries.

## §D. `mergeSectionBased` Preservation — Design Decision

The existing `internal/merge/strategies.go mergeSectionBased` parses
markdown by `## ...` headings and merges sections independently. The
decision: how does the merge recognize the new managed block as a
*preserved* section vs a regular mergeable section?

### D.1 Option (a) — Explicit registration (default per spec.md)

The merge carries a registered list of managed-section headings:

```go
var managedSectionHeadings = []string{
    "## Project-Specific Configuration (Harness-Generated)",
    "## MOAI:LEARNED-WORKFLOW",
    "## MOAI:LEARNED-WORKFLOW-LOCAL",  // per NEEDS CLARIFICATION H-1
}
```

Sections in this list are treated as opaque preserved units: when the
upstream/local conflict inside the marker boundaries is non-trivial, the
merge surfaces a conflict rather than auto-resolving.

**Pro**: explicit allow-list; adding a new managed type requires updating
the registration (forces conscious review).
**Con**: registration maintenance — each new managed block type needs a
new entry.

### D.2 Option (b) — Auto-recognition via marker pair

Any section whose body is bounded by a `<!-- moai:*-start -->` /
`<!-- moai:*-end -->` marker pair is auto-recognized as a managed section.

**Pro**: no registration maintenance — any future managed type with the
marker-pair form is auto-recognized.
**Con**: couples merge behavior to marker syntax; a malformed marker pair
silently degrades the section to regular merge semantics.

### D.3 Default + NEEDS CLARIFICATION

The spec.md default leans (a) for conservatism (explicit allow-list).
NEEDS CLARIFICATION H-2 in plan.md surfaces this to the orchestrator's
AskUserQuestion round. The chosen option is non-blocking — either works,
and the writer API treats the merge as a black box.

## §E. Alternatives Considered (and rejected)

### E.1 Alternative — Author a parallel writer from scratch

**Rejected**: a parallel writer would duplicate the well-tested regex /
atomic-replace logic of `InjectMarker`, doubling the maintenance surface.
The generalize-don't-rewrite decision (§B.1) is preferred.

### E.2 Alternative — Single-surface writer (CLAUDE.md only)

**Rejected**: the design SSOT §4 Learned-zone row specifies TWO digest
surfaces (CLAUDE.md managed block + CLAUDE.local.md append-only). They
serve different tiers (4 vs 3) with different evidence thresholds (10 vs
5) and different write semantics (CRUD vs append-only). A single-surface
writer would collapse the tier differentiation (REQ-HEV2-025/026).

### E.3 Alternative — Store digest bullets in a separate JSON file

**Rejected**: the design SSOT §4 specifies the digest layer as the
always-loaded CLAUDE.md block, NOT a separate file. A separate JSON file
would require a new always-loaded surface (defeating the budget-control
intent — the digest layer IS the budget-controlled surface).

### E.4 Alternative — Full ledger in CLAUDE.md (no 2-layer split)

**Rejected**: this is the failure mode the 2-layer Recall contract (A3)
exists to prevent. The design SSOT §5 A3 / risk grid HIGH "CLAUDE.md
contamination / budget overflow" names this as the top risk. The 2-layer
split (digest summary in CLAUDE.md + full evidence in ledger) is the
mitigation.

### E.5 Alternative — No L5 approval gate (autonomous writes)

**Rejected**: the design SSOT §5 Loop 2 + "인간은 목표와 기준 제공"
require L5 approval for every CLAUDE.md / CLAUDE.local.md write. An
autonomous write path would be a reward-hacking surface (the model could
write its own report card). REQ-HEV2-032 binds this; AP-HEV2-003 names
the violation.

### E.6 Alternative — Wholesale block rewrite on every Curator run

**Rejected**: per design SSOT §2 "Curator is forbidden from full rewrite,
bullet CRUD only". Wholesale rewrite causes context-collapse for the L5
reviewer. §B.2 codifies per-bullet CRUD as the default.

### E.7 Alternative — Negative-evidence registry built in this SPEC

**Rejected**: the SSOT §7 M3 explicitly assigns the negative-evidence
registry + re-proposal cooldown to EVOLVE-003 (design delta A7). This
SPEC names only the *search-interface slot* in the Recall contract ledger
layer (REQ-HEV2-016); the registry itself is EVOLVE-003.

## §F. Forward-Looking Considerations

### F.1 EVOLVE-003 dependencies on this SPEC

EVOLVE-003 (tier↔surface mapping activation + gates + re-proposal
suppression) will:
- Activate the `auto_detection` Tier-4 surface (it's an Evolvable-zone
  surface, NOT a Learned-zone surface — this SPEC does not build it).
- Activate L2 Canary + L3 Contradiction gates (the writer API stays
  unchanged; the gates wrap the writer).
- Build the negative-evidence registry (this SPEC's Recall contract
  reserves the ledger-layer slot for it).
- Add the permission-surface Frozen-guard expansion (design delta A1).

The typed writer API in this SPEC is designed to be forward-compatible:
EVOLVE-003 can wrap the writer with gates without modifying the writer
itself (the writer is the inner write primitive; the gates are outer
decorators).

### F.2 EVOLVE-004 dependencies on this SPEC

EVOLVE-004 (`/moai harness evolve | promote | demote | freeze | unfreeze`
console verbs) will:
- Drive the typed writer via the Curator pipeline (the verbs are CLI
  surfaces over the writer API).
- Surface the lineage + snapshot history in `status` / `doctor`.

The writer API in this SPEC exposes enough surface (per-bullet CRUD,
snapshot directory pointer, lineage entry) for EVOLVE-004 to drive it
without further writer-side changes.

### F.3 EVOLVE-005 dependencies on this SPEC

EVOLVE-005 (Recall wiring + typed parser + template deployment) will:
- Wire Phase −1 digest-layer consumption (read the CLAUDE.md block).
- Wire Phase −1 ledger-layer search (grep routing-ledger.jsonl, lineage,
  neg-evidence registry).
- Verify the empty-marker template deployment.

The 2-layer Recall contract formalized in this SPEC (§C.4 of spec.md)
names the layers and the `ledger_key` linkage; EVOLVE-005 wires their
consumption. The split ensures this SPEC can ship the write layer
without coupling to the router-side read path.

### F.4 The GLM observation-only carve-out

The design SSOT §2 + STOP lesson note that GLM sessions are observation-
only in later loops — observation is always accepted, but Tier 3+
promotion *proposal generation* is Opus/Fable-only. This SPEC's writer
API is model-agnostic (it accepts a typed `BlockContent`); the model
gate lives in the Curator pipeline upstream of the writer. The writer
itself does not check `model_class`. This is correct: the writer is the
mechanical write primitive; the model gate is a Curator-pipeline policy
concern.

## §G. Open Architecture Questions (NEEDS CLARIFICATION inputs)

These are the architecture-level inputs to the plan.md §H NEEDS
CLARIFICATION items:

1. **H-1 marker naming** — the choice between `MOAI:LEARNED-WORKFLOW-LOCAL`
   vs `MOAI:LEARNED-LOCAL` vs the parenthetical form. Architecturally
   neutral (the writer accepts marker names per `BlockType`). The choice
   affects human readability only.
2. **H-2 mergeSectionBased recognition** — option (a) explicit
   registration vs option (b) marker-pair auto-recognition (§D above).
   Architecturally consequential for the merge code but not for the
   writer.
3. **H-3 debug CLI verb** — option (a) no CLI verb (Go-test-driven M7
   verification only) vs option (b) a minimal `moai harness curator
   debug-write` verb. Architecturally minor (a CLI verb is a thin wrapper
   over the writer API).

All three are orchestrator-mediated (AskUserQuestion before Implementation
Kickoff Approval). The defaults in spec.md are the conservative path.
