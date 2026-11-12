# Merge train query — #2388 AC-1c

`queue-query PROJECT` reads a leader-selected GitHub Issue comment and projects its
merge train. It never selects or claims an Issue, writes a journal/cache/TODO/SQLite
store, posts a comment, updates a branch or grants integration permission.

```json
{"schema_version":1,"owner":"modu-ai","repo":"moai-adk","base_branch":"v2026.11.0","ledger_issue":1975,"ledger_comment_id":6780607225}
```

Pass the JSON on stdin. The example comment is a historical pinned source, not an
automatic selection of the latest train. The leader chooses the current comment ID.
The result uses the normal `queue.query` protocol envelope. Invalid input, an API
failure, malformed/ambiguous order, changed head/base/state or an edited ledger is
an error; the CLI does not return an apparently complete partial train.

## Projection contract

- Source: exactly one Markdown table whose first two columns are `순서 | PR` or
  `position | PR`. Positions must be 1..N without gaps, with unique positive PR
  numbers. `#2428 (#2393)` is accepted in a PR cell. Code fences are ignored. An
  explicit empty table is an empty queue; missing tables are errors. Other table
  columns are historical prose and never replace current remote observations.
- Order is the source table's order, even when PR numbers are not sorted. The
  result preserves ledger positions and closed entries; only preceding open PRs
  produce `waiting_predecessor`.
- `session` and `ownership` are the linked Issue's existing lane-status claim
  holder/lease/handoff projection, not the PR author or an observation of a pane.
  Expiry remains visible and never causes a takeover. Missing ownership is null.
  Per-AC multi-owner claims are not interpreted as a new ownership protocol.
- `head`, `base`, mergeability, draft and open/closed state come from the PR API.
  `Refs #N` uses the existing leader parser. `Blocked-By:` references reuse
  queue-next's parser and are re-read remotely; open dependencies are listed.
- `verdict` reuses lane-status after filtering to the exact first-line grammar
  `AUDIT #<PR> (<key>) head <40hex> base <40hex> — verdict: PASS|FAIL`.
  Both head and current base must match. Conditional verdicts, quoted examples,
  incidental SHA mentions and other PRs are excluded. Only the first line is
  projected. Existing PR-comment precedence over Issue comments is preserved;
  each source is sorted/deduplicated by comment ID as in lane-status.
- `ci` reports `rc-ci` and `windows-build` separately, on the exact head, from
  GitHub Actions check runs. Paginated attempts are compared by check ID so an old
  success cannot mask a newer running/failed attempt. Missing is `not_run`;
  running, failure, cancelled, skipped and unknown never become success. A PR run
  has no `windows-build` (#2554, user decision U10: it runs after merge), so a
  missing `windows-build` or one its job condition skipped (`skipped`) does not
  block; one that ran on the head and is not `success` does. `rc-ci` must be
  `success`.
- `blocked_reasons` lists every observed impediment: predecessor, closed/draft PR,
  unknown/expired ownership, open dependency, missing/stale/failed audit,
  unsuccessful CI, merge conflict, behind base, forge block or unknown mergeability.
  An empty list is an observation, not a replacement for the leader's independent
  audit, exact-head status/ruleset checks and final integration decision.
- Ledger Issue/comment identity, source body SHA256 and edit time are returned.
  Each PR is checked again after its reads; the ledger is checked again at the end.
  GitHub offers no transaction over these reads: this is a bounded observation,
  not a guarantee that nothing changes after the call. Re-query before acting.
- Bounds: at most 50 rows, 50 dependencies per row, 50 pages per list, cycle
  detection, and pagination confined to its original endpoint. These conservative
  limits are not measurements. Existing forge retry/timeout/redaction applies.

## Research and verification

Adopt the existing forge GET transport (queue-next), lane-status comment ordering,
claim projection and audit projection. Reject a new persistent queue store: the
remote leader comment is the ordering source. No dependencies were added.

Official GitHub REST documentation was actually retrieved and read on 2026-11-10.
The sources confirm PR head/base and state fields, Issue comments as the PR
conversation, and check-run status/conclusion/head SHA plus pagination. Adopt GET
only; do not use merge/write endpoints. Receipts below hash the retrieved HTML,
not the rendered page. Local fixture verification is separate from live service
validation (product live query and Windows runtime are not_run).

| Source | Retrieval UTC | HTML SHA256 |
| --- | --- | --- |
| https://docs.github.com/en/rest/pulls/pulls | 2026-11-10T23:47:51.461501+00:00 | `73386c4edabefc1e97aeac143ae015ff4172dcbb86115d2a6603e945212eaa53` |
| https://docs.github.com/en/rest/checks/runs | 2026-11-10T23:47:51.726930+00:00 | `151bf45c0bb69974f12a2b88fd5114f202f090b2053d8af05d0e9c9fcab1458b` |
| https://docs.github.com/en/rest/issues/comments | 2026-11-10T23:47:52.041166+00:00 | `63ff36d68796051c6928a3e1f85afff2e5e2d9d11121a995d77f3e175ac08aed` |

Target tests use RecordedTransport (every request must match GET) and a temporary
Git repository. The binary test clears its environment and uses a fixture gh
executable; no installed home or credentials are read. They cover explicit order,
predecessors, head/base/PR mismatch, conditional/quoted verdicts, missing and newer
CI attempts, pagination/cycles, changed ledger/head, expired claims, dependencies,
closed predecessors, malformed input, and absence of a new project store.
