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

# SPEC-HARNESS-EVOLVE-002 — Curator Editable Surfaces (Loop 2 / write layer)

## HISTORY

| Date | Version | Change | Author |
|--------|---------|--------|--------|
| 2026-07-12 | 0.1.0 | Initial plan-phase draft (Tier L, 5 content artifacts: spec.md + plan.md + acceptance.md + design.md + research.md, plus progress.md §E skeleton). M2 of the HARNESS-EVOLVE Epic per the approved design SSOT `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (§4 3-Zone edit-surface contract — Learned-zone row line ~161; §5 Loop 2 + A3 two-layer Recall contract line ~112; §7 M2 milestone line ~270-272; §7 risk grid + success metrics). Implements the Loop 2 write layer: generalizes `internal/harness/layer3.go InjectMarker` into a typed managed-block writer; adds the `MOAI:LEARNED-WORKFLOW` digest block (CLAUDE.md, Tier 4), the append-only Learned section (CLAUDE.local.md, Tier 3), formalizes the 2-layer Recall contract (A3 — digest always-loaded + ledger on-demand grep), extends `mergeSectionBased` preservation, and extends snapshot/rollback/lineage to cover the new surfaces. Consumes EVOLVE-001's routing ledger as the ledger-layer input. 3 NEEDS-CLARIFICATION items tracked in plan.md §H. | manager-spec |

## §A. Context and Intent

`SPEC-HARNESS-EVOLVE-001` (completed, sync `f12ebc7ea` + backfill `b242450ed`) built
the **observation layer** of the self-evolving harness: the
`routing-ledger.jsonl` schema v1, the `internal/harness/routing/` Go writer/reader,
and the Stop-hook outcome capture. Observation alone changes nothing — the harness
still has no permitted write surface for the Curator (Loop 2). Design gap G2
("Curator edit surface limited to skill frontmatter 2 fields") and G3
("CLAUDE.md / CLAUDE.local.md evolution path absent") remain open (design SSOT §3).

This SPEC implements **Loop 2 (write layer)** of the 3-Loop self-evolving harness
(design SSOT §5, grounded in Lilian Weng's "Harness Engineering for
Self-Improvement"). It provides the typed, gated, snapshot-backed write surfaces
the Curator needs to persist distilled workflow knowledge — WITHOUT activating
the tier↔surface mapping, gate enforcement, or re-proposal suppression that
consumes those surfaces (those are `SPEC-HARNESS-EVOLVE-003`).

**Boundary principle.** This SPEC authors ONLY the write-path machinery:
- the typed managed-block writer (generalizing `InjectMarker`),
- the two Learned-zone surfaces (`MOAI:LEARNED-WORKFLOW` digest block +
  append-only CLAUDE.local.md Learned section),
- the 2-layer Recall *contract* naming (the digest-layer summary format and the
  ledger-layer search interface — the ledger itself is EVOLVE-001's routing
  package; the negative-evidence *registry* is EVOLVE-003),
- `mergeSectionBased` preservation of the managed block,
- snapshot / rollback / lineage extension to cover the new surfaces.

It performs NO tier promotion, NO gate activation, NO re-proposal suppression,
NO console verb addition — those are EVOLVE-003 / EVOLVE-004 territory
(see §E Exclusions).

Three anti-fabrication principles bind the design (inherited from EVOLVE-001 §A
+ `.claude/rules/moai/core/verification-claim-integrity.md` §1.1):

1. **Machine signals only** — Curator write authority and rollback triggers
   derive from mechanical state (tier counts, gate results, manual verbs),
   never from model self-report. The writer is invoked by the existing Curator
   pipeline, not by free-form model prose.
2. **Privacy / template neutrality** — the managed block carries distilled
   *generic* workflow knowledge; NEVER verbatim user text or internal SPEC IDs /
   REQ tokens / dates / commit SHAs. The block marker ships to
   `internal/template/templates/` so it MUST satisfy template-internal-content
   isolation (CLAUDE.local.md §25 — consult
   `.moai/docs/template-internal-isolation-doctrine.md`).
3. **Evidence-or-null** — digest bullets carry a `ledger_key` linking to a
   ledger-layer entry; where evidence is absent, the key is `null`, never
   an inferred value.

**L5 approval invariant.** Every CLAUDE.md / CLAUDE.local.md Curator write is
gated by L5 `AskUserQuestion` approval routed through the orchestrator (no
autonomous write path exists, no exception — design SSOT §5 Loop 2 + "인간은
목표와 기준 제공"). The writer is a subagent-side function; the approval channel
is the orchestrator's (subagent-boundary discipline per C-HRA-008).

## §B. Scope Summary

**In scope**:
- Typed managed-block writer package `internal/harness/curator/` generalizing
  the existing `internal/harness/layer3.go InjectMarker` into a typed
  `WriteManagedBlock(path, blockType, content)` API with a `BlockType` enum
  (`LEARNED_WORKFLOW` + `HARNESS_GENERATED` for backward-compat with the
  existing `## Project-Specific Configuration (Harness-Generated)` block).
- `MOAI:LEARNED-WORKFLOW` managed block in `CLAUDE.md` (Tier 4, Learned-zone
  digest layer): per-bullet CRUD operations (add / update / delete a single
  bullet without rewriting the whole block), ≤3,000-character digest-budget
  enforcement (reusing `internal/config/token_budget_guard.go`
  `measureAlwaysLoaded`), ≤20-bullet cap, `ledger_key` cross-layer linkage to
  the ledger layer.
- Append-only Learned section in `CLAUDE.local.md` (Tier 3): append-only writer
  with pattern-key deduplication.
- 2-layer Recall contract (A3) formalization: digest layer (always-loaded —
  CLAUDE.md block summary only) + ledger layer (on-demand grep search —
  `routing-ledger.jsonl`, lineage manifest, negative-evidence registry
  *search interface*, auto-memory). Principle: "remember everything ✗, search
  when needed ○" — detail lives in the ledger, only the summary is always-loaded.
- `mergeSectionBased` preservation extension: the new managed block is
  recognized as a preserved section during `moai update` so the merge does not
  clobber the block.
- Snapshot / rollback / lineage extension: the existing
  `internal/harness/applier.go createSnapshot` + `RestoreSnapshot` +
  `internal/harness/lineage.go WriteLineageEntry` machinery is extended to
  cover the new write surfaces as distinct restore units with byte-identical
  rollback.
- Template-First: the empty managed-block marker (heading + start/end markers +
  zero bullets) ships to `internal/template/templates/CLAUDE.md`; the writer Go
  code ships to `internal/harness/curator/`. Block CONTENT never ships.

**Preserve**:
- `internal/harness/layer3.go InjectMarker` — NOT deleted; the typed writer
  generalizes its mechanism. The existing `HARNESS_GENERATED` block type keeps
  the legacy `InjectMarker` behavior verbatim (backward compatibility —
  existing callers' behavior is preserved byte-identical).
- `internal/harness/routing/` (EVOLVE-001 output) — read-only consumption of
  the routing ledger by the Recall contract is fine; the ledger writer is
  untouched.
- `internal/harness/observer.go` + `.moai/harness/usage-log.jsonl` — untouched
  (separate observation surface, EVOLVE-001 REQ-HEV-009 discipline continued).
- The 5-layer safety pipeline (`layer1.go` … `layer5.go`) — this SPEC adds
  Curator write machinery but does NOT activate L2 Canary or L3 Contradiction
  (those activations are EVOLVE-003).
- `internal/merge/strategies.go mergeSectionBased` — extended to recognize the
  new managed block as a preserved section, NOT rewritten.
- Template neutrality (§25): the MECHANISM ships to templates (Go binary +
  empty block marker); the block DATA never does.

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 Typed Managed-Block Writer (generalize `InjectMarker`)

- **REQ-HEV2-001** (Ubiquitous): The `internal/harness/curator` package shall
  provide a typed managed-block writer
  `WriteManagedBlock(path string, blockType BlockType, content BlockContent) error`
  that reads the target file, locates the marker block matching `blockType`,
  replaces the block content atomically (or appends a fresh block when absent),
  and writes the file back with its pre-block and post-block bytes preserved
  verbatim.
- **REQ-HEV2-002** (Ubiquitous): The `BlockType` enum shall include at minimum
  `BlockTypeLearnedWorkflow` (the new `MOAI:LEARNED-WORKFLOW` block) and
  `BlockTypeHarnessGenerated` (the existing
  `## Project-Specific Configuration (Harness-Generated)` block — preserved
  for backward compatibility so existing `InjectMarker` callers continue to
  work through the typed writer).
- **REQ-HEV2-003** (Compound — idempotent): **While** a managed block of the
  given `blockType` already exists at `path`, **When** the writer is invoked
  with `content` byte-identical to the existing block body, the writer shall
  produce zero byte-diff on the file (idempotent re-application is a no-op).
- **REQ-HEV2-004** (Unwanted behavior — byte preservation): The writer shall
  not modify any byte outside the matched marker-block boundaries (the
  pre-block prefix and post-block suffix are preserved verbatim, including
  their newline structure).
- **REQ-HEV2-005** (Event-driven): **When** the target file lacks any marker
  block of the given `blockType`, the writer shall append the block with a
  separating blank line, ensuring the file's existing trailing newline
  convention is respected (no double-blank-line insertion, no missing
  newline at end of file).

### C.2 `MOAI:LEARNED-WORKFLOW` digest block (Tier 4, Learned-zone digest layer)

- **REQ-HEV2-006** (Ubiquitous — marker contract): The
  `BlockTypeLearnedWorkflow` block shall carry the marker contract: a
  `## MOAI:LEARNED-WORKFLOW` heading immediately followed by an HTML-comment
  start marker `<!-- moai:learned-start -->` and end marker
  `<!-- moai:learned-end -->`, with the bullet body between them. The heading
  + start/end markers form an atomic match group (consistent with the existing
  `markerBlockPattern` in `layer3.go`).
- **REQ-HEV2-007** (Ubiquitous — bullet CRUD): The typed writer shall expose
  per-bullet operations `AddBullet`, `UpdateBullet`, `DeleteBullet` that
  rewrite only the targeted bullet line (matched by `ledger_key`), preserving
  every other bullet in the block unchanged. A whole-block `ReplaceAllBullets`
  operation is ALSO provided for the rare wholesale-rewrite path, but the
  per-bullet operations are the default Curator interface (context-collapse
  avoidance — design SSOT §2 "Curator is forbidden from full rewrite, bullet
  CRUD only").
- **REQ-HEV2-008** (Compound — digest budget): **While** the would-be-written
  block content exceeds the 3,000-character digest budget (measured via the
  existing `internal/config/token_budget_guard.go` `measureAlwaysLoaded`
  function, which counts the CLAUDE.md-managed-block contribution to the
  always-loaded surface), **When** the writer evaluates the proposed write,
  the writer shall reject the write with a typed `ErrDigestBudgetExceeded`
  error and NOT touch the file.
- **REQ-HEV2-009** (Compound — bullet cap): **While** the proposed bullet
  count exceeds 20, **When** the writer evaluates the proposed write, the
  writer shall reject with a typed `ErrBulletCapExceeded` error and NOT touch
  the file.
- **REQ-HEV2-010** (Ubiquitous — ledger reference): Each digest bullet shall
  carry a `ledger_key` token in a trailing `<!-- key: ... -->` HTML comment
  (or equivalent stable location) linking the digest bullet to a ledger-layer
  entry — a routing-ledger pattern key, a lineage manifest SHA, or (once
  EVOLVE-003 ships it) a negative-evidence registry key. The token is the
  cross-layer linkage that makes the 2-layer Recall contract navigable.
- **REQ-HEV2-011** (Unwanted behavior — anti-fabrication): The block body
  shall not carry verbatim user request text, internal SPEC IDs
  (`SPEC-V3R6-*` / `SPEC-HARNESS-EVOLVE-*` etc.), REQ/AC tokens
  (`REQ-HEV2-*` / `AC-HEV2-*`), internal session dates (`2026-07-12` ISO
  form), or commit SHAs. The block carries *distilled generic workflow
  knowledge* only (template-internal-content isolation, §25). This binds the
  Curator's proposal content, the writer's input validation, and the empty
  marker shipped to the template tree.

### C.3 Append-only Learned section in `CLAUDE.local.md` (Tier 3)

- **REQ-HEV2-012** (Ubiquitous — append-only): The `CLAUDE.local.md` Learned
  section writer shall append new entries to the section without modifying or
  deleting any existing entry's bytes (append-only contract — Tier 3 is the
  permanent-record layer, distinct from the digest-layer CRUD in §C.2).
- **REQ-HEV2-013** (Ubiquitous — section marker contract): The
  `CLAUDE.local.md` Learned section shall carry the marker contract: a
  `## MOAI:LEARNED-WORKFLOW-LOCAL` heading + HTML-comment start/end markers
  `<!-- moai:learned-local-start -->` / `<!-- moai:learned-local-end -->`.
  *(Marker naming is a NEEDS CLARIFICATION candidate — see plan.md §H-1; the
  default proposed here matches the digest-block naming convention.)*
- **REQ-HEV2-014** (Compound — dedup): **While** an existing entry in the
  Learned section carries the same `ledger_key` as the proposed append,
  **When** the append is invoked, the writer shall no-op (deduplicated append
  returns `ErrDuplicateAppend` so the caller can distinguish from a
  successful fresh append; no bytes are written).

### C.4 2-Layer Recall Contract (A3) formalization

- **REQ-HEV2-015** (Ubiquitous — digest layer): The digest layer of the Recall
  contract SHALL comprise the always-loaded summary surface — the
  `MOAI:LEARNED-WORKFLOW` managed block in `CLAUDE.md` only. The digest layer
  is loaded into context automatically by virtue of being in `CLAUDE.md`;
  it carries summaries, never full evidence.
- **REQ-HEV2-016** (Ubiquitous — ledger layer): The ledger layer of the Recall
  contract SHALL comprise the on-demand grep-search surfaces:
  `.moai/state/routing-ledger.jsonl` (EVOLVE-001 output),
  `.moai/state/<learning-history>/manifest.jsonl` (lineage, this SPEC),
  negative-evidence registry (the registry *itself* is EVOLVE-003; this SPEC
  names only the *search interface* slot in the contract), and auto-memory
  (`~/.claude/projects/<hash>/memory/`). The ledger layer is consulted only
  on explicit Phase −1 grep search, never auto-loaded in full.
- **REQ-HEV2-017** (Compound — cross-layer linkage): **When** the Curator
  emits a digest bullet, the bullet **shall** carry a `ledger_key` linking to
  a ledger-layer entry; **While** no machine signal underpins the bullet
  (early-tier observation, evidence-or-null), the `ledger_key` shall be
  `null` and the bullet marked provisional. The Recall contract is navigable
  digest→ledger via this key, never the reverse (the ledger is the source of
  truth; the digest is the summary).
- **REQ-HEV2-018** (Ubiquitous — principle codification): The Recall contract
  codifies the principle "remember everything ✗, search when needed ○"
  (design SSOT §1 / §5 A3): the digest carries *summaries only*; detail,
  evidence, lineage, and failure history live in the ledger. The writer API
  enforces this — there is no `WriteFullEvidenceToDigest` code path.

### C.5 `mergeSectionBased` preservation

- **REQ-HEV2-019** (Compound — preservation): **While** `moai update` runs
  `internal/merge/strategies.go mergeSectionBased` over `CLAUDE.md`, **When**
  the upstream template tree carries an empty `MOAI:LEARNED-WORKFLOW` block
  marker and the local copy carries a populated block, the merge **shall**
  recognize the managed block as a preserved section and retain the local
  (populated) block verbatim when no upstream/local content conflict exists
  inside the marker boundaries.
- **REQ-HEV2-020** (Unwanted behavior — no silent clobber): The merge shall
  not silently empty or clobber the local populated managed block. A genuine
  upstream/local content conflict inside the marker boundaries is surfaced as
  a merge conflict (standard `mergeSectionBased` behavior), NOT auto-resolved
  by taking either side.

### C.6 Snapshot / rollback / lineage extension

- **REQ-HEV2-021** (Ubiquitous — snapshot extension): The existing
  `internal/harness/applier.go createSnapshot` machinery shall snapshot the
  CLAUDE.md managed block and the CLAUDE.local.md Learned section as *distinct
  restore units* — each surface has its own entry in the snapshot manifest,
  so a rollback can target one surface without reverting the other.
- **REQ-HEV2-022** (Unwanted behavior — byte-identical rollback): The
  `RestoreSnapshot` path shall produce a byte-identical restoration of the
  managed block / Learned section — the post-rollback bytes match the
  pre-write bytes exactly (no partial state, no bullet half-deleted, no
  marker orphaned). The snapshot manifest records the byte-length of each
  restore unit so post-rollback integrity is verifiable.
- **REQ-HEV2-023** (Ubiquitous — lineage recording): Each Curator write to a
  Learned surface SHALL append a `LineageEntry` to the existing
  `.moai/state/<learning-history>/manifest.jsonl` recording: target surface
  (`claude.md.learned-workflow` / `claude.local.md.learned-local`), bullets
  added/updated/deleted (by `ledger_key`), `ledger_key` refs, outcome
  (`applied` / `rolled-back` / `rejected`), and the snapshot directory
  pointer when applicable. The lineage is the audit trail that makes
  Loop 2 introspectable.
- **REQ-HEV2-024** (Unwanted behavior — evidence-or-null): The `LineageEntry`
  shall carry `null` in any evidence-reference field where no machine signal
  underpins the entry, never an inferred or fabricated value (inherits
  EVOLVE-001 REQ-HEV-006 evidence-or-null discipline).

### C.7 Tier-differentiated Curator writes

- **REQ-HEV2-025** (Capability gate — Tier 4): **Where** a pattern's evidence
  count reaches the Tier 4 threshold (default 10 observations, per the
  existing 4-tier learning ladder `[1,3,5,10]`), the Curator **shall** propose
  a CLAUDE.md managed-block write (the Tier 4 surface). The threshold itself
  is owned by the existing learner — this REQ binds only the surface
  selection, not the threshold value.
- **REQ-HEV2-026** (Capability gate — Tier 3): **Where** a pattern's evidence
  count reaches the Tier 3 threshold (default 5 observations), the Curator
  **shall** propose a CLAUDE.local.md append (the Tier 3 surface).
- **REQ-HEV2-027** (Unwanted behavior — no self-tier-escalation): The Curator
  shall not write to a higher tier than the pattern's evidence count
  qualifies for (a 6-observation pattern cannot go to Tier 4; it stays
  Tier 3). Self-tier-escalation is a reward-hacking shape and is blocked at
  the writer API (the writer accepts an explicit `tier` argument and rejects
  mismatches with `ErrTierNotQualified`).

### C.8 Template-First + template neutrality

- **REQ-HEV2-028** (Ubiquitous — Template-First): Every edit to the template
  tree (`internal/template/templates/CLAUDE.md`) shall be made Template-First
  — template source FIRST, then the live `.claude/CLAUDE.md` / `CLAUDE.md`
  copy, then `make build` to recompile embedded assets (per CLAUDE.local.md §2
  `//go:embed all:templates`).
- **REQ-HEV2-029** (Ubiquitous — empty-marker shipping): The template tree
  `internal/template/templates/CLAUDE.md` SHALL ship an EMPTY
  `MOAI:LEARNED-WORKFLOW` block marker (the `## MOAI:LEARNED-WORKFLOW`
  heading + `<!-- moai:learned-start -->` + `<!-- moai:learned-end -->` with
  zero bullets between the markers). The block CONTENT (populated bullets)
  never ships — it is per-project learned state, not template-distributable.
- **REQ-HEV2-030** (Unwanted behavior — template isolation): The empty marker
  block in the template tree shall carry NO internal SPEC IDs / REQ tokens /
  internal dates / commit SHAs / archive paths / memory paths (template-
  internal-content isolation, CLAUDE.local.md §25). The template-distributed
  marker is a generic empty section with the heading + markers only.

### C.9 Machine-signal-only + L5 approval gate

- **REQ-HEV2-031** (Unwanted behavior — machine-signal-only): The Curator's
  write authority shall derive from mechanical state — tier counts from the
  existing learner aggregation, gate results from the existing safety
  pipeline, manual `rollback` / `promote` / `demote` verbs — NEVER from model
  self-report. The writer API does not accept free-text model prose as input.
- **REQ-HEV2-032** (Compound — L5 approval): **While** the Curator proposes a
  CLAUDE.md or CLAUDE.local.md write, the proposal **shall** route through L5
  `AskUserQuestion` approval via the orchestrator's user-interaction channel
  (the writer is a subagent-side function and cannot call `AskUserQuestion`
  directly per the orchestrator-subagent boundary). **When** the orchestrator
  returns an approval token, the writer executes; **When** the orchestrator
  returns a rejection, the writer records the rejection in lineage and does
  NOT touch the file. No autonomous write path exists.
- **REQ-HEV2-033** (Unwanted behavior — rollback trigger): The rollback
  trigger shall be mechanical — regression-gate failure, L3 contradiction
  (once EVOLVE-003 activates L3), or a manual `moai harness rollback` verb —
  NEVER model self-report. The writer does not expose a "model decided to
  roll back" entry point.

### C.10 Go quality + subagent boundary

- **REQ-HEV2-034** (Ubiquitous — coverage and quality): The new
  `internal/harness/curator/` package and the extended
  `internal/harness/{applier,lineage,layer3}.go` and
  `internal/merge/strategies.go` shall reach ≥ 90% statement coverage, use
  table-driven tests with `t.TempDir()` isolation, wrap errors with `%w`,
  and set no OTEL environment variables via `t.Setenv` in parallel tests.
- **REQ-HEV2-035** (Unwanted behavior — subagent boundary): The
  `internal/harness/curator/` package and its call sites in
  `internal/harness/applier.go` shall NOT invoke `AskUserQuestion` (subagent
  boundary per C-HRA-008 / archived-agent-rejection.md §C). L5 approval is
  routed through the orchestrator; the Curator returns a proposal artifact
  and the orchestrator runs the `AskUserQuestion` round.
- **REQ-HEV2-036** (Unwanted behavior — no new hook surface): This SPEC shall
  add NO new hook wrapper script, NO `settings.json` / `settings.json.tmpl`
  hook registration change, and NO new gate. The typed writer is invoked via
  the existing Curator pipeline (the applier path) — the hook surface is
  unchanged.

## §D. Reference — typed writer contract (SSOT)

### D.1 `WriteManagedBlock` API sketch

```go
package curator

// BlockType enumerates the managed-block types the typed writer can manipulate.
type BlockType int

const (
    BlockTypeLearnedWorkflow  BlockType = iota  // MOAI:LEARNED-WORKFLOW (this SPEC)
    BlockTypeHarnessGenerated                   // ## Project-Specific Configuration (Harness-Generated) — backward compat
)

// BlockContent is the typed payload for a managed-block write.
type BlockContent struct {
    Bullets []Bullet           // ordered list; order is preserved on write
    Tier    int                // 3 (CLAUDE.local.md) or 4 (CLAUDE.md); writer rejects tier/surface mismatch
}

// Bullet is a single digest-layer entry.
type Bullet struct {
    LedgerKey  string  // cross-layer linkage to ledger-layer entry; "" or null marker when provisional
    Text       string  // distilled generic workflow knowledge (anti-fabrication binding — REQ-HEV2-011)
    Provisional bool   // true when ledger_key is null (early-tier observation)
}

// WriteManagedBlock reads path, locates the marker block matching blockType,
// replaces the block content atomically (or appends when absent), and writes
// the file back with pre-block and post-block bytes preserved verbatim.
// Returns ErrDigestBudgetExceeded / ErrBulletCapExceeded / ErrTierNotQualified
// without touching the file when the proposed write violates a guard.
func WriteManagedBlock(path string, blockType BlockType, content BlockContent) error

// AddBullet / UpdateBullet / DeleteBullet — per-bullet CRUD operations
// (default Curator interface; the writer locates the bullet by LedgerKey
// and rewrites only the targeted line).
func AddBullet(path string, blockType BlockType, bullet Bullet) error
func UpdateBullet(path string, blockType BlockType, ledgerKey string, newText string) error
func DeleteBullet(path string, blockType BlockType, ledgerKey string) error
```

### D.2 Marker contract

| Surface | File | Heading | Start marker | End marker |
|---------|------|---------|--------------|------------|
| Digest (Tier 4) | `CLAUDE.md` | `## MOAI:LEARNED-WORKFLOW` | `<!-- moai:learned-start -->` | `<!-- moai:learned-end -->` |
| Append-only (Tier 3) | `CLAUDE.local.md` | `## MOAI:LEARNED-WORKFLOW-LOCAL` | `<!-- moai:learned-local-start -->` | `<!-- moai:learned-local-end -->` |
| Legacy (backward compat) | `CLAUDE.md` | `## Project-Specific Configuration (Harness-Generated)` | `<!-- moai:harness-start ... -->` | `<!-- moai:harness-end -->` |

### D.3 Digest budget measurement

The ≤3,000-character digest budget is measured by reusing
`internal/config/token_budget_guard.go` `measureAlwaysLoaded(repoRoot string) (total int, surface []string, err error)`.
The function already enumerates `CLAUDE.md` as an always-loaded surface; this
SPEC extends the per-section attribution so the contribution of the
`MOAI:LEARNED-WORKFLOW` block is measurable distinctly, then enforces the
≤3,000-char budget at the writer API (REQ-HEV2-008).

### D.4 `LineageEntry` extension (sketch)

The existing `internal/harness/types.go LineageEntry` is extended with
Learned-surface fields (additive — existing fields preserved):

```go
type LineageEntry struct {
    // ...existing fields preserved verbatim...
    LearnedSurface  string   // "claude.md.learned-workflow" | "claude.local.md.learned-local" | "" (legacy entries)
    BulletsChanged  []string // ledger_keys of bullets added/updated/deleted
    SnapshotDir     string   // pointer to the snapshot restore unit ("" when no snapshot taken)
}
```

The full machine-verifiable AC matrix (AC-HEV2-001 … AC-HEV2-NNN) lives in
`acceptance.md` (SSOT). Every REQ maps to at least one AC; cross-file
registrations (template `CLAUDE.md` empty marker, `mergeSectionBased`
extension, `LineageEntry` extension, snapshot manifest entry) are pinned as
SEPARATE baseline-0 ACs per the reachability discipline (inherited from
EVOLVE-001 + the `feedback_ac_token_presence_not_reachability` lesson).

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — Tier↔surface mapping activation (EVOLVE-003)

- NO harness.yaml v2 schema, NO `auto_detection` block registration, NO
  tier-count rule changes, NO permission-surface Frozen-guard expansion
  (design deltas A1, A6, A7 land in `SPEC-HARNESS-EVOLVE-003`). This SPEC
  builds the writer API and accepts an explicit `tier` argument (REQ-HEV2-027
  rejects mismatches), but the *activation* of tier↔surface mapping in the
  harness config is EVOLVE-003.

### Out of Scope — Gate activation (EVOLVE-003)

- NO L2 Canary activation (shadow-apply + doctor + regression gate), NO L3
  Contradiction activation (Frozen-rules contradiction check). The 5-layer
  pipeline code stays as-is; this SPEC's writer is invoked via the existing
  applier path. Gate *activation* is EVOLVE-003.

### Out of Scope — Re-proposal suppression (EVOLVE-003 / A7)

- NO negative-evidence registry construction, NO re-proposal cooldown logic.
  This SPEC *names the search-interface slot* for the negative-evidence
  registry in the Recall contract ledger layer (REQ-HEV2-016) and ensures
  `LineageEntry` can reference registry keys, but the registry ITSELF
  (the data structure, the writer, the cooldown logic) is EVOLVE-003.

### Out of Scope — Console verbs (EVOLVE-004)

- NO `/moai harness evolve | promote | demote | freeze | unfreeze` verbs,
  NO `status` / `doctor` extension for Learned-surface views
  (`SPEC-HARNESS-EVOLVE-004`). This SPEC authors the write-path machinery
  only; the human-facing console surface that drives it is EVOLVE-004.

### Out of Scope — Recall wiring + typed parser + template deployment (EVOLVE-005)

- NO Phase −1 Harness Recall wiring (the *consumption* of the digest layer
  by the router), NO Phase Ω routing-bias consumption, NO
  `harness-spec.yaml` typed Go parser, NO template-deployment verification
  of the empty marker beyond the §C.8 marker-contract checks
  (`SPEC-HARNESS-EVOLVE-005`). This SPEC *formalizes* the 2-layer Recall
  contract (§C.4) but does NOT wire its consumption.

### Out of Scope — Loop 0 / Loop 1 changes (EVOLVE-001 boundary)

- NO changes to the EVOLVE-001 observation layer. The routing-ledger writer
  (`internal/harness/routing/`) is untouched; the ledger-layer search
  interface (REQ-HEV2-016) consumes the ledger read-only via the existing
  `internal/harness/routing/reader.go`.

### Out of Scope — New hook surface

- NO new hook wrapper script, NO `settings.json` / `settings.json.tmpl`
  hook registration change, NO new gate (REQ-HEV2-036). The writer is
  invoked via the existing applier path inside the Curator pipeline.

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site
  4-locale updates are a follow-up sync/docs concern.

## §F. Cross-References

- `.moai/reports/harness-self-evolving-redesign-final-20260712.html` — design
  SSOT (§3 gaps G2/G3, §4 3-Zone edit-surface contract Learned-zone row, §5
  Loop 2 + A3 two-layer Recall contract, §7 M2 milestone + risk grid +
  success metrics).
- `.moai/specs/SPEC-HARNESS-EVOLVE-001/spec.md` — upstream Epic predecessor
  (Loop 0 observation layer; `depends_on` target). This SPEC consumes
  EVOLVE-001's routing ledger read-only via the ledger-layer search interface.
- `internal/harness/layer3.go InjectMarker` — the sole existing CLAUDE.md
  mechanical write path, GENERALIZED into the typed writer (REQ-HEV2-001).
- `internal/merge/strategies.go mergeSectionBased` — the existing
  `moai update` section-preservation logic, EXTENDED to recognize the new
  managed block (REQ-HEV2-019/020).
- `internal/config/token_budget_guard.go measureAlwaysLoaded` — the existing
  always-loaded surface budget measurement, REUSED for the ≤3K digest-budget
  enforcement (REQ-HEV2-008).
- `internal/harness/applier.go createSnapshot` + `RestoreSnapshot` — the
  existing snapshot/rollback machinery, EXTENDED to cover the new surfaces
  (REQ-HEV2-021/022).
- `internal/harness/lineage.go WriteLineageEntry` — the existing lineage
  writer, EXTENDED with Learned-surface fields (REQ-HEV2-023/024).
- `internal/harness/types.go LineageEntry` — the existing lineage record
  type, EXTENDED additively (§D.4).
- `internal/template/templates/CLAUDE.md` — the template source for the empty
  `MOAI:LEARNED-WORKFLOW` block marker (REQ-HEV2-029).
- CLAUDE.local.md §2 (Template-First) + §25 (Template Internal-Content
  Isolation) — mirror + neutrality discipline (REQ-HEV2-028/029/030).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — the
  no-unobserved-claim invariant the machine-signal-only rule
  (REQ-HEV2-031) + evidence-or-null rule (REQ-HEV2-024) mechanize at the
  write layer.
- `.claude/rules/moai/workflow/archived-agent-rejection.md` §C — subagent
  boundary (REQ-HEV2-035); the Curator returns proposals, the orchestrator
  runs `AskUserQuestion` (REQ-HEV2-032).
- `SPEC-HARNESS-EVOLVE-003..005` (unauthored) — Epic successors consuming
  these write surfaces.
- `plan.md` / `acceptance.md` / `design.md` / `research.md` — implementation
  plan + AC matrix + architecture decisions + codebase investigation (SSOT).
