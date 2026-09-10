# M1 — Tooling baseline (card t332, SPEC-BACKLOG-HYGIENE-001)

Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t332`
Branch: `WT-backlog-hygiene`
HEAD at baseline capture: `6165f9f5e` (after `git merge origin/develop`, absorbing `ee50984ab`)
Primary checkout resolved from `git rev-parse --path-format=absolute --git-common-dir`
→ `/Users/goos/MoAI/moai-adk-go/.git`, parent `/Users/goos/MoAI/moai-adk-go`

## Step 1 — refs refreshed once and pinned (REQ-BH-009, closes plan.md B6)

```
$ git fetch origin develop main
From https://github.com/modu-ai/moai-adk
 * branch                develop    -> FETCH_HEAD
 * branch                main       -> FETCH_HEAD

$ git rev-parse origin/develop origin/main
ee50984abe4f11ac337382b48a26328f091e200a
48239c7dc7428c8751a04f6321887c2d36123884
```

Fetch time (UTC): **2026-08-30T11:16:22Z**

| Ref | Pinned SHA |
|---|---|
| `origin/develop` (integration branch) | `ee50984abe4f11ac337382b48a26328f091e200a` |
| `origin/main` (release branch) | `48239c7dc7428c8751a04f6321887c2d36123884` |

Every landing query in M3 runs against these two SHAs as literals. No landing query names a branch.

## Step 2 — landing method: **path A (direct git)**

Governing `strings` measurement, taken once, before any card was read:

```
$ strings ~/go/bin/moai | grep -c 'worktree_base_branch'
0
$ ~/go/bin/moai version
moai-adk v3.1.2
```

The installed binary predates `260ea5369` and cannot resolve the integration-branch key, so it
answers the landed question against `origin/main` only. **Governing count = 0** (path A: the single
pre-sweep measurement). Under AC-BH-008 no card entry may cite `moai todo pr`'s landed column, and
none does.

Path B (rebuild + reinstall) was **not** taken. Grounds: nine other lanes are live in this batch and
a reinstall replaces the shared installed binary mid-batch, mutating state this read-only card has
no business mutating. Path A needs no install and each verdict cites its own commands.

### Control queries (the positive half of AC-BH-008)

```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt342\b' --oneline
ee50984ab docs(SPEC-ARTIFACT-STATELESS-001): post-merge --strict measurement on develop (t357)
2cc154d55 docs(SPEC-ARTIFACT-STATELESS-001): M3 evidence + AC judgment runner (t357)
15453140a merge(WT-moving-ref-guard): SPEC-MOVING-REF-GUARD-001 — moving-ref invariant guard, advisory emission (t342)
38f937a4f docs(SPEC-MOVING-REF-GUARD-001): re-record MERGE_BASELINE_SHA after the second develop absorption (t342)
de5cc7b08 fix(SPEC-MOVING-REF-GUARD-001): emit MovingRefUnpinned advisory so --strict stops gating on the corpus (t342)

$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt342\b' --oneline
(empty)
```

Non-empty against pinned `develop`, empty against pinned `main` — the asymmetry the binary lag
hides is reproduced, so the method is not vacuous.

## Step 3 — live worktree list

Captured with `git worktree list` at baseline time; full output at `$R/00-worktree-list.txt`.
M3 decides `in-flight-unlanded` against that recorded file, never a re-run.

In-scope cards with a live worktree (branch + tip SHA, from the captured list):

| Card | Worktree | Branch | Tip |
|---|---|---|---|
| t154 | `.claude/worktrees/t154` | `WT-lint-heading` | `dbb87f14f` |
| t216 | `.claude/worktrees/t216` | `WT-hook-wiring-drift` | `8aa96bfb1` |
| t295 | — | — | (no worktree; t313's tree is the related one) |
| t313 | `.claude/worktrees/t313` | `WT-worktree-baseref` | `3fd8b5072` |
| t337 | `.claude/worktrees/t337` | `WT-windows-stamp-liveness` | `c72a517c3` |

(The table above is the M1 pre-read; M3 re-derives it per card from the same captured file and is
the authority. Cards outside this list carry no worktree in the captured output.)

## Finding B1' — a second lag, one layer down: the queue store migrated to SQLite

Discovered while capturing the M2 digest, and it invalidates AC-BH-006's deciding command.

```
$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/state/todo/
-rw-r--r--  225280 Aug 30 20:15 backlog.db          <- live
-rw-------  141146 Aug 29 20:35 backlog.json        <- frozen
-rw-------  155043 Aug 27 23:01 backlog.json.migrated

$ sqlite3 .../backlog.db "select count(*) from items;"          -> 97
$ jq -r '.items|length' .../backlog.json                        -> 96
$ sqlite3 .../backlog.db "select id,state from items where id='t363';"  -> t363|queued
$ jq -r '.items[]|select(.id=="t363")' .../backlog.json                 -> (empty)
$ sqlite3 .../backlog.db "select state from items where id='t294';"     -> picked
$ jq -r '.items[]|select(.id=="t294")|.state' .../backlog.json          -> queued
```

Migrating commit: `3cb258d62 merge(WT-todo-sqlite): integrate card t306 — todo queue JSON to SQLite
plus state-directory rename (SPEC-TODO-SQLITE-001)`. Implementation:
`internal/kanban/backlog_sqlite.go`, `internal/kanban/state_dir.go`.

`backlog.json` is a leftover of the pre-migration store. It is never written now, so the
SPEC-mandated digest over it returns the plan-phase value **whatever the sweep does**:

```
$ jq -S -c '[.items[] | {id, state, text}] | sort_by(.id)' .../backlog.json | shasum -a 256
56e1387eecce52f5b58076cba14e9519bef77b241e240f4b92749026365b346b  -
```

That is byte-identical to the figure acceptance.md recorded on 2026-08-29 — a criterion that cannot
red. **This is the card's own subject reproducing inside the card's own acceptance criteria**: a
measurement whose premise rotted while the sentence citing it stayed unchanged. It is recorded here
rather than smoothed over, and AC-BH-006 is repaired (acceptance.md v0.3.0) rather than satisfied
vacuously.

### The repaired observable, with both directions measured

```
sqlite3 -json <primary>/.moai/state/todo/backlog.db \
  "select id,state,text from items order by id;" | shasum -a 256
```

`findings` is a **separate table** from `items`, so the projection excludes relate-appended findings
structurally — the same property the JSON `.items[]` projection had, now enforced by the schema
(`internal/kanban/backlog_sqlite.go`; `.schema findings` shows `subject_id, related_id, relation,
source, score, note, at`).

| Probe | Digest | Verdict |
|---|---|---|
| live baseline, run 1 | `86a3fb05…dc20` | — |
| live baseline, run 2 | `86a3fb05…dc20` | stable |
| **negative control** — insert a `relate`-shaped row into `findings` | `86a3fb05…dc20` | **unchanged**, as required |
| positive control A — append ` MUTANT` to one card's `text` | `010e05a6…fa20` | changed |
| positive control B — flip one card's `state` to `dropped` | `9024d77c…d270` | changed |

All controls ran against scratch copies under the session scratchpad
(`.../scratchpad/ctl.db`, `ctl2.db`); the real store was read and never written.

## Gaps

- The M1 worktree pre-read table above is a convenience index, not the deciding artifact — M3
  decides per card against `00-worktree-list.txt`.
- Path B was not exercised, so no post-rebuild `strings` count exists. AC-BH-008's path-B conjunct
  is vacuously satisfied by antecedent, and the governing count is the path-A one recorded above.

## Residual risk

- The pinned SHAs are correct as of the recorded fetch time. Work landing on `develop` after
  2026-08-30T11:16:22Z is outside this sweep by construction, and a card reading `not-landed` here
  may have landed since. The fetch time is recorded so a reader can bound that window rather than
  guess it.
- `sqlite3` reads the live store while other sessions may write it. The digest is a point-in-time
  read; M5 re-captures it and the comparison is what decides, not either capture alone.
