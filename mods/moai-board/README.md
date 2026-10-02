# moai-board

A read-only side panel for Claude Code that shows the MoAI kanban queue, the factory lanes and
SPEC documents in one pane. A **mod**: a plugin of function hooks that runs inside the session
(Claude Code 2.1.287; the plugin API is early access and moves between releases).

Prototype for card t1436, `SPEC-MOAI-BOARD-MOD-001`. It lives in this folder only: it is not
embedded in the `moai` binary, not deployed by `moai init` or `moai update`, and not part of the
release archives. Where it finally loads from (user, project or repository) waits on card t1434.

## What it does

`/moai-board` opens the pane `moai-board` with three tabs (hotkeys `1`, `2`, `3`) and a refresh
button (`r`):

| Tab | Shows | Reads |
|---|---|---|
| Queue | Picked / Queued / Held counts, the cards grouped by state, a card's full text | `moai gtd list --json` (text list `--limit 0` as the fallback) |
| Lanes | factory cards by owner with state, stage, SPEC id and a "lease expired" marker; sessions with a heartbeat in the last 24 h (newest first, 20 at most, no alive/dead verdict) | `moai factory status --json`, `moai session list --json` |
| SPEC | SPECs filtered by status (default draft and in-progress), 15 per page; one file of a SPEC at a time as Markdown | `moai spec status --list`, then the file through the engine's file calls |

The pane polls every 15 seconds at the fastest, one poll at a time, and only while it is open.
A failed read shows a one-line cause and keeps the last good data, dimmed. `moai` missing from
`PATH` is one of those causes; the session carries on.

## The boundary

Three actions only: **refresh**, **open** and **pick**. Everything else (editing, dropping,
completing, holding, moving, merging, pushing, pull requests, writing a file) is out of scope.

**Pick** promotes one `queued` card to `picked` with `moai gtd next <id> --expect <prefix>`. It
runs only after you press the pick button on that card and answer the confirmation dialog with
`Pick`. Any other answer, `Cancel`, a dismissed dialog, or no one to ask (`claude -p`) runs
nothing. A lane session cannot mutate the queue (`moai` refuses); the refusal text is shown as is.

Every command is an argv list from a fixed table in `hooks/data.ts` and goes through one call site
in `hooks/register.tsx`; there is no shell string, no network call, no file write.

## Run it

```sh
claude --plugin-dir /absolute/path/to/mods/moai-board
# then type /moai-board
```

The engine lays its typings into `.claude-plugin/types/` of a mod it loads this way; the repository
`.gitignore` ignores that folder. The mod's own contract, `types/index.d.ts`, is tracked.

## Check it

From the repository root:

```sh
claude plugin validate mods/moai-board
```

Engine tests (hook dispatch, timers, the pick flow, the pane mounted on terminal and desktop). The
runner is refused while the account's rollout switch is off; an empty config directory has run it
regardless:

```sh
mkdir -p /tmp/mbm-claude-cfg
CLAUDE_CONFIG_DIR=/tmp/mbm-claude-cfg claude plugin test mods/moai-board
```

Pure tests (parsers, reducers, the SPEC guard, the chunker). They need `bun`, which nothing else in
this repository uses, so they are developer-local evidence only. Judge them from the junit file:

```sh
bun test mods/moai-board/tests/pure/ --reporter=junit --reporter-outfile=/tmp/mbm-junit.xml
```

Test file names decide the runner: `*.test.ts` and `*.test.tsx` under `tests/` are engine tests,
`*.spec.ts` under `tests/pure/` are pure tests (and are excluded from `tsconfig.json`).

## Layout

```
.claude-plugin/plugin.json   manifest; names the state contract under "types"
hooks/hooks.json             one hooks module
hooks/register.tsx           the only file that spells the engine calls: hooks, timer, press handlers
hooks/data.ts                argv table, run classification, queue and lane parsers, pick helpers
hooks/specs.ts               SPEC list, id and file allow-lists, real-path guard, Markdown chunker
hooks/view.tsx               the three tabs, drawn with Box, Text, Button and Markdown only
types/index.d.ts             the plugin's state contract
tests/                       engine tests; tests/pure/ holds the bun tests
```

The engine refuses the engine handle passed into a function imported from another file, so the
helper modules receive plain functions (`run`, `stat`, `read`) or the resolved element table and
never the handle itself.
