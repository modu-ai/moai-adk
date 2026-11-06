---
card: t2128
reviewed_sha: c5bc16ab2a9711f4e10abe98b8d3ed48173a661e
reviewer: lane-28 self-review (Claude)
codex_review: not run — the stage codex budget is used (rule d-20261106T080039Z-305c)
---

# t2128 card review

## Why this is a self-review

Standing rule `d-20261106T080039Z-305c` allows at most 2 codex calls per audit stage, and the auditor makes them. The sync stage used both:
- iter-1 delta from 04d8c27a7e: pass, receipt `rcpt-e703dc932cb9971ce8459362`
- iter-2 delta from 4e7aed6246: pass, receipt `rcpt-eb59e780ddb40a6c320d00c0`

So the lane did not call `codex_review` here, and those two receipts are this card's codex evidence. This file is advisory. The binding verdict is `sync-audit-iter2.md`, summarized in `verdict.md`.

## Diff reviewed (04d8c27a7e..c5bc16ab2a)

**`internal/config/agentic_factory_settings.go`**
- `afsHasSecretKey` now returns `(found, cycle)`.
- It tracks nodes on the current path to detect a cycle, and nodes already walked clean so a shared anchor is walked only once.
- `guardSecrets` reports a cycle on `webhook.routes[i]` with no value echoed.
- The unused `slices` import is removed.

**`internal/config/agentic_factory_settings_cycle_test.go`** (new)
- A subprocess regression test: re-exec, env guard, 60s bound.
- A shared-anchor test that must not be rejected, and a secret-through-anchor test that must still be refused.
- A 7-case direct unit test.
- The child writes check their errors (`_, _ =`), fixed in `fa2a30ca6e`.

## Lane observations

**Commit order.** The RED test came before the fix: `d5e205208c` precedes `15d6a8648a`. The reproduction in progress.md §2.1 is `fatal error: stack overflow` on the unfixed code.

**Lane-run results.**
- `go test -count=1 ./internal/config/` gave `ok … 22.910s`.
- `GOOS=windows GOARCH=amd64 go build ./...` gave exit 0.

**Lint.** `golangci-lint run ./internal/config/` with local v2.15.0 reports 7 `govet inline` findings, all in files this card does not touch. The auditor found the same 7 on the base export. CI-pinned v2.1.6, built with Go 1.26.8, reports `0 issues.` on HEAD (auditor run).

**Security.** A cyclic alias now fails as a config error instead of crashing the process, and the error text never echoes a value. Expansion-heavy input (10^12) is walked in linear time (auditor measurement).

## Debt carried (non-blocking, from the audits)

- **F2:** a `secret` key behind an alias key or a complex key goes undetected. This predates the card.
- **F3:** when a route holds both a cycle and a secret, the cycle is reported first.
- **F4:** a cycle outside `webhook.routes` is not read by this reader. The resolver's map decode reports it cleanly.
- **F5:** the 7 pre-existing govet `inline` findings, which appear on local v2.15.0 only.
