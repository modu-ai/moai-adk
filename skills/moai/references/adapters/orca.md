# Adapter: orca

Orca is a desktop app with a CLI that manages git worktrees, terminals, agent sessions and an embedded browser. MoAI does not copy Orca's instructions: Orca prints a guide that matches the installed version, so read that guide instead of trusting a stored copy.

## When to use it

- Use Orca only when the user chose it for this project and `orca` answers on this machine; installed is not chosen.
- Lane worktrees, reading and sending to lane terminals, searching lane sessions (`orca search`), linking Issues and pull requests to a workspace (`orca reference`), and checking a page in Orca's embedded browser.
- Plain shell, git and HTTP work does not need Orca.
- MoAI uses Orca through its CLI only (user decision, Issue #2634 comment 6845653775).

## Open the version-matched guide first

Run these before relying on any flag, and prefer `--json` for agent calls:

- `orca skills get orca-cli` — the CLI guide for this Orca version (`orca skills list` names the other guides, for example `orchestration`).
- `orca agent-context --json` — the machine-readable command schema; when the help text and the schema disagree, the schema wins.
- `orca status --json` — whether the runtime is reachable.

Never invent a flag from memory. Resolve the executable the way Orca's skill stub does: `ORCA_CLI_COMMAND` when set, then `orca-dev`, then `orca-ide` on Linux, otherwise `orca`.

## Lines MoAI keeps

- Read a terminal before sending to it, and never overwrite another session's unsubmitted input.
- Name every terminal by its handle; without a handle a command acts on the active terminal, which may be someone else's.
- Never answer another session's permission or trust prompt, change its permission mode, stop it, close its terminal or remove its worktree without the user's confirmation.
- `accepted: true` proves only that the input was accepted; read the target again, and never resend on silence.
- Publishing skill links or artifacts stays off unless a human turned it on in Orca's settings; never ask an agent to turn it on.

The full session-control rules for lanes and the leader are in the MODE-AGENT and MODE-LEADER workflows. Orca present is what puts the leader in the command mode, whatever host it runs on; the leader clears a lane only under the clear contract of MODE-LEADER's `supervise` step (Issue #2899 G2, G4).

## When Orca is missing

Use `git worktree` and the host's own terminal instead, and record the Orca step as `not_run` with the reason (`orca` not on PATH, or the runtime not reachable).

## Sources

- https://orca.dev/docs/cli/overview and https://orca.dev/docs/cli/skills (read 2026-11-13; `onorca.dev` redirects there).
- `orca --help`, `orca skills list`, `orca skills get orca-cli` and `orca agent-context --json` of orca 1.4.257 on this machine (read 2026-11-13). `agent-context` is in the CLI but not on the web pages read.
