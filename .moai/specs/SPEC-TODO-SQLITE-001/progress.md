# SPEC-TODO-SQLITE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-27T09:46:32Z   # refreshed at run-phase M1 (finding N2): the
#   original 18:05 stamp predated the audit-driven artifact revisions. The value is now the
#   plan commit a4c135d2e's own committer timestamp (2026-08-27T18:46:32+09:00), the moment
#   every plan artifact was final. Measured: git log -1 --format=%cI a4c135d2e
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md
card: t306 (absorbs t309)
branch: WT-todo-sqlite @ d29b8942e (develop-based integration line)

## §E.2 Run-phase Evidence

### Milestone status

| Milestone | Status | Commit |
|---|---|---|
| M1 driver adoption + storage skeleton | complete | `3d24cf6df` |
| M2 store guts + migration state machine | complete | `83a1d492a` |
| M3 directory rename (t309) + fallback read | not started | — |
| M4 consumer sweep | not started | — |
| M5 templates + export-json + docs | not started | — |
| M6 cross-platform / race / gates | not started | — |

M2 absorbed the storage half of plan.md M3 (the lazy migration state machine,
REQ-TOSQ-011..014). The split was forced by a real seam rather than chosen: the
store swap and the migration are inseparable if the tree is to stay green — a
store that writes the database while reading the legacy JSON is the silent
divergence class this SPEC names as its enemy (plan.md §G). M3 therefore keeps
the DIRECTORY half it was really about: the `.moai/state/kanban` →
`.moai/state/todo` rename absorbed from t309, the registry census, and the
fallback read (REQ-TOSQ-015).

### PROCESS DEFECT — foreign commits on an actively-owned worktree

Between the M1 commit and the M2 commit, a foreign actor rewrote history in
this worktree. Observed, not inferred:

- M1 was committed by this agent as `b6fd82dbd`. A later
  `git merge-base --is-ancestor b6fd82dbd HEAD` reported NOT an ancestor; the
  same content now sits at `3d24cf6df` under a different parent.
- The M2 working tree was committed by another actor as `83a1d492a`
  ("M2 WIP — ..."), not by this agent.

No code was lost — every authored symbol was re-verified present
(`LoadPure`, `EnginePath`, `assertBacklogParity`, `quarantineLegacyBacklog`,
`TestConcurrencyStress`, `TestMigrationPartialFailureRemovesArtifacts`,
`TestLoadPureNeverMovesBytes`, `queueStateBytes`), the migration/spec/go.mod
changes are intact, and the tree was re-verified green at the rewritten HEAD
rather than assumed green.

What WAS lost is the M2 commit message body, which carried the mutation
evidence below. That is why the evidence is recorded here instead: a commit
body another actor can replace is not a durable evidence surface.
`agent-common-protocol.md` § Background Agent Execution requires this be
reported to the lead rather than continued past quietly.

### Safety mechanisms fired MECHANICALLY (operator's binding condition)

Every mechanism below was verified by MUTATING the implementation and observing
the guard go red, then restoring. A test that only asserts the happy path does
not discharge the condition, so none of these rows rests on one.

| # | Mechanism | Mutation applied | Observed on mutant |
|---|---|---|---|
| 1 | busy_timeout ≥ 5000 + WAL (REQ-TOSQ-003) | `backlogBusyTimeoutMS = 100`; WAL pragma removed | `TestBacklogEnginePragmas` FAIL — `journal_mode = "delete"`, `busy_timeout = 100` |
| 2 | corrupt/not-a-database classification (REQ-TOSQ-006) | `classifyBacklogEngineCode` returns nil for corrupt | `TestBacklogEngineNotADatabaseMapsToCorrupt` + `TestClassifyBacklogEngineCode` FAIL |
| 3 | refuse-never-repair on unknown schema (C-6) | `os.Remove(e.dbPath)` added to the refusal | `TestBacklogEngineRefusesUnknownSchemaVersionWithoutDestroying` FAIL — "the refusal removed the file, violating C-6" |
| 4 | whole-RMW under the lock (REQ-TOSQ-008) | lock acquire/release removed from `Mutate` | `TestConcurrencyStress` FAIL — `t7` issued 5×, `t16`/`t20`/`t18`/`t21` 2×, `t8` 3×, `t11` 4× |
| 5 | partial-artifact cleanup on aborted cutover (REQ-TOSQ-012) | `removeBacklogDBArtifacts` early-returns | `TestMigrationPartialFailureRemovesArtifacts` FAIL — "partial artifact backlog.db survived the aborted cutover" |
| 6 | quarantine never overwritten (REQ-TOSQ-014 / C-6) | existing-`.migrated` guard removed | `TestMigrationNeverOverwritesExistingQuarantine` FAIL — quarantine sha256 changed, "the original rollback source was destroyed" |
| 7 | parity verified BEFORE authority flips (REQ-TOSQ-011) | writer drops `spec_id` | migration ABORTS: "parity check failed, legacy file left authoritative: item 1 (t3): spec_id SPEC-EXAMPLE-001 != <nil>"; with the gate ALSO disabled the corruption reaches the queue (`spec_id` null-shape changed) — the gate is what stands between them |
| 8 | pure read never migrates (REQ-TOSQ-009 / REQ-WTQ-001) | `LoadPure` delegates to the adopting `load()` | `TestLoadPureNeverMovesBytes` FAIL ×3 subtests + the console's own guards `TestTodoSectionReadsThroughToProjectLocalQueue` / `TestConsoleRoutesLeaveBacklogUntouched` FAIL — "the console moved or removed the local queue" |

Mutation-run log: `.moai/state/verify/t306-m2/red.log`.

### A guard that did NOT bite, and what was done about it

The first form of the aborted-cutover assertion was VACUOUS. It seeded a
malformed legacy file and asserted no partial database survived — but a
malformed seed fails at the parse, before any database is created, so the
assertion could not distinguish a working cleanup from an absent one.
Disabling `removeBacklogDBArtifacts` left it green (mutant B, red.log).

It was replaced by `TestMigrationPartialFailureRemovesArtifacts`, which seeds a
legacy file carrying a DUPLICATE id: that parses fine, reaches the insert, and
trips `UNIQUE(id)` with the database already on disk — the only input that
actually exercises the removal path. Re-running the same mutation against the
replacement fires it (row 5 above).

### Design deviation recorded — `seq` is a POSITION, not the id's number

design.md §2 annotates the `items.seq` column "mirrors t<N>; preserves insertion
order". Those two clauses cannot both hold: `todo move`
(`internal/cli/todo_edit_move.go`) reorders `rec.Items`, after which array order
and id order differ. REQ-TOSQ-004 binds ARRAY order — "reproducing exactly the
array order the legacy JSON file preserved" — so `seq` carries the 1-based array
position and the id's number is not stored separately.

Id integrity is unaffected: it rests on `UNIQUE(id)` plus `meta.last_seq`
(REQ-TOSQ-005), neither of which reads `seq`. Guarded by
`TestReorderedItemsRoundTrip`, which moves the last card to the front and
asserts the order round-trips as `t3, t1, t2`. This is an implementation choice
inside a binding requirement, not a requirement change — spec.md is untouched.

### Behavioral-contract ports (REQ-TOSQ-007 / REQ-TOSQ-010)

Twenty-two existing tests failed on the storage swap. Every one was a PHYSICAL
assertion (`os.ReadFile` of the JSON queue); no behavioral assertion — flag,
stdout, exit code, refusal semantics — changed. Each was re-pointed at the new
backing with its MEANING preserved:

- `queueDigest` / `readBacklogBytes` now digest the stored RECORD in canonical
  document form via `LoadPure`, so "a refused verb changed nothing" is still the
  assertion; it just no longer measures database page layout and WAL checkpoint
  timing. The wording moved from "byte-identical" to "record-identical" to say
  what is actually checked.
- `TestTodoPR_QueueDirUnchanged` stats `store.EnginePath()` (the new physical
  carrier) instead of `Path()`.
- The document-shape contract (`findings` top-level key, exactly five per-item
  keys) is asserted on the canonical serialization of the stored record, which
  is the same shape the M5 downgrade export must regenerate.

One test changed MEANING rather than mechanism, and the change is recorded
rather than buried: `TestBacklogWrite_UnwritableDirFailsTempCreation` revoked a
directory's write bit from inside the mutation callback, which failed the old
temp-file write. The engine has no temp-file step and an already-open descriptor
keeps writing correctly through a directory that turns read-only underneath it,
so that scenario now produces no error and nothing to assert. It was re-pointed
at the hazard that still exists — a directory unwritable BEFORE the store opens
— and split in two, because the failure arrives from different places on the two
paths: the write path fails at the LOCK artifact (measured:
`open board lock ...: permission denied`) and never reaches the engine, while
the read path takes no lock and fails at the engine. Both are asserted
(`TestBacklogWrite_UnwritableDirFailsLoudly`,
`TestBacklogRead_UnwritableDirSurfacesEngineFailure`).

### Measurements (this run, this tree)

| Measurement | Command | Observed |
|---|---|---|
| Binary size baseline (pre-driver) | `go build -o /tmp/t306-baseline-moai ./cmd/moai` | 77,719,874 B |
| Binary size after driver (C-4) | `go build -o /tmp/t306-m1-moai ./cmd/moai` | 83,631,442 B — delta +5,911,568 B = **+5.64 MiB**, within the +12 MiB budget |
| Driver pin | `go get modernc.org/sqlite` | `modernc.org/sqlite v1.57.0` (+ libc v1.74.4, mathutil v1.7.1, memory v1.11.0, go-strftime v1.0.0, bigfft) |
| Concurrency (AC-TOSQ-009) | `go test -race ./internal/kanban -run TestConcurrencyStress` | PASS — 8 writers × 6 adds = 48 distinct ids, 48 stored items, last_seq 48, 0 collisions, 0 lost updates |
| kanban package | `go test ./internal/kanban/ -count=1` | `ok ... 12.363s` |
| cli packages | `go test ./internal/cli/... -count=1` | exit 0 — `ok internal/cli 248.615s` + 14 sub-packages ok |
| web / statusline / hook | `go test ./internal/web/... ./internal/statusline/... ./internal/hook/... -count=1` | all ok |
| Cross-compile vet | `GOOS={linux,windows,darwin} GOARCH=amd64 go vet ./internal/kanban/...` | exit 0 on all three (compile evidence only — the windows BEHAVIORAL verdict is CI's, per D.3) |

## §E.3 Run-phase Audit-Ready Signal

<pending run-phase>

## §E.4 Sync-phase Audit-Ready Signal

<pending sync-phase>
