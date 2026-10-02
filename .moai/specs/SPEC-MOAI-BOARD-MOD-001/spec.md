---
id: SPEC-MOAI-BOARD-MOD-001
title: "moai-board: read-only queue / lane / SPEC side-panel mod for Claude Code (MVP)"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: "mods/moai-board"
lifecycle: spec-anchored
tags: "claude-code-mod, plugin, pane, kanban, queue, lanes, spec-viewer, read-only, fail-soft"
tier: M
related_specs: [SPEC-KANBAN-BOARD-001]
---

# SPEC: moai-board — read-only queue / lane / SPEC side-panel mod (MVP)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-02 | manager-spec | Initial plan-phase authoring for card t1436 (Class C). Design source: report proposal "moai-board: queue/lane/SPEC side panel" (High). Tier M. Measurements taken at tree `802a72235`. |

## 1. Problem and measured basis

The operator reads the queue, the factory lanes and SPEC documents through three separate terminal commands and re-types them to follow progress. Claude Code 2.1.287 can host a **mod** — a plugin of function hooks that runs inside the session and draws a pane. This SPEC defines an MVP mod, `/moai-board`, that shows the three views in one side panel and offers three actions only: **refresh**, **open**, **pick**.

Everything below was measured in this session at tree `802a72235536958ada5b7cd5876a168e4b8c325f` (branch `WT-moai-board-mod`, clean at start) with the installed `moai v3.2.0-rc.25` and `claude 2.1.287`. Figures are a snapshot of a moving queue, not constants; run-phase re-measures them (plan.md §C). Rows marked *scratch* were computed by a throwaway script outside the repository over the command's captured output; the plain command is the evidence, the script only counted.

| # | Command (argv, no shell) | Observed | Notes |
|---|---|---|---|
| M-1 | `claude --version` | `2.1.287 (Claude Code)`; `claude plugin test` and `claude plugin validate` both exist | the typings header reads "Written by Claude Code 2.1.287" |
| M-2 | `moai gtd list --json` | exit 0; stdout **1,938,159 bytes**, one line; stderr 164 bytes (a notice that the SQLite store answers); wall 1.96 s and 1.46 s on two runs (2.68 s on a third) | top-level keys: `project_uuid, version, last_seq, items, findings, archived, runtime`. *scratch* bytes per key: `items` 194,613 (152 items), `findings` 1,651 (10), `archived` 1,524,666 (998), `runtime` 247,030. `items` by state: dropped 89, picked 17, hold 2, queued 44. The 63 non-dropped items total 56,781 bytes; their text is 94-732 chars (max over all items 2,198). **`items` is 10% of the payload; 79% is `archived`, which the board never shows.** |
| M-3 | `moai gtd list` (text form) | exit 0; stdout **39,534 bytes**, 78 lines; wall 1.14 s | 63 rows (`id<TAB>state<TAB>[by=… lease=…<TAB>]text`), 14 `\t↳` continuation lines, 1 footer `89 dropped (hidden — see: moai todo list --dropped)`. *scratch:* id+state set **identical** to the JSON's 63 non-dropped items and in the **same order**; no live text contains a newline or tab. Default `--limit` is 100; `--limit 0` is unbounded |
| M-4 | `moai factory status --json` | exit 0; 8,803 bytes; wall 2.10 s; `"unavailable": []` | 17 cards: 10 legacy `completed`, 5 `assigned`, 1 `plan`, 1 `picked`. Per-card keys include `run_id, card_id, state, stage, owner, lease_holder, lease_expires_at, lease_expired, decision_gate, spec_id, legacy`. One non-legacy card (`t1399`, owner `lane-3`) reads `lease_expired: true`. `moai factory runs` has **no `--json` form** (text only; 13 rows, one `active live`) |
| M-5 | `moai session list --json` | exit 0; 30,132 bytes; wall 0.18 s; **95 entries** | *scratch:* heartbeat ≤ 15 min: 15, ≤ 60 min: 23, ≤ 24 h: 35, oldest 144 h. The same `pid` appears under several `session_id`s (e.g. 40177), so a pid cannot tell live from dead. Keys: `session_id, spec_id, phase, started_at, last_heartbeat, pid, host, cwd` |
| M-6 | `moai spec status --list` | exit 0; 65,886 bytes, 1,014 lines; wall 0.31 s | header + rule + 1,012 rows; one row is `_archive unknown`; 1,011 `SPEC-` rows: completed 780, implemented 143, draft 25, archived 31, superseded 15, in-progress 15, rejected 2 (*scratch*). The `Modified` column reads `2026-10-02 17:28` on every row — checkout time, not authorship. **The list depends on the tree it is run from:** the same command with cwd = the primary checkout printed 678 lines / 44,141 bytes |
| M-7 | files of `spec/plan/acceptance/design/research/progress.md` under `.moai/specs` of this tree | largest 133,596 bytes; 2,936 of 4,129 files exceed 10,000 bytes; none exceeds 4 MiB (*scratch*) | the 10,000-char Markdown element limit (M-12) bites on most files |
| M-8 | `claude plugin validate <stub>` / `--json` on a throwaway stub mod in the scratchpad | exit 0; prints `./register.ts hooks: …` and `./register.ts calls: $.command.register, $.process.run, $.ui.ask`; the same two lines are in `contents[].notes` of `--json`. A hooks.json whose `modules` was a string was **refused** ("expected array"); `modules` is an array of exactly one path | the `calls:` line is a machine-readable list of every `$` noun the module calls |
| M-9 | `claude plugin test <stub>` | with zero test files: **exit 1** (`no *.test.ts or *.test.tsx under …`). With tests: `process.run` is stubbed by `on('process.run', …)` returning `{ value: … }`, the argv is assertable, and `$.ui.ask` is stubbed by `on('tool.call', { tool: 'AskUserQuestion' }, …)` returning `{ result: { questions, answers } }`; 3 of 3 stub tests passed | an empty sweep cannot pass vacuously; both the process boundary and the confirmation dialog are testable |
| M-10 | `moai gtd next t99999 --expect zz` **run from this lane session** | exit 1; stdout 0 bytes; stderr 272 bytes: `Moai next: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next.` | `MOAI_FACTORY_ROLE=lane` here. A successful pick and the not-found error path were **not** observable from a lane (Gap G-2) |
| M-11 | packaging checks at the tree | `git check-ignore -v` over six mod paths (`mods/moai-board/` + `.claude-plugin/plugin.json`, `hooks/hooks.json`, `hooks/register.tsx`, `hooks/board.test.ts`, `types/index.d.ts`, `tsconfig.json`) → exit 1 (none ignored); control `git check-ignore -v mods/moai-board/node_modules/x` → exit 0 (matched `.gitignore:365`). `…/.claude-plugin/types/claude-code/index.d.ts` → exit 1 (**not** ignored). `.goreleaser.yml` archive `files:` are `README.md`, `LICENSE` only; `grep -n mods .goreleaser.yml` → exit 1. `ls` of the repository root shows no `.go` file, and a `go:embed` pattern cannot reach outside its package directory (Go language rule, not re-tested here); the template embed root is `internal/template/templates` (`//go:embed all:templates`). `template-neutrality-check.yaml` path filters name only `internal/template/templates/**` and two tests; `codeql.yml` matrix is `language: [go]`; `ci.yml` path filter names `**/*.go` | `git ls-files -- '*.ts'` lists four tracked `.ts` files (`docs-site/api/i18n-detect.ts` + three testdata): no TypeScript plugin precedent, no root `package.json` |
| M-12 | platform contract read from the 2.1.287 typings (`claude-code.d.ts`) | see §4 | the typings are the authority; the reference note says the API is early access and moves between releases |
| M-13 | `claude plugin test <stub>` repeated later in the same session | the first four runs executed (M-9). After a stub test file that imports a sibling module and builds a 2 MB payload was added, four consecutive runs exited **1** (the last one after the SPEC lint below) with `claude plugin test: hooks modules are turned off in this process: the rollout switch served off, and a plugin's tests run only while it is on`. `claude plugin validate` on the same stub kept exiting 0 | the test runner is gated by a server-side rollout switch that **flipped within one session**. A refusal is not a failure of the mod and not a pass (disposition: acceptance.md §A.2) |

## 2. Boundary — read-only first

[HARD] The mod is **read-only except for one action**. The complete action allow-list is **refresh**, **open**, **pick**:

- **refresh** re-runs the read commands. **open** shows a card or a SPEC detail in the pane.
- **pick** promotes one `queued` card to `picked` with `moai gtd next <id> --expect <prefix>`. It exists only as an operator press followed by an explicit `$.ui.ask` confirmation. This keeps the kanban rule that **promotion is the operator's act** (`.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act): nothing in the mod promotes a card on its own, ranks cards, or picks for the operator.

Everything else is out of scope (§6): editing, dropping, done, hold/unhold, move, unpick, merge, push, pull requests, external issues, and any write to a file.

## 3. Requirements (GEARS)

- **REQ-MBM-001** (Ubiquitous) — The mod shall register one slash command named `moai-board` (the name carries no colon) and one pane with id `moai-board`; running the command shall open the pane and have no other effect.
- **REQ-MBM-002** (Ubiquitous) — The mod shall offer exactly three user actions — refresh, open, pick — and no control for editing, dropping, completing, holding, reordering, unpicking, merging, pushing, creating a pull request, or touching an external issue; the mod shall run no command that is not in its fixed argv table, and shall run the pick command from no timer, no session-start hook, no render hook, and no hook other than the press handler of the pick button.
- **REQ-MBM-003** (Event-driven) — When the operator presses pick on a card whose state is `queued`, the mod shall ask for confirmation through `$.ui.ask`, naming the card id and the first line of its text, and shall run the pick command only when the answer equals the confirm label exactly.
- **REQ-MBM-004** (Event-driven) — When the confirmation is dismissed, answered with any text other than the confirm label, or cannot be asked because no person is present, the mod shall run no process and change nothing.
- **REQ-MBM-005** (Event-driven) — When the pick command exits non-zero or cannot start, the mod shall show the first 400 characters of its stderr (or the start error) in the pane and in a toast, shall not retry, and shall refresh; when it exits 0 the mod shall refresh and shall report the pick as unconfirmed if the card still reads `queued`.
- **REQ-MBM-006** (State-driven) — While the pane is open, the mod shall poll the read commands the visible tab needs, plus the queue for the summary line, at an interval no shorter than 15,000 ms, with at most one poll in flight and a per-command timeout of 20,000 ms; while the pane is closed the mod shall hold no running timer; the SPEC list shall be read on tab open and on refresh only.
- **REQ-MBM-007** (Ubiquitous) — The queue view shall be built from the non-dropped items only: dropped items, `archived`, `runtime` and unknown keys shall be discarded at parse time, the summary line shall show the counts In progress (`picked`), Queued and Held, and a poll whose raw stdout equals the previous one shall be neither re-parsed nor re-rendered.
- **REQ-MBM-008** (Event-driven) — When the queue read reports truncated stdout, is not valid JSON, or has no `items` array, the mod shall read `moai gtd list --limit 0` instead and show the queue with a notice that detail columns are limited; when that read also fails the queue tab shall show the degraded state of REQ-MBM-010.
- **REQ-MBM-009** (Ubiquitous) — The Lanes tab shall list factory cards that are not legacy and not `completed`, grouped by `owner`, with state, stage, SPEC id and an explicit "lease expired" marker, followed by sessions from `moai session list --json` whose heartbeat is within 24 hours, newest first, capped at 20 rows, each shown with its heartbeat age; the tab shall not label any session alive or dead.
- **REQ-MBM-010** (Event-driven) — When a command cannot start (including `moai` missing from PATH), exits non-zero, times out, or returns output the mod cannot parse, the mod shall show a one-line plain-language cause in the affected tab, keep the last good data dimmed with its age, and let no exception leave a hook, so that Claude Code continues unaffected.
- **REQ-MBM-011** (Ubiquitous) — The SPEC tab shall list SPECs from `moai spec status --list` (id and status; every row that is not a `SPEC-` id discarded), default to the statuses draft and in-progress, offer a filter for every other status, and page 15 rows at a time.
- **REQ-MBM-012** (Event-driven) — When the operator opens a SPEC, the mod shall read only `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md` or `progress.md` of a SPEC id that matches the SPEC-id pattern, through a path whose resolved real path lies under the resolved real path of `<session root>/.moai/specs`; it shall show one file at a time as Markdown elements of at most 9,000 characters each, split at blank lines and never inside a fenced code block, at most 12 elements per file followed by a truncation notice naming the file; an id, file name or path failing any check shall not be read.
- **REQ-MBM-013** (Ubiquitous) — The mod shall build every command as an argv list from the fixed table through a single function and never as a shell string, shall validate the card id before use and refuse a `--expect` prefix that begins with `-`, and shall make no network call, no file write, no tool call, no prompt submission, and no model, agent, settings or environment access.
- **REQ-MBM-014** (Ubiquitous) — The mod shall live in `mods/moai-board/` as a standalone plugin directory, shall not be embedded in or deployed by `moai init` or `moai update`, shall not appear in release archives, and the repository `.gitignore` shall ignore the engine-written `mods/*/.claude-plugin/types/`.
- **REQ-MBM-015** (Ubiquitous) — The pane shall draw three tabs — Queue, Lanes, SPEC — with hotkeys `1`, `2`, `3` and a refresh button with hotkey `r`, using only the `Box`, `Text`, `Button` and `Markdown` elements of the surface table, sized to the pane's body columns, and shall hold its view and data state in `$.state`, not in module variables.

## 4. Platform constraints (from the 2.1.287 typings; the contract is early access)

- `$.process.run(argv, init)`: argv list, no shell; cwd defaults to the session's; default timeout 30 s, ten minutes at most; stdout is cut at **4,194,304 bytes** with `isStdoutTruncated`; rejects when the command cannot start or is still running at the timeout; resolves with any exit code.
- `$.fs.read(path)`: a relative path is under the session's working directory, an absolute path is used as given; over 4 MiB rejects. `$.fs.stat(path, { resolve: true })` answers `realPath`; the documented robust guard is an allow-list on `realPath` under a root resolved the same way. `$.session.root()` is the session's project root — where it started, or where `/cd`, a host directory change or a worktree move took it.
- `$.ui.ask(question, options)`: 2-4 option labels; rejects when dismissed and in a `-p` run. `$.ui.open({ id })`: id is 1-64 of letters, digits, `_`, `-`; opened by the person's command it is placed at any width, unasked it waits below 144 columns; docked beside the transcript from 110 columns in fullscreen, otherwise inline.
- A `Markdown` element carries at most 10,000 characters. `Box`, `Text`, `Button`, `Markdown` exist on every surface; `Input` and `Client` do not.
- A hook has a **10,000 ms budget of its own time** per dispatch; a `$` call in flight does not count against it. `$.clock.every` timers run until cancelled or the module reloads. A hook that throws is skipped and the chain continues.
- A command name is letters, digits, `_`, `-` (up to 64): `/moai-board`, never `/moai:board`. `hooks/hooks.json` names exactly one hooks module in a `modules` array.

## 5. Decisions and open decisions

**D-1 — Location (decided, with evidence).** The prototype lives in `mods/moai-board/` at the repository root. Evidence is M-11: no `.gitignore` rule matches the mod's own paths (a control path under `node_modules/` does match); `.goreleaser.yml` archives carry `README.md` and `LICENSE` only; no Go package can embed a root-level directory and the template embed root is `internal/template/templates`; the neutrality and CodeQL workflows are scoped to that tree and to Go; the repository has no root `package.json` and no TypeScript plugin precedent, so the mod brings its own `tsconfig.json`. One collision **is** found: the engine lays its typings under `<mod>/.claude-plugin/types/` at every load (reference note), and nothing ignores that path — REQ-MBM-014 adds the ignore rule. Final deployment scope (user, project, or repository-distributed) depends on card t1434, which is not done; D-1 moves with it (Q1).

**Open decisions** — recorded without a preferred answer in `decision-index.md` and routed there for the operator:

- **Q1** — final deployment scope after the prototype (t1434 not done).
- **Q2** — whether the text list should be the primary queue source instead of a fallback (M-2 vs M-3).
- **Q3** — the polling interval value above the 15,000 ms floor.
- **Q4** — whether the SPEC tab reads the session root only or also the primary checkout (M-6).
- **Q5** — whether pick is offered in a lane session, where the CLI refuses it (M-10).
- **Q6** — which statuses count as "active" for the SPEC tab's default filter (REQ-MBM-011 names draft and in-progress).

**Escalations (leader-first list: confidence < 0.5, push, PR, deletion, external issues):** none is required by this card. Two run-phase items need a person and are listed so they are not discovered late: (1) the first interactive launch and the live pick (AC-MBM-014, AC-MBM-015) need an interactive terminal and a **non-lane** session; (2) the live pick mutates the real queue and needs an operator-admitted throwaway card.

## 6. Out of Scope

### Out of Scope — actions beyond refresh / open / pick

- Editing a card, `drop`, `undrop`, `done`, `hold`, `unhold`, `move`, `unpick`, `claim`, `landed`, `answer`, and every other `moai gtd` or `moai factory` mutation.
- Merge, push, pull-request creation, external issue creation, and any `git` command.
- Writing, creating, deleting or renaming any file from the mod.

### Out of Scope — display ideas deferred from the design report

- Usage / context-window warning band, toasts for lane events, Codex review display, Jev buttons, the health line.
- The "landing-pending" counter of the report's mock: it needs per-card pull-request and landing state (`moai gtd pr`, a git query per card), which this MVP does not run.
- Card relations (`findings`), the dropped-card view, search, sorting by anything other than the emitted order.

### Out of Scope — deployment and distribution

- Installing the mod through `moai init`, `moai update`, a marketplace, or the user's mods folder; deciding its final load scope (card t1434).
- Adding a CI job, a Go test, or a release step for the mod; running `tsc` in CI.
- Changing any `moai` Go code (for example a slimmer `--json` form of `moai gtd list`); that would be a separate card.

## 7. Gaps — unobserved, not passes

- **G-1 Interactive behaviour.** Pane rendering, docking at ≥ 110 columns, hotkeys, scrolling and `Markdown` painting were not observed; no interactive terminal exists here. `claude plugin test` exercises hooks and the returned tree, "never a surface's paint" (typings header).
- **G-2 Live pick.** A successful `moai gtd next <id> --expect <prefix>` and its not-found error were not observed: this session is a lane and the CLI refuses queue mutation (M-10). Only the refusal path is measured.
- **G-3 Parse cost in the engine.** Wall times above are `moai`'s own. The CPU time of `JSON.parse` over ~1.9 MB inside the mod's isolate, against the 10 s hook budget, is unmeasured; AC-MBM-006 measures completion on a synthetic 2 MB payload, not the engine's CPU time.
- **G-4 Growth of the JSON path.** `archived` is 1,524,666 of 1,938,159 bytes and the cap is 4,194,304, i.e. 46% used. The growth rate was not measured; the time until the cap is reached is unknown. REQ-MBM-008 exists because the margin is finite, not because a date was computed.
- **G-5 `moai` on the engine's PATH.** `which moai` resolves in this shell; the PATH of the Claude Code process that will host the mod was not inspected.
- **G-6 Engine-written typings.** `validate` and `test` were observed **not** to write `.claude-plugin/types` into the stub; that a session load writes it into a `--plugin-dir` folder is documented, not observed. REQ-MBM-014's ignore rule is defensive on that documentation.
- **G-7 Type check.** `tsc` is not on PATH (`which tsc` empty); a type check needs the engine-laid typings and a TypeScript ≥ 5.4, so it is advisory here, not an acceptance gate.
- **G-8 `$.state` size.** No size limit is documented in the typings; the reduced queue is ~57 KB (M-2) and is the largest value stored. Unmeasured against the host.
- **G-9 Desktop surface.** Tests mount the pane on `terminal` and `desktop`; `vscode` and `mobile` were not exercised.
- **G-10 Early-access API.** The contract moved between releases before; a later Claude Code build may change names used here. Nothing guards against that besides `validate` and `test` on the build in use.
- **G-11 Rollout switch.** The same switch that gates `claude plugin test` gates loading a hooks module in a session (`reference.md`: a plugin "whose module was not loaded, … the switch being off included"). It was observed off for the test runner (M-13) and its value for a live session was not read. If it is off for the operator, the mod loads nothing and no AC beyond `validate` can run; that outcome is unknown, not a mod defect.
- **G-12 Importing plugin modules from a test.** Whether a `*.test.ts` may import a sibling module of the plugin (the pure parsers) was **not observed**: the runner went off (M-13) before the first such test ran. The plan's M1 settles it first and falls back to exercising the pure functions through the hooks if it cannot.
