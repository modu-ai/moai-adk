# t366 — CLI surface census and Option A feasibility

Tree: worktree `.claude/worktrees/t366`, HEAD `d7010f86a`. Fills the census gap named in
`discovery.md` § Q2.

## How large the affected class is

    $ grep -rn 'Use:[[:space:]]*"' --include='*.go' internal/cli/ | grep -v _test.go | wc -l
    223                     # cobra command definitions

    $ grep -rln 'cobra.Command{' --include='*.go' internal/cli/ | grep -v _test.go | wc -l
    107                     # files defining at least one

    $ grep -rln "findProjectRootFn\|os.Getwd()" --include='*.go' internal/cli/ | grep -v _test.go | wc -l
    68                      # files reading project root or cwd (of 199 non-test files)

Attribution note: 223 counts `Use:` fields, so it includes nested subcommands and slightly
over-counts against "commands a user types". 68 is a FILE count, not a command count. Neither
is a per-subcommand census of "reads the tree and its answer depends on compiled-in rules" —
that census does not exist and would have to be produced if the chosen repair keys on a list
of commands (Option C). What both figures do establish, and all they establish, is that the
class is large enough that a per-command patch is the wrong shape.

## The invocation seam is unoccupied

    $ grep -rn 'PersistentPreRun' --include='*.go' internal/cli/ cmd/ | grep -v _test.go
    internal/cli/inventory.go:255   (a comment referring to root.go wiring)
    internal/cli/root.go:127        worktree.WorktreeCmd.PersistentPreRunE = ...

`rootCmd` (`internal/cli/root.go:18`) defines `Use`, `Short`, `Long`, `Version`, and `Run` —
and NO `PersistentPreRun` / `PersistentPreRunE`. The only existing persistent hook is scoped
to the `worktree` subtree. So Option A adds a root-level hook to an empty slot rather than
composing with an existing one.

Cobra caveat that must reach the SPEC: `PersistentPreRunE` does NOT chain by default — a
subcommand defining its own replaces the parent's. The `worktree` subtree already defines one
at `root.go:127`, so a root-level advisory would NOT fire for `moai worktree ...` unless that
handler is amended too. Any acceptance criterion asserting "every subcommand announces the
lag" must therefore either name `worktree` as an exception or require the chain be made
explicit.

## A skip list already exists

`trivialCommands` (`internal/cli/root.go:46`) lists the paths that must not pay initialization
cost: `--version`, `version`, `-v`, `help`, `--help`, `-h`, `completion`, and the three
launchers `cc` / `cg` / `glm`. The launchers matter for a different reason than cost — they
end in `syscall.Exec` and replace the process, so an advisory printed there is written by a
process about to be discarded. Whatever Option A's skip set turns out to be, this list is the
natural starting point rather than a new one.

## Gaps

- No measurement of what a per-invocation git comparison costs on the hot path. The
  session-start caller bounds the same comparison at 250 ms
  (`internal/hook/session_start_binary_lag.go`, `binaryLagJoinBound`), which is a precedent
  for the shape but not a measurement of the cost.
- No count of how many of the 223 command definitions are reachable as user-typed commands.
