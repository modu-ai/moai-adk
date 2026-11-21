# Adapter: orca

Orca is a desktop app with a CLI that manages git worktrees, terminals, agent sessions and an embedded browser. MoAI does not copy Orca's instructions: Orca prints a guide that matches the installed version, so read that guide instead of trusting a stored copy.

## When to use it

- Use Orca only when the user chose it for this project and `orca` answers on this machine; installed is not chosen.
- Lane worktrees, reading and sending to lane terminals, searching lane sessions (`orca search`), linking a pull request to a workspace (`orca reference`), and checking a page in Orca's embedded browser.
- Plain shell, git and HTTP work does not need Orca.
- MoAI uses Orca through its CLI only.

## Open the version-matched guide first

Run these before relying on any flag, and prefer `--json` for agent calls:

- `orca skills get orca-cli` — the CLI guide for this Orca version (`orca skills list` names the other guides, for example `orchestration`).
- `orca agent-context --json` — the machine-readable command schema; when the help text and the schema disagree, the schema wins.
- `orca status --json` — whether the runtime is reachable.

Never invent a flag from memory. Resolve the executable the way Orca's skill stub does: `ORCA_CLI_COMMAND` when set, then `orca-dev`, then `orca-ide` on Linux, otherwise `orca`.

## Messaging between sessions

When the host has no cross-session message (or it is refused or held), send with `orca orchestration send` to the session's terminal handle, and let the leader read replies with `orca orchestration check`. This is the second step of the channel order in `flows/team-messages`; the ledger note alone is the third. A message is a notice and the ledger is the record.

## Lines MoAI keeps

- Read a terminal before sending to it, and never overwrite another session's unsubmitted input.
- Name every terminal by its handle; without a handle a command acts on the active terminal, which may be someone else's.
- Never answer another session's permission or trust prompt, change its permission mode, stop it, close its terminal or remove its worktree without the user's confirmation.
- `accepted: true` proves only that the input was accepted; read the target again, and never resend on silence.
- Publishing skill links or artifacts stays off unless a human turned it on in Orca's settings; never ask an agent to turn it on.

## What Orca's pages say, and what they leave out

Read on 2026-11-20 (UTC, from `date -u` in the same step as each read): https://orca.dev/docs/cli/orchestration (2026-11-20T22:46:19Z), https://orca.dev/docs/cli/reference (2026-11-20T22:46:24Z), https://orca.dev/docs/agents/native-chat (2026-11-20T22:46:39Z) and https://orca.dev/docs/cli/automations (2026-11-20T22:46:40Z). None of the four pages names a version; `orca --version` printed 1.4.267 on this machine (2026-11-20T22:47:36Z), so a line below that cites the local help is the installed CLI's own text, not a document.

- Mail and `check`: the orchestration page says "Default check is the bound Run's oldest unacked Delivery (FIFO)" and "--peek / --all do not consume mail" (https://orca.dev/docs/cli/orchestration). The local help of `orca orchestration check --help` adds that a consuming check marks the batch it returns as handled, that no `--ack` is needed, and that a script that only watches an inbox must use `--peek` and never `--ack`, or it takes mail the agent never sees; a lost answer is replayed with the `--retry-request` id it printed.
- The pages do not document `delivered_at`, any mail state (stored, delivered, turn start, work done) or what happens to Runs, Tasks and mail after an Orca restart, and the local help of `orca orchestration send --help` does not mention `delivered_at` either. A rule that needs one of them is written as measured on this machine and marked unverified; a null `delivered_at` is never a reason to type the same text again.
- Tracked work and one-off prompts: the orchestration page says to use a Run, Tasks and workers "when you need ownership, completion tracking, or a DAG" and "For one-off prompts, use orca terminal send" (https://orca.dev/docs/cli/orchestration). So supervised work goes as a Run, a Task and a Dispatch, a plain notice goes as mail, and a typed line to a lane you are watching goes through `orca terminal send` after the read-before-send lines above.
- Unverified facts. A missing `hostScope`, native chat and sleep or restore are unverified, because the reference page says "Treat a missing host identity or scope as unverifiable, not local" and the native-chat page says of sleep only that "closed chats or sleeping workspaces wait until reopened or awakened" (https://orca.dev/docs/cli/reference, https://orca.dev/docs/agents/native-chat). A missing `hostScope` is not proof that a lane's terminal exited, and a lane keeps its card and its reservation while a fact is unverified; an unverified fact never authorises a recovery.
- Screens: a screen read is the frame that is rendered now, and a cursor of `terminal read` is a position in the output stream: "Use --screen when you need the currently rendered frame" and "Save nextCursor from one stream read, then pass it back with --cursor" (https://orca.dev/docs/cli/reference); the pages report no caret position.

## When Orca is missing

Use `git worktree` and the host's own terminal instead, and record the Orca step as `not_run` with the reason (`orca` not on PATH, or the runtime not reachable).

Orca is optional for every workflow. Nothing in the `moai` tool runs `orca`, and a workflow names Orca only in steps that carry their own "without Orca" rule. This table says what each workflow does on each side; `works` runs as written, `replaced` runs through another channel, `unsupported` has no way yet and is left for the leader to track.

| Workflow | With Orca | Without Orca | Host differences |
|---|---|---|---|
| READ | works (Orca not used) | works | parallel reads: Claude Code several agents in one message |
| PROJECT | works | works | entry `/moai project` |
| UPDATE | works when an Orca session has `CLAUDE_CONFIG_DIR` set: the update preview and the skill copy follow that variable, so `/moai` lands in that account's own folder, the preview names it and the next preview has nothing left to apply (measured in a temporary profile; no live Orca profile session was run, and that Orca sets the variable is unverified) | works | files go to the host's own settings folder, which an Orca profile can move away from `~/.claude` |
| PLAN | works | works | the `moai-plan` agent runs through the host's own subagent tool |
| RUN | works (Orca terminals and worktrees are not used) | works | isolation: Claude Code `isolation: worktree` |
| AUDIT | works | works | read-only is enforced by the agent file on Claude Code only; the verdict's first line has one format on every host |
| SYNC | works | works | none |
| WORK | works (Orca not used) | works | none |
| LEADER | works: the host's message tool carries the dispatch first, and `moai lane-send` (a lane terminal read and a send) is the fallback, with `/clear` and screen reads | replaced: the host's message tool carries the dispatch | the cross-session message has no `moai` command on any host yet |

Nothing is unsupported with Orca today. The settings folder of an Orca profile (`CLAUDE_CONFIG_DIR`) is followed by the update preview and the skill copy (host-folder resolver; measured with `update-plan` and `update-apply` in a temporary profile with a clean environment: the preview names the profile's folder, `SKILL.md` lands there, nothing is created in the home's `.claude` folder and the second preview has nothing to apply); no live Orca profile session was run, so that Claude Code in such a session reads the moved folder rests on https://code.claude.com/docs/en/claude-directory.md, the page the resolver cites. Orca's pages say each Claude account keeps its own config folder but never name `CLAUDE_CONFIG_DIR` (https://orca.dev/docs/settings, read 2026-11-21T01:32:47Z; https://orca.dev/docs/agents/supported, read 2026-11-21T01:32:45Z), so that Orca sets it is unverified here; it is the report behind Issue #2936. A workflow without a row, or a row without one of the three words, fails the repository scan.

## Sources

- https://orca.dev/docs/cli/overview and https://orca.dev/docs/cli/skills (read 2026-11-13; `onorca.dev` redirects there).
- The four pages above and, of orca 1.4.267 on this machine, `orca --version`, `orca orchestration check --help`, `orca orchestration send --help` and `orca skills get orchestration` (read 2026-11-20; the times are in the section above).
- `orca --help`, `orca skills list`, `orca skills get orca-cli` and `orca agent-context --json` of orca 1.4.257 on this machine (read 2026-11-13). `agent-context` is in the CLI but not on the web pages read.
