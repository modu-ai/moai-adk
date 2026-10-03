# SPEC-MERGE-WINDOW-QUEUE-001 — Acceptance

> Verification layer: Given-When-Then, binary-testable. Requirements live in spec.md §C. Tests use
> `t.TempDir()` project roots and injected liveness / gh runners; no real remote is pushed.

## §D AC Matrix

### Queue (REQ-MWQ-001 … -009)

- **AC-MWQ-001** (maps REQ-MWQ-001) — Given a window record file written in today's shape (holder fields only), When
  it is read and then acquired/released by the new code, Then it reads as an empty queue, the
  holder fields round-trip unchanged, and no new key is written unless a ticket or lease exists.
- **AC-MWQ-002** (maps REQ-MWQ-002) — Given session A holds the window, When sessions B then C invoke
  `acquire --wait` (B strictly first), Then status lists B at position 1 and C at position 2, and
  both commands are still blocking.
- **AC-MWQ-002a** (maps REQ-MWQ-002) — Given N concurrent `acquire --wait` callers released by a barrier against a
  held window (cross-process, `-count` repeated), When the queue is read, Then it holds N distinct
  tickets with no duplicate and no loss.
- **AC-MWQ-003** (maps REQ-MWQ-003) — Given A holds and B, C are queued, When A releases, Then B is the holder, the
  queue is [C], and B's blocked acquire returns success.
- **AC-MWQ-003a** (maps REQ-MWQ-003) — Given A holds and dead-owner ticket D precedes live ticket B, When A releases,
  Then B is the holder, D is removed, and the release output names D as dropped.
- **AC-MWQ-004** (maps REQ-MWQ-004) — Given the holder's owning process is gone and B is queued, When C runs
  `status` (or any acquire), Then B is promoted and the displaced holder is recorded.
- **AC-MWQ-005** (maps REQ-MWQ-005) — Given A holds indefinitely, When B runs `acquire --wait=<short bound>`, Then
  B exits non-zero after the bound, the message names A, B's last position, and the bound, and B's
  ticket is no longer in the queue.
- **AC-MWQ-006** (maps REQ-MWQ-006) — Given a holder with lease, a hold policy, and two tickets, When `status` and
  `status --json` run, Then both show holder, lease expiry, policy with reason, and both tickets in
  order with positions and liveness.
- **AC-MWQ-007** (maps REQ-MWQ-007) — Given a held window, When `acquire` runs without `--wait`, Then stdout, stderr,
  exit code, and the record are byte-identical to the pre-change baseline captured on the
  plan tree (research.md §R1 tree), and the queue is empty.
- **AC-MWQ-008** (maps REQ-MWQ-008) — Given B is queued behind A, When B invokes `acquire --wait` again, Then the
  queue still holds exactly one ticket for B at its original position; and Given C is queued after
  B, When any promotion occurs, Then C is never promoted while B is live and queued.
- **AC-MWQ-009** (maps REQ-MWQ-009) — Given a configured non-zero lease and an expired holder lease with a live owner
  process, When B (queued) polls, Then B is promoted and the displaced holder is recorded; and
  Given lease duration zero, When the same record ages past any duration, Then the holder stays.

### Policy (REQ-MWQ-010 … -013)

- **AC-MWQ-010** (maps REQ-MWQ-010) — Given no policy record, When `moai integration policy` reads, Then it reports
  `open`.
- **AC-MWQ-011** (maps REQ-MWQ-011) — Given policy `hold:release-cut` and a free window, When B runs `acquire`, Then
  it refuses naming `release-cut`; When B runs `acquire --wait`, Then B is queued and not promoted;
  When the leader sets `open`, Then B is promoted on its next poll; and Given A held before the
  hold, Then A remains holder.
- **AC-MWQ-012** (maps REQ-MWQ-012) — Given `MOAI_FACTORY_ROLE` set to the lane value, When the session runs
  `policy hold`, Then it refuses and the policy record is unchanged; Given no role value, Then the
  write succeeds.
- **AC-MWQ-013** (maps REQ-MWQ-013) — Given the edited doctrine, When
  `grep -n "지명만이 근거" AGENTS.local.md` and `grep -n "리더 공지가 여전히 첫 번째 층"
  .claude/rules/local/gitflow-lane-protocol.md` run, Then both return no match, and a grep for the
  promotion-as-authorization sentence returns one match in each file.

### Re-measure (REQ-MWQ-020 … -023)

- **AC-MWQ-020** (maps REQ-MWQ-020) — Given a card worktree that absorbed develop at SHA X, When the re-measure verb
  runs a command, Then the record's `tree` equals `git rev-parse HEAD^{tree}` and `base` equals X.
- **AC-MWQ-021** (maps REQ-MWQ-021) — Given a command exiting 3, When the verb runs it, Then the record carries exit
  code 3 as observed (not a caller argument); and Given a record with test count 0 and no CI run
  id, Then it is classified invalid.
- **AC-MWQ-022** (maps REQ-MWQ-022) — Given a valid record and an unmoved base, When the in-window merge path runs,
  Then no test command is executed inside the window (instrumented runner records zero test
  invocations between acquire and release) and the merge tree equals the record tree.
- **AC-MWQ-023** (maps REQ-MWQ-023) — Given a valid record with base X and develop advanced to Y, When the in-window
  merge path runs, Then no merge commit is created, the window is released and the next ticket
  promoted, and the message names X and Y.

### Completion gate (REQ-MWQ-030 … -033)

- **AC-MWQ-030** (maps REQ-MWQ-030) — Given a card with no re-measure record, When `moai factory complete <card>`
  runs, Then it refuses, the card state version is unchanged, and the integration branch tip is
  unchanged.
- **AC-MWQ-031** (maps REQ-MWQ-031) — Given complete previously wrote `merge-record.txt`, When that file is offered
  as the re-measure, Then the gate rejects it.
- **AC-MWQ-032** (maps REQ-MWQ-032) — Given a merge whose tree equals the record tree with exit 0 and test count 5,
  Then the reader accepts; Given a file that merely contains the merge SHA prefix, Then the reader
  rejects (this is RED on the plan tree — research.md §R1 E5).
- **AC-MWQ-033** (maps REQ-MWQ-033) — Given a card without a record, When `moai factory merge ready` runs, Then the
  verdict names the re-measure condition as failing alongside the three existing conditions.

### Push (REQ-MWQ-040 … -044)

- **AC-MWQ-040** (maps REQ-MWQ-040) — Given threshold 20 and 7 unpushed commits, When `moai integration push` runs,
  Then no push is attempted and the output names 7 and 20; Given threshold 0 and 1 commit, Then
  the count condition passes.
- **AC-MWQ-041** (maps REQ-MWQ-041) — Given the injected CI reader reports `failure` for the remote tip, Then push
  refuses; Given no completed run, Then push proceeds.
- **AC-MWQ-042** (maps REQ-MWQ-042) — Given the injected gh runner errors, Then push refuses and names the fault.
- **AC-MWQ-043** (maps REQ-MWQ-043) — Given lane role, or a held window, Then push refuses; and the push invocation
  recorded by the injected git runner never contains `--force` or `-f`.
- **AC-MWQ-044** (maps REQ-MWQ-044) — Given all conditions pass, When push runs against a local bare remote in
  `t.TempDir()`, Then exactly one push occurs, a fetch follows, and the output reports remote tip
  == pushed tip.

### Doctrine and compatibility (REQ-MWQ-050, -051)

- **AC-MWQ-050** (maps REQ-MWQ-050) — Given `make build` after the template edit, When
  `grep -n "announcement to the lead" internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md .claude/rules/moai/workflow/kanban-dispatch-mechanics.md`
  runs, Then no match (RED on plan tree, §R1 E8), and the template neutrality guard passes.
- **AC-MWQ-051** (maps REQ-MWQ-051) — Given holder-only records (held/free/stale), When the integration-lock guard
  evaluates a `git merge` in the existing test table, Then every decision equals the baseline.

## §E Edge cases

- Waiter killed after promotion → next observer promotes onward (covered by AC-MWQ-004 shape).
- Holder re-acquires (refresh) while queue non-empty → holder unchanged, queue unchanged.
- `--force` with a queue → forcer becomes holder; queue order preserved; displacement recorded.
- Unreadable record → hard error as today, never treated as free.

## §F Traceability

Each REQ-MWQ-NNN maps to AC-MWQ-NNN with the same number; REQ-MWQ-002 additionally maps to
AC-MWQ-002a and REQ-MWQ-003 to AC-MWQ-003a.

| REQ group | REQs | ACs |
|---|---|---|
| Queue | 001-009 | 001-009, 002a, 003a |
| Policy | 010-013 | 010-013 |
| Re-measure | 020-023 | 020-023 |
| Completion gate | 030-033 | 030-033 |
| Push | 040-044 | 040-044 |
| Doctrine / compat | 050-051 | 050-051 |

## §G Quality gates / Definition of Done

- All ACs above GREEN on the run HEAD; evidence in progress.md §E.2 with command + verbatim output.
- `go test -race` on `internal/kanban` and the touched `internal/cli` tests, AC-MWQ-002a repeated
  with `-count`.
- `go vet` + CI-version `golangci-lint` on touched packages; `make build` clean.
- spec-lint clean for this SPEC.
- Sync phase closes before the merge into develop (AGENTS.local.md §4.1).
