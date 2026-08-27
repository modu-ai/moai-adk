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
| M3 directory rename (t309) + fallback read | complete | `8910c337c` |
| M4 consumer sweep | complete | `8910c337c` |
| M5 export-json + downgrade docs | complete | `ffe33ac09` |
| M6 cross-platform / race / gates | complete | `d19187327` |

M3 and M4 landed in one commit: the sweep is what makes the rename true, and a
commit that renamed the directory while consumers still named the old one would
be a broken tree by construction. The template edits ride M4 (Template-First
plus `make build`) rather than M5, because the three skill docs name the
directory the rename moved.

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

### M3-M6 safety mechanisms fired MECHANICALLY

Continuing the table above; same discipline, same restore-after.

| # | Mechanism | Mutation applied | Observed on mutant |
|---|---|---|---|
| 9 | stale-copy policy (REQ-TOSQ-015) | the `dirExists(current)` early return removed | `TestStateDirBothPresentLeavesLegacyUntouched` FAIL — "last_seq = 42, want 99 — the current directory must win" |
| 10 | fallback read on refused relocation (AC-TOSQ-008) | refusal returns the new dir instead of the legacy one | `TestStateDirRelocationRefusedFallsBackToLegacyRead` FAIL — "resolved to .../state/todo, want the legacy directory served in place" |
| 11 | pure resolver never relocates (REQ-TOSQ-015) | the `!adopt` early return removed | `TestStateDirPureResolutionNeverRelocates` FAIL — "want the legacy directory observed in place" |
| 12 | export survives later verbs (REQ-TOSQ-013/016) | state D quarantines unconditionally, ignoring the in-flight marker | `TestTodoExportJSON_SurvivesSubsequentVerbs` FAIL — "the export was renamed or removed by a later verb"; `TestMigrationBothPresentPrefersDatabase/a_downgrade_export_is_left_exactly_where_it_was_put` FAIL |

Mutation-run logs: `.moai/state/verify/t306-m3/red.log`, `.moai/state/verify/t306-m5/red.log`.

### A second vacuous guard, found the same way

The M5 state-D subtest asserting "an export is left alone" passed under mutant
12 on its first form. It reused a root that had already migrated, so a
`backlog.json.migrated` existed — and the SEPARATE "never overwrite an existing
quarantine" guard silently did the work. The subtest asserted a mechanism it
was not exercising.

Rebuilt on a fresh root where no quarantine exists, it fires. Two instances of
this shape in one SPEC (the other is recorded above, at M2) is the reason every
row in both tables cites the mutant that fired it: a guard that has never been
seen red has proven nothing, whatever its green says.

### A defect the export surfaced

Writing `export-json` exposed a real fault in the M2 state-D rule, not a test
problem. A database beside a `backlog.json` is ambiguous — pre-cutover legacy
stranded by a crash between the migration commit and the quarantine, or a
downgrade export the operator just asked for. M2 quarantined unconditionally,
so `moai todo export-json` followed by ANY `todo` verb renamed the operator's
only downgrade artifact to `.migrated`. They would not have found out until the
older binary came up to an empty queue.

Closed by recording an in-flight marker (`meta.legacy_quarantine_pending`) as
the migration's last outstanding step, cleared once the rename lands. Only a
set marker means "finish this quarantine"; an export never sets one. The
state-D test now asserts BOTH directions, and reconstructs the crash case
faithfully — legacy file restored AND the marker set, which is what a real
crash leaves. Asserting either direction alone passes a store that ignores the
marker entirely.

### AC matrix (acceptance.md §D)

Every row's command was run in this tree, this run. Evidence logs under
`.moai/state/verify/t306-m*/`.

| AC | Status | Command | Observed |
|----|--------|---------|----------|
| AC-TOSQ-001 | PASS | `go test ./internal/kanban -run TestMigrationParity` | PASS with named subtests: item count, per-item field equality, spec_id null shape, insertion order preserved, findings order and tuples, last_seq above max-present survives, physical schema present |
| AC-TOSQ-002 | PASS | `go test ./internal/kanban -run TestMigrationIDContinuity` | PASS — first id after the cutover is `t43` (persisted mark 42 + 1), second `t44` |
| AC-TOSQ-003 | PASS | `go test ./internal/kanban -run TestMigrationQuarantinesLegacyByteIdentical` | PASS — `backlog.json` absent, `backlog.json.migrated` sha256-equal to the seed |
| AC-TOSQ-004 | PASS | `go test ./internal/kanban -run TestMigrationBothPresentPrefersDatabase` | PASS, both subtests — interrupted migration completes its quarantine; a downgrade export is left exactly where it was put |
| AC-TOSQ-005 | PASS | `go test ./internal/kanban -run 'TestMigrationMalformedAbortsWithoutDestroying\|TestMigrationPartialFailureRemovesArtifacts'` | PASS — malformed seed: structured error, seed sha256 unchanged, no db artifact; duplicate-id seed (the case that actually reaches the cleanup path): `IsBacklogIDConflict`, no `.db`/`-wal`/`-shm` residue, seed unchanged, no quarantine |
| AC-TOSQ-006 | PASS | `go test ./internal/kanban -run TestStateDirRelocatesLegacyWithRegistryFiles` | PASS — census: 3 registry files + `backlog.json` all under `.moai/state/todo/`; legacy dir gone; sampled record byte-unchanged; `RecordPath` follows |
| AC-TOSQ-007 | PASS | `go test ./internal/kanban -run TestStateDirBothPresentLeavesLegacyUntouched` | PASS — todo dir wins on both resolvers and through a real read + write; legacy census and queue bytes unchanged |
| AC-TOSQ-008 | PASS | `go test ./internal/kanban -run TestStateDirRelocationRefusedFallsBackToLegacyRead` | PASS — refusal FIRED by revoking the state parent's write bit; no error surfaces, queue served from the old layout, no partial new directory |
| AC-TOSQ-009 | PASS | `go test -race ./internal/kanban -run TestConcurrencyStress` | PASS — 8 writers x 6 adds = 48 distinct ids, 48 stored items, last_seq 48, 0 collisions, 0 lost updates; race detector clean |
| AC-TOSQ-010 | PASS | (a) `go test ./internal/cli/... ./internal/web/... -count=1`  (b) `go test ./internal/cli -run 'TestTodoVerbSurfaceZeroDelta\|TestTodoVerbExitCodesUnchanged'` | (a) exit 0 — every behavioral guard green UNMODIFIED, including the console's own `TestTodoSectionReadsThroughToProjectLocalQueue` / `TestConsoleRoutesLeaveBacklogUntouched` and the bare-join convention walk. (b) PASS — the frozen 14-verb x flag table shows zero deltas, `export-json` is the single permitted addition, and 13 representative invocations return their pre-swap success/refusal verdict. **This half was MISSING from the first version of this matrix — see the correction below.** |
| AC-TOSQ-011 | PASS | `go test ./internal/cli -run TestTodoExportJSON` | PASS ×4 — round trip through the legacy record shape with post-cutover state changes present; the live store stays authoritative; no temp residue; the export survives later verbs |
| AC-TOSQ-012 | PASS | `go test ./internal/statusline -run TestResolveBacklogCounts_LatencyBudget` | 500-item fixture, 41 renders — median 433.459us, p95 604.541us, max 716.833us. Budget: median <=10ms, ceiling 25ms |
| AC-TOSQ-013 | PASS (compile evidence only) | `GOOS={windows,linux,darwin} GOARCH=amd64 go vet ./internal/...` + `GOOS=darwin GOARCH=arm64` | exit 0 on all four. The windows BEHAVIORAL verdict is CI's, per acceptance.md D.3 — see Gaps |
| AC-TOSQ-014 | PASS | `go test ./internal/kanban -cover` + per-file statement count from the profile | package `internal/kanban` 87.6% (was 82.0% before the M6 unit tests); changed-path files 506/582 = 86.9%. Both clear the 85% bar |
| AC-TOSQ-015 | PASS | dual-spelling grep over `internal/ pkg/ cmd/` excluding `_test.go` and `kanban-board` | Zero `"state", "kanban"` segment joins. Two `state/kanban` occurrences remain, both comments and both on the allowlist: `state_dir.go:7` (the fallback reader's own justification, naming this SPEC) and `board_test.go:128` (naming the board dir that deliberately did NOT rename). Template tree: zero |
| AC-TOSQ-016 | PASS | `make build` then `strings bin/moai \| grep -c` | exit 0; embedded bundle carries `state/todo/backlog.json` 4x and `state/kanban/backlog.json` 0x |
| AC-TOSQ-017 | PASS | `go test ./internal/kanban -run 'TestMigrationParity/physical_schema_present\|TestReorderedItemsRoundTrip'` | PASS — meta rows (`schema_version`, `last_seq`) present; a `todo move`-shaped reorder round-trips as `t3, t1, t2`, so array position is the stored order |
| AC-TOSQ-018 | PASS | `go test ./internal/kanban -run TestBacklogEnginePragmas` | PASS — `journal_mode` reads `wal`, `busy_timeout` >= 5000, asserted on the pool AND across connection churn |

### CORRECTION — AC-TOSQ-010 was first reported PASS on half its criterion

The first version of this matrix marked AC-TOSQ-010 PASS citing only "the
existing suites are green". That is one of the two halves acceptance.md asks
for. The AC's green-evidence cell reads:

> `go test ./internal/cli -run 'TestTodo' ./internal/web ...` PASS **+
> surface-diff test asserts zero-delta**

and its scenario names "a frozen verb-surface comparison table (verb x flags x
exit codes) generated pre/post swap shows zero deltas". No such test existed.
Marking the AC PASS on the strength of the surface LOOKING unchanged was an
unobserved claim — precisely the failure this SPEC's own doctrine names, and
the same shape as the two vacuous guards recorded above, this time in my
report rather than in a test.

The gap was raised by the lead. It is now closed by
`internal/cli/todo_surface_test.go`.

**The frozen table's baseline is measured, not remembered.** Its provenance is
a diff of the surface DECLARATIONS between the branch point and HEAD:

```
git diff 7ed6edb3e..HEAD -- 'internal/cli/todo*.go' | grep -E '^[+-]' \
  | grep -vE '^(\+\+\+|---)' \
  | grep -E 'Use:|Flags\(\)\.|Args:|cobra\.(NoArgs|ExactArgs|...)'
```

which returns exactly two lines, both introducing `export-json`:

```
+		Use:   "export-json",
+		Args: cobra.NoArgs,
```

Every other verb, flag, and arity declaration on this branch is byte-unchanged.
That is the zero-delta proof; the test then freezes it going forward, reading
the LIVE cobra tree rather than source text so it also catches what a source
grep would miss — a flag registered elsewhere, a shorthand added, a default
changed.

Fired by mutation (logs: `.moai/state/verify/t306-ac010/red.log`):

| Mutation | Observed |
|---|---|
| verb renamed (`list` -> `ls`) | FAIL — "verb \"list\" is GONE from the surface"; "verb \"ls\" appeared and is not this SPEC's declared addition" |
| flag dropped (`add --force`) | FAIL — "verb \"add <text>\" re-flagged: frozen [force=bool(false) pick=bool(false)] / live [pick=bool(false)]" |
| shorthand added (`relate --note` -> `-n`) | FAIL — "re-flagged: frozen [note=string() ...] / live [note=string()/-n ...]" |
| `export-json` unregistered (vacuity check) | FAIL — "export-json is not registered; the downgrade route is unreachable"; "surface holds 14 verbs, want 15" |

The fourth is the one that matters for this matrix's own honesty: without it
the table would pass on a tree where the addition had never been wired in.


### M6 additions

- **Lock-held relocation** (plan.md M6, design.md R6). POSIX and Windows
  disagree: renaming a directory holding an open, locked file is permitted on
  one and refused on the other. `TestStateDirRelocationUnderHeldLock` therefore
  asserts the platform-NEUTRAL invariants — the queue stays readable and
  complete, exactly ONE directory holds it (no half-moved debris), the session
  records travel with it, and a retry after the lock releases lands on the
  current directory. Observed on darwin/arm64: `relocated=true`. A test written
  to assert either platform's specific outcome would be a regression on the
  other.
- **Coverage** (C-1). The M6 unit tests are not padding: the findings helper
  family (`Names`, `SamePairAs`, `HasFindingTuple`, `AppendFindingOnce`,
  `FindingsNaming`, `RemoveFindingsNaming`, `HasAgentFindingForPair`) and the
  count surface were reachable only from OTHER packages' tests, so their rules
  were asserted through a command surface — which means a change to a rule plus
  a compensating change to its command could pass together. They now have
  direct tests stating the rules themselves (unordered pair matching,
  tuple-once excluding timestamp and note, the 1-based index `todo unrelate`
  addresses).
- **Lint** (E5). `golangci-lint run` over `internal/{kanban,statusline,web,cli}`
  — 0 issues. Two were found and fixed rather than suppressed: an unused
  helper I had written and never called, and an unchecked `Fprintf` on the
  command's own stream.
- **Secured** (D.6). Grep for SQL built by concatenation or `Sprintf` over
  `internal/kanban/*.go`: clean. Every value travels through a placeholder.
- **Subagent boundary**. `AskUserQuestion` / `mcp__askuser` in the touched
  production code: 0 matches.


## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-27
run_status: complete
ac_pass_count: 18
ac_fail_count: 0
milestones_complete: M1, M2, M3, M4, M5, M6
run_commit_shas: 3d24cf6df (M1), 83a1d492a (M2), 447f517fe (evidence), 8910c337c (M3+M4), ffe33ac09 (M5), d19187327 (M6)
coverage_internal_kanban: 87.6%          # go test -cover ./internal/kanban/...
coverage_changed_path_files: 86.9% (506/582 statements, from the profile)
coverage_internal_statusline: 91.3%      # go test -cover ./internal/statusline/...
coverage_internal_cli_affected_paths: 94.5% (358/379 statements — kanban.go 98.2%, todo.go 94.3%, todo_export.go 77.1%; C-1 floor for affected command paths is 90%)
coverage_internal_cli_package: 79.5%     # pre-existing baseline for the whole package, not a changed-path figure
coverage_internal_web_package: 66.8%     # pre-existing baseline; this branch touched 3 files in it
race_detector: clean (go test -race ./internal/kanban/...)
lint: golangci-lint run ./internal/{kanban,statusline,web,cli}/... — 0 issues
cross_platform_vet: windows/amd64, linux/amd64, darwin/amd64, darwin/arm64 — all exit 0
affected_suite: 33 packages ok, exit 0
binary_size_delta: +5,911,568 B (+5.64 MiB) vs the 77,719,874 B pre-driver baseline; C-4 budget +12 MiB
new_warnings_or_lints_introduced: none
preserve_list_post_run: no file outside the SPEC's scope envelope was modified
push_state: NOT pushed, no PR — integration is the lane orchestrator's act

### Gaps — what was NOT observed

Stated explicitly, because an empty Gaps section would itself be a claim.

- **Windows behavioral verdict.** Local evidence for windows/amd64 stops at
  `go vet` compile success. No windows binary was executed, so the lock-held
  rename divergence (design.md R6), `LockFileEx` sharing-violation behavior,
  and the windows lock-release path are UNVERIFIED behaviorally here. This is
  the indirect verification acceptance.md D.3 accepts in advance; the verdict
  belongs to the CI windows job on the push, which has not happened.
- **Ten-lane process-level contention.** AC-TOSQ-009 was measured in-process
  (8 goroutine writers under `-race`). No multi-PROCESS stress run was
  executed, so real factory-fleet contention across separate `moai` processes
  is approximated, not observed. acceptance.md D.3 accepts this approximation;
  fleet-scale validation remains dogfood observation.
- **SC-3 live rehearsal transcript.** The seed to migrate to mutate to export
  chain is proven by `TestTodoExportJSON_RoundTripsThroughTheLegacyShape`
  driving the real CLI verbs. A hand-run transcript on a scratch root outside
  the test harness was NOT captured.
- **The real operator queue was never touched.** Every test used `t.TempDir()`.
  No command in this run was pointed at the live `.moai/state/kanban/` queue,
  so the production queue's own migration has NOT been exercised — it will
  happen on the first `moai todo` command after this lands, which is the
  behavior under test but not a thing this run observed.
- **CI full-suite verdict.** Only affected packages were run locally (load
  discipline). `go test ./...` was NOT run; that verdict is CI's.
- **`todo_export.go` sits at 77.1%.** The affected-path aggregate (94.5%) clears
  C-1's 90% floor, but this new file does not reach it alone. The residue is
  three I/O error branches inside `writeExportAtomic` — a failed `Write`, a
  failed `Close`, and a `Replace` that fails for a reason other than the target
  being a directory. Reaching them needs an injected-failure seam in production
  code, and adding a seam to move a number is the over-engineering this
  project's constitution forbids. The reachable failure paths ARE covered: an
  unwritable directory, an uncreatable parent, a target that is a directory,
  and the no-residue guarantee on each.
- **Per-file coverage below the bar.** Two changed-path files sit under 85%
  individually — `backlog_migrate.go` 80.2%, `todo_root.go` 79.6% — while the
  changed-path total (86.9%) and the package (87.6%) clear it. C-1 is worded
  "overall on changed-path files", so this reads as PASS; recording the
  per-file figures rather than only the aggregate, since the aggregate is what
  passes and the per-file numbers are what a reader would want to challenge.

### Residual risk

- The migration is one-way per queue root and fires unattended on the first
  `todo` verb. Its safety rests on the parity check running BEFORE authority
  flips — verified by mutation (row 7) — but a defect in the parity comparison
  itself would pass its own check. The comparison is exhaustive by field rather
  than by count, and all ten of its axes are fired directly
  (`TestMigrationParityCatchesTamperedRecord`), which is the strongest local
  guard available short of a second independent implementation.
- The state-D marker distinguishes an interrupted migration from an export. A
  database written by some future path that sets the marker and then never
  clears it would quarantine a legitimate `backlog.json` on the next open. Only
  the migration writes it today, and it is cleared in the same function.
- `-wal` and `-shm` siblings are new artifacts in a directory operators and
  scripts read. Backup tooling that copies `backlog.db` alone, without the WAL,
  captures a queue missing its most recent commits. The downgrade doc names all
  five artifacts for exactly this reason, but tooling outside this repository
  cannot be made to read it.


## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-08-27
sync_status: complete
sync_commit_sha: pending-backfill      # backfilled by the lane orchestrator in the
#   immediately following commit — a commit cannot carry its own sha.
card: t306 (absorbs t309)
branch: WT-todo-sqlite
spec_status_transition: in-progress -> completed (3-phase close: plan -> run -> sync;
#   the `completed` transition rides THIS sync commit, per the 3-phase close convention)
spec_version: 0.1.0 -> 0.1.1
changelog_entry_position: CHANGELOG.md `## [Unreleased]` / `### Added`, first entry
ac_pass_count: 18                      # AC-TOSQ-001..018
ac_fail_count: 0

### Landed run-phase commits

All reachable from `origin/develop`; the card's work was integrated ahead of this
sync commit, so the run-phase evidence describes a tree that is already merged.

| Commit | Carries |
|---|---|
| `3030df58b` | plan artifacts |
| `3d24cf6df` | M1 — SQLite driver adoption (`modernc.org/sqlite`, pure Go) + storage engine |
| `83a1d492a` | M2 — store guts + lazy migration state machine |
| `447f517fe` | M1+M2 evidence |
| `8910c337c` | M3 — state-directory rename (`.moai/state/kanban` -> `.moai/state/todo`, absorbing card t309) + M4 consumer sweep |
| `ffe33ac09` | M5 — `export-json` downgrade route + docs |
| `d19187327` | M6 — cross-platform, race and quality gates |
| `1c5840558` | SHA backfill |
| `a8099e43e` | rescue commit — verb-surface guard `todo_surface_test.go` + companions |
| `3cb258d62` | merge commit on `develop` |

### Post-merge verification

Measured by the lane orchestrator against the MERGED tree (`3cb258d62`), not
carried over from the run-phase tree. Cited here with that attribution rather
than re-executed: the machine is shared with eight concurrent lanes, and a
re-run under that load would measure the machine, not the code.

| Dimension | Command | Observed |
|---|---|---|
| kanban package | `go test ./internal/kanban/...` | ok 13.2s |
| statusline package | `go test ./internal/statusline/...` | ok 15.3s |
| web package | `go test ./internal/web/...` | ok 3.6s |
| hook package | `go test ./internal/hook/...` | ok 27.1s |
| cli package | `go test ./internal/cli/...` | ok 220.4s + 16 sub-packages all ok |
| Cross-platform vet | `GOOS=windows go vet` / `GOOS=linux go vet` | rc=0 on both |
| Coverage (C-1 floor 85%) | `go test -cover ./internal/kanban/...` | 87.7% of statements |
| Lint | `golangci-lint run ./internal/kanban/...` | 0 issues |
| Format | `gofmt -l` on card-attributable files | 0 unformatted |
| Live queue untouched | inspection of `.moai/state/kanban/` | no `.db` artifact; primary checkout clean |
| Binary provenance | `strings bin/moai \| grep -F 3cb258d62` | hit — built from the merged tree |

### Sync-phase checks run in THIS tree

| Check | Command | Observed |
|---|---|---|
| B12-1 duplicate-entry guard | `grep -c 'SPEC-TODO-SQLITE-001' CHANGELOG.md` | `0` — no prior entry, emission safe |
| B12-2 AC count | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `18` (AC-TOSQ-001..018), matching the CHANGELOG entry's stated count |
| B12-3 path verification | `ls .moai/docs/todo-queue-storage.md internal/template/templates/.moai/docs/todo-queue-storage.md` | both present, 113 lines, 5911 B |
| Template mirror parity | `diff -q` on the pair above | identical — the M5 mirror is intact, verified rather than assumed |
| REQ-TOSQ-018 | `grep -rn 'state/kanban' internal/template/templates/` | 1 hit, in `todo-queue-storage.md` prose naming the OLD directory as history ("Earlier releases called it ..."). Not a live path reference; REQ-TOSQ-018 holds |

### Documentation synchronized

- `CHANGELOG.md` — one `### Added` entry under `## [Unreleased]`.
- docs-site 4-locale (`ko` / `en` / `ja` / `zh`), `utility-commands/moai-todo.md` and
  `advanced/moai-web-console.md`: the queue path and storage form corrected to
  `.moai/state/todo/backlog.db`, and the additive `export-json` verb added to the
  CLI-surface table. These pages stated `.moai/state/kanban/backlog.json`, which
  the M3 rename and the M1/M2 storage swap made factually false — the correction
  is card-attributable, not a drive-by edit.
- `.moai/docs/todo-queue-storage.md` (REQ-TOSQ-017) landed at M5 and is unchanged
  here; its template mirror was verified intact rather than re-copied.

### Gaps — what this sync phase did NOT observe

Stated explicitly, because an empty Gaps section would itself be a claim.

- **No verification was re-executed in this tree.** Every figure in the
  post-merge table above is the lane orchestrator's measurement, cited with that
  attribution. This sync phase ran only the greps, `ls`, and `diff` recorded in
  the sync-phase table — no test, no build, no lint.
- **`sync_commit_sha` is a placeholder.** It is `pending-backfill` until the
  orchestrator's following commit; any reader resolving this SPEC's sync commit
  before that backfill lands will not find it here.
- **The docs-site build was not run.** The 4-locale edits are text corrections to
  existing tables and sentences; no Hugo build, link check, or locale-parity
  linter was executed against them in this tree.
- **The README verb enumeration was left alone.** `README*.md` line 364 lists ten
  `moai todo` verbs and omits `pr`, `relate`, `unrelate`, `why` — a gap that
  predates this card. `export-json` was NOT added there, because adding one verb
  to a list already missing four would leave it no more accurate and would put a
  non-card-attributable edit in this commit. Recorded as follow-up work.
- **Every run-phase Gap in §E.3 remains open.** The windows behavioral verdict,
  ten-lane process-level contention, the SC-3 live rehearsal transcript, the real
  operator queue's own migration, and the CI full-suite verdict were not observed
  here either; this phase added no evidence on any of them.
