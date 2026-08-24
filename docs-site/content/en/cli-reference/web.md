---
title: moai web Console
weight: 50
draft: false
description: "The moai web command that starts the local operations console — flags, routes, port-reclaim behavior."
---

`moai web` starts **MoAI Web Console**, the local operations screen. It lets you view the project's SPEC catalog, the Kanban chain, and session, goal and verification state in a browser, and edit profile preferences and project settings from the same screen.

The screen layout and what each area reads are covered in [MoAI Web Console](/en/advanced/moai-web-console/). This page covers the command itself — flags, routes and port handling.

## Overview

```bash
moai web [OPTIONS]
```

The console **binds to loopback (127.0.0.1) only**. There is no external database, no authentication and no network exposure.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port <N>` | `3041` | The TCP port to bind on 127.0.0.1 |
| `--no-open` | `false` | Do not open the browser automatically |
| `--no-reuse` | `false` | Do not reclaim the port from a stale moai instance; fail on a port conflict instead |

## Examples

```bash
moai web                 # Bind 127.0.0.1:3041 and open a browser
moai web --port 9000     # Bind a different port
moai web --no-open       # Start without opening a browser
moai web --no-reuse      # Fail instead of reclaiming a port that is in use
```

## Port-reclaim behavior

If a stale moai instance already holds the target port, the default behavior terminates that instance and rebinds. A process that is **not** moai is never terminated; in that case the command reports an error and suggests a different port via `--port`. Passing `--no-reuse` makes it fail without reclaiming even a stale moai instance.

If opening the browser fails, the server stays up regardless. Open the address printed in the terminal by hand.

## Routes

The console serves the following paths. The five read-only screens refuse any method other than GET with a 405.

| Path | Method | What it does |
|------|--------|--------------|
| `/` | GET | Overview — stat tiles, Kanban chain, in-progress SPECs, attention list, sessions |
| `/kanban` | GET | Chain session board plus the SPEC pipeline |
| `/specs` | GET | SPEC catalog. `?q=` searches, `?status=` filters, `?id=` opens the detail |
| `/monitor` | GET | Sessions, goals, verification, epics |
| `/settings` | GET | The nine settings tabs. `?tab=` selects the tab, `?profile=` the profile being edited |
| `/todo` | GET | The backlog queue, read-only — every card in all three states (`queued` · `picked` · `dropped`) |
| `/events` | GET | SSE stream — carries refresh signals only |
| `/save` | POST | Save settings |
| `/profile/create` · `/profile/rename` · `/profile/delete` | POST | Profile lifecycle |
| `/glm-key/reveal` | POST | Reveal the stored GLM API key |
| `/__shutdown__` | POST | Shut the server down |

`/events` is a stream that holds the connection open. Opening it directly with `curl` never completes the response, so being cut off by a timeout is the correct behavior.

## Shutting down

Press `Ctrl+C` in the terminal, or use the shutdown button at the foot of the rail. Either way, in-flight requests finish before the server exits.

## Scope of the profile record

Switching profiles in the console records that choice for the current project in `~/.moai/claude-profiles/launch.yaml`. That value is used when you run `moai cc` without `-p` in the same project.

Everything the console reads and writes is scoped to the current project, so the profile shown on screen and the profile actually recorded are always the same. One exception: opening the console inside a session started with `moai cc -p X` shows `X` regardless of the record, because `CLAUDE_CONFIG_DIR` is already fixed.

Selection order and constraints are covered in [profile management](/en/cli-reference/profile/).

---

Related: [MoAI Web Console](/en/advanced/moai-web-console/) · [profile management](/en/cli-reference/profile/) · [CLI overview](/en/getting-started/cli/)
