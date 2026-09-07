# Upstream issue draft — Claude Code worktree guard: heredoc brace asymmetry

> DRAFT — do not submit from this card. Submission is the maintainer's act
> after the card lands on the remote (card t512 note). Target: Claude Code
> upstream. — moai-adk lane, 2026-09-07

---

**Title**: Worktree-isolation Bash check folds command substitutions in
quoted-delimiter heredoc bodies but refuses braces there, though such bodies
cannot expand

**Environment**: Claude Code worktree-isolated session (macOS, worktree under
the project's `.claude/worktrees/`). Originally reported against
`moai-adk` by jjjh7401 (issue #1659, 2026-08-26) and reproduced independently
by the moai-adk maintainers.

## The asymmetry, measured

Paired probes in a worktree-isolated session, differing ONLY in the heredoc
body's content:

| # | Command shape | Result |
|---|---|---|
| 1 | `cat > target.txt <<'EOF'` whose body line is a brace-wrapped JSON pair, e.g. `{"key": "value"}` | **Refused** — `This session is isolated in the worktree …, but this command is too complex to verify that it stays inside the worktree.` The command never executed (the target file was confirmed absent afterward). |
| 2 | Same heredoc shape; the body line instead contains a command substitution, e.g. `value: $(echo x)` | **Executed** — exit 0, the body written through verbatim (quoted-delimiter semantics preserved). |

The refusal's own remediation ("Split it into plain, separate commands") does
not apply naturally here: the command IS plain — a single `cat` with a heredoc
— and the offending content is inert body text.

## Why the brace form is provably inert

For a quoted delimiter (`<<'EOF'`), POSIX/bash performs **no expansion of any
kind** inside the body: no parameter expansion, no command substitution, no
arithmetic expansion, no brace expansion. A brace character in such a body is
a literal character and cannot be brace expansion. The guard itself already
relies on this fact: probe 2 shows it folds a command substitution appearing
in exactly that position. Refusing braces in the same position is therefore
internally inconsistent — the analyzer treats one inert construct as folded
data and another as live syntax. (This argument and the original observation
are jjjh7401's, from moai-adk issue #1659 — credited here as the source.)

Note the asymmetry does NOT ask for any change to **unquoted**-delimiter
bodies (`<<EOF`): those bodies genuinely expand, and erring toward refusal
there is correct behavior.

## Suggested directions (either would resolve it)

1. **Fold braces in quoted-delimiter heredoc bodies** — apply the same
   folding the analyzer already performs for command substitutions in that
   position.
2. **Name the offending construct in the refusal message** — e.g. point at
   the heredoc body / the brace so users can restructure (move the content to
   a file write) instead of guessing. The current message invites users to
   hunt the complexity through an entire multi-line command.

## Reproduction

Any worktree-isolated session; a single `cat`-with-quoted-delimiter-heredoc
whose body contains a `{...}` line reproduced the refusal in every observed
session (the maintainer reproduction session here, plus the reporter's
independent observations). The companion repo (moai-adk) keeps a pinned
evidence ledger with the verbatim refusal, both probes, and the non-execution
proof (target file absent) should more detail help.
