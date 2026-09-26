# acceptance.md — SPEC-FACTORY-RECORD-001 (card t1239)

## §A. Scope of verification

The criteria verify the F1 record layer: schema migration, the transition API and its concurrency
control, the three evidence readers, the lease, the hint gates, the three `moai factory` commands,
the queue boundary, and the mirror write on the two existing dispatch paths. F2 and F3 behaviour is
not verified here.

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
red because: no transition API and no transition event kind exist.

R-07
command: grep -n "OpenFactory" internal/cli/gtd.go
stdout:  (empty)
exit:    1
red because: the queue dispatch path writes only the queue runtime report, never factory.db.

R-08
command: grep -rln "hint_after" internal
stdout:  (empty)
exit:    1
red because: no hint, decision, SHA, or contract-reference column exists.
```

Positive control for the empty greps: R-01 and R-03 use the same tool against the same files and
print a match, so an empty result in R-02 / R-04-R-08 is an absence in this tree, not a broken probe.

### C.2 AC matrix

| AC | REQ | Milestone | Class | RED cell | Green path |
|----|-----|-----------|-------|----------|------------|
| AC-001 | REQ-FR-002 | M1 | release-blocking | R-01 | M1 migration; test reads `schema_version=4` and every pre-seeded row intact |
| AC-002 | REQ-FR-001 | M1 | release-blocking | R-01, R-08 | M1 DDL; `PRAGMA table_info(cards)` lists every F1 column |
| AC-003 | REQ-FR-003 | M1 | release-blocking | R-06 | M2 legacy-state refusal |
| AC-004 | REQ-FR-005 | M2 | release-blocking | R-03, R-06 | M2 transaction; version+1 and one event |
| AC-005 | REQ-FR-006 | M2 | release-blocking | R-03 | M2 version compare |
| AC-006 | REQ-FR-004, REQ-FR-007 | M2 | release-blocking | R-06 | M2 table; every non-edge refused |
| AC-007 | REQ-FR-005 | M2 | release-blocking | R-03 | M2; exactly one of two racing writers wins |
| AC-008 | REQ-FR-008, REQ-FR-011 | M3 | release-blocking | R-06 | M3 E-ENTRY |
| AC-009 | REQ-FR-009, REQ-FR-011 | M3 | release-blocking | R-06 | M3 E-VERDICT |
| AC-010 | REQ-FR-010, REQ-FR-011 | M3 | release-blocking | R-06 | M3 E-MERGE |
| AC-011 | REQ-FR-012, REQ-FR-013 | M4 | release-blocking | R-02, R-05 | M4 acquire / renew |
| AC-012 | REQ-FR-014 | M4 | release-blocking | R-02 | M4 expiry return |
| AC-013 | REQ-FR-015 | M4 | release-blocking | R-02 | M4 merging → blocked |
| AC-014 | REQ-FR-016, REQ-FR-023 | M5 | release-blocking | R-08 | M5 `assign --after` gate |
| AC-015 | REQ-FR-017 | M5 | release-blocking | R-08 | M5 `prefer` stored and shown |
| AC-016 | REQ-FR-018, REQ-FR-023 | M5 | release-blocking | R-04 | M5 `decide --gate kickoff` |
| AC-017 | REQ-FR-019 | M5 | release-blocking | R-04 | M5 `decide --choice` |
| AC-018 | REQ-FR-020 | M3/M5 | release-blocking | R-04 | M5 `decide --gate push` |
| AC-019 | REQ-FR-021 | M3/M5 | release-blocking | R-04 | M5 no-remote close |
| AC-020 | REQ-FR-022 | M5 | regression-guard | (green today) | must stay green |
| AC-021 | REQ-FR-022, REQ-FR-023 | M5 | release-blocking | R-04 | M5 queue-picked precondition |
| AC-022 | REQ-FR-023, REQ-FR-024 | M5 | release-blocking | R-04 | M5 `status` |
| AC-023 | REQ-FR-025 | M6 | release-blocking | R-07 | M6 mirror write |
| AC-024 | REQ-FR-001 | M1/M5 | release-blocking | R-08 | M1 columns + M5 `assign --contract-ref` |

AC-020 is a regression guard: the queue schema is frozen and its guard is green on `553e224f3`, so it
cannot be adopted as RED; it proves F1 did not break it.

## §D. Acceptance Criteria (Given-When-Then)

- **AC-001** (maps REQ-FR-002) Given a factory database created by the current code at schema version 3 and seeded with
  one row in each of `runs`, `workers`, `cards`, `events`, and `resume_handoffs`, When it is opened by
  the F1 store, Then `meta.schema_version` reads `4`, every seeded row is present with its original
  values, every new `cards` column of an existing row reads `''`, and opening it a second time changes
  nothing (no error, no duplicate column, identical row count).

- **AC-002** (maps REQ-FR-001) Given no factory database, When the F1 store creates one, Then `PRAGMA table_info(cards)`
  lists every column named in `design.md` § Schema, and `meta.schema_version` reads `4`. A database
  whose version reads `5` is refused with an unsupported-version error.

- **AC-003** (maps REQ-FR-003) Given a card row whose `state` is `in_progress` (a value outside the F1 set), When any
  transition other than `abandoned` is requested, Then it is refused with a legacy-state error and the
  row and the `events` count are unchanged; When `decide --choice abandon` is requested, Then the row
  reads `abandoned` with its version incremented.

- **AC-004** (maps REQ-FR-005) Given a card in `plan` at version `v`, When a legal transition with expected version `v`
  and valid evidence is applied, Then the row reads the target state at version `v+1`, exactly one new
  `events` row of kind `card.transition` exists whose payload names the card, `from`, `to`, `v+1`, and
  the actor; and a fault injected after the state update but before the event insert leaves both the
  row and the event count exactly as before (neither is committed).

- **AC-005** (maps REQ-FR-006) Given a card at version `v`, When a transition is requested with expected version `v-1`,
  Then it is refused with a stale-version error, and a byte comparison of the card row and a count of
  `events` before and after show no change.

- **AC-006** (maps REQ-FR-004, REQ-FR-007) Given a card placed in each of the 19 states in turn, When every one of the 19 target
  states is requested from it with valid evidence and the current version, Then exactly the pairs listed
  in `design.md` § Transition Table (plus the lease-expiry edges when the lease is expired) are accepted
  and every other pair is refused with an illegal-transition error; the test enumerates all 361 pairs
  and reports the accepted count, which equals the table's edge count.

- **AC-007** (maps REQ-FR-005) Given one card at version `v` and two store connections in separate goroutines, When both
  request a legal transition with expected version `v` at the same time, Then exactly one succeeds,
  the other returns a stale-version error, the row reads version `v+1`, and exactly one
  `card.transition` event exists. Repeated 50 times under `-race`, with zero double-success runs.

- **AC-008** (maps REQ-FR-008, REQ-FR-011) Given a card in `plan` with a fixture worktree, When entry into `plan-audit` is requested
  with (a) a SHA that does not exist, (b) a SHA that exists but is not an ancestor of the worktree
  HEAD, (c) a valid SHA and an artifact path absent at that commit, Then each is refused with an
  evidence error and the row is unchanged; When requested with a valid ancestor SHA and a path present
  at that commit, Then the card reads `plan-audit` with `evidence_sha` and `evidence_path` recorded.
  The API offers no parameter through which a caller can assert the evidence passed.

- **AC-009** (maps REQ-FR-009, REQ-FR-011) Given a card in `plan-audit` with `evidence_sha = S`, When `kickoff` is requested and the
  verdict file (a) is absent, (b) lacks the `verdict:` line, (c) reads `verdict: FAIL`, (d) reads
  `verdict: PASS` with `audited_sha:` naming a different commit, Then each is refused and the row is
  unchanged; When the file reads `verdict: PASS-WITH-DEBT` and `audited_sha: S`, Then the card reads
  `kickoff` with `decision_gate = kickoff`; and the return edge to `plan` is accepted with the
  `verdict: FAIL` file of case (c).

- **AC-010** (maps REQ-FR-010, REQ-FR-011) Given a card in `merging` and a fixture repository with an integration branch, When
  `merged-local` is requested with (a) a single-parent commit, (b) a merge commit whose tree differs
  from its second parent's tree, (c) a merge commit not reachable from the integration branch, (d) a
  valid merge commit and a re-measure file that does not contain its SHA, Then each is refused and the
  row is unchanged; When requested with a valid merge commit and a re-measure file naming it, Then the
  card reads `merged-local` with `merge_sha`, `merge_tree`, and `remeasure_path` recorded.

- **AC-011** (maps REQ-FR-012, REQ-FR-013) Given a card in `assigned` with owner label `worker-1`, When the lease is requested by a
  label absent from `workers`, or by `worker-2` registered in `workers`, Then both are refused; When
  requested by `worker-1` registered in `workers`, Then the card reads `leased` with holder `worker-1`
  and an expiry equal to the heartbeat time plus the lease duration; When `worker-1` renews, Then the
  expiry advances and `workers.heartbeat_at` for `worker-1` equals the new heartbeat time; When
  `worker-2` renews, Then it is refused and neither table changes.

- **AC-012** (maps REQ-FR-014) Given a card in `run` with `stage = run`, a worktree path, an `evidence_sha`, and a lease
  whose expiry is in the past (injected clock), When any transition on that card is next requested,
  Then the card first reads `assigned` with an empty holder and `stage`, `worktree_path`, and
  `evidence_sha` unchanged, one `lease.expired` event is appended, the requested transition returns a
  lease-expired error, and a hash of the fixture worktree's `git rev-parse HEAD`, branch list, and
  `git status --porcelain` output is identical before and after. A subsequent lease by the owner
  followed by `leased → run` is accepted.

- **AC-013** (maps REQ-FR-015) Given a card in `merging` whose lease has expired, When any transition on it is next
  requested, Then the card reads `blocked` (not `assigned`) and one `lease.expired` event naming the
  interrupted merge is appended.

- **AC-014** (maps REQ-FR-016, REQ-FR-023) Given card `c2` recorded in `picked` with `--after c1`, When `assign c2 --to worker-1` is
  run while (a) `c1` has no factory record, (b) `c1` is in `sync`, Then (a) is refused with an
  unknown-predecessor error and (b) with a predecessor-unmerged error, and `c2` stays `picked`; When
  `c1` reads `merged-local`, Then the same command moves `c2` to `assigned` with owner `worker-1`.

- **AC-015** (maps REQ-FR-017) Given `assign c3 --prefer backend=codex`, When `status --json` is read, Then the card's
  `prefer` field is exactly `backend=codex`; and `assign c3 --to worker-1` (a worker whose backend is
  not codex) succeeds — the hint refuses nothing.

- **AC-016** (maps REQ-FR-018, REQ-FR-023) Given cards `k1`, `k2`, `k3` in `kickoff` and card `r1` in `run`, When
  `moai factory decide k1 k2 r1 --gate kickoff --choice approve` is run, Then `k1` and `k2` read `run`
  with `decider = human`, `r1` is reported as refused (not in `kickoff`) and is unchanged, and `k3` is
  unchanged; When `decide k3 --gate kickoff --choice reject` is run, Then `k3` reads `blocked`; When
  any `decide` is run with `--decider llm`, Then it is refused and nothing changes.

- **AC-017** (maps REQ-FR-019) Given a card moved from `run` to `needs-decision` with a question, When it is read, Then
  `decision_question` holds the question, `decision_resume = run`, and the lease is cleared; When
  `decide --choice resume` is run, Then the card reads `assigned` with `stage = run`; given another
  card in `needs-decision` whose resume state is `merged-local`, `resume` returns it to `merged-local`;
  `--choice block` yields `blocked` and `--choice abandon` yields `abandoned`.

- **AC-018** (maps REQ-FR-020) Given a card in `merged-local` with `merge_sha = M` and a fixture remote whose
  remote-tracking integration ref does not contain `M`, When `decide <card> --gate push` is run, Then it
  is refused and the card stays `merged-local`; after the fixture remote-tracking ref is advanced to
  contain `M`, the same command moves the card to `pushed`. No `git fetch` is run by the command (the
  fixture ref is advanced by the test).

- **AC-019** (maps REQ-FR-021) Given a fixture repository with no remote and a card in `merged-local`, When
  `decide <card> --gate push` is run, Then the card reads `done` and its transition event payload
  contains `no remote — no CI verdict`; given the same card in a repository with a remote, the direct
  `merged-local → done` edge is refused.

- **AC-020** (maps REQ-FR-022) Given the queue schema guard, When `go test ./internal/kanban -run 'TestBacklogDowngrade_PreChangeBinaryStillServes' -count=1`
  runs, Then it prints `--- PASS`; and given a fixture `backlog.db`, When `assign`, `status`, and
  `decide` have each run against it, Then the `sql` column of every `sqlite_master` row is byte-identical
  to its value before, and the `items` table's rows are byte-identical.

- **AC-021** (maps REQ-FR-022, REQ-FR-023) Given a fixture queue where `q1` is `queued`, `q2` is `dropped`, `q3` does not exist, and
  `q4` is `picked`, When `assign` is run for each, Then `q1`, `q2`, and `q3` are refused with an error
  naming the queue state (or its absence) and create no factory record, and `q4` is recorded as
  `picked`.

- **AC-022** (maps REQ-FR-023, REQ-FR-024) Given a run with one card in each of `picked`, `leased` (expired lease), and
  `needs-decision`, When `moai factory status` and `moai factory status --json` run, Then the text output
  lists all three card ids with their state, stage, version, owner, pending gate, and hints, marks the
  expired lease `expired`, and the JSON parses with one object per card carrying those fields; and the
  factory database file's SHA-256 and the `cards` / `events` row contents are identical before and after
  both runs (the expired card still reads `leased`).

- **AC-023** (maps REQ-FR-025) Given a fixture factory run and queue, When the queue dispatch path (`moai gtd` dispatch)
  assigns card `d1` to lane `worker-2`, Then factory.db holds `d1` in `assigned` with owner `worker-2`,
  the queue runtime report holds the same assignment row it holds today (asserted against a
  characterization test written before the change), and the dispatch output is unchanged; When
  factory.db is made unwritable, Then the dispatch still succeeds, the queue runtime row is still
  written, and stderr carries a `FACTORY_RECORD_UNAVAILABLE` line.

- **AC-024** (maps REQ-FR-001) Given `assign c5 --contract-ref SPEC-EXAMPLE-001,<64-hex>,2026-09-26T09:00:00Z`, When
  `status --json` is read, Then the card's contract reference carries exactly those three values; and
  a contract reference with a digest that is not 64 hexadecimal characters is refused.

## §D.1 Mutation criteria

- Removing the `version = expected` clause makes AC-005 and AC-007 fail.
- Accepting any verdict token as passing makes AC-009 case (c) fail; ignoring `audited_sha` makes case
  (d) fail.
- Dropping the tree-equality check makes AC-010 case (b) fail.
- Letting `status` apply expiry makes AC-022 fail (the file hash changes).
- Returning `merging` to `assigned` on expiry makes AC-013 fail.
- Writing a queue item row from `assign` makes AC-020 fail.

## §D.2 Severity

All criteria except AC-020 are release-blocking. AC-020 is a regression guard.

## §D.3 Traceability

| REQ | AC |
|-----|----|
| REQ-FR-001 | 002, 024 |
| REQ-FR-002 | 001 |
| REQ-FR-003 | 003 |
| REQ-FR-004 | 006 |
| REQ-FR-005 | 004, 007 |
| REQ-FR-006 | 005 |
| REQ-FR-007 | 006 |
| REQ-FR-008 | 008 |
| REQ-FR-009 | 009 |
| REQ-FR-010 | 010 |
| REQ-FR-011 | 008, 009, 010 |
| REQ-FR-012 | 011 |
| REQ-FR-013 | 011 |
| REQ-FR-014 | 012 |
| REQ-FR-015 | 013 |
| REQ-FR-016 | 014 |
| REQ-FR-017 | 015 |
| REQ-FR-018 | 016 |
| REQ-FR-019 | 017 |
| REQ-FR-020 | 018 |
| REQ-FR-021 | 019 |
| REQ-FR-022 | 020, 021 |
| REQ-FR-023 | 014, 016, 021, 022 |
| REQ-FR-024 | 022 |
| REQ-FR-025 | 023 |

## §D.4 Definition of Done

- Every release-blocking AC shows a `--- PASS:` line in the run-phase evidence, with a non-zero swept
  count, measured on the run-phase HEAD.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test -cover ./internal/homestate/...` reports ≥ 85%.
- `golangci-lint run` reports no NEW issue in touched packages.
- AC-020 still passes.
- `plan.md` §C decisions confirmed or overridden by the lead before Implementation Kickoff Approval.
