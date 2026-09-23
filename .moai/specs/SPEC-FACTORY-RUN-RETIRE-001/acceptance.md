# acceptance.md — SPEC-FACTORY-RUN-RETIRE-001 (card t1107)

## §A Scope of Verification

Verification layer for factory run retirement. The requirements layer (GEARS) lives in `spec.md`
§B; this file carries only observable, binary-testable criteria in Given-When-Then form.

Every criterion below is satisfied by a **recorded command and its verbatim output** in
`progress.md` §E.2, attributed to that run and that tree. A source citation satisfies no criterion
in this file.

## §B Test-environment constraint (binds every AC)

[HARD] Any test or manual exercise that opens factory state runs with `HOME`, `MOAI_HOME`, and
`MOAI_CLAUDE_BIN` pointed at a sandbox, and with a project directory **outside this repository's
worktree set**. Mechanism: `homestate.CanonicalProjectRoot` converges a linked worktree onto the
primary checkout, so a fixture rooted inside a worktree writes to the developer's real
`factory.db`. "Use a temp dir" alone does not satisfy this — the temp dir must also not be inside a
linked worktree, and the environment must be scrubbed in the same compound invocation
(`unset … && <command>`).

Evidence that the isolation held: the only factory state created by the exercise is under the
sandbox `MOAI_HOME`, and its project key is not this repository's.

## §C Two-cell adoption — RED-now and green path

[HARD] Per `verification-completeness.md` §2, every release-blocking criterion below is adopted
with a pair of cells: a RED-now observation on the pre-implementation tree, and the milestone that
flips it. The RED-now cells live in the evidence ledger at §C.1 and are cited by id from the matrix.

**Document-level tree pin**: every RED-now cell below was measured on commit **`bb5b8f9d1`**
(branch `WT-factory-run-retire`). No criterion carries its own pin, so this document-level pin binds
all of them.

### C.1 RED-now evidence ledger

Each entry carries the four required elements: a single-invocation read-only command, its verbatim
stdout, its exit code, and the tree SHA (the document pin above). Every RED below is red because
**the symbol, column, or surface it names does not exist yet** — not because of unrelated files, and
not because the criterion is unsatisfiable.

| id | command | verbatim stdout | exit | why red |
|----|---------|-----------------|------|---------|
| R-01 | `grep -c lead_pid internal/homestate/factory.go` | `0` | 1 | the owner-identity column is not in the `runs` DDL |
| R-02 | `grep -c "factorySchemaVersion = 3" internal/homestate/factory.go` | *(empty)* | 1 | schema is still at version 2 |
| R-03 | `grep -rn migrateFactoryV2ToV3 internal/homestate` | *(empty)* | 1 | no v2→v3 migration exists, so no legacy row can be reconciled |
| R-04 | `grep -c retired internal/homestate/runtime.go` | `0` | 1 | no `retired` status value and no `run.retired` event are written anywhere |
| R-05 | `grep -rn ReconcileActiveRuns internal/homestate` | *(empty)* | 1 | the reconciler does not exist |
| R-06 | `grep -c "Use: \"runs\"" internal/cli/factory_handoff_recover.go` | `0` | 1 | the `moai factory runs` operator surface does not exist |
| R-07 | `grep -rln factory_run_retire test/integration/harness` | *(empty)* | 1 | no test sits on the one path the three-OS CI job runs |
| R-08 | `grep -c "AC-011 PASS" .moai/specs/SPEC-FACTORY-RUN-RETIRE-001/progress.md` | `0` | 1 | no door invocation has been recorded |

Deliberately **not** used as a RED cell: any `go test -run <NewTestName>` selector. A Go test binary
given a selector that matches zero tests exits 0 and prints `ok` — a vacuous green dressed as a red,
which is the exact failure `verification-completeness.md` §2 warns about.

## §C.2 AC Matrix

| AC | Requirement | Milestone | Severity | RED-now | Green path (what flips it) |
|----|-------------|-----------|----------|---------|----------------------------|
| AC-001 | REQ-001/002 | M1 | release-blocking | R-01, R-02 | M1 adds the columns and the stamp; the query returns a non-zero pid and version 3 |
| AC-002 | REQ-010 | M1 | release-blocking | R-04 | M1 adds the `retired` transition and the `run.retired` event |
| AC-003 | REQ-003/003b | M2 | release-blocking | R-05 | M2 adds the classifier; the PID-reuse case returns `dead` |
| AC-004 | REQ-004 | M2/M3 | release-blocking | R-05 | M2/M3 retire dead owners; the join succeeds |
| AC-005 | REQ-005 | M2/M3 | release-blocking | R-05 | M2/M3 leave a live owner's run `active` |
| AC-006 | REQ-005 | M2 | release-blocking | R-05 | M2 leaves an indeterminate owner's run `active` |
| AC-007 | REQ-006 | M2/M6 | release-blocking | R-03 | M1 migration + M2 peer fallback retire the legacy rows |
| AC-008 | REQ-007 | M3 | release-blocking | R-05 | M3 adds classifications to the ambiguity error text |
| AC-009 | REQ-014 | M3 | release-blocking | R-05 | M3 preserves both sentinels through the new path |
| AC-010 | REQ-008/009 | M4 | release-blocking | R-06 | M4 adds `moai factory runs` and its refuse-live branch |
| AC-011 | REQ-011 | M5 | release-blocking | R-08 | M5 executes `moai glm -f` and records the result |
| AC-012 | REQ-011 | M5 | release-blocking | R-08 | M5 executes `moai codex -f` and records the result |
| AC-013 | REQ-013 | M6 | leg 1 release-blocking / leg 2 post-merge | R-07 | M6 places the exercise under the three-OS path; leg 1 green is `--- PASS: TestFactoryRunRetire` locally, leg 2 is the develop-push run |
| AC-014 | REQ-012 | M1-M6 | release-blocking | R-05 | each exercise records its isolation evidence |
| AC-015 | REQ-004/005 | M6 | release-blocking | R-05 | M6 mutation, both directions |
| AC-016 | REQ-002b/002c | M1 | release-blocking | R-01 | M1 restamp makes stamp and lead peer name one process |

## §D Acceptance Criteria (Given-When-Then)

- **AC-001** Given a sandbox project with no factory state, When a factory run start is recorded,
  Then the `runs` row for that run id carries a non-zero `lead_pid` and a non-empty
  `lead_process_start`, read back with `sqlite3 <factory.db> "SELECT run_id,lead_pid,lead_process_start FROM runs"`,
  and the database reports `schema_version = 3`.

- **AC-002** Given an active run whose owner is dead, When it is retired, Then the `runs` row still
  exists with `status='retired'` (`SELECT count(*) FROM runs WHERE run_id=?` returns 1), and an
  `events` row of kind `run.retired` exists for that run id.

- **AC-003** Given a recorded owner identity whose PID is live but whose process-start fingerprint
  differs from the recorded one (the PID-reuse shape), When the classifier runs, Then it returns
  `dead` — and given the same PID with the matching fingerprint it returns `live`. A classifier that
  consults only the PID fails this criterion. The fingerprints are supplied as fixture values, not
  harvested by racing a real PID-reuse: on linux the fingerprint has one-second resolution
  (`ps -o lstart=`), so a real same-second reuse is indistinguishable there by construction. The
  test asserts the REQ-003b direction explicitly — an indistinguishable fingerprint pair returns
  `live`, never `dead`.

- **AC-004** Given a sandbox holding two `active` runs whose owners are both dead, When a worker
  joins with no `--factory-run`, Then the join succeeds, and the `runs` table afterwards shows the
  run it joined as the sole `active` row with the other transitioned to `retired`.

- **AC-005** Given two `active` runs, one owned by a process that is still alive for the whole
  exercise and one owned by a dead process, When reconciliation runs, Then the live-owner run is
  still `active` afterwards and the dead-owner run is `retired`. The live owner is a controlled
  long-lived process held open across the measurement, not an assumption.

- **AC-006** Given an `active` run whose owner identity cannot be probed (no stamp and no lead peer),
  When reconciliation runs, Then that run is still `active` afterwards and is reported
  `indeterminate`.

- **AC-007** Given a factory database created at schema version 2 holding two unstamped `active`
  rows, each with a registered `role='lead'` peer whose PID is dead, When a worker joins with no
  `--factory-run`, Then the database migrates to version 3, both legacy rows are retired through the
  peer fallback, and the join no longer fails with `AMBIGUOUS_FACTORY`. This is the migration path:
  a run phase that only prevents new duplicates does not satisfy it.

- **AC-008** Given two `active` runs that both survive reconciliation (one live owner, one
  indeterminate), When resolution runs, Then it fails with an error whose text contains
  `AMBIGUOUS_FACTORY`, both surviving run ids, and each one's classification.

- **AC-009** Given zero `active` runs, When resolution runs, Then it fails with `NO_ACTIVE_FACTORY`;
  and given two `active` runs with live owners, Then it fails with `AMBIGUOUS_FACTORY`. Neither case
  returns a run id.

- **AC-010** Given a sandbox with one live-owner and one dead-owner `active` run, When
  `moai factory runs` is invoked, Then its output names both runs with their status and owner
  classification; and When `moai factory runs --retire <live-owner-run-id>` is invoked, Then it exits
  non-zero, the run remains `active`, and the message names liveness as the reason.

- **AC-011** Given the isolated sandbox, When `moai glm -f` is **executed** (not read) and the
  `runs` table is captured immediately afterwards, Then `progress.md` records the invocation, its
  exit code, and the verbatim table — and states whether the glm door exhibits the same
  record-without-retirement behaviour as the `cc` door. The criterion is satisfied only by a
  recorded invocation; quoting `internal/cli/glm.go:242` does not satisfy it.

- **AC-012** Given the isolated sandbox, When `moai codex -f` is **executed** and the `runs` table is
  captured immediately afterwards, Then `progress.md` records the invocation, its exit code, and the
  verbatim table, with the same source-citation exclusion as AC-011.

- **AC-013** — two legs with different timing; only the first gates the merge.

  **Leg 1 — pre-merge, release-blocking.** Given the liveness-predicate and reconciler exercise
  placed at `test/integration/harness/it08_factory_run_retire_test.go` behind
  `//go:build integration`, When `go test -tags=integration -v ./test/integration/harness/ -run TestFactoryRunRetire`
  is run locally on darwin, Then the output contains a `--- PASS: TestFactoryRunRetire` line.

  The `-v` and the `--- PASS:` line are the load-bearing part, not decoration: a Go test binary
  given a selector that matches **zero** tests exits 0 and prints `ok`, so a bare `ok` would satisfy
  a weaker wording while the three-OS job silently runs nothing. This leg asserts the test is
  actually *selected*, which is the property the post-merge leg depends on.

  **Leg 2 — post-merge, confirmation only, NOT a merge gate.** Given leg 1 passed and the card's
  branch has merged, When the develop push triggers `test-integration`, Then jobs
  `Integration Tests (ubuntu-latest)`, `(macos-latest)`, and `(windows-latest)` each report a
  result, recorded by run id in `progress.md`.

  **Why leg 2 cannot gate the merge.** `ci.yml` triggers on `push: [main, develop]` and
  `pull_request: [main]` — there is no `pull_request` trigger for `develop`, and this project does
  not push `WT-` branches, so a card branch gets no CI run before it merges (`spec.md` §A.2 carries
  the line citations and the confirming run `35802361895`, whose three Integration Tests jobs each
  reported `success` on a develop push). Writing leg 2 as a merge gate would make the criterion
  false about its own timing. Pre-merge coverage is darwin only; ubuntu / macos / windows are
  post-merge, and `spec.md` §F carries that as named residual risk.

- **AC-014** Given any exercise in AC-001..AC-012, When it completes, Then the factory state it
  created lives only under the sandbox `MOAI_HOME` and carries a project key that is not this
  repository's — recorded as the directory listing plus the key.

- **AC-016** Given a run recorded through a **spawn-shaped** launch (the Windows shape, exercised
  under test by driving the restamp path directly rather than by requiring a Windows host), When the
  run row and the run's `role='lead'` peer are both read, Then the `lead_pid` /
  `lead_process_start` on the row equal the `pid` / `process_start` on the peer — the two sources
  name one process. And given a **replace-shaped** launch (the POSIX `syscall.Exec` shape), Then the
  same equality holds without a restamp. A design in which the row names the launcher while the peer
  names the child fails this criterion: that divergence is the D1 defect this SPEC was revised to
  remove, and it is the shape that would retire a live session's run.

## §D.1 Mutation criteria (both directions)

- **AC-015a** Given the implemented reconciler, When the liveness guard is removed so that a `live`
  or `indeterminate` owner is treated as retirable, Then at least one test FAILS. A guard whose
  removal leaves the suite green is not tested.

- **AC-015b** Given the implemented reconciler, When the retirement step is removed so that a
  provably-dead owner's run stays `active`, Then at least one test FAILS.

Both directions are required: AC-015b alone would pass on a reconciler that retires everything, and
AC-015a alone would pass on one that retires nothing.

## §D.2 Severity

All sixteen criteria are **release-blocking**, including AC-013 leg 1: each either protects a live
lead's run or establishes that a dead lead's run actually leaves `active`.

**AC-013 leg 2 is the one non-gating obligation.** It is not release-blocking and is never recorded
as a pass at integration time, because no CI run exists for a card branch before it merges — the
three-OS result lands on the develop push that follows. It is recorded as **pending** at close and
**confirmed** when the run id is read. This is a timing fact about the CI wiring (`spec.md` §A.2),
not a weakened criterion: leg 1 gates the merge on the property leg 2 depends on — that the test is
actually selected by the integration path.

## §D.3 Traceability

REQ-001→AC-001 · REQ-002→AC-001 · REQ-002b→AC-016 · REQ-003→AC-003 · REQ-003b→AC-003 ·
REQ-004→AC-004/AC-015b · REQ-005→AC-005/AC-006/AC-015a · REQ-006→AC-007 · REQ-007→AC-008 ·
REQ-008→AC-010 · REQ-009→AC-010 · REQ-010→AC-002 · REQ-011→AC-011/AC-012 · REQ-012→AC-014 ·
REQ-013→AC-013 (both legs) · REQ-014→AC-009.

Every one of the 16 REQs in `spec.md` §B has at least one AC; every one of the 16 ACs traces to at
least one REQ. Both counts sit exactly at the Tier M budget ceiling (16 requirements, 16 acceptance
criteria, applied independently per `spec-workflow.md` § SPEC Complexity Tier).

## §D.4 Definition of Done

- All 16 release-blocking criteria (AC-001..AC-016, AC-013 counted at leg 1) recorded PASS in
  `progress.md` §E.2 with command plus verbatim output. AC-013 leg 2 is recorded **pending with its
  reason**, never as a pass, and is confirmed by run id after the develop push.
- `go test ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/...` passes in the run
  tree; CI supplies the full-suite and cross-platform verdict.
- `golangci-lint run` reports zero findings on the changed packages.
- No file from the t1082 or t1109 sets appears in `git diff --name-only` against the base.
- The `AMBIGUOUS_FACTORY` and `NO_ACTIVE_FACTORY` sentinels still appear in resolution failures
  (AC-009), so no downstream matcher is silently broken.
