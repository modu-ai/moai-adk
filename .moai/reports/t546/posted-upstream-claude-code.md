> **Version scope of this report.** Every observation in the probe table below is an
> observation of **specific builds at specific dates** — reporter: Claude Code 2.1.238
> (2026-08-24, from the environment block of the originating report), maintainer
> reproduction sessions: 2026-09-07 (build not recorded) and **2026-09-08 on 2.1.263**
> (fresh observation, refusal text byte-identical). Behavior in other builds is not
> asserted; if a later build has changed this, the report should be read as historical.

**Environment**: Claude Code worktree-isolated session (macOS, worktree under the
project's `.claude/worktrees/`). Originally reported against `moai-adk` by jjjh7401
(https://github.com/modu-ai/moai-adk/issues/1659, filed 2026-08-26) and reproduced
independently by the moai-adk maintainers — including a fresh reproduction on
2.1.263 the day this report was submitted.

## The asymmetry, measured

Paired probes in a worktree-isolated session, differing ONLY in the heredoc body's
content:

| # | Command shape | Result |
|---|---|---|
| 1 | `cat > target <<'EOF'` whose body line is a brace-wrapped JSON pair, e.g. `{"key": "value"}` | **Refused** — `This session is isolated in the worktree …, but this command is too complex to verify that it stays inside the worktree.` The command never executed (the target file was confirmed absent afterward). Re-confirmed 2026-09-08 on 2.1.263; the refusal also fired when the target was **outside** the worktree (`/tmp/…`), so the verdict does not depend on the target path. |
| 2 | Same heredoc shape; the body line instead contains a command substitution, e.g. `value: $(echo x)` | **Executed** — exit 0, the body written through verbatim (quoted-delimiter semantics preserved). Observed 2026-09-07. |

The refusal's own remediation ("Split it into plain, separate commands") does not
apply naturally here: the command IS plain — a single `cat` with a heredoc — and the
offending content is inert body text.

## Why the brace form is provably inert

For a quoted delimiter (`<<'EOF'`), POSIX/bash performs **no expansion of any kind**
inside the body: no parameter expansion, no command substitution, no arithmetic
expansion, no brace expansion. A brace character in such a body is a literal
character and cannot be brace expansion. The guard itself already relies on this
fact: probe 2 shows it folds a command substitution appearing in exactly that
position. Refusing braces in the same position is therefore internally inconsistent —
the analyzer treats one inert construct as folded data and another as live syntax.
(This argument and the original observation are jjjh7401's, from
https://github.com/modu-ai/moai-adk/issues/1659 — credited here as the source.)

Note the asymmetry does NOT ask for any change to **unquoted**-delimiter bodies
(`<<EOF`): those bodies genuinely expand, and erring toward refusal there is correct
behavior.

## Suggested directions (either would resolve it)

1. **Fold braces in quoted-delimiter heredoc bodies** — apply the same folding the
   analyzer already performs for command substitutions in that position.
2. **Name the offending construct in the refusal message** — e.g. point at the
   heredoc body / the brace so users can restructure (move the content to a file
   write) instead of guessing. The current message invites users to hunt the
   complexity through an entire multi-line command.

## Reproduction

Any worktree-isolated session; a single `cat`-with-quoted-delimiter-heredoc whose
body contains a `{...}` line reproduced the refusal in every observed session —
2026-08 (reporter, 2.1.238), 2026-09-07 (maintainer session), and 2026-09-08
(2.1.263, target path independence additionally observed). The companion repo
(moai-adk) keeps a pinned evidence ledger with the verbatim refusal, both probes,
and the non-execution proof (target file absent) should more detail help.
