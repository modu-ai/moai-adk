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

---

## Reds excluded at merge time, and why each is not this card's

Three tests are red on the absorbed tree. Each is attributed by measurement, not by argument: this
branch's own contribution over `develop` is ten files —

    git diff --name-only e969dc07d a9fa07a9a
      → .moai/reports/t437/*  ·  internal/cli/todo{,_help_store_test}.go
        internal/hook/session_start_kanban.go  ·  internal/web/screens{.templ,_templ.go}

— and none of them is an input to any of the three.

**`TestManifestHashFormat`** (`internal/template`) — `CATALOG_HASH_UNSTABLE` for `sync-auditor`.
Present before this card's first absorb and unchanged by it. Owned by card t443.

**`TestCatalogHashCoversSkillSubfiles`** (`internal/template`) — `CATALOG_HASH_SKINNY`: the `moai`
catalog hash no longer covers its deployed directory tree. Attributed to the t191 absorb by
**delta**, which is the only way it could be attributed at all: on the tree absorbed from
`a239cf050` this package had exactly ONE red (`TestManifestHashFormat`); after absorbing
`e969dc07d` it had two. Keeping that intermediate measurement is what made the second red
attributable rather than merely present. t191 added skill subfiles under
`templates/.claude/skills/moai/workflows/project/` without regenerating `catalog.yaml`. The repair
belongs to card t436, whose `--entry moai` regeneration covers this too — deliberately NOT done
here, because two cards regenerating the same line would collide.

**`TestAlwaysLoadedTokenBudget`** (`internal/config`) — the always-loaded surface measures 76,939
tokens against a 76,400 budget, 539 over. The surface is `.claude/rules/moai/**.md` without a
`paths:` restriction, plus `CLAUDE.md`, `AGENTS.md`, and the active output style
(`token_budget_guard.go:220`). This branch edits none of them, so the surface bytes are identical
to `develop`'s and the result is too. Owned by card t453, which raises the ceiling to 77,200.

That last one is worth a note beyond its ownership: 77,200 leaves 261 tokens of headroom against a
surface whose largest single entry is 66,255 bytes. The guard will fire again on the next card that
adds a line to any always-loaded rule, and raising the ceiling each time is not a strategy.

## How the re-measurement scope was chosen

Not from a summary. The packages to re-run were derived from what the absorb actually carried:

    git diff --name-only <pre-absorb> <post-absorb> | grep '^internal/' | sed 's|/[^/]*$||' | sort -u

That measurement found `internal/template/templates/` in the absorbed set, which a summary of the
same absorb had omitted — and running the template package on that basis is what surfaced
`TestCatalogHashCoversSkillSubfiles`. Had the scope been taken from the summary, that red would
have merged unnoticed.
