# OPTIMIZE

Entered with `/moai optimize` in Claude Code or `$moai optimize` in Codex CLI (leader decision on Issue #2724). It shows how much space the repository's leftovers hold and which of them could go. It is a preview only: this route deletes, moves and writes nothing.

## Input

The project, the local ref of its integration branch, the build-folder roots the user names (none by default), and what the forge says about the worktrees' branches — independent-audit verdicts, live claims and Issues on HOLD — each with its source.

## Output

The preview table of `moai optimize-plan`, its totals, and, for each kind of row, what applying it would need.

## Procedure

1. **Integration ref.** Use the local ref of the integration branch (for example `origin/<branch>`). The preview reads no network; when the ref may be stale, say so instead of judging around it.
2. **Observations.** Read from the forge, read-only, and write one JSON file in an ignored folder of the project: `schema_version` 1, `audits` {`source`, `items` [{`pr`, `head`, `verdict_comment`, `verdict`}]} with the 40-hex head each verdict names and `PASS` or `FAIL`, `claims` {`source`, `items` [{`issue`, `claim_comment`}]} for every live claim, and `holds` {`source`, `items` [{`issue`}]} for Issues on HOLD or `needs:decision`. A part you could not read stays out of the file; the preview then keeps the rows that need it.
3. **Build roots.** Ask which folders outside the worktrees hold build output to list (for example `/tmp` for `/tmp/moai-rc-target-*`). There is no default root; list none the user did not name.
4. **Run.** `moai optimize-plan <project> --integration-ref <ref> [--build-root <dir>]… [--observations <file>]`, with `--retention-days` only when the user gives one (default 7). Quote the envelope's `ok`; a failure is reported as it is.
5. **Show.** One table: kind, path, size, last change, retention, action, reasons and how to restore it; then `totals` with `freeable_bytes`. Say plainly that a worktree is a `remove` row only with an audit `PASS` for its exact HEAD, that a squash merge counts as merged only in this preview, and that time alone never makes a worktree a candidate.
6. **Stop.** Apply nothing in this route. When the user wants space back, name the next step and wait for their yes: worktree rows go through `moai worktree-gc-plan` and then `moai worktree-gc-apply` with that plan's digest and the user's permission reference — it counts only ancestor merges, so a squash-merged worktree stays there; build folders outside the worktrees and reports have no apply step yet, and are never removed by a shell command instead.

## Exceptions and Permissions

Deleting is a `guarded` choice: only the user decides it, after seeing this preview. Never touch the real HOME, keys or global settings, never delete a file, branch or worktree from this route, and never stand in for the missing apply step with `rm`, `git worktree remove` or `git clean`. A forge read that failed is reported as unread, never as an empty list.

## Completion Check

The table shown is the output of one `optimize-plan` run on the named ref and observations, quoted with its `preview_digest`. Never report space as freed, or any write, by this route.
