# moai-status

A Claude Code **mod** (plugin of function hooks) for the MoAI session: a
usage/context warning strip above the prompt with a spinner suffix, a toast for
every inbound cross-session delivery, and a health/status warning line fed by a
background timer. Strictly additive observer — absent or disabled, nothing
changes; it draws, it never acts.

SPEC: `.moai/specs/SPEC-MOAI-STATUS-MOD-001/` (card t1437).

## The three features

1. **Usage/context warning strip** — the engine pushes its usage figures as
   `session.measure` after each turn (never polled). The mod classifies them
   against thresholds mirroring the in-repo gates' *defaults* (context soft
   50 for windows ≥ 500k tokens else 90, critical at min(95, autocompact+10)
   clamped up to soft; rate holds 90 for `five_hour`, 95 for `seven_day`) and
   holds the classification in `$.state`. The AbovePrompt hook draws a one-line
   strip ahead of the upstream tree when something warns; the Spinner hook
   appends a short marker (` · ctx 91%`) to the incoming suffix only.
2. **Lane notification toast** — every `session.receive` delivery raises one
   toast line (`peer: hello lane`) and the delivery always passes: the handler
   returns `next(e)` on every path and never produces `{ consumed }`.
3. **Health/status warning line** — a `$.clock.every` timer (60 s, floor 15 s)
   runs three diagnostics one at a time under a single-flight gate with a 20 s
   timeout: `moai doctor --check "Binary Freshness"`, `--check "MCP Server
   Version"`, and `moai memory doctor --json`. Warnings pin one line under the
   prompt (behind binary with both SHAs, stale MCP server with pid and
   commits, over-cap memory store with count/cap); a clean cycle clears the
   line; an unknown source shows `name ?` and never reads as healthy.

## Boundary (spec.md §2)

No hook alters turn flow (every handler passes its event through); no network,
file write, tool call, prompt submission, or env/settings access; the only
child-process calls are the fixed argv table's three diagnostics through the
single `$.process.run` site in `hooks/register.ts`. `moai doctor --check "MCP
Server Version"` deletes the CLI's own dead PID-stamp files as a side effect of
the called command, not of the mod.

## Launch

```sh
claude --plugin-dir <absolute path to mods/moai-status>
```

Interactive behavior (the strip's paint, the toast on a real delivery) is the
manual check of acceptance.md AC-MSM-013, performed by an operator with a
second session able to send a message.

## Test and validate

```sh
# pure (bun — developer-local evidence, never engine or CI evidence)
bun test mods/moai-status/tests/pure/

# engine (the authority for hooks, state, timers and render dispatch)
mkdir -p /tmp/msm-claude-cfg-empty
CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status

# manifest, hooks, $-noun calls and the state contract
CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin validate mods/moai-status
```

Layout: `hooks/register.ts` is the only file that spells `$.…` (the engine
refuses `$` passed into an imported function); `hooks/data.ts` and
`hooks/health.ts` stay `$`-free and pure (bun-tested under `tests/pure/`,
named `*.spec.ts` so the engine runner's `*.test.ts` glob skips them); the
state contract lives in `types/index.d.ts`, named by `plugin.json`'s `types`
pointer. The engine lays its own typings under `.claude-plugin/types/` at load
(gitignored).
