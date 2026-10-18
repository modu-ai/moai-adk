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
