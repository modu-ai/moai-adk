verdict: PASS-WITH-DEBT
audited_sha: 55cac554f56b02bd4ace7e6c29952fa952bb1844
required_backend: codex pass
convergence_overall: pass

# t1858 — merge verdict (Class B, no SPEC)

Card: t1858 (run tn1p4e, lane-31). Branch `WT-release-checks-ssot`, merge base `3642a3a9b`. This file is the SPEC-less card's merge evidence (`.moai/reports/<card-id>/verdict.md`); the audit it cites is `.moai/reports/t1858/sync-audit.md`.

## Claim
`scripts/release.sh` no longer names any required CI check; it reads the `release_gate` list of `.github/required-checks.yml`, so `internal/config` `TestNoHardcodedContexts` passes again, and the release gate demands the same three names with a fail-closed reader.

## Evidence (deciding lines; full output in the cited files)
- Cause and RED: `.moai/reports/t1858/progress.md` § Cause. At the merge base `3642a3a9b`, `go test ./internal/config -run TestNoHardcodedContexts -count=1` FAILED with `scripts file .../scripts/release.sh must not contain hardcoded context "Test (ubuntu-latest)"`.
- Sync audit, three iterations (`.moai/reports/t1858/sync-audit.md`, iteration history inside): iter1 FAIL @f9458ae18 (F1 BLOCKING), iter2 FAIL @16ba65ac9 (F5, F6), iter3 PASS-WITH-DEBT @55cac554f.
- Required backend codex, iteration 3: `audit_multi` receipt `rcpt-2d74fd9b7e1845c52910b486` and `codex_audit` receipt `rcpt-395a7db84a6dbcc53c1ad17a`, both `codex_verdict: pass` (read from `.moai/state/audit-receipts/receipts/`), no findings.
- Card review (advisory): `.moai/reports/t1858/card-review.md`, verdict pass, no findings.
- Lane re-measure at `f85666c0b` (this run): `go test ./internal/config -count=1` ok; `go test ./internal/template -run 'TestReleaseScript|TestTestInputFilters|BranchProtection' -count=1 -v` all PASS; `gofmt -l internal/config internal/template` empty; `golangci-lint run ./internal/config/ ./internal/template/` 0 issues; `scripts/ci-mirror/validate-required-checks.sh` All validations passed; `grep -c 'Test (ubuntu-latest)' scripts/release.sh` 0.

## Baseline-attribution
Audited commit `55cac554f56b02bd4ace7e6c29952fa952bb1844` (tree `.moai/worktrees/t1858`). Differences after it are confined to `.moai/reports/t1858/`. The landing re-measure against the absorbed tip is recorded below when it runs.

## Gaps
- Hosted CI has not run on this change; the new `go_code` filter entry is covered by its guard test, not by an observed classification.
- Audit tool build lag: the `moai` binary behind `audit_multi`/`codex_audit` was built from `ee981a932`, an ancestor of the audited commit.
- Landing re-measure against the current main tip not yet run (see below).

## Residual-risk
- Debt carried by the audit (MINOR, not reachable on the real SSoT): M1 inline value after `release_gate:` not refused; M2 refusal message carries preceding names; M3 a nested item or lone quote is demanded as a name (fails closed).
- The three-name floor lives in Go tests (`internal/config` pin and the real-SSoT script tests), not in the script: changing the demanded set takes a deliberate two-place edit.
