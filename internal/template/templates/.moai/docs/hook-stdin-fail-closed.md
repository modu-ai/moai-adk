# Hook stdin parse failure: fail-closed on decision events

Operator document named in the hook's own denial reason by its path,
`.moai/docs/hook-stdin-fail-closed.md`. When a hook denies a decision with
the reason `fail-closed: hook stdin could not be parsed as JSON
(.moai/docs/hook-stdin-fail-closed.md)`, this page explains what happened and
how to recover.

## What changed

Every `moai hook <event>` invocation reads its payload from stdin as JSON.
When that payload cannot be parsed — malformed JSON, a truncated stream, a
payload cut off by the 5 MiB read limit, or a JSON nesting depth beyond the
decoder's limit — the dispatcher used to answer every event the same way: a
stderr warning plus a default (`{}`) output at exit 0. For an **observation**
event (`PostToolUse`, `SessionStart`, etc.) that is still correct — no guard
was going to act on the payload anyway, so a default output changes nothing.

For a **decision** event — `PreToolUse`, `PermissionRequest`, `Stop`,
`UserPromptSubmit` — a default output of `{}` reads to the host as "no
opinion," which on an unconfirmed permission mode is equivalent to an
approval. A hook whose stdin fails to parse on one of these four events now
denies instead, still at exit 0, still without cobra usage noise or a fake
tool failure.

Codex's `Stop` event is exempt from this fail-closed behavior and keeps the
default-output path, because the Codex host has no bound on how many
consecutive `Stop` blocks it will accept — an unparseable payload there could
otherwise deny every turn indefinitely with no way for the session to end.

## What the denial reason tells you, and does not

The denial's reason string carries exactly three things: the `fail-closed`
marker, a fixed statement that the cause was a stdin parse failure, and this
document's path. It carries no recovery steps and none of the failing
payload's content — the reason string is read by the model that produced the
tool call, so recovery instructions or excerpted input would hand a
model-controlled channel a way to see, or work around, why it was denied.

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
action is needed — an occasional truncated or malformed payload from a
model-constructed tool call is expected and is exactly the case this
behavior exists to catch.

## Why fail-closed, in one line

A wrong denial is loud — the model and the operator both see the reason and
can act on it. A wrong silent pass is not: nothing records that a guard was
skipped. This asymmetry is why decision events fail closed even though it
costs more to recover from a persistent parse failure than an environment
toggle would.
