# Internal: git-safety

When to read: before you commit, push, merge or switch branches in a worktree.

1. Write only while you hold the card: run `moai card check` for the files you are about to write before every commit and every push. A refusal stops the write and is reported as it is (the WORK route, `/moai agent`).
2. Keep the trailer paragraph last: the `Moai-Spec` and `Moai-Session` trailers and the attribution line, nothing after them; no card number goes into a commit message, a branch name or a pull request. Confirm each new commit with `git interpret-trailers --parse`. Write the message in the project's `language.commit_messages` (read it from the project profile).
3. Work in your own worktree, with its one build folder inside it (`workflows/run.md`, step 1). Never write in another worktree and never share a build folder with another worktree; after switching worktrees, trust only a run whose `Compiling` line names the worktree you are checking. New worktrees go to the project's `.claude/worktrees/<name>` (`workflows/run.md`, step 1); never move an existing worktree, and never touch `~/.moai/worktrees/projects/`, which the Go generation uses.
4. Never force-push, `reset --hard`, `clean`, delete a branch or rewrite history that others may have. Never use a bare `git stash`; the stash stack is shared between worktrees.
5. A lane never pushes (`flows/next-flow`). After a push the flow does make — the landing's remote copy, or a release pull request's branch — read the remote head back (`git ls-remote`); a release pull request with a conflict runs no CI until the conflict is resolved.
6. Take every SHA you write down from command output, never typed from memory.
7. Enter and leave a worktree with the host's built-in worktree tool where it has one; where it has none, name the worktree on every git call as `git -C <worktree> <verb>`, one git command per call (user decision). The commit guard reads the branch in the folder `-C` names; a line that commits or pushes in two folders is refused, so run each on its own line.
8. In a checkout other sessions share, stage by explicit path — never `git add -A`, `git add .` or `git commit -a` — and re-read `git status`, the branch and `HEAD` immediately before staging, committing or pushing; never switch, create, reset, stash, rebase or merge branches in the shared primary checkout.

| Host | Enter and leave a worktree | Git inside it |
| --- | --- | --- |
| Claude Code | `EnterWorktree` (`path` for an existing one) and `ExitWorktree`; subagents with `isolation: worktree`. `EnterWorktree` with a `name` branches from the remote default branch, so for another base create the worktree with `git worktree add` first and enter it by `path`; give `git worktree add` an absolute path and follow `workflows/work.md` (never by `name`). | Plain `git <verb>`, or `git -C <own worktree> <verb>`. Its built-in guard refuses a line that runs git next to a shell expansion (`$VAR`, `$?`, `$(…)`) or `-C` to the shared checkout; put such work in a script file under `target/`. |
| Cursor CLI | `-w` / `--worktree` at start only. | Inside a session started there, plain `git <verb>`; otherwise `git -C <worktree> <verb>`. |
| Kiro CLI | None documented. | `git -C <worktree> <verb>`. |

Sources, read 2026-11-13: https://code.claude.com/docs/en/worktrees.md and https://code.claude.com/docs/en/tools-reference.md; https://cursor.com/docs/cli/reference/parameters; https://kiro.dev/docs/reference/cli-commands/. The Claude Code guard behaviour was measured on 2.1.323.
