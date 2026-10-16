# t1823 sync-audit verdict

Written by lane-38 as the lane's completion verdict on mechanical evidence (`factory-dispatch-detail.md` § The verdict's home). The independent evidence beneath it: the required-backend codex receipt, the test-executing `sync-auditor` run, and the lane's own re-measurements. Details: `.moai/reports/t1823/card-review.md`, `.moai/reports/t1823/auditor-run-report.md`, `.moai/reports/t1823/progress.md`.

verdict: PASS
audited_sha: 7c76f7147e3b44fb5d8b89114f9ee74da063ac56
convergence_overall: pass
required_backend: codex pass

## Claim

Card t1823's two red tests are repaired: `TestAuditLagUsesBinlagSeam` (coordinate drift from t1619/t1589) and `TestManagedTUINeverReachedWithoutOptIn/opted_in_probes_once` (production-path regression from `063a7620f`, repaired in the managed divert). The production change is confined to `internal/cli/codex_launcher.go`.

## Evidence

- Required backend codex: `codex_audit` native / baseBranch → `verdict: pass`, no findings, receipt `rcpt-c285169dc42467aba2878c41` (HEAD `1a82b5991`; the audited commits are unchanged since, the later commit adds evidence files only).
- Card-scope `codex_review` → pass, no findings (advisory).
- `sync-auditor` spawn, test-executing: C1–C5 PASS, no blocking findings, two low optional findings dispositioned in card-review.md.
- Lane re-measurements at the final state: four target tests PASS (`ok 2.759s`); `golangci-lint run --new-from-rev=ec5a4f392 --timeout=10m ./internal/cli/...` → `0 issues.`; `go vet` and `gofmt -l` no output.
- A/B of the introducing commit: `063a7620f^` PASS, base FAIL (progress.md §4).

## Authorship disclosure

`54944481c` is a LANE direct edit (test coordinate map only); `f37be81f7` is the `manager-develop` spawn's work. The leader ruled the first need not be redone. Both are inside the audited range.

## Baseline-attribution

Tests and lint ran in `.moai/worktrees/t1823`; the audit tools reported their judging build as `3a8689d6b`, a strict ancestor of HEAD, and the diff since touches only the paths listed in card-review.md.

## Gaps

- The full `./internal/cli` package was not run; the t1823 landing remeasure (and its recorded failing-test count) is still ahead and gated by the go-test-heavy slot.
- Managed divert not run against a real codex binary; RED of the new test not re-observed by the lane.
- codex judged statically. The sync-auditor's second receipt cited in its final message was not verified and is not relied on.

## Residual-risk

- `./internal/cli` is not shown green by this verdict; carried failures named in other cards' skips still apply.
- The coordinate want map is brittle and will rot again on the next edit above its lines.
