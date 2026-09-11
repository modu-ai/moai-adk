# t629 — cost of widening the secret scan on this repository

card: t629
worktree: .claude/worktrees/t629, HEAD at measurement feeecc980 (tree = local develop 80e9e0039)
measured: 2026-09-10 17:37-17:39, lane-5

The card requires that widening the scan's scope be costed with measured numbers, and that any
proposal to narrow the scope again for cost be put to the operator rather than decided by the
lane. This file is the measurement; it proposes nothing.

## Claim

On this repository, the incremental scan over a typical recent range takes a fraction of a
second, while the full `--all` scan takes over a minute — and the full scan finds credential-shaped
matches in 15 commits, 14 of which are not reachable from HEAD. An incremental scan anchored on HEAD
would never reach those 14.

## Evidence

All commands are plain `git`, run in this worktree, using the exact regex from `review.md`.
`--format=%H` prints only commit SHAs, so no patch text — and no matched content — was printed or
stored. Timing is wall clock, taken with timestamps printed immediately before and after each scan.

```
$ git rev-list --all --count                                  → 12157
$ git rev-list --count d060e0d13..HEAD                        → 55
$ git rev-list HEAD  (to file)                                → 7067 commits
```

| scan | commits in scope | wall time | exit | commits matching the regex |
|---|---|---|---|---|
| incremental, `d060e0d13..HEAD` | 55 | **0.207 s** (1789029467.534 → .741) | 0 | 0 |
| full, `--all` | 12,157 | **76.084 s** (1789029467.750 → 543.834) | 0 | **15** |

Reachability of the 15 matching commits, by comparing SHA lists only (`grep -F -x -f`):

| reachable from HEAD | reachable only through other refs |
|---|---|
| **1** | **14** |

**Measurement condition — contended, not a clean benchmark.** Load averages were
18.99 / 14.05 / 10.28 immediately before the scans and **34.23** / 17.92 / 11.95 immediately after.
Other lanes were active, and the 76-second full-history scan itself likely contributed to the rise.
The 76 s figure is therefore an inflated, load-dependent number; the ratio between the two scans
is the more robust reading than either absolute time.

## Baseline-attribution

Every figure was measured in this run, on this worktree, at the SHAs named. None is carried over.

## Gaps

- **The content of the 15 matching commits was deliberately not examined.** They may be synthetic
  test fixtures, documentation examples, or real leaks; which is unknown. Their SHAs are kept only
  in the session scratchpad, not in this repository, so that this file does not become a map of
  where credential-shaped strings live. Following up on them is outside this card and is the
  operator's call.
- No clean-machine timing; the full-scan time was taken under contention.
- The cost of alternative designs (for example, scanning only commits newly reachable from any ref
  since the last scan) was not measured.
- `--all` here includes this repository's many worktree and card branches; a user project with few
  branches would see a much smaller ratio. Not measured.

## Residual-risk

A single contended timing can mislead in either direction. Any design decision that turns on the
absolute cost should be re-measured on an idle machine first.
