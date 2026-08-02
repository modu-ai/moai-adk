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

## Console Screens

The Console's interface language is chosen from the selector at the right of the header, among English · 한국어 · 日本語 · 中文. The screens below were captured with it set to English.

The header places the project name, the current profile, and a summary of the main settings (`lang · model · effort · dev`) side by side. Below it come the profile selector and the Profiles card, followed by six tabs: Identity · Language · LLM · 3rd Party LLM · Agents · Report. Edited values are written with the Save settings button at the bottom.

![MoAI Web Console initial screen: the project name and profile in the header, the Profiles card, the six tabs, the Display name field on the Identity tab, and the Save settings button](/images/profile/web-console-overview.png)

From the Profiles card you can switch profiles, Delete one, and create a new one by entering a New profile name and pressing Create profile. Choosing a different profile also changes the profile shown in the header. Below is the Language tab opened after switching to the `moai-cowork` profile.

![The Language tab of the Console after switching to the moai-cowork profile, with four fields: Conversation language, Commit message language, Code comment language, and Documentation language](/images/profile/web-console-switch.png)

The LLM tab is where you change Permission mode, Model, and Effort level. These are the same values the "Model Settings" step of the terminal wizard covers.

![The LLM tab of the moai-adk profile, with three fields: Permission mode, Model, and Effort level](/images/profile/web-console-llm-tab.png)

## Scope of the Profile Record

When you switch profiles in the Console, that choice is left in `~/.moai/claude-profiles/launch.yaml` as the current project's record. This value is used when you run `moai cc` without `-p` in the same project.

{{< callout type="note" >}}
Per-project records ship in the next release. The currently released version handles only a single global record, with no distinction between projects.
{{< /callout >}}

Both the values the Console reads and the ones it writes are based on the current project, so the profile shown on screen and the profile actually recorded are always the same. However, when you open the Console inside a session started with `moai cc -p X`, `CLAUDE_CONFIG_DIR` is already fixed, so it shows `X` as-is regardless of the record.

The resolution order and its limitations are covered in detail in [Profile Management](/en/cli-reference/profile#automatic-profile-selection).

---

Related: [Profile Management](/cli-reference/profile) · [CLI Overview](/getting-started/cli)
