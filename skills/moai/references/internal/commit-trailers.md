# Internal: commit-trailers

When to read: before you write a commit message for a change that belongs to a SPEC, and before you post an audit verdict.

Every commit an agent makes ends with two identity trailers in its last paragraph (user decision 2026-11-17). A card number never goes into a commit, a branch name or a pull request: the SPEC id names the work.

```
Moai-Spec: SPEC-<yyyymmdd>-<slug>
Moai-Session: <this session's id>:<worktree key>
```

The `Moai-*` lines and any `Co-Authored-By` line are one last paragraph with no blank line between them. Git reads trailers from the last paragraph only, so a blank line before `Co-Authored-By` leaves the `Moai-*` lines unread: `moai trailer-check` then reports `missing_spec`, and the commit-msg hook refuses the message as `split_trailer_block`.

```
fix(scope): subject

Body.

Moai-Spec: SPEC-<yyyymmdd>-<slug>
Moai-Session: <this session's id>:<worktree key>
Moai-Verified: tests
Co-Authored-By: <name> <email>
```

1. `moai trailers --spec <SPEC-id>`, run in the worktree you commit from, prints both lines in `text`; put them in the trailer block, next to any `Co-Authored-By` line. Never type the value yourself.
2. The session id is the one the host gives this session (`CLAUDE_CODE_SESSION_ID`; `moai session-current` reads the same): not a secret, 8 to 64 hex digits and `-`. It is not stable for a whole stretch of work: Claude Code gives a new id after `/clear` (its environment-variable reference says the variable is "updated on `/clear`"), so one author's commits may carry several values. Comparing lane names stays the fallback. When the host gives none, or the folder is in no git worktree, `moai trailers` says so (`session_missing`): commit with `Moai-Spec` only and say in the report why the session line is missing. That change's authors are then unknown to the session that records its verdict, which says so in the verdict note.
3. Why the worktree key (leader decision 2026-11-17): a Claude Code subagent inherits its parent's session id, so the id alone cannot tell an author from the audit agent it started. The worktree key is the first 12 hex digits of the sha256 of the worktree's canonical git directory (`git rev-parse --absolute-git-dir`): each worktree of a repository has its own, and one worktree always gets the same. So an audit agent runs in a worktree of its own, never in the author's.
4. One SPEC per commit. A commit that belongs to no SPEC does not go on a `spec/` branch.
5. An audit verdict names its auditor the same way: the line below the verdict's first line is `Moai-Session: <value>`, the value `moai trailers` prints in the auditor's own worktree (`workflows/audit.md`, "Verdict note"). No command reads that line on a card: the session that records the verdict compares it with every `Moai-Session` of the change's commits, and a match is an author, never an independent PASS. A bare session id from before the worktree key counts its whole host session as one author. Both values are self-declared: an author can write any `Moai-Session` value, so the comparison catches a slip, not a deliberate lie, which the independent audit and the lander's review still have to.
6. Before asking for an audit, run `moai trailer-check <base>..HEAD` (`--branch <name>` when the checkout is detached, as in CI): on a `spec/` branch a commit without `Moai-Spec`, and anywhere a malformed value, fails it; a missing `Moai-Session` and a bare legacy id are reported only. It exits 1 when anything fails, with the findings still printed.
