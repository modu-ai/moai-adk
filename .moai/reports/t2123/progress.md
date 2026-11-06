# t2123 progress (#1969 Windows race residuals, split of t2119) — slice ① only

## Claim
Slice ① (`/bin/sh` absent on Windows, 19 failures in the CI analysis) is addressed in the three places the analysis names:
1. `internal/factorymsg/factory_run_retire_test.go` `deadIdentity` and `internal/cli/factory_runs_command_test.go` `deadOwnerIdentity` started `/bin/sh -c "exit 0"` to obtain an exited process. They now run the test binary itself with `-test.run=^$` (`os.Args[0]`), which exits at once on every platform. The helper in `internal/cli` is shared by many tests (including the t2104 rebind tests).
2. `internal/agenticfactory/effect/internal/procrun`: every case drives the runner through a POSIX shell script (`/bin/sh`, `head`, `tr`, `sleep`). The runner is platform neutral; the fixture is not. A `TestMain` skips the package on Windows with the stated reason, instead of failing.
Slices ② (POSIX permission bits, 46), ③ (home guard / TEMP, 54, needs a Windows reproduction) and ④ (pre-commit e2e 12, ~200 "other", classify first) are NOT done here.

## Evidence
- Source: `.moai/reports/pr-sweep-20261106/pr1969-ci-analysis.md` (t2077 tree), row "Windows: `/bin/sh` 없음(19건)": `start short-lived process: exec: "/bin/sh": executable file not found in %PATH%` (cli 11, factorymsg 5, procrun 3).
- `go test ./internal/factorymsg -run 'Retire|Dead' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/factorymsg  29.233s`
- `go test ./internal/cli -run 'TestFactoryRuns' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/cli  24.697s`
- `go test ./internal/agenticfactory/effect/internal/procrun -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/agenticfactory/effect/internal/procrun  0.427s`
- `GOOS=windows go vet ./internal/factorymsg` and `./internal/agenticfactory/effect/internal/procrun` printed nothing; `gofmt -l` on the touched files printed nothing.

## Baseline-attribution
Tree t2123 (branch WT-windows-race-bin), commands run in this run against this tree.

## Gaps
- No RED run: the failure exists only on Windows, which this lane cannot run; the changes are verified to keep the macOS behavior (tests above) and to cross-compile for Windows. The CI Windows race matrix is the verdict.
- The cli "11" sites: only `deadOwnerIdentity` was changed. Other `/bin/sh` references in cli tests (`codex_receipt_detail_test.go`, `factory_operational_*`, `launch_session_pid_exec_posix_test.go`) were not individually checked against the 11; the `_posix` file is build-constrained by name. They need a Windows run to be confirmed or ruled out.
- Slices ②③④ not started; see Residual-risk.
- Focused runs only, no whole suite.

## Residual-risk
- If Windows still reports `/bin/sh` failures in cli after this change, they come from the call sites listed in Gaps.
- Skipping the whole `procrun` package on Windows removes its coverage there; acceptable because the runner has no Windows-specific code path and the fixture cannot run.
- ② needs a per-test survey across ~10 packages (mission 21, cli 11, homestate, factory, graph, yamlpatch, civerdict, cli/update, template, web) before it can be a bounded slice.
