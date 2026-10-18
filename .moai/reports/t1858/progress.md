# t1858 — local main red: TestNoHardcodedContexts (Class B, run)

Card: t1858 (run tn1p4e, lane-31, P0). Tree `.moai/worktrees/t1858`, branch `WT-10-18-class` (to be renamed), base `3642a3a9b`.

## Cause (established before leaving run)

### Claim
`scripts/release.sh` carries the forbidden literal `Test (ubuntu-latest)` outside `.github/required-checks.yml`, which `internal/config.TestNoHardcodedContexts` forbids anywhere under `scripts/` (non-test). The literal landed with t1833 (`4e72e171f`, `315fc6dc9`); t1833's landing re-measure ran `internal/template` only and so never ran the `internal/config` scan that reads this non-Go file.

### Evidence (this run, this tree `3642a3a9b`)
```
go test ./internal/config -run TestNoHardcodedContexts -count=1
--- FAIL: TestNoHardcodedContexts (0.01s)
    required_checks_test.go:121: scripts file .../scripts/release.sh must not contain hardcoded context "Test (ubuntu-latest)"
FAIL	github.com/modu-ai/moai-adk/internal/config	0.400s
```
Offending lines (`grep -n` on `scripts/release.sh`): `:240` (comment naming the job), `:244` `CI_REQUIRED_CHECKS=("Lint" "Test (ubuntu-latest)" "Build (linux/amd64)")`. The same RED was observed by the leader on `.claude/worktrees/main` @`3642a3a9b`.

The test (`internal/config/required_checks_test.go:85-128`) forbids exactly one string, `Test (ubuntu-latest)`, in `internal/cli/branch_protection.go`, `internal/cli/github_init.go`, and every non-`/test/`, non-`_test.sh` file under `scripts/`.

### `:323` `Release Verify (…)` question
Not subject to this rule: those three names are job names of the release-verify matrix workflow, are not contexts in `.github/required-checks.yml`, and do not contain the one forbidden string. Left unchanged (scope discipline).

### Constraints that shape the repair
- The gate tests (`internal/template/release_ci_gate_test.go`) pin the gate to exactly three names (`Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`) with 3-name fixtures; they run `scripts/release.sh` from the real repo inside a scratch clone that holds only `CHANGELOG.md` and `README.md` (`release_script_test.go:33`). So the SSoT must be located relative to the script file, not the working-directory repo.
- `branches.main.contexts` cannot be read whole: it also lists `Release PR Multi-OS Gate` (PR-only), `Analyze (Go) (go)` and four more Build legs, none of which a push-to-main gate may demand without changing release behavior. `branches."ci/**"` (7 names) would make the gate stricter, also a behavior change.
- `yq` is not guaranteed on every machine that runs the gate tests; the SSoT is a repo-controlled file in one fixed shape.

## Repair design
Add a top-level `release_gate:` list to `.github/required-checks.yml` (the SSoT) holding the three names; mirror it as `ReleaseGate []string` in `internal/config.RequiredChecks`; `release.sh` reads that list (dependency-free `awk`, fail-closed when none parse) and no longer names any check literally, including in comments. A new `internal/config` test asserts `release_gate` is non-empty and a subset of `branches.main.contexts`, so it cannot drift from the required set. Behavior of the gate is unchanged (same three names).

## Gaps (so far)
- No fix applied yet; the repair, its tests, and the re-measure are recorded below as they are produced.
- t1853 (lane-26): `git -C ../t1853 diff --quiet 3642a3a9b -- scripts/release.sh .github/required-checks.yml` reported no difference at the time of writing, so there is no same-line overlap *now*; whether t1853 edits these files later is unobserved.

## Run evidence (run phase, cycle_type=tdd; base tree be83fea45)

- RED (before struct value/YAML block): `go test ./internal/config -run TestRequiredChecks_ReleaseGateSubsetOfMain -count=1`
  -> `--- FAIL ... required_checks_release_gate_test.go:17: release_gate must be a non-empty list in .github/required-checks.yml` / `FAIL github.com/modu-ai/moai-adk/internal/config 0.298s`
- GREEN: same command -> `--- PASS: TestRequiredChecks_ReleaseGateSubsetOfMain (0.00s)` / `ok ... internal/config 0.122s`
- `go test ./internal/config -count=1` -> `ok github.com/modu-ai/moai-adk/internal/config 13.534s` (includes TestNoHardcodedContexts, now PASS)
- `bash -n scripts/release.sh` -> exit 0; `grep -c 'Test (ubuntu-latest)' scripts/release.sh` -> 0; no `Build (linux` / `"Lint"` literal left either
- awk read in isolation against the real SSoT -> `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`
- `go test ./internal/template -run 'TestReleaseScript|TestTestInputFilters' -count=1 -v` -> 7 TestReleaseScript* and 2 TestTestInputFilters* all PASS; `ok ... internal/template 29.735s`
- `./scripts/ci-mirror/validate-required-checks.sh` (the other SSoT consumer) -> exit 0, all validations passed
- `gofmt -l internal/config` -> empty; `go vet ./internal/config` -> exit 0

Gaps: the script's own fail-closed `die` branches (file missing, zero names) were not driven end to end; only the awk yielding 0 lines on a block-less file was probed. The internal/cli tests that mention release.sh were not run (they reference the script for unrelated reasons). Full repository suite not run (scoped by instruction).

## Fail-closed evidence (follow-up, test internal/template/release_ci_gate_ssot_test.go)

`TestReleaseScriptCIGateSSoTFailsClosed` runs a COPY of the real script from `<tmp>/scripts/release.sh` with `<tmp>/.github/required-checks.yml` set per case (`--dry-run`, stub gh): SSoT absent, no release_gate block, empty release_gate block (all exit 1 with `[FAIL] CI gate: ...` and no check-runs call), and a control with a copy of the real SSoT (exit 0, `CI gate:     verified`).

Sensitivity: first draft matched on message text only and PASSED with both `die` guards neutered to `log_warn` (the text still printed, exit 1 came from a later failure) - a wrong-reason pass. Fixed by requiring the `[FAIL]` prefix and no `check-runs` gh call. Re-run with both guards neutered in the real script (temporary, restored with `git checkout -- scripts/release.sh`): the three refusal rows FAIL ("a refused gate must not read the check runs; gh calls: [api --paginate ... check-runs ...]"), control PASS. Unmutated: all four PASS; `git diff -- scripts/release.sh` empty.

## Sync evidence

Files changed in sync (docs only; no spec/scripts/internal/.github/test/rules touched):
- `CHANGELOG.md` - one entry under `## [Unreleased]` / `### Fixed` (card t1858).
- `.moai/docs/batch-failure-bisect-runbook.md` - three statements that named `CI_REQUIRED_CHECKS` as the home of the required check names now name the `release_gate` list of `.github/required-checks.yml` (a rename is a one-place edit there; missing/empty file makes the script refuse).

Commands run: `grep -n '^### Fixed' CHANGELOG.md`; `grep -n 'CI_REQUIRED_CHECKS\|release_gate' .moai/docs/batch-failure-bisect-runbook.md`; `git diff --stat`; `ls .github/required-checks.yml internal/config/required_checks_release_gate_test.go`. No markdown or docs lint exists in the Makefile (grep for markdownlint/docs-check/lint-md returned nothing), so none was run.

## Sync audit round 1: FAIL, repair spawn blocked (lane-31)

Verdict: `.moai/reports/t1858/sync-audit.md` = FAIL @`f9458ae18`; required backend codex `fail`, receipts `rcpt-b6a0f22a9eee365cd060f299` (audit_multi) and `rcpt-b2c03755b996d645bf1ea108` (standalone codex). Defects: F1 BLOCKING (the gate only checks "non-empty": a `release_gate` shortened to `- Lint` lets a tag through on a lone green Lint; reproduced by the auditor's awk probes and by codex end to end), F2 SHOULD-FIX (CRLF, trailing comment, single quote misparse), F3 MINOR (stale `CI_REQUIRED_CHECKS` comment and die text), F4 MINOR (subset test too weak).

Repair plan (the lane's judgment, ladder terminal step): pin the demanded set in two CI-enforced layers without a check literal in `scripts/`: (a) `internal/config` test asserting `release_gate` equals exactly the three names, no duplicates; (b) `internal/template` tests driving the REAL script against the REAL SSoT with one name missing at a time and a Lint-only page; plus strict awk (reject CRLF residue, trailing comment, single quote, duplicate name/key) and fixed wording.

Blocked, verbatim deny from the real `Agent(manager-develop)` spawn attempt:
```
PreToolUse:Agent hook error: AUDIT_RECEIPT_VIOLATION: manager-develop cannot be spawned while an audit PASS stands unproven in /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858. Outstanding: sync-auditor / unknown-spec / verdict line missing. Re-run the audit through codex_audit or audit_multi and let the auditor end with a verdict line citing the receipt.
```
Facts read from the tree (read-only, nothing deleted or edited): `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json` = agent_id `a39f63408f3184be8`, cause `verdict line missing`, rejected_at `2026-10-18T01:13:24.954843Z`, reentry_warned true; `starts/` holds 0 files, `rejections/` holds 1 file. The same auditor agent_id ended with a FAIL verdict line (`AUDIT-VERDICT: FAIL spec=t1858 receipts=…`), committed with this note. The lane does not clear the record, does not substitute another agent type, and does not edit code itself; it asks the leader to clear the record under standing c305, after which the repair spawn runs.

## Sync-audit repair r1 (run phase, cycle_type=tdd; base f9458ae18)

Defects from `.moai/reports/t1858/sync-audit.md`: F1 BLOCKING (gate only demanded a non-empty list), F2 SHOULD-FIX (awk misparsed malformed lists), F3 MINOR (stale `CI_REQUIRED_CHECKS` wording), F4 MINOR (weak subset test).

- F1a/F4: `internal/config/required_checks_release_gate_test.go` now pins `release_gate` to exactly `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)` with no duplicates (CI-enforced floor). RED (SSoT temporarily shortened to `- Lint`, restored with a file copy): `release_gate lacks the pinned check "Test (ubuntu-latest)": [Lint]` / `... "Build (linux/amd64)"` / `release_gate must hold exactly [...], got [Lint]`; GREEN on the real SSoT.
- F1b: `TestReleaseScriptRealSSoTDemandsEveryPinnedCheck` (real script, real SSoT): Lint missing, Test missing, Build missing, Lint-only green page, each exit 1 with the exact missing names. RED with the SSoT shortened to `- Lint`: test_missing, build_missing, only_lint_green exit 0 (want 1); SSoT restored.
- F2: awk strips `\r`, refuses trailing ` #` comment, single-quoted entry, duplicate name, duplicate `release_gate:` key with `[FAIL] CI gate: malformed release_gate in <file>: <reason>`. Five new rows in `TestReleaseScriptCIGateSSoTFailsClosed` (RED before the script change; CRLF row asserts a CRLF copy of the real SSoT still passes). Sensitivity on mutated copies of the script (restored from a backup, `cmp` identical): removing any one guard fails exactly its own row.
- F3: comment and die text now point at the `release_gate` list of `.github/required-checks.yml`; no test asserted the old phrase. CHANGELOG sentence extended to mention malformed lists and the pin.
- Not touched: `Release Verify (...)` legs, `TestNoHardcodedContexts`, `required_checks_test.go`.
- Limit: the script cannot know the floor without naming the checks, so the floor is the CI-required Go test plus the real-SSoT script tests, not the script.

## Sync-audit repair r2 (run phase, cycle_type=tdd; base 16ba65ac9)

Defects from iteration 2 of `.moai/reports/t1858/sync-audit.md`: F5 SHOULD-FIX (`.github/required-checks.yml` absent from the `go_code` filter, so an SSoT-only PR gets the skip-marker as `Test (ubuntu-latest)` and the pin tests never run), F6 SHOULD-FIX (empty `release_gate` entries silently dropped).

- F5: `'.github/required-checks.yml'` added to `go_code` in `.github/test-input-filters.yml` with a comment naming the readers; the same path added to `requiredGoCodeRoots` in `internal/template/test_input_filter_parity_test.go`. RED (entry in the guard list, absent from the filter): `go_code list lost required root ".github/required-checks.yml"` (`TestTestInputFiltersCoversGoTestInputs` FAIL). GREEN after the filter edit. Sensitivity: filter entry removed again (backup, `cmp` restored) -> same FAIL.
- F6: awk in `scripts/release.sh` now refuses an empty entry (`-`, `- `, `- ""`, `- "  "`), a space-padded quoted name, and any non-comment, non-entry line inside the block (`-Lint`), each as `[FAIL] CI gate: malformed release_gate in <file>: <reason>`; the `[[ -n "$check_name" ]] &&` skip is gone (an empty awk result dies before the loop, so the loop cannot add an empty element). `- name`, `- "name"`, tab indentation, CRLF, comment and blank lines inside the block still pass. Seven rows added to `TestReleaseScriptCIGateSSoTFailsClosed` (six refusals plus a tab-indent pass). RED before the script change: the six refusal rows FAIL. Sensitivity (script mutated one guard at a time, backup restored, `cmp` identical): empty-entry guard off -> 4 rows FAIL; space-padded off -> 1 row; unrecognized-line off -> 1 row.
- CHANGELOG sentence extended to list the new refusals. Not touched: `Release Verify (...)` legs, `TestNoHardcodedContexts`, `required_checks_test.go`.
- Not addressed (auditor classed unrealistic): a `---` document marker mid-list still ends the block.
