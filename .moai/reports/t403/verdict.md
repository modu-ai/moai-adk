# t403 verdict — moai todo large-output silent truncation

Card: t403 (moai todo 대용량 출력 조용히 절단 + 'Full output saved' 라벨 미전달)
Branch: WT-todo-truncation (worktree .claude/worktrees/t403; base b6231290d = develop tip at entry; absorbs WT-todo-drop-render @ a7465699a via merge da212b32a — same render function, composed behavior is what ships)
Verdict: FIXED — the list render is bounded (default 20 rows), `--limit` adjusts / `0` lifts, and a truncated listing states the withheld count on stderr.

## Claim

1. The defect as dispatched is diagnosed: "Full output saved" is the CONSUMER-side (agent harness) truncation label — it appears nowhere in this repository (grep across internal/ pkg/ cmd/ = 0 hits). The moai-side cause is that `runTodoList` rendered every visible row with full text and no bound, so a large queue pushed truncation downstream to the reading harness, where rows vanished from the visible output with no withheld count from the command itself.
2. The repair extends the codebase's own established contract — the history verb's REQ-TAQ-007/008 (bounded default of 20, `--limit <n>`, `0 = unbounded`, withheld count on stderr) — to the list surface: default bound 20, withheld notice `list: <n> rows withheld — showing <k> of <total> (--limit 0 lists all)` on stderr, negative `--limit` refused.
3. `--json` ignores `--limit`: the structured record stays the full read (a bounded JSON would be the same silent truncation).
4. The bound composes with the t384 dropped-collapse: dropped count line stays on stdout, withheld notice on stderr.
5. Bare `moai todo` and `moai todo list` carry the same bound and notice (single runTodoList).

## Evidence

Commands run in this worktree (cwd = worktree root), in order:

1. Premise measurement (this run, live queue, installed binary pre-fix): `moai todo list --json | jq` state distribution = `{"dropped":20,"picked":8,"queued":74}` (102 cards; 82 live) and `wc -c` = `194146` bytes — even post-t384, the live render alone crosses harness truncation.
2. Label-absence measurement: `grep -rn 'Full output' internal/ --include='*.go'` = 0 hits; `grep -rn 'truncat' internal/cli/ internal/kanban/` — the only truncate-notice contract in the todo tree is history's REQ-TAQ-008.
3. RED (before the fix): `go test ./internal/cli/ -run 'TestTodoList_(DefaultBoundedWithWithheldNotice|BareInvocationBoundedToo|LimitRaisesLowersLifts|LimitNegativeRefused|JSONIgnoresLimit|LimitComposesWithDroppedCollapse)'` → FAIL (unbounded render, no withheld notice).
4. GREEN after the fix: `go test ./internal/cli/ -run 'TestTodoList|TestTodoVerb|TestTodoBare|TestTodoDrop|TestTodoUndrop|TestTodoUndone|TestTodoHistory|TestLiveReaders'` → `ok ... 58.007s`.
5. Full package re-measurement: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 354.236s`.
6. Gates: `go vet ./internal/cli/` rc=0 · `golangci-lint run internal/cli/` = `0 issues.` · gofmt on touched files clean · `GOOS=windows go vet ./internal/cli/` OK · `make build` succeeded (catalog.yaml content-unchanged).
7. Doc mirror parity: `diff` live vs template todo.md → identical.

## Baseline-attribution

All figures measured in this run, in worktree .claude/worktrees/t403 (branch WT-todo-truncation), against the tree at merge da212b32a (develop tip b6231290d + absorbed t384 fix a7465699a). The RED run predates the fix in the same tree; the final GREEN run is the tree being committed. The 82-live/194KB figure is a live-queue measurement via the installed pre-fix binary, cited as the defect's magnitude only — the repair's correctness rests on the package tests.

## Gaps

- The live queue was NOT re-read post-fix through a rebuilt binary (fix travels via develop merge; the installed binary predates it). The repaired behavior is proven by the package tests, not by a live-binary rerun.
- The `pr` and `history` verbs' own output bounds were not touched; `history` already bounded (REQ-TAQ-007). `pr` renders one row per card unbounded — out of this card's dispatched scope (todo axis), left as-is.
- Whether the operator's or lead's scripts parse the withheld stderr line is unknown (new stderr text; stdout contract unchanged for bounded queues ≤ 20 rows).

## Residual-risk

- Row-bounding is not byte-bounding: 20 rows of very long card texts can still render large. The withheld notice states rows, not bytes — consistent with history's contract, but a byte-huge queue of ≤ 20 live cards still pushes size downstream. Accepted: matching the established contract; a byte bound would clip card texts.
- A consumer reading `list` stdout for MORE than 20 live rows without `--limit 0` now silently misses rows unless it reads the stderr notice — that is the repair's intended contract change, documented in the skill doc and surface-test declaration.
- The default view shows the OLDEST 20 visible rows (file order); a pick candidate beyond the bound is only visible via `--limit 0` / `--json`. The withheld notice names the escape hatch.
