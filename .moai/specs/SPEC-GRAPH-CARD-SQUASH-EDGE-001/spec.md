---
id: SPEC-GRAPH-CARD-SQUASH-EDGE-001
title: "graph card-file layer: walk every commit so squash landings produce card edges"
version: "0.1.0"
status: draft
created: 2026-10-07
updated: 2026-10-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph"
lifecycle: spec-anchored
tags: "graph, card-file-edges, squash-landing, git-log, freshness-fingerprint, github-flow"
tier: M
---

# SPEC-GRAPH-CARD-SQUASH-EDGE-001 — Card-file edges for squash landings

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-07 | Initial plan-phase authoring (card t1559, lane-3 self-dispatch run tmhxo0). Defect mechanism and the squash-attribution premise verified by executed measurement against the card worktree at df0c417e9. | manager-spec |

## §1 Context and Problem

`internal/graph/card_file.go` derives card→file edges from committed evidence: the file-local
walk `walkCardMerges` runs `git log --merges --format=%H%x00%s HEAD` (card_file.go:49), and
`CardFileEdges` keeps only the commits whose subject the caller-supplied attributor
(`factory.AttributeSubject`) maps to exactly one card. The `--merges` parent-count filter
predates the 2026-10-05 GitHub Flow cutover, when card landings were `--no-ff` merge commits.
Under the current flow a card lands as a **squash commit — a single-parent commit** — so the
walk never sees it and the layer produces no edges for it.

### Executed baseline (card worktree, HEAD df0c417e9, 2026-10-07)

| Measurement | Command | Observed |
|---|---|---|
| HEAD is a squash landing | `git rev-list --parents -1 HEAD` | `df0c417e97178056353c0dcbd39b41d9c07fc7cb cb2a011d033527c6f1b9f311c7209579eba2c9f5` — HEAD plus exactly one parent |
| HEAD subject carries the card + PR groups | `git log -1 --format=%s` | `fix(ci): compile-guard the FIFO regression test for windows + re-anchor the codemaps stamp (card t1563) (#1782)` |
| The attributor already reads this shape | HEAD subject vs form-2b `\([^()]*t[0-9]+[^()]*\) \(#[0-9]+\)$` | 1 match |
| Commits the current walk skips | `git rev-list --count HEAD` vs `git rev-list --count --merges HEAD` | 13127 vs 2664 — 10463 single-parent commits invisible to the walk |

Consequences: `CardFileEdges` (consumer at internal/graph/graph.go:239) emits no edges for
squash-landed cards; `CardMergeFingerprint` / `CardAttributedMergeSHAs` (consumer at
internal/graph/meta.go:59, freshness source `card-merges`) miss squash landings, so the
fingerprint under-converges and the graph-freshness check cannot stabilize after a squash
landing.

The attribution side needs no change: form 2 (`(card tN)` trailing), form 2b
(`(card tN) (#PR)`), form 2c (comma form) and form 1 (`fix(tN):` scope) all attribute
single-parent landing subjects already — the `--merges` filter is the only blocker.

### Design invariants to preserve (card_file.go doc comments, MU-86/MU-87)

1. Attribution stays the caller-supplied single point (`CardFileAttributor` wired to
   `factory.AttributeSubject`); the layer never grows a second matcher.
2. Walk breadth remains "reachable from HEAD by ANY parent path" — the absorb-merge
   second-parent path keeps being walked (MU-86).
3. Absorb-direction merges attribute nothing (the non-attribution rule mechanically refuses
   them; MU-87).
4. Deterministic output — git log order plus the existing sort; two runs byte-identical.

## §2 Requirements (GEARS)

### §2.1 Walk semantics

- REQ-GCSE-001: The card→file edge layer's history walk shall consider every commit reachable from HEAD by any parent path — merge and single-parent commits alike — with subject attribution as the only card filter (anchor: the file-local walk function, renamed `walkCardCommits`).
- REQ-GCSE-002: **When** a card-attributed single-parent landing (a squash commit whose subject the attributor maps to exactly one card) is reachable from HEAD, the layer shall produce one card→file edge per (landing, changed file) pair from the landing's first-parent diff, with the landing commit's abbreviated SHA as the evidence pointer.
- REQ-GCSE-003: **When** a card-attributed single-parent landing enters the reachable history, the fingerprint input (`CardAttributedMergeSHAs`) shall include that landing's full SHA, so the `card-merges` freshness source converges after squash landings.
- REQ-GCSE-004: **When** the broadened walk reaches a parentless (root) commit, the layer shall contribute no edge for it and shall return no error — the existing first-parent-diff failure guard (the per-commit `continue`) stays the mechanism.

### §2.2 Invariants preserved

- REQ-GCSE-005: The layer shall keep the caller-supplied `CardFileAttributor` as its single attribution point and shall not grow a second subject matcher.
- REQ-GCSE-006: The walk shall keep full reachable breadth — including the absorb-merge second-parent path (MU-86) — so `TestGraphCardFileEdgesSeeAbsorbedMerge` behavior is unchanged.
- REQ-GCSE-007: Absorb-direction merges shall attribute nothing.
- REQ-GCSE-008: Two runs of `CardFileEdges` and of `CardMergeFingerprint` over the same tree and reachable history shall return byte-identical output.

### §2.3 Tests and hygiene

- REQ-GCSE-009: A RED-first reproduction test `TestGraphCardFileEdgesSeeSquashLanding` shall fail on the pre-fix tree and pass after the fix, with the test commit preceding the fix commit so the RED→GREEN ordering is witnessed by the commit graph.
- REQ-GCSE-010: The six existing card-file tests shall keep passing with their fixture patterns unmodified.
- REQ-GCSE-011: The layer's file-top comment and doc comments shall state commit-level semantics — the walk function renamed `walkCardCommits`, the `CardFileEdges` / `CardAttributedMergeSHAs` / `CardMergeFingerprint` doc comments updated — no exported identifier shall be renamed, and no file outside `internal/graph/card_file.go` and `internal/graph/card_file_test.go` shall change.
- REQ-GCSE-012: The affected package shall pass `go vet ./internal/graph/...`, `golangci-lint run ./internal/graph/...`, and `go test -timeout 30m ./internal/graph/...`.

## §3 Non-Functional Constraints

- Performance (measured): the broadened walk grows the first-parent diff workload from 681 to
  5,851 runs (≈8.6×) — measured by extracting the attribution implementation and counting
  attributed commits at the fixed authoring base df0c417e9 (codex extraction, 2026-10-07); the
  audit's independent consistency probe (1,303/2,664 merges vs 6,591/13,127 card-token
  subjects) confirms the scale and direction. No commit-limit, pagination, or caching is
  engineered (§4); the wall-clock cost of one full affected-package run is recorded once at
  plan M3 so the acceptance picture carries a measured number.
- Security lens (dispatch `--security`): the walk runs the same `git` subprocess with one flag
  removed — no new input surface; the NUL-separated `%x00` subject parsing and the `-z`
  first-parent-diff parsing stay as-is; `repoRoot` and `landedBranch` provenance unchanged.
- Determinism: no wall-clock and no map-iteration-order output; git log order plus the existing
  three-key sort.
- Attribution semantics — intermediate-commit attribution is ACCEPTED (lane decision, recorded
  2026-10-07): every intermediate branch commit carries the card id by the traceability
  mandate, and pre-cutover `--no-ff` merges make those commits reachable from `main`, so the
  broadened walk attributes them — a card's edges now span its intermediate commits, not only
  the landing, on `main` as well as in in-flight worktrees. This is semantically consistent
  landed evidence (the card did touch those files) and a ONE-TIME historical enrichment: under
  the current squash-landing flow each new landing attributes exactly one new commit, so no
  ongoing churn follows. No walk-side filter is added — attribution remains the single filter
  (REQ-GCSE-001; a walk-side subject filter is plan §G's first anti-pattern).

## §4 Out of Scope

### Out of Scope — consumer files and naming

- `internal/graph/meta.go` and `internal/graph/graph.go` stay untouched — the freshness source
  name `card-merges` keeps its spelling even though the source now covers squash landings too
  (a naming imprecision recorded in plan.md §B; a later cosmetic rename is behavior-neutral).
- No exported identifier is renamed: `CardAttributedMergeSHAs` and `CardMergeFingerprint` keep
  their names (renaming would force changes in files this SPEC excludes).

### Out of Scope — attribution engine

- `internal/factory/prlink_landed.go` attribution forms are not modified — forms 1/2/2b/2c
  already attribute squash landing subjects (measured, §1).

### Out of Scope — performance engineering

- No commit-count limits, pagination, or caching for the broadened walk.

### Out of Scope — workflow doctrine

- The GitHub Flow landing-chain documents and any CI workflow changes.

## §5 Cross-References

- `SPEC-TODO-CARD-ISSUANCE-001` (REQ-TCI-016) — the card→file edge layer this SPEC repairs.
- `SPEC-WORKTREE-SQUASH-MERGE-001` — the sibling squash-detection problem in worktree cleanup;
  same GitHub Flow squash-landing root cause, different layer.
- `SPEC-V3R6-GRAPH-FRESHNESS-001` — the graph freshness discipline whose convergence this SPEC
  restores for squash landings.
