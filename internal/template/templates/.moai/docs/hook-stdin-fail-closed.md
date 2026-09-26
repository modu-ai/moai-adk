# Hook stdin parse failure: fail-closed on decision events

Operator document named in the hook's own denial reason by its path,
`.moai/docs/hook-stdin-fail-closed.md`. When a hook denies a decision with a
reason that starts `fail-closed: hook stdin could not be parsed as JSON` and
ends `(.moai/docs/hook-stdin-fail-closed.md)`, this page explains what
happened and how to recover.

## What changed

Every `moai hook <event>` invocation reads its payload from stdin as JSON.
When that payload cannot be parsed — malformed JSON, a truncated stream, a
payload cut off by the 5 MiB read limit, or a JSON nesting depth beyond the
decoder's limit — the dispatcher used to answer every event the same way: a
stderr warning plus a default (`{}`) output at exit 0. For an **observation**
event (`PostToolUse`, `SessionStart`, etc.) that is still correct — no guard
was going to act on the payload anyway, so a default output changes nothing.

For a **decision** event — `PreToolUse`, `PermissionRequest`, `Stop`,
`UserPromptSubmit` — a default output of `{}` means the hook expresses no
opinion, so the host's normal permission flow decides instead — which, in a
mode that does not prompt, lets the call through. A hook whose stdin fails to
parse on one of these four events now denies instead, still at exit 0, still
without cobra usage noise or a fake tool failure.

Codex's `Stop` event is exempt from this fail-closed behavior and keeps the
default-output path, because the Codex host has no bound on how many
consecutive `Stop` blocks it will accept — an unparseable payload there could
otherwise deny every turn indefinitely with no way for the session to end.

## Stop under Claude Code: MoAI's own limit

Claude Code's `Stop` is denied fail-closed too, but MoAI does not rely on the
host's limit on consecutive `Stop` blocks to end that loop. Measured, that
limit ended the turn only when no tool use came in between the blocks, and a
session started with a raised limit (kanban, factory, or an infinite goal)
kept going to its turn limit. So MoAI counts the loop itself:

- **N=8.** The first 8 consecutive `Stop` calls whose stdin cannot be parsed
  are denied. From the 9th on, the hook answers with no opinion (`{}`), which
  lets the turn end, writes one stderr line, and records the release in
  `.moai/logs/codex-adapter.jsonl` under the key
  `stdin-parse-stop-cap-released`.
- **The release persists.** Once past 8, every further unparseable `Stop`
  keeps getting the no-opinion answer — a payload that cannot be parsed does
  not reveal where one turn ends and the next begins. Tool use in between does
  not reset the count.
- **Only two things reset it:** a `Stop` whose stdin parses, which deletes
  the count, and 60 minutes passing since the count was last updated, after
  which the record counts as absent and is swept.
- **What is counted together.** The count is kept per session: the
  `CLAUDE_CODE_SESSION_ID` value in the hook's environment when it is set and
  not blank, otherwise the Claude Code process that owns the session (found by
  walking up from the hook past any wrapper shells). If neither can be
  determined, the count is not applied and every unparseable `Stop` stays
  denied.
- **Where the count lives.** One small JSON file per session under
  `.moai/state/stop-parse-cap/`, named by a hash of the session key — the
  session id itself is not stored. If that directory or a record cannot be
  used (not a directory, a symbolic link, unwritable), the count is not
  applied and the denial stays; the hook says so on stderr.
- **Independent of the host's limit.** The limit of 8 is fixed in MoAI. It
  does not read `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` or any other setting, and
  no setting changes it.

Only `Stop` is limited this way. The other three decision events block a tool
call or a prompt rather than the end of a turn, so their denial cannot keep a
turn going on its own. Under Codex none of this applies — Codex's `Stop` is
exempt, as described above.

## What the denial reason tells you, and does not

The denial's reason string carries the `fail-closed` marker, a fixed
statement that the cause was a stdin parse failure, and this document's path.
Under Claude Code it also tells the model not to edit hook scripts or
settings files to get past the denial, and to stop and tell a human
operator: a model handed the bare denial has been seen reading the hook
configuration and trying to edit it. The reason carries no recovery steps and
none of the failing payload's content — the reason string is read by the
model that produced the tool call, so recovery instructions or excerpted
input would hand a model-controlled channel a way to see, or work around, why
it was denied.

## Recovery

There is no configuration switch that turns this denial back into a pass —
an earlier design that allowed opting out was found to be defeatable (a value
added and later removed to a settings file still reached the hook's
environment for the rest of that session) and was removed. If a decision
event's stdin persistently fails to parse — most likely because the host
changed its payload format in a way MoAI does not yet handle — recover with
one of:

1. **Update MoAI.** Run `moai update` to pick up a binary that understands
   the new payload shape. This is the preferred recovery: it restores the
   guard rather than removing it.
2. **Disable hooks**, only if an update is not immediately available and the
   denial is blocking work you need to do right now. On Claude Code, set
   `disableAllHooks: true` in the relevant `settings.json`. This turns off
   every hook in that settings file's scope, not only the one that is
   denying — including MoAI's observation hooks — so treat it as a temporary
   escape hatch, not a fix. Re-enable it once the underlying cause is
   resolved. Codex-side hook disablement is host-specific and is not
   documented here.

If you see this denial only occasionally and updating clears it, no further
action is needed — the most plausible model-driven cause is a tool call
whose input exceeds the 5 MiB read limit, and catching that case is exactly
what this behavior exists to do.

## Why fail-closed, in one line

A wrong denial is loud — the model and the operator both see the reason and
can act on it. A wrong silent pass is not: nothing records that a guard was
skipped. This asymmetry is why decision events fail closed even though it
costs more to recover from a persistent parse failure than an environment
toggle would.
