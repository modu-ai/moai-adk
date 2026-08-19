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
moai cc [-p profile] [-w [name]] [-- claude-args...]
```

Removes GLM-specific environment variables from `.claude/settings.local.json`, resets team mode if it was enabled, and then runs Claude Code.

| Flag | Description |
|--------|------|
| `-p, --profile <name>` | Use a named Claude profile (`~/.moai/claude-profiles/<name>/`) |
| `--permission-mode <mode>` | Specify the permission mode |
| `-b, --bypass` | Shorthand for `--permission-mode bypassPermissions` |
| `-c, --continue` | Continue the previous session |
| `-m, --model <model>` | Override the model selection |
| `-w, --worktree [name]` | Launch inside an isolated git worktree (`.claude/worktrees/<name>/`) — name omitted means auto-generated |
| `--chrome` / `--no-chrome` | Toggle the Chrome MCP |
| `-k, --kanban [SPEC-ID]` | Enter as the kanban lead — seeds the `plan → run → sync` chain in this session. With a SPEC-ID attached, that SPEC is the target |
| `-k --name <role>` | Join an open kanban run as a companion session. Roles are `plan` · `run` · `sync`. If a live session already holds the role name, the next number is attached (`plan-1`, `plan-2`, …) |
| `-f, --factory [N]` | Enter as the **factory lead** — a factory run opening N lanes (`lane-1`…`lane-N`). Omit N and the run starts with a single lane (`lane-1`), grown one at a time with the incremental form below. The lead deals the cards the operator picks to free lanes over cross-session messages |
| `-f lane-<n>` | Bring up one more lane (`lane-<n>`) and attach it to the lead socket of a running factory. A label already held by a live session bumps to the next free number. `moai glm -f lane-<n>` behaves the same on the GLM backend |
| `-k <N>` / `-k <N> --name lane-<i>` | The v1.2.0 combined form, still valid — `-k <N>` is the lead of an N-lane run, `-k <N> --name lane-<i>` is lane `<i>` within it. `-k --name lane-<i>` without N defaults to 8 lanes |

{{< callout type="info" >}}
`-k` is the kanban chain token, and `-f` is the dedicated entry token for **Factory Mode**. One `-k` is still read three ways — no argument or a SPEC-ID is the kanban lead, `--name <role>` is a kanban companion, and a number is a lane run. A launch takes one entry token only, so `-k` together with `-f` is an error. The mixed-backend launcher `moai cg` refuses both modes (the factory-side refusal sentinel is `FACTORY_MODE_UNSUPPORTED_BACKEND`). For the full contract see [Kanban Mode](/en/advanced/kanban-mode) and [manager-lead Lead Coordinator](/en/advanced/manager-lead).
{{< /callout >}}

A card travels differently here than in kanban. In kanban one card moves across the `plan → run → sync` columns, while in a factory one card goes whole to one lane and passes the three phases serially inside it. Each phase is spawned by that session as `Agent()` subagents, and write-capable spawns are isolated with `isolation: "worktree"`. A lane runs at most 10 concurrent subagents, and the launcher seeds that value into lanes and companion sessions through `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, so N lanes sharing one machine's capacity is guaranteed by configuration rather than by operator restraint. Never bring the lanes up all at once — start the first, confirm it is actually producing output, then activate the rest.

The backend mix is decided by token availability first. One starting point is the lead on GLM, plan on Claude (Opus), run on GLM, and sync on Claude (Opus), placing Opus only on the phases where judgment is heavy. A different combination, or unifying on a single backend, is equally fine.

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

## Isolated worktree (`-w` flag)

All three launchers accept `-w [name]` to start the session inside an isolated git worktree, collapsing the two-step `cd` then launch into a single command.

```bash
moai cc -w feat-login    # Start in .claude/worktrees/feat-login/
moai cc -w               # Auto-generated name
moai glm -w feat-login   # Same for the GLM backend
moai cg -w feat-login    # Same for the hybrid
```

Behavior:

- The worktree path is `.claude/worktrees/<name>/`. `<name>` is a **worktree name** — not a branch name and not a SPEC ID.
- If a worktree of that name already exists it is **reused rather than recreated**, so this doubles as the re-entry path into a tree an earlier session was working in.
- Omitting the name lets Claude Code generate one.
- The `-w=name`, `--worktree name`, and `--worktree=name` spellings are all accepted and mean the same thing.
- Arguments after `--` pass through to Claude Code untouched and are unaffected by this rewrite.

{{< callout type="info" >}}
Naming the worktree after the SPEC ID (`moai cc -w SPEC-XXX-001`) lets a session handoff bring the next session back into the same working tree with one line.
{{< /callout >}}

## Related documents

- [CG Mode (Claude + GLM)](/en/multi-llm/cg-mode)
- [Profile Management](/en/cli-reference/profile)
- [Security Notes](/en/advanced/security-notes) — GLM credential path security model
- [CLI Overview](/en/getting-started/cli)
