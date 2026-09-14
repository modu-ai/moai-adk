# t542 verdict — orphan queue directories under `~/.moai/todo`

- Card: t542 (clear the residue; the fallback's own guard is t536's)
- Worktree `.claude/worktrees/t542`, branch `WT-orphan-queue-sweep`
- Base: `1d150a27d` (= `origin/develop`). Local `develop` was `5ddccacc9`, 30
  commits ahead; this branch has not absorbed it. The window absorb is owed.
- Unpushed. The lane does not push.
- Operator approval for the home-directory deletion obtained immediately before
  executing it, per the card's [HARD] clause. Scope approved: the 341 `001-*`
  directories only.

## Claim

1. **The card's inventory is stale, and in the direction that matters.** The
   card records 343 directories and 946 files, measured 2026-09-07. This run
   measured **344 directories and 949 files**. The difference is one directory,
   `youtube-4ee46d07`, with an mtime of **2026-09-09T03:31:05Z**.

2. **Inflow is therefore NOT closed.** The card's premise — "유입은 이미
   막혔고 남은 것은 잔재다", resting on the t422 guard landing in develop at
   `a1cba5425` on 2026-09-02 04:14 — is falsified by a directory created a week
   after that landing. Whatever t422 closed, it did not close the path that
   produced this one.

3. **The new orphan is not test-shaped.** The 341 swept directories are named
   `001-<8 hex>`, the shape `t.TempDir()` yields. `youtube-4ee46d07` is named
   after a project. It is empty — zero rows in every table — so nothing was
   lost, but its existence is the evidence that the home fallback still
   resolves for a real project.

4. **Three storage layouts are present, and the card measured one.** The card
   read the 211 `backlog.json` files. There are also **133 `backlog.db` SQLite
   queues** it did not open, and a third, nested `.moai/state/kanban/` layout
   used by the two production-origin directories. Data risk was assessed across
   all three here, not one.

5. **Zero operator cards, across every layout.** 339 card rows total: 209
   `first card` and 128 `first card` (JSON and SQLite respectively), plus
   `fix the drift found in t151` and `e2e probe card one`, one each. Every row
   is a test or probe fixture.

6. **The sweep removed exactly what was approved.** 344 directories → 3;
   11M → 88K. The three production-origin directories and their contents are
   intact.

## Evidence

| Claim | Command and observation | Path |
|---|---|---|
| Directory count | `find ~/.moai/todo -mindepth 1 -maxdepth 1 -type d \| wc -l` → `344` (card: 343) | `01-inventory-full.txt` |
| File census | `find … -type f \| sed 's\|.*/\|\|' \| sort \| uniq -c` → `339 backlog.lock`, `211 backlog.json`, `133 backlog.db-wal`, `133 backlog.db-shm`, `133 backlog.db` = 949 (card: 946) | `02-file-types.txt` |
| Size | `du -sh ~/.moai/todo` → `11M` | `03-du.txt` |
| The new orphan | `find … -type d -exec stat -f '%Sm %N' …` sorted descending → newest is `2026-09-09T03:31:05Z /Users/goos/.moai/todo/youtube-4ee46d07`; next newest `2026-09-02T03:53:09Z` | `01-inventory-full.txt` |
| Non-`001-*` directories | `find … -type d -exec basename {} \; \| grep -v '^001-'` → `proj-325ca0b6`, `t203-probe-d7a16ea2`, `youtube-4ee46d07` | `05-non-001-dirs.txt` |
| JSON card texts | `find … -name backlog.json -print0 \| xargs -0 grep -ho '"text":[^,}]*' \| sort \| uniq -c` → `209 "first card"`, `1 "fix the drift found in t151"`, `1 "e2e probe card one"` | `04-json-card-texts.txt` |
| SQLite card texts | `os.walk` over the tree (see note), 133 `backlog.db` opened read-only → `items: 128`, `archived_items: 0`, `findings: 0`, `archived_findings: 0`; every item text `first card`; 49 databases predate the `archived_*` tables | `06-sqlite-card-texts.txt` |
| `youtube-4ee46d07` is empty | `sqlite3 …/youtube-4ee46d07/.moai/state/todo/backlog.db "select 'items', count(*) …"` → `items\|0`, `archived_items\|0`, `findings\|0`, `archived_findings\|0` | transcript |
| Production-origin contents | `head -c 500` on both nested `backlog.json` → `e2e probe card one` (state `queued`, added 2026-08-23T11:23:23Z) and `fix the drift found in t151` (state `picked`, added 2026-08-23T16:20:49Z) | transcript |
| Target list frozen and validated | `find … -name '001-*'` → 341 lines; `grep -cv '^/Users/goos/\.moai/todo/001-[0-9a-f]\{8\}$'` → `0` non-conforming | `07-delete-targets.txt` |
| The removal | `tr '\n' '\0' < 07-delete-targets.txt \| xargs -0 rm -rf` → exit 0 | transcript |
| After | directory count `3`; `du -sh` → `88K`; remaining files `1 backlog.db`, `1 backlog.db-shm`, `1 backlog.db-wal`, `2 backlog.json`, `2 backlog.lock` | `08-inventory-after.txt`, `09-file-types-after.txt`, `10-du-after.txt` |

## Baseline-attribution

- Every figure above was measured in this run against this machine's live
  `~/.moai/todo`. None is carried from the card.
- The deletion list is the file the deletion consumed — the same
  `07-delete-targets.txt` was validated by regex and then piped into `rm`, so
  the enumerated set and the removed set cannot diverge.
- The SQLite census used `os.walk` rather than `glob('**/backlog.db')`. The
  glob form was tried first and returned **0 databases**, because Python's glob
  does not descend into dot-directories and the queues live under `.moai/`. An
  empty result from a scan that cannot see the target is not a measurement of
  zero; the walk is the attributable form.
- Card texts were read from both formats before any deletion, never after.

## Gaps

- **The producer of `youtube-4ee46d07` was not identified.** Its mtime and its
  name are all this run establishes. Which command, which session, and which
  project path triggered the home fallback on 2026-09-09 was not traced —
  no transcript scan, no log correlation.
- **Whether t422's guard is incomplete or simply does not cover this path is
  unknown.** The guard was not read in this card, and no reproduction of the
  fallback was attempted. Claim 2 states that inflow is open, not why.
- **The `backlog.lock` count is unexplained.** 339 locks against 344
  directories before the sweep; which five lacked one, and why, was not
  examined.
- **The two `backlog.db-wal` / `-shm` sidecars were not inspected for
  uncheckpointed content.** The read-only connections would have replayed a WAL
  on open, so the row counts are believed complete, but the sidecars were not
  independently read.
- **Nothing was backed up before deletion.** The inventory and the card-text
  census are the record; the directories themselves are gone. That was the
  approved action, and it is worth stating plainly rather than implying the
  evidence is a restore path.
- **Recurrence is not measured.** Whether new orphans appear after this sweep
  is a future observation, not one this run can make.

## Residual-risk

- **The residue will come back.** The sweep removed the accumulation, not the
  path that fills it. With inflow open (claim 2), the count grows again from 3,
  and nothing currently watches it.
- **The preserved three are load-bearing for another card.** They are the only
  surviving physical evidence that the home fallback reached a real project
  path. Anyone tidying `~/.moai/todo` to zero destroys t536's material.
- **A future orphan may not be empty.** All three surviving directories, and
  every one swept, held fixtures. `youtube-4ee46d07` happened to be empty; the
  same path reached by a project mid-session would carry that project's queue
  outside the project.

## Disposition

- Deleted: the 341 `001-*` directories (operator-approved scope).
- Preserved: `proj-325ca0b6`, `t203-probe-d7a16ea2`, `youtube-4ee46d07` — kept
  until t536 closes, per the card.
- Claim 2 (inflow still open, with a dated instance) is outside this card's
  scope, which is residue only. It belongs to t536 and is reported to the lead
  rather than acted on here.

## Addendum — 2026-09-14: recurrence observed (first post-sweep orphan)

The Gaps section recorded "Recurrence is not measured" and Residual-risk
predicted "the count grows again from 3, and nothing currently watches it."
Both are closed by observation ~19.5 hours after the sweep (lane-3, card t542
continuation dispatched by the lead; totals updated, no deletion performed).

**Observation.** The queue holds **4** directories, not the 3 this verdict
left. The newcomer is `001-3d9b96cd` — birth 2026-09-14 04:19:58 +0900
(`added_at 2026-09-13T19:19:58Z`, identical), 8K. Nested
`.moai/state/kanban/` layout (`backlog.json` + `backlog.lock`), one item:
`first card`, state `queued` — a fresh fixture scaffold, zero operator data.

**What this strengthens.** Claim 2 (inflow open) no longer rests on a single
non-test-shaped specimen. `001-3d9b96cd` is exactly the `001-*` test shape
the t422 guard was supposed to stop, arriving twelve days after the guard
landed in develop (`a1cba5425`, 2026-09-02 04:14) — test-origin inflow is
open too. It also uses the nested kanban layout previously seen only on the
two production-origin directories, so the nested-layout producer now emits
`001-*` names as well. That is an observation, not a producer
identification; the producer-unidentified Gap above stands.

**Disposition.** NOT deleted. The 2026-09-13 operator approval covered the
341 `001-*` directories enumerated that day; this one did not exist yet, and
the card's [HARD] clause requires re-approval immediately before any
home-directory deletion. Preserved alongside the other three as t536
material — arguably the sharpest of the four specimens, since it shows the
guard failing at its own stated job. Deletion awaits the operator's call,
routed through the lead.

**Totals after this addendum:** 4 directories, 96K — `001-3d9b96cd` (8K,
recurrence specimen), `youtube-4ee46d07` (non-test-shaped specimen),
`proj-325ca0b6`, `t203-probe-d7a16ea2` (production-origin). Commands and
verbatim outputs: `11-recurrence-20260914.txt`.

🗿 MoAI
