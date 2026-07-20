---
id: SPEC-HARNESS-EVOLVE-002
title: "Curator Editable Surfaces — Loop 2 (write layer) of the self-evolving harness"
version: "0.1.0"
status: in-progress
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

# SPEC-HARNESS-EVOLVE-002 — Codebase Investigation

> Counterpart to `spec.md` (requirements SSOT), `plan.md` (implementation
> approach), `design.md` (architecture decisions). This document owns the
> codebase investigation: existing-asset inventory, signature maps, boundary
> reference, and the factual basis for every claim about "what already
> exists". Per `feedback_ac_token_presence_not_reachability`, every "asset
> exists" claim in this document was verified at plan-phase authoring time
> via direct grep / Read / find — citation-grade evidence, not memory recall.

## §A. Existing-Asset Inventory (the ~70% that already exists)

The design SSOT §3 states "harness assets의 약 70%는 이미 존재한다". This
section is the precise inventory of that 70% — verified 2026-07-12 at
plan-phase authoring time.

### A.1 The sole CLAUDE.md mechanical write path — `InjectMarker`

**Path**: `internal/harness/layer3.go`
**Signature** (verified verbatim):

```go
// InjectMarker injects (or replaces) the harness block in the file at
// claudeMdPath. If a block already exists (regardless of its specID), it is
// replaced atomically with the new block built from specID, domain, and
// importPaths. The function is idempotent: re-running with same or different
// specIDs produces exactly one block per file.
func InjectMarker(claudeMdPath, specID, domain string, importPaths []string) error
```

**Mechanism** (verified via direct Read):
- Reads the file via `os.ReadFile`.
- Matches the existing block via `markerBlockPattern` (a `regexp.MustCompile`
  that matches the heading + start marker + body + end marker as one atomic
  group).
- If matched: `ReplaceAllString` swaps the block content atomically.
- If not matched: appends the block with a separating newline.
- Writes back via `os.WriteFile` (mode 0o644).

**`markerBlockPattern`** (verified verbatim):

```go
var markerBlockPattern = regexp.MustCompile(
    `(?s)## Project-Specific Configuration \(Harness-Generated\)\n<!-- moai:harness-start[^>]*-->.*?<!-- moai:harness-end -->`,
)
```

**`buildMarkerBlock`** (verified verbatim) — builds the legacy block with
specID, domain, harness level, updated date, and `See @<path>` import refs.

**Call sites** (re-verified 2026-07-12 at plan-audit iter-1 via
`grep -rn "InjectMarker" internal/ --include="*.go" | grep -v "_test.go"` —
verbatim 10-line output):

```
internal/cli/harness_route.go:120:	// live call path that wires the previously-orphaned InjectMarker (layer3)
internal/cli/harness/install.go:5:// internal/harness.InjectMarker (layer3, CLAUDE.md marker block) and
internal/cli/harness/install.go:79:	// existing InjectMarker installer (layer3). Idempotent — re-running
internal/cli/harness/install.go:85:	if err := harness.InjectMarker(claudeMdPath, opts.SpecID, opts.Domain,
internal/harness/layer3.go:21:// InjectMarker injects (or replaces) the harness block in the file at
internal/harness/layer3.go:26:func InjectMarker(claudeMdPath, specID, domain string, importPaths []string) error {
internal/harness/layer3.go:28:		return errors.New("InjectMarker: empty path")
internal/harness/layer3.go:31:		return errors.New("InjectMarker: empty specID")
internal/harness/layer3.go:35:		return fmt.Errorf("InjectMarker: read %s: %w", claudeMdPath, err)
internal/harness/layer3.go:50:		return fmt.Errorf("InjectMarker: write %s: %w", claudeMdPath, err)
```

Classification of the 10 non-test matches:

- **ONE production caller** — `internal/cli/harness/install.go:85`. This is
  the live `moai harness install` call path (the cobra registration at
  `internal/cli/harness_route.go:124` wires `NewInstallCmd()`; the
  `harness_route.go:120` comment documents this as the SPEC-V3R6-HARNESS-
  ACTIVATION-WIRING-001 wiring of the previously-orphaned InjectMarker).
- **Definition** — `internal/harness/layer3.go:26` (the function itself) +
  internal error messages at `layer3.go:21,28,31,35,50`.
- **Documentation comments** — `install.go:5,79` + `harness_route.go:120`
  (they describe the install path's relationship to InjectMarker).

The plan-audit iter-1 finding D1 corrected an earlier false premise in this
section ("only test call sites"). There is exactly ONE production caller
(`install.go:85`). This is load-bearing for M1: the generalization MUST
preserve the byte-identical behavior of the `## Project-Specific
Configuration (Harness-Generated)` block that `moai harness install`
injects into CLAUDE.md. The legacy `InjectMarker` is preserved as a thin
backward-compat wrapper over `curator.WriteManagedBlock(path,
BlockTypeHarnessGenerated, ...)` (REQ-HEV2-002) AND its one live caller
(`install.go:85`) must keep producing the same block byte-for-byte. See
plan.md §F M1 backward-compat verification list + AC-HEV2-003.

### A.2 The `mergeSectionBased` section-preservation logic

**Path**: `internal/merge/strategies.go:469-471`
**Signature** (verified verbatim):

```go
// mergeSectionBased performs section-based merge for CLAUDE.md files.
func mergeSectionBased(base, current, updated []byte) (*MergeResult, error)
```

**Call sites** (verified):

```
internal/merge/three_way.go:61     — production caller (the 3-way merge driver)
internal/merge/strategies_test.go:342, 369 — tests
```

**Mechanism** (per the comment + the `mergeSectionBased` name): parses
markdown by `## ...` heading and merges sections independently. This SPEC
extends the recognition to treat the new managed-block heading as a
preserved section (REQ-HEV2-019, design.md §D).

### A.3 The always-loaded budget measurement

**Path**: `internal/config/token_budget_guard.go`
**Key signatures** (verified verbatim):

```go
func estimateTokens(b []byte) int                                          // line 39
func findRepoRoot(start string) (string, bool)                             // line 45
func frontmatterHasPaths(data []byte) bool                                 // line 62
func hasPathsRestriction(path string) bool                                 // line 88
func alwaysLoadedSurface(repoRoot string) ([]string, error)                // line 101
func memoryHead(data []byte) []byte                                        // line 132
func measureAlwaysLoaded(repoRoot string) (total int, surface []string, err error)  // line 151
```

**Usage in this SPEC**: `measureAlwaysLoaded` is REUSED for the
≤3,000-character digest-budget enforcement (REQ-HEV2-008). The function
enumerates always-loaded surfaces including CLAUDE.md; this SPEC may add a
per-section attribution extension so the `MOAI:LEARNED-WORKFLOW` block's
contribution is measurable distinctly (additive extension — no breaking
change to the return type).

**Test coverage baseline**: `internal/config/token_budget_guard_test.go`
exists — the existing tests must continue to pass.

### A.4 The snapshot / rollback machinery

**Path**: `internal/harness/applier.go`
**Key signatures** (verified verbatim):

```go
func (a *Applier) Apply(proposal Proposal, evaluator SafetyEvaluator, snapshotBase string, sessions []Session) error  // line 282
func (a *Applier) applyWithRegressionGate(proposal Proposal, snapshotDir string) error                                // line 432
func (a *Applier) measurementProjectRoot(snapshotDir string) string                                                   // line 521
func measurementRoot(snapshotDir string) string                                                                       // line 540
func (a *Applier) createSnapshot(proposal Proposal, snapshotBase string) (string, error)                              // line 568
func RestoreSnapshot(snapshotDir string) error                                                                        // line 622
```

**`createSnapshot` mechanism** (verified verbatim):
- Generates an ISO-date directory name under `snapshotBase`.
- Reads the original file via `os.ReadFile(proposal.TargetPath)`.
- Writes the backup with the original filename as `backupName`.
- Writes a `manifest.json` recording `ProposalID`, `CreatedAt`,
  `Files: [{OriginalPath, BackupName}]`.

**`RestoreSnapshot` mechanism** (verified): reads `manifest.json`, restores
original files from the backup.

**Usage in this SPEC**: extended to snapshot the CLAUDE.md managed block +
CLAUDE.local.md Learned section as distinct restore units (REQ-HEV2-021).
The `@MX:ANCHOR` on `RestoreSnapshot` (line 620) marks it as an invariant
contract — the extension must preserve the existing signature for backward
compatibility.

**`@MX:ANCHOR` annotations** (verified):

```
line 620: // @MX:ANCHOR: [AUTO] RestoreSnapshot is the core function of rollback functionality.
line 621: // @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, harness CLI rollback, Phase 5 IT
```

### A.5 The lineage writer/loader

**Path**: `internal/harness/lineage.go`
**Key signatures** (verified verbatim):

```go
// WriteLineageEntry는 manifestPath에 단일 LineageEntry를 append한다 (부모 디렉토리 자동 생성).
// @MX:ANCHOR: [AUTO] WriteLineageEntry는 lineage 기록의 단일 진입점.
// @MX:REASON: [AUTO] fan_in >= 3: applier.go(Apply accept+reject), lineage_test.go, harness CLI(future)
func WriteLineageEntry(manifestPath string, entry LineageEntry) error

// LoadManifest는 manifestPath의 모든 LineageEntry를 write 순서대로 읽는다.
// @MX:ANCHOR: [AUTO] LoadManifest는 lineage 조회의 단일 진입점.
// @MX:REASON: [AUTO] fan_in >= 3: lineage_test.go(여러 AC), harness CLI status(future), Phase 5 IT
func LoadManifest(manifestPath string) ([]LineageEntry, error)
```

**Mechanism** (verified verbatim):
- `WriteLineageEntry`: auto-creates parent dir, defaults `Timestamp` to
  `time.Now().UTC()` when zero, marshals entry to JSON, appends `\n`,
  opens with `O_APPEND|O_CREATE|O_WRONLY` 0o644, writes. (Mirrors
  `learner.go:145 WritePromotion` idiom.)
- `LoadManifest`: returns `([]LineageEntry{}, nil)` when file absent
  (backward compat), skips blank/malformed lines.

**Usage in this SPEC**: the `LineageEntry` struct (defined in
`internal/harness/types.go:521`) is EXTENDED additively with
`LearnedSurface`, `BulletsChanged`, `SnapshotDir` fields (spec.md §D.4).
The writer/loader signatures stay byte-identical.

### A.6 The 4-tier learning ladder

**Path**: `internal/harness/learner.go` (and the `scorer_engine.go` /
`outcome.go` siblings).

The 4-tier ladder has thresholds `[1, 3, 5, 10]` per the design SSOT:
- Tier 1 (1 observation) → auto-memory temporary
- Tier 2 (3 observations) → auto-memory temporary (reinforced)
- Tier 3 (5 observations) → CLAUDE.local.md permanent append (this SPEC's
  append-only writer)
- Tier 4 (10 observations) → CLAUDE.md managed block (this SPEC's
  digest writer)

This SPEC does NOT touch the threshold values or the aggregation logic
(`learner.go`). It binds only the surface selection (REQ-HEV2-025/026) —
when the existing learner reports "pattern P reached Tier N", the Curator
selects the corresponding surface.

### A.7 The 5-layer safety pipeline

**Paths**: `internal/harness/layer1.go` (Frozen-guard), `layer2.go`
(Canary), `layer3.go` (the marker injector — generalized in this SPEC),
`layer5.go` (AskUserQuestion approval). L3 (Contradiction) lives in
`internal/harness/safety/canary_veto.go` per the grep result.

This SPEC does NOT activate L2 Canary or L3 Contradiction (those are
EVOLVE-003). The typed writer is invoked via the existing applier path
inside the Curator pipeline.

### A.8 The routing-ledger observation layer (EVOLVE-001 output)

**Path**: `internal/harness/routing/`
**Files** (verified via `find internal/harness/routing/ -type f`):

```
internal/harness/routing/types.go
internal/harness/routing/types_test.go
internal/harness/routing/writer.go
internal/harness/routing/writer_reader_test.go
internal/harness/routing/reader.go
internal/harness/routing/edge_test.go
internal/harness/routing/pending.go
internal/harness/routing/pending_test.go
internal/harness/routing/outcome.go
internal/harness/routing/outcome_test.go
internal/harness/routing/digest.go
internal/harness/routing/subagent_boundary_test.go
```

**Usage in this SPEC**: read-only consumption via the ledger-layer search
interface (REQ-HEV2-016). The routing writer is untouched (EVOLVE-001
boundary).

## §B. Boundary Reference — `SPEC-HARNESS-EVOLVE-001`

EVOLVE-001 `status: completed` (sync `f12ebc7ea` + backfill `b242450ed`,
verified at delegation time). The boundary contract this SPEC inherits:

### B.1 What EVOLVE-001 produced (this SPEC consumes read-only)

- `.moai/state/routing-ledger.jsonl` — append-only JSONL, schema v1,
  gitignored runtime state. Fields per EVOLVE-001 §D.1 (the 13-field
  schema including design deltas A2 `loop_iterations` / `goal_converged`
  / `convergence_class` and A4 `delegations[]`).
- `.moai/state/routing-pending-<session>.json` — per-session pending-row
  store.
- `internal/harness/routing/` — Go writer/reader package with filters
  (subcommand / outcome / time window).
- Stop-hook outcome capture via the HOI-gated `harness-observe-stop`
  handler (additive, fail-open).
- CLI verbs `moai harness ledger record | evidence | list`.
- Workflow skill recording obligations in
  `.claude/skills/moai/SKILL.md` + `workflows/{plan,run,sync}.md`.

### B.2 What EVOLVE-001 explicitly excluded (this SPEC's boundary)

From EVOLVE-001 §E "Out of Scope — Curator writes and Learned surfaces
(EVOLVE-002)":

> - NO writes to CLAUDE.md, CLAUDE.local.md, or any `MOAI:LEARNED-WORKFLOW`
>   managed block. The typed managed-block writer, the append-only Learned
>   section writer, the 2-layer Recall contract, and snapshot/rollback/
>   lineage extension are `SPEC-HARNESS-EVOLVE-002` territory.
> - NO digest-layer emission: this SPEC produces the ledger (원장 layer)
>   only; nothing is loaded into always-on context.

This SPEC builds exactly those excluded pieces. The boundary is clean:
EVOLVE-001 owns the ledger writer; EVOLVE-002 owns the digest writer +
the 2-layer contract + the snapshot/rollback/lineage extension for the
new surfaces.

### B.3 Inherited anti-fabrication principles

From EVOLVE-001 §A (this SPEC inherits verbatim):

1. **Machine signals only** — `outcome` derives from exit codes, audit
   scores, gate results; never from model self-report.
2. **Privacy / template neutrality** — verbatim user text never persists.
3. **Evidence-or-null** — absent evidence is `null`, never inferred.

This SPEC applies the same three principles to the digest-layer write
path (REQ-HEV2-011, REQ-HEV2-024, REQ-HEV2-031).

### B.4 EVOLVE-001's `era: V3R6` + frontmatter style

This SPEC mirrors EVOLVE-001's frontmatter structure:
- `era: V3R6` set explicitly (avoids transient V3R2-R4 misclassification
  per `lifecycle-sync-gate.md` H-2 — the progress.md will be minimal at
  plan start).
- `tier: L` (EVOLVE-001 was Tier M; this SPEC is Tier L per the design
  SSOT §7 M2).
- `depends_on: [SPEC-HARNESS-EVOLVE-001]` — explicit dependency
  declaration per the Phase 0.5 Depends_on Pre-flight Check.
- 12 canonical frontmatter fields per
  `.claude/rules/moai/development/spec-frontmatter-schema.md`.

## §C. Boundary Reference — Template Internal-Content Isolation (§25)

Per CLAUDE.local.md §25 (consulted at
`.moai/docs/template-internal-isolation-doctrine.md`), the template tree
(`internal/template/templates/`) is the distribution surface — it ships to
external users and MUST NOT carry moai-adk internal development traces.

### C.1 Forbidden content classes in the template `CLAUDE.md` block

The empty `MOAI:LEARNED-WORKFLOW` marker shipped to
`internal/template/templates/CLAUDE.md` MUST NOT contain:

- Internal SPEC IDs (`SPEC-HARNESS-EVOLVE-*`, `SPEC-V3R6-*`, etc.)
- Internal REQ/AC tokens (`REQ-HEV2-*`, `AC-HEV2-*`)
- Internal session dates (`2026-07-12` ISO form)
- Internal commit SHAs (`[0-9a-f]{7,8}`)
- Audit citations ("Audit N Finding AX")
- Archive paths (`.moai/backups/`)
- Memory paths (`~/.claude/projects/-Users-goos-...`)

### C.2 Allowed content in the template block marker

- The heading `## MOAI:LEARNED-WORKFLOW` (MoAI-ADK system identifier — allowed)
- The HTML-comment start/end markers (generic mechanism)
- An (optional) one-line generic-prose description of the block's purpose
  (e.g. "<!-- Distilled workflow knowledge, populated by the harness
  Curator. Editable via bullet CRUD; do not edit inline. -->") —
  generic mechanism description, allowed

### C.3 The populated block content NEVER ships

The populated digest (with bullets) is per-project learned state. The
template ships only the EMPTY marker (heading + markers + zero bullets).
REQ-HEV2-029 binds this; AP-HEV2-006 names the violation.

### C.4 CI guard

`internal/template/internal_content_leak_test.go` +
`.github/workflows/template-neutrality-check.yaml` enforce the
forbidden-classes regex tree-wide. The new template `CLAUDE.md` block
marker MUST pass these guards (AC-HEV2-038). The run-phase M6 milestone
verifies this.

## §D. Boundary Reference — `moai spec lint` baseline

`moai spec lint` is available at `/Users/goos/go/bin/moai` (verified).
The lint validates:

- EARS/GEARS modality compliance (SHALL, WHEN, WHILE, WHERE, IF)
- REQ ID uniqueness
- AC→REQ coverage (100% required)
- Frontmatter schema validation (the 12 canonical fields)
- Dependency DAG (no cycles, all deps exist — `depends_on:
  [SPEC-HARNESS-EVOLVE-001]` will be resolved against the EVOLVE-001 dir)
- Out of Scope section presence (`### Out of Scope — <topic>` H3
  sub-headings with `-` bullets, per the `OutOfScopeRule`)
- Zone registry cross-references

The spec.md authored here is structured to pass `moai spec lint` AND
`moai spec lint --strict` (no `LegacyEARSKeyword` — the SPEC uses GEARS
notation throughout, no residual `IF/THEN`).

## §E. Working-Tree State at Plan-Phase Authoring (2026-07-12)

### E.1 Active parallel work streams (DO NOT TOUCH)

These uncommitted / local-ahead items are a SEPARATE work stream
(v3.0-tokenomics):

```
M  .claude/skills/moai-domain-html-report/SKILL.md
M  .moai/config/sections/llm.yaml
M  internal/template/templates/.claude/skills/moai-domain-html-report/SKILL.md
?? .moai/harness/proposals/
?? .moai/reports/harness-self-evolving-redesign-20260712.html
?? .moai/reports/harness-self-evolving-redesign-final-20260712.html
?? .moai/reports/model-tier-redesign-20260712.html
?? .moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/
```

Plus 1+ local-ahead commit(s) on `main` (the most recent at authoring
time: `97dd0dbf9 fix(statusline): use recycle icon for cache_hit segment`).

### E.2 Commit hygiene (this SPEC)

- Plan-phase commit: `git add .moai/specs/SPEC-HARNESS-EVOLVE-002/`
  (specific path ONLY — never `git add -A` / `git add .`).
- Conventional Commit subject: `feat(SPEC-HARNESS-EVOLVE-002): plan-phase
  artifacts (L, 5 artifacts)` per the Status Transition Ownership Matrix
  `(none) → draft` row.
- `🗿 MoAI` trailer on every commit.
- Commit PROMPTLY after authoring (4 active sessions on this checkout;
  `git stash -u` / reset / clean absorption risk per
  `feedback_spec_artifact_commit_before_parallel_work`).

### E.3 Active sessions (Pre-Spawn Sync Check)

Per the agent-common-protocol.md § Pre-Spawn Sync Check, the orchestrator
runs `moai session list --json --filter-spec=SPEC-HARNESS-EVOLVE-002`
before run-phase spawn. At plan-phase authoring, no other session is
working on this SPEC ID (the parallel sessions are on the v3.0-tokenomics
stream).

## §F. Gaps Surfaced by Investigation

### F.1 `InjectMarker` has ONE production caller (plan-audit iter-1 D1 correction)

> **Correction (iter-1)**: an earlier draft of this section asserted
> `InjectMarker` had NO production caller. That was a false premise —
> re-verification at plan-audit iter-1 found ONE live production caller.

Verified via `grep -rn "InjectMarker" internal/ --include="*.go" | grep -v "_test.go"`
(verbatim 10-line output — see §A.1 for the full block):

- **Production caller** — `internal/cli/harness/install.go:85`, the live
  `moai harness install` path (cobra-registered at `harness_route.go:124`).
  This call injects the `## Project-Specific Configuration (Harness-Generated)`
  block into CLAUDE.md.
- The remaining 9 non-test matches are the function definition
  (`layer3.go:21,26,28,31,35,50`) + documentation comments
  (`install.go:5,79` + `harness_route.go:120`).

**Risk conclusion (corrected)**: the generalization in M1 is NOT
low-risk-by-absence — there IS a production caller whose byte-identical
behavior MUST be preserved. M1 MUST preserve the byte-identical
`## Project-Specific Configuration (Harness-Generated)` block produced by
the `install.go:85` caller. The legacy `InjectMarker` entry point is
preserved as a thin backward-compat wrapper over
`curator.WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)`
(REQ-HEV2-002), AND its one live caller must keep producing the same
block byte-for-byte (a `moai harness install` smoke test on a fixture
project verifies this — see plan.md §F M1 backward-compat verification
list + AC-HEV2-003).

### F.2 `measureAlwaysLoaded` may need a per-section attribution extension

The existing function returns `(total int, surface []string, err error)`
— a total budget number + the list of always-loaded surfaces. It does
NOT attribute the budget per-section within a single surface. The
≤3K-budget enforcement on the `MOAI:LEARNED-WORKFLOW` block (REQ-HEV2-008)
requires either:
- (a) measuring the block's contribution distinctly (per-section
  attribution extension — additive), OR
- (b) measuring the whole CLAUDE.md before vs after the proposed write
  and computing the delta (simpler, no function extension).

Run-phase M2 will choose between (a) and (b). The simpler (b) is the
default — write a temp file with the proposed block, run
`measureAlwaysLoaded` on it, compute the delta. This avoids modifying
the existing function.

### F.3 `LineageEntry` struct location + `@MX:ANCHOR` discipline

The struct is at `internal/harness/types.go:521` with `@MX:ANCHOR` marks
(fan_in >= 3). The additive extension in M5 (spec.md §D.4) MUST preserve
the existing fields verbatim — the `@MX:ANCHOR` discipline requires that
the existing field set stays stable so the existing callers
(`WriteLineageEntry`, `LoadManifest`, the applier) continue to work.

### F.4 `mergeSectionBased` recognition mechanism — open question

The merge parses by `## ...` heading. Whether the new managed-block
heading is auto-recognized or requires explicit registration is
NEEDS CLARIFICATION H-2 in plan.md. The investigation confirms BOTH
options are technically feasible; the choice is policy (explicit allow-
list vs marker-pair auto-recognition).

## §G. Cross-References

- `.moai/specs/SPEC-HARNESS-EVOLVE-001/spec.md` — boundary reference (§B
  above).
- `.moai/reports/harness-self-evolving-redesign-final-20260712.html` —
  design SSOT (§3 ~70% inventory, §4 3-Zone contract, §5 3-Loop, §7 M2).
- `.moai/docs/template-internal-isolation-doctrine.md` — §25 template
  neutrality (§C above).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter
  schema + Status Transition Ownership Matrix.
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — `era: V3R6`
  explicit-override rationale.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` —
  Section A-E delegation template (Tier L → REQUIRED at run-phase).
- `internal/harness/layer3.go` — `InjectMarker` source (§A.1).
- `internal/merge/strategies.go` — `mergeSectionBased` source (§A.2).
- `internal/config/token_budget_guard.go` — budget measurement source (§A.3).
- `internal/harness/applier.go` — snapshot/rollback source (§A.4).
- `internal/harness/lineage.go` — lineage writer/loader source (§A.5).
- `internal/harness/types.go` — `LineageEntry` struct (§A.5, §F.3).
- `internal/harness/routing/` — EVOLVE-001 output (§A.8).
