verdict: PASS
audited_sha: e7ec3140c8b1f6d017799043360c14991caae8e9

# t2133 card review, iteration 2

Reviewers: sync-auditor (claude-opus-5-5, in-session) and one approved Codex audit (`mcp__moai__codex_audit`, native, baseBranch, project_root = this worktree). Commit reviewed: e7ec3140c8b1f6d017799043360c14991caae8e9 (diff from 04d8c27a7e, 3 files).

Codex receipt: `rcpt-1975ef14d9c8638ef237d672`. Codex verdict pass, findings none, model gpt-6.1-sol, review base main (merge base 04d8c27a7e). Approval: d-20261106T163140Z-56c1. No audit_multi, no GLM, no reused receipt.

## Reviewed
- internal/testguard/homeguard.go: adds `CheckSeamHomesInTemp` / `checkSeamHomesInTemp`. `judgeSandbox` now delegates to `judgeSandboxIn` with an injectable temp dir and makes the same decision as before (re-read in this run).
- internal/testguard/homeguard_test.go: adds `TestCheckSeamHomesInTemp` (4 subtests on a Windows-like injected layout). `TestHomeGuard_AcceptsTempDir` now calls the exempting check.
- internal/cli/init_home_guard_test.go: `prepareSafeInitHome` now calls the exempting check.

## Measured in this run
- `go test -count=1 -run 'HomeGuard|CheckSeam|Sandbox' ./internal/testguard/` exits 1. Every focused test passes. The only failure is TestHomeGuard_RepositoryHasNoNewHomeAccess (F3).
- `GOOS=windows go vet ./internal/testguard/ ./internal/cli/` exits 0, and gofmt prints nothing.

## Findings
- F1 (optional, low, in-diff): the exemption covers all of os.TempDir rather than the test's own root. The real home is not exposed.
- F2 (optional, trivial, in-diff): comment typo "positive case:t.TempDir()" at homeguard_test.go:331.
- F3 (outside, info): the repo-scan test fails on two `~/` literals in files this diff does not touch. The scanner is untouched too, and iteration 1 reproduced the failure at base. The leader should route it separately.

Verdict PASS, score 0.90. The required Codex gate is met. Full evidence: `.moai/reports/t2133/verdict.md`. Iteration 1: `.moai/reports/t2133/sync-audit-verdict-iter1.md`.
