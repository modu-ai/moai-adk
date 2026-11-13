# Internal: git-safety

When to read: before you commit, push, merge or switch branches in a worktree.

1. Write only on an `owner` verdict: run `moai-cli claim-verify` before every commit and every push. Any other verdict stops the write and is reported as it is (MODE-AGENT workflow, `claim-verify` step).
2. Keep the trailer paragraph last: the claim token `Moai-Claim: <issue>#<claim id>` and the attribution line, nothing after them. Confirm each new commit with `git interpret-trailers --parse`.
3. Work in your own worktree with your own build folder. Never write in another worktree and never share a build folder between worktrees.
4. Never force-push, `reset --hard`, `clean`, delete a branch or rewrite history that others may have. Never use a bare `git stash`; the stash stack is shared between worktrees.
5. After every push, read the remote head back (`git ls-remote`) and the pull request's merge state; a conflicting pull request runs no CI until the tip is merged in.
6. Take every SHA you write down from command output, never typed from memory.
