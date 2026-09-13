## Summary

`moai worktree clean --json` is documented as a read-only reporter. It emits no JSON at all, and it mutates state: it prints `✓ Cleaned stale worktree references` and performs a prune.

Found while working card **t567** (reproducing a worktree that disappeared without a removal command). The flag was reached for precisely because its help promised it would look without touching — which is the reflex this defect punishes.

## What the help says

```
$ moai worktree clean --help
    --json         Report every non-protected worktree and its state as JSON; removes nothing
```

## What actually happens

```
$ moai worktree clean --json > /tmp/t567-clean.json 2>/tmp/t567-clean.err
$ echo "exit=$?"
exit=0
$ wc -c /tmp/t567-clean.json /tmp/t567-clean.err
     296 /tmp/t567-clean.json
       0 /tmp/t567-clean.err
$ cat /tmp/t567-clean.json
╭───────────────────────────────────────╮
│  ✓ Cleaned stale worktree references  │
╰───────────────────────────────────────╯
```

Observed vs. documented:

| Axis | Documented | Observed |
|---|---|---|
| Output format | JSON report of every non-protected worktree and its state | A human-rendered box banner. Zero JSON. 296 bytes total. |
| Side effects | "removes nothing" | Prints `Cleaned stale worktree references`; a prune ran |
| Exit code | — | 0 |
| stderr | — | empty |

## Blast radius, as measured

A second invocation was diffed against an inventory taken immediately before and after:

```
$ git worktree list | awk '{print $1}' | sort > /tmp/before.txt
$ moai worktree clean --json
╭───────────────────────────────────────╮
│  ✓ Cleaned stale worktree references  │
╰───────────────────────────────────────╯
$ git worktree list | awk '{print $1}' | sort > /tmp/after.txt
$ comm -23 /tmp/before.txt /tmp/after.txt
(no output)
```

No **live** worktree was removed. The mutation observed is a prune of dangling administrative entries — entries whose directory was already gone. That is still a mutation, and it is still the opposite of "removes nothing".

This measurement bounds one run on one machine. It does not establish that the path can never remove a live worktree.

## Why it matters

`--json` is the flag an operator or an agent reaches for to *inspect* worktree state before deciding anything — the safe half of a command whose other half deletes directories. A flag advertised as inert that prints a "Cleaned" banner trains the reader either to distrust the help or to ignore the banner. Both are worse than a flag that plainly says what it does.

Downstream, any tooling that parses `--json` output gets a box-drawing banner where an object was promised.

## Environment

- `moai-adk v3.2.0-rc.8`
- `git 2.50.1 (Apple Git-155)`
- macOS (darwin 25.6.0)
- Invoked as the installed `moai` on PATH; the build was not pinned to a source tree, so this is attributed to that installed build.

## Expected

Either of these would be consistent; the current state is neither:

1. `--json` emits the documented JSON report and performs no prune, or
2. the help stops claiming `--json` removes nothing and names the prune.

## Related

- Card t567 — verdict at `.moai/reports/t567/verdict.md` (claim 4), which also records two sweep paths (`prMergeCleanup`, `moai worktree clean`) that enumerate every worktree through `git worktree list --porcelain`, L1 `.claude/worktrees/*` trees included.

🗿 MoAI
