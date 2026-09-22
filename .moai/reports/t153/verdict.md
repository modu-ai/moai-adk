# t153 — `moai todo` discard verbs (drop / undrop)

Class B, run → sync in-lane. The card follows the gap t119 recorded: the
store carries `BacklogStateDropped` with no CLI verb reaching it, which is
why the four dropped cards on the live queue (`t6 t7 t10 t18`) were written
into the file by hand.

- branch: `WT-todo-drop`
- base: `origin/release/v3.1.1` @ `ca7e15fa2` (checked — `EnterWorktree`
  opened on `4100d8767` as warned, moved with `git fetch` +
  `git merge --ff-only` **before** any edit, so no rebase was needed this time)
- commit: `3f8952dc1`
- pushed: **no** (per the dispatch)

## What shipped

| Verb | Form |
|---|---|
| `drop` | `moai todo drop <n> "<reason>" [--expect <prefix>]` |
| `undrop` | `moai todo undrop <n> [--expect <prefix>]` |

`drop` moves a **queued** card to `dropped` and prefixes its text with
`[DROPPED — <reason>] `, the convention the hand-written cards already use.
The card stays in the file — `done` removes a finished card, `drop` keeps a
discarded one visible with its reason — and it stops being a pick candidate
(bare `next` lists queued cards only).

`undrop` reverses it. **The state is the authority, not the marker**: a card
marked dropped by hand with no marker in its text undrops with its text
untouched, so the four existing cards are recoverable through the CLI rather
than by another hand edit.

## Exact reversal — how it is held

The [HARD] requirement is that `undrop` reverse `drop` precisely. Three
refusals exist to hold it, and each would look merely cautious without the
reason:

| Refusal | Why reversal needs it |
|---|---|
| only a **queued** card may be dropped | dropping a picked card would lose `state` and `spec_id` with nowhere to record them — the per-item field set is frozen at five fields — so `undrop` could not restore what `drop` took. `unpick` then `drop` is the two-step path, and both steps reverse. |
| the reason may not contain `]` | the marker is parsed off at the first `] `, so such a reason would strip the wrong span and silently corrupt the restored text. |
| an already-dropped card is refused, not re-marked | a second marker would leave `undrop` restoring the card to a still-marked state. |

With those in place the round trip is **byte-exact**, asserted directly:
`TestTodoDrop_UndropIsAnExactReversal` captures the queue file before the
drop and compares bytes after the undrop. The real binary reproduces it
(`diff` → identical).

## Contract carried over from t119

- every refusal returns from inside `BacklogStore.Mutate`'s callback, which
  writes nothing → the queue file stays byte-identical (8 refusal cases
  asserted);
- `--expect <prefix>` on both verbs guards an id typed from a stale listing;
  on `undrop` it matches the card's **current** (marked) text, which is what
  a listing shows;
- nothing is inferred.

## Deliberately excluded

No automatic drop, staleness heuristic, duplicate detection, or absorption of
one card into another. Those collide with the [HARD] clauses in
`workflows/todo.md` and `kanban-dispatch.md`; a doctrine change comes first.
The doc surface states the same boundary for the two new verbs.

## Files

| File | Change |
|---|---|
| `internal/cli/todo_drop.go` | new — both verbs + `stripTodoDropMarker` |
| `internal/cli/todo_drop_test.go` | new — 7 tests |
| `internal/cli/todo.go` | `AddCommand` wiring (2 lines) |
| `.claude/skills/moai/workflows/todo.md` | two table rows, the [HARD] operator-act note extended to all four correction verbs, and the `state` bullet noting a dropped card stays and is recoverable |
| `internal/template/templates/.claude/skills/moai/workflows/todo.md` | mirror (byte-identical) |

`make build` run. `catalog.yaml` unchanged, correctly — the catalog hashes a
skill's root `SKILL.md` only, not `workflows/*.md`.

## Evidence

| Claim | Command | Observed |
|---|---|---|
| new tests pass | `go test ./internal/cli/ -run TestTodoDrop -v -count=1` | 7/7 PASS |
| whole todo surface green | `go test ./internal/cli/ -run TestTodo -count=1` | `ok … 12.136s` |
| affected package green | `go test ./internal/cli/ -count=1 -timeout 580s` | `rc=0`, 0 `--- FAIL` — log `.moai/state/verify/t153/cli-full.log` |
| templates neutral + parity | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` | `ok … 21.184s` |
| static analysis / formatting | `go vet ./internal/cli/`, `gofmt -l` on the three Go files | clean |
| mirror parity | `diff` local vs template `todo.md` | identical |
| real binary, round trip | `bin/moai todo drop 1 "전제 반증" --expect 전제` then `undrop 1`, `diff` of the queue file before/after | `dropped t1 전제 반증된 카드 (reason: 전제 반증)` → `undropped t1 전제 반증된 카드`; file **IDENTICAL** |
| dropped card leaves the candidate set | `bin/moai todo next` after the drop | only `t2` listed |
| refusal exit codes | picked-drop / undrop-of-queued / double-drop, measured directly | `rc=1` each, with a named reason; the valid drop is `rc=0` |

## Gaps (not observed)

- Cross-platform: not run under `GOOS=windows` / `linux`. No path or syscall
  surface is added (string prefix + the existing store), but the matrix
  verdict is CI's.
- No new concurrency test — both verbs go through the same `Mutate` lock path
  the existing verbs use.
- Full repository suite not run locally, per the standing rule.
- The four live dropped cards were **read** to confirm the marker convention;
  no live queue card was mutated by this work.

## Residual risk

- The marker lives in the card text because the per-item field set is frozen.
  A card whose original text already begins with `[DROPPED — …] ` (written by
  hand, then dropped again through the CLI) would undrop one marker deep and
  leave the hand-written one. The double-drop refusal covers the CLI path;
  the hand-written-then-CLI-dropped path is not reachable today, since a
  hand-marked card also carries state `dropped` and is refused.
- `drop` refuses a picked card. That is deliberate (see the reversal table)
  but it is a behavioural constraint an operator may hit: the recovery is
  `unpick <n>` then `drop <n> "<reason>"`.
- A `[MoAI Security Guardian] sql-injection` finding fired on the markdown doc
  edit. False positive — the changed file is documentation and contains no
  SQL; nothing was suppressed or changed in response.
