---
id: SPEC-V3R6-CI-BASELINE-DRIFT-001
title: "CI baseline drift cleanup — plan"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P3
phase: v3.0.0
module: "internal/cli, internal/config, internal/statusline, internal/cli/wizard, internal/constitution, internal/template, internal/merge, internal/tmux"
lifecycle: spec-anchored
tier: S
tags: "ci, lint, baseline, drift, technical-debt, tier-s, v3.0"
---

# Plan — CI baseline drift cleanup

## Section A — Context and Scope

CI baseline drift accumulated across the v3.0.0-rc1 development cycle is the cleanup target. The drift surfaced as 4 CI failures at SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 PR #1048 admin merge (commit `767bc04a4`, 2026-05-23). This SPEC is a Tier S technical-debt restoration — no behavior changes, only baseline alignment.

Three categories (per `spec.md` §A.1-A.3):
- **Category A**: 27 golangci-lint issues
- **Category B**: TestStatus golden v2.17.0 → v3.0.0-rc1
- **Category C**: ConfigManager race on `sunsetNoticeOnce sync.Once`

## Section B — Approach

### B.1 — Sequencing rationale (M1 → M2 → M3)

M1 (TestStatus golden) is the fastest and lowest-risk milestone: 3 file edits with mechanical version string substitution. Surfaces any other golden drift early.

M2 (Lint 27 → 0/scope-deferred) is the largest milestone in terms of file count but each fix is isolated. Splitting by lint category (errcheck → ineffassign → staticcheck → unused) allows incremental verification.

M3 (ConfigManager race) is the deepest milestone: root-cause fix to `internal/config/sunset_notice.go` requires synchronization primitive selection and a regression test to prevent re-introduction.

### B.2 — Per-category fix strategy

**errcheck (8 sites)** — `internal/cli/harness_mute.go` (4) + `internal/template/seq_thinking_retire_audit_test.go` (2 defer-close):
- Strategy: explicit `if _, err := fmt.Fprintln(...); err != nil { ... }` or assign to `_` only when justified
- For `defer f.Close()` in test files: wrap with `_ = f.Close()` or local closure that captures result

**ineffassign (1 site)** — `internal/cli/agent_lint_test.go:2008`:
- Strategy: examine the loop body; either remove the ineffectual `p :=` or use the value

**staticcheck (5 sites)**:
- `internal/merge/confirm.go:559,637` — QF1012 mechanical rewrite (`b.WriteString(fmt.Sprintf(...))` → `fmt.Fprintf(&b, ...)`)
- `internal/tmux/session_sensitive_test.go:35,80,114` — SA4032: build constraints exclude windows. Either remove `runtime.GOOS == "windows"` branches (dead under build tag) or relax build constraints. M2 recommendation: remove the dead checks (lower risk)

**unused (13 sites)** — see §D.1 for per-entry scope-defer decisions

### B.3 — Sunset notice race fix design (M3)

Root cause: `internal/config/sunset_notice.go:37` rewrites `sunsetNoticeOnce sync.Once` from test code without synchronization. Multiple parallel tests racing.

**Fix options evaluated**:
1. **Add `sync.Mutex` around `sunsetNoticeOnce` access** — minimal change, preserves once-per-process semantics
2. **Replace `sync.Once` with `atomic.Bool` + explicit reset under mutex** — clearer test-reset semantics
3. **Remove `resetSunsetNoticeOnce()` and refactor tests to use a new `Loader` instance per test** — eliminates global state but requires test refactoring

**Selected**: Option 1 (mutex). Reason: minimal blast radius, preserves REQ-MIG003-018 contract, fits Tier S scope. Implementation:
```go
var (
    sunsetNoticeMu   sync.Mutex
    sunsetNoticeOnce sync.Once
)

func emitSunsetDormantNotice(sectionsDir string) {
    sunsetNoticeMu.Lock()
    once := sunsetNoticeOnce
    sunsetNoticeMu.Unlock()
    once.Do(func() { /* existing body */ })
}

func resetSunsetNoticeOnce() {
    sunsetNoticeMu.Lock()
    defer sunsetNoticeMu.Unlock()
    sunsetNoticeOnce = sync.Once{}
}
```

Note: the captured-`once` pattern above ensures `Do()` runs on the snapshot, avoiding lock-held callback. Alternative: lock-while-Do (acceptable for slog.Info side-effect).

### B.4 — Behavior preservation guarantees

- `moai status` output: only the version-line substring `v2.17.0` → `v3.0.0-rc1` changes; column widths, box drawing characters, all other content identical (verified by `diff` of golden files post-edit)
- DORMANT notice (REQ-MIG003-018): at-most-once-per-process contract preserved (sync.Once semantics retained, just protected by mutex)
- No public API signature changes (REQ-CBD-004)
- No file moves, no package renames, no exported symbol removals

### B.5 — Verification batch (parallel)

Per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution, M1+M2+M3 completion verification batched as single-turn multi-Bash:
1. `golangci-lint run --timeout=2m`
2. `go test -race -count=1 ./internal/config/...`
3. `go test -count=1 -run TestStatus ./internal/cli/...`
4. `go test -count=1 ./...`
5. `grep -rn "AskUserQuestion\|mcp__askuser" internal/harness/ internal/hook/ | grep -v "_test.go"` (C-HRA-008 sentinel)
6. `go run ./cmd/moai --version`
7. `go vet ./...`

## Section C — Pre-flight checklist

Before M1 start, manager-develop verifies:
- [ ] `git status` clean (no uncommitted changes on target files)
- [ ] HEAD on main or fresh feat/SPEC-V3R6-CI-BASELINE-DRIFT-001 branch
- [ ] `golangci-lint --version` reports v1.55+ (or current project pinned version per CI config)
- [ ] `go test -race -count=1 ./internal/config/...` reproduces the race (M3 baseline confirmed)
- [ ] Three golden files exist at `internal/cli/testdata/status-{nocolor,light,dark}.golden`
- [ ] `pkg/version.Version` reads `v3.0.0-rc1` (sanity check on SoT)

## Section D — Milestones

### M1 — TestStatus golden update (lowest risk, fastest)

**Deliverables**:
1. `internal/cli/testdata/status-nocolor.golden` — line 6 version string `v2.17.0` → `v3.0.0-rc1`
2. `internal/cli/testdata/status-light.golden` — same
3. `internal/cli/testdata/status-dark.golden` — same
4. Regenerate goldens via `go test -update -run TestStatus ./internal/cli/...` if such flag exists, OR mechanical edit per AC

**Verification**: `go test -count=1 -run TestStatus ./internal/cli/...` PASS
**AC covered**: AC-CBD-001
**Estimated edits**: 3 lines across 3 files

### M2 — Lint baseline restoration (27 → ≤2 deferred)

Split into 4 sub-milestones by lint category:

**M2.1 — errcheck (8 → 0)**:
- `internal/cli/harness_mute.go:141,145,173,186` — explicit error handling for `fmt.Fprintln/Fprintf` to `cmd.OutOrStdout()`
- `internal/template/seq_thinking_retire_audit_test.go:95,135` — `defer` close pattern fix
- *(2 sites may need re-survey — total errcheck count from output is 8, M2.1 verification must reduce to 0)*

**M2.2 — ineffassign (1 → 0)**:
- `internal/cli/agent_lint_test.go:2008` — remove ineffectual reassignment

**M2.3 — staticcheck (5 → 0)**:
- `internal/merge/confirm.go:559,637` — QF1012 rewrite (`b.WriteString(fmt.Sprintf(...))` → `fmt.Fprintf(&b, ...)`)
- `internal/tmux/session_sensitive_test.go:35,80,114` — SA4032 remove dead-under-build-tag checks

**M2.4 — unused (13 → ≤2 deferred)**:
- **Remove** (safe deletions):
  - `internal/constitution/validator.go:151` — `var hardRuleRegexp` (1 entry, no follow-up plan)
- **Scope-defer with `//nolint:unused`** (work-in-progress modules):
  - `internal/cli/wizard/review.go:15-17,48,119,152` — Review wizard module is feature-complete but not wired into init flow (6 entries). DEFER decision: retain with `//nolint:unused // wired-in via SPEC-V3R6-INIT-WIZARD-REVIEW-001 follow-up` directive on each symbol. Create follow-up SPEC stub.
  - `internal/statusline/renderer.go:145,682` — v3 statusline variant (2 entries). DEFER decision: retain with `//nolint:unused // selected via SPEC-V3R6-STATUSLINE-V3-SELECTOR-001 follow-up` directive. Create follow-up SPEC stub.
- **Remove or refactor** (case-by-case):
  - `internal/cli/branch_protection.go:35,38` — `ttyConfirmer` struct + method (2 entries). Inspect: if it shadows a planned interface implementation, defer; otherwise remove.
  - `internal/cli/init_layout.go:21,61` — `renderInitHeader`, `renderInitNextSteps` (2 entries). Inspect: if init UX redesign is planned, defer; otherwise remove.

**Verification**: `golangci-lint run --timeout=2m` reports ≤2 issues (only deferred entries with `//nolint:unused` retained)
**AC covered**: AC-CBD-002, AC-CBD-003, AC-CBD-004, AC-CBD-005
**Estimated edits**: ~15-20 lines across ~10 files

### M3 — ConfigManager race fix

**Deliverables**:
1. `internal/config/sunset_notice.go` — add `sunsetNoticeMu sync.Mutex`, wrap `emitSunsetDormantNotice` and `resetSunsetNoticeOnce` with mutex (per §B.3 selected design)
2. `internal/config/sunset_notice_race_test.go` (NEW, optional per REQ-CBD-006) — deterministic race reproducer: spawn N goroutines, each calling `resetSunsetNoticeOnce()` + `emitSunsetDormantNotice(tmpDir)` concurrently; assert no race
3. Optional refactor: rename `sunsetNoticeOnce` access to package-private helper if call sites multiply

**Verification**:
- `go test -race -count=1 ./internal/config/...` PASS (no races)
- Existing tests that call `resetSunsetNoticeOnce()` (e.g., `TestEmitSunsetDormantNotice*`) still PASS
- DORMANT notice still emits at most once per process (manual verification: run twice in same process, assert single emission)

**AC covered**: AC-CBD-006, AC-CBD-007
**Estimated edits**: ~10-15 lines in `sunset_notice.go` + ~40-60 lines for optional race reproducer test

## Section E — Risks and Mitigations

### Risk R1: Hidden errcheck/staticcheck sites discovered during fix
**Likelihood**: Medium
**Impact**: Scope creep — final lint count may exceed initial 27
**Mitigation**: Re-run `golangci-lint run --timeout=2m` after each sub-milestone (M2.1, M2.2, M2.3, M2.4). If new issues surface in unchanged files, document in `progress.md` and either fix in this SPEC or document as deferred per §D.1.

### Risk R2: M2.4 unused-defer decisions disputed
**Likelihood**: Medium
**Impact**: Re-delegation cycle if plan-auditor rejects defer rationale
**Mitigation**: Each deferred symbol gets `//nolint:unused // <reason>` with explicit follow-up SPEC ID. Follow-up SPEC stub created in same commit. Rationale: "feature-complete but unwired" is a legitimate technical-debt category and removal would lose implementation context.

### Risk R3: M3 mutex design alternative preferred
**Likelihood**: Low
**Impact**: M3 fix-forward if reviewer prefers atomic.Value or test-instance refactor
**Mitigation**: Document design decision in commit body (Notable Implementation Decision). All 3 options preserve REQ-MIG003-018; option 1 is simplest. If reviewer rejects, fix-forward swap is mechanical.

### Risk R4: Race reproducer test (REQ-CBD-006) is flaky
**Likelihood**: Low-Medium
**Impact**: Flake fails CI intermittently, masking real failures
**Mitigation**: REQ-CBD-006 is `Optional` (Where possible). If race reproducer is flaky, mark with `t.Skip("flaky race reproducer — see SPEC-V3R6-CI-BASELINE-DRIFT-001 R4")` and rely on `go test -race` over the full `./internal/config/...` package as primary verification.

### Risk R5: pkg/version SoT changes mid-SPEC
**Likelihood**: Very Low
**Impact**: Golden files updated to one version, SoT bumps to another
**Mitigation**: M1 is the first milestone — fast turnaround. If SoT bumps during the SPEC, re-run M1 with the new value before final commit.

### Risk R6: ConfigManager race surfaces additional racing globals
**Likelihood**: Low
**Impact**: M3 expands beyond `sunsetNoticeOnce`
**Mitigation**: After M3 mutex fix, re-run `go test -race -count=1 ./internal/config/...`. If additional races surface, scope-decide: (a) fix in this SPEC if mechanical, (b) defer to follow-up SPEC `SPEC-V3R6-CONFIG-RACE-AUDIT-001` if architectural.

## Section F — Out-of-Scope Reminders

Per `spec.md` Exclusions, the following are explicitly NOT in scope:
- v3.0.0-rc2/GA release version bump
- Behavioral changes to `moai status` or any CLI command
- REQ-MIG003-018 contract redesign
- `internal/cli/wizard/review.go` removal (deferred per §D.1)
- pkg/version ldflags mechanism changes
- Broader CI workflow modifications
- Dependency upgrades, Go toolchain bumps, golangci-lint config changes
