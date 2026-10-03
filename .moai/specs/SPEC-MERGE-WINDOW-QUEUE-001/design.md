# SPEC-MERGE-WINDOW-QUEUE-001 — Design

> Observable shapes only. Function names and internal types are the run phase's choice.

## D1 Window record — additive fields

The existing record (`.moai/state/integration-lock.json` in the primary checkout) gains optional
fields; absent fields read as today:

| Field | Meaning |
|---|---|
| `queue[]` | ordered tickets: `session_id`, `session_name`, `card`, `enqueued_at`, `waiter_pid`, `waiter_start` (process start time), `heartbeat_at`, `state` (`waiting` / `between-slices` / `reserved`), `reentry_deadline` (between-slices), `ready_deadline` + `front_used` (reserved) |
| `lease_expires_at` | holder lease expiry (on by default, 30 min; zero disables) |

The policy (`open` | `hold` + reason + setter + instant) is a sibling record beside the window
record, so the guard's read of the window record is untouched (REQ-MWQ-025).

A published record may have **no holder and a non-empty queue** in exactly two cases: the policy is
`hold` (REQ-MWQ-008), or every queued ticket is `reserved` and not yet ready (REQ-MWQ-020). In every
other case, a mutation that frees the holder promotes in the same mutation.

## D2 Lifecycle

```
acquire --wait ─► [free, open, no ready ticket] ─────────────────────► holder
              └► otherwise ─► ticket@tail (waiting, heartbeating)
                    ├─ promoted (open policy) ───────────────────────► holder
                    ├─ slice ends ─► between-slices ─(re-invoke ≤ deadline)─► waiting, same position
                    ├─ total bound elapses ─► withdrawn, exit≠0
                    └─ waiter gone / heartbeat stale / deadline passed ─► dropped at next mutation
release / owner dead / lease expired ─► [open] promote first live+ready ticket
                                       └► [hold] no successor; queue intact until `policy open`
```

Timeout vs promotion: both are written inside the mutation; the first written wins. A waiter that
reads "promoted" after its own bound has elapsed releases immediately and reports it.

## D3 Lane sequence (new order)

```
lane worktree:   git merge develop (local), commit
                 moai integration remeasure -- <cmd>   (clean tree before+after, HEAD unchanged;
                 record keyed by HEAD^{tree}, base = absorbed develop SHA)
                 moai integration acquire --wait --card <id>   (background, or --slice loop)
integration wt:  identity check (record.base == develop tip AND card tip tree == record.tree;
                 under candidate CI: REQ-CCI-011 landing check)
                 ├─ equal ─► git merge --no-ff ─► verify merge^{tree} == record.tree ─► release
                 └─ moved ─► release (next ready promoted) ─► reserved ticket ─► re-absorb, re-measure
reserved:        not ready ─(record on current tip)─► ready ─► next promotion once
                 ├─ not ready in 30 min ─► dropped ("readiness bound")
                 └─ base moved again after its front promotion ─► tail, waiting
```

`moai factory complete` follows the same in-window steps and refuses up front without a valid
record (REQ-MWQ-021). `merge-record.txt` stays as a merge-identity file and never counts as the
re-measure (REQ-MWQ-022).

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
repository's row is `go test -json` (count = `pass` events of tests; markers `no tests to run`,
`[no test files]`). Rows for other runners are additive and neutral.

## D5 Policy verb

`moai integration policy open` / `moai integration policy hold --reason <text>` / bare
`moai integration policy` (read). Lane-role sessions are refused on the write forms. `policy open`
runs the promotion mutation itself.

## D6 Compatibility matrix

| Caller | Before | After |
|---|---|---|
| `acquire` (no `--wait`) | refuse on live holder | identical to the committed fixture |
| `release` | clears record | clears holder; promotes under `open`, keeps queue under `hold` |
| `status` | holder only | holder + lease + policy + queue + recorded commands |
| session-end automerge | acquire force=false | identical (no `--wait`) |
| PreToolUse guard | reads holder | identical decision on holder-only records |
| `factory complete` | writes stand-in, substring gate | requires keyed record, structural gate |
