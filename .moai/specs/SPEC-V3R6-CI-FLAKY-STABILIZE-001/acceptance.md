# Acceptance Criteria — SPEC-V3R6-CI-FLAKY-STABILIZE-001

## §D — AC Matrix

| AC | Requirement | Verification | Severity | Gate |
|----|-------------|--------------|----------|------|
| AC-CFS-001 | REQ-CFS-001/002/003 | `go test -race ./internal/spec/ -count=20` → 0 failures, 0 race findings | MUST-PASS | local |
| AC-CFS-002 | REQ-CFS-005 | `go test -race ./internal/spec/ -run TestClose_FullClose_ProducesCommit -count=20` → all PASS | MUST-PASS | local |
| AC-CFS-003 | REQ-CFS-004 | `grep -L 't.Parallel()' ...` proof: `TestClose_FullClose_ProducesCommit` body contains NO `t.Parallel()` call | MUST-PASS | local |
| AC-CFS-004 | REQ-CFS-001 | `performAtomicClose` passes `filepath.Rel(baseDir, ...)`-derived paths to `git add` (code inspection) | MUST-PASS | local |
| AC-CFS-005 | REQ-CFS-002 | `Close()` normalizes `baseDir` via `filepath.Abs` before git ops (code inspection) | MUST-PASS | local |
| AC-CFS-006 | REQ-CFS-006/009 | `grep -c 'runtime.GOOS == "windows"' internal/cli/coverage_improvement_test.go` ≥ 1; every `ConfirmMerge`-reaching caller without `--yes` has the skip | MUST-PASS | local |
| AC-CFS-007 | REQ-CFS-007/008 | `merge.ConfirmMerge` contains a `runtime.GOOS == "windows"` + `isatty.IsTerminal(os.Stdin.Fd())` guard returning an error before `tea.NewProgram().Run()` (code inspection) | MUST-PASS | local |
| AC-CFS-008 | REQ-CFS-008 | `grep -c 'mattn/go-isatty' go.mod` unchanged (no NEW dependency added); `git diff go.mod` shows no new require line | MUST-PASS | local |
| AC-CFS-009 | REQ-CFS-010 | `go test -race ./...` → 0 failures (excluding out-of-scope docs-i18n / mirror-drift packages) | MUST-PASS | local |
| AC-CFS-010 | REQ-CFS-010 | `GOOS=windows GOARCH=amd64 go build ./...` exits 0 (compiles for Windows) | MUST-PASS | local |
| AC-CFS-011 | REQ-CFS-006/007 | Windows CI job (`windows-latest`) completes `internal/cli` + `internal/harness` UNDER the 600s timeout, 0 hangs | MUST-PASS | **CI-deferred** |
| AC-CFS-012 | REQ-CFS-011 | The full CI matrix (ubuntu + macos + windows) Test job is GREEN on `origin/main` after merge | MUST-PASS | **CI-deferred** |

## §D.1 — Given-When-Then Scenarios

### Scenario 1 — FLAKY-1: parallel-pressure race elimination (local-verifiable)
- **Given** `internal/spec` has 15 `t.Parallel()` sibling tests and the close orchestrator stages via `git add` with `cmd.Dir = baseDir`,
- **When** the suite runs `go test -race ./internal/spec/ -count=20` under the race detector,
- **Then** zero race findings and zero failures are reported, because `performAtomicClose` now passes repo-root-relative paths (via `filepath.Rel`) consistently resolvable regardless of `cmd.Dir` filesystem timing, and `baseDir` is absolute.

### Scenario 2 — FLAKY-1: atomicity preserved
- **Given** `TestClose_FullClose_ProducesCommit` documents the close transaction as atomic and intentionally serial,
- **When** the M1 fix is applied,
- **Then** the test body still contains NO `t.Parallel()` call — the fix targets path resolution, not test parallelism.

### Scenario 3 — FLAKY-2: Windows non-TTY guard (library-level, CI-verifiable)
- **Given** a Windows runner with non-TTY stdin and a test that reaches `merge.ConfirmMerge`,
- **When** `ConfirmMerge` is invoked,
- **Then** it returns a fail-open error BEFORE `tea.NewProgram().Run()` (no `ReadConsole` block), because the `runtime.GOOS == "windows" && !isatty.IsTerminal(os.Stdin.Fd())` guard short-circuits the TUI loop.

### Scenario 4 — FLAKY-2: test-level skips (CI-verifiable)
- **Given** `coverage_improvement_test.go` callers that reach `ConfirmMerge` without `--yes`,
- **When** the suite runs on `windows-latest`,
- **Then** each such caller is skipped via the `runtime.GOOS == "windows"` guard matching the existing `target_coverage_test.go:266` idiom, so `internal/cli` completes under the 600s timeout.

## §D.2 — Edge Cases

- **`baseDir` already absolute**: `filepath.Abs` on an already-absolute path is idempotent — no regression.
- **`progressMDPath == ""`** (SPEC without progress.md): the relative-path resolution must guard the empty-path case (skip `filepath.Rel` on empty) — staging proceeds with spec.md only.
- **`filepath.Rel` error** (paths on different volumes — Windows edge): fall back to the absolute path or surface a clear error; do NOT silently swallow.
- **Non-Windows, non-TTY** (e.g. CI on linux with piped stdin reaching `ConfirmMerge`): the M3 guard is gated on `runtime.GOOS == "windows"` so linux/darwin behavior is unchanged (the bubbletea hang is Windows-specific).
- **TTY present on Windows** (developer running interactively): `isatty.IsTerminal` returns true → guard does not fire → normal interactive TUI works.

## §D.3 — Windows Verifiability Caveat (HONESTY GATE)

[HARD] FLAKY-2 CANNOT be fully verified in the local development environment — there is no Windows runner. Local darwin/linux passing is **necessary but NOT sufficient** for FLAKY-2.

- AC-CFS-001 through AC-CFS-010 are LOCAL gates — verifiable on darwin/linux at run-phase.
- **AC-CFS-011 and AC-CFS-012 are CI-DEFERRED gates** — the `windows-latest` CI job completing under timeout (AC-CFS-011) and the full matrix Test job going GREEN (AC-CFS-012) are the TRUE completion proof for FLAKY-2.
- The SPEC MUST NOT be marked `completed` on local evidence alone. Closure requires observing the post-merge CI run's Windows job pass. If the first post-merge CI run still fails Windows, this is a run-phase blocker requiring re-iteration (the library-level guard in M3 is the defense-in-depth backstop precisely for this risk).

## §D.4 — Definition of Done

- [ ] M1-M4 milestones complete.
- [ ] AC-CFS-001..010 (local gates) all PASS on darwin.
- [ ] `go test -race ./...` GREEN locally (excluding out-of-scope packages).
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- [ ] `golangci-lint run` clean.
- [ ] No new dependency in `go.mod` (AC-CFS-008).
- [ ] `TestClose_FullClose_ProducesCommit` remains non-parallel (AC-CFS-003).
- [ ] Out-of-scope boundary respected: no `docs-site/`, `internal/template/templates/`, or mirror-drift fixture changes.
- [ ] **CI-deferred**: AC-CFS-011 / AC-CFS-012 confirmed on the post-merge CI run (Windows job under timeout + full matrix GREEN). Until observed, FLAKY-2 is "fix applied, awaiting CI confirmation" — NOT closed.
