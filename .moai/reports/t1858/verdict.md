verdict: PASS-WITH-DEBT
audited_sha: 90c92ea98dde923ddfbebc9cff96f54a17931e3e
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

- HEAD confirmation after absorbing main (`.moai/reports/t1858/sync-audit-head-confirmation.md`): `verdict: PASS` @`90c92ea98dde923ddfbebc9cff96f54a17931e3e` (merge of main `09f4959a5` into the card branch; one CHANGELOG.md conflict resolved by keeping both entries). The card's code files are byte-identical to the iteration-3 audited versions; `git diff --stat main HEAD` lists only the card's own files. `audit_multi` receipts `rcpt-dddc850b167379461c7d7aa8` and `rcpt-2141289dadbbc5ba3c1bfeee`, both `codex_verdict: pass`, no findings. The label here stays PASS-WITH-DEBT because the iteration-3 debts M1-M3 are carried unchanged.

## Baseline-attribution
Audited commit `90c92ea98dde923ddfbebc9cff96f54a17931e3e` (tree `.moai/worktrees/t1858`, the merged HEAD). Differences after it are confined to `.moai/reports/t1858/`. The sync audit proper (iteration 3) read `55cac554f56b02bd4ace7e6c29952fa952bb1844`; the confirmation above binds the merge. The landing re-measure runs against the committed tree of this HEAD and is keyed to it, so its command and result are not appended here (that would change the tree); they are in the lane's landing report to the leader and in the re-measure verb's own record.

## Gaps
- Hosted CI has not run on this change; the new `go_code` filter entry is covered by its guard test, not by an observed classification.
- Audit tool build lag: the `moai` binary behind `audit_multi`/`codex_audit` was built from `ee981a932`, an ancestor of the audited commit.
- The landing re-measure against the absorbed tip had not run when this file was written (it runs after this file is committed; see Baseline-attribution).

## Residual-risk
- Debt carried by the audit (MINOR, not reachable on the real SSoT): M1 inline value after `release_gate:` not refused; M2 refusal message carries preceding names; M3 a nested item or lone quote is demanded as a name (fails closed).
- The three-name floor lives in Go tests (`internal/config` pin and the real-SSoT script tests), not in the script: changing the demanded set takes a deliberate two-place edit.
