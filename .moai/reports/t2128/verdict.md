---
card: t2128
verdict: PASS
audited_sha: c5bc16ab2a9711f4e10abe98b8d3ed48173a661e
score: 0.93
source: .moai/reports/t2128/sync-audit-iter2.md
---

# t2128 verdict

| Item | Value |
|---|---|
| Binding audit | `sync-audit-iter2.md`, PASS 0.93, independent Claude sync-auditor (opus) |
| Audited commit | `c5bc16ab2a9711f4e10abe98b8d3ed48173a661e` |
| Iteration history | iter-1 FAIL 0.87 @4e7aed6246, one blocking finding (F1, errcheck at cycle_test.go:51/:54); repaired in `fa2a30ca6e` |
| Codex receipts | `rcpt-e703dc932cb9971ce8459362` (iter-1), `rcpt-eb59e780ddb40a6c320d00c0` (iter-2); both pass, 0 findings; 2 calls, within the rule `d-20261106T080039Z-305c` limit |
| Card review | `card-review.md` (lane self-review; codex budget used) |

## Deciding evidence (verbatim lines carried from the audits and the lane runs)

- RED on the unfixed code: `runtime: goroutine stack exceeds 1000000000-byte limit` / `fatal error: stack overflow` (progress.md §2.1; the auditor re-ran it on a scratch extract of `d5e205208c`).
- After the fix: `go test -count=1 ./internal/config/` → `ok  github.com/modu-ai/moai-adk/internal/config	11.580s` (auditor, iter-2) and `22.910s` (lane).
- `GOOS=windows GOARCH=amd64 go vet ./internal/config/` → exit 0 (auditor). `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild exit=0` (lane).
- CI-equivalent lint (golangci-lint v2.1.6, Go 1.26.8, auditor) → `0 issues.` on HEAD, and errcheck 2 on the iter-1 commit.

## Gaps

- No Windows execution of the subprocess test; Windows vet only.
- CI was not run. CI lint conditions were reproduced locally.
- The moai binary serving the audit tools was built from `04d8c27a7e`, an ancestor of HEAD; the card changes no moai CLI behaviour.

## Residual risk

The F2–F4 debt is listed in `card-review.md`. F2 means a `secret` key written as an alias key or a complex key is not refused; this predates the card.
