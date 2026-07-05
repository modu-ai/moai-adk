# SPEC-PRECOMMIT-001 — Acceptance Criteria

## §D AC Matrix

| AC | Requirement(s) | Type | Blocking |
|----|----------------|------|----------|
| AC-PC-001 | REQ-PC-001, REQ-PC-002 | installer create | yes |
| AC-PC-002 | REQ-PC-004 | foreign-hook preserve | yes |
| AC-PC-003 | REQ-PC-003 | marker overwrite | yes |
| AC-PC-004 | REQ-PC-015, REQ-PC-016 | byte-identity | yes |
| AC-PC-005 | REQ-PC-008, REQ-PC-010a, REQ-PC-010b | gofmt block | yes |
| AC-PC-006 | REQ-PC-012 | bypass env | yes |
| AC-PC-007 | REQ-PC-019 | cross-platform build | yes |
| AC-PC-008 | REQ-PC-009 | no-staged-Go no-op | yes |
| AC-PC-009 | REQ-PC-011a, REQ-PC-011b | go vet block | yes |
| AC-PC-010 | REQ-PC-006 | `--no-hooks` skip | yes |
| AC-PC-011 | REQ-PC-014 | 2-tier boundary (no ci-local) | yes |
| AC-PC-012 | REQ-PC-018, REQ-PC-007 | install assertion (POSIX shebang + marker + 4 tokens) | yes |
| AC-PC-013 | REQ-PC-013 | toolchain-absent graceful skip | yes |
| AC-PC-014 | REQ-PC-005 | non-fatal install failure | yes |
| AC-PC-015 | REQ-PC-017 | config-agnostic install (skip/warn/enforce identical) | yes |

## Given-When-Then Scenarios

### AC-PC-001 — New pre-commit hook created when none exists

- **Given** a git repository whose `.git/hooks/pre-commit` does not exist,
- **When** `InstallPreCommitHook(false)` runs (via `moai init` / `moai update`),
- **Then** `.git/hooks/pre-commit` is created with mode `0755`, its content equals
  `preCommitHookContent`, and the call returns `nil`.

### AC-PC-002 — Foreign (non-marker) pre-commit hook preserved

- **Given** a `.git/hooks/pre-commit` whose first 3 lines do NOT contain
  `# MoAI-ADK pre-commit hook` (a user's own hook),
- **When** `InstallPreCommitHook(false)` runs,
- **Then** the file is left byte-unchanged AND the call returns `ErrUserHookExists`;
  the optional wrapper prints a "existing pre-commit hook preserved" note and does NOT abort init/update.

### AC-PC-003 — MoAI-marker pre-commit hook overwritten safely

- **Given** a `.git/hooks/pre-commit` whose first 3 lines contain
  `# MoAI-ADK pre-commit hook` (a stale MoAI hook),
- **When** `InstallPreCommitHook(false)` runs,
- **Then** the file is overwritten with the current `preCommitHookContent` and the call returns `nil`.

### AC-PC-004 — Template ↔ constant byte-identity

- **Given** the template `internal/template/templates/.git_hooks/pre-commit` and the
  constant `preCommitHookContent`,
- **When** the new byte-identity test runs (`go test ./internal/cli/ -run TemplateMatchesConstant` or the pre-commit-specific test),
- **Then** the two are byte-identical and the test passes (the test skips gracefully
  when the template file is absent, matching the pre-push test's tarball-safe behavior).

### AC-PC-005 — Staged un-gofmt'd Go file blocks the commit

- **Given** an installed MoAI pre-commit hook and a staged `.go` file whose content is
  not `gofmt`-clean,
- **When** the hook runs (`git commit` in a temp repo),
- **Then** the hook prints a gofmt remediation hint to stderr and exits 1 (commit blocked).

### AC-PC-006 — `SKIP_MOAI_PRECOMMIT=1` bypass

- **Given** an installed MoAI pre-commit hook and a staged un-gofmt'd `.go` file,
- **When** the hook runs with `SKIP_MOAI_PRECOMMIT=1` in the environment,
- **Then** the hook prints a bypass notice to stderr and exits 0 without running any
  check (commit allowed).

### AC-PC-007 — Cross-platform build

- **Given** the new installer + constant,
- **When** `GOOS=windows GOARCH=amd64 go build ./...` runs (and the native `go build ./...`),
- **Then** both exit 0 (no OS-specific primitives; no build-tag split required).

### AC-PC-008 — No staged Go files → fast no-op

- **Given** an installed MoAI pre-commit hook and a commit that stages only non-Go
  files (or nothing),
- **When** the hook runs,
- **Then** the hook exits 0 immediately without invoking `gofmt` or `go vet`.

### AC-PC-009 — `go vet` failure blocks the commit

- **Given** an installed MoAI pre-commit hook and a staged `.go` file that is
  `gofmt`-clean but triggers a `go vet` diagnostic in its package,
- **When** the hook runs,
- **Then** the hook prints a `go vet` hint to stderr and exits 1 (commit blocked).

### AC-PC-010 — `--no-hooks` skip

- **Given** `moai init` / `moai update` invoked with `--no-hooks`,
- **When** `installPreCommitHookOptional(projectRoot, true, out)` runs,
- **Then** no `.git/hooks/pre-commit` is written and no install message is printed (no-op).

### AC-PC-011 — 2-tier boundary (no heavy CI in commit tier)

- **Given** the installed pre-commit hook content,
- **When** its body is inspected,
- **Then** it contains NO `make ci-local`, NO `golangci-lint`, and NO `go test`
  invocation token (the fast subset is `gofmt -l` + `go vet` only).

### AC-PC-012 — Install assertion (POSIX shebang + marker + 4 tokens)

- **Given** a fresh repository,
- **When** the pre-commit hook is installed,
- **Then** the installed file begins with the `#!/bin/sh` POSIX shebang (REQ-PC-007) and
  its content contains all four tokens: the `# MoAI-ADK pre-commit hook` marker,
  `gofmt -l`, `go vet`, and `SKIP_MOAI_PRECOMMIT`.

### AC-PC-013 — Toolchain-absent graceful skip (highest blast radius — all 16-language users)

- **Given** neither `go` nor `gofmt` is resolvable on `PATH` (a non-Go project, or a
  Go project on a machine without the toolchain),
- **When** the installed pre-commit hook runs on a commit that stages one or more `.go`
  files,
- **Then** each check is independently skipped (its `command -v` guard fails) and the
  hook exits 0 — a non-Go environment must NEVER have every `git commit` blocked by an
  absent `go vet` / `gofmt`. (This is the deployment-safety guarantee for the 16-language
  template audience; the hook ships to all moai users, most of whom have no Go toolchain.)

### AC-PC-014 — Non-fatal install failure (general error path)

- **Given** a repository whose `.git/hooks/` directory cannot be written (e.g. a
  read-only mount or a permission-denied path) and no foreign-hook condition applies,
- **When** `installPreCommitHookOptional(projectRoot, false, out)` runs during
  `moai init` / `moai update`,
- **Then** a non-fatal warning is written to the output writer AND init/update completes
  normally (the optional wrapper swallows the error, never aborts). This exercises the
  REQ-PC-005 general install-failure path — distinct from AC-PC-002, which covers ONLY
  the `ErrUserHookExists` foreign-hook branch that REQ-PC-005 explicitly excludes.

### AC-PC-015 — Config-agnostic install (skip / warn / enforce identical)

- **Given** `git_strategy.<mode>.hooks.pre_commit` set to each of `skip`, `warn`, and
  `enforce` in turn,
- **When** the pre-commit installer runs,
- **Then** the pre-commit hook is installed identically in all three cases (same `0755`
  file, same `preCommitHookContent`) — the config tri-state is NOT read as an install
  gate (REQ-PC-017). The installer is config-agnostic; the tri-state's semantic is a
  deferred hook-run-time severity dial, out of scope here.

## Edge Cases

- Staged file staged for deletion (`--diff-filter` excludes `D` via `ACM`) → not passed
  to gofmt/vet (a deleted file cannot be format-checked).
- Path with spaces in a staged Go filename → hook iteration must be whitespace-safe
  (run-phase implementation concern; test with a spaced path if feasible).
- `gofmt` present but `go` absent (or vice versa) → each check independently guarded by
  `command -v` (REQ-PC-013); the present tool still runs, the absent one is skipped.
- Re-running `moai update` on a repo that already has the current MoAI pre-commit hook →
  idempotent overwrite (AC-PC-003 path), returns `nil`.

## Quality Gate Criteria

- All 15 ACs PASS (all blocking).
- `go test ./...` green; `internal/cli` coverage for added functions ≥ 85%.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `golangci-lint run` introduces no NEW findings.
- Pre-push installer suite (`TestPrePushTemplateMatchesConstant` + pre-push installer
  tests) remains green (no regression).

## Definition of Done

- Template `internal/template/templates/.git_hooks/pre-commit` + `preCommitHookContent`
  constant exist and are byte-identical (AC-PC-004), verified after `make build`.
- `PreCommitInstaller` + `installPreCommitHookOptional` implemented and wired into
  `moai init` (`init.go`) and `moai update` (`update.go`) behind `--no-hooks`.
- All 15 ACs PASS with evidence recorded in the run-phase self-verification matrix.
- No modification to pre-push code/template/test and no config default changes.
- Changes committed path-limited with a Conventional-Commit subject; no unrelated
  parallel-session files included.
