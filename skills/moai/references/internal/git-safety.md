# Internal: git-safety

When to read: before you commit, push, merge or switch branches in a worktree.

1. Write only on an `owner` verdict: run `moai-cli claim-verify` before every commit and every push. Any other verdict stops the write and is reported as it is (MODE-AGENT workflow, `claim-verify` step).
2. Keep the trailer paragraph last: the claim token `Moai-Claim: <issue>#<claim id>` and the attribution line, nothing after them. Confirm each new commit with `git interpret-trailers --parse`.
3. Work in your own worktree with your own build folder. Never write in another worktree and never share a build folder between worktrees.
4. Never force-push, `reset --hard`, `clean`, delete a branch or rewrite history that others may have. Never use a bare `git stash`; the stash stack is shared between worktrees.
5. After every push, read the remote head back (`git ls-remote`) and the pull request's merge state; a conflicting pull request runs no CI until the tip is merged in.
6. Take every SHA you write down from command output, never typed from memory.
7. Enter and leave a worktree with the host's built-in worktree tool where it has one; where it has none, name the worktree on every git call as `git -C <worktree> <verb>`, one git command per call (user decision, Issue #2673). The commit guard reads the branch in the folder `-C` names; a line that commits or pushes in two folders is refused, so run each on its own line.

| Host | Enter and leave a worktree | Git inside it |
| --- | --- | --- |
| Claude Code | `EnterWorktree` (`path` for an existing one) and `ExitWorktree`; subagents with `isolation: worktree`. `EnterWorktree` with a `name` branches from the default branch, so for another base create the worktree with `git worktree add` first and enter it by `path`. | Plain `git <verb>`, or `git -C <own worktree> <verb>`. Its built-in guard refuses a line that runs git next to a shell expansion (`$VAR`, `$?`, `$(…)`) or `-C` to the shared checkout; put such work in a script file under `target/`. |
| Codex CLI | `--worktree` or `/worktree` starts the session in a worktree; no documented way out. | Inside a session started there, plain `git <verb>`; otherwise `git -C <worktree> <verb>`. |
| Cursor CLI | `-w` / `--worktree` at start only. | Same as Codex CLI. |
| OMP | None for the main session; the task tool's `isolated` option runs a subagent in its own worktree. | `git -C <worktree> <verb>`. |
| Kiro CLI | None documented. | `git -C <worktree> <verb>`. |

Sources, read 2026-11-13: https://code.claude.com/docs/en/worktrees.md and https://code.claude.com/docs/en/tools-reference.md; https://learn.chatgpt.com/docs/developer-commands?surface=cli; https://cursor.com/docs/cli/reference/parameters; https://github.com/can1357/oh-my-pi; https://kiro.dev/docs/reference/cli-commands/. The Claude Code guard behaviour was measured on 2.1.323 (Issue #2673 comment 6849413590).
