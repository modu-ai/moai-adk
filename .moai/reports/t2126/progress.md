# t2126 progress — PR #1969 Codex P2 (checkpoint GIT_DIR inheritance)

Class B, no SPEC. Base 04d8c27a7e, branch WT-pr-codex-repair-3. AF-Flow: local only, no push.

## Reproduction (RED) — commit e9ad9bd01d
`TestConfirmRemoteIgnoresInheritedGitDir` (internal/agenticfactory/checkpoint/gitremote_env_test.go): two real repos (work, other) each with a bare origin; only other's origin holds WT-x at sha; `GIT_DIR=other/.git` inherited; `ConfirmRemote(GitRemote{Dir: work})`.
Command: `go test -count=1 -run TestConfirmRemoteIgnoresInheritedGitDir ./internal/agenticfactory/checkpoint`
Observed (EXPECTED_RED, assertion failure, exit 1): `ConfirmRemote err = <nil>, want ErrNotAcked (the checkpoint repository's origin lacks the branch)`.
(An earlier run failed to compile — a promoted-field struct literal in my test; that was TOOL_FAILURE, fixed before the RED above.)

## Fix
`GitRemote.git` sets `c.Env = gitenv.Env()` (existing helper that scrubs repo-scoping vars). One production line + import.

## GREEN
- `go test -count=1 ./internal/agenticfactory/checkpoint ./internal/gitenv` -> both `ok`.
- `go vet ./internal/agenticfactory/...` -> exit 0.
- `go test -count=1 ./internal/agenticfactory/...` -> all packages `ok` (runner 73.7s, mergeq 199.9s).

## Gaps
- Mutant probe = the RED run itself (fix absent); no separate mutant of the fix line.
- The worktree guard refused one compound shell command (add+commit chain); re-issued as separate plain commands.
- Independent audit: sync-audit-iter1.md PASS 0.93 @8ecd4bfcc0 (published as verdict.md, commit 8795490c5b); stage chain run->sync->sync-audit->merge-ready v11 (one rebind-evidence to the audited commit).
- card-review.md not produced: codex_audit 1회 pass·지적 0(rcpt-ee212793db05707dcbc2a1da)이 리뷰를 대신함, Codex 한도 보존 (leader instruction).
- Follow-up candidate (not fixed): ctxmodel/ctxmodel.go:310 GitVerifier.Lookup runs git cat-file without Env (read only, not reproduced); leader asks the operator whether to issue a card.
