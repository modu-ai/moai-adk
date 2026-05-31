# Implementation Plan — SPEC-V3R6-CI-FLAKY-STABILIZE-001

## §A — Context

Two CI-only flaky test failures keep `origin/main` red. Both are evidence-diagnosed (file + line citations in §B Known Issues). This plan sequences the fixes lowest-risk-and-most-verifiable first (FLAKY-1, deterministically reproducible locally) before the Windows fixes (FLAKY-2, verifiable only on the CI Windows runner).

Development mode: per `.moai/config/sections/quality.yaml`. Cycle type chosen by orchestrator at run-phase (`tdd` recommended — each fix has a clear before/after test assertion).

Git: Hybrid Trunk, Tier M → main-direct (no PR). Note: current branch carries a local-ahead commit `3dcd58677` from a parallel session (disjoint scope) — do NOT touch it.

## §B — Known Issues (evidence-bound)

### FLAKY-1
- `internal/spec/closer.go` `Close()` — `baseDir` defaults to `"."` (relative) at lines ~126-128.
- `internal/spec/closer.go` `performAtomicClose()` (~line 259) — `git add` / `git restore --staged` receive `state.SpecMDPath` / `state.ProgressMDPath` as ABSOLUTE paths while `cmd.Dir = baseDir` (`runGitInDir`, ~line 372). Absolute-path-vs-`cmd.Dir` resolution is the race surface under the CI race-detector.
- `internal/spec/closer_test.go` — 15 `t.Parallel()` siblings; `TestClose_FullClose_ProducesCommit` (~line 900) is intentionally NON-parallel (atomic transaction; comment at line 901-902).
- Local state: `go test -race ./internal/spec/ -run TestClose_FullClose_ProducesCommit -count=5` → PASS (race is CI-only).

### FLAKY-2
- `internal/merge/confirm.go` `ConfirmMerge()` (~line 838) — calls `tea.NewProgram(m).Run()` at ~line 854-855 with no TTY guard. Blocks on Windows `ReadConsole` in non-TTY CI.
- `internal/cli/update.go` — `runTemplateSyncWithProgress()` (~line 897) calls `merge.ConfirmMerge(analysis)` at ~line 928 ONLY when `!autoConfirm` (the `--yes` flag, line 901/917). A second call site at line 592 (`autoConfirm` branch).
- `internal/cli/target_coverage_test.go` (~line 266-268) — `TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck` ALREADY has `if runtime.GOOS == "windows" { t.Skip("charmbracelet/bubbletea blocks on Windows console ReadConsole in non-TTY CI") }`. This is the proven skip idiom + precedent.
- `internal/cli/coverage_improvement_test.go` — most `runTemplateSyncWithProgress` callers set `--yes=true` (e.g. `VersionMismatch` line 4308, `TemplatesOnlySkipsBinary` line 6311), so they take the skip branch and do NOT reach `ConfirmMerge`. The hanging caller is whichever reaches `ConfirmMerge` with `autoConfirm=false` — M2 audit determines the exact set.
- Dependency: `github.com/mattn/go-isatty v0.0.22` is ALREADY in `go.mod` and used at `update.go:151`, `init.go:279/315` (`isatty.IsTerminal(os.Stdin.Fd())`). No new dependency required.

## §C — Pre-flight

- [ ] `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` — surface divergence before any commit (per agent-common-protocol §Pre-Spawn Sync Check).
- [ ] Confirm `mattn/go-isatty` import idiom by reading `internal/cli/update.go:151`.
- [ ] Confirm `TestClose_FullClose_ProducesCommit` non-parallel comment intact before editing closer.go.

## §D — Constraints

- [HARD] Preserve `TestClose_FullClose_ProducesCommit` as NON-parallel (atomicity invariant).
- [HARD] Do NOT introduce a new dependency for the TTY guard — reuse `mattn/go-isatty`.
- [HARD] Do NOT touch `docs-site/`, `internal/template/templates/`, or mirror-drift fixtures (sibling SPEC scope).
- [HARD] FLAKY-2 Windows behavior is verified by the CI Windows job, not local darwin runs.
- After fixing any test, run the FULL `go test ./...` to catch cascading failures (CLAUDE.local.md §6).

## §E — Self-Verification (run-phase)

Parallel read-only batch on completion:
1. `go test -race ./internal/spec/ -count=20`
2. `go test -race ./internal/spec/ -run TestClose_FullClose_ProducesCommit -count=20`
3. `go test ./internal/cli/ -run 'TestRunTemplateSyncWithProgress|TestRunUpdate' -v`
4. `go test -race ./...`
5. `GOOS=windows GOARCH=amd64 go build ./...` (cross-compile sanity — proves it compiles for Windows; does NOT prove the runtime guard works)
6. `golangci-lint run --timeout=2m`
7. `grep -n 'runtime.GOOS == "windows"' internal/cli/coverage_improvement_test.go` (confirm skips added)

## §F — Milestones (priority-ordered, no time estimates)

### M1 — FLAKY-1: relative-path resolution in performAtomicClose (Priority High)
- Normalize `baseDir` via `filepath.Abs` in `Close()` (REQ-CFS-002).
- In `performAtomicClose`, resolve `state.SpecMDPath` / `state.ProgressMDPath` to paths RELATIVE to `baseDir` via `filepath.Rel` before `git add` (REQ-CFS-001).
- Apply the same relative resolution to the `git restore --staged` rollback path (REQ-CFS-003).
- Keep `TestClose_FullClose_ProducesCommit` non-parallel (REQ-CFS-004).
- Verify: items 1, 2 of §E GREEN. Est ~20 LOC in `closer.go`.
- Rationale for first: deterministically verifiable locally; lowest risk; unblocks the ubuntu/macos CI race failures immediately.

### M2 — FLAKY-2: Windows test skips (Priority High)
- Audit ALL `runTemplateSyncWithProgress` (and other `ConfirmMerge`-reaching) callers in `coverage_improvement_test.go` (REQ-CFS-009).
- Add `if runtime.GOOS == "windows" { t.Skip(<bubbletea/ReadConsole reason>) }` to each caller that reaches `ConfirmMerge` without `--yes` (REQ-CFS-006), matching the existing `target_coverage_test.go:266` idiom verbatim.
- Verify: item 7 of §E. Low risk (test-only change).

### M3 — FLAKY-2: library-level TTY guard in merge/confirm.go (Priority Medium)
- In `ConfirmMerge`, before `tea.NewProgram().Run()`, add a guard: on Windows + non-TTY stdin, return a fail-open error instead of entering the TUI loop (REQ-CFS-007).
- Reuse `isatty.IsTerminal(os.Stdin.Fd())` (REQ-CFS-008) — no new dependency.
- Defense-in-depth backstop: even if a future test slips the M2 net, the library no longer hangs.
- Verify: items 4, 5, 6 of §E. Est 1 file, ~5-10 LOC.

### M4 — Green gate (Priority High)
- Run full §E batch (items 1-7).
- `go test -race ./...` local GREEN (REQ-CFS-010).
- Record in acceptance.md that Windows CI job-green is the deferred true gate for FLAKY-2 (REQ-CFS-011).
- Status transition draft → in-progress occurs at M1 commit (manager-develop owns).

## §G — Anti-Patterns to Avoid

- Making `TestClose_FullClose_ProducesCommit` parallel "to test the race" — breaks the atomicity intent; the race is in path resolution, not the test's parallelism.
- Naming a single test as THE FLAKY-2 trigger without the M2 audit — the `--yes` flag analysis shows most callers skip `ConfirmMerge`; only the audit identifies the real set.
- Claiming FLAKY-2 closed on local darwin pass — local cannot exercise the Windows `ReadConsole` path.
- Replacing the merge TUI or upgrading bubbletea — out of scope; minimal guard only.

## §H — Cross-References

- `internal/spec/closer.go`, `internal/spec/closer_test.go`
- `internal/merge/confirm.go`
- `internal/cli/update.go`, `internal/cli/coverage_improvement_test.go`, `internal/cli/target_coverage_test.go`
- Sibling: `SPEC-V3R6-DOCS-I18N-PARITY-001` (docs-i18n), `SPEC-V3R6-MAIN-RED-REMEDIATION-001` (template mirror-drift)
- `internal/cli/CLAUDE.md` § Cross-platform; `internal/spec/CLAUDE.md` § git log / cmd.Dir patterns
