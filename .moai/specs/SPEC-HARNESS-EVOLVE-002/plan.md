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

# SPEC-HARNESS-EVOLVE-002 — Implementation Plan

> Counterpart to `spec.md` (SSOT for requirements) and `acceptance.md` (SSOT
> for acceptance criteria). This document owns the implementation approach,
> milestones, technical approach, risks, and NEEDS-CLARIFICATION items.

## §A. Context

### A.1 Delegation targets

- **Plan-phase authoring**: `manager-spec` (this document).
- **Plan-phase audit**: `plan-auditor` (independent, bias-prevention).
- **Run-phase implementation**: `manager-develop` (cycle_type per
  `quality.yaml` — default `tdd`; this SPEC extends existing Go code in
  `internal/harness/`, so the brownfield pre-RED analysis applies per the
  TDD workflow's brownfield enhancement rule).
- **Sync-phase documentation**: `manager-docs`.

### A.2 Working location + branch policy

Per the Tier classification (Tier L), this SPEC follows **Route B — PR route**
by default (`spec-workflow.md` § SPEC Phase Discipline). However, the local
1-person-OSS Hybrid Trunk policy (CLAUDE.local.md §23) permits direct-to-main
for all tiers when the developer opts in. The actual route is decided at
Implementation Kickoff Approval.

- **Working tree**: main checkout (Step 1 plan-phase per Step-1 main-checkout
  invariant).
- **Branch (Route A)**: `main` direct.
- **Branch (Route B)**: `feat/SPEC-HARNESS-EVOLVE-002` at run-phase Step 2.

### A.3 Current state at plan-phase authoring

- EVOLVE-001 `status: completed` (sync `f12ebc7ea` + backfill `b242450ed`,
  verified at delegation time). The routing-ledger writer
  (`internal/harness/routing/`) is live and producing rows when HOI is
  opted-in (dev-repo dogfood-enable).
- Local-ahead commit `97dd0dbf9` (statusline) on `main` — unrelated work
  stream, not blocked by this SPEC.
- Parallel uncommitted work stream in flight (v3.0-tokenomics: statusline,
  html-report skill, llm.yaml, model-tier-redesign report,
  SPEC-MODEL-TIER-PLANTYPE-001 plan artifacts). This SPEC MUST NOT touch
  any of those paths — commit hygiene uses specific-path `git add`
  (`.moai/specs/SPEC-HARNESS-EVOLVE-002/` only).

### A.4 SPEC artifact paths

- `spec.md` — ~330 lines (this plan's SSOT counterpart)
- `plan.md` — this file
- `acceptance.md` — AC matrix SSOT
- `design.md` — architecture decisions + alternatives
- `research.md` — codebase investigation + asset inventory
- `progress.md` — §E lifecycle skeleton (manager-spec authors §E.1 only)

### A.5 PRESERVE list (files this SPEC must NOT modify)

- `internal/harness/routing/**` — EVOLVE-001 output (read-only consumption).
- `internal/harness/observer.go` + `.moai/harness/usage-log.jsonl` — separate
  observation surface.
- `internal/harness/layer1.go` / `layer2.go` / `layer5.go` — safety pipeline
  code unchanged (L2/L3 activation is EVOLVE-003).
- `internal/harness/learner.go` — tier-count aggregation unchanged (tier-
  count rule changes are EVOLVE-003).
- All v3.0-tokenomics work-in-progress paths (statusline/*, html-report skill,
  llm.yaml, model-tier reports, SPEC-MODEL-TIER-PLANTYPE-001/).
- `internal/harness/layer3.go InjectMarker` — generalized, NOT deleted; the
  legacy entry point is preserved as a thin wrapper over the typed writer
  for backward compatibility.

## §B. Known Issues (auto-injection per manager-develop-prompt-template §B)

- **B1 Cross-platform build tags**: the new `internal/harness/curator/` package
  uses only stdlib `os` / `regexp` / `errors` / `fmt` — no syscall, no
  cross-platform build tags needed. Verify with `GOOS=windows GOARCH=amd64
  go build ./...` at run-phase.
- **B2 Cross-SPEC policy conflict pre-scan**: this SPEC EXTENDS existing
  machinery (InjectMarker, mergeSectionBased, createSnapshot, WriteLineageEntry).
  Run `grep -r "InjectMarker\|mergeSectionBased\|createSnapshot\|WriteLineageEntry"
  internal/` at M1 to enumerate ALL call sites before generalizing — the
  generalization must be backward-compatible (existing callers keep working
  byte-identical via the `BlockTypeHarnessGenerated` legacy enum value).
- **B3 C-HRA-008 subagent boundary**: the new `internal/harness/curator/`
  package MUST NOT call `AskUserQuestion` (REQ-HEV2-035). L5 approval is
  routed via the orchestrator. CI guard: `internal/harness/curator/subagent_boundary_test.go`
  grep 0 matches.
- **B4 Frontmatter canonical schema**: N/A (plan-phase artifact, no
  frontmatter changes during run-phase).
- **B5 CI 3-tier awareness**: spec-lint, golangci-lint, and Test run
  independently. Distinguish the W1/W2 chicken-and-egg pattern (template
  edit triggers template-neutrality-check.yaml + internal_content_leak_test.go)
  vs NEW defects.
- **B6 spec-lint heading convention**: the spec.md §E uses `### Out of Scope — <topic>`
  H3 sub-headings (verified). Run `moai spec lint` after plan-phase commit.
- **B7 observer.go / capture path resolution**: N/A — this SPEC does not
  touch observer.go or the capture path.
- **B8 Working tree hygiene**: `.moai/state/routing-ledger.jsonl`,
  `.moai/state/routing-pending-*.json`, `.moai/state/<learning-history>/`,
  `.moai/harness/usage-log.jsonl` are runtime-managed — DO NOT touch in tests
  (use `t.TempDir()` isolation per REQ-HEV2-034).
- **B9 Git commit + push performed directly**: per Hybrid Trunk 1-person OSS
  (CLAUDE.local.md §23), manager-develop commits + pushes directly within
  this SPEC scope. Conventional Commits (`feat(SPEC-HARNESS-EVOLVE-002):
  M{N} <subject>`), `🗿 MoAI` trailer. Never `--no-verify` (a warn-only
  pre-commit hook is normal).
- **B10 Untouched paths PRESERVE**: see §A.5 above. Take extra care during
  parallel-session work (4 active sessions on this checkout per the
  delegation prompt).
- **B11 AskUserQuestion prohibited**: returns blocker report on missing
  input; never free-form prose questions.
- **B12 CHANGELOG emission**: N/A (manager-docs owns sync-phase).

## §C. Pre-flight Check List (run before any code change)

```bash
# 1. Current branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build feasibility
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Existing lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -10

# 4. Enumerate ALL call sites of the to-be-generalized functions
grep -rn "InjectMarker\b" internal/ --include="*.go" | grep -v "_test.go"
grep -rn "mergeSectionBased\b" internal/ --include="*.go" | grep -v "_test.go"
grep -rn "createSnapshot\b\|RestoreSnapshot\b" internal/ --include="*.go" | grep -v "_test.go"
grep -rn "WriteLineageEntry\b\|LoadManifest\b" internal/ --include="*.go" | grep -v "_test.go"

# 5. Measure current CLAUDE.md always-loaded budget (baseline for ≤3K enforcement)
go test -run TestMeasureAlwaysLoaded ./internal/config/ -v 2>&1 | tail -20

# 6. Verify EVOLVE-001 ledger is producing rows (dogfood-enabled dev repo)
test -f .moai/state/routing-ledger.jsonl && wc -l .moai/state/routing-ledger.jsonl

# 7. Check the existing 4-tier learning ladder thresholds
grep -n "1.*3.*5.*10\|thresholds.*\[1" internal/harness/learner.go | head -5

# 8. Check retired/superseded SPECs of affected packages (B2)
grep -r "Retired\|TestHarnessRetirement\|superseded" internal/harness/ | head -5
```

## §D. Constraints (DO NOT VIOLATE)

- **PRESERVE**: every file in §A.5. The generalization of `InjectMarker`,
  `mergeSectionBased`, `createSnapshot`, `WriteLineageEntry` MUST be
  backward-compatible — existing callers continue to work byte-identical.
- **PRESERVE**: the existing `## Project-Specific Configuration (Harness-Generated)`
  block in CLAUDE.md (produced by `InjectMarker`). The `BlockTypeHarnessGenerated`
  enum value preserves this block's behavior verbatim.
- **DO NOT** introduce `AskUserQuestion` calls into `internal/harness/curator/`
  or any harness package (REQ-HEV2-035, C-HRA-008).
- **DO NOT** modify `.moai/state/routing-ledger.jsonl`,
  `.moai/state/routing-pending-*.json`, or `.moai/harness/usage-log.jsonl`
  in tests — use `t.TempDir()`.
- **DO NOT** add a new hook wrapper, a new settings.json hook registration,
  or a new gate (REQ-HEV2-036).
- **DO NOT** activate L2 Canary, L3 Contradiction, or re-proposal suppression
  — those are EVOLVE-003 (§E Out of Scope).
- **DO NOT** introduce internal SPEC IDs (`SPEC-HARNESS-EVOLVE-*`), REQ/AC
  tokens, internal dates, or commit SHAs into
  `internal/template/templates/CLAUDE.md` (template-internal-content isolation,
  §25 / REQ-HEV2-030).
- **DO NOT** use `--no-verify` on git commits; never force-push to main.
- **DO NOT** use `git add -A` or `git add .` — always specific-path
  (`git add .moai/specs/SPEC-HARNESS-EVOLVE-002/` for plan-phase,
  `git add internal/harness/curator/ internal/harness/layer3.go ...` for
  run-phase). Kitchen-sink hazard with the parallel v3.0-tokenomics work
  stream in flight.
- **REQUIRED**: Conventional Commit subjects —
  - Plan-phase (this commit): `feat(SPEC-HARNESS-EVOLVE-002): plan-phase artifacts (L, 5 artifacts)`
  - Run-phase M{N}: `feat(SPEC-HARNESS-EVOLVE-002): M{N} <subject>` or `fix(SPEC-HARNESS-EVOLVE-002): M{N} <subject>`
  - Sync-phase: `docs(SPEC-HARNESS-EVOLVE-002): sync-phase artifacts`
- **REQUIRED**: `🗿 MoAI` trailer on every commit.
- **REQUIRED**: `make build` after ANY edit to
  `internal/template/templates/CLAUDE.md` (recompiles embedded assets via
  `//go:embed all:templates`).

## §E. Self-Verification (run-phase deliverable — owned by manager-develop)

Each E-item is reported per the verification-claim-integrity 5-section format
(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

- **E1 AC Binary PASS/FAIL Matrix** — every AC-HEV2-XXX in acceptance.md
  with PASS/FAIL + verification command + actual output.
- **E2 Cross-Platform Build** — `go build ./...` AND
  `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- **E3 Coverage measurement** — `go test -cover ./internal/harness/curator/... ./internal/harness/...`
  ≥ 90% statement coverage on new + extended packages.
- **E4 Subagent Boundary Grep (C-HRA-008)** —
  `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ | grep -v "_test.go" | grep -v "// "`
  yields 0 matches.
- **E5 Lint Status** — `golangci-lint run --timeout=2m` exit 0; NEW vs
  pre-existing baseline distinguished.
- **E6 Branch HEAD + Push state** — list of new commit SHAs, result of
  `git push origin <branch>`.
- **E7 Blocker Report** — any NEEDS-CLARIFICATION item from §H that the
  orchestrator has not yet resolved, returned as a structured blocker (NOT
  asked via `AskUserQuestion`).

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Typed Managed-Block Writer foundation

- Author `internal/harness/curator/writer.go` with the
  `WriteManagedBlock(path, blockType, content)` API, `BlockType` enum
  (`BlockTypeLearnedWorkflow`, `BlockTypeHarnessGenerated`), and
  `BlockContent` / `Bullet` types per spec.md §D.1.
- Generalize the marker-block regex from `layer3.go markerBlockPattern` into
  a per-`BlockType` marker registry (each type maps to its heading + start
  marker + end marker per spec.md §D.2).
- Refactor `layer3.go InjectMarker` into a thin wrapper that calls
  `curator.WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)`. Existing
  tests in `layer3_test.go` MUST continue to pass byte-identical (B2
  backward compatibility).
- Per-bullet CRUD operations: `AddBullet`, `UpdateBullet`, `DeleteBullet`
  (REQ-HEV2-007). Each locates the bullet by `LedgerKey` and rewrites only
  the targeted line.
- Idempotency (REQ-HEV2-003), byte-preservation (REQ-HEV2-004), append-mode
  (REQ-HEV2-005).
- Table-driven tests with `t.TempDir()` isolation. Coverage ≥ 90%.

### M2 — `MOAI:LEARNED-WORKFLOW` digest block + budget/cap enforcement

- Author `BlockTypeLearnedWorkflow` marker contract (heading + start/end
  markers per spec.md §D.2). Marker-pattern unit tested.
- Wire the ≤3,000-character digest-budget enforcement via
  `internal/config/token_budget_guard.go measureAlwaysLoaded` (REQ-HEV2-008).
  The function may need a per-section attribution extension so the
  `MOAI:LEARNED-WORKFLOW` block's contribution is measurable distinctly —
  this extension is additive (no breaking change to the existing function's
  return type).
- Wire the ≤20-bullet cap (REQ-HEV2-009).
- Wire the `ledger_key` cross-layer linkage on each bullet (REQ-HEV2-010)
  in the trailing HTML comment form `<!-- key: <ledger_key> -->`.
- Anti-fabrication input validation (REQ-HEV2-011): the writer rejects any
  bullet text matching the forbidden patterns (internal SPEC IDs, REQ/AC
  tokens, ISO dates, commit SHAs) per the §25 forbidden-classes regex.

### M3 — CLAUDE.local.md append-only Learned section

- Author `BlockTypeLearnedLocal` (or equivalent) for the
  `## MOAI:LEARNED-WORKFLOW-LOCAL` section (marker naming per NEEDS
  CLARIFICATION H-1; default proposed in spec.md §C.3).
- Append-only writer (REQ-HEV2-012): the writer appends new entries between
  the start/end markers without modifying existing entries' bytes.
- Pattern-key deduplication (REQ-HEV2-014): append with an existing
  `ledger_key` returns `ErrDuplicateAppend` without writing.

### M4 — `mergeSectionBased` preservation

- Extend `internal/merge/strategies.go mergeSectionBased` to recognize the
  `MOAI:LEARNED-WORKFLOW` heading as a preserved managed section (REQ-HEV2-019).
  The recognition mechanism (explicit registration vs auto-recognition) is
  per NEEDS CLARIFICATION H-2.
- Test: when upstream template carries an empty marker and local carries a
  populated block, merge preserves local verbatim (no clobber — REQ-HEV2-020).
- Test: when upstream and local both carry content and the content conflicts
  inside the marker boundaries, the merge surfaces a conflict (NOT auto-
  resolved).

### M5 — Snapshot / rollback / lineage extension

- Extend `internal/harness/applier.go createSnapshot` to snapshot the
  CLAUDE.md managed block and CLAUDE.local.md Learned section as distinct
  restore units (REQ-HEV2-021). The snapshot manifest gains per-surface
  entries with byte-length recorded.
- Extend `RestoreSnapshot` for byte-identical rollback of each surface
  (REQ-HEV2-022). Post-rollback integrity verified via byte-length check.
- Extend `internal/harness/types.go LineageEntry` with `LearnedSurface`,
  `BulletsChanged`, `SnapshotDir` fields (additive — spec.md §D.4).
- Extend `internal/harness/lineage.go WriteLineageEntry` callers in the
  applier path to populate the new fields on every Curator write
  (REQ-HEV2-023).
- Evidence-or-null (REQ-HEV2-024): `null` in evidence-reference fields where
  no machine signal exists.

### M6 — Template-First + template neutrality + 2-layer Recall contract

- Template-First edit: add the EMPTY `MOAI:LEARNED-WORKFLOW` block marker to
  `internal/template/templates/CLAUDE.md` (REQ-HEV2-029). The marker carries
  heading + start/end markers + zero bullets + NO internal SPEC IDs / REQ
  tokens / dates / SHAs (REQ-HEV2-030, §25 neutrality).
- `make build` to recompile embedded assets.
- Run `internal/template/internal_content_leak_test.go` +
  `.github/workflows/template-neutrality-check.yaml` — MUST PASS.
- Author the 2-layer Recall contract documentation in the Curator package
  godoc (REQ-HEV2-015..018). The contract names the digest layer, the ledger
  layer surfaces, the cross-layer `ledger_key` linkage, and the "remember
  everything ✗, search when needed ○" principle. (The contract is
  formalized as types + godoc; the *consumption wiring* is EVOLVE-005.)

### M7 — Integration verification (spec.md M2 verification target)

- End-to-end test: AddBullet → VerifyBudgetEnforced → snapshot → DeleteBullet
  → RestoreSnapshot → byte-identical-rollback check → LineageEntry audit
  trail (REQ-HEV2-003..024).
- Template-merge round-trip: populate local CLAUDE.md block → run
  `mergeSectionBased` against template empty-marker → verify preservation
  (REQ-HEV2-019/020).
- Tier-differentiated Curator proposal test: tier-3-qualified pattern →
  CLAUDE.local.md append; tier-4-qualified pattern → CLAUDE.md write;
  under-tier pattern → `ErrTierNotQualified` (REQ-HEV2-025..027).
- L5 approval gate test: writer returns proposal artifact → orchestrator
  AskUserQuestion simulation → approval token → writer executes; rejection
  → writer records `rejected` in lineage + does NOT touch file
  (REQ-HEV2-032).

## §G. Anti-Patterns

- **AP-HEV2-001 — Wholesale block rewrite instead of bullet CRUD**: the
  Curator default interface is per-bullet operations (REQ-HEV2-007).
  `ReplaceAllBullets` exists but is the rare wholesale-rewrite path, NOT the
  default. Wholesale rewriting on every Curator run causes context-collapse
  (design SSOT §2 — "Curator is forbidden from full rewrite").
- **AP-HEV2-002 — Verbatim user text in digest bullets**: the digest carries
  *distilled generic workflow knowledge*; verbatim user text or internal
  SPEC IDs / REQ tokens / dates / SHAs in the block is a §25 violation
  (REQ-HEV2-011). The writer's input-validation regex rejects these.
- **AP-HEV2-003 — Autonomous CLAUDE.md write without L5 approval**: no
  autonomous write path exists (REQ-HEV2-032). A code path that calls
  `WriteManagedBlock` without first obtaining an L5 approval token is a
  violation; the orchestrator owns the approval channel.
- **AP-HEV2-004 — Breaking `InjectMarker` backward compatibility**: the
  legacy `InjectMarker` entry point is preserved as a thin wrapper over the
  typed writer (REQ-HEV2-002). Existing `layer3_test.go` tests MUST continue
  to pass byte-identical.
- **AP-HEV2-005 — Touching the routing-ledger writer**: EVOLVE-001's
  `internal/harness/routing/` is read-only consumption (REQ-HEV2-016). Any
  edit to the routing writer is a boundary violation.
- **AP-HEV2-006 — Shipping populated block content to template**: the
  template tree ships an EMPTY marker only (REQ-HEV2-029). Populated bullets
  in `internal/template/templates/CLAUDE.md` is a §25 violation.
- **AP-HEV2-007 — Silently clobbering local block during `moai update`**:
  the merge recognizes the managed block as a preserved section (REQ-HEV2-019).
  Auto-resolving a content conflict by taking the upstream empty block is
  silent clobber and is prohibited (REQ-HEV2-020).
- **AP-HEV2-008 — Partial rollback state**: rollback MUST be byte-identical
  (REQ-HEV2-022). A rollback that leaves a bullet half-deleted or a marker
  orphaned is a partial-state violation.
- **AP-HEV2-009 — Self-tier-escalation**: a 6-observation pattern stays
  Tier 3; it cannot escalate itself to Tier 4 without 4 more observations
  (REQ-HEV2-027). Self-tier-escalation is a reward-hacking shape.
- **AP-HEV2-010 — Inferred evidence in lineage**: where no machine signal
  exists, the lineage records `null` (REQ-HEV2-024). Inferred or fabricated
  values in evidence-reference fields violate the evidence-or-null principle.

## §H. NEEDS CLARIFICATION (orchestrator AskUserQuestion before run-phase)

The following items require orchestrator-mediated user clarification before
Implementation Kickoff Approval. The orchestrator runs the AskUserQuestion
rounds and re-delegates with the answers injected. These markers are the
plan-phase inputs to that flow (per `moai-workflow-spec` § NEEDS CLARIFICATION
Marker Convention).

### H-1 — CLAUDE.local.md section marker naming

The spec.md proposes `## MOAI:LEARNED-WORKFLOW-LOCAL` +
`<!-- moai:learned-local-start -->` / `<!-- moai:learned-local-end -->` for
the append-only Tier 3 surface (REQ-HEV2-013). Alternatives:
- (a) `## MOAI:LEARNED-LOCAL` (shorter, parallels the digest block name)
- (b) `## MOAI:LEARNED-WORKFLOW (local)` (parenthetical, matches the
  existing `## Project-Specific Configuration (Harness-Generated)` style)
- (c) the proposed default `## MOAI:LEARNED-WORKFLOW-LOCAL`

The choice is non-blocking — the writer API accepts the marker names as
configurable per `BlockType`. The default is `(c)` per spec.md; the
orchestrator's AskUserQuestion round confirms or substitutes.

### H-2 — `mergeSectionBased` recognition mechanism

`internal/merge/strategies.go mergeSectionBased` parses markdown by `## ...`
headings. Two integration options for the new managed block:
- (a) **Explicit registration**: the merge code carries a registered list of
  managed-section headings (`Project-Specific Configuration (Harness-Generated)`,
  `MOAI:LEARNED-WORKFLOW`, ...) and treats them as opaque preserved units.
- (b) **Auto-recognition via marker pair**: any section whose body is bounded
  by a `<!-- moai:*-start -->` / `<!-- moai:*-end -->` marker pair is auto-
  recognized as a managed section.

Option (a) is more conservative (explicit allow-list) but requires updating
the registration on each new managed block type. Option (b) is more general
but couples merge behavior to marker syntax. The spec.md default leans (a)
for conservatism; the orchestrator's AskUserQuestion round confirms.

### H-3 — Debug CLI verb scope

This SPEC authors the write-path machinery. The SSOT §6 console verbs
(`evolve` / `promote` / `demote` / `freeze` / `unfreeze`) are EVOLVE-004.
However, a debug-only verb `moai harness curator write` (or similar) might
be useful for M7 integration verification. Options:
- (a) NO CLI verb in this SPEC — M7 integration verification drives the
  writer via Go test calls only. (Default — keeps this SPEC focused on the
  write-path machinery.)
- (b) A minimal debug CLI verb `moai harness curator debug-write` for
  integration verification + future EVOLVE-004 reuse.

The spec.md default is (a); the orchestrator's AskUserQuestion round
confirms.

## §I. Risks (from SSOT §7 risk grid, this SPEC's slice)

- **HIGH — CLAUDE.md contamination / budget overflow**: mitigation is
  REQ-HEV2-008 (≤3K budget enforcement via `measureAlwaysLoaded`) +
  REQ-HEV2-009 (≤20-bullet cap) + REQ-HEV2-018 (digest carries summaries
  only, detail in ledger) + REQ-HEV2-011 (anti-fabrication input validation).
- **HIGH — Wrong rule self-reinforcement**: mitigation is REQ-HEV2-031
  (machine-signal-only write authority) + REQ-HEV2-033 (mechanical rollback
  trigger) + REQ-HEV2-032 (L5 approval gate, no autonomous write path).
- **MID — Parallel session ledger/block race**: mitigation is REQ-HEV2-021
  (distinct restore units) + the existing Pre-Spawn Sync Check discipline
  (agent-common-protocol.md § Pre-Spawn Sync Check) + the append-only
  jsonl pattern for ledger writes (inherited from EVOLVE-001
  REQ-HEV-007) + per-session pending isolation.
- **MID — Template learning-data leak**: mitigation is REQ-HEV2-029 (empty
  marker only) + REQ-HEV2-030 (template isolation) + REQ-HEV2-011 (input
  validation). CI guards: `internal/template/internal_content_leak_test.go`
  + `.github/workflows/template-neutrality-check.yaml`.
- **LOW — L5 approval fatigue**: mitigation is the tier-differentiated
  proposal rate (most patterns stay Tier 1-2 auto-memory, never reach the
  L5 gate) + the EVOLVE-004 `evolve` batch-review verb (later Epic
  milestone).

## §J. Cross-References

- `spec.md` — requirements SSOT (GEARS REQ-HEV2-001..036)
- `acceptance.md` — AC matrix SSOT (AC-HEV2-001..NNN)
- `design.md` — architecture decisions + alternatives considered
- `research.md` — codebase investigation + existing-asset inventory
- `progress.md` — §E lifecycle skeleton (§E.1 plan-phase audit-ready signal
  populated by manager-spec at plan-phase; §E.2-§E.4 left as placeholder
  headings per the progress.md Section Map in spec-frontmatter-schema.md)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status
  Transition Ownership Matrix — `(none) → draft` row (plan-phase commit
  subject pattern: `feat(SPEC-HARNESS-EVOLVE-002): plan-phase artifacts (L, 5 artifacts)`)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` —
  Section A-E delegation template (this SPEC is Tier L → REQUIRED full
  template at run-phase)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier —
  Tier L (5 artifacts) classification
- `feedback_ac_token_presence_not_reachability` lesson — AC reachability
  discipline (every cross-file registration pinned as SEPARATE baseline-0 AC)
- `feedback_spec_artifact_commit_before_parallel_work` lesson — commit
  promptly after authoring (4 active sessions; specific-path `git add` only)
