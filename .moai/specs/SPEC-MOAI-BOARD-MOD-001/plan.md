# Plan — SPEC-MOAI-BOARD-MOD-001

Measurements are cited as `M-n` / `G-n` from `spec.md` (§1, §7); tree `802a72235`. Development mode: `tdd` (`.moai/config/sections/quality.yaml` → `development_mode: tdd`): each milestone writes its tests first and observes them red.

## §A Context

**Card.** t1436, Class C (design change). Operator constraints from the leader dispatch are authoritative: data through `$.process.run` (argv list) and `$.clock.every`; pane through `$.ui.open` plus a `ui.render` hook on `{ component: 'Pane', requestId }`; SPEC text through `$.fs.read` and `Markdown` elements; read-only first; the commands are `/moai-board` (no colon). Plugin load scope waits on card t1434, so the mod lives in a standalone directory of this repository (`spec.md` D-1).

**Tier judgment — M.** Files touched: 13 under `mods/moai-board/` (9 source and config, 4 test files) plus the root `.gitignore` = 14, inside the Tier M band (5-15). Size: an estimate of 600-900 lines including tests, inside 300-1,000 (an estimate, not a measurement). Not constitutional, no Go code, no change to any distributed template. REQ 15 of 16, AC 15 of 16. Tier L was weighed and rejected: no file-count or LOC signal reaches it, and a `design.md` / `research.md` pair would restate §1 and §B. Threshold for the plan audit: 0.80.

**Provisional parameters are named constants.** Every numeric value in §G is either a measured bound or a constant the operator's verdict (`decision-index.md`) sets at Kickoff. The tests assert the floors and properties of REQ-MBM-006/008/012, never a particular default, so M1-M4 do not depend on any open verdict.

## §B Architecture

### B.1 Files (the mod, a plugin of function hooks)

```
mods/moai-board/
  .claude-plugin/plugin.json   name moai-board, version, description, author (validate warns without one — M-8 stub)
  hooks/hooks.json             { "modules": ["./register.tsx"] }  — an array of exactly one path (M-8)
  hooks/register.tsx           register(): command + pane lifecycle + the Pane render hook + press handlers
  hooks/data.ts                argv table; the single runMoai(); result classification; queue / lane parsers and reducers
  hooks/specs.ts               SPEC list parser; id and file allow-lists; real-path guard; Markdown chunker
  hooks/view.tsx               tab row, summary line, Queue / Lanes / SPEC views
  types/index.d.ts             PluginState for the 'moai-board' plugin + shared types
  tests/pure/data.spec.ts      PURE (bun): queue/lane parsers and reducers, runner classification, pick argv/outcome, interval clamp
  tests/pure/specs.spec.ts     PURE (bun): list parser, id/file allow-lists, injected-io path guard, chunker
  tests/engine.test.ts         ENGINE (claude plugin test): hook dispatch, polling timers, pick flow with ui.ask stub, fail-soft hooks
  tests/board.test.tsx         ENGINE: Pane mounted on terminal and desktop
  tsconfig.json                the options of the typings header, include [".claude-plugin/types","hooks","types","tests"], exclude ["tests/pure"] (bun:test is not among the engine typings)
  README.md                    launch, test and validate commands; the read-only boundary
```

`$.state` keys: `view` (tab, page, selected card or SPEC, file, filters), `queue`, `lanes`, `specs`, `notice`. `plugin` and `key` are literals in source (the typings require it; `validate` lists them).

### B.2 Data flow

```
tick / refresh / tab open
  → runMoai(argv-from-table)                 sole $.process.run site, timeoutMs 20000
  → classify: ok | exit(code, stderr) | timeout | cannot-start | truncated
  → parse (pure)  → reduce (pure)            drops dropped / archived / runtime at parse time
  → raw === previous raw ? stop              no parse, no state write, no render
  → $.state.set(queue|lanes|specs)           the Pane render hook re-draws its readers
```

A render hook never writes state and never calls `$.process.run`; handlers (press, timer, command) do. UI state is read from `$.state` (a hot reload loses module variables); the previous raw stdout is the one module variable, and losing it costs one re-parse.

### B.3 The fixed argv table (REQ-MBM-002, -013)

| Name | argv | When |
|---|---|---|
| queue-json | `moai gtd list --json` | poll, refresh, queue tab |
| queue-text | `moai gtd list --limit 0` | only when queue-json is truncated/invalid (REQ-MBM-008) |
| lanes | `moai factory status --json` | Lanes tab visible, refresh |
| sessions | `moai session list --json` | Lanes tab visible, refresh |
| spec-list | `moai spec status --list` | SPEC tab open, refresh |
| **pick** | `moai gtd next <id> --expect <prefix>` | confirmed pick only — the one write-capable entry |

`<id>` must match `^[A-Za-z0-9][A-Za-z0-9_-]{0,31}$` (measured card ids all match `^t[0-9]+$`, M-2 *scratch*); `<prefix>` is the first 40 code points of the card text as polled and must not begin with `-` (measured: no live text starts with `-` and the shortest live text is 94 characters, M-2 *scratch*; a card failing either check has no pick button). No other argv is built anywhere.

### B.4 Tab model

| Tab | Content | Sources |
|---|---|---|
| Queue (hotkey 1) | summary `Picked N · Queued N · Held N` (`Picked` is the state `picked`, an operator promotion — not "in progress"); rows grouped picked → queued → hold, each group in the emitted order (M-3: JSON order = text order); a row shows id and the text cut to the body width; `open` shows the full text as chunked Markdown, `added_at`, `spec_id`; a `queued` card also shows **pick** (`p`) | queue-json, fallback queue-text |
| Lanes (2) | factory cards grouped by owner with state, stage, SPEC id and the "lease expired" marker (M-4: `t1399`); below, sessions with a heartbeat within 24 h, newest first, ≤ 20, shown as "heartbeat Xm ago" — never alive/dead (M-5: pid reuse, 144 h-old entries) | lanes, sessions |
| SPEC (3) | status filter chips (default draft + in-progress), 15 rows per page; `open` shows a file selector and the file as Markdown elements | spec-list, `$.fs.read` |

Global buttons: refresh (`r`), back (`b`). Elements are `Box`, `Text`, `Button`, `Markdown` only (present on every surface, `spec.md` §4), sized to `e.props.bodyColumns`. The command `/moai-board` registers at `session.start`, opens the pane (`$.ui.open({ id: 'moai-board', title: 'moai-board' })`) and starts the poll; a `ui.close` hook for the id cancels the timer.

### B.5 Polling policy (REQ-MBM-006, -007)

Measured cost: `moai gtd list --json` is 1.5-2.7 s wall and 1,938,159 bytes (M-2); the board needs 56,781 bytes of it (63 non-dropped items). Policy: one timer from `$.clock.every`; a tick is skipped while a poll is in flight (single flight); the interval is never below 15,000 ms (about 5.6 times the slowest measured JSON call, 2.68 s); each `$.process.run` carries `timeoutMs: 20000`, inside the 30 s default and independent of the 10 s hook budget (a `$` call in flight does not count against it, `spec.md` §4); the stdout string is compared with the previous one before any `JSON.parse`; only non-dropped items are kept in `$.state`. Lanes and sessions are fetched only while the Lanes tab is visible; the SPEC list only on tab open and refresh; nothing runs while the pane is closed.

### B.6 Error, empty and degraded states (REQ-MBM-008, -010)

| Condition | Detected by | Shown (plain language) |
|---|---|---|
| `moai` not on PATH | `$.process.run` rejects (cannot start) | "moai was not found on PATH, so the board cannot read the queue." |
| non-zero exit | `exitCode !== 0` | "`moai gtd list` failed (exit N): <first 200 chars of stderr>" |
| timeout | `$.process.run` rejects at `timeoutMs` | "`moai gtd list` took longer than 20 s and was stopped." |
| JSON too large | `isStdoutTruncated` (4 MiB cap; 46% used, G-4) | fallback to queue-text with the notice "Showing a reduced list: the full queue data is over the 4 MiB read limit." |
| invalid JSON / no `items` array | parse | same fallback; if queue-text also fails: "The queue data could not be read." |
| empty queue | `items` has no non-dropped entry | "The queue is empty." (a state, not an error) |
| stale data | last good `at` older than the interval | last good rows stay, dimmed, with "updated Xm ago" |

Every hook body runs inside one guard that turns an exception into a `notice` and returns a normal value, so a mod failure never reaches the session; the Pane render hook falls back to a one-line `Text` tree.

### B.7 SPEC tab mechanics (REQ-MBM-011, -012)

- **Listing.** `moai spec status --list`; keep rows whose first column is a SPEC id per the SPEC-id rule of spec.md §3 (one regex, stated once) and whose second column is a status word; this drops the header, the rule line, `_archive`, `SPEC-GITHUB-WORKFLOW` and `SPEC-I18N-001-ARCHIVED` (M-14: 1,010 kept at `5f6c7d343`); ignore the `Modified` column (M-6: checkout time). Statuses seen at the M-6 tree: completed 780, implemented 143, draft 25, archived 31, superseded 15, in-progress 15, rejected 2.
- **Which tree.** `moai spec status --list` follows the **cwd of the process**; the mod uses the session's (the engine default). From a worktree the list is that worktree's tree (1,010 SPEC ids at `5f6c7d343`, M-14), from the primary checkout another (678 lines, M-6). The pane header shows the session root so the operator can see which tree is read; whether to union the primary is Q4.
- **Reading.** `root = await $.session.root()`; `base = root + '/.moai/specs'`; `file` ∈ the six allowed names; id satisfies the same SPEC-id rule as the list (spec.md §3); the guard takes its `stat`/`read` as an injected `io` argument so bun can test it; `$.fs.stat(path, { resolve: true })` and `$.fs.stat(base, { resolve: true })`; the read happens only when `kind` is a regular file and the file's `realPath` starts with the base's `realPath` plus the separator; the read target is the `realPath`. `..`, absolute names, extra segments, lowercase ids and symlinks leaving the base fail before any read.
- **Chunking.** Blocks are split at blank lines outside fenced code; blocks are packed greedily to ≤ 9,000 characters (the 10,000 limit minus a 1,000 margin for re-fencing and counting UTF-16 units); a block over the limit is split at line boundaries; a split inside a fence closes it at the chunk end and reopens it (same info string) at the next start; a single line over the limit is cut at the limit without splitting a surrogate pair; at most 12 chunks per file (the measured maximum file, 133,596 bytes, needs 15), then a notice naming the file path. Measured: 2,936 of 4,129 SPEC artifacts exceed 10,000 bytes (M-7), so chunking is the common case.

### B.8 Pick flow (REQ-MBM-003 to -005)

0. **One call site.** The argv is built by `buildPickArgv(id, prefix)` in `hooks/data.ts` and that function — defined as `const buildPickArgv = (…) =>`, so that the text `buildPickArgv(` matches calls only — is called from exactly one place, `onPickPress` in `hooks/register.tsx`; `session.start`, `command.run`, the timer and the render hook never reference it (AC-MBM-003 counts the references; the engine test `dispatch:` records every argv across those dispatches).
1. Press on a `queued` card with a valid id and prefix → `$.ui.ask("Pick <id>? <first line of text, ≤120 chars>", ["Pick", "Cancel"])` (the confirm label is the constant `Pick`).
2. Rejection (dismissed, `-p` run), any text other than `Pick`, or `Cancel` → no process, no state change.
3. `Pick` → one `runMoai` of the pick argv. Non-zero exit or cannot start → first 400 chars of stderr (or the start error) into the pane and a toast; no retry. Measured refusal text from a lane (M-10) is shown as-is.
4. Exit 0 → immediate refresh; the card still reading `queued` → "pick reported success but the card still reads queued (unconfirmed)".
5. The list the operator pressed on can be stale by up to one interval; `--expect` makes the CLI refuse a pick whose card text changed (documented in `moai gtd next --help`: refuses unless the text starts with the prefix; **not exercised from a lane**, G-2).

## §C Pre-flight (re-measure at run-phase entry; a different value stops the run and reports)

Each is a plain command whose output is read, not remembered. Values below are the plan-time snapshots.

| # | Command | Snapshot at `802a72235` |
|---|---|---|
| C1 | `claude --version` | `2.1.287 (Claude Code)` |
| C2 | `claude plugin test --help` and `claude plugin validate --help` | both exist; if the build changed, re-read the typings header first |
| C3 | `moai gtd list --json` (stdout bytes, wall time) | 1,938,159 bytes; 1.46-2.68 s |
| C4 | `moai gtd list` (stdout bytes) | 39,534 bytes, 63 rows |
| C5 | `moai factory status --json` | 8,803 bytes, 17 cards |
| C6 | `moai spec status --list` run in the session's cwd | 1,014 lines in this tree |
| C7 | `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts` | exit 1 (not ignored) until M5 |
| C8 | `claude plugin test <any folder holding one trivial test>` | **expected to execute; the rollout switch may serve off (M-13)** — record the outcome either way |
| C9 | `git log --oneline -1 -- mods/` and `ls mods` | nothing yet |
| C10 | `bun --version`, then `bun test mods/moai-board/tests/pure/ --reporter=junit --reporter-outfile=/tmp/mbm-junit.xml` | `1.4.2`; exit 1 "did not match any test files" until M1 (M-15). bun is developer-local — no CI or tooling file references it |

If C8 refuses, M1 proceeds with `claude plugin validate` only; the pure criteria run under bun (C10) and are lane/developer evidence; every engine-runner criterion stays **UNOBSERVED** until the engine runner executes, and the SPEC stays open (acceptance.md §A.2, §F).

## §D Constraints (not renegotiated in run-phase)

1. Read-only except pick; the action allow-list is refresh / open / pick (REQ-MBM-002).
2. argv lists only, one `$.process.run` call site, no shell string (REQ-MBM-013).
3. No network, no file write, no tool call, no prompt submission, no model/agent/settings/env access — enforced by `claude plugin validate`'s `calls:` line (AC-MBM-002).
4. The mod is never embedded in `internal/template/templates/` and never deployed by `moai init` / `moai update` (REQ-MBM-014).
5. No change to any Go file, any `moai` command, or any rule file (`AGENTS.md` §5 scope discipline).
6. Test files are named by runner: pure tests are `*.spec.ts` under `tests/pure/` (bun; never `*.test.ts`, which the engine runner globs), engine tests are `*.test.ts` / `*.test.tsx` directly under `tests/`.
7. Every commit on the branch names card `t1436` in its message; the evidence path is `.moai/reports/t1436/verdict.md` (written by the lane, not by this plan).

## §E Self-verification (plan-phase)

Run by the plan author and recorded in `progress.md` §E.1: ID pattern check (`PASS`), ID uniqueness (`ls .moai/specs | grep -c MOAI-BOARD` → 0 before authoring), frontmatter field presence, `moai spec lint SPEC-MOAI-BOARD-MOD-001` (result in §E.1), `OutOfScopeRule` heading shape, artifact set = Tier M (spec, plan, acceptance) plus `progress.md` and `decision-index.md` (decision gate on).

## §F Milestones (ordered by likelihood of change: contracts and interfaces first, mechanics last)

**M1 — Contracts and skeleton (highest change likelihood).** `types/index.d.ts` (`PluginState` for `moai-board`: `view`, `queue`, `lanes`, `specs`, `notice`; the parsed-item types), the argv table and `runMoai` signature, `plugin.json`, `hooks.json`, `tsconfig.json`, an empty `register.tsx` that registers `/moai-board` and opens an empty pane. The first pure spec is written and observed red under bun (C10) before any parser exists. Exit: `claude plugin validate mods/moai-board` exit 0 and its `hooks:` line names `command.run{command=moai-board}`.

**M2 — Queue and Lanes data path (TDD).** Tests first, red observed: pure under bun — parsers and reducers for queue-json, queue-text, factory status, sessions; classification of runner results; interval clamp and single-flight guard; unchanged-raw skip; a 2 MB synthetic payload; engine under `claude plugin test` — polling timers, timer cancel on `ui.close`, the `dispatch:` and `poll:` tests. Exit: the named tests of AC-MBM-005a, -006, -007, -008, -009a pass under bun; AC-MBM-005b is observed only when the engine runner executes.

**M3 — Pane UI and pick (TDD).** The Pane render hook and views for Queue and Lanes; tab row, summary, row `open` and detail; the pick flow of §B.8 with `$.ui.ask` stubbed through `tool.call` (M-9); the fail-soft hooks. Exit: AC-MBM-004a passes under bun; AC-MBM-004b, -009b and -012 are observed only when the engine runner executes (terminal and desktop).

**M4 — SPEC tab (TDD).** List parser and filters, id and file allow-lists, real-path guard with a stubbed `fs.stat` (`realPath` outside the base is denied), chunker properties (≤ 9,000, blank-line boundaries, fences closed and reopened, 12-chunk cap with notice) — all pure, under bun. Exit: AC-MBM-010, -011 named tests pass under bun.

**M5 — Packaging and handoff.** Add `mods/*/.claude-plugin/types/` to `.gitignore`; write `README.md` (launch: `claude --plugin-dir <absolute path to mods/moai-board>`; test and validate commands; the read-only boundary and the pick confirmation); re-run the regression guards of AC-MBM-013; leave the two manual items (AC-MBM-014, -015) as a checklist for an operator in an interactive, non-lane session. Exit: AC-MBM-001..013 evidence recorded in `progress.md` §E.2 with commands and verbatim output, each labeled pure (bun), engine, or UNOBSERVED.

## §G Parameters

| Constant | Value | Basis |
|---|---|---|
| `POLL_MIN_MS` (floor) | 15,000 | REQ-MBM-006; ≈ 5.6 × the slowest measured JSON call (2.68 s, M-2) |
| `POLL_INTERVAL_MS` | set by Q3 at Kickoff | EVIDENCE-NEEDED (G-3); tests assert only the floor |
| `CMD_TIMEOUT_MS` | 20,000 | inside the 30 s engine default; > 7 × the slowest measured call |
| `CHUNK_MAX` | 9,000 | Markdown limit 10,000 (M-12) minus re-fence margin |
| `MAX_CHUNKS` | 12 | bounds one view to ≈ 108,000 characters; the measured largest file needs 15 |
| `PAGE_SIZE` | 15 | pane height; not measured against a surface (G-1) |
| `HEARTBEAT_WINDOW_H` / `SESSION_CAP` | 24 / 20 | M-5: 35 sessions within 24 h; the cap keeps the tab short |
| `STDERR_SHOW` | 400 | enough for the 272-byte measured refusal (M-10) |
| `EXPECT_PREFIX_LEN` | 40 code points | live text minimum 94 (M-2 *scratch*) |

## §H Risks

| Risk | Effect | Handling |
|---|---|---|
| Early-access API (G-10) | a later build renames a noun | `validate` and `test` on the build in use; names kept in one place per file |
| Rollout switch off (M-13, G-11) | tests cannot run, or the mod loads nothing | AC disposition (acceptance.md §A); never reported as pass |
| JSON path reaches the 4 MiB cap (G-4) | queue tab loses detail | REQ-MBM-008 fallback; slimmer `--json` flag is a separate card (out of scope) |
| Parse cost in the isolate (G-3) | hook budget (10 s of own time) overrun | one parse per changed payload; reduce immediately; measured on a 2 MB payload in test |
| Pick from the wrong session | CLI refuses in a lane (M-10) | shown as-is; Q5 |
| `$.state` size (G-8) | host rejects a large value | only 63 live items kept; value ≈ 57 KB |
| SPEC tab shows a different tree than expected (M-6) | operator reads stale SPECs | session root shown in the pane header; Q4 |
| Engine-written typings land in the repo (G-6) | untracked 747 KB file | `.gitignore` rule in M5 |

## §I Anti-patterns

- Polling while the pane is closed, or two polls at once.
- Keeping raw payloads, `archived` or `runtime` in `$.state`.
- A second `$.process.run` call site, or an argv assembled from a template string.
- Reading a SPEC path before both `stat` calls resolved, or reading the unresolved spelling.
- Reporting an AC as passed because the runner refused or printed nothing (acceptance.md §A).
- Presenting `bun test` results as engine evidence or as CI evidence (spec.md G-13), or naming a pure test file `*.test.ts`.
- A colon in a command name; an `import()` in a hooks module (it does not load).

## §J Cross-references

`.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act · `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier · `.claude/rules/moai/development/verification-completeness.md` §1-§2 (observed failure, two-cell adoption) · `.claude/rules/moai/core/verification-claim-integrity.md` §1, §2.2 (tool provenance: the `claude` build is `2.1.287`, the `moai` build `v3.2.0-rc.25`) · `SPEC-KANBAN-BOARD-001` (the board model this mod displays; no code shared).
