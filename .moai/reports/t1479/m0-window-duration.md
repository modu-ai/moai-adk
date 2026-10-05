# M0 — in-window duration measurement (card t1479, SPEC-MERGE-WINDOW-QUEUE-001 plan.md M0)

Measured on branch `WT-merge-window-queue` at HEAD `5c406769d` (this measurement precedes the M1
code commit; recorded 2026-10-05, run-phase). The in-window path the SPEC's merge verb executes is
the seconds-long section REQ-MWQ-017 defines after every gate has passed: the tree-identity check
(`git rev-parse` on the record tree and the pinned tree) plus `git merge --no-ff <pinned SHA>` plus
the merged-tree comparison.

## Method

A fixture repository under `/tmp/t1479-m0/fixture` (main + a card branch `cand` five commits ahead
that `main` absorbs; each run merges and then resets). The measured span covers exactly the five
git subprocess calls of the in-window section — two `rev-parse` tree reads, one `rev-parse` of the
cand tip, `git merge --no-ff -q`, and the merged-tree `rev-parse` — with no other work inside it.
Timing via `time.perf_counter_ns()` in Python, subprocess spawn cost included (the real verb pays
it too). Raw driver: `/tmp/t1479-m0/m0-window-timing2.py`.

## Command

```
python3 /tmp/t1479-m0/m0-window-timing2.py
```

Verbatim output (this run, fixture repository, 20 runs):

```
n=20 median=534673.5us max=721861us min=419785us
raw(us): [493403, 458182, 491586, 588347, 603518, 721861, 569022, 571165, 419785, 566854, 465611, 497594, 553303, 464858, 517994, 600827, 551353, 572937, 510162, 484656]
```

## Result

| Metric | Value |
|---|---|
| n | 20 |
| median | 534.7 ms |
| maximum | 721.9 ms |
| minimum | 419.8 ms |

## Shorten-only rule check (AC-MWQ-008 M0 evidence clause)

The measured in-window duration is under 1 second. The shipped 30-minute lease default (REQ-MWQ-008)
is far above it and the 15 s heartbeat / 60 s liveness-window values (leader decision Q14) are also
above it: **no default is lowered** — the shorten-only rule lowers nothing because every current
value already exceeds the measurement by orders of magnitude. The measurement closes the plan-time
gap (spec.md §F row 1): the in-window re-measure figure that motivated the SPEC was a leader
estimate at up to 30 minutes under load; the window as this SPEC builds it holds for hundreds of
milliseconds.
