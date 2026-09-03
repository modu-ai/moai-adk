# t437 — `moai todo --help` stale store claim + Go-layer completeness gap

Card: t437 · Branch: `WT-todo-help-stale` · Base: local `develop` b7462203a

## Claim

1. `moai todo --help` told the reader the queue lives at `.moai/state/todo/backlog.json`
   (and, on the home-fallback path, `~/.moai/todo/<project-key>/backlog.json`). The
   canonical store is `backlog.db`; a `backlog.json` beside it is an export or a legacy
   leftover whose contents can be arbitrarily stale (SPEC-BACKLOG-JSON-DISCLOSURE-001, t395).
   The string is compiled into the shipped binary, so every distributed user saw it.
2. The t395 completeness sweep could not see this. It walks
   `internal/template/templates/` only, so a Go string literal is outside its reach by
   construction — its green said nothing about this layer.
3. D7 (carried over from t395): the same stale assertion sat in three source comments.

## Evidence

Premise measurement, before repair:

    $ grep -rn -E 'state/todo/backlog\.json|moai/todo/<project-key>/backlog\.json' \
        --include="*.go" --include="*.templ" internal/ cmd/ pkg/
    internal/web/screens_templ.go:1242   (comment, generated)
    internal/web/screens.templ:286       (comment, source)
    internal/cli/todo.go:96              (help text — pattern 1)
    internal/cli/todo.go:101             (help text — pattern 2)
    internal/hook/session_start_kanban.go:194  (comment)
    + three `_test.go` sites, which carry the patterns as checkers/fixtures (correct)

Repair: five non-test sites, all now naming `backlog.db`. The help text additionally
states that a `backlog.json` beside the database is not the queue.

Guard: `internal/cli/todo_help_store_test.go` — the same two t395 patterns, run over
`internal/`, `cmd/`, `pkg/` (non-test `.go`/`.templ`, template tree excluded as t395's
subject), plus a point-of-use assertion on the rendered `Long` string.

RED proof (mutation — both repair sites reverted independently):

- `mutant-a-red.log` — todo.go reverted: 2 test failures, 5 messages, both patterns named
- `mutant-b-red.log` — three comments reverted: 3 hits named with file:line
- `green-restored.log` — `ok github.com/modu-ai/moai-adk/internal/cli 0.960s`

Rendered output after repair: `help-render.log`.

## Baseline-attribution

Measured in this run, in this tree (`.claude/worktrees/t437`, base b7462203a):

    go vet ./internal/cli/... ./internal/web/... ./internal/hook/...   → rc 0
    go test ./internal/cli/ -count=1                                  → ok, 465.677s
    go test ./internal/web/... ./internal/template/... ./internal/kanban/... -count=1 → ok
    go test ./internal/hook/... -count=1 → one FAIL (below)

## Gaps

- `internal/hook/quality` FAILs on `TestGateGraphFreshness_AllLayersFreshNotice`. This is
  an inherited red owned by card t304, not produced here: `git diff --stat` shows this
  branch touches four files, none in `internal/hook/quality`, and the one file under
  `internal/hook` is a comment-only change in a different package.
- No cross-platform (windows/linux) run; CI on the develop push is the verdict there.
- `make build` was not run — no template source changed, so the embed check is unaffected.
  The help string is Go source and is picked up by the next ordinary build.

## Residual-risk

- The guard is line-scoped: a claim whose path is split across two lines is invisible to it,
  the same limit the t395 sweep carries.
- The guard reads shape, not intent. A future *correct* mention of `state/todo/backlog.json`
  in non-test Go source (an export or migration doc-comment, say) would be flagged and would
  need an explicit allowlist entry — deliberate, so the exception is stated rather than
  silently absorbed.

---

## What the integration tree showed that no card could show alone

This card's sweep walks non-test Go and templ sources under `internal/`, `cmd/`, and `pkg/`. In the
card's own branch that population is the branch's own code. In the integration tree it is every
card that has landed since — here, cards t305 and t360, which the same lane wrote and merged
first.

That widened population is the point. Either of those cards could have introduced a fresh
`state/todo/backlog.json` claim in non-test Go source; neither card's own suite would have said so,
because the guard did not exist in their trees and their own tests do not look for it. Measured on
the absorbed tree, the sweep passes:

    go test ./internal/cli/ -run 'TestTodoHelp_NamesTheDatabaseStore|TestTodoStoreClaims_NoStaleGoSourceSite'
      → both PASS

So the pass is not a restatement of the card's own green — it is an observation about code the card
never saw, which is exactly the gap the integration branch exists to close: every card was green on
its own, and nobody looked at the combined state.

The same run confirms the sibling t395 sweep still holds on the merged tree
(`TestBacklogJSONDisclosure_EmbeddedTemplatesMatchSource`,
`TestBacklogJSONDisclosure_TemplateMirrorIsComplete`), which matters because this card added a
second sweep over a different layer; two sweeps that disagree would be worse than one.

The limit stays what it was: the sweep sees the layers it walks. A claim landing in a shell script,
a template, or a Markdown rule is outside both sweeps, and neither this observation nor the guard
says anything about those.
