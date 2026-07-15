---
title: moai cc / cg / glm Launchers
weight: 15
draft: false
---

`moai cc`, `moai cg`, and `moai glm` are three launchers that run Claude Code with different backend configurations. All three adjust settings and then `exec` to replace the current process with Claude Code. Since which model does which work is what drives cost, the launcher choice is the first tokenomics decision.

## The three launchers compared

| Launcher | Backend | Purpose |
|------|--------|------|
| `moai cc` | Claude only | Standard execution — every agent uses Claude models |
| `moai glm` | GLM only | Every agent uses GLM models via the Z.AI proxy |
| `moai cg` | Claude + GLM hybrid | Leader is Claude, teammates are GLM (60-70% cost reduction) |

## moai cc — Claude backend

```bash
moai cc [-p profile] [-- claude-args...]
```

Removes GLM-specific environment variables from `.claude/settings.local.json`, resets team mode if it was enabled, and then runs Claude Code.

| Flag | Description |
|--------|------|
| `-p, --profile <name>` | Use a named Claude profile (`~/.moai/claude-profiles/<name>/`) |
| `--permission-mode <mode>` | Specify the permission mode |
| `-b, --bypass` | Shorthand for `--permission-mode bypassPermissions` |
| `-c, --continue` | Continue the previous session |
| `-m, --model <model>` | Override the model selection |
| `--chrome` / `--no-chrome` | Toggle the Chrome MCP |

The permission mode is one of `default`, `acceptEdits` (project default), `plan`, `auto`, `bypassPermissions`, `dontAsk`. The `auto` mode runs a background classifier that inspects actions and requires a Team plan + Sonnet/Opus 4.6 or later.

## moai glm — GLM backend

```bash
moai glm setup <api-key>   # Save API key (first time only)
moai glm                   # Run with the GLM backend
moai glm -p work           # Run with the 'work' profile
moai glm status            # Check credential status
```

Reads GLM credentials from `~/.moai/.env.glm`, injects environment variables such as `ANTHROPIC_AUTH_TOKEN` and `ANTHROPIC_BASE_URL`, and then runs Claude Code.

| Subcommand | Description |
|-------------|------|
| `moai glm setup [api-key]` | Save the GLM API key |
| `moai glm status` | Show the current GLM credential status |

{{< callout type="warning" >}}
GLM does not support the `auto` permission mode (it is a third-party provider). If you need `auto`, use `moai cc` or `moai cg`. Also, Z.AI has a low concurrent-request limit (1-3 in-flight on the paid tier), so for parallel multi-agent execution the `moai cg` hybrid mode is more stable.
{{< /callout >}}

## moai cg — Claude + GLM hybrid

```bash
moai cg [-p profile]
```

CG stands for "Claude + GLM", a cost-optimized team configuration.

- **Leader** (current tmux pane): uses Claude models (opus/sonnet)
- **Teammates** (new tmux panes): use GLM models via the Z.AI proxy

On launch it validates the tmux session, removes the GLM environment in the leader pane (Claude), injects the GLM environment into the tmux session (teammates), and sets `teammateMode=tmux` and `team_mode: cg`.

**Prerequisites**:

1. Set the GLM API key with `moai glm setup <api-key>`
2. Run inside a tmux session for per-pane environment isolation

## Profiles (`-p` flag)

All three launchers accept `-p <name>` to select a named profile, which sets `CLAUDE_CONFIG_DIR` to `~/.moai/claude-profiles/<name>/`. Use this to keep multiple accounts / setting sets separate.

## Related documents

- [CG Mode (Claude + GLM)](/en/multi-llm/cg-mode)
- [Profile Management](/en/cli-reference/profile)
- [Security Notes](/en/advanced/security-notes) — GLM credential path security model
- [CLI Overview](/en/getting-started/cli)
