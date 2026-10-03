# SPEC-MERGE-WINDOW-QUEUE-001 — Acceptance

> Verification layer: Given-When-Then, binary-testable. Requirements live in spec.md §C. Tests use
> `t.TempDir()` project roots, an injected clock, and injected process-liveness and runner seams;
> no real remote is pushed. One AC per REQ, same number; an AC may carry several scenarios.

## §D AC Matrix

### Queue

- **AC-MWQ-001** (maps REQ-MWQ-001) — Given a record in today's shape (holder fields only), When it
  is read and then acquired/released by the new code, Then it reads as an empty queue, the holder
  fields round-trip unchanged, and no queue key is written unless a ticket exists; and Given a
  ticket written by the new code, Then it carries session, name, card, enqueue instant,
  owning-session pid with `pid_source: session-owner`, `branch`, `branch_source`, `worktree` (equal
  to what a direct `acquire` with the same arguments would record), waiter pid, waiter start time,
  and heartbeat instant.
- **AC-MWQ-002** (maps REQ-MWQ-002) — Scenario 1: Given A holds, When B then C run
  `acquire --wait` (B strictly first), Then status lists B at 1 and C at 2 and both block. Scenario
  2: Given N concurrent `acquire --wait` callers released by a barrier (cross-process, repeated with
  `-count`), Then the queue holds N distinct tickets, no duplicate, no loss. Scenario 3: Given a
  bare `--wait`, Then the bound is 60m from enqueue (injected clock). Scenario 4: Given a blocking
  waiter, When the clock advances 15 s, Then the ticket's heartbeat instant advances.
- **AC-MWQ-003** (maps REQ-MWQ-003) — Scenario 1: Given B's waiter process is killed (SIGKILL; owning
  session still live), When C enqueues, Then B's ticket is dropped and C's output names B with
  reason "waiter gone". Scenario 2: Given a live process with B's waiter pid but a different start
  time, Then B is dropped. Scenario 3: Given B's last heartbeat at −59 s, Then B stays; at −61 s,
  Then B is dropped. Scenario 4: Given B's owning session is gone while its waiter is alive, Then B
  is dropped with reason "owner gone". Scenario 5: Given B's ticket was dropped for a stale
  heartbeat while B's waiter still runs, When B's waiter next polls, Then it exits non-zero naming
  that reason and the queue holds no new ticket for B.
- **AC-MWQ-004** (maps REQ-MWQ-004) — Given A holds indefinitely, When B runs `acquire --wait=2m`
  (injected clock), Then at 1m59s B still waits and at 2m B exits non-zero naming A, its last
  position, and the bound, and B's ticket is gone.
- **AC-MWQ-005** (maps REQ-MWQ-005) — Scenario 1: Given the test hook orders "promote B" before
  "B bound elapsed" in the mutation, Then B ends as holder and, observing its elapsed bound, releases
  immediately (C promoted) and exits non-zero saying it released. Scenario 2: Given the reverse
  order, Then B's ticket is withdrawn and C is promoted; B is never recorded as holder.
- **AC-MWQ-006** (maps REQ-MWQ-006) — Scenario 1: Given A holds, B and C wait, When A releases, Then
  B is holder, queue [C], and B's acquire returns success. Scenario 2 (owner-pid anchor): Given B
  was promoted and its waiter process has exited, Then the holder record carries B's owning-session
  pid with `pid_source: session-owner`, B reads as live (not stale), a mutation by C (enqueue or
  heartbeat) does not displace B, and B's own `release` succeeds. Scenario 3: Given A's owning
  process is gone and B queued, When C runs `status`, Then B is promoted and A is recorded as
  displaced. Scenario 4: Given A's lease expired with a live owner and B queued, When B's waiter
  polls, Then B is promoted and A is recorded as displaced. Scenario 5 (target copy — fixture): Given
  B enqueued with `--card t0002` while the configured git-flow develop branch is `develop`, When B is
  promoted inside A's `release`, Then the holder record's `branch` is `develop`, `branch_source` is
  `config`, and `worktree` equals the ticket's; and the factory-complete window phase run as B
  treats the window as its own (the `lock.Branch != ""` held-by-us path) rather than resolving a new
  branch.
- **AC-MWQ-007** (maps REQ-MWQ-007) — Given policy `hold`, A holding and B queued, When A releases,
  Then the window has no holder and the queue is [B]; Given a stale holder under hold, When any
  mutation runs, Then the holder is cleared with no successor; When the leader runs `policy open`,
  Then B is holder after that same mutation.
- **AC-MWQ-008** (maps REQ-MWQ-008) — Given the shipped default and an injected clock, When A
  acquires, Then the lease expiry is +30m; When A runs `status` at +20m, Then it moves to +50m; and
  Given duration 0, Then a holder with a live owner stays past any age. M0 evidence: the file
  `.moai/reports/t1479/m0-window-duration.md` exists, committed before the M1 code commit, naming
  its command, N, median, and maximum (shorten-only rule checked against it).
- **AC-MWQ-009** (maps REQ-MWQ-009) — Given a holder with lease, policy `hold:release-cut`, and two
  tickets, When `status` and `status --json` run, Then both show holder, lease expiry, policy and
  reason, and both tickets with position and liveness; and Given a ticket whose waiter is gone,
  When `status` runs, Then its output names the dropped ticket and the printed queue omits it.
- **AC-MWQ-010** (maps REQ-MWQ-010) — Given the fixture record
  `.moai/reports/t1479/baseline-acquire-nowait/record.json` (live holder pid 1), When `acquire`
  runs without `--wait` in both human and `--json` forms, Then each exits 1 with empty stdout, the
  stderr equals that directory's `human.stderr` / `json.stderr` after the README's two
  normalizations, and the record is unchanged with an empty queue.
- **AC-MWQ-011** (maps REQ-MWQ-011) — Given B queued behind A, When B invokes `acquire --wait` again
  while its waiter is alive, Then exactly one ticket for B remains at its position; Given C after B,
  Then C is never promoted while B is live; Given policy `hold`, A released and B queued, When a
  caller runs `acquire` without `--wait`, Then it refuses and B stays first; Given `--force` with a
  non-empty queue, Then the forcer holds, the queue order is preserved, and the displacement is
  recorded.

### Policy and doctrine

- **AC-MWQ-012** (maps REQ-MWQ-012) — Given no policy record, Then `policy` reads `open`; Given
  `policy hold --reason release-cut`, When B runs `acquire`, Then it refuses naming `release-cut`;
  When B runs `acquire --wait`, Then B is queued; Given `MOAI_FACTORY_ROLE` set to the lane value,
  When that session runs `policy hold`, Then it refuses and the policy record is unchanged.
- **AC-MWQ-013** (maps REQ-MWQ-013) — For F in `AGENTS.local.md` and
  `.claude/rules/local/gitflow-lane-protocol.md`, each command prints the stated count:
  `grep -c "대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다" F` → `1`;
  `grep -c "moai integration merge --card 또는 그것을 부르는 moai factory complete 로만 병합한다" F` → `1`;
  `grep -c "self-dispatch lane 예외 — 병합 창.*moai integration merge" F` → `1` (the self-dispatch
  clause names the verb complete calls; RED on the plan tree: `0` in both files);
  `grep -c "지명만이 근거" F` → `0`;
  `grep -c "리더 공지가 여전히 첫 번째 층" F` → `0`.
  (RED on the plan tree: `지명만이 근거` → `1` in AGENTS.local.md, `리더 공지가 여전히 첫 번째 층` → `1`
  in the protocol file — research.md §R1 E7 and the iteration-2 audit evidence.)

### Re-measure and merge verb

- **AC-MWQ-014** (maps REQ-MWQ-014) — Given a card worktree that absorbed develop at X, When the
  re-measure verb runs, Then the record's `tree` equals `git rev-parse HEAD^{tree}`, `base` equals
  X, and the build identity is present; Given `workflow.candidate_ci.enabled: true`, Then a valid
  local-form record is rejected and a record carrying a green candidate run id is accepted; Given
  the key false or absent, Then the reverse holds; both decisions come from one verifier.
- **AC-MWQ-015** (maps REQ-MWQ-015) — Each row below is a scenario; "valid" means the verifier
  accepts the record:

  | Command | Exit | Expected |
  |---|---|---|
  | `go test -json ./pkg/...` reporting 4 passes | 0 | valid, count 4 |
  | `go test -json -run '^NONE$' ./pkg/...` (zero tests reported) | 0 | invalid |
  | `go test ./pkg/...` (no `-json`) | 0 | invalid (structured output supported, not requested) |
  | output containing `[no test files]` | 0 | invalid |
  | `true` (no recognized report) | 0 | valid, command + exit only |
  | any command | 3 | invalid; exit 3 recorded as observed |
- **AC-MWQ-016** (maps REQ-MWQ-016) — Given an untracked file in the card worktree, When the verb
  starts, Then it refuses and writes no record; Given a command that modifies a tracked file, When
  the run ends, Then no record is written; Given a command that commits, Then no record is written
  (HEAD changed).
- **AC-MWQ-017** (maps REQ-MWQ-017) — Scenario 1: Given lane B holds the window, a valid record for
  card t0002 and an unmoved base, When `moai integration merge --card t0002` runs, Then it resolves
  the card's `WT-` branch, creates one `--no-ff` merge commit on the integration branch whose second
  parent is the pinned SHA and whose tree equals the record's tree, releases the window, and the
  instrumented runner records zero test invocations. Scenario 2 (pinning): Given the test hook
  advances the `WT-` branch by one commit after the identity check, Then the merge commit's second
  parent is still the pinned SHA and the recorded git invocation is `merge --no-ff <sha>`, never the
  branch name. Scenario 3: Given `workflow.candidate_ci.enabled` false, Then the injected
  landing-check seam records zero calls; Given it true, Then it records exactly one call, for the
  pinned SHA, before the integration branch ref moves. Scenario 4 (non-holder): Given C does not hold
  the window, When C runs the verb, Then it refuses and the window record's bytes are unchanged.
- **AC-MWQ-018** (maps REQ-MWQ-018) — Each row is a scenario run as holder B with C queued; in every
  row no merge commit remains on the integration branch, the window is released, C is promoted (or,
  in the last row, the policy is `hold`), and the exit code is distinct from every other row's:

  | Cause | Setup | Extra expectation |
  |---|---|---|
  | base moved | develop advanced X → Y after the record | message names X and Y; B re-acquiring after re-measure is at the tail |
  | tree mismatch | pinned tree ≠ record tree | — |
  | landing refusal | landing-check seam refuses (key true) | integration tip unchanged |
  | merge failure | conflicting change on develop and a matching record | `git merge --abort` ran; `git status --porcelain` in the integration worktree is empty; no `MERGE_HEAD` |
  | other error | merge-step runner returns an unexpected error | — |
  | abort leaves dirt | abort seam leaves an untracked file | policy is `hold` with a reason naming the worktree; C is not promoted |

### Completion gate

- **AC-MWQ-019** (maps REQ-MWQ-019) — Scenario 1: Given a card with no record, When `moai factory
  complete <card>` runs, Then it refuses, the card state version is unchanged, and the integration
  tip is unchanged. Scenario 2 (one merge path): Given a valid record and an unmoved base, When
  complete runs, Then the instrumented merge-step seam records exactly one call, no other merge
  invocation occurs, and the card is merged-local. Scenario 3 (moved base): Given develop advanced
  past the record's base, When complete runs, Then no merge commit is created, the card state
  version is unchanged, and the exit code is the re-measure-and-re-acquire code. Scenario 4 (verb
  then complete): Given the lane merged card t0002 through `moai integration merge` and develop then
  advanced further, When complete runs, Then the merge-step seam records zero calls, no re-measure
  is required, and the card is merged-local with the existing merge commit's SHA.

- **AC-MWQ-020** (maps REQ-MWQ-020) — Given complete wrote `merge-record.txt`, When that file is
  offered as the re-measure, Then the gate rejects it.
- **AC-MWQ-021** (maps REQ-MWQ-021) — Given a merge whose tree equals a valid record's tree, Then
  the reader accepts; Given a file that only contains the merge SHA prefix, Then it rejects (RED on
  the plan tree, research.md §R1 E5); Given a card without a record, When `moai factory merge
  ready` runs, Then the verdict names the record condition failing next to the three existing
  conditions and prints the recorded command for a card that has one.

### Distribution and compatibility

- **AC-MWQ-022** (maps REQ-MWQ-022) — After `make build`, for F in
  `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md` and
  `.claude/rules/moai/workflow/kanban-dispatch-mechanics.md`: `grep -ci "announc" F` → `0` (RED on
  the plan tree: `1`, line 120, research.md §R1 E8); `grep -c "moai integration acquire --wait" F`
  → at least `1`; `grep -c "moai integration policy" F` → at least `1`;
  `grep -c "moai integration merge" F` → at least `1`; and the template neutrality guard passes.
- **AC-MWQ-023** (maps REQ-MWQ-023) — Given holder-only records (held / free / stale), When the
  integration-lock guard evaluates a `git merge` in its existing test table, Then every decision
  equals the baseline.

## §E Edge cases

- Holder re-acquires (refresh) while the queue is non-empty → holder unchanged, lease renewed, queue
  unchanged.
- Unreadable record → hard error as today, never treated as free.
- A waiter exits normally without promotion (its Bash call ended) → ticket dropped at the next
  mutation; a later `acquire --wait` enqueues at the tail.

## §F Traceability

One AC per REQ with the same number (AC-MWQ-001 … AC-MWQ-023 ↔ REQ-MWQ-001 … REQ-MWQ-023).

| REQ group | REQs | ACs |
|---|---|---|
| Queue | 001-011 | 001-011 |
| Policy and doctrine | 012-013 | 012-013 |
| Re-measure and merge verb | 014-018 | 014-018 |
| Completion gate | 019-021 | 019-021 |
| Distribution / compat | 022-023 | 022-023 |

## §G Quality gates / Definition of Done

- All ACs GREEN on the run HEAD; evidence in progress.md §E.2 with command + verbatim output.
- `go test -race` on `internal/kanban` and touched `internal/cli` tests; AC-MWQ-002 scenario 2
  repeated with `-count`.
- `go vet` + CI-version `golangci-lint` on touched packages; `make build` clean.
- spec-lint clean for this SPEC.
- Sync phase closes before the merge into develop (AGENTS.local.md §4.1).
