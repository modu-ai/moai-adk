# SPEC-MERGE-WINDOW-QUEUE-001 — Design

> Observable shapes only. Function names and internal types are the run phase's choice.

## D1 Window record — additive fields

The existing record (`.moai/state/integration-lock.json` in the primary checkout) gains optional
fields; absent fields read as today:

| Field | Meaning |
|---|---|
| `queue[]` | ordered tickets: `session_id`, `session_name`, `pid` (owning session, `pid_source: session-owner`), `card`, `enqueued_at` |
| `lease_expires_at` | holder lease expiry (only when a non-zero lease duration is configured) |

A record with no holder but a non-empty queue is valid only transiently inside a mutation; every
published record with a non-empty queue has a holder (promotion happens in the same mutation that
frees the holder).

The policy (`open` | `hold` + reason + setter + instant) is a sibling record beside the window
record so that the guard's read of the window record is untouched (REQ-MWQ-051).

## D2 Lifecycle

```
acquire --wait  ─► [free & open] ──────────────► holder
                └► [held or hold] ─► ticket@tail ─► poll ─┬─► promoted ─► holder
                                                         └─► bound elapsed ─► ticket removed, exit≠0
release         ─► holder cleared ─► first live ticket promoted (dead tickets dropped, named)
stale/expired   ─► observed by any acquire/poll/status ─► head promoted, displaced holder recorded
```

Promotion makes the waiter the holder in the record; the waiting process observes it on its next
poll and returns success. A waiter whose process died before observing promotion leaves a held
record whose owner is dead — the next observer treats it as stale and promotes onward (no wedge).

## D3 Lane sequence (new order)

```
lane worktree:   git merge develop (local)  ─►  moai integration remeasure -- <cmd>
                 (record keyed by HEAD^{tree}, base = absorbed develop SHA)
                 moai integration acquire --wait --card <id>
integration wt:  check develop == record.base AND card tip tree == record.tree
                 ├─ equal  ─► git merge --no-ff ─► verify merge^{tree} == record.tree ─► release
                 └─ moved  ─► release (next promoted) ─► report requeue (re-absorb, re-measure)
```

`moai factory complete` follows the same in-window steps and refuses up front without a valid
record (REQ-MWQ-030). The merge identity it records (`merge-record.txt`) stays as a merge-identity
file but is never accepted as the re-measure (REQ-MWQ-031).

## D4 Re-measure record

Written under the card's evidence directory, keyed by the candidate tree SHA. Fields: `tree`,
`base` (absorbed integration-branch commit), `command`, `exit_code`, `test_count` or
`ci_run_id` + `ci_conclusion`, `measured_at`, `measured_by` (build provenance of the moai binary,
per verification-claim-integrity §2.2). The executing verb captures command and exit code; the
test-count source is decision-index Q6.

## D5 Policy verb

`moai integration policy open` / `moai integration policy hold --reason <text>` / bare
`moai integration policy` (read). Lane-role sessions are refused on the write forms.

## D6 Push verb

`moai integration push [--json]`:

1. refuse if lane role; refuse if window held;
2. pre-push final read: local integration tip, integration worktree has no `MERGE_HEAD`;
3. count = `rev-list --count origin/<b>..<b>`; compare with threshold (0 = no condition);
4. CI read of the remote tip via the existing gh conclusion mapping: failure → refuse; no completed
   run → proceed; unreadable → refuse naming the fault;
5. `git push origin <b>` (never `--force`); fetch; report remote tip == pushed tip.

## D7 Compatibility matrix

| Caller | Before | After |
|---|---|---|
| `acquire` (no `--wait`) | refuse on live holder | identical |
| `release` | clears record | clears record, promotes head if any |
| `status` | holder only | holder + lease + policy + queue |
| session-end automerge | acquire force=false | identical (no `--wait`) |
| PreToolUse guard | reads holder | identical decision on holder-only records |
| `factory complete` | writes stand-in, substring gate | requires keyed record, structural gate |
