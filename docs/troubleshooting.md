# Troubleshooting

When a Claude Code session behaves in a way you do not expect, this page helps you find out whether MoAI ADK, another customization or Claude Code itself is the cause. Take the steps in order: each one narrows the search.

The flags, variables and commands below are Claude Code's own. They come from its official documentation (`debug-your-config`, `cli-reference`, `env-vars`, `hooks`, `authentication`, `auto-mode-classifier-billing`, `claude-projects`, `self-hosted-environments` and `desktop-linux` at https://code.claude.com/docs/en/), read on 2026-11-22. If a name no longer works, check the current page.

## 1. Start in safe mode

Run:

```sh
claude --safe-mode
```

Safe mode starts a session with your customizations turned off: CLAUDE.md files, skills, plugins, hooks, MCP servers, custom commands and agents, output styles, the status line and auto memory do not load. That includes everything MoAI ADK installed — its skills, agents, rules and hooks. Login, model choice, built-in tools and permissions keep working. Managed settings from your organization still apply.

- If the problem goes away in safe mode, the cause is one of the customizations it turned off. Go on to step 2 to find which one.
- If the problem stays, the cause is not MoAI ADK or another customization: look at Claude Code itself, your login, your network or your organization's managed settings.

## 2. Reproduce with a clean configuration folder

Claude Code keeps your settings, session history, plugins and login in its configuration folder, `~/.claude` unless `CLAUDE_CONFIG_DIR` points elsewhere. To try a clean one, point `CLAUDE_CONFIG_DIR` at a new folder and turn off the project's configuration:

```sh
CLAUDE_CONFIG_DIR=~/claude-clean/config claude --setting-sources user
```

- `--setting-sources user` turns off project configuration such as `.claude/settings.json`, `CLAUDE.md` and `.mcp.json`.
- On the first launch you go through the first-run setup screens again. That is the sign the clean folder is in use. You log in again in that folder: the login and, on macOS, the Keychain entry belong to the configuration folder.
- If the problem is gone, copy `~/.claude.json` and the files in `~/.claude` into the clean folder one at a time until it comes back. The last file you copied holds the cause.

Change `CLAUDE_CONFIG_DIR`, never `HOME`. Claude Code documents `CLAUDE_CONFIG_DIR` as the way to keep a separate configuration (it is also how you keep two accounts side by side). Changing `HOME` moves much more than Claude Code's files: your git, SSH and shell configuration go with it, so the reproduction no longer matches your real setup.

## 3. Write a debug file

Debug output goes to a file, not to the terminal:

```sh
claude --debug-file ~/claude-debug.txt
```

- `--debug-file <path>` writes the debug log to that path and turns debug mode on. Without it, `claude --debug` writes to `~/.claude/debug/<session-id>.txt`. Inside a session, `/debug` turns debug logging on and asks Claude to read the log for you.
- For more detail, start the session with `CLAUDE_CODE_DEBUG_LOG_LEVEL=verbose`. The levels are `verbose`, `debug` (the default), `info`, `warn` and `error`.
- Hooks show up in the log. A hook that ran leaves a line such as `Hooks: PostToolUse:Write [<command>] finished with status 0 (31ms)` (Claude Code v2.1.296 or later). A run can also end as `timed out after <N>ms`, `cancelled`, `moved to the background` or `failed to start`. A MoAI ADK hook that blocked a call, or never ran, is visible here.

## 4. Clean a debug file before sharing it

A debug file records hook output, command lines and file paths from your session. Read it before sharing it with anyone, and remove API keys, tokens, passwords and private paths (your user name, customer or project names) first. Share only the part around the problem.

## 5. The auto mode classifier notice behind a gateway

When you use auto mode through an LLM gateway or proxy, Claude Code may show a notice titled "Auto mode classifier request charges", saying the session "isn't eligible". It means the server-side checks do not reach the session, so Claude Code makes its own classifier requests and they count toward your token usage. Auto mode keeps working.

- On a Pro, Max, Ultra or Team plan you never see this notice. `/status` shows an **Auto mode server** row: `Enabled`, or `Disabled` once the session has fallen back.
- The lasting fix is a gateway that forwards the `anthropic-beta` header and the `safeguards` body field unchanged.
- Setting `CLAUDE_CODE_AUTO_MODE_SERVER=0` before you start the session (in your shell or in the `env` settings key) keeps the notice from appearing. Claude Code calls this a temporary setting that may be removed in a later release.

MoAI ADK does not change this setting for you.

## 6. See what the session loaded

These commands show what Claude Code actually loaded, so you can tell whether a MoAI ADK file was read:

| Command | Shows |
| --- | --- |
| `/context` | everything in the context window, with the source of each subagent, memory file and skill |
| `/memory` | the memory and CLAUDE.md files in the user and project scopes |
| `/skills` | the skills from project, user and plugin sources |
| `/hooks` | the active hooks; a hook you defined that is missing here was not loaded |
| `/permissions` | the allow and deny rules in effect |
| `/status` | the active settings sources, and whether managed settings apply |
| `/doctor` | a setup checkup: installation, invalid settings files, duplicate subagent names |
| `claude doctor` | read-only installation and settings diagnostics, without starting a session |

Settings apply in this order: managed settings first, then local over project over user. Flags and environment variables can override them again.

## 7. Where MoAI ADK does not reach

Some places do not run MoAI ADK's hooks or read its settings. MoAI ADK's checks do not protect work done there.

- **Cloud threads.** Claude Code's cloud threads do not pick up anything from the Claude Code setup on your own machine. They read the repository's `CLAUDE.md` and rules, but they do not run hooks from `.claude/settings.json` and do not apply its permission rules or `env`. MoAI ADK's hooks and settings have no effect there.
- **Self-hosted runners.** MoAI ADK does not support self-hosted runners. A runner can point `core.hooksPath` at its own hooks directory, which skips the repository's git hooks, and its git proxy option replaces the global git configuration with no backup.
- **Linux desktop.** MoAI ADK does not support the Claude desktop app on Linux, which is a beta. Use the Claude Code CLI on Linux instead.
