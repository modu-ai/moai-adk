# SPEC-MERGE-WINDOW-QUEUE-001 — Acceptance

> Verification layer: Given-When-Then, binary-testable. Requirements live in spec.md §C. Tests use
> `t.TempDir()` project roots, an injected clock, and injected process-liveness and runner seams;
> no real remote is pushed. One AC per REQ, same number; an AC may carry several scenarios.

## §D AC Matrix

### Queue

- **AC-MWQ-001** (maps REQ-MWQ-001) — Given a record in today's shape (holder fields only), When it
  is read and then acquired/released by the new code, Then it reads as an empty queue, the holder
  fields round-trip unchanged, and no queue key is written unless a ticket exists; and Given a
  ticket written by the new code, Then it carries session, name, card, enqueue instant, waiter pid,
  waiter start time, heartbeat instant, and state.
- **AC-MWQ-002** (maps REQ-MWQ-002) — Scenario 1: Given A holds, When B then C run
  `acquire --wait` (B strictly first), Then status lists B at 1 and C at 2 and both block. Scenario
  2: Given N concurrent `acquire --wait` callers released by a barrier (cross-process, repeated with
  `-count`), Then the queue holds N distinct tickets, no duplicate, no loss. Scenario 3: Given a
  bare `--wait`, Then the ticket's total bound is 60m from first enqueue (injected clock). Scenario
  4: Given a blocking waiter, When the clock advances one heartbeat interval, Then the ticket's
  heartbeat instant advances.
- **AC-MWQ-003** (maps REQ-MWQ-003) — Given B queued at position 2 with `--slice 5m`, When the slice
  ends, Then B exits with the still-queued code naming position 2 and the ticket is `between-slices`;
  When B re-invokes `acquire --wait --slice 5m` at +119 s, Then the queue still holds exactly one
  ticket for B at position 2 and its state is `waiting`; Given instead a re-invocation at +121 s
  after the drop mutation ran, Then B gets a new ticket at the tail.
- **AC-MWQ-004** (maps REQ-MWQ-004) — Scenario 1 (killed waiter): Given B's waiter process is killed
  (SIGKILL; owning session still live), When C enqueues, Then B's ticket is dropped and C's output
  names B with reason "waiter gone". Scenario 2 (pid reuse): Given a live process with B's pid but a
  different start time, Then B is dropped. Scenario 3: Given B's last heartbeat at −59 s, Then B
  stays; at −61 s, Then B is dropped (60 s window, injected clock).
  Scenario 4: Given a `between-slices` ticket past its re-entry deadline, Then it is dropped.
- **AC-MWQ-005** (maps REQ-MWQ-005) — Given A holds indefinitely, When B runs `acquire --wait=2m`
  (injected clock), Then at 1m59s B still waits and at 2m B exits non-zero naming A, its last
  position, and the bound, and B's ticket is gone.
- **AC-MWQ-006** (maps REQ-MWQ-006) — Scenario 1: Given the test hook orders "promote B" before
  "B bound elapsed" in the mutation, Then B ends as holder and, observing its elapsed bound, releases
  immediately (C promoted) and exits non-zero saying it released. Scenario 2: Given the reverse
  order, Then B's ticket is withdrawn and C is promoted; B is never recorded as holder.
- **AC-MWQ-007** (maps REQ-MWQ-007) — Scenario 1: Given A holds, B and C wait, When A releases, Then
  B is holder, queue [C], B's acquire returns success. Scenario 2: Given A's owning process is gone
  and B queued, When C runs `status`, Then B is promoted and A is recorded as displaced. Scenario 3:
  Given A's lease expired with a live owner and B queued, When B polls, Then B is promoted and A is
  recorded as displaced.
- **AC-MWQ-008** (maps REQ-MWQ-008) — Given policy `hold`, A holding and B queued, When A releases,
  Then the window has no holder and the queue is [B]; When A's replacement observation finds a stale
  holder under hold, Then the holder is cleared with no successor; When the leader runs
  `policy open`, Then B is holder after that same mutation.
- **AC-MWQ-009** (maps REQ-MWQ-009) — Given the shipped default and an injected clock, When A
  acquires, Then the lease expiry is +30m; When A runs `status` at +20m, Then it moves to +50m; and
  Given duration 0, Then a holder with a live owner stays past any age. M0 evidence: the file
  `.moai/reports/t1479/m0-window-duration.md` exists, committed before the M1 code commit, naming
  its command, N, median, and maximum (shorten-only rule checked against it).
- **AC-MWQ-010** (maps REQ-MWQ-010) — Given a holder with lease, policy `hold:release-cut`, and a
  `waiting` and a `reserved` ticket, When `status` and `status --json` run, Then both show holder,
  lease expiry, policy and reason, both tickets with position, state, and liveness, and the recorded
  re-measure command verbatim for the holder and the ready ticket.
- **AC-MWQ-011** (maps REQ-MWQ-011) — Given the fixture record
  `.moai/reports/t1479/baseline-acquire-nowait/record.json` (live holder pid 1), When `acquire`
  runs without `--wait` in both human and `--json` forms, Then each exits 1 with empty stdout, the
  stderr equals that directory's `human.stderr` / `json.stderr` after the README's two
  normalizations, and the record is unchanged with an empty queue.
- **AC-MWQ-012** (maps REQ-MWQ-012) — Given B queued behind A, When B invokes `acquire --wait`
  again, Then exactly one ticket for B remains at its position; Given C after B, Then C is never
  promoted while B is live and ready; Given `--force` with a non-empty queue, Then the forcer holds,
  the queue order is preserved, and the displacement is recorded.

### Policy and doctrine

- **AC-MWQ-013** (maps REQ-MWQ-013) — Given no policy record, Then `policy` reads `open`; Given
  `policy hold --reason release-cut`, When B runs `acquire`, Then it refuses naming `release-cut`;
  When B runs `acquire --wait`, Then B is queued; Given `MOAI_FACTORY_ROLE` set to the lane value,
  When that session runs `policy hold`, Then it refuses and the policy record is unchanged.
- **AC-MWQ-014** (maps REQ-MWQ-014) — When the following run, Then each prints the stated count:
  `grep -c "대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다" AGENTS.local.md` → `1`;
  `grep -c "대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다" .claude/rules/local/gitflow-lane-protocol.md` → `1`;
  `grep -c "지명만이 근거" AGENTS.local.md` → `0`;
  `grep -c "리더 공지가 여전히 첫 번째 층" .claude/rules/local/gitflow-lane-protocol.md` → `0`.
  (RED on the plan tree: research.md §R1 E7 shows line 221 carrying `지명만이 근거`.)

### Re-measure

- **AC-MWQ-015** (maps REQ-MWQ-015) — Given a card worktree that absorbed develop at X, When the
  re-measure verb runs, Then the record's `tree` equals `git rev-parse HEAD^{tree}`, `base` equals
  X, and the build identity is present; Given `workflow.candidate_ci.enabled: true`, Then a valid
  local-form record is rejected and a record carrying a green candidate run id is accepted; Given
  the key false or absent, Then the reverse holds; both decisions come from one verifier.
- **AC-MWQ-016** (maps REQ-MWQ-016) — Each row below is a scenario; "valid" means the verifier
  accepts the record:

  | Command | Exit | Expected |
  |---|---|---|
  | `go test -json ./pkg/...` reporting 4 passes | 0 | valid, count 4 |
  | `go test -json -run '^NONE$' ./pkg/...` (zero tests reported) | 0 | invalid |
  | `go test ./pkg/...` (no `-json`) | 0 | invalid (structured output supported, not requested) |
  | output containing `[no test files]` | 0 | invalid |
  | `true` (no recognized report) | 0 | valid, command + exit only |
  | any command | 3 | invalid; exit 3 recorded as observed |
- **AC-MWQ-017** (maps REQ-MWQ-017) — Given an untracked file in the card worktree, When the verb
  starts, Then it refuses and writes no record; Given a command that modifies a tracked file, When
  the run ends, Then no record is written; Given a command that commits, Then no record is written
  (HEAD changed).
- **AC-MWQ-018** (maps REQ-MWQ-018) — Given a valid record and an unmoved base, When the in-window
  merge path runs, Then the instrumented runner records zero test invocations between acquire and
  release and the merge tree equals the record tree.
- **AC-MWQ-019** (maps REQ-MWQ-019) — Given a record with base X and develop advanced to Y, When the
  in-window merge path runs, Then no merge commit is created, the window is released and the next
  ready ticket promoted, the caller holds a `reserved` ticket, and the message names X and Y.
- **AC-MWQ-020** (maps REQ-MWQ-020) — Scenario 1 (non-blocking): Given A `reserved` and not ready,
  B `waiting`, When the holder releases, Then B is promoted. Scenario 2 (front-once): Given A then
  produces a record on the current tip while C and D wait, When the holder releases, Then A is
  promoted before C and D. Scenario 3 (bound): Given A never re-measures, When 30m pass, Then A is
  dropped with reason "readiness bound". Scenario 4 (second own-re-measure move): Given A was
  requeued once, and develop moves again while A is re-measuring, When A's merge path runs, Then
  A's new ticket is `waiting` at the tail. Scenario 5 (move while waiting does not count): Given A
  is ready at the front while B holds, When B's merge moves develop, Then A keeps its front position,
  its own-re-measure requeue count is unchanged, and after re-measuring A is promoted before every
  other ticket. Scenario 6 (bound): Given A has been requeued three consecutive times of any kind,
  Then A is at the tail and the window record's log carries an entry naming A and the count 3.

### Completion gate

- **AC-MWQ-021** (maps REQ-MWQ-021) — Given a card with no record, When `moai factory complete
  <card>` runs, Then it refuses, the card state version is unchanged, and the integration tip is
  unchanged.
- **AC-MWQ-022** (maps REQ-MWQ-022) — Given complete wrote `merge-record.txt`, When that file is
  offered as the re-measure, Then the gate rejects it.
- **AC-MWQ-023** (maps REQ-MWQ-023) — Given a merge whose tree equals a valid record's tree, Then
  the reader accepts; Given a file that only contains the merge SHA prefix, Then it rejects (RED on
  the plan tree, research.md §R1 E5); Given a card without a record, When `moai factory merge
  ready` runs, Then the verdict names the record condition failing next to the three existing
  conditions and prints the recorded command for a card that has one.

### Doctrine and compatibility

- **AC-MWQ-024** (maps REQ-MWQ-024) — After `make build`, for F in
  `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md` and
  `.claude/rules/moai/workflow/kanban-dispatch-mechanics.md`: `grep -ci "announc" F` → `0` (RED on
  the plan tree: `1`, line 120, research.md §R1 E8); `grep -c "moai integration acquire --wait" F`
  → at least `1`; `grep -c "moai integration policy" F` → at least `1`; and the template neutrality
  guard passes.
- **AC-MWQ-025** (maps REQ-MWQ-025) — Given holder-only records (held / free / stale), When the
  integration-lock guard evaluates a `git merge` in its existing test table, Then every decision
  equals the baseline.

## §E Edge cases

- Holder re-acquires (refresh) while the queue is non-empty → holder unchanged, lease renewed, queue
  unchanged.
- Unreadable record → hard error as today, never treated as free.
- Policy set to `hold` while a waiter is between slices → ticket kept, no promotion.

## §F Traceability

One AC per REQ with the same number (AC-MWQ-001 … AC-MWQ-025 ↔ REQ-MWQ-001 … REQ-MWQ-025).

| REQ group | REQs | ACs |
|---|---|---|
| Queue | 001-012 | 001-012 |
| Policy and doctrine | 013-014 | 013-014 |
| Re-measure | 015-020 | 015-020 |
| Completion gate | 021-023 | 021-023 |
| Doctrine / compat | 024-025 | 024-025 |

## §G Quality gates / Definition of Done

- All ACs GREEN on the run HEAD; evidence in progress.md §E.2 with command + verbatim output.
- `go test -race` on `internal/kanban` and touched `internal/cli` tests; AC-MWQ-002 scenario 2
  repeated with `-count`.
- `go vet` + CI-version `golangci-lint` on touched packages; `make build` clean.
- spec-lint clean for this SPEC.
- Sync phase closes before the merge into develop (AGENTS.local.md §4.1).
