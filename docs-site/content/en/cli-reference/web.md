---
title: moai web Console
weight: 50
draft: false
---

`moai web` launches the **MoAI Web Console**, a browser-based settings editor. It reuses the same validation and persistence logic as the terminal profile wizard (`moai profile`), letting you edit profile preferences and the project's user / language / statusline sections from a web UI.

## Overview

```bash
moai web [OPTIONS]
```

The Console binds to **loopback only (127.0.0.1)**. There is no external database, no auth, and no network exposure. By default, if the target port is held by a stale moai instance, the Console terminates that instance and rebinds. A non-moai (foreign) process is never terminated — the Console reports an error and suggests `--port`.

## Flags

| Flag | Description |
|------|-------------|
| `--port <N>` | TCP port to bind on 127.0.0.1 (default: `3041`) |
| `--no-open` | Do not auto-open the browser |
| `--no-reuse` | Do not reclaim the port from a stale moai instance; fail on any port conflict |

## Examples

```bash
moai web                 # bind 127.0.0.1:3041 and open the browser
moai web --port 9000     # bind a different port
moai web --no-open       # start without launching a browser
moai web --no-reuse      # fail instead of reclaiming a busy port
```

## What it edits

The Web Console edits:

- **Profile preferences** — per-profile settings such as model, language, and display options
- **Project settings** — the user / language / statusline sections under `.moai/config/sections/`

Saves go through the same validation as the terminal wizard, so results are consistent whichever path you use.

---

Related: [Profile Management](/cli-reference/profile) · [CLI Overview](/getting-started/cli)
