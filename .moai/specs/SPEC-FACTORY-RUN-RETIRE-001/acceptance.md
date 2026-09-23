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

## §C AC Matrix

| AC | Requirement | Milestone | Severity | Verification |
|----|-------------|-----------|----------|--------------|
| AC-001 | REQ-001/002 | M1 | MUST | schema read + stamped-row query |
| AC-002 | REQ-010 | M1 | MUST | row survives retirement + event row present |
| AC-003 | REQ-003 | M2 | MUST | classifier unit test, PID-reuse case |
| AC-004 | REQ-004 | M2/M3 | MUST | dead-owner run retired, join succeeds |
| AC-005 | REQ-005 | M2/M3 | MUST | live-owner run NOT retired |
| AC-006 | REQ-005 | M2 | MUST | indeterminate-owner run NOT retired |
| AC-007 | REQ-006 | M2/M6 | MUST | v2 legacy DB reconciled via peer fallback |
| AC-008 | REQ-007 | M3 | MUST | surviving-ambiguity error text |
| AC-009 | REQ-014 | M3 | MUST | fail-closed regression, both directions |
| AC-010 | REQ-008/009 | M4 | MUST | `moai factory runs` list + refuse-live |
| AC-011 | REQ-011 | M5 | MUST | executed `moai glm -f`, captured `runs` |
| AC-012 | REQ-011 | M5 | MUST | executed `moai codex -f`, captured `runs` |
| AC-013 | REQ-013 | M6 | MUST | darwin + linux + windows CI result |
| AC-014 | REQ-012 | M1-M6 | MUST | isolation evidence per exercise |
| AC-015 | REQ-004/005 | M6 | MUST | mutation, both directions |

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
  consults only the PID fails this criterion.

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

- **AC-013** Given the merged branch, When CI runs, Then the liveness-predicate and reconciler tests
  report a result on darwin, linux, and windows. The originating reproduction was darwin-only; a
  darwin-only test result does not satisfy this criterion.

- **AC-014** Given any exercise in AC-001..AC-012, When it completes, Then the factory state it
  created lives only under the sandbox `MOAI_HOME` and carries a project key that is not this
  repository's — recorded as the directory listing plus the key.

## §D.1 Mutation criteria (both directions)

- **AC-015a** Given the implemented reconciler, When the liveness guard is removed so that a `live`
  or `indeterminate` owner is treated as retirable, Then at least one test FAILS. A guard whose
  removal leaves the suite green is not tested.

- **AC-015b** Given the implemented reconciler, When the retirement step is removed so that a
  provably-dead owner's run stays `active`, Then at least one test FAILS.

Both directions are required: AC-015b alone would pass on a reconciler that retires everything, and
AC-015a alone would pass on one that retires nothing.

## §D.2 Severity

All criteria above are MUST. There are no nice-to-have criteria in this SPEC: every one of them
either protects a live lead's run or establishes that a dead lead's run actually leaves `active`.

## §D.3 Traceability

REQ-001→AC-001 · REQ-002→AC-001 · REQ-003→AC-003 · REQ-004→AC-004/AC-015b · REQ-005→AC-005/AC-006/AC-015a ·
REQ-006→AC-007 · REQ-007→AC-008 · REQ-008→AC-010 · REQ-009→AC-010 · REQ-010→AC-002 ·
REQ-011→AC-011/AC-012 · REQ-012→AC-014 · REQ-013→AC-013 · REQ-014→AC-009.

Every REQ in `spec.md` §B has at least one AC; every AC traces to at least one REQ.

## §D.4 Definition of Done

- All 15 ACs recorded PASS in `progress.md` §E.2 with command plus verbatim output.
- `go test ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/...` passes in the run
  tree; CI supplies the full-suite and cross-platform verdict.
- `golangci-lint run` reports zero findings on the changed packages.
- No file from the t1082 or t1109 sets appears in `git diff --name-only` against the base.
- The `AMBIGUOUS_FACTORY` and `NO_ACTIVE_FACTORY` sentinels still appear in resolution failures
  (AC-009), so no downstream matcher is silently broken.
