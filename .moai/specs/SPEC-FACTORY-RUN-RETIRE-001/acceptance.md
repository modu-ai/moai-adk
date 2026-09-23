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
| R-02 | `grep -c "factorySchemaVersion = 3" internal/homestate/factory.go` | `0` | 1 | schema is still at version 2 |
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
| AC-010 | REQ-008, REQ-005 | M4 | release-blocking | R-06 | M4 adds `moai factory runs` and its refuse branch, refusing both `live` and `indeterminate` |
| AC-011 | REQ-011 | M5 | release-blocking | R-08 | M5 executes all four lead doors, pane door included, one evidence row each |
| AC-012 | REQ-002d | M5 | release-blocking | R-08 | M5 exercises the no-identity refusal and shows no launcher-stamped run survives it |
| AC-013 | REQ-013 | M6 | leg 1 release-blocking / leg 2 post-merge | R-07 | M6 places the exercise under the three-OS path; leg 1 green is `--- PASS: TestFactoryRunRetire` locally, leg 2 is the develop-push run |
| AC-014 | REQ-012 | M1-M6 | release-blocking | R-05 | each exercise records its isolation evidence |
| AC-015 | REQ-004/005 | M6 | release-blocking | R-05 | M6 mutation, both directions |
| AC-016 | REQ-002b, REQ-013 | M1 | release-blocking | R-01 | M1 restamp makes stamp and lead peer name one process, driven through the build-tag-free seam on all three shapes |
| AC-017 | REQ-005 | M2/M4 | release-blocking | R-05 | M2/M4 retire only on positive `dead`; an unenumerated classification declines by default |

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

- **AC-010** Given a sandbox holding one live-owner, one dead-owner, and one **indeterminate-owner**
  `active` run, When `moai factory runs` is invoked, Then its output names all three with their
  status and owner classification; and When `--retire` is invoked against each in turn, Then the
  live-owner run is refused, **the indeterminate-owner run is refused**, and only the dead-owner run
  is retired — each refusal exiting non-zero, leaving the run `active`, and naming its classification
  as the reason.

  The indeterminate leg is the D14 defect: an earlier draft bound the command to `live` alone, which
  let a deliberate `--retire` reach a run whose owner could not be probed — the same live-run
  retirement the reconciler is forbidden from performing, arrived at through the operator surface
  instead. REQ-005 now binds every path, so this criterion and AC-006 assert one invariant on two
  surfaces rather than two invariants that can drift apart.

- **AC-011** Given the isolated sandbox, When **each** lead door is **executed** (not read) and the
  `runs` table is captured immediately afterwards, Then `progress.md` carries one row per door —
  invocation, exit code, verbatim `runs` table, and the stamped `lead_pid` / `lead_process_start`
  against the run's `role='lead'` peer identity. The doors, matching `spec.md` §A.1:

  | Door | Shape | Obligation |
  |---|---|---|
  | `moai cc -f` | replace | executed |
  | `moai glm -f` | replace | executed |
  | `moai codex -f` | replace | executed |
  | `moai codex -f --spawn` | **pane** | executed — this is the D11 door; it is the one whose stamp must survive the launcher exiting |

  The criterion is satisfied only by recorded invocations; quoting `internal/cli/glm.go:242` or
  `codex_launcher.go:230` does not satisfy it. The two `codex_direct_*` sites are **not** listed as
  executed: `spec.md` §A.1 asserts each is covered because its shape matches an executed door, and
  that assertion — not silence — is what discharges them.

- **AC-012** Given a launch through a door that does not replace the launching process, When the
  session identity cannot be obtained (the pane-identity resolver exhausting its deadline, or the
  spawned child failing its liveness probe), Then the launch is refused with a non-zero exit **and**
  no `runs` row is left carrying the launching process's PID — verified by querying
  `lead_pid` for that run id after the refusal and finding either no row or a row not naming the
  launcher. A refusal that cleans up the pane but leaves the run stamped fails this criterion:
  that residue is the D11 defect with the launch merely failing earlier.

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
  `Integration Tests (ubuntu-latest)`, `(macos-latest)`, and `(windows-latest)` each report
  conclusion **`success`** — recorded by run id in `progress.md`, together with the per-job
  conclusions read from `gh run view <id> --json jobs`.

  "Reports a result" is deliberately **not** the condition: a FAILING job also reports a result, so
  that wording would be satisfied by the very outcome this criterion exists to catch. (This was the
  wording at `320cdeb90`; re-read there rather than assumed resolved.)

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

- **AC-016** Given a run recorded through each of the three launch shapes, When the run row and the
  run's `role='lead'` peer are both read, Then the row's `lead_pid` / `lead_process_start` equal the
  peer's `pid` / `process_start` — the two sources name one process — for **replace** (no restamp
  needed), **spawn** (restamped to the child), and **pane** (restamped to the tmux pane process).

  **The restamp seam must be build-tag-free for this criterion to be assertable at all.** A restamp
  living only in `launch_exec_windows.go` sits behind `//go:build windows`, so a darwin host cannot
  even compile a call to it, and the criterion would be unrunnable on the one platform where the
  pane door actually exists. The seam therefore takes an already-resolved `(runID, pid, fingerprint)`
  and carries no build tag; the platform files and the pane door each supply the identity they
  resolved. This criterion drives that seam directly with fixture identities, so it runs on any host
  and asserts all three shapes — no Windows host and no live tmux server required.

  **Second leg — the seam is actually called at every non-replace call site.** Driving the seam with
  fixture identities proves the seam *behaves*; it says nothing about whether each door *reaches* it.
  A door that never calls the seam leaves its run stamped with the launching process and passes the
  first leg unchanged — which is exactly how defect D18 survived three audits. This leg therefore
  asserts, at source level, that a restamp-seam call is present at each of the three non-replace call
  sites: `internal/cli/launch_exec_windows.go` (spawn), `internal/cli/codex_direct_windows.go`
  (spawn), and `internal/cli/codex_launcher.go` (pane). The two replace-shaped sites
  (`launch_exec_posix.go`, `codex_direct_posix.go`) are asserted **absent** from that required set —
  a replace-shaped door needs no restamp, and demanding one there would make the leg false about the
  design it is checking.

  The assertion is **source-level**, reading the call sites rather than executing the doors, because
  two of the three sit behind `//go:build windows`: the darwin host this card runs on can neither
  execute them nor compile a call to them. That host constraint is what kept the gap invisible, so
  the leg is written to hold without it — no Windows host, no live tmux server.

  **Mutation direction, one site at a time.** Delete the seam call at `launch_exec_windows.go` and
  re-run: this leg turns red. Restore it, delete the one at `codex_direct_windows.go`, re-run: red.
  Restore, delete the one at `codex_launcher.go`, re-run: red. All three deletions are probed
  **individually and separately** — an assertion that only goes red when all three are missing stays
  green while two sites are correct and one is silently not, which is the vacuous shape this SPEC
  rejects throughout. A leg that survives any single-site deletion has not closed D18.

  A design in which the row names the launcher while the peer names the session fails this
  criterion. That divergence is the D1 defect on the spawn shape and the D11 defect on the pane
  shape — the same failure twice, and the shape that retires a live session's run.

- **AC-017** Given a classification value the retirement code does **not** enumerate — a fourth
  state introduced by the test, beyond `live` / `dead` / `indeterminate` — When it is put through
  every retirement path (reconciler, migration pass, and `--retire`), Then each path declines to
  retire, without having been told about that value.

  **This is the criterion that separates a positive rule from a reject-list, and nothing else in
  this file does.** An enumeration written "reject `live`, reject `indeterminate`" passes AC-005,
  AC-006 and AC-010 unchanged — those name the two states it happens to list — and then silently
  retires the fourth state the day one is added. Only a probe with an unenumerated value can tell
  the two implementations apart.

  **Mutant probe.** Rewrite the guard as `if c == live || c == indeterminate { refuse }` and re-run:
  AC-005, AC-006 and AC-010 stay green, AC-017 turns red. That divergence is the whole point of the
  criterion — if a mutant can satisfy every other criterion while violating REQ-005, the rest of the
  file is too shallow to adopt the requirement on its own.

## §D.1 Mutation criteria (both directions)

- **AC-015a** Given the implemented reconciler, When the liveness guard is removed so that a `live`
  or `indeterminate` owner is treated as retirable, Then at least one test FAILS. A guard whose
  removal leaves the suite green is not tested.

- **AC-015b** Given the implemented reconciler, When the retirement step is removed so that a
  provably-dead owner's run stays `active`, Then at least one test FAILS.

Both directions are required: AC-015b alone would pass on a reconciler that retires everything, and
AC-015a alone would pass on one that retires nothing.

## §D.2 Severity

All 17 criteria are **release-blocking**, including AC-013 leg 1: each either protects a live
lead's run or establishes that a dead lead's run actually leaves `active`.

**AC-013 leg 2 is the one non-gating obligation.** It is not release-blocking and is never recorded
as a pass at integration time, because no CI run exists for a card branch before it merges — the
three-OS result lands on the develop push that follows. It is recorded as **pending** at close and
**confirmed** when the run id is read. This is a timing fact about the CI wiring (`spec.md` §A.2),
not a weakened criterion: leg 1 gates the merge on the property leg 2 depends on — that the test is
actually selected by the integration path.

## §D.3 Traceability

REQ-001→AC-001 · REQ-002→AC-001 · REQ-002b→AC-016 · REQ-002d→AC-012 · REQ-003→AC-003 ·
REQ-003b→AC-003 · REQ-004→AC-004/AC-015b · REQ-005→AC-005/AC-006/AC-010/AC-015a/**AC-017** ·
REQ-006→AC-007 · REQ-007→AC-008 · REQ-008→AC-010 · REQ-010→AC-002 · REQ-011→AC-011 ·
REQ-012→AC-014 · REQ-013→AC-013 (both legs) + AC-016 (the seam) · REQ-014→AC-009.

`REQ-009` is absent by design — retired into REQ-005 at v0.4.0, its number left as a gap rather than
closed by renumbering (`spec.md` §B).

Every one of the 16 REQs in `spec.md` §B has at least one AC; every one of the 17 ACs traces to at
least one REQ. The AC count exceeds the Tier M ceiling of 16, so this SPEC is **Tier L** (ceilings
25/25) — see § D.5 for the arithmetic and `spec.md` §H for the tier-change record.

## §D.4 Definition of Done

- All 17 release-blocking criteria (AC-001..AC-017, AC-013 counted at leg 1) recorded PASS in
  `progress.md` §E.2 with command plus verbatim output. AC-013 leg 2 is recorded **pending with its
  reason**, never as a pass, and is confirmed by run id after the develop push.
- The AC-017 mutant probe run and recorded: with the reject-list mutant in place, AC-005 / AC-006 /
  AC-010 green and AC-017 red. A run that cannot show that divergence has not exercised REQ-005.
- `go test ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/...` passes in the run
  tree; CI supplies the full-suite and cross-platform verdict.
- `golangci-lint run` reports zero findings on the changed packages.
- No file from the t1082 or t1109 sets appears in `git diff --name-only` against the base.
- The `AMBIGUOUS_FACTORY` and `NO_ACTIVE_FACTORY` sentinels still appear in resolution failures
  (AC-009), so no downstream matcher is silently broken.

## §D.5 Budget record — the count was written first, and the tier followed it

Scope **increased** at v0.4.0 by operator decision (the pane door), and again at v0.5.0 (AC-017).
Tier-up is pre-authorized, so this section records the arithmetic rather than negotiating it: the
counts are what the scope needs, and `tier: L` follows from them.

**Requirements: 16.**

- **+1** `REQ-002d` — the refuse-when-no-identity path. New behaviour, distinct trigger; it cannot
  fold into REQ-002b, which states a restamp.
- **−1** `REQ-009` retired into `REQ-005`. This consolidation stands on its **own merit**, not on
  the count — and D14's sharpened framing makes that clearer than it was at v0.4.0. The invariant
  wants to be stated **once and positively** (retire only on a positive `dead`); re-splitting it
  across a reconciler requirement and a command requirement is precisely what let the two copies
  drift to different reject-lists. Restoring REQ-009 would reintroduce the defect, so the merge is
  now the only correct shape regardless of budget.

**Acceptance criteria: 17 — over the Tier M ceiling of 16.**

- **+1** `AC-012` reassigned to the REQ-002d refusal path.
- **−1** the codex-door execution folded into `AC-011`, one evidence row per lead door. Substantive:
  the old AC-011/AC-012 pair named **two** doors while REQ-011 governs **five call sites across four
  doors**, so the split was already incomplete against its own requirement.
- **+1** `AC-017` — the unenumerated-classification probe. This is the criterion the SPEC could not
  do without: every other criterion is satisfied by a reject-list mutant that violates REQ-005, so
  without AC-017 the positive rule is asserted and never tested.

**Consequence: this SPEC is Tier L.** 17 acceptance criteria exceed the Tier M ceiling, so the tier
rises to L (ceilings 25/25) and the artifact set gains `design.md` + `research.md`. The
plan-auditor PASS threshold rises with it, **0.80 → 0.85**.

The LOC and file-count guidance still reads Tier M (~8 files, well under 1000 LOC); the **budget**
is what carries this SPEC over, and per `spec-workflow.md` the ceilings are the binding constraint —
"exceeding either ceiling is a signal to tier up or to split the SPEC, not to relax the budget". No
criterion was dropped, merged, or renumbered to avoid the tier change.
