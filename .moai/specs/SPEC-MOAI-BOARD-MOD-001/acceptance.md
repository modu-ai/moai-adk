# Acceptance — SPEC-MOAI-BOARD-MOD-001

**Measurement tree (document-level pin, binds every criterion that carries no pin of its own):** `802a72235536958ada5b7cd5876a168e4b8c325f`, branch `WT-moai-board-mod`. Judging build: `claude 2.1.287`, `moai v3.2.0-rc.25` (tool provenance, `verification-claim-integrity.md` §2.2). Every RED-now cell below is a command that was run in this session on that tree; its stdout is verbatim (bounded) and its exit code is its own field. Paths in stdout are the worktree's absolute paths.

## §A Dispositions that apply to every criterion

1. **Named tests, counted.** A test-bearing criterion is judged on `claude plugin test mods/moai-board` by counting lines of the form `(pass) <prefix>…` against the expected count **and** observing zero `(fail)` lines. A count of 0 is not a pass: an empty sweep asserts nothing (`verification-completeness.md` §1.1). The runner itself exits 1 when no test file exists (M-9), and the RED-now cells observe that.
2. **A refused runner is UNOBSERVED.** If the runner prints `hooks modules are turned off in this process: the rollout switch served off` (observed four times in one session, M-13), the criterion is **unobserved** — not pass, not fail. Run-phase retries and records each attempt with the output. A test-bearing criterion closes only on one execution, at a recorded tree SHA, in which every named test passes. If the switch stays off, the lane reports to the leader; neither the lane nor this document decides a waiver.
3. **Two-cell adoption.** Each criterion carries a RED-now cell and a green path naming the milestone that flips it (`verification-completeness.md` §2). Criteria that cannot be red by construction are labeled **regression-guard** and carry a positive control.
4. **Manual criteria are Gap-class.** AC-MBM-014 and -015 need an interactive terminal; they are not recorded as passes until a person performs them, and they are not release-blocking (`verification-completeness.md` §2.1, undecidable disposition).

## §D AC matrix

| AC | Requirements | RED-now (tree `802a72235`) | GREEN (milestone) |
|---|---|---|---|
| AC-MBM-001 | REQ-MBM-001, REQ-MBM-015 | `claude plugin validate mods/moai-board` → exit 1 "File not found" | exit 0 and a `hooks:` line naming `command.run{command=moai-board}` (M1) |
| AC-MBM-002 | REQ-MBM-013 | same command exit 1; forbidden-call count `0`, control count `0` | forbidden `0` and control `1` (M2-M4) |
| AC-MBM-003 | REQ-MBM-002, REQ-MBM-013 | `grep -rn 'process\.run(' mods/moai-board …` → exit 2 | exactly 1 match, in `hooks/data.ts` (M2) |
| AC-MBM-004 | REQ-MBM-002, REQ-MBM-003, REQ-MBM-004, REQ-MBM-005 | `claude plugin test mods/moai-board` exit 1 "no such plugin folder"; `(pass) pick:` count `0` | 7 passes, 0 fails (M3) |
| AC-MBM-005 | REQ-MBM-002, REQ-MBM-006 | same runner exit 1; `(pass) poll:` count `0` | 5 passes (M2) |
| AC-MBM-006 | REQ-MBM-007 | same; `(pass) queue:` count `0` | 4 passes (M2) |
| AC-MBM-007 | REQ-MBM-008 | same; `(pass) queue-fallback:` / `queue-text:` count `0` | 4 + 1 passes (M2) |
| AC-MBM-008 | REQ-MBM-009 | same; `(pass) lanes:` count `0` | 4 passes (M2, M3) |
| AC-MBM-009 | REQ-MBM-010 | same; `(pass) failsoft:` count `0` | 5 passes (M2) |
| AC-MBM-010 | REQ-MBM-011 | same; `(pass) specs:` count `0` | 3 passes (M4) |
| AC-MBM-011 | REQ-MBM-012 | same; `(pass) specs-guard:` / `specs-chunk:` count `0` | 4 + 4 passes (M4) |
| AC-MBM-012 | REQ-MBM-015, REQ-MBM-001 | same; `(pass) board:` count `0` | 3 passes on terminal and desktop (M3) |
| AC-MBM-013 | REQ-MBM-014 | `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts` → exit 1, stdout empty | exit 0 naming the new rule (M5); three regression guards stay as measured |
| AC-MBM-014 | REQ-MBM-001, REQ-MBM-015 | not performed | manual launch observed (operator, interactive terminal) |
| AC-MBM-015 | REQ-MBM-003, REQ-MBM-004, REQ-MBM-005 | not performed (a lane cannot mutate the queue, M-10) | manual live pick observed (operator, non-lane session) |

## §D.1 AC detail

### AC-MBM-001 — the mod validates and registers `/moai-board`
- **Given** `mods/moai-board/` with a manifest, a `hooks/hooks.json` naming one module, and the hooks module,
- **When** `claude plugin validate mods/moai-board` runs,
- **Then** it exits 0, prints no error, and its `./register.tsx hooks:` line contains `command.run{command=moai-board}` and `session.start`; no registered name in the source contains a colon.
- Judging: `claude plugin validate mods/moai-board` then `grep -rn "name: '[^']*:" mods/moai-board/hooks` → no match.
- **RED-now** — command `claude plugin validate mods/moai-board`; exit **1**; stdout:
  ```
  Validating plugin manifest: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1436/mods/moai-board

  ✘ Found 1 error:

    ❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1436/mods/moai-board

  ✘ Validation failed
  ```
- Mutant: a module that registers `moai:board` → the `hooks:` line names it and the colon grep matches. A `modules` string instead of an array → validate exits 1 (observed on the stub, M-8).

### AC-MBM-002 — the module calls only the allowed `$` nouns
- **Given** the finished hooks module,
- **When** `claude plugin validate mods/moai-board` prints its `calls:` line,
- **Then** the line contains `$.process.run`, `$.ui.ask`, `$.ui.open`, `$.clock.every`, `$.fs.read`, `$.fs.stat`, `$.session.root`, `$.state` and `$.command.register` and **no** `$.http`, `$.fs.write`, `$.tool.call`, `$.prompt.submit`, `$.model`, `$.agent`, `$.settings`, `$.env`, `$.session.append` or `$.process.spawn`.
- Judging pair (both required): `claude plugin validate mods/moai-board | grep -c -E '\$\.(http|fs\.write|tool\.call|prompt\.submit|model|agent|settings|env|session\.append|process\.spawn)'` → `0`; and `claude plugin validate mods/moai-board | grep -c -F '$.process.run'` → `1`.
- **RED-now** — the validate command is the AC-MBM-001 cell (exit 1). Judging pair on the absent mod: forbidden count stdout `0` (pipeline exit 1), control count stdout `0` (pipeline exit 1). **The forbidden count alone is already `0` now — it is vacuous by itself; the control count (`0` now, `1` when green) is the cell that flips**, which is why the two are one criterion.
- Mutant probe (run-phase, in a scratch copy of the mod): add one `$.http.fetch(...)` call → forbidden count `1`; remove all `$.process.run` calls → control count `0`.
- Limit: `calls:` lists `$` nouns, not argv values; AC-MBM-003 and the tests cover argv.

### AC-MBM-003 — one `$.process.run` call site
- **Given** the finished source,
- **When** `grep -rn 'process\.run(' mods/moai-board --exclude-dir=.claude-plugin --exclude-dir=node_modules --exclude='*.test.*'` runs,
- **Then** it prints exactly one line, in `mods/moai-board/hooks/data.ts`, and the argv table in that file contains `gtd`, `next` only in the pick entry.
- **RED-now** — command as above; exit **2**; stderr `ugrep: warning: mods/moai-board: No such file or directory`; stdout empty.
- Mutant: a second `process.run(` anywhere in `hooks/` → two lines. The engine-written typings under `.claude-plugin/` also contain the text `process.run(`, which is why that directory is excluded.

### AC-MBM-004 — pick: confirmation, no-op paths, failure and unconfirmed handling
- **Given** the process boundary stubbed through `on('process.run', …)` returning `{ value }` and the dialog stubbed through `on('tool.call', { tool: 'AskUserQuestion' }, …)` (both forms observed working, M-9),
- **When** pick is pressed on a `queued` card,
- **Then** these seven tests pass: `pick: confirm label runs exactly one argv` (argv equals `moai gtd next <id> --expect <prefix>`), `pick: cancel runs no process`, `pick: other text runs no process`, `pick: rejected ask (no person) runs no process`, `pick: non-zero exit shows stderr and does not retry`, `pick: exit 0 with card still queued reads unconfirmed`, `pick: id or prefix failing the check offers no pick`.
- Judging: `claude plugin test mods/moai-board 2>&1 | grep -c -E '^\(pass\) pick:'` → `7`; `claude plugin test mods/moai-board 2>&1 | grep -c -E '^\(fail\)'` → `0`.
- **RED-now** — command `claude plugin test mods/moai-board`; exit **1**; stdout `claude plugin test: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1436/mods/moai-board: no such plugin folder`. Judging count: stdout `0`.
- Mutant: pick that runs before the ask resolves → `cancel runs no process` fails; a retry loop → `does not retry` fails.

### AC-MBM-005 — polling and no autonomous mutation
- **Given** a mocked clock (`mock.clock`) and stubbed processes,
- **When** the pane opens, ten intervals elapse, and the pane closes,
- **Then** five tests pass: `poll: no process while the pane is closed`, `poll: single flight`, `poll: interval never below 15000`, `poll: timer cancelled on ui.close`, `poll: only read argv across 10 ticks` (every recorded argv is in the read-only table; none contains `next`).
- Judging: `… | grep -c -E '^\(pass\) poll:'` → `5`; zero `(fail)`.
- **RED-now** — the AC-MBM-004 runner cell (exit 1); `(pass) poll:` count `0`.
- Mutant: a timer that survives `ui.close` → `timer cancelled` fails; a tick that ignores an in-flight poll → `single flight` fails.

### AC-MBM-006 — queue reduction
- **Given** payloads shaped like M-2 (`items`, `findings`, `archived`, `runtime`; states dropped/picked/hold/queued), including a synthetic payload built to ≥ 1,900,000 characters,
- **When** the reducer runs,
- **Then** four tests pass: `queue: drops dropped, archived, runtime`, `queue: summary counts` (fixture of 17 picked / 44 queued / 2 held reads `In progress 17 · Queued 44 · Held 2`), `queue: unchanged raw is not re-parsed` (a parse counter stays at 1 across two identical polls), `queue: 2 MB payload reduces` (completes and returns the right counts within the runner's own timeout; it does **not** claim engine CPU time, G-3).
- Judging: `… | grep -c -E '^\(pass\) queue:'` → `4`.
- **RED-now** — runner exit 1; count `0`. Mutant: keeping `archived` in the reduced value → `drops …` fails.

### AC-MBM-007 — truncation and invalid-JSON fallback
- **Given** a result with `isStdoutTruncated: true`, one with invalid JSON, one with an object lacking `items`, and a text list built from the M-3 shape (rows with and without `by=`/`lease=`, `\t↳` continuation lines, the `N dropped (hidden …)` footer),
- **When** the queue is read,
- **Then** `queue-fallback: truncated stdout`, `queue-fallback: invalid json`, `queue-fallback: no items array` each cause exactly one extra read of `moai gtd list --limit 0` and show the reduced-list notice; `queue-fallback: both fail is degraded` shows the degraded text; `queue-text: rows, continuation lines, footer` yields 63 rows for a 63-row fixture and none for the continuation or footer lines.
- Judging: `… | grep -c -E '^\(pass\) queue-fallback:'` → `4`; `… | grep -c -E '^\(pass\) queue-text:'` → `1`.
- **RED-now** — runner exit 1; both counts `0`. Mutant: a parser that treats the footer as a card → the 63-row assertion fails.

### AC-MBM-008 — Lanes tab
- **Given** factory-status and session fixtures shaped like M-4 and M-5,
- **When** the Lanes view is built,
- **Then** `lanes: groups by owner, drops legacy and completed`, `lanes: lease expired marker` (the `lease_expired: true` card carries the words "lease expired"), `lanes: sessions window, order, cap` (heartbeat within 24 h only, newest first, at most 20), `lanes: never says alive or dead` (the rendered text contains neither word) pass.
- Judging: `… | grep -c -E '^\(pass\) lanes:'` → `4`.
- **RED-now** — runner exit 1; count `0`. Mutant: a view that prints "live" for a recent heartbeat → the wording test fails.

### AC-MBM-009 — fail-soft
- **Given** the stubbed process boundary rejects (cannot start), returns exit 2, rejects at the timeout, and returns unparseable stdout,
- **When** each tab refreshes,
- **Then** `failsoft: moai not found`, `failsoft: non-zero exit`, `failsoft: timeout`, `failsoft: unparseable output keeps last good data dimmed`, `failsoft: no hook rejects` pass: each shows a one-line cause, no dispatch rejects, and a previously good dataset stays visible dimmed with its age.
- Judging: `… | grep -c -E '^\(pass\) failsoft:'` → `5`.
- **RED-now** — runner exit 1; count `0`. Mutant: dropping the guard around the render hook → `no hook rejects` fails.

### AC-MBM-010 — SPEC list
- **Given** the text of M-6 (header, rule, 1,011 `SPEC-` rows, `_archive unknown`),
- **When** it is parsed,
- **Then** `specs: list rows, header/rule/_archive dropped` yields 1,011 rows for that fixture, `specs: default filter` yields draft + in-progress (25 + 15 = 40 for the M-6 counts), `specs: pages of 15` yields 3 pages for 40 rows.
- Judging: `… | grep -c -E '^\(pass\) specs:'` → `3`.
- **RED-now** — runner exit 1; count `0`. Mutant: counting the header row as a SPEC → 1,012 and the first test fails.

### AC-MBM-011 — SPEC read guards and chunking
- **Given** ids and paths including `../x`, `/etc/passwd`, `spec-lower-001`, `SPEC-A-001/../../x`, a symlink whose `realPath` lies outside the base (stubbed `fs.stat`), and a 133,596-character document with fences,
- **When** a SPEC is opened,
- **Then** `specs-guard: id pattern` (the four bad ids are refused; `SPEC-MOAI-BOARD-MOD-001` accepted), `specs-guard: file allow-list` (only the six names), `specs-guard: real path outside base denied`, `specs-guard: nothing read before both stats` (no `fs.read` is recorded when either stat is missing or outside) pass; and `specs-chunk: each chunk ≤ 9000`, `specs-chunk: splits at blank lines`, `specs-chunk: fences closed and reopened`, `specs-chunk: 12-chunk cap with notice` (the 133,596-character fixture shows 12 chunks and a notice naming the file) pass.
- Judging: `… | grep -c -E '^\(pass\) specs-guard:'` → `4`; `… | grep -c -E '^\(pass\) specs-chunk:'` → `4`.
- **RED-now** — runner exit 1; both counts `0`. Mutant: a guard comparing the unresolved spelling → `real path outside base denied` fails; a chunker that cuts inside a fence → `fences closed and reopened` fails.

### AC-MBM-012 — the pane draws
- **Given** the `Pane` component mounted through the plugin on `terminal` and on `desktop` (`$.ui.mount({ plugin: 'moai-board', surface, component: 'Pane', requestId: 'moai-board', props })`) with queue fixtures in state,
- **When** the view renders,
- **Then** `board: pane draws on terminal and desktop` (tab buttons keyed for Queue, Lanes, SPEC; the summary line reads `In progress 17 · Queued 44 · Held 2`; the engine accepts the tree — an invalid tree would not draw), `board: tab buttons switch the view`, `board: refresh button re-runs the read argv` pass on both surfaces.
- Judging: `… | grep -c -E '^\(pass\) board:'` → `3`.
- **RED-now** — runner exit 1; count `0`. Mutant: an element the surface table lacks (for example `Svg`, which the terminal table lacks) → the engine refuses the tree and `pane draws` fails.
- Limit: paint, docking, scrolling and hotkey delivery are not exercised (G-1).

### AC-MBM-013 — packaging
- **Given** the finished mod and the `.gitignore` edit,
- **When** these run: `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts` (a), `git check-ignore -v mods/moai-board/hooks/register.tsx` (b), `git ls-files -- 'internal/template/templates/*moai-board*'` (c), `grep -n mods .goreleaser.yml` (d),
- **Then** (a) exits 0 and names the new rule; (b) exits 1 (the rule does not swallow the source); (c) prints nothing; (d) exits 1.
- **RED-now (a)** — command `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts`; exit **1**; stdout empty.
- (b), (c), (d) are **regression-guards**, green today and incapable of red until someone mispaces the mod: (b) exit 1 today (measured, M-11); (c) exit 0, stdout empty today, with control `git ls-files -- 'internal/template/templates/.claude/rules/moai/core/zone-registry.md'` printing that path, exit 0; (d) exit 1 today. Mutant: a file added under `internal/template/templates/` named with `moai-board` makes (c) print its path.

### AC-MBM-014 — manual: first interactive launch (Gap-class, not release-blocking)
- **Given** an operator in an interactive terminal,
- **When** they run `claude --plugin-dir <absolute path to mods/moai-board>` and type `/moai-board`,
- **Then** the pane opens with three tabs, the summary line shows non-zero counts, refresh updates it, closing the pane stops the poll, and an unplugged `moai` (PATH without it) shows the plain-language error while the session keeps working. The result is recorded in `progress.md` §E.2 with the date and the `claude` build; until then it is **unobserved**.

### AC-MBM-015 — manual: live pick (Gap-class, not release-blocking)
- **Given** a non-lane session and one throwaway card the operator admitted for the purpose,
- **When** the operator presses pick, answers `Cancel` (card must stay queued), presses again and answers `Pick`,
- **Then** the first leaves the queue unchanged; the second runs `moai gtd next <id> --expect <prefix>`, and `moai gtd list` then shows the card `picked`. In a lane session the CLI's lane-boundary refusal (M-10) appears in the pane. Until performed it is **unobserved**; it mutates the real queue and is never run automatically.

## §E Edge cases (covered by the tests above or listed as Gaps)

Empty queue (a state, not an error); only held cards; a 2,198-character card text with fences; hot reload while a poll is in flight (the poll result writes state, the timer is re-created by the new environment); `-p` run (the ask rejects → no-op); `/cd` moving the session root (the SPEC tab reads the new root); a pane closed while a poll is in flight (the result is stored, nothing is drawn); the `vscode` and `mobile` surfaces (G-9, unexercised).

## §F Definition of Done

1. AC-MBM-001..013 evidence in `progress.md` §E.2: command, verbatim output, exit code, tree SHA, `claude` and `moai` builds.
2. Every test-bearing criterion observed on an executing runner (§A.2), or the refusal reported to the leader.
3. `git status --short` shows only files under `mods/moai-board/`, `.gitignore`, and this SPEC's directory.
4. AC-MBM-014 and -015 listed as Gap-class in the completion report, not as passes.
5. Commit messages carry `t1436`; the verdict lives at `.moai/reports/t1436/verdict.md` (written by the lane).
