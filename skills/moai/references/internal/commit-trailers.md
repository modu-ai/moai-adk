# Internal: commit-trailers

When to read: before you write a commit message for a change that belongs to a SPEC, and before you post an audit verdict.

Every commit an agent makes ends with two identity trailers in its last paragraph (user decision U-B, 2026-11-17). They replace the planned `Moai-Claim: C-<n>#<gen>`: a card number is personal and never goes into a commit, a branch name or a pull request.

```
Moai-Spec: SPEC-<yyyymmdd>-<slug>
Moai-Session: <this session's id>
```

1. `moai trailers --spec <SPEC-id>` prints both lines in `text`; put them in the trailer block, next to any `Co-Authored-By` line. Never type the session id yourself.
2. The session id is the one the host gives this session (`CLAUDE_CODE_SESSION_ID`, then `CODEX_THREAD_ID`; `moai session-current` reads the same): stable for the host session, not a secret, 8 to 64 hex digits and `-`. When the host gives none, `moai trailers` says so (`session_missing`): commit with `Moai-Spec` only and say in the report that the session id was missing. The audit gate then reads that pull request's authors as unknown, and a tier 1 or 2 pull request waits.
3. One SPEC per commit. A commit that belongs to no SPEC does not go on a `spec/` branch.
4. An audit verdict names its auditor's session the same way: a `Moai-Session: <auditor session id>` line below the verdict's first line. The audit gate compares it, and the auditor's lane name, with every `Moai-Session` of the pull request's commits; a match is an author, never an independent PASS.
5. Before asking for an audit, run `moai trailer-check <base>..HEAD`: on a `spec/` branch a commit without `Moai-Spec`, and anywhere a malformed value, fails it; a missing `Moai-Session` is reported only.
