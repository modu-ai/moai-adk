# Acceptance Criteria — SPEC-V3R5-ATOMIC-WRITE-001

All acceptance criteria are binary (PASS/FAIL) and independently verifiable via a single shell command. Each AC maps to one or more REQ entries in `spec.md` §4.

## Binary AC Matrix

| AC | Maps to REQ | Verification Command | Expected Output |
|----|-------------|---------------------|-----------------|
| AC-AWR-001 | REQ-AWR-001, REQ-AWR-005 | `grep -A 30 '^func atomicWrite' internal/migrate/hook_cleanup.go` | Output contains all of: `CreateTemp`, `Write`, `Sync`, `Close`, `Rename`, `Chmod`. Output does NOT contain: `os.WriteFile`. |
| AC-AWR-002 | REQ-AWR-002 | `grep -n '^func atomicWrite' internal/migrate/hook_cleanup.go` | Exactly one match. The signature on that line equals `func atomicWrite(path string, data []byte, perm os.FileMode) error`. |
| AC-AWR-003 | REQ-AWR-001, REQ-AWR-002, REQ-AWR-003 | `go test -run 'TestAtomicWrite' ./internal/migrate/...` | All test cases pass. The test file includes at minimum: happy-path write, mid-write error → temp removed, mid-rename error → temp removed, dest already exists → replaced atomically. |
| AC-AWR-004 | REQ-AWR-002 | `go test ./...` | Full suite passes. No existing migrate / hook test fails after the fix. |
| AC-AWR-005 | REQ-AWR-004 | `go test -run 'TestAtomicWritePreservesPerm' ./internal/migrate/...` | Test verifies that calling `atomicWrite(path, data, 0o644)` results in `os.Stat(path).Mode().Perm() == 0o644`, not `0o600`. |
| AC-AWR-006 | REQ-AWR-006, REQ-AWR-005 | `grep -B 4 '^func atomicWrite' internal/migrate/hook_cleanup.go` | Output contains `@MX:WARN` and `@MX:REASON` on lines immediately preceding the function declaration. |
| AC-AWR-007 | Coverage gate per `.moai/config/sections/quality.yaml` Go threshold | `go test -cover ./internal/migrate/...` | Output reports coverage ≥ 85% for `internal/migrate`. |
| AC-AWR-008 | NFR — cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | Exit code 0. No `undefined: ...` errors for the modified file. |

## Test Scenarios (Given-When-Then)

### Scenario 1 — Happy path write

**Given** a temporary directory `dir` and a destination path `dir/settings.json` that does not yet exist
**When** `atomicWrite(dir/settings.json, []byte("{\"hooks\":{}}"), 0o644)` is called
**Then** the function returns `nil`, AND `os.ReadFile(dir/settings.json)` returns `{"hooks":{}}`, AND `os.Stat(dir/settings.json).Mode().Perm() == 0o644`, AND the directory `dir` contains no files matching `.hook_cleanup_tmp_*`.

### Scenario 2 — Destination already exists, content is replaced atomically

**Given** a destination file containing `OLD` with mode `0o644`
**When** `atomicWrite(path, []byte("NEW"), 0o644)` is called
**Then** the function returns `nil`, AND the file content is exactly `NEW` (no `OLD` residue, no partial concatenation), AND the mode remains `0o644`.

### Scenario 3 — Mid-write error leaves destination untouched

**Given** a destination file containing `OLD` with mode `0o644`, AND a forced error injected at the `Write` step (test harness uses a custom `io.Writer` substitute or restricts disk quota)
**When** `atomicWrite(path, data, 0o644)` is called
**Then** the function returns a non-nil error wrapping the message `atomicWrite: write temp`, AND the destination file content is still exactly `OLD` (unchanged), AND no `.hook_cleanup_tmp_*` file remains in the directory.

### Scenario 4 — Rename failure leaves destination untouched and removes temp

**Given** a destination file containing `OLD`, AND a forced rename error (test harness uses a read-only parent directory or a path that cannot be renamed)
**When** `atomicWrite(path, []byte("NEW"), 0o644)` is called
**Then** the function returns a non-nil error wrapping the message `atomicWrite: rename`, AND the destination file content is still exactly `OLD`, AND no `.hook_cleanup_tmp_*` file remains.

### Scenario 5 — Permission preservation regardless of CreateTemp default

**Given** a fresh destination path
**When** `atomicWrite(path, []byte("X"), 0o644)` is called
**Then** the function returns `nil`, AND `os.Stat(path).Mode().Perm() == 0o644` (not `0o600` which is `os.CreateTemp`'s default).

### Scenario 6 — Caller `CleanupUserSettings` regression

**Given** a project root `dir` with `.claude/settings.json` containing one retired hook event entry under `hooks.Notification`
**When** `CleanupUserSettings(dir)` is called (which invokes `atomicWrite` at `hook_cleanup.go:109`)
**Then** the function returns `nil`, AND the resulting `.claude/settings.json` parses as valid JSON, AND the `hooks.Notification` key is absent, AND a `migration-<YYYY-MM-DD>.json` archive file exists under `.moai/archive/hooks/v3.0/` containing the removed entry.

## Quality Gates

| Gate | Threshold | Verification |
|------|-----------|--------------|
| Test suite | All pass | `go test ./...` exit 0 |
| Package coverage (`internal/migrate`) | ≥ 85% | `go test -cover ./internal/migrate/...` reports coverage line ≥ 85.0% |
| Cross-platform build | exit 0 | `GOOS=windows GOARCH=amd64 go build ./...` |
| Lint baseline | No NEW issues | `golangci-lint run --timeout=2m` — NEW issues in `internal/migrate/` = 0 |
| Race detector | All pass | `go test -race ./internal/migrate/...` |
| Function naming-semantic match | Match | `grep -A 30 '^func atomicWrite' internal/migrate/hook_cleanup.go` contains `Rename` and does NOT contain `os.WriteFile` |
| @MX annotation | Present | `grep -B 4 '^func atomicWrite' internal/migrate/hook_cleanup.go` contains both `@MX:WARN` and `@MX:REASON` |

## Definition of Done

- [ ] All 8 AC entries above PASS via the listed commands
- [ ] All 6 Given-When-Then scenarios are implemented as Go tests in `internal/migrate/hook_cleanup_test.go`
- [ ] `go test -cover ./internal/migrate/...` ≥ 85%
- [ ] `go test ./...` PASS
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] `golangci-lint run --timeout=2m` — zero NEW issues attributable to `internal/migrate/`
- [ ] `@MX:WARN` + `@MX:REASON` annotation present
- [ ] The 24 dirty files listed in `spec.md` EXCL-AWR-004 are unchanged
- [ ] Commit message follows Conventional Commits: `fix(SPEC-V3R5-ATOMIC-WRITE-001): replace os.WriteFile with atomic tmp+rename pattern in internal/migrate/hook_cleanup.go`
- [ ] Commit trailer: `🗿 MoAI <email@mo.ai.kr>`
