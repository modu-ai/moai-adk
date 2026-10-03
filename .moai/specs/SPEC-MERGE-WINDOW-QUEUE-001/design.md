# SPEC-MERGE-WINDOW-QUEUE-001 — Design

> Observable shapes only. Function names and internal types are the run phase's choice.

## D1 Window record — additive fields

The existing record (`.moai/state/integration-lock.json` in the primary checkout) gains optional
fields; absent fields read as today:

| Field | Meaning |
|---|---|
| `queue[]` | FIFO tickets: `session_id`, `session_name`, `card`, `enqueued_at`, `pid` + `pid_source: session-owner` (the owning session, resolved as `acquire` resolves it today), `branch` + `branch_source` + `worktree` (the integration target, resolved at enqueue as `acquire` resolves it today), `waiter_pid`, `waiter_start` (process start time), `heartbeat_at` |
| `lease_expires_at` | holder lease expiry (on by default, 30 min; zero disables) |

Promotion copies the ticket's `session_id`, `session_name`, `card`, `pid`, `pid_source`, `branch`,
`branch_source`, and `worktree` into the holder fields. The holder therefore keeps the exact
semantics `Stale()` and `releasableBy` rely on (`internal/kanban/integration_lock.go:172-180`,
`322-329`) — liveness and self-release follow the owning session, never the waiter process, which
exits as soon as `acquire --wait` returns — and the target `factory complete` reads
(`lock.Branch != ""` at `factory_card.go:1335`, `BranchSource` at `:1351`).

Every window verb is a queue mutation: `acquire`, `release`, `status`, and `policy` apply the drop
rules and promotion before doing their own work. `merge` is the exception in order only: it reads
the record to decide holdership first, refuses a non-holder or an expired-lease holder with the
record unchanged, and applies drops (and renews its own lease) only after that check passes.

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
moai integration merge --card <id>
  read record ─► not holder, or own lease expired ─► refuse, record unchanged (no drops, no promotion)
  holder ─► renew lease, apply drops
  ├─ resolve WT- branch (SPEC-CANDIDATE-CI-001 REQ-CCI-004 contract); pin SHA = branch tip (read once)
  │   pre-merge checks, in order — each failure: release, promote next, distinct exit
  ├─ (1) record valid (REQ-014/015)?                  ─► no ─► exit RECORD-INVALID
  ├─ (2) record.base == develop tip?                  ─► no ─► exit RE-MEASURE (names both SHAs) ─► lane re-measures, tail
  ├─ (3) record.base is an ancestor of SHA?           ─► no ─► exit ANCESTRY
  ├─ (4) SHA^{tree} == record.tree?                   ─► no ─► exit TREE-MISMATCH
  ├─ (5) landing check(SHA) (no-op when key off)      ─► refused ─► exit LANDING-REFUSED
  ├─ (6/7) git merge --no-ff SHA ─► fails ─► git merge --abort
  │                                      ├─ worktree clean ─► exit MERGE-FAILED (next promoted)
  │                                      └─ still dirty    ─► policy hold(reason, setter=merge step+card) ─► exit MERGE-DIRTY
  ├─ (8) after the merge commit exists: merge^{tree} ≠ record.tree or any lookup error
  │         ─► leave the merge commit, policy hold(reason names cause + merge SHA) ─► release ─► exit POST-MERGE
  ├─ (9) any other error before the merge ─► exit OTHER (next promoted)
  └─ success ─► release, exit 0
```

With checks (2)-(4) passing, `git merge --no-ff SHA` onto a tip equal to `record.base` produces
`SHA^{tree}`, so cause 8's tree mismatch is unreachable by construction; cause 8 still exists for
lookup and I/O errors after the commit, and its `hold` stops the queue on a develop the leader must
inspect. Exit-code names are placeholders; the run phase assigns thirteen distinct values (cause 13: a path the pinned SHA adds already exists in the integration worktree as an ignored or untracked file — refused before `git merge`, bytes untouched; cause 10: pinned SHA equals `record.base`, nothing to merge, refused before `git merge`; cause 11: the merge verb applies complete's card gate and requires the requested card to equal the window card, before develop moves; cause 12: integration worktree dirty before the merge; dirty after the merge is cause 8). The stale
candidate is rebuilt only after the RE-MEASURE exit, never eagerly at queue entry
(SPEC-CANDIDATE-CI-001's assignment to this SPEC).

`moai factory complete` has no merge of its own any more, and every gate precedes any move of
develop:

```
factory complete <card>
  ├─ (1) card gates (today's T14 preconditions): merge-ready, caller holds the card lease,
  │       lease unexpired, version as read ─► any fails ─► refuse; develop and card unchanged
  ├─ (2) adoption: develop holds a merge commit whose 2nd parent == WT- branch CURRENT tip
  │       and whose tree == a VALID record's tree ─► record merged-local from it (no merge step)
  │       (a branch with commits after that merge is not adopted → continue)
  ├─ (3) no valid record for the current candidate tree ─► refuse; develop and card unchanged
  └─ (4) call the merge step above (its release deferred until the transitions below are done)
        ├─ success ─► merging → merged-local transitions ─► release
        │     └─ transition fails (version/lease changed after (1)) ─► leave merge commit,
        │        policy hold(post-merge-transition-conflict, merge SHA) ─► release ─► exit TRANSITION-CONFLICT
        └─ any failure ─► card unchanged, exit with the step's code
```

`merge-record.txt` stays as a merge-identity file and never counts as the re-measure (REQ-MWQ-020).

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
