# t621 — todo queue-root resolution, second layer

Base: local `develop` `9935e4e3e` · branch `WT-todo-root-rekey` · worktree `.claude/worktrees/t621`

The card named two residues left by t549 on the branch where a home directory
DOES resolve: (A) the home root is re-keyed, (B) the fallback's layout
detection misses the db-only queue. Both reproduce. The repair turned out to be
one layer up from both symptoms, and the lead approved the widened scope with
three conditions; each is discharged below.

## Reproduction (measured before any edit)

`internal/kanban/home_root_rekey_repro_test.go`, on the branch reached by a
non-git base inside `os.TempDir()` that the temp-origin discriminant does not
classify temporary, with a resolvable home.

**A — re-keying.** `homeTodoQueueRoot` returns `~/.moai/todo/<key(base)>`.
Consumers extend a root with `BacklogPathForRoot`, which re-derives the home
location from whatever root it receives, so the project key is computed from a
home path:

```
canonical  .../002/.moai/db/001-61ad6b5e/todo/backlog.json
resolved   .../002/.moai/db/001-61ad6b5e-628f478f/todo/backlog.json
                            ^^^^^^^^^^^^^^^^^^^^ the doubled key
```

**B — layout blindness.** `fallbackTodoQueueRoot` decided read-through with
`os.Stat` on the `backlog.json` name. The engine's steady state is a sibling
`backlog.db` with no json beside it, so a queue in the normal layout was
invisible: both resolvers read 0 items where 1 was queued.

## Why the repair is the layer, not the two predicates

A re-keyed root does not merely name an odd directory — it FORKS the queue,
because not every surface goes through these resolvers:

| surface | how it resolves | reads |
|---|---|---|
| `moai todo` | `ResolveTodoQueueRootAdopting` → `BacklogPathForRootAdopting` | `db/<key>-<hash>/todo` |
| statusline | state anchor → `BacklogCountsForRoot` (`internal/statusline/backlog.go:48`, `landed.go:210`) | `db/<key>/todo` |
| console watch | `kanban.StateDirForRoot(projectRoot)` (`internal/web/events.go:176`) | `db/<key>/todo` |

The only root whose project key is the project's own is the launch base. And
the home redirection, the temporary-origin refusal, and the adoption of every
legacy location — including this file's own former `~/.moai/todo/<key>`
fallback, which `legacyHomeStateDirsForRoot` already enumerates — all live in
`resolveStateDir`, which is anchored (`state_dir.go` `@MX:ANCHOR`) as the single
directory-layer resolver precisely so two copies of that policy cannot drift.
The home-fallback ROOT layer was the second copy that had drifted.

Repair: both resolvers answer the launch base on every non-git branch;
`fallbackTodoQueueRoot` and `adoptLocalTodoQueue` are deleted with the branch
they served. `homeTodoQueueRoot` stays as the legacy location's name (the
adoption source), no longer a resolution target.

Behaviour change is narrow: only the base-inside-`os.TempDir()`-but-not-
temp-origin branch answers differently. The queue still lands under
`~/.moai/db/<key>/todo` for a non-git project — the same directory, now named by
one key instead of two.

## Condition ① — RED before, GREEN after

Production file swapped to its `HEAD` copy, tests unchanged, then swapped back.

| test | unrepaired | repaired |
|---|---|---|
| `TestT621_HomeBranchRootIsNotReKeyed` | FAIL | PASS |
| `TestT621_StatuslineAnchorAndCommandPathReadOneQueue` | FAIL | PASS |
| `TestT621_FallbackReadsThroughToADbOnlyQueue` | FAIL | PASS |
| `TestT621_AdoptionSeesADbOnlyLocalQueue` | FAIL | PASS |

The statusline test is the one the lead asked for: it writes through the
command path and reads through the project anchor, so it fails exactly when the
two disagree.

`TestT621_HomeBranchRootIsNotReKeyed` passed on BOTH trees in its first form —
it asserted only that the two resolvers agreed and each read 3 items, which a
pair agreeing on one wrong location satisfies. It was strengthened to assert
the path IDENTITY against the project's own queue, and only then separated the
two states. Recorded because a test named for a defect that cannot detect it is
worse than no test.

## Condition ② — the two AC producers, original vs. replacement

**AC-WTQ-008** (`TestResolveTodoQueueRootAdopting_AdoptsLocalQueue`) — criterion:
adopt-NOT-shadow; the command path surfaces the project's existing cards rather
than an empty queue beside them.

| | assertion | status |
|---|---|---|
| was | `root == ~/.moai/todo/<key>` | a re-keyed root; asserting it pins the fork |
| was | `os.Stat(local)` is NotExist after adoption | `seedLocalQueue` resolves through `BacklogPathForRoot`, which since the home-state migration already writes into the home database — so "it moved" now means it moved OFF the canonical path |
| is | `root == dir`, queue under `home`, 3 items, all `queued` | the criterion itself: same count, same states, readable through the resolved root |

**REQ-THG-005 / AC-THG-003** (`TestTodoQueueRoot_NonTempNonGitKeepsHomeFallback`)
— criterion: the guard's trigger is a temporary origin, never the absence of
git; a non-git base KEEPS its home queue.

| | assertion | status |
|---|---|---|
| was | `ResolveTodoQueueRoot(dir) == ~/.moai/todo/<key>` | a claim about where the QUEUE lives, asserted on the root value, which stopped tracking it |
| is | `BacklogPathForRoot(root)` is under `home` and NOT under `dir`, for both entry points; the 2 seeded cards are surfaced | the same claim, on the value that now carries it |

Both replacements PASS on the unrepaired tree as well. That is the point: the
property they assert held before AND after, so rewriting them removed no failing
signal — the defect is carried by the four t621 tests above, which do separate
the states.

Three further tests were re-aimed rather than relaxed:

- `TestResolveTodoQueueRoot_FallbackNoGit` — now asserts the base plus the home
  property on the queue path. FAILs on the unrepaired tree (it pins the new
  contract).
- `TestResolveTodoQueueRoot_PopulatedFallbackWins` → `…_LegacyHomeFallbackQueueIsAdopted`.
  "Which of two locations wins" cannot be carried into a design with one. The
  half that must survive is that a queue already at the old fallback location is
  not stranded; it is reached through `legacyHomeStateDirsForRoot`, asserted with
  a precondition that the scan really names the fixture. `MOAI_HOME` is set
  because that scan resolves its home through `paths.MoaiHome`, which does not
  read this package's `HomeDirFn` seam.
- `TestAdoptingAndPureResolversAgreeWhenAdoptionFails` (+ its `_NonTemp` copy).
  Its central assertion became a TAUTOLOGY under the repair — the two resolvers
  now share one body, so comparing their returns holds with every branch below
  deleted. Re-aimed at what can still fail: the temp-base copy asserts each
  entry point READS the card; the non-temp copy blocks the relocation one layer
  down (`relocateQueueArtifacts`) and asserts a refused relocation still serves
  the cards from where they already are.

## Condition ③ — call sites of the deleted functions

```
HEAD:  internal/kanban/todo_root.go                 9   (definitions + calls)
       internal/kanban/todo_root_contract_test.go   1
       internal/kanban/todo_root_temp_guard_test.go 1
after: internal/kanban/todo_root_contract_test.go   1   prose only
       internal/kanban/todo_root_nontemp_copy_test.go 1 prose only
       internal/kanban/todo_root_temp_guard_test.go 1   prose only
```

Zero code call sites remain; the three hits are comments recording the history.
One of them described the deleted function in the present tense and was
corrected.

## Verification

Repaired tree unless stated.

| check | result |
|---|---|
| `go build ./...` | clean |
| `go vet ./internal/kanban/` | clean |
| `gofmt -l internal/kanban/` | empty |
| `golangci-lint run ./internal/kanban/...` | 0 issues |
| `go test ./internal/kanban/` | ok (169.9s) |
| `go test ./internal/web/ ./internal/statusline/` | ok, ok |
| `go test` — codexadapter, codexwiring, feedback, permission, profile, settings, migration/migrations, cli/uikit, cli/update/merge, cli/worktree | ok (10/10) |
| invisible-character scan (`U+200B–200F`, `U+2060–206F`, `U+FEFF`, `U+00A0`) over the 6 changed files | 0, with a positive control proving the pattern matches |

`internal/cli` is the remaining dependent package and runs in the integration
window.

### Known red, not attributable

`internal/hook` `TestSessionStart_DeferredScanDoesNotBlockReturn` fails 3/3 on
the REPAIRED tree and 3/3 on the UNREPAIRED tree (same worktree, only
`todo_root.go` swapped — and that file is the change's entire production
surface). Pre-existing on `9935e4e3e`.

## Residues

- `pathInsideTempDir` is now reached only from tests. It is kept because t549's
  reachability preconditions assert against it; it names a condition, not a
  branch.
- The `~/.moai/todo/<key>` legacy location is adopted through
  `legacyHomeStateDirsForRoot`, whose home comes from `paths.MoaiHome` while
  this package's own resolution uses the `HomeDirFn` seam. The two disagree
  under a test stub, and would disagree in production only if `MOAI_HOME` were
  set to something other than `<home>/.moai`. Not a defect reached by this card;
  worth a look.
