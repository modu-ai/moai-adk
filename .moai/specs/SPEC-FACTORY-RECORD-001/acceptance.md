# acceptance.md — SPEC-FACTORY-RECORD-001 (card t1239)

## §A. Scope of verification

The criteria verify the F1 record layer: schema migration, the transition table and its concurrency
control, the three evidence readers and the reserved CI edges, the auditor verdict-line producer, the
lease and the decision-pending states, the hint gates, the three `moai factory` commands, the queue
boundary, and the mirror write on both existing dispatch paths. F2 and F3 behaviour is not verified
here.

## §B. Test-environment constraint (binds every AC)

- Every test builds its own project root with `t.TempDir()` and its own fixture git repository; the
  MoAI home is redirected to a temporary directory. No test reads or writes the real `~/.moai`.
- CLI-level tests run only with a `-run` filter naming the AC tests
  (`go test ./internal/cli -run 'TestFR_AC0' -count=1`). Whole-package `./internal/cli` and
  `./internal/hook` runs are prohibited for this card (they write the real profile-lease database).
- **Pass convention:** an AC passes only when its test prints a `--- PASS: TestFR_AC0NN_…` line and
  the run's swept count is non-zero. A run that prints `[no tests to run]` or selects fewer tests than
  the AC names is a gap, not a pass.
- CLI assertions check output content and database content. `moai factory <anything>` exits 0 today
  even for a nonexistent subcommand (research.md R9), so an exit code alone decides nothing.
- Lease expiry is driven by an injected clock, never by sleeping.

## §C. Two-cell adoption

### C.1 RED-now evidence ledger

Tree for every entry: `553e224f3` (document-level pin; binds every entry below). Commands are run from
the worktree root.

```
R-01
command: grep -n "const factorySchemaVersion" internal/homestate/factory.go
stdout:  20:const factorySchemaVersion = 3
exit:    0
red because: F1 requires schema version 4; version 3 has no F1 columns.

R-02
command: grep -c "lease_expires_at" internal/homestate/factory.go
stdout:  0
exit:    1
red because: no lease column exists in the factory schema.

R-03
command: grep -n "version=cards.version+1" internal/homestate/runtime.go
stdout:  68:owner_label=excluded.owner_label,state=excluded.state,version=cards.version+1,evidence_path=excluded.evidence_path,updated_at=excluded.updated_at`,
exit:    0
red because: the only cards writer increments the version without comparing it (no expected-version
             clause), so a stale write is accepted.

R-04
command: grep -rn 'Use: *"status"' internal/cli/factory_handoff_recover.go
stdout:  (empty)
exit:    1
red because: the factory command tree has no status, assign, or decide subcommand.

R-05
command: grep -rn "UPDATE workers" internal --include='*.go'
stdout:  (empty)
exit:    1
red because: nothing ever refreshes workers.heartbeat_at.

R-06
command: grep -rln "card.transition" internal
stdout:  (empty)
exit:    1
red because: no transition API, no transition table, and no transition event kind exist.

R-07
command: grep -n "OpenFactory" internal/cli/gtd.go
stdout:  (empty)
exit:    1
red because: the queue dispatch path writes only the queue runtime report, never factory.db.

R-08
command: grep -rln "hint_after" internal
stdout:  (empty)
exit:    1
red because: no hint, decision, SHA, or contract-pointer column exists.

R-09
command: grep -c "audited_sha" .claude/agents/moai/plan-auditor.md
stdout:  0
exit:    1
red because: no auditor emits the machine-readable audited-SHA line the E-VERDICT gate reads.

R-10
command: grep -n "OpenFactory" internal/cli/goal.go
stdout:  (empty)
exit:    1
red because: the auto-mission dispatch path writes only the queue runtime report, never factory.db.

R-11
command: grep -rln "failure_reason" internal/homestate
stdout:  (empty)
exit:    1
red because: no failure-reason field exists, so a failed card records no reason.
```

Positive control for the empty greps: R-01 and R-03 use the same tool against the same tree and print a
match, so an empty result in R-02 / R-04-R-11 is an absence in this tree, not a broken probe.

### C.2 AC matrix

| AC | REQ | Milestone | Class | RED cell | Green path |
|----|-----|-----------|-------|----------|------------|
| AC-001 | REQ-FR-001, REQ-FR-002 | M1 | release-blocking | R-01, R-08 | M1 migration and DDL |
| AC-002 | REQ-FR-003 | M1/M2 | release-blocking | R-06 | M2 legacy-state refusal |
| AC-003 | REQ-FR-005 | M2 | release-blocking | R-03, R-06 | M2 transaction; version+1 and one event |
| AC-004 | REQ-FR-005 | M2 | release-blocking | R-03 | M2 version compare |
| AC-005 | REQ-FR-004 | M2 | release-blocking | R-06 | M2 table; 65 accepted, 296 refused |
| AC-006 | REQ-FR-005 | M2 | release-blocking | R-03 | M2; exactly one of two racing writers wins |
| AC-007 | REQ-FR-007, REQ-FR-010 | M3 | release-blocking | R-06 | M3 E-ENTRY |
| AC-008 | REQ-FR-008, REQ-FR-010 | M3 | release-blocking | R-06 | M3 E-VERDICT |
| AC-009 | REQ-FR-009, REQ-FR-010 | M3 | release-blocking | R-06 | M3 E-MERGE |
| AC-010 | REQ-FR-012 | M4 | release-blocking | R-02, R-05 | M4 acquire / renew |
| AC-011 | REQ-FR-013 | M4 | release-blocking | R-02 | M4 expiry return |
| AC-012 | REQ-FR-014 | M4 | release-blocking | R-02 | M4 merging → blocked |
| AC-013 | REQ-FR-018, REQ-FR-013 | M4 | release-blocking | R-02, R-06 | M4 kickoff releases the lease |
| AC-014 | REQ-FR-016, REQ-FR-017, REQ-FR-023 | M5 | release-blocking | R-08 | M5 `assign --after / --prefer` |
| AC-015 | REQ-FR-019, REQ-FR-023 | M5 | release-blocking | R-04 | M5 `decide --gate kickoff` |
| AC-016 | REQ-FR-020, REQ-FR-018 | M5 | release-blocking | R-04 | M5 `decide --choice resume/block/abandon` |
| AC-017 | REQ-FR-020, REQ-FR-015 | M4/M5 | release-blocking | R-04, R-11 | M5 `unblock`, M4 `failed` |
| AC-018 | REQ-FR-021 | M3/M5 | release-blocking | R-04 | M5 `decide --gate push` |
| AC-019 | REQ-FR-006 | M2 | release-blocking | R-06 | M2 reserved-edge refusal |
| AC-020 | REQ-FR-011, REQ-FR-008 | M3b | release-blocking | R-09 | M3b producer lines + emitter regenerated |
| AC-021 | REQ-FR-022 | M5 | regression-guard | (green today) | must stay green |
| AC-022 | REQ-FR-022, REQ-FR-023 | M5 | release-blocking | R-04 | M5 queue-picked precondition |
| AC-023 | REQ-FR-023, REQ-FR-024, REQ-FR-001 | M5 | release-blocking | R-04 | M5 `status` |
| AC-024 | REQ-FR-025 | M6 | release-blocking | R-07 | M6 queue-dispatch mirror |
| AC-025 | REQ-FR-025 | M6 | release-blocking | R-10 | M6 auto-mission-dispatch mirror |

AC-021 is a regression guard: the queue schema is frozen and its guard is green on `553e224f3`, so it
cannot be adopted as RED; it proves F1 did not break it.

## §D. Acceptance Criteria (Given-When-Then)

- **AC-001** (maps REQ-FR-001, REQ-FR-002) Given a factory database created by the current code at
  schema version 3 and seeded with one row in each of `runs`, `workers`, `cards`, `events`, and
  `resume_handoffs`, When it is opened by the F1 store, Then `meta.schema_version` reads `4`, every
  seeded row is present with its original values, every new `cards` column of an existing row reads
  `''`, and opening it a second time changes nothing (no error, no duplicate column, identical row
  count); and given no database, When the F1 store creates one, Then `PRAGMA table_info(cards)` lists
  every column named in `design.md` § Schema (including `failure_reason` and the four contract-pointer
  columns), and a database whose version reads `5` is refused with an unsupported-version error.

- **AC-002** (maps REQ-FR-003) Given a card row whose `state` is `in_progress` (a value outside the F1
  set), When any transition other than `abandoned` is requested, Then it is refused with a legacy-state
  error and the row and the `events` count are unchanged; When `decide --choice abandon` is requested,
  Then the row reads `abandoned` with its version incremented.

- **AC-003** (maps REQ-FR-005) Given a card in `plan` at version `v`, When a legal transition with
  expected version `v` and valid evidence is applied, Then the row reads the target state at version
  `v+1`, exactly one new `events` row of kind `card.transition` exists whose payload names the card,
  `from`, `to`, `v+1`, and the actor; and a fault injected after the state update but before the event
  insert leaves both the row and the event count exactly as before (neither is committed).

- **AC-004** (maps REQ-FR-005) Given a card at version `v`, When a transition is requested with expected
  version `v-1`, Then it is refused with a stale-version error, and a byte comparison of the card row
  and a count of `events` before and after show no change.

- **AC-005** (maps REQ-FR-004) Given a fixture repository **with** a remote whose remote-tracking
  integration ref contains every merge SHA used, and a card placed in each of the 19 states in turn
  with an unexpired lease where the state is lease-holding, When every one of the 19 target states is
  requested from it with the current version and a fixture that satisfies that row's guard (for T4a-f
  the card's `stage` equals the target; for T22a-e `decision_resume` names the target; for evidence
  rows the evidence is valid), Then exactly **65** pairs are accepted and **296** are refused, the
  accepted set equals the rows of `design.md` § Transition Table counted there, T18, T19 and T20 are
  among the refused, and the test fails if the table it enumerates contains a duplicate pair; and given
  a `leased` card whose `stage` is `sync`, When `leased → plan` or `leased → run` is requested, Then each
  is refused (guard mismatch) while `leased → sync` is accepted.

- **AC-006** (maps REQ-FR-005) Given one card at version `v` and two store connections in separate
  goroutines, When both request a legal transition with expected version `v` at the same time, Then
  exactly one succeeds, the other returns a stale-version error, the row reads version `v+1`, and exactly
  one `card.transition` event exists. Repeated 50 times under `-race`, with zero double-success runs.

- **AC-007** (maps REQ-FR-007, REQ-FR-010) Given a card in `plan` with a fixture worktree, When entry
  into `plan-audit` is requested with (a) a SHA that does not exist, (b) a SHA that exists but is not an
  ancestor of the worktree HEAD, (c) a valid SHA and an artifact path absent at that commit, Then each is
  refused with an evidence error and the row is unchanged; When requested with a valid ancestor SHA and
  a path present at that commit, Then the card reads `plan-audit` with `evidence_sha` and
  `evidence_path` recorded. The API offers no parameter through which a caller can assert the evidence
  passed.

- **AC-008** (maps REQ-FR-008, REQ-FR-010) Given a card in `plan-audit` with `evidence_sha = S`, When
  `kickoff` is requested and the verdict file (a) is absent, (b) lacks the `verdict:` line, (c) reads
  `verdict: FAIL`, (d) reads `verdict: PASS` with `audited_sha:` naming a different commit, (e) carries
  two conflicting `verdict:` lines, Then each is refused and the row is unchanged; When the file reads
  `verdict: PASS-WITH-DEBT` and `audited_sha: S`, Then the card reads `kickoff` with
  `decision_gate = kickoff`; and the return edge to `plan` is accepted with the `verdict: FAIL` file of
  case (c).

- **AC-009** (maps REQ-FR-009, REQ-FR-010) Given a card in `merging` and a fixture repository with an
  integration branch, When `merged-local` is requested with (a) a single-parent commit, (b) a merge
  commit whose tree differs from its second parent's tree, (c) a merge commit not reachable from the
  integration branch, (d) a valid merge commit and a re-measure file that does not contain its SHA, Then
  each is refused and the row is unchanged; When requested with a valid merge commit and a re-measure
  file naming it, Then the card reads `merged-local` with `merge_sha`, `merge_tree`, and
  `remeasure_path` recorded.

- **AC-010** (maps REQ-FR-012) Given a card in `assigned` with owner label `worker-1`, When the lease is
  requested by a label absent from `workers`, or by `worker-2` registered in `workers`, Then both are
  refused; When requested by `worker-1` registered in `workers`, Then the card reads `leased` with holder
  `worker-1` and an expiry equal to the heartbeat time plus the lease duration; When `worker-1` renews,
  Then the expiry advances and `workers.heartbeat_at` for `worker-1` equals the new heartbeat time; When
  `worker-2` renews, Then it is refused and neither table changes.

- **AC-011** (maps REQ-FR-013) Given a card in `run` with `stage = run`, a worktree path, an
  `evidence_sha`, and a lease whose expiry is in the past, When any transition on that card is next
  requested, Then the card first reads `assigned` with an empty holder and `stage`, `worktree_path`, and
  `evidence_sha` unchanged, one `lease.expired` event is appended, the requested transition returns a
  lease-expired error, and a hash of the fixture worktree's `git rev-parse HEAD`, branch list, and
  `git status --porcelain` output is identical before and after. A subsequent lease by the owner followed
  by `leased → run` is accepted.

- **AC-012** (maps REQ-FR-014) Given a card in `merging` whose lease has expired, When any transition on
  it is next requested, Then the card reads `blocked` (not `assigned`) and one `lease.expired` event
  naming the interrupted merge is appended.

- **AC-013** (maps REQ-FR-018, REQ-FR-013) Given a `leased` card in `plan-audit` whose holder is
  `worker-1`, When T7 moves it to `kickoff`, Then `lease_holder` and `lease_expires_at` read `''` and
  `decision_gate = kickoff`; When the injected clock is advanced to twice the lease duration and
  `moai factory decide <card> --gate kickoff --choice approve` is run, Then the card reads `assigned` with
  `stage = run` and `decider = human`, no `lease.expired` event exists for that card, and the event log
  holds exactly one `card.transition` for the decision; the same holds for a card in `needs-decision`
  decided `resume` after the same clock advance.

- **AC-014** (maps REQ-FR-016, REQ-FR-017, REQ-FR-023) Given card `c2` recorded in `picked` with
  `--after c1`, When `assign c2 --to worker-1` is run while (a) `c1` has no factory record, (b) `c1` is
  in `sync`, Then (a) is refused with an unknown-predecessor error and (b) with a predecessor-unmerged
  error, and `c2` stays `picked`; When `c1` reads `merged-local`, Then the same command moves `c2` to
  `assigned` with owner `worker-1`; and given `assign c3 --prefer backend=codex`, `status --json` shows
  `prefer` exactly `backend=codex`, and `assign c3 --to worker-1` (a worker whose backend is not codex)
  succeeds — the hint refuses nothing.

- **AC-015** (maps REQ-FR-019, REQ-FR-023) Given cards `k1`, `k2`, `k3` in `kickoff` and card `r1` in
  `run`, When `moai factory decide k1 k2 r1 --gate kickoff --choice approve` is run, Then `k1` and `k2`
  read `assigned` with `stage = run` and `decider = human`, `r1` is reported as refused (not in
  `kickoff`) and is unchanged, and `k3` is unchanged; When `decide k3 --gate kickoff --choice reject` is
  run, Then `k3` reads `blocked`; When any `decide` is run with `--decider llm`, Then it is refused and
  nothing changes.

- **AC-016** (maps REQ-FR-020, REQ-FR-018) Given a card moved from `run` to `needs-decision` with a
  question, When it is read, Then `decision_question` holds the question, `decision_resume = run`, and
  the lease is cleared; When `decide --choice resume` is run, Then the card reads `assigned` with
  `stage = run`; given another card in `needs-decision` whose resume state is `merged-local`, `resume`
  returns it to `merged-local`; `--choice block` yields `blocked`; and given a plain v4 card in `run`
  (not legacy, not decision-pending), `--choice abandon` yields `abandoned` with its version incremented
  by exactly one and every other column unchanged.

- **AC-017** (maps REQ-FR-020, REQ-FR-015) Given a card in `blocked` with `stage = sync`, When
  `decide --choice unblock` is run, Then it reads `assigned` with `stage = sync`, and `unblock` on a card
  not in `blocked` is refused; and given a leased card in `run`, When the holder requests `failed` with an
  empty reason, Then it is refused and unchanged; When requested with reason `build broken`, Then the card
  reads `failed` with `failure_reason = build broken` and an empty lease, and every later transition from
  it (including `abandoned`) is refused.

- **AC-018** (maps REQ-FR-021) Given a card in `merged-local` with `merge_sha = M` and a fixture remote
  whose remote-tracking integration ref does not contain `M`, When `decide <card> --gate push` is run,
  Then it is refused and the card stays `merged-local`; after the fixture ref is advanced to contain `M`,
  the same command moves the card to `pushed`, and no `git fetch` is run by the command; and given a
  fixture repository with no remote, `decide <card> --gate push` moves the card to `done` with the event
  payload containing `no remote — no CI verdict`, while in the repository with a remote the direct
  `merged-local → done` edge is refused.

- **AC-019** (maps REQ-FR-006) Given a card in `pushed` and a card in `ci-green` (placed by fixture), When
  `pushed → ci-green` and `ci-green → done` are requested with the current version, and additionally with
  a caller-supplied file claiming a successful CI run for the merge SHA, Then every request is refused
  with a reserved-edge error whose message names F3, and both rows and the event count are unchanged.

- **AC-020** (maps REQ-FR-011, REQ-FR-008) Given the run-phase tree, When `grep -c 'audited_sha:' <file>` and
  `grep -c 'verdict: <PASS|PASS-WITH-DEBT|FAIL>' <file>` are run on each of
  `.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md`, their two template mirrors
  under `internal/template/templates/.claude/agents/moai/`, and both copies of
  `audit-artifact-convention.md`, Then each of the twelve runs prints a count of at least 1; `make agents-emit-check` exits 0
  and the two emitted `.toml` files contain `audited_sha`; the template-neutrality guard test passes; and
  a verdict file written exactly as the instruction's example (with `audited_sha` set to a fixture card's
  `evidence_sha`) is accepted by the E-VERDICT reader for `plan-audit → kickoff`; and, as a regression
  guard on the existing `AUDIT-VERDICT:` last-non-empty-line convention of SPEC-CODEX-AUDIT-GATE-AXES-001
  (research.md R16), checked with that SPEC's own reader `ParseVerdictLine`
  (`internal/auditreceipt/store.go:325`): (i) in each of the four agent files the
  `### [HARD] Cite your audit receipt` block still carries the `AUDIT-VERDICT:` literal and no
  `verdict:` / `audited_sha:` instruction appears after that block; (ii) a test that builds an
  auditor final message exactly as the edited instructions direct — report body, then the
  `AUDIT-VERDICT: PASS spec=<fixture> receipts=none` line last — gets `ok == true` from
  `ParseVerdictLine`, while the same message with a `verdict: PASS` line appended after it gets
  `ok == false` (the failure the guardrail prevents, observed once); and (iii)
  `go test ./internal/auditreceipt -run TestParseVerdictLine -count=1` still prints `--- PASS`.

- **AC-021** (maps REQ-FR-022) Given the queue schema guard, When
  `go test ./internal/kanban -run 'TestBacklogDowngrade_PreChangeBinaryStillServes' -count=1` runs, Then
  it prints `--- PASS`; and given a fixture `backlog.db`, When `assign`, `status`, and `decide` have each
  run against it, Then the `sql` column of every `sqlite_master` row is byte-identical to its value
  before, and the `items` table's rows are byte-identical.

- **AC-022** (maps REQ-FR-022, REQ-FR-023) Given a fixture queue where `q1` is `queued`, `q2` is
  `dropped`, `q3` does not exist, and `q4` is `picked`, When `assign` is run for each, Then `q1`, `q2`, and
  `q3` are refused with an error naming the queue state (or its absence) and create no factory record,
  and `q4` is recorded as `picked`.

- **AC-023** (maps REQ-FR-023, REQ-FR-024, REQ-FR-001) Given a run with one card in each of `picked`,
  `leased` (expired lease), and `needs-decision`, and a fourth card assigned with
  `--contract-ref SPEC-EXAMPLE-001,<64-hex>,2026-09-26T09:00:00Z,<64-hex>`, When `moai factory status`
  and `moai factory status --json` run, Then the text output lists all four card ids with their state,
  stage, version, owner, pending gate, hints, and contract pointer, marks the expired lease `expired`,
  and the JSON parses with one object per card carrying those fields (the fourth card's pointer carries
  exactly the four values); the factory database file's SHA-256 and the `cards` / `events` row contents
  are identical before and after both runs (the expired card still reads `leased`); and a
  `--contract-ref` whose digest or event locator is not 64 hexadecimal characters is refused.

- **AC-024** (maps REQ-FR-025) Given a fixture factory run and queue, and a characterization test
  written before the change that records the queue runtime row and the command output of the queue
  dispatch path (`moai gtd` dispatch, `gtd.go`), When that path assigns card `d1` to lane `worker-2`,
  Then factory.db holds `d1` in `assigned` with owner `worker-2`, and the queue runtime row and the
  command output equal the characterized ones; When the factory-record write for card `d3` is made to
  fail by a test seam that injects a write error (not by file permissions, which the store re-applies on
  open), Then the dispatch still succeeds, the queue runtime row is still written, stderr carries a
  `FACTORY_RECORD_UNAVAILABLE` line, the unavailable-record log under the factory state directory holds
  one entry naming `d3`, `worker-2`, the run, and the error, and `moai factory status` (text and JSON)
  reports that unreconciled entry; When the seam is removed and a further factory-record write
  succeeds for the same run, Then exactly one `record.drift` event names `d3`, `worker-2`, and the factory
  record's state for `d3` (absent), the log entry reads reconciled, `status` no longer reports it, and a
  second successful write appends no further `record.drift` event for `d3`.

- **AC-025** (maps REQ-FR-025) Given a fixture auto mission whose next operation is a dispatch of card
  `d2` to lane `worker-3` (the `goal.go` owner-adapter path), and a characterization test of that path
  written before the change, When the operation executes, Then factory.db holds `d2` in `assigned` with
  owner `worker-3`, the queue runtime row and the mission operation receipt equal the characterized ones;
  When the factory-record write for a second dispatched card `d4` is made to fail by the same injected
  write error, Then the operation still completes, the queue runtime row is still written, stderr
  carries a `FACTORY_RECORD_UNAVAILABLE` line, the unavailable-record log holds an entry naming `d4`
  and `worker-3`, and the next successful write for that run appends one `record.drift` event naming
  `d4` and marks the entry reconciled.

## §D.1 Mutation criteria

- Removing the `version = expected` clause makes AC-004 and AC-006 fail.
- Accepting a legacy-state row's non-abandon transition makes AC-002 fail.
- Adding or dropping any transition-table edge, or ignoring a T4 / T22 guard, changes AC-005's 65 count.
- Accepting any verdict token as passing makes AC-008 case (c) fail; ignoring `audited_sha` makes case
  (d) fail.
- Dropping the tree-equality check makes AC-009 case (b) fail.
- Keeping the lease on `kickoff` makes AC-013 fail (a `lease.expired` event appears and approval is
  refused).
- Accepting `pushed → ci-green` on a supplied file makes AC-019 fail.
- Removing the producer lines from any one of the six files makes AC-020 fail.
- Letting `status` apply expiry makes AC-023 fail (the file hash changes).
- Returning `merging` to `assigned` on expiry makes AC-012 fail.
- Writing a queue item row from `assign` makes AC-021 fail.
- Wiring only one dispatch path makes AC-024 or AC-025 fail.
- Reporting a factory-record failure only on stderr (no log entry), or never emitting `record.drift`
  on the next successful write, makes AC-024 and AC-025 fail; emitting it on every write makes AC-024's
  "no further event" clause fail.
- Appending any line after the `AUDIT-VERDICT:` instruction in an auditor's final message makes
  AC-020 clause (ii) fail.

## §D.2 Severity

All criteria except AC-021 are release-blocking. AC-021 is a regression guard.

## §D.3 Traceability

| REQ | AC |
|-----|----|
| REQ-FR-001 | AC-001, AC-023 |
| REQ-FR-002 | AC-001 |
| REQ-FR-003 | AC-002 |
| REQ-FR-004 | AC-005 |
| REQ-FR-005 | AC-003, AC-004, AC-006 |
| REQ-FR-006 | AC-019 |
| REQ-FR-007 | AC-007 |
| REQ-FR-008 | AC-008, AC-020 |
| REQ-FR-009 | AC-009 |
| REQ-FR-010 | AC-007, AC-008, AC-009 |
| REQ-FR-011 | AC-020 |
| REQ-FR-012 | AC-010 |
| REQ-FR-013 | AC-011, AC-013 |
| REQ-FR-014 | AC-012 |
| REQ-FR-015 | AC-017 |
| REQ-FR-016 | AC-014 |
| REQ-FR-017 | AC-014 |
| REQ-FR-018 | AC-013, AC-016 |
| REQ-FR-019 | AC-015 |
| REQ-FR-020 | AC-016, AC-017 |
| REQ-FR-021 | AC-018 |
| REQ-FR-022 | AC-021, AC-022 |
| REQ-FR-023 | AC-014, AC-015, AC-022, AC-023 |
| REQ-FR-024 | AC-023 |
| REQ-FR-025 | AC-024, AC-025 |

## §D.4 Definition of Done

- Every release-blocking AC shows a `--- PASS:` line (or, for AC-020's grep and emitter parts, the
  command output) in the run-phase evidence, with a non-zero swept count, measured on the run-phase HEAD.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test -cover ./internal/homestate/...` reports ≥ 85%.
- `golangci-lint run` reports no NEW issue in touched packages.
- `make agents-emit-check` exits 0.
- AC-021 still passes.
- `plan.md` §C decisions confirmed or overridden by the lead before Implementation Kickoff Approval.
