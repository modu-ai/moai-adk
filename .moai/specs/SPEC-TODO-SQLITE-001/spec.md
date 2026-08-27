---
id: SPEC-TODO-SQLITE-001
title: "SQLite-backed backlog queue store with .moai/state/todo rename"
version: "0.1.0"
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, todo, sqlite, storage, migration, backlog, queue"
related_specs: [SPEC-KANBAN-TODO-CLI-001, SPEC-WEB-TODO-QUEUE-001]
tier: L
---

# SPEC-TODO-SQLITE-001 — SQLite-backed backlog queue store with .moai/state/todo rename

## A. Purpose and Scope

Factory card t306 (absorbing card t309): replace the backlog queue's single-JSON-file
persistence (`.moai/state/<dir>/backlog.json`, measured 118 KB / 82 items / 1 finding,
live-growing) with a single-file SQLite store, and rename the project-local state
directory `.moai/state/kanban/` to `.moai/state/todo/`. The rename rides this card because
a schema/path redefinition moment is the only zero-cost point for it.

In scope:

- The backlog queue store (`internal/kanban/backlog_store.go`): Load/Mutate/Add and the
  queued-count helpers, over a SQLite database at `.moai/state/todo/backlog.db`.
- Lazy, lossless one-time migration from legacy `backlog.json` to the database.
- The directory rename, including the per-session registry files (`<uuid>.json`,
  `internal/kanban/record.go` family) that share the directory, and a fallback READ of
  the old directory when only it exists.
- Consumer sweep: `internal/cli/todo*`, `internal/cli/graph.go`,
  `internal/cli/kanban.go` (companion/lead registry path helpers), `internal/web`
  (events watch map, view model, queue read seam), `internal/statusline/backlog.go`,
  `internal/hook/session_start_*`.
- Template copies of user-facing workflow docs that name the path (Template-First),
  plus a documented JSON export path for downgrade rollback.

Out of scope: see §F.

## B. Requirements (GEARS)

Numbering note: requirement IDs are contiguous (REQ-TOSQ-001..018) per this
repository's modern-SPEC convention; grouping into §B.1–§B.5 below is thematic only,
never numeric.

### B.1 Storage engine

- REQ-TOSQ-001: The backlog queue store shall persist every element of the queue
  record — items, findings, the persisted sequence high-water mark (`last_seq`),
  and a schema version marker — inside one SQLite database file at
  `<queue-root>/.moai/state/todo/backlog.db`.

- REQ-TOSQ-002: **Where** the build runs with `CGO_ENABLED=0` (the repository's CI
  cross-build posture, `.github/workflows/ci.yml` GOOS matrix), the binary shall compile
  and pass its package tests for darwin/arm64, darwin/amd64, linux/amd64, and
  windows/amd64 using a pure-Go SQLite driver.

- REQ-TOSQ-003: **While** a database connection is open, the store shall configure
  SQLite for a many-writer desktop environment: journal mode WAL and a busy timeout of
  at least 5000 ms, so a writer contending with another process waits rather than
  surfacing a spurious failure.

- REQ-TOSQ-004: `Load` shall return items in insertion order, reproducing exactly the
  array order the legacy JSON file preserved (queued, picked, and dropped cards alike —
  positions survive state changes and removals of other rows).

- REQ-TOSQ-005: The store shall enforce id integrity at the storage layer itself: a
  UNIQUE constraint on the item id, and `last_seq` advanced transactionally with item
  insertion inside one write transaction, such that no committed mutation can mint a
  duplicate or reused id even when a process dies mid-mutation.

- REQ-TOSQ-006: **When** the store meets an engine-level fault (lock timeout /
  SQLITE_BUSY beyond the busy window, corruption, constraint violation), it shall map
  the fault onto the package's named error taxonomy (in the style of
  `IsBoardLockHeld`) and surface the named error to the caller; under no such fault
  shall the store delete or overwrite either the database or any quarantined legacy
  artifact as a recovery action.

### B.2 Behavior preservation

- REQ-TOSQ-007: Every existing `moai todo` verb (`add [--pick] [--force]`, `list`,
  `done`, `next [<n>]`, `unpick`, `drop`, edit/move, `analyze`, `relate`, `unrelate`,
  `why`, `pr`) shall retain identical command-line flags, stdout rendering, and exit
  codes across the storage swap. No verb is renamed, removed, or re-flagged.

- REQ-TOSQ-008: **While** multiple processes mutate the same queue concurrently (factory
  mode runs up to ten lanes against one queue), mutations shall serialize exactly as the
  legacy store serialized them — the existing sibling advisory file lock
  (`backlog.lock`, held across the entire read-modify-write) SHALL remain the outer
  serialization mechanism, with the engine's own locking as a backstop — such that N
  concurrent adds yield N unique sequential ids and zero lost updates.

- REQ-TOSQ-009: The statusline backlog count (rendered per refresh interval) shall read
  the store through a constant-cost path whose per-render added latency stays within the
  budget fixed in §C-2, failing open to an empty display exactly as today.

- REQ-TOSQ-010: The set of exported symbols outside `package kanban` that consume the
  store (`BacklogStore`, `BacklogPathForRoot`, `ResolveTodoQueueRoot[Adopting]`,
  `RecordPath`, `QueuedBacklogCountForRoot`, the record types) shall keep their existing
  signatures and observable contracts; callers swap nothing but their import-time
  expectations of where bytes live.

### B.3 Migration

- REQ-TOSQ-011: **When** the database file is absent and a legacy `backlog.json` exists
  at the resolved queue root, the first operation that opens the store shall migrate
  losslessly: item count AND field-level equality per item, findings including tuple and
  order, and `last_seq`, verified BEFORE the new layout becomes authoritative.

- REQ-TOSQ-012: **When** the migration parity check fails, the store shall abort the
  cutover, leave the legacy JSON authoritative and untouched, remove any partially
  written database artifact, and surface a structured error naming both files — the
  caller keeps serving the queue it had.

- REQ-TOSQ-013: **When** both layouts are present under one queue root — a database
  alongside an un-quarantined legacy JSON, reachable after a crash between commit and
  quarantine — the store shall treat the database as authoritative on every read and
  write path, and shall attempt completion of the quarantine as an idempotent
  best-effort step retried on each subsequent open whose failure shall not fail the
  open itself.

- REQ-TOSQ-014: On a successful migration the legacy file shall be RENAMED to
  `backlog.json.migrated` (byte-preserved quarantine), never deleted. No code path in
  this SPEC deletes operator queue data.

- REQ-TOSQ-015: **Where** path resolution is root-parametric (`BacklogPathForRoot`
  plus the pure-vs-adopting resolvers), the relocation rules of this requirement shall
  apply uniformly at every resolved root — primary checkout and home-based fallback
  alike. **When** only the old directory `.moai/state/kanban/` exists under a root, the
  first adopting open shall relocate its contents — the queue file AND the per-session
  registry files (`<uuid>.json`) — to `.moai/state/todo/`. **While** both directories
  exist, every code path shall resolve queue and registry data from
  `.moai/state/todo/` only and shall leave `.moai/state/kanban/` strictly untouched
  (stale-copy policy mirroring `adoptLocalTodoQueue`'s leave-the-original rule).

### B.4 Rollback to plain JSON

- REQ-TOSQ-016: The binary shall provide a deliberate JSON regeneration entry point
  (additive CLI verb; the existing verb surface is untouched per REQ-TOSQ-007) that
  writes the CURRENT live database state out as a valid legacy-format `backlog.json`
  at the same queue root, so a downgraded earlier binary — which reads only that file
  name and ignores the presence of a `.db` — resumes full-fidelity service.

- REQ-TOSQ-017: The downgrade procedure shall be documented (docs tree, with the
  exact commands and the meaning of each artifact: `.db`, `-wal`/`-shm`,
  `backlog.json.migrated`). A config knob selecting storage engines is DELIBERATELY
  rejected: two always-live engines reintroduce the silent-divergence class this SPEC
  exists to close.

### B.5 Source and template hygiene

- REQ-TOSQ-018: After implementation, no active template source under
  `internal/template/templates/` and no production Go source outside the intentional
  old-directory FALLBACK READER(s) shall contain the literal `state/kanban`;
  template edits land template-first and the embedded bundle is rebuilt
  (`make build`) before commit.

## C. Constraints

- C-1 Coverage: package `internal/kanban` ≥ 85% statements overall on changed-path
  files; affected command paths in `internal/cli` ≥ 90% (per TRUST 5 Tested gate).
- C-2 Statusline budget: added per-render latency of the backlog-count read ≤ 10 ms
  median and ≤ 25 ms hard ceiling versus the current whole-file JSON read, measured
  against a generated fixture of ≥ 500 items on the development machine; measurement
  evidence recorded in progress notes.
- C-3 Cross-platform: windows/amd64 remains release-blocking; verification uses
  `GOOS=windows go vet ./internal/...` for compile evidence PLUS the CI windows job as
  the behavioral verdict (local vet proves compilation only).
- C-4 Engine constraints: pure-Go driver only (CI builds with CGO disabled — measured);
  binary-size delta versus the 62 MB shipped baseline is measured at M1 and recorded;
  a delta above +12 MB triggers a driver-configuration review rather than silent
  acceptance.
- C-5 Privacy: test fixtures are SYNTHESIZED from the observed schema shape
  (field names, value domains, mixed states); production queue contents (operator card
  texts) are never committed to the repository.
- C-6 Lossless doctrine: migration, fallback reads, and quarantine steps never destroy
  or truncate operator data under any input, including malformed legacy files (abort +
  surface, never repair-by-delete — mirrors the existing malformed-load contract).

## D. Success Criteria

- SC-1: All ACs in acceptance.md pass, including migration parity, concurrency
  stress, cross-compile, and template-literal cleanliness.
- SC-2: `go test ./internal/kanban/... ./internal/cli/... ./internal/web/...`
  (affected packages) green; `golangci-lint run` clean on touched packages.
- SC-3: The live-operator upgrade path is rehearsal-provable end to end on a scratch
  root: legacy JSON seeded → first command migrates + quarantines → ids continue past
  the legacy high-water mark → export-json reproduces a valid legacy file.
- SC-4: No `[NEEDS CLARIFICATION]` markers remain in plan/research artifacts at
  kickoff.

## E. Background

The queue is the delegation channel between sessions (lead, foreman loop, operator
picks all read ONE queue), which is why the store shares a directory with the
per-session kanban records and why both travel together under the rename. The id
space is anchored by the persisted high-water mark `last_seq` — never max-present-id —
because done-cards delete rows and a derived mark would reuse ids. Two prior layers of
precedent constrain this work: `adoptLocalTodoQueue` (lossless adopt-or-shadow rules)
and `legacyBacklogLockFileName` (quarantine-not-delete of superseded artifacts).
Measured ground truth lives in research.md; physical schema and algorithms in design.md.

## F. Exclusions

### Out of Scope — Board storage

- The board state directory `.moai/state/kanban-board/` (`boardDirSegments`) keeps its
  name and its JSON persistence; only `.moai/state/kanban/` relocates.

### Out of Scope — Session-record storage model

- Per-session registry files stay individual JSON documents addressed by filename
  (`record.go` API unchanged); converting them to SQL tables buys nothing their
  random-key access pattern cannot already express.

### Out of Scope — Bulk migration of dormant fallback roots

- The ~200 accumulated home-side project roots (`~/.moai/todo/<key>/…`) migrate lazily,
  each on ITS OWN first open; no scanner sweeps them, and unopened stale roots remain
  valid-for-their-era forever.

### Out of Scope — Multi-backend abstraction

- No pluggable-storage interface, engine-selection config knob, or second
  always-maintained engine behind the store. One store, one engine, one documented
  export route back to JSON.

### Out of Scope — Web SSE event-key renaming

- The console's event channel KEY `"kanban"` (frontend-visible contract) keeps its
  name; only the watched filesystem path changes. UI label cleanup is separate work.

### Out of Scope — Performance beyond stated budgets

- No query-layer optimization work beyond the hot paths named in §B.2/C-2; the access
  pattern (hundreds of rows, sub-millisecond queries) leaves nothing that pays back
  deeper engineering.

### Out of Scope — Historical SPEC documents

- Earlier SPEC bodies (`SPEC-KANBAN-*-001`, …) that quote the old path are historical
  records and are not rewritten.

## G. History

| Date | Change |
|------|--------|
| 2026-08-27 | Initial draft (t306 plan-phase; absorbs t309 directory rename). |
| 2026-08-27 | Plan-audit iter-1 revision: contiguous REQ renumbering (001..018), shall-modality restored in REQ-013/015, statusline budget cite fixed to C-2, tier frontmatter added, consumer inventory extended (cli/kanban.go registries). |
