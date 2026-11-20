# Adapter: orca

Orca is a desktop app with a CLI that manages git worktrees, terminals, agent sessions and an embedded browser. MoAI does not copy Orca's instructions: Orca prints a guide that matches the installed version, so read that guide instead of trusting a stored copy.

## When to use it

- Use Orca only when the user chose it for this project and `orca` answers on this machine; installed is not chosen.
- Lane worktrees, reading and sending to lane terminals, searching lane sessions (`orca search`), linking Issues and pull requests to a workspace (`orca reference`), and checking a page in Orca's embedded browser.
- Plain shell, git and HTTP work does not need Orca.
- MoAI uses Orca through its CLI only (user decision).

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

The full session-control rules for lanes and the leader are in the MODE-AGENT and MODE-LEADER workflows. Orca present is what puts the leader in the command mode, whatever host it runs on; the leader clears a lane only under the clear contract of MODE-LEADER's `supervise` step.

## When Orca is missing

Use `git worktree` and the host's own terminal instead, and record the Orca step as `not_run` with the reason (`orca` not on PATH, or the runtime not reachable).

Orca is optional for every workflow. Nothing in the `moai` tool runs `orca`, and the two modes name Orca only in steps that carry their own "without Orca" rule. This table says what each workflow does on each side; `works` runs as written, `replaced` runs through another channel, `unsupported` has no way yet and is left for the leader to track.

| Workflow | With Orca | Without Orca | Host differences |
|---|---|---|---|
| READ | works (Orca not used) | works | parallel reads: Claude Code several agents in one message, Codex one agent at a time and wait for all |
| SETUP | unsupported: an Orca profile starts Claude Code with `CLAUDE_CONFIG_DIR` set to its own folder, and `moai update` looks only at `~/.claude`, so `/moai` can be missing in that session while the preview says nothing to apply (Issue #2936) | works | files go to the host's own settings folder, which an Orca profile can move away from `~/.claude` |
| PROJECT | works | works | entry `/moai project` (Claude Code), `$moai project` (Codex CLI) |
| UPDATE | unsupported: the same gap as SETUP; `moai update` does not see the settings folder of an Orca profile (Issue #2936) | works | same settings-folder point as SETUP |
| PLAN | works | works | the `moai-plan` agent runs through the host's own subagent tool |
| RUN | works (Orca terminals and worktrees are not used) | works | isolation: Claude Code `isolation: worktree`; Codex, Kiro and Cursor work in one worktree, one change at a time |
| VERIFY | works | works | heavy checks take turns through `heavy-slot`, shared by the whole machine |
| AUDIT | works | works | read-only is enforced by the agent file on Claude Code only; the verdict's first line has one format on every host |
| OPTIMIZE | works | works | entry `/moai optimize` (Claude Code), `$moai optimize` (Codex CLI) |
| SYNC | works | works | none |
| MODE-AGENT | works: a leader reads and sends lane screens through Orca | replaced: `lane-next` self-pick and Issue and pull request comments; the wake is Claude Code `/loop`, Codex `loop_line` and the Stop hook | none |
| MODE-LEADER | works: the leader reads and sends lane screens through Orca | replaced: with a timer (Claude Code `/loop 5m`) it leads through Issue and pull request comments alone, with no `clear_redispatch` and no `restart`; with neither Orca nor a timer it refuses to enter; unsupported on Codex, which has no in-session timer | none |

Unsupported today, tracked in Issue #3005: a Codex or OMP leader's periodic run without Orca. Settled in Issue #3005 P10 and P11 (OMP 18.18.7, read in its installed documents and source on 2026-11-15; no live OMP session was run): the entry is `/skill:moai <text>` and the text after the name reaches the skill as plain text; OMP reads skills from `~/.agents/skills` and from `.claude/skills`, `.agents/skills` and `.omp/skills` of the project, and from `~/.claude/skills` only when `enabledProviders` names `claude`, `skills.enableClaudeUser` is on or `CLAUDE_CONFIG_DIR` is set; an OMP lane keeps going with goal mode and ends it with the `goal` tool's `drop`, since OMP has no interval timer and no empty-read count. Tracked in Issue #2936: the settings folder of an Orca profile (`CLAUDE_CONFIG_DIR`), which `moai update` does not see, so SETUP and UPDATE are unsupported with Orca until that repair lands. The table is pinned by `tests/orca_optional_3005.rs`: a workflow without a row, or a row without one of the three words, fails the test.

## Sources

- https://orca.dev/docs/cli/overview and https://orca.dev/docs/cli/skills (read 2026-11-13; `onorca.dev` redirects there).
- `orca --help`, `orca skills list`, `orca skills get orca-cli` and `orca agent-context --json` of orca 1.4.257 on this machine (read 2026-11-13). `agent-context` is in the CLI but not on the web pages read.
