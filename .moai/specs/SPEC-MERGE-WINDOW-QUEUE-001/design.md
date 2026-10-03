# SPEC-MERGE-WINDOW-QUEUE-001 — Design

> Observable shapes only. Function names and internal types are the run phase's choice.

## D1 Window record — additive fields

The existing record (`.moai/state/integration-lock.json` in the primary checkout) gains optional
fields; absent fields read as today:

| Field | Meaning |
|---|---|
| `queue[]` | FIFO tickets: `session_id`, `session_name`, `card`, `enqueued_at`, `pid` + `pid_source: session-owner` (the owning session, resolved as `acquire` resolves it today), `waiter_pid`, `waiter_start` (process start time), `heartbeat_at` |
| `lease_expires_at` | holder lease expiry (on by default, 30 min; zero disables) |

Promotion copies the ticket's `session_id`, `session_name`, `card`, `pid`, and `pid_source` into the
holder fields. The holder therefore keeps the exact semantics `Stale()` and `releasableBy` rely on
(`internal/kanban/integration_lock.go:172-180`, `322-329`): liveness and self-release follow the
owning session, never the waiter process, which exits as soon as `acquire --wait` returns.

The policy (`open` | `hold` + reason + setter + instant) is a sibling record beside the window
record, so the guard's read of the window record is untouched (REQ-MWQ-023).

A published record has **no holder and a non-empty queue** only while the policy is `hold`
(REQ-MWQ-007). Under `open`, every mutation that frees the holder promotes in the same mutation.

## D2 Lifecycle

```
acquire --wait ─► [free, open, empty queue] ─────────────────────► holder
              └► otherwise ─► ticket@tail (waiter heartbeats every 15 s)
                    ├─ promoted (open policy; owner pid stamped on holder) ─► holder
                    ├─ bound elapses ─► withdrawn, exit≠0
                    └─ owner gone / waiter gone / heartbeat > 60 s ─► dropped at next mutation
release / owner dead / lease expired ─► [open] promote first live ticket
                                       └► [hold] no successor; queue intact until `policy open`
```

Timeout vs promotion: both are written inside the mutation; the first written wins. A waiter that
reads "promoted" after its own bound has elapsed releases immediately and reports it.

## D3 Lane sequence

```
lane worktree:   git merge develop (local), commit
                 moai integration remeasure -- <cmd>   (clean tree before+after, HEAD unchanged;
                 record keyed by HEAD^{tree}, base = absorbed develop SHA)
                 moai integration acquire --wait --card <id>   (run in the background)
when holder:     moai integration merge --card <id>
                 ├─ resolve WT- branch (SPEC-CANDIDATE-CI-001 REQ-CCI-004 contract)
                 ├─ record.base == develop tip AND branch tip tree == record.tree ?
                 │   ├─ yes ─► landing check (REQ-CCI-011; no-op when candidate_ci off)
                 │   │         ─► git merge --no-ff ─► merge^{tree} == record.tree ─► release, exit 0
                 │   └─ base moved ─► release (next live ticket promoted)
                 │                    ─► exit re-measure-and-re-acquire code (names both SHAs)
                 │                    ─► lane re-absorbs, re-measures, acquire --wait at the tail
```

The stale candidate is rebuilt only after this exit, never eagerly at queue entry
(SPEC-CANDIDATE-CI-001's assignment to this SPEC).

`moai factory complete` refuses up front without a valid record (REQ-MWQ-019) and performs the same
in-window steps. `merge-record.txt` stays as a merge-identity file and never counts as the
re-measure (REQ-MWQ-020).

## D4 Re-measure record

Written under the card's evidence directory, keyed by the candidate tree SHA. Common fields:
`tree`, `base`, `measured_at`, `measured_by` (moai build identity, per verification-claim-integrity
§2.2). Two forms, read by one verifier; `workflow.candidate_ci.enabled` (SPEC-CANDIDATE-CI-001
REQ-CCI-023) selects which is required:

- local form — `command`, `exit_code` (captured by the verb), `test_count` when a recognized
  structured report exists; invalid on: non-zero exit, reported zero, runner empty-sweep marker,
  structured-capable tool run without structured output;
- candidate-CI form — `ci_run_id`, `ci_verdict` (green per SPEC-CANDIDATE-CI-001 REQ-CCI-009).

Recognized runners and empty-sweep markers form a per-runner table owned by the verb; this
repository's row is `go test -json` (count = test `pass` events; markers `no tests to run`,
`[no test files]`). Rows for other runners are additive and neutral.

## D5 Policy verb

`moai integration policy open` / `moai integration policy hold --reason <text>` / bare
`moai integration policy` (read). Lane-role sessions are refused on the write forms. `policy open`
runs the promotion mutation itself.

## D6 Compatibility matrix

| Caller | Before | After |
|---|---|---|
| `acquire` (no `--wait`) | refuse on live holder | identical to the committed fixture; also refused while a live ticket is queued |
| `release` | clears record | clears holder; promotes under `open`, keeps queue under `hold` |
| `status` | holder only | holder + lease + policy + queue |
| session-end automerge | acquire force=false | identical (no `--wait`) |
| PreToolUse guard | reads holder | identical decision on holder-only records |
| `factory complete` | writes stand-in, substring gate | requires keyed record, structural gate |
| lane merge into develop | manual `git merge --no-ff` in the integration worktree | `moai integration merge --card <id>` |
