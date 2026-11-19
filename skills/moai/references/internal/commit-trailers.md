# Internal: commit-trailers

When to read: before you write a commit message for a change that belongs to a SPEC, and before you post an audit verdict.

Every commit an agent makes ends with two identity trailers in its last paragraph. A card number is personal and never goes into a commit, a branch name or a pull request.

```
Moai-Spec: SPEC-<yyyymmdd>-<slug>
Moai-Session: <this session's id>:<worktree key>
```

1. `moai trailers --spec <SPEC-id>`, run in the worktree you commit from, prints both lines in `text`; put them in the trailer block, next to any `Co-Authored-By` line. Never type the value yourself.
2. The session id is the one the host gives this session (`CLAUDE_CODE_SESSION_ID`, then `CODEX_THREAD_ID`; `moai session-current` reads the same): not a secret, 8 to 64 hex digits and `-`. It is not stable for a whole stretch of work: Claude Code gives a new id after `/clear` (its environment-variable reference says the variable is "updated on `/clear`"), so one author's commits may carry several values. Comparing lane names stays the fallback. When the host gives none, or the folder is in no git worktree, `moai trailers` says so (`session_missing`): commit with `Moai-Spec` only and say in the report why the session line is missing. The audit gate then reads that pull request's authors as unknown, and a tier 1 or 2 pull request waits.
3. Why the worktree key: a Claude Code subagent inherits its parent's session id, so the id alone cannot tell an author from the audit agent it started. The worktree key is the first 12 hex digits of the sha256 of the worktree's canonical git directory (`git rev-parse --absolute-git-dir`): each worktree of a repository has its own, and one worktree always gets the same. So an audit agent runs in a worktree of its own, never in the author's.
4. One SPEC per commit. A commit that belongs to no SPEC does not go on a `spec/` branch.
5. An audit verdict names its auditor the same way: a `Moai-Session: <value>` line below the verdict's first line, the value `moai trailers` prints in the auditor's own worktree. The audit gate compares it, and the auditor's lane name, with every `Moai-Session` of the pull request's commits; a match is an author, never an independent PASS. A bare session id from before the worktree key (one transition, until the open pull requests that carry it have merged) still counts its whole host session as one author. Both values are self-declared: an author can write any `Moai-Session` value, and the gate trusts it the way it trusts lane names; it catches a slip, not a deliberate lie, which the independent audit and the leader's review still have to.
6. Before asking for an audit, run `moai trailer-check <base>..HEAD` (`--branch <name>` when the checkout is detached, as in CI): on a `spec/` branch a commit without `Moai-Spec`, and anywhere a malformed value, fails it; a missing `Moai-Session` and a bare legacy id are reported only. It exits 1 when anything fails, with the findings still printed.
