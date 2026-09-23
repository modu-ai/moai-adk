---
title: moai cc / glm Launchers
weight: 15
draft: false
---

`moai cc` and `moai glm` launch Claude Code with an explicitly selected backend. Legacy CG configurations require migration before launch.

## Launcher comparison

| Launcher | Backend | Purpose |
|------|--------|------|
| `moai cc` | Claude only | Standard execution — every agent uses Claude models |
| `moai glm` | GLM only | Every agent uses GLM models via the Z.AI proxy |

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
| `--chrome` / `--no-chrome` | Passed through to Claude Code unchanged; the launcher adds neither, so `/chrome` can attach unless you pass `--no-chrome` |
| `-k, --kanban [SPEC-ID]` | Enter as the kanban lead — seeds the `plan → run → sync` chain in this session. With a SPEC-ID attached, that SPEC is the target |
| `-k --name <role>` | Join an open kanban run as a companion session. Roles are `plan` · `run` · `sync`. If a live session already holds the role name, the next number is attached (`plan-1`, `plan-2`, …) |
| `-f, --factory` | Enter as the **factory lead** — opens a factory run with one worker (`worker-1`). The lead deals the cards the operator picks to free workers over cross-session messages |
| `-f worker` | Auto-join one more worker at the next free number and attach it to the lead socket of a running factory |
| `-f worker-<n>` | Bring up exactly that worker (`worker-<n>`) as one more. A number already held by a live canonical worker bumps to the next free number; a number held by a live legacy worker (`agent-<n>`/`lane-<n>`) is refused by name. `moai glm -f worker` / `-f worker-<n>` behave the same on the GLM backend |
| `-k <N>` / `-k <N> --name worker-<i>` | The v1.2.0 combined form, still valid — `-k <N>` is the lead of an N-worker run, `-k <N> --name worker-<i>` is worker `<i>` within it. `-k --name worker-<i>` without N defaults to 8 workers |
| `-f agent` · `-f lane-<n>` · `--name lane-<n>` | **Deprecated aliases.** Each behaves exactly like `-f worker` / `-f worker-<n>` / `--name worker-<n>`, but every launch prints a hint to switch to the canonical spelling |

{{< callout type="info" >}} `-k` is the kanban chain token, and `-f` is the dedicated entry token for **Factory Mode**. One `-k` is still read three ways — no argument or a SPEC-ID is the kanban lead, `--name <role>` is a kanban companion, and a number is a worker run. A launch takes one entry token only, so `-k` together with `-f` is an error. For the full contract see [Kanban Mode](/en/advanced/kanban-mode) and [manager-lead Lead Coordinator](/en/advanced/manager-lead). {{< /callout >}}

A card travels differently here than in kanban. In kanban one card moves across the `plan → run → sync` columns, while in a factory one card goes whole to one worker and passes the three phases serially inside it. Each phase is spawned by that session as `Agent()` subagents, and write-capable spawns are isolated with `isolation: "worktree"`. A worker runs at most 10 concurrent subagents, and the launcher seeds that value into workers and companion sessions through `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, so N workers sharing one machine's capacity is guaranteed by configuration rather than by operator restraint. Never bring the workers up all at once — start the first, confirm it is actually producing output, then activate the rest.

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
GLM does not support the `auto` permission mode. Use an eligible Claude session for that mode. CG is retired and provides no concurrency alternative.
{{< /callout >}}

## CG retirement and migration

It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session. [CG retirement and migration](/en/multi-llm/cg-mode/) CG is retired; use `moai migrate cg` to preview explicit migration choices.

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

This writes `llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, and `llm.gateway.teammate_provider: inherit`. It removes the old hybrid role assignment; it does not preserve a Claude leader with GLM teammate panes.

The `claude-glm` target describes a Claude leader with GLM teammates in tmux. Its apply and launch paths are currently unavailable because the TEAMMATE integration gate has not passed. Preview is available. Installing tmux or setting `verified: true` does not open this gate.

## Profiles (`-p` flag)

Both launchers accept `-p <name>` to select a named profile, which sets `CLAUDE_CONFIG_DIR` to `~/.moai/claude-profiles/<name>/`. Use this to keep multiple accounts / setting sets separate.

## Isolated worktree (`-w` flag)

Both launchers accept `-w [name]` to start the session inside an isolated git worktree, collapsing the two-step `cd` then launch into a single command.

```bash
moai cc -w feat-login    # Start in .claude/worktrees/feat-login/
moai cc -w               # Auto-generated name
moai glm -w feat-login   # Same for the GLM backend
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

- [CG retirement and migration](/en/multi-llm/cg-mode/)
- [Profile Management](/en/cli-reference/profile)
- [Security Notes](/en/advanced/security-notes) — GLM credential path security model
- [CLI Overview](/en/getting-started/cli)
