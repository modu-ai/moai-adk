# t119 — `moai todo` correction verbs (edit / move)

Card scope as narrowed by the lead: **the prerequisite CLI verbs only**. No
automatic analysis, absorption, or drop logic — those collide with the [HARD]
clauses in `workflows/todo.md` and `kanban-dispatch.md`, and a doctrine change
must come first.

- branch: `WT-t119`
- base: `origin/release/v3.1.1` @ `b317f47c4` (rebased after the lead's base
  correction; the work was authored on `4100d8767` and rebased, not redone —
  the rebase applied cleanly, one commit, no conflicts)
- commit: `1ceac9377`
- pushed: **no** (per the dispatch)

## What shipped

| Verb | Form |
|---|---|
| `edit` | `moai todo edit <n> "<text>" [--expect <prefix>]` |
| `move` | `moai todo move <n> (--top \| --bottom \| --before <m> \| --after <m>)` |

Both are operator acts, both reversible — the property the lead named as the
card group's core risk:

- **edit** rewrites only `text`. `id`, `added_at`, `state`, and `spec_id`
  survive, so a correction no longer churns the card's identity the way
  `done` + re-add does. The confirmation prints the **prior** text alongside
  the new one, so a wrong edit is reversed by editing back. `--expect
  <prefix>` mirrors `next --expect`: an id typed from a stale listing cannot
  silently rewrite a different card.
- **move** is a pure permutation of the item slice — nothing dropped,
  duplicated, or altered — so a wrong move is reversed by another move.
  Exactly one destination flag is required; zero or two is a malformed
  invocation, never a guess the CLI resolves.

Every refusal returns from inside `BacklogStore.Mutate`'s callback, which
writes nothing, so the queue file stays byte-identical on a miss, a mismatch,
or a malformed invocation.

## Deliberately excluded

- **drop / undrop.** The store carries `BacklogStateDropped` with no CLI verb
  reaching it (which is why the four dropped cards in the live backlog were
  hand-decided), but "drop" is neither text correction nor reordering — the
  two capabilities the card names. Left for a follow-up card, where the
  reversibility condition the lead set would apply (an `unpick`-shaped
  `undrop`).
- Any inference about what to correct or where a card belongs.

## Files

| File | Change |
|---|---|
| `internal/cli/todo_edit_move.go` | new — both verbs + `applyTodoMove` |
| `internal/cli/todo_edit_move_test.go` | new — 8 tests |
| `internal/cli/todo.go` | `AddCommand` wiring (2 lines) |
| `.claude/skills/moai/workflows/todo.md` | two table rows + a [HARD] operator-act boundary note |
| `internal/template/templates/.claude/skills/moai/workflows/todo.md` | mirror (byte-identical) |

`make build` run (templates embedded). `internal/template/catalog.yaml`
unchanged, correctly: the catalog hashes a skill's root `SKILL.md` only
(`gen-catalog-hashes.go` `resolveHashSourcePath`), not `workflows/*.md`.

## Evidence

| Claim | Command | Observed |
|---|---|---|
| new tests pass | `go test ./internal/cli/ -run 'TestTodo' -count=1` | `ok … 6.718s` |
| affected package green (whole `internal/cli`) | `go test ./internal/cli/ -count=1 -timeout 580s` | `rc=0`, 0 `--- FAIL` — log: `.moai/state/verify/t119/cli-full-rebased.log` |
| templates neutral + parity | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` | `ok … 23.001s` |
| static analysis | `go vet ./internal/cli/` | clean |
| formatting | `gofmt -l` on the three touched Go files | clean |
| mirror parity | `diff` local vs template `todo.md` | identical |
| real binary, happy path | `bin/moai todo add/edit/move/list` in a throwaway git repo | `edited t1 fix the gate (was fix teh gate)` / `moved t3 1 third` / `moved t1 3 fix the gate`, final order `t3,t2,t1` |
| real binary, refusals | `move 1` (no dest), `edit 9`, `move 1 --before 1`, `edit 1 "  "` | all `rc=1` with a named reason |

## Gaps (not observed)

- Cross-platform: not run under `GOOS=windows` / `linux`. The change adds no
  path or syscall surface (slice permutation + the existing store), but the
  Windows matrix verdict is CI's.
- Concurrency: no new cross-process race test. Both verbs go through the same
  `Mutate` lock path the existing verbs use; the lock itself is covered by the
  store suite.
- Full repository suite: not run locally, per the standing rule — CI on the
  integration PR is the verdict.

## Residual risk

- `move` addresses positions in the **whole** item list (queued + picked +
  dropped), while `add` reports a position among **queued** cards only. The
  two numbers can differ on a queue holding picked cards. Handled by making
  the landing position explicit in the move confirmation, not by reconciling
  them — reconciling would change `add`'s long-standing output.
- A pre-existing failure was observed on the OLD base
  (`TestGLMTask_ForwardsSystemModelAndMaxTokens`, reproduced on the untouched
  `WT-t146` tree at `4100d8767`); it passes on `origin/release/v3.1.1`, so it
  was fixed on the release branch and is unrelated to this card.
