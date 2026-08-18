# t104 — Lane dispatch exit-first worktree procedure — Implementation Evidence

Card: t104 · Branch: `WT-t104` · Base: `origin/release/v3.1.1` @ `1f2780227` · Impl commit: `77d09fc5a`

## 1. Claim

1. `kanban-dispatch.md` now codifies the **exit-first step**: a lane still anchored in a previous card's worktree MUST `ExitWorktree` back to the primary checkout before creating the new card worktree (`EnterWorktree(<card-id>)` cannot run from inside a worktree session; skipping the exit lands the new card's work on the old card's branch).
2. The dispatch block's `wt` field spec now states that `wt` names the **new** card's worktree, never a previous card's tree, and may carry the exit-first instruction chain (`ExitWorktree` → `EnterWorktree(<card-id>)` → `git branch -m WT-<card-id>`).
3. Additions are offset by meaning-preserving trims in the same file — **net −1 byte (23,610 → 23,609)** — satisfying the same-file offset discipline for an always-loaded rule.
4. Template mirror is byte-identical; `make build` rerun; always-loaded token budget guard PASSES.

## 2. Evidence

Worktree base (dispatch premise was `ba679791c`; remote had advanced — remote tip used per the dispatch's own `base: origin/release/v3.1.1` field):

```
$ git rev-parse --short origin/release/v3.1.1
1f2780227
$ git log --oneline ba679791c..1f2780227
1f2780227 merge: pr1566 — register moai mcp on root — review-PASS(hub) · coderabbit review-completed(status f7de9eb85)
```

File size (offset discipline; original measured before edit):

```
$ wc -c .claude/rules/moai/workflow/kanban-dispatch.md   # before edit
   23610
$ wc -c .claude/rules/moai/workflow/kanban-dispatch.md   # after edit + 24 trims
   23609
$ git diff --stat -- .claude/rules/moai/workflow/kanban-dispatch.md | tail -1
 1 file changed, 19 insertions(+), 18 deletions(-)   (first pass; final: 54 ins/52 del across the 2-file pair)
```

Mirror parity (after `cp` to template source):

```
$ diff -q .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md && echo MIRROR-IDENTICAL
MIRROR-IDENTICAL
```

Always-loaded budget guard (authoritative Go test, env-scrubbed compound form):

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget|TestAlwaysLoadedSurfaceEnumeration|TestMeasureAlwaysLoaded_WithMemory' -count=1 -v
--- PASS: TestAlwaysLoadedTokenBudget (0.00s)
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.00s)
--- PASS: TestAlwaysLoadedSurfaceEnumeration (0.01s)
--- PASS: TestMeasureAlwaysLoaded_WithMemory (0.00s)
ok  github.com/modu-ai/moai-adk/internal/config	0.466s
```

Mirror-drift guard:

```
$ unset … && go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror' -count=1 -v
--- PASS: TestLateBranchTemplateMirror (0.00s)
--- PASS: TestRuleTemplateMirrorDrift (0.00s)
ok  github.com/modu-ai/moai-adk/internal/template	0.502s
```

Template neutrality self-check (added lines scanned for SPEC IDs / REQ tokens / dates / commit SHAs / card ids):

```
$ git diff -U0 -- …kanban-dispatch.md | grep -E '^\+' | grep -cE 'SPEC-…|REQ-…|20[0-9]{2}-…|[0-9a-f]{7,40}|…t[0-9]{2,}…'
0
```

Build:

```
$ make build
go build -ldflags "… -X …/pkg/version.Commit=1f2780227 …" -o bin/moai ./cmd/moai   # exit 0
```

## 3. Baseline-attribution

- Token baseline measured on THIS tree (WT-t104 @ `1f2780227`), before the edit, by replicating `measureAlwaysLoaded` (no-`paths:` rules + CLAUDE.md + moai.md + repo-root MEMORY.md[absent→0], bytes/4):
  `217,753 + 20,380 + 65,238 = 303,371 bytes / 4 = 75,842 tokens` vs budget `76,000` (`AlwaysLoadedTokenBudget`, `internal/config/token_budget_guard.go:32`) → **headroom ≈ 158 tokens (~632 bytes)**.
  Post-edit the file is 1 byte smaller, so the surface is unchanged or lower; the authoritative post-edit verdict is the Go test PASS above, run after the edit.
- **Premise correction (dispatch said "여유 ≤3토큰")**: measured headroom on this tree is ≈158 tokens, not ≤3. The ≤3 figure did not reproduce; the offset discipline was applied anyway (net −1 byte).

## 4. Gaps

- Full `go test ./...` suite NOT run locally (lane-local verification discipline — CLAUDE.local.md §4; the card touches docs/templates only, and the two owning packages config/template were run directly).
- No CI run applicable at this stage: release-branch pushes sit outside the `branches: [main]` CI triggers (noted in `token_budget_guard.go:28-29`), and this card integrates via the batch release branch, not its own PR.
- Hub review pending — the meaning-preservation of the 24 trims is asserted by the implementer and awaits reviewer confirmation.

## 5. Residual-risk

- The char/4 estimator carries the guard's documented ±15%; absolute token numbers are tripwire readings, not accounting.
- Headroom (~158 tokens) remains thin — any future always-loaded growth in ANY of the 14 no-`paths:` rule files (not just this one) can consume it; the root fix (stub + lazy-load diet for large always-loaded rules) is a separate card per `token_budget_guard.go:29-31`.
- The exit-first rule text documents the procedure; the t79-style incident specifics (card ids, dates) were deliberately kept OUT of the rule + mirror (template-neutrality forbidden classes), so the case study lives only in the card text and this report.
