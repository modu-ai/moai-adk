# t1425 — measurements the lane took itself (after the run phase, before the sync audit)

tree: worktree `.moai/worktrees/t1425`, HEAD `9fbf6d0a8`, branch `WT-harness-prune-single-writer`, clean (`git status --short` empty before and after).
judging tools: `go` toolchain of this machine; no installed `moai` build was used to judge anything.

## 1. How long one prune takes on a copy of the live log (the hook timeout is 5 s)

Why it matters: the harness-observe hooks run with `"timeout": 5` and `"async": true` (`internal/template/templates/.claude/settings.json.tmpl` lines 147, 204, 223, 272). A pruner or a lock waiter that outlives the timeout is killed. A pruner killed after the archive append and before the rewrite leaves the stale events in the log, so the next pruner archives them again.

Method: a throwaway test (deleted afterwards, never committed) read the live log `/Users/goos/MoAI/moai-adk-go/.moai/harness/usage-log.jsonl` read-only (65,804,810 bytes at that moment), wrote a copy to a temp dir per case, and timed ONE `PruneStaleEntries(30)` with `nowFn` shifted forward by N days. Run under the slot lease `go-test-internal-harness` (acquired exit 0, released exit 0, status `free` afterwards), `timeout 600 go test -timeout 8m -count=1`.

| now shifted by | elapsed (one prune) | log bytes after | meaning |
|---|---|---|---|
| +0 days | 548.795 ms | 65,804,810 (unchanged) | nothing stale: read-only scan |
| +2 days | 528.790 ms | 65,804,810 (unchanged) | nothing stale: read-only scan |
| +6 days | 1.788 s | 57,575,630 | about 12.5 percent of the bytes stale: archive append + whole-log rewrite |

Reading: on this machine at that moment the worst measured case (12.5 percent stale, full rewrite) is 1.79 s, about 36 percent of the 5 s timeout. The 2026-10-04T09:18Z case expires roughly one day of events, which lies between the first two rows and the third. NOT measured: the same prune on a machine under the load of the incident, and a lock waiter queued behind a prune on a loaded machine. Measured once per row, not repeated.

## 2. Guard re-run by the lane

Command (one invocation, slot lease held): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 600 go test -timeout 8m -count=3 -run TestPrune -v ./internal/harness/` redirected to a scratch file.
Observed: `ok  github.com/modu-ai/moai-adk/internal/harness  1.269s`; 30 `--- PASS` lines, 0 `--- FAIL` or `FAIL` lines; `TestPruneConcurrentProcessesSingleRewrite` passed three times (0.23 s, 0.28 s, 0.30 s).

## 3. Documentation impact (why no docs agent was spawned)

`grep -rln -i -E 'usage-log|retention|PruneStale|prune-state'` over `docs-site/content`, `.claude/rules`, `.claude/skills`, the template rules and `.moai/harness/README.md` found only unrelated retention topics and one artifact list (`docs-site/content/en/workflow-commands/moai-harness.md:170`, a directory listing). No page describes the prune interval or its mechanics, and the new `<log>.prune-state` file is an internal state file next to the log. Nothing in the repository's documentation needs to change. CHANGELOG.md is not written by the lane (leader writes it at release time, per the leader's standing decision on the previous card).

## 4. Gaps in this file

Not run: the live 70 MB log with hundreds of hooks; Windows tests; the `internal/cli` tests that merely mention the usage log (listed by the run agent in `run-evidence.md`).
