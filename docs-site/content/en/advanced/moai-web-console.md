---
title: MoAI Web Console
weight: 85
draft: false
---
# MoAI Web Console

**MoAI Web Console** is the browser-based way to edit the same profile·project settings that the terminal wizard edits. One `moai web` command opens the console, and changing values with the mouse passes through the same validation as the terminal wizard before saving. It's faster and less error-prone than hand-editing YAML files directly.

![MoAI Web Console — project name and profile in header, profile bar with add/rename/delete, nine setting tabs](/img/moai-web-console.png)

## What It's For

The web console edits two kinds of settings.

### Profile preference editing

Edit per-profile values like model·inference strength·language·display settings. When using multiple profiles in one project — for example `medium` normally, `max` for difficult work — switch profiles in the profile bar and adjust each profile's detailed values in the tabs. The add / rename / delete controls sit next to the profile selector itself, so the full profile lifecycle is on one row — no separate profile-management card.

### Project section editing

Edit user / language / statusline sections under `.moai/config/sections/` in the web UI. No need to open YAML in an editor from the terminal.

When you save, changes pass through the same validation as the terminal wizard (`moai profile`), so results are identical whichever way you edit. Wizard users who prefer it can stay with the wizard; users who prefer a visual overview can use the console — both edit the same settings files.

## Console Structure — 9 Tabs

The console is organized into nine tabs, in this order:

1. **Identity** — display name and project-level identity fields
2. **Language** — conversation / commit-message / code-comment / documentation language
3. **LLM** — permission mode, model, reasoning effort (the session-level "how sessions start" trio)
4. **3rd Party LLM** — GLM model selection per tier, reasoning effort per tier, and the GLM API key
5. **Workflow** — execution mode, default mode, agentic-loop and loop-prevention knobs
6. **Git & Worktree** — `git_strategy.mode`, the three per-profile `merge_method` values, and the worktree / branch-guard toggles
7. **Audit** — audit model and the per-backend gates (claude / codex / glm)
8. **Agents** — per-agent profile/model assignment
9. **Report** — report format and output preferences

The tab list is the single source of panel order — what you see across the top matches what renders below, and the workflow / git-worktree / audit split keeps any one tab from overcrowding. Two earlier tabs (Workflow, Git & Worktree, Audit) replace what used to be a single overloaded Workflow tab plus a `git_strategy` section that had fields but no UI; restoring that surface is part of the redesign.

### Widget honesty

Every field renders a widget that matches its real value domain:

- **Bool fields** render as a two-option radio group (used / not-used), not a checkbox. A checkbox communicates "off and on" implicitly; the radio pair makes the active choice visible.
- **Closed-set fields** (e.g. `execution_mode`, `audit.model`, the `harness.*` selectors) render as a select or radio group, and the apply layer rejects any value outside the declared set on save. A free-text box over a closed set invites invalid input — the console no longer presents that invitation.
- **Open-value fields** (repository paths, output directories, the GLM API key) remain free text, because their domain genuinely is open.

### GLM honesty badge

On the 3rd Party LLM tab, the reasoning-effort runtime channel is a single session-level env var — there is no per-tier delivery path. The per-tier effort values you set are therefore **store-only**: they persist in your config, but the runtime reads only the session-level effort collapsed to z.ai's three states. The tab carries a badge naming the **applied source** (the session-level effort) so the store-only nature of the per-tier values is stated, not implied.

## 4-Locale UI

The console's interface language is chosen from the selector at the upper right of the header: **English · 한국어 · 日本語 · 中文**. The same four languages MoAI-ADK supports. When you choose Korean, menus·labels·help text all switch to Korean. It's the same language set as the docs-site's [4-locale documentation](/en/resources/i18n-docs/), so you can review settings in your native language.

## Getting Started

From the project directory, one command opens the console.

```bash
moai web
```

By default it binds to `127.0.0.1:3041` and the browser opens automatically. If the port is already in use by another moai instance, that instance is terminated before rebinding. If an external process (not moai) is using the port, it doesn't terminate and errors instead — in that case, use `--port` for a different port.

```bash
moai web --port 9000     # Bind to a different port
moai web --no-open       # Start without opening browser
```

Detailed flags and behavior are in [CLI reference — moai web](/en/cli-reference/web/).

## Security Model

The web console is built for local development convenience, with security designed accordingly.

### Loopback-only

The console binds **only to `127.0.0.1`**. It's NOT exposed to external network interfaces. Other user accounts on the same machine or remote hosts cannot reach it.

### No database (no-DB)

No separate database server is started. Values the console reads and writes are only the current project's config files (`.moai/config/sections/`) and profile files. When you close the console, only the config file changes you saved remain.

### No authentication (no-auth)

Because only the local user can reach it (loopback-only), there's no authentication layer like login or tokens. Use it immediately without complex credential management.

{{< callout type="info" >}}
**Loopback-only is the premise of no-auth.** Exposing the console via reverse proxy or `0.0.0.0` binding to external networks is NOT officially supported. If you need remote editing, forward the local port safely via SSH tunnel.
{{< /callout >}}

## Profile Recording Scope

When you switch profiles in the console, that selection is recorded to `~/.moai/claude-profiles/launch.yaml` as the current project's record. When you run `moai cc` without `-p` in the same project, this value is used. Because both the value the console reads and writes are based on the current project, the profile shown on screen and the profile actually recorded are always the same.

Selection order and constraints are covered in detail in [Profile management](/en/cli-reference/profile/).

## When to Use Terminal Wizard vs. Console

Both handle the same settings files, but which is more convenient depends on your workflow.

- **Terminal wizard** (`moai profile`, `moai update -c`) — Initial setup changing one value at a time, scripted settings run in CI, when interactive guidance is needed.
- **Web console** (`moai web`) — When comparing and editing multiple sections on one screen, when switching profiles frequently to compare values, when you want fast edits without worrying about YAML indentation.

Save results are identical. Choose whichever feels right.

## Related Documents

- [CLI reference — moai web](/en/cli-reference/web/) — Flag and behavior details
- [Profile management](/en/cli-reference/profile/) — Profile auto-selection and recording scope
- [Harness profiles and evaluation](/en/advanced/harness-profiles/) — Profile matrix and model·inference-strength assignment
