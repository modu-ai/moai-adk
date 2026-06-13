# Acceptance Criteria — SPEC-GITSTRATEGY-SAVE-ISOLATION-001

> PRIMARY AC: `TestWriteProjectConfigSectionIsolation` passes. No-regression AC: the SAVE-WIRING round-trip tests stay green. The reproduction test is the failing test itself (CLAUDE.md §7 Rule 4 — reproduction-first).

## §A. Acceptance Criteria (Given-When-Then)

### AC-GSI-001 — Primary: project-config write does not touch git_strategy (PRIMARY)

- **Given** a project whose `git-strategy.yaml` on disk contains only `git_strategy:\n  sentinel: DO_NOT_TOUCH\n`
- **When** a scoped project-config write runs (`writeProjectConfig` → `SetSection("quality")` + `SetSection("git_convention")` → `Save()`, NOT touching git_strategy)
- **Then** `git-strategy.yaml` content is byte-identical to before the write (the sentinel survives; no expansion into the default tree)
- **Verify**: `go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation` exits 0 (was RED before the fix).
- Maps: REQ-GSI-001, REQ-GSI-002.

### AC-GSI-002 — No-regression: git_strategy set→save→reload round-trip

- **Given** a config manager with `git-strategy.yaml` present and a probe git_strategy with non-default values (`Mode="personal"`, `Team.Hooks.PrePush="enforce"`)
- **When** `SetSection("git_strategy", probe)` then `Save()` then a fresh `Load()`
- **Then** the fresh load recovers `Mode == "personal"` and `Team.Hooks.PrePush == "enforce"`
- **Verify**: `go test ./internal/config/ -run TestConfigManagerSaveGitStrategyRoundTrip` exits 0 (stays GREEN — must not regress).
- Maps: REQ-GSI-003.

### AC-GSI-003 — No-regression: greenfield Save creates git-strategy.yaml

- **Given** a fresh project root with NO `git-strategy.yaml` on disk
- **When** `Load()` (no SetSection) then `Save()`
- **Then** `git-strategy.yaml` is created carrying the top-level `git_strategy:` key
- **Verify**: `go test ./internal/config/ -run TestConfigManagerSaveCreatesGitStrategyFile` exits 0 (stays GREEN — the file-absent create disjunct must hold).
- Maps: REQ-GSI-004.

### AC-GSI-004 — No caller regression across all Save() consumers

- **Given** the four production callers of `ConfigManager.Save()` (`internal/web` ×2, `internal/cli/profile_setup.go`, `internal/profile/sync.go`)
- **When** the full affected-package test suite runs
- **Then** every package passes with no call-site signature change
- **Verify**: `go test ./internal/config/... ./internal/web/... ./internal/cli/... ./internal/profile/...` exits 0.
- Maps: REQ-GSI-005.

### AC-GSI-005 — Guard test integrity (not weakened)

- **Given** the fix is applied
- **When** inspecting `internal/web/projectconfig_scope_test.go`
- **Then** the sentinel byte-equality assertion (`if string(after) != string(before[name])`) is intact, no `t.Skip` was added, and the assertion was not loosened
- **Verify**: `git diff` shows zero changes to `projectconfig_scope_test.go` (the reproduction passes because Save was fixed, not because the test was changed).
- Maps: REQ-GSI-006.

## §B. Edge Cases

- **EC-1 — git_strategy explicitly set THEN out-of-scope save**: if a session does `SetSection("git_strategy", x)` and later `Save()`, git-strategy.yaml IS written (dirty flag true) — the round-trip (AC-GSI-002) covers this.
- **EC-2 — existing file, no git_strategy modification**: existing `git-strategy.yaml` with user content (not just sentinel) is left byte-unchanged on a Save() that did not modify git_strategy (the isolation invariant generalizes beyond the sentinel fixture).
- **EC-3 — dirty flag reset**: after a successful `Save()` that wrote git-strategy.yaml, a subsequent `Save()` with no further `SetSection("git_strategy")` does NOT rewrite it (flag reset post-Save). Race-safe under `m.mu`.
- **EC-4 — concurrent SetSection/Save**: the dirty flag is mutated only under the existing `m.mu` lock; `go test -race ./internal/config/...` is clean.

## §C. Quality Gate Criteria

- `go test ./internal/config/... ./internal/web/... ./internal/cli/... ./internal/profile/...` — GREEN
- `go test -race ./internal/config/...` — GREEN
- `go vet ./internal/config/...` — clean
- `golangci-lint run ./internal/config/... ./internal/web/...` — clean
- Coverage for `internal/config` does not decrease versus the pre-fix baseline.

## §D. Definition of Done

- [ ] AC-GSI-001 (primary reproduction) flips RED → GREEN
- [ ] AC-GSI-002 + AC-GSI-003 (SAVE-WIRING no-regression) stay GREEN
- [ ] AC-GSI-004 (no caller regression) GREEN across all four Save() consumers
- [ ] AC-GSI-005 guard test body byte-unchanged (`git diff` clean for projectconfig_scope_test.go)
- [ ] No change to git_strategy schema / validators / defaults / loader contract / templates (EX-1..EX-3)
- [ ] No SEC-HARDEN file touched (EX-4)
- [ ] Fix is git_strategy-scoped, not a generic Save() redesign (EX-5)
- [ ] `go test -race`, `go vet`, `golangci-lint` all clean
- [ ] Coverage for `internal/config` not decreased
