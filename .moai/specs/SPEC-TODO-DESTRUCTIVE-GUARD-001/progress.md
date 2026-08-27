# Progress — SPEC-TODO-DESTRUCTIVE-GUARD-001

Card: t330 · Tier M · Route A (Tier S/M default)

## §E.1 Plan-phase Audit-Ready Signal

| Item | Value |
|---|---|
| SPEC ID | SPEC-TODO-DESTRUCTIVE-GUARD-001 (regex self-check: `PASS`) |
| Tier | M — 10 files, est. 400-700 LOC (`plan.md` §A) |
| Artifacts | spec.md + plan.md + acceptance.md (Tier M set) + this file |
| Requirements | 16 (at ceiling 16) |
| Acceptance criteria | 16 (at ceiling 16) |
| Base tree | `812ee01fc`, branch `WT-todo-destructive-guard` |
| Status | `draft` |

Decisions settled at plan-phase, each from a cited measurement:

- **Decision 1** — additive archive, not a fourth `BacklogState`. `spec.md` §B.1 (M1/M2/M3); `plan.md` §B. Confirmed independently by plan-audit iter1; not reopened.
- **Decision 2** — t330 owns the reversal and the refusal seam; t331 owns the persisted landing-state field. `spec.md` §B.2; `plan.md` §C.
- **Decision 3** — the archive is deliberately **included** in `export-json` (the downgrade route), with a stderr disclosure of the downgrade loss; live-queue readers exclude it. `spec.md` §C.5.

Measurements recorded at plan-phase that contradict the framing the card was dispatched with:

1. The existing landed primitive fails in **both** reachable modes, and correcting the ref moves the failure rather than removing it. As shipped (`LandedRef = "origin/main"`, `prlink_landed.go:28`) it answers **false** for every develop-integrated card — `origin/main` names t306 in 0 commits — so a default-on refusal would block everything. After the obvious ref correction it answers **true** on any mention — `origin/develop` names t306 in 13 commits, the earliest being the run commit `3cb258d62` — so it would have passed the premature `done` silently (`spec.md` §A.4). The opt-in ruling rests on the predicate answering the wrong question, not on the ref.
2. `LandedRef` is stale under the develop-integration git-flow. Not fixed here — shared with `moai todo pr`; declared out of scope in `spec.md` §D and left as a candidate follow-up card.
3. There is no live JSON engine. `openEngine` (`backlog_store.go:437-455`) falls through to the SQLite engine on every path, so a "both live backends" comparison is unreachable (`plan.md` §D).
4. The `moai todo` surface is **15** verbs, not 14 — the doctrine table omits `why` (`todo.go:137-141`).
5. `--expect` is carried by `next`, `edit`, `drop`, `undrop` — **not** by `move`.

### plan-audit history

| Iteration | Verdict | Score | Disposition |
|---|---|---|---|
| iter1 | FAIL | 0.75 (Tier M threshold 0.80) | 5 blocking defects (D1-D5) + 3 non-blocking (D6-D8); all addressed — see below |
| iter2 | PASS-WITH-DEBT | 0.9375 (monotonic +0.1875, no dimension regressed) | Clarity 0.75→1.00, Testability 0.50→0.75, Traceability 0.75→1.00; 7/7 MUST-PASS. Decision 3 reviewed and approved. 3 debts (S1, N1, N2) + 2 optional (N3, N4) — all 5 landed in iter3 |

iter3 debt closure: S1 (the two `acceptance.md` citations the D7 sweep missed, plus the false "four refreshed" claim corrected above) · N1 (AC-TDG-015 captures stdout/stderr separately — a merged `2>&1` would pass a disclosure printed to the machine-read stdout line) · N2 (AC-TDG-007 asserts the exact output `t1: no findings`, since `todo_why.go:34-35` echoes the id and defeats a grep) · N3 (`move` declares four flags, not two) · N4 (softened "the only point we control" — the exported artifact is also ours and could carry the warning; not built, not foreclosed).

D1 (spec.md §A.4 rewritten with both modes) · D2/D4 (Decision 3: REQ-TDG-015, AC-TDG-015, M5 rewritten to budget the disclosure) · D3 (REQ-TDG-006, AC-TDG-006, §A.3 rewritten to the reachable configuration) · D5 (15-verb list re-derived from `AddCommand`) · D6 (`--expect` set corrected — and `move` removed, which the defect report itself carried wrongly) · D7 (citations refreshed in spec.md and plan.md; t306 count 10→13, my original was `head -10` truncation) · D8 (REQ-TDG-016, AC-TDG-016: restore empties the entry).

> Correction (iter3): the iter2 line above originally read "four citations refreshed", which was an unobserved completion claim — two of the four (`todo.go:341` and `todo.go:352-354`) survived untouched in `acceptance.md` because the D7 sweep covered `spec.md` and `plan.md` only. Caught by iter2 audit as S1 and landed in iter3. The failure is the one D7 itself named, and it survived because the report read finished.

## §E.2 Run-phase Evidence

Cycle: TDD (RED-GREEN-REFACTOR). Every criterion below was verified in an isolated repository per `acceptance.md` §A.1 — either a Go fixture over `t.TempDir()` + `CLAUDE_PROJECT_DIR`, or a `mktemp -d` + `git init` shell repository. **No criterion touched this repository's live queue.**

Verbatim logs: `.moai/reports/t330/evidence/` (the tracked copy — the `.moai/state/verify/t330/` originals are gitignored and would not resolve at audit time once this worktree is disposed).

### RED, before any implementation

`.moai/reports/t330/evidence/red-cli.txt` — the CLI acceptance suite run at `abd4fbbbd` with no implementation present. 10 of 16 tests fail; the remaining 6 are the freeze and base-holds assertions `acceptance.md` marks as such. Three passed at base for the WRONG reason and were tightened before implementation rather than after:

| Test | Base-tree pass reason | What was tightened |
|---|---|---|
| `TestTodoUndone_ReissuedIDRefuses` | the parent command's mistyped-verb guard emits an error carrying `t1`, so an id-only assertion matched | now requires the refusal to name the reissue, not merely echo the id; re-measured RED at base before implementing |
| `TestTodoDoneUndone_RefusalsWriteNothing` | every path refuses at base (unknown flag / unknown verb) and so writes nothing | disclosed, not rewritten: the criterion asserts byte-identity, and `acceptance.md` AC-TDG-011 states the absent-id path already holds at base. Its post-implementation value comes from the paired positive tests (AC-TDG-008/009), which are genuinely red at base |
| `TestTodoDoneUndone_NeverPrompt` | `acceptance.md` AC-TDG-012 states this holds for `done` at base | disclosed; the criterion extends the guarantee to the new verb and flags |

The base verb surface was read from the live tree rather than transcribed — the RED log records the CLI's own listing: `add, analyze, done, drop, edit, export-json, list, move, next, pr, relate, undrop, unpick, unrelate, why` (15 verbs, no `undone`), matching `internal/cli/todo.go` `AddCommand`.

### AC matrix — 16/16 PASS

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-TDG-001 | PASS | `go test ./internal/cli/ -run TestTodoUndone_RestoresTheCard` | ok — `undone t1 alpha work`, `list` names t1 again. RED at base: `undone` unregistered |
| AC-TDG-002 | PASS | `-run TestTodoDone_UndoneRoundTripIsByteIdentical` + `./internal/kanban/ -run TestBacklogArchive_RestorePreservesPositions` | ok — serialization byte-identical across done+undone with a recorded finding; the kanban half archives a MIDDLE card and asserts `reflect.DeepEqual(before, after)` |
| AC-TDG-003 | PASS | `-run TestTodoDone_ArchivesRatherThanDiscards`; `sqlite3 backlog.db "select id, position from archived_items"` | archive ids `[t1]` in `export-json`; the database query returned `t1\|0` (`evidence/ac005.txt`) |
| AC-TDG-004 | PASS | `grep -n "CHECK (state IN" internal/kanban/backlog_sqlite.go`; `-run 'TestBacklogArchive_StateEnumUnchanged\|TestBacklogArchive_PerItemContractFrozen'` | ONE occurrence, line 113, reading `CHECK (state IN ('queued','picked','dropped'))`; enum still 3 values; `reflect.TypeOf(BacklogItem{}).NumField() == 5`. `archived_items` deliberately carries no state CHECK, so the grep stays single-hit |
| AC-TDG-005 | PASS | `-run TestBacklogArchive_SchemaVersionNotBumped`; `bash evidence/ac005.sh` | stamp reads `"1"` on a database holding an archived row; a binary built from `812ee01fc` ran `todo list` on that database with **rc=0**, saw the live t2, did NOT see the archived t1, and could still `add` (rc=0) |
| AC-TDG-006 | PASS | `-run TestTodoUndone_SurvivesMigrationFromLegacyJSON` + `./internal/kanban/ -run TestBacklogArchive_SurvivesLegacyMigration` | a JSON-only queue migrates on `done`, `backlog.db` appears, the round trip is byte-identical; a legacy file already CARRYING archived rows migrates them losslessly (the parity assertion was extended to cover the archive) |
| AC-TDG-007 | PASS | `-run TestTodoDone_ArchivedRowsInvisibleToLiveReaders` | `list`, `next`, `analyze` do not name t1; `why t1` emits exactly `t1: no findings`; re-adding the archived text is not reported as a duplicate |
| AC-TDG-008 | PASS | `-run TestTodoDone_ExpectRefusesMismatch` | non-zero exit, stderr names the observed prefix `alpha work`, record byte-identical |
| AC-TDG-009 | PASS | `-run 'TestTodoDone_RequireLandedRefusesWhenNotLanded\|TestTodoDone_RequireLandedProceedsWhenInconclusive'` | refusal names `origin/main` and leaves the record byte-identical; an unanswerable query proceeds (`done t1` on stdout) and surfaces `could not answer` on stderr |
| AC-TDG-010 | PASS | `-run TestTodoDone_NoLandingQueryWithoutTheFlag` | 0 subprocesses without the flag; exactly 1, to `git`, with it |
| AC-TDG-011 | PASS | `-run TestTodoDoneUndone_RefusalsWriteNothing` | all five refusal paths (absent id, `--expect` mismatch, `--require-landed`, `undone` of an unarchived id, `undone` into a reissued id) exit non-zero with the record byte-identical |
| AC-TDG-012 | PASS | `-run TestTodoDoneUndone_NeverPrompt` | six paths with stdin closed; every one returned a determinate exit, logged verbatim in the test output |
| AC-TDG-013 | PASS | `-run TestTodoUndone_ReissuedIDRefuses` + `./internal/kanban/ -run 'TestBacklogArchive_RestoreRefusesAReissuedID\|TestBacklogArchive_LastSeqClearsArchivedIDs'` | refusal names the id AND the reissue, record byte-identical, live card untouched. The allocator provably cannot reissue: the high-water mark now clears archived ids, asserted directly |
| AC-TDG-014 | PASS | `cmp` on the two `todo.md` paths; `grep -c undone` on each | `cmp` rc=0; both files carry 4 occurrences of `undone` |
| AC-TDG-015 | PASS | `-run 'TestTodoExportJSON_CarriesArchiveAndDisclosesDowngrade\|TestTodoExportJSON_NoDisclosureWithoutArchivedRows'` | streams captured separately: the archive carries t1 and `items` carries t2; stdout is the structured `exported …` line and contains no disclosure; stderr names `archived` and `discard`. An archive-free export discloses nothing |
| AC-TDG-016 | PASS | `-run TestTodoUndone_EmptiesTheArchiveEntry` + `./internal/kanban/ -run TestBacklogArchive_RestoreIsNotRepeatable` | archive empty after restore; a second `undone` refuses with the record byte-identical; re-archiving works, double-archiving refuses |

### `LandedRef` pre-flight (plan.md §C, required by acceptance.md §C)

Performed and recorded. `internal/kanban/prlink_landed.go:28` reads `const LandedRef = "origin/main"` — unchanged, as shipped. It was **not** corrected: the constant is shared with `moai todo pr`, so changing it alters that verb's answers, and `spec.md` §A.4 already shows the correction would not unlock a default-on check. `--require-landed` therefore ships opt-in, and both its help text and the doctrine name the ref its answer is about.

### Deviations from plan.md, each forced by a requirement

1. **Archive entry shape.** `plan.md` §E M1 sketches `Archived []BacklogItem` plus `ArchivedFindings []BacklogFinding`. Implemented instead as ONE field, `Archived []BacklogArchiveEntry`, each entry carrying the card, its findings, and the POSITION each held. Forced by REQ-TDG-002: `done` splices a row out of the middle, so a positionless restore appends and the record comes back with the same cards in a different order — equivalent-looking, not byte-identical. Still additive top-level fields plus new tables per REQ-TDG-004; the five-field per-item contract and the three-value enum are untouched.
2. **AC-TDG-010's `git` shim.** The criterion describes a shim on `PATH`. Implemented against the package's own process seam (`todoRunCommand`) instead — the same measurement, in-process. This is not merely equivalent: it corrected a wrong-reason pass. The landing query runs `git` in the PROCESS working directory, not the queue root, so a ref seeded in the fixture repo was ignored and AC-TDG-009 was passing off whatever this repository's own `origin/main` happens to contain. Both landing criteria now pin the answer.
3. **Two pre-existing freeze tests updated.** `TestBacklogEngineSchemaShape` (exact table set) and `TestTodoVerbSurfaceZeroDelta` (verb x flag surface) both fail on any addition — by design. Both were extended to DECLARE this change's additions rather than edited to absorb them: the verb-surface table stays pinned at its branch point with the additions listed beside it, so the file still records which change introduced what. Both remain exact.

### Not attributable to this change

`go test ./internal/cli/` initially reported 9 doctor-test failures. Cause measured, not assumed: `✗ Error: no readable binary to judge at <worktree>/bin/moai (11 committed artifacts to compare)` — the Agent Emit Embed check needs a built binary. After `make build`, `go test ./internal/cli/ -run 'TestRunDoctor|TestDoctorCmd'` returned **rc=0** (`evidence/doctor-after-build.txt`). The pre-build log is retained at `evidence/cli-full-before-binary-build.txt`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha:
  m1: 05d2c234b   # M1 archive storage shape
  m2: 3a0ce021c   # M2-M5 undone, done's guards, export disclosure
  m3: 3f1a7b896   # M6 doctrine + template mirror
  m_final: b08c7a6e4   # the progress.md commit, backfilled (a commit cannot name its own SHA)
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
base_tree: abd4fbbbd
branch: WT-todo-destructive-guard
preserve_list_post_run_count: 0    # no file outside plan.md §A touched
l44_pre_commit_fetch: not-run      # see Gaps
l44_post_push_fetch: not-run       # nothing pushed; integration is the lead's
new_warnings_or_lints_introduced: 0
lint: "golangci-lint run --timeout=5m ./internal/kanban/... ./internal/cli/... -> 0 issues"
vet: "go vet ./... -> rc=0"
cross_platform_build:
  darwin_arm64: "go build ./... -> rc=0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... -> rc=0"
  windows_tests: not-compiled      # gap, see below
tests:
  kanban: "go test ./internal/kanban/ -count=1 -cover -> ok, coverage 87.0%"
  cli: "go test ./internal/cli/ -count=1 -cover -> ok, coverage 79.8%"
total_run_phase_files: 13
m1_to_mN_commit_strategy: "three commits, one per milestone group (M1 / M2-M5 / M6), each naming card t330; no force-push, no --amend"
```

### Files touched (13)

Implementation (6): `internal/kanban/backlog_store.go`, `internal/kanban/backlog_sqlite.go`, `internal/kanban/backlog_migrate.go`, `internal/cli/todo.go`, `internal/cli/todo_export.go`, and the new `internal/cli/todo_undone.go`.
Tests (4): new `internal/kanban/backlog_archive_test.go` and `internal/cli/todo_undone_test.go`; updated `internal/kanban/backlog_sqlite_test.go` and `internal/cli/todo_surface_test.go`.
Doctrine (2): both `todo.md` paths. SPEC (1): `spec.md` frontmatter `draft → in-progress` on the M1 commit.

`plan.md` §A anticipated 10 files; the three extra are `todo_undone.go` (folded into the CLI row there) and the two pre-existing freeze tests this change's additions oblige to declare them.

### Gaps — what was explicitly NOT observed

- **Full-suite verdict.** Only `./internal/kanban/...` and `./internal/cli/...` were run, per `plan.md` §F.5 and CLAUDE.local.md §4 (concurrent lanes running the full suite drove machine load to 413). The whole-module verdict is CI's, on a clean environment against the pushed head. **Nothing was pushed**, so no CI run exists for this work yet.
- **`internal/cli` package coverage (79.8%)** is below the 85% package target. It was NOT measured at the base tree, so no claim is made in either direction about whether this change moved it. What WAS measured is the new code: `newTodoDoneCmd` 100%, `todoRequireLanded` 100%, `newTodoUndoneCmd` 100%, `discloseArchiveDowngradeCost` 100%, `todoWriteLine` 66.7% (the uncovered arm is a stream-write failure), `RestoreCard` 100%, `ArchivedIndex` 100%, `ArchiveCard` 93.8%, `readArchive` 76.9%, `writeArchive` 66.7% (uncovered arms are SQL-failure paths).
- **No `-race` run.** The archive rides the existing `Mutate` lock and adds no goroutine, so no new concurrency surface was introduced — but that is an argument, not a measurement.
- **Windows tests not compiled or run.** Only `GOOS=windows go build` was verified; a green cross-build does not establish that the tests compile or pass there.
- **The downgrade check used `812ee01fc`**, the base `acceptance.md` names — not an actual released v3.0.x / v3.1.x binary. A release older than that base was not tested.
- **`l44` fetch not run.** This session worked inside an isolated worktree, which the Pre-Edit Sync Check exempts, and pushed nothing. A divergence check belongs to whoever integrates the branch.
- **`moai spec lint` / `moai spec audit` not run** on the SPEC artifacts after the frontmatter transition.

### Residual risk — what could still be wrong despite the above

- **Restore position is exact only for the immediate round trip.** Once other cards have come and gone, the recorded index is clamped into range (asserted by `TestBacklogArchive_RestoreClampsWhenTheQueueMovedOn`). REQ-TDG-002 asks for byte-identity against the state immediately before the `done`, which is what is delivered; a restore taken later lands at a defensible position, not a provably original one.
- **The archive grows without bound.** Declared out of scope (`spec.md` §D); a long-lived queue's `backlog.db` and every `export-json` artifact grow monotonically with the number of finished cards.
- **The landed predicate still answers the wrong question.** Documented in the help text and the doctrine, and shipped opt-in — but an operator who reads `--require-landed` as "the work shipped" will be misled exactly as the t306 incident was. Closing that needs t331's persisted landing-state field.
- **Downgrade loss is narrower than feared, and was measured.** The old binary writes only the tables it knows, so `archived_*` SURVIVED its `add` (`select id from archived_items` returned `t1` afterwards). The disclosed loss is therefore the JSON path specifically: an export read and rewritten by a pre-archive binary drops the field. A user who exports, downgrades, and later re-upgrades keeps the archive as long as `backlog.db` still exists.
- **`export-json` now writes an `archived` key** that a third-party consumer of the legacy format may not expect. No such consumer is known, and the field is additive, but the artifact's shape did change.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
