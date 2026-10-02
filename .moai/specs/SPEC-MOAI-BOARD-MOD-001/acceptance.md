# Acceptance — SPEC-MOAI-BOARD-MOD-001

**Measurement tree (document-level pin, binds every criterion that carries no pin of its own):** `5f6c7d343eeb1b36b647c88a8b6501b94e9e0969`, branch `WT-moai-board-mod` (revision 0.2.0; the 0.1.0 cells were measured at `802a72235` and every cell below was **re-run at this tree**). Judging builds: `claude 2.1.287`, `moai v3.2.0-rc.25`, `bun 1.4.2` (tool provenance, `verification-claim-integrity.md` §2.2). Every RED-now cell is a command run in this session; its stdout is verbatim (bounded) and its exit code is its own field. Paths in stdout are the worktree's absolute paths.

## §A Dispositions that apply to every criterion

1. **Two runners, labeled.** A test-bearing criterion is either **pure** — judged under `bun test`, switch-independent, covering parsers, reducers, filters, the chunker, guards and argv builders — or **engine** — judged under `claude plugin test`, covering anything that needs `on`, `mock.clock`, a `$.ui.ask` stub, hook dispatch or `mount`. **Pure evidence is lane/developer evidence only**: `bun` is a user-local binary and no CI or tooling file of this repository references it (M-15, G-13); it proves the pure functions, never hook dispatch or paint. The engine runner stays the authority for the engine criteria.
2. **A refused engine runner is UNOBSERVED.** If `claude plugin test` prints `hooks modules are turned off in this process: the rollout switch served off` (every run after the first four of this session, `--help` included, M-13), the engine criteria — AC-MBM-004b, -005b, -009b, -012 — are **UNOBSERVED-until-runner-executes**: not pass, not fail, **and no RED is recorded for them**, because while the switch is off the runner returns exit 1 for an absent mod and would for a finished one. Their first RED is the first run that executes tests; it is recorded when it can be observed, never reconstructed. Each attempt is recorded with its output.
3. **Closure.** An engine criterion closes only on one executing run, at a recorded tree SHA, in which every named test passes. **While any engine criterion is UNOBSERVED the SPEC stays open**, unless the leader or operator waives it in writing; neither the lane nor this SPEC decides a waiver (decision-index Q7).
4. **Judging recipes.** *Pure:* `rm -f /tmp/mbm-junit.xml`, then `bun test mods/moai-board/tests/pure/ --reporter=junit --reporter-outfile=/tmp/mbm-junit.xml` (must exit 0), then for each prefix `grep -c '<testcase name="<prefix>' /tmp/mbm-junit.xml` against the expected count and `grep -c '<failure' /tmp/mbm-junit.xml` → `0`. (In a non-TTY run bun prints no `(pass)` lines — observed, M-15 — so the junit file is the record.) *Engine:* `claude plugin test mods/moai-board 2>&1 | grep -c -E '^\(pass\) <prefix>'` against the expected count and `… | grep -c -E '^\(fail\)'` → `0`. A count of 0 is not a pass: an empty sweep asserts nothing (`verification-completeness.md` §1.1); bun exits 1 and the engine runner exits 1 when no test file exists (M-15, M-9).
5. **Two-cell adoption.** Each criterion carries a RED-now cell and a green path naming the milestone that flips it (`verification-completeness.md` §2). Criteria that cannot be red by construction are labeled **regression-guard** and carry a positive control.
6. **Manual criteria are Gap-class.** AC-MBM-014 and -015 need an interactive terminal; they are not recorded as passes until a person performs them, and they are not release-blocking (`verification-completeness.md` §2.1, undecidable disposition).

## §D AC matrix

| AC | Requirements | Runner | RED-now (tree `5f6c7d343`) | GREEN (milestone) |
|---|---|---|---|---|
| AC-MBM-001 | REQ-MBM-001, REQ-MBM-015 | validate | `claude plugin validate mods/moai-board` → exit 1 "File not found" | exit 0 and a `hooks:` line naming `command.run{command=moai-board}` (M1) |
| AC-MBM-002 | REQ-MBM-013 | validate | same command exit 1; forbidden-call count `0`, control count `0` | forbidden `0` and control `1` (M2-M4) |
| AC-MBM-003 | REQ-MBM-002, REQ-MBM-013 | structural grep | `grep -rn …` over `mods/moai-board` → exit 2 for each of three greps | one `process.run(` line, one `'next'` line, one `buildPickArgv(` line (M2, M3) |
| AC-MBM-004a | REQ-MBM-002, REQ-MBM-003, REQ-MBM-004, REQ-MBM-005 | pure (bun) | bun exit 1 "did not match any test files"; `pick-pure:` count: no junit file (exit 2) | 5 tests, 0 failures (M3) |
| AC-MBM-004b | REQ-MBM-003, REQ-MBM-004 | engine | **UNOBSERVED** (runner refused) | 4 tests, 0 failures, on an executing run (M3) |
| AC-MBM-005a | REQ-MBM-006 | pure (bun) | as AC-MBM-004a; `poll-pure:` count: no junit file | 2 tests (M2) |
| AC-MBM-005b | REQ-MBM-002, REQ-MBM-006 | engine | **UNOBSERVED** | 4 tests incl. `dispatch:` (M2) |
| AC-MBM-006 | REQ-MBM-007 | pure (bun) | as above; `queue:` count: no junit file | 4 tests (M2) |
| AC-MBM-007 | REQ-MBM-008 | pure (bun) | as above; `queue-fallback:` / `queue-text:` | 4 + 1 tests (M2) |
| AC-MBM-008 | REQ-MBM-009 | pure (bun) | as above; `lanes:` | 4 tests (M2) |
| AC-MBM-009a | REQ-MBM-010 | pure (bun) | as above; `failsoft-pure:` | 4 tests (M2) |
| AC-MBM-009b | REQ-MBM-010 | engine | **UNOBSERVED** | 2 tests (M3) |
| AC-MBM-010 | REQ-MBM-011 | pure (bun) | as above; `specs:` | 3 tests (M4) |
| AC-MBM-011 | REQ-MBM-012 | pure (bun) | as above; `specs-guard:` / `specs-chunk:` | 4 + 4 tests (M4) |
| AC-MBM-012 | REQ-MBM-015, REQ-MBM-001 | engine | **UNOBSERVED** | 3 tests on terminal and desktop (M3) |
| AC-MBM-013 | REQ-MBM-014 | git / grep | `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts` → exit 1, stdout empty | exit 0 naming the new rule (M5); three regression guards stay as measured |
| AC-MBM-014 | REQ-MBM-001, REQ-MBM-015 | manual | not performed | manual launch observed (operator, interactive terminal) |
| AC-MBM-015 | REQ-MBM-003, REQ-MBM-004, REQ-MBM-005 | manual | not performed (a lane cannot mutate the queue, M-10) | manual live pick observed (operator, non-lane session) |

**The shared pure RED-now cell** (every `pure (bun)` row): command `bun test mods/moai-board/tests/pure/ --reporter=junit --reporter-outfile=/tmp/mbm-junit.xml` after `rm -f /tmp/mbm-junit.xml`; exit **1**; stdout:
```
bun test v1.4.2 (744846f84)
The following filters did not match any test files in --cwd="/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1436":
 mods/moai-board/tests/pure/
6573 files were searched [77.00ms]

note: Tests need ".test", "_test_", ".spec" or "_spec_" in the filename (ex: "MyApp.test.ts")
```
and the judging count `grep -c '<testcase name="queue:' /tmp/mbm-junit.xml` → stderr `ugrep: warning: /tmp/mbm-junit.xml: No such file or directory`, exit **2**. The reason for the red is the one the work removes: no pure test file exists.

**The engine rows have no RED-now.** `claude plugin test mods/moai-board` re-run at this tree printed `claude plugin test: hooks modules are turned off in this process: the rollout switch served off, and a plugin's tests run only while it is on`, exit 1, and so did `claude plugin test --help`. That output is not a RED for these criteria (§A.2); the stale 0.1.0 cell ("no such plugin folder") no longer reproduces and was removed.

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
- **RED-now** — the validate command is the AC-MBM-001 cell (exit 1). Judging pair on the absent mod: forbidden count stdout `0`, pipeline exit 1; control count stdout `0`, pipeline exit 1. **The forbidden count alone is already `0` now — vacuous by itself; the control count (`0` now, `1` when green) is the cell that flips**, which is why the two are one criterion.
- Mutant probe (run-phase, scratch copy): add one `$.http.fetch(...)` call → forbidden count `1`; remove all `$.process.run` calls → control count `0`. (The plan-audit observed this on a scratch stub and that the engine refuses `const h = $.http` and `$[noun]` evasions.)
- Limit: `calls:` lists `$` nouns, not argv values; AC-MBM-003 and the tests cover argv.

### AC-MBM-003 — structural call-site checks (one `process.run`, one `next`, one pick builder call)
- **Given** the finished source under `mods/moai-board/hooks`,
- **When** these run: (i) `grep -rn 'process\.run(' mods/moai-board --exclude-dir=.claude-plugin --exclude-dir=node_modules --exclude='*.test.*' --exclude='*.spec.*'`; (ii) `grep -rn -E "['\"]next['\"]" mods/moai-board/hooks`; (iii) `grep -rn 'buildPickArgv(' mods/moai-board/hooks`,
- **Then** (i) prints exactly one line, in `hooks/data.ts`; (ii) prints exactly one line, in the pick entry of the argv table in `hooks/data.ts`; (iii) prints exactly one line, the call inside `onPickPress` in `hooks/register.tsx` (the builder is defined as `const buildPickArgv = (` so the definition and the import do not match `buildPickArgv(`).
- **RED-now** — (i) `grep -rn 'process\.run(' mods/moai-board --exclude-dir=.claude-plugin --exclude-dir=node_modules --exclude='*.test.*' --exclude='*.spec.*'`; exit **2**; stdout empty; stderr `ugrep: warning: mods/moai-board: No such file or directory`. (ii) `grep -rn -E "['\"]next['\"]" mods/moai-board/hooks`; exit **2**; stdout empty; stderr `ugrep: warning: mods/moai-board/hooks: No such file or directory`. (iii) `grep -rn 'buildPickArgv(' mods/moai-board/hooks`; exit **2**; stdout empty; stderr `ugrep: warning: mods/moai-board/hooks: No such file or directory`.
- Mutants: a second `process.run(` anywhere in `hooks/` → (i) two lines; a `session.start`, `command.run` or render hook that calls `buildPickArgv(` → (iii) two lines (this is the structural half of REQ-MBM-002's "no pick from session.start, render or command hooks"). The engine typings under `.claude-plugin/` also contain `process.run(`, which is why that directory is excluded.
- Limit: (iii) counts call sites, not which hook holds them, and a call through an alias (`const b = buildPickArgv; b(…)`) evades it; the `'next'` literal of (ii) and the engine test `dispatch:` (AC-MBM-005b) cover those.

### AC-MBM-004a — pick, pure part (argv, id/prefix check, confirm label, failure and unconfirmed outcomes)
- **Given** the pure functions of `hooks/data.ts`,
- **When** they are called with fixtures,
- **Then** five tests pass: `pick-pure: argv is moai gtd next id expect prefix` (the list is exactly `['moai','gtd','next',<id>,'--expect',<prefix>]`), `pick-pure: id or prefix failing the check offers no pick` (an id outside `^[A-Za-z0-9][A-Za-z0-9_-]{0,31}$`, a prefix beginning with `-`), `pick-pure: only the confirm label confirms` (`Pick` confirms; `Cancel`, `pick`, a leading-space ` Pick` and the empty string do not), `pick-pure: non-zero exit shows 400 chars of stderr and does not retry`, `pick-pure: exit 0 with card still queued reads unconfirmed`.
- Judging: pure recipe (§A.4), prefix `pick-pure:` → `5`, `<failure` → `0`.
- **RED-now** — the shared pure cell above.
- Mutant: an argv builder that appends a shell string → the first test fails; `isConfirmed` using `includes` → the confirm-label test fails.

### AC-MBM-004b — pick, engine part (confirmation flow) — UNOBSERVED
- **Given** the process boundary stubbed through `on('process.run', …)` returning `{ value }` and the dialog stubbed through `on('tool.call', { tool: 'AskUserQuestion' }, …)` (both forms observed working on a stub while the runner executed, M-9),
- **When** pick is pressed on a `queued` card,
- **Then** four tests pass: `pick: confirm label runs exactly one argv`, `pick: cancel runs no process`, `pick: other text runs no process`, `pick: rejected ask runs no process` (the stubbed ask rejects, as in a `-p` run).
- Judging: engine recipe, prefix `pick:` → `4`, zero `(fail)`.
- **RED-now: UNOBSERVED** (§A.2). Mutant: pick that runs before the ask resolves → `cancel runs no process` fails.

### AC-MBM-005a — polling, pure part
- **Given** the pure interval and single-flight helpers,
- **When** they are called,
- **Then** `poll-pure: interval is clamped to the 15000 floor` (any requested value below 15000, including 0 and negatives, returns 15000; larger values pass through) and `poll-pure: a second poll while one is in flight is refused` pass.
- Judging: pure recipe, prefix `poll-pure:` → `2`.
- **RED-now** — the shared pure cell. Mutant: a clamp that returns the requested value → the first test fails.

### AC-MBM-005b — polling timers and dispatch, engine part — UNOBSERVED
- **Given** a mocked clock (`mock.clock`) and stubbed processes,
- **When** the pane opens, ten intervals elapse, and the pane closes,
- **Then** four tests pass: `poll: no process while the pane is closed`, `poll: timer cancelled on ui.close`, `poll: only read argv across 10 ticks` (every recorded argv is in the read-only table; none contains `next`), and `dispatch: no pick argv from session.start, command.run or render` (the argv recorded across a `session.start` dispatch, a `/moai-board` `command.run`, and a mounted Pane render contains no `next`; this is the test REQ-MBM-002's hook clause needs).
- Judging: engine recipe, prefixes `poll:` → `3` and `dispatch:` → `1`.
- **RED-now: UNOBSERVED** (§A.2). Mutants: a timer that survives `ui.close` → `timer cancelled` fails; a `session.start` hook that runs the pick argv → `dispatch:` fails (and AC-MBM-003 (iii) shows a second call line).

### AC-MBM-006 — queue reduction (pure)
- **Given** payloads shaped like M-2 (`items`, `findings`, `archived`, `runtime`; states dropped/picked/hold/queued), including a synthetic payload built at run time to ≥ 1,900,000 characters (not committed),
- **When** the reducer runs,
- **Then** four tests pass: `queue: drops dropped, archived, runtime`, `queue: summary counts` (a fixture of 17 picked / 44 queued / 2 held reads `Picked 17 · Queued 44 · Held 2`), `queue: unchanged raw is not re-parsed` (a parse counter stays at 1 across two identical inputs), `queue: 2 MB payload reduces` (completes and returns the right counts; it does **not** claim engine CPU time, G-3).
- Judging: pure recipe, prefix `queue:` → `4`.
- **RED-now** — the shared pure cell. Mutant: keeping `archived` in the reduced value → `drops …` fails; labeling the count "In progress" → `summary counts` fails.

### AC-MBM-007 — truncation and invalid-JSON fallback (pure)
- **Given** a result with `isStdoutTruncated: true`, one with invalid JSON, one with an object lacking `items`, and a text list built from the M-3 shape (rows with and without `by=`/`lease=`, `\t↳` continuation lines, the `N dropped (hidden …)` footer),
- **When** the pure read planner and text parser run,
- **Then** `queue-fallback: truncated stdout`, `queue-fallback: invalid json`, `queue-fallback: no items array` each make the planner return the `moai gtd list --limit 0` argv exactly once and flag the reduced-list notice; `queue-fallback: both fail is degraded` returns the degraded state; `queue-text: rows, continuation lines, footer` yields 63 rows for a 63-row fixture and none for the continuation or footer lines.
- Judging: pure recipe, `queue-fallback:` → `4`, `queue-text:` → `1`.
- **RED-now** — the shared pure cell. Mutant: a parser that treats the footer as a card → the 63-row assertion fails.

### AC-MBM-008 — Lanes tab (pure)
- **Given** factory-status and session fixtures shaped like M-4 and M-5 and a fixed `now`,
- **When** the lane reducers and the row-text builder run,
- **Then** `lanes: groups by owner, drops legacy and completed`, `lanes: lease expired marker` (the `lease_expired: true` card's text contains "lease expired"), `lanes: sessions window, order, cap` (heartbeat within 24 h only, newest first, at most 20), `lanes: never says alive or dead` (no row text contains either word) pass.
- Judging: pure recipe, prefix `lanes:` → `4`.
- **RED-now** — the shared pure cell. Mutant: a builder that prints "live" for a recent heartbeat → the wording test fails.

### AC-MBM-009a — fail-soft, pure part (cause classification)
- **Given** run results shaped as: cannot start (rejected), exit 2 with stderr, rejected at the timeout, unparseable stdout,
- **When** the pure classifier builds the one-line cause,
- **Then** `failsoft-pure: moai not found`, `failsoft-pure: non-zero exit`, `failsoft-pure: timeout`, `failsoft-pure: unparseable output` pass: each yields a one-line plain-language cause and, for the last, keeps the previous good dataset marked dimmed with its age.
- Judging: pure recipe, prefix `failsoft-pure:` → `4`.
- **RED-now** — the shared pure cell. Mutant: a classifier that throws on a rejected run → the first test fails.

### AC-MBM-009b — fail-soft, hook part — UNOBSERVED
- **Given** the stubbed process boundary rejects, returns exit 2, and returns unparseable output while the pane is mounted,
- **When** each tab refreshes,
- **Then** `failsoft: no hook rejects` (no dispatch rejects) and `failsoft: last good data stays visible dimmed` (the mounted tree still carries the earlier rows, dimmed) pass.
- Judging: engine recipe, prefix `failsoft:` → `2`. **RED-now: UNOBSERVED** (§A.2). Mutant: dropping the guard around the render hook → `no hook rejects` fails.

### AC-MBM-010 — SPEC list (pure)
- **Given** a text built in the shape of M-14: the header, the rule line, 1,010 SPEC-id rows (25 `draft`, 15 `in-progress`, 970 `completed`), `SPEC-GITHUB-WORKFLOW implemented`, `SPEC-I18N-001-ARCHIVED archived`, and `_archive unknown`,
- **When** it is parsed under the SPEC-id rule of spec.md §3 (`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`),
- **Then** `specs: list rows, header rule and non-ids dropped` yields exactly 1,010 rows (the two `SPEC-`-prefixed non-ids, the header, the rule and `_archive` are dropped and none is openable), `specs: default filter` yields draft + in-progress = 40 rows, `specs: pages of 15` yields 3 pages for 40 rows.
- Judging: pure recipe, prefix `specs:` → `3`. The live count the fixture mirrors: `moai spec status --list | grep -c -E '^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}[[:space:]]'` printed `1010` at `5f6c7d343` (M-14); a run-phase re-measure may differ by the SPECs added since.
- **RED-now** — the shared pure cell. Mutant: filtering by "begins with `SPEC-`" → 1,012 and the first test fails.

### AC-MBM-011 — SPEC read guards and chunking (pure)
- **Given** ids and paths including `../x`, `/etc/passwd`, `spec-lower-001`, `SPEC-A-001/../../x`, an injected `io` whose `stat` reports a `realPath` outside the base, and a 133,596-character document with fences,
- **When** a SPEC is opened through the guard and chunked,
- **Then** `specs-guard: id pattern` (the four bad ids are refused; `SPEC-MOAI-BOARD-MOD-001` accepted), `specs-guard: file allow-list` (only the six names), `specs-guard: real path outside base denied`, `specs-guard: nothing read before both stats` (the injected `read` is never called when either stat is missing or outside) pass; and `specs-chunk: each chunk at most 9000`, `specs-chunk: splits at blank lines`, `specs-chunk: fences closed and reopened`, `specs-chunk: 12-chunk cap with notice` (the 133,596-character fixture shows 12 chunks and a notice naming the file) pass.
- Judging: pure recipe, `specs-guard:` → `4`, `specs-chunk:` → `4`.
- **RED-now** — the shared pure cell. Mutant: a guard comparing the unresolved spelling → `real path outside base denied` fails; a chunker that cuts inside a fence → `fences closed and reopened` fails.

### AC-MBM-012 — the pane draws — UNOBSERVED
- **Given** the `Pane` component mounted through the plugin on `terminal` and on `desktop` (`$.ui.mount({ plugin: 'moai-board', surface, component: 'Pane', requestId: 'moai-board', props })`) with queue fixtures in state,
- **When** the view renders,
- **Then** `board: pane draws on terminal and desktop` (tab buttons keyed for Queue, Lanes, SPEC; the summary line reads `Picked 17 · Queued 44 · Held 2`; the engine accepts the tree — an invalid tree would not draw), `board: tab buttons switch the view`, `board: refresh button re-runs the read argv` pass on both surfaces.
- Judging: engine recipe, prefix `board:` → `3`. **RED-now: UNOBSERVED** (§A.2). Mutant: an element the surface table lacks (for example `Svg`, which the terminal table lacks) → the engine refuses the tree and `pane draws` fails.
- Limit: paint, docking, scrolling and hotkey delivery are not exercised (G-1).

### AC-MBM-013 — packaging
- **Given** the finished mod and the `.gitignore` edit,
- **When** these run: `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts` (a), `git check-ignore -v mods/moai-board/hooks/register.tsx` (b), `git ls-files -- 'internal/template/templates/*moai-board*'` (c), `grep -n mods .goreleaser.yml` (d),
- **Then** (a) exits 0 and names the new rule; (b) exits 1 (the rule does not swallow the source); (c) prints nothing; (d) exits 1.
- **RED-now (a)** — re-run at `5f6c7d343`: command `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts`; exit **1**; stdout empty.
- (b), (c), (d) are **regression-guards**, green today and incapable of red until someone mispaces the mod, re-run at `5f6c7d343`: (b) exit 1; (c) exit 0, stdout empty, with control `git ls-files -- 'internal/template/templates/.claude/rules/moai/core/zone-registry.md'` printing that path, exit 0; (d) exit 1. Mutant: a file added under `internal/template/templates/` named with `moai-board` makes (c) print its path.

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

1. AC-MBM-001..013 evidence in `progress.md` §E.2: command, verbatim output, exit code, tree SHA, `claude`, `moai` and (for pure criteria) `bun` builds, each labeled **pure (bun)**, **engine**, or **UNOBSERVED**.
2. **Every engine criterion (AC-MBM-004b, -005b, -009b, -012) is observed on an executing engine run, or the SPEC is not done.** A reported refusal does not close it: while any engine criterion is UNOBSERVED the SPEC stays open, unless the leader or operator waives it in writing (§A.3; neither the lane nor this SPEC decides a waiver). Pure criteria closed under bun do not substitute for them.
3. `git status --short` shows only files under `mods/moai-board/`, `.gitignore`, and this SPEC's directory.
4. AC-MBM-014 and -015 listed as Gap-class in the completion report, not as passes.
5. Commit messages carry `t1436`; the verdict lives at `.moai/reports/t1436/verdict.md` (written by the lane).
