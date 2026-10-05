# t384 verdict — moai todo dropped-card render

Card: t384 (moai todo가 dropped 카드를 영구 렌더 — 목록 길이가 실부하와 어긋남)
Branch: WT-todo-drop-render (worktree .claude/worktrees/t384, base a1cba5425 = develop tip at entry)
Verdict: FIXED — default list render hides dropped cards behind a count line; `list --dropped` is the recovery surface.

## Claim

1. The default render (`moai todo` bare, `moai todo list`) no longer renders dropped cards; it renders live cards (queued + picked) plus exactly one count line naming `moai todo list --dropped`.
2. `list --dropped` renders the discarded set with its `[DROPPED — ...]` markers (the undrop recovery surface).
3. `list --json` is byte-unchanged in shape: the structured record keeps every card including dropped ones.
4. The store is untouched — t153's drop/undrop exact-reversal contract holds.
5. The new flag is a declared surface addition (todo_surface_test.go permittedFlagAdditions), not a silent re-flag.

## Evidence

Commands run in this worktree (cwd = worktree root), in order:

1. RED (reproduction, before the fix):
   `go test ./internal/cli/ -run 'TestTodoList_(DroppedHiddenByDefault|DroppedFlagRendersDroppedOnly|DroppedFlagEmptySaysSo|AllDroppedDefaultView|JSONKeepsDroppedCards)' -v`
   → 4 FAIL / 1 PASS: dropped card t3 rendered in the default view (`t3 dropped [DROPPED — superseded by another card] gamma discard me` printed in output); `list --dropped` = `unknown flag: --dropped`; JSON test passed pre-fix (pins the JSON truth).
2. GREEN after the fix:
   `go test ./internal/cli/ -run 'TestTodoList_|TestTodoVerb|TestTodoDrop|TestTodoUndrop|TestTodoBare|TestTodoUndone'`
   → `ok github.com/modu-ai/moai-adk/internal/cli 20.960s`
3. Package re-measurement (first attempt, before golden update):
   `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/`
   → FAIL `TestLiveReadersUnchangedByHistoryVerb` (todo_history_test.go:538: live reader list.txt drifted from the golden) — the golden pinned the defective render.
4. Golden re-capture: `internal/cli/testdata/golden/live-readers/list.txt` rewritten to the post-fix render (provenance note added in TestLiveReadersUnchangedByHistoryVerb doc comment; other five goldens untouched).
5. Final package re-measurement:
   `unset MOAI_KANBAN ... && go test ./internal/cli/`
   → `ok github.com/modu-ai/moai-adk/internal/cli 408.900s`
6. Gates: `go vet ./internal/cli/` rc=0 · `golangci-lint run internal/cli/` = `0 issues.` · `gofmt -l` on touched files: clean · `GOOS=windows go vet ./internal/cli/ && GOOS=darwin go vet ./internal/cli/` = CROSS-COMPILE-OK · `make build` succeeded (catalog.yaml content-unchanged; workflows/*.md is not catalog-hashed).
7. Doc mirror parity: `diff .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md` → identical (MIRROR-IDENTICAL).

## Baseline-attribution

All figures above were measured in this run, in worktree .claude/worktrees/t384 (branch WT-todo-drop-render), against base a1cba5425 (develop tip at worktree creation). The RED run predates the fix in the same tree; the final GREEN run is the tree being committed. No figure is carried from another package, tree, or session.

## Gaps

- The live 40KB queue's post-fix rendering was NOT re-measured against the installed `~/go/bin/moai` binary (the fix travels via develop merge → release; the local binary is v3.1.2 pre-fix). The defect's live presence was observed at session start via the installed binary (`moai todo` = 40.1KB output with dropped cards rendered); the fix's live effect is proven by the package tests, not by an installed-binary rerun.
- Web console rendering was deliberately NOT changed (it renders all three states in a table with explicit state labels — a different surface from the card's complaint). Untouched by scope discipline; not verified beyond reading its tests.
- `internal/web`, `internal/statusline`, `internal/kanban` were not re-measured: no code in them changed. The counts reader (`BacklogCountsForRoot`) already excluded dropped cards pre-fix (read, not changed).

## Residual-risk

- JSON consumers that treated every `items[]` row as live work would still miscount — unchanged behavior, but the doc now states the filter-by-`state` contract explicitly.
- The count-line wording (`N dropped (hidden — see: moai todo list --dropped)`) is pinned only by substring assertions (`1 dropped`, `--dropped`), so cosmetic rewording would not fail tests — deliberate, but it means the recovery-path pointer itself is not byte-pinned.
- If the operator's scripts parse `list` output expecting dropped rows, they break — that is the repair's intent, declared in the surface-test comment and the skill doc.
