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
| `-f, --factory` | Enter as the **factory leader**. It takes no argument. The leader deals the cards the operator picks, whole, to free lanes over cross-session messages, and lanes join with `-l` |
| `-l, --lane` | Join the running factory as a **lane**, claiming the next `lane-<n>` number automatically (one past the highest live lane). It takes no argument and is refused when no factory is running. `moai glm -l` and `moai codex -l` behave the same |
| `--leader <name>` | Used with `-l` or `--lane` only. Picks which leader session to join (default `leader`; the former spelling `lead` is refused). When the run's record is missing or retired while a live leader exists, the join verifies that leader (pid + process-start) and restores its run |
| `--factory-run <run-id>` | With `-l`: join the run with this id. It cannot be combined with `--leader` |
| `--clear-policy <value>` | With `moai cc -l` / `moai glm -l`: how the lane clears its context after a card (`clear-each` default, `clear-when-full`, `relaunch`) |
| `--no-auto-dispatch` | With `moai cc -l` / `moai glm -l`: start the lane in manual mode. The default is a self-dispatch lane that leases the next queued card itself |

{{< callout type="info" >}} There are two entry tokens, `-f` (leader) and `-l` (lane), and neither takes an argument. A launch carries one token only, so `-f` together with `-l` is an error, and every form that attaches a value (`-f <value>`, `-l lane-2`) or asks for a leader on Codex (`-f` on `moai codex`) is refused with a one-line error. The retired `-k` entry is refused the same way, with a message pointing to `-f` and `-l`. For the full contract see [Factory Mode](/en/advanced/factory-mode) and [manager-lead Leader Coordinator](/en/advanced/manager-lead). {{< /callout >}}

A card goes whole into one lane and passes the three phases `plan → run → sync` serially inside it. Each phase is spawned by that session as `Agent()` subagents, and write-capable spawns are isolated with `isolation: "worktree"`. A lane runs at most 10 concurrent subagents, and the launcher plants this value in the lane session as `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, so N lanes sharing the machine's capacity is guaranteed by configuration rather than by operator restraint. Do not switch on all lanes at once: bring up the first lane, confirm it has actually started producing output, and then start the rest.

The leader and the lanes can run on different backends. The backend mix is decided by token availability first. One starting point is to place Opus only where judgment is heavy and run implementation-heavy lanes on GLM. A different combination, or unifying on a single backend, is equally fine.

The permission mode is one of `default`, `acceptEdits` (the `moai init` default), `plan`, `auto`, `bypassPermissions`, `dontAsk`. The `auto` mode runs a background classifier that inspects actions; supported plans and models are listed in the [Claude Code permission modes documentation](https://code.claude.com/docs/en/permission-modes).

## moai glm — GLM backend

```bash
moai glm setup <api-key>   # Save API key (first time only)
moai glm --key <api-key>   # Save API key via flag (same storage as setup)
moai glm                   # Run with the GLM backend
moai glm -p work           # Run with the 'work' profile
moai glm status            # Check credential status
```

Reads GLM credentials from `~/.moai/.env.glm`, injects environment variables such as `ANTHROPIC_AUTH_TOKEN` and `ANTHROPIC_BASE_URL`, and then runs Claude Code.

| Subcommand | Description |
|-------------|------|
| `moai glm setup [api-key]` | Save the GLM API key |
| `moai glm --key <api-key>` | Save the GLM API key via flag (both forms write the same file; the last save wins) |
| `moai glm status` | Show the current GLM credential status |

The Jev (TypeSafe) credential is stored with `moai jev --key <credential>` — written to `~/.moai/.env.typesafe` (mode 0600), the same file `moai doctor` and the web console read. Running `moai jev` without `--key` prints help only.

{{< callout type="warning" >}}
GLM does not support the `auto` permission mode. Use an eligible Claude session for that mode. CG is retired and provides no concurrency alternative.
{{< /callout >}}

## CG retirement and migration

It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session. [CG retirement and migration](/en/multi-llm/cg-mode/) CG is retired; use `moai migrate cg` to preview explicit migration choices.

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

This writes `llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, and `llm.gateway.teammate_provider: inherit`. It removes the old hybrid role assignment; it does not preserve a Claude leader (CG leader) with GLM teammate panes.

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
