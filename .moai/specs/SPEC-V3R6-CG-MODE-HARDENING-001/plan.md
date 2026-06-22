# Implementation Plan — SPEC-V3R6-CG-MODE-HARDENING-001

## §A. Context

Hardening of the `moai cg` (Claude leader + GLM teammate) launch mode. Ten requirements span four concerns: a cross-platform launch-safety fix, a CG-settings atomicity/ordering cluster, the headline detector-SSOT realignment, and supporting doc/precondition/security/coverage fixes. All defects were re-verified against the cited source during plan-phase (see spec.md §A.1).

Tier classification: **M** (standard). Justification:
- Multi-file: `internal/cli/launcher.go`, `internal/cli/glm.go`, `internal/cli/settings.go`, `internal/tmux/cg_detect.go`, `internal/config/validation.go`, plus new/updated test files and a build-tag split.
- A non-trivial behavioral redesign (detector SSOT) that must reconcile with a sibling SPEC's owned code.
- A documentation edit with a conditional template-sync + `make build` branch.
- Security validation logic at a config boundary.
- Not Tier L: the change is contained to the CG/GLM launch path; no cross-cutting architecture rewrite, no new subcommand, no migration. Not Tier S: more than a single-file mechanical fix; it spans 5 production files + a detector redesign + tests.

## §B. Known Issues / Risks

- **R1 — Sibling-SPEC collision**: `cg_detect.go` and `REQ-WTL-009` belong to `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`. REQ-CGH-006 must reconcile, not delete. Mitigation: keep the sibling's `IsCGMode` test green; add new tests rather than rewriting old ones; preserve the drift-warning emission semantics.
- **R2 — Windows build verification can't run natively on macOS dev box**: `GOOS=windows GOARCH=amd64 go build ./...` is a cross-compile-only check; runtime behavior on Windows is not exercised by CI here. Mitigation: mirror the proven `update.go:461` pattern verbatim and reuse the `team_launch_posix.go`/`team_launch_windows.go` build-tag idiom that already ships.
- **R3 — Atomicity helper refactor blast radius**: routing three existing mutators (`removeGLMEnv`, `syncPermissionModeToSettingsLocal`, `ensureSettingsLocalJSON`/`injectGLMEnvForTeam`) through one helper risks regressing user-key preservation. Mitigation: characterization tests on each mutator FIRST (DDD PRESERVE), then refactor.
- **R4 — Detector SSOT change is environment-sensitive**: `tmux show-environment` requires a real tmux session; tests must inject a fake session-env reader rather than depend on a live tmux. Mitigation: introduce a NEW package-level injectable seam `var sessionEnvReaderFn = <tmux show-environment reader>` (no existing equivalent — the worktree layer calls `tmux.NewSessionManager()` directly without a recording-fake var) and override it with a recording fake in tests. The 2-arg `IsCGMode(settingsPath, stderrSink)` signature is preserved (all 10 call sites are 2-arg).
- **R5 — URL allowlist over-restriction**: a too-strict `api.z.ai`-only allowlist would break legitimate user overrides. Mitigation: allow the canonical family AND any explicitly user-configured `llm.yaml` override that is a well-formed `https://` URL; the constraint is "well-formed https + not obviously hijacked", not "z.ai only".

## §C. Pre-flight

- [ ] Confirm `internal/cli` + `internal/tmux` baseline test suites are green (`go test ./internal/cli/... ./internal/tmux/...`).
- [ ] Confirm `GOOS=windows GOARCH=amd64 go build ./...` currently succeeds (to attribute any new break to REQ-CGH-001).
- [ ] Capture baseline coverage for the six cited functions (recorded in spec.md §A.1).
- [ ] Read the sibling SPEC `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` acceptance for the `IsCGMode` contract before touching `cg_detect.go`.

## §D. Constraints

Inherited from spec.md §C (C-1 subagent boundary, C-2 envkeys, C-3 settings helpers, C-4 cross-platform, C-5 no GLM integration tests in dev project, C-6 template neutrality, C-7 no sibling-SPEC regression).

Development mode: per `.moai/config/sections/quality.yaml development_mode`. The atomicity-helper refactor (M2) is brownfield → DDD characterization-first; new logic (URL validation, detector SSOT) is greenfield → TDD.

## §E. Self-Verification

Plan-phase audit-ready signal lives in `progress.md §E.1`. Run-phase (§E.2/§E.3) and sync-phase (§E.4) are owned by manager-develop / manager-docs respectively.

## §F. Milestones (priority-ordered, no time estimates)

- **M1 — Cross-platform launch safety (REQ-CGH-001)**. Split `launchClaudeDefault`'s exec into a POSIX `syscall.Exec` path and a Windows spawn-child-and-exit path. Prefer the `team_launch_posix.go`/`team_launch_windows.go` build-tag idiom OR an inline `runtime.GOOS == "windows"` guard mirroring `update.go:461`. Add `GOOS=windows` build to the verification batch. (TDD where unit-testable; build-check is the primary gate.)

- **M2 — CG settings atomicity + ordering (REQ-CGH-002, REQ-CGH-003, REQ-CGH-005)**. 
  1. DDD PRESERVE: characterization tests for `removeGLMEnv`, `syncPermissionModeToSettingsLocal`, `ensureSettingsLocalJSON` capturing current user-key-preservation behavior.
  2. Introduce a single locked + atomic (`temp-file + os.Rename`, mirroring `saveLLMSection`) settings helper in `internal/cli/settings.go`; wire flock from `internal/cli/team_spawn_lock_unix.go` (`lockFile`/`unlockFile`, `//go:build !windows`, with a `team_spawn_lock_windows.go` companion).
  3. Reorder `applyCGMode`: leader cleanup (strip GLM creds + set `teammateMode="tmux"`) as a SINGLE RMW at the TOP, before tmux injection. Collapse the `removeGLMEnv`(set `""`) + `ensureSettingsLocalJSON`(set `"tmux"`) double write.
  4. Verify characterization tests still pass (PRESERVE invariant).

- **M3 — Detector SSOT realignment (REQ-CGH-006)** [HEADLINE]. Make `IsCGMode` authoritative on `llm.yaml team_mode == "cg"` and/or the tmux SESSION env (`tmux show-environment`), not the leader process env. Reconcile `REQ-WTL-009` drift warning (keep it meaningful against the new signal). Add injection seam for the session-env reader; keep sibling-SPEC tests green.

- **M4 — tmux availability precondition (REQ-CGH-008)**. Add `Detector.IsAvailable()` check inside `applyCGMode` when `InTmuxSession()` is true; emit the clear "tmux not installed" error with install guidance.

- **M5 — GLM base_url validation (REQ-CGH-007)**. Add URL validation to the LLM config validation layer (`internal/config/validation.go` LLM section): well-formed `https://` + host allowlist (canonical `api.z.ai` family OR explicit user override). Reject with a clear error on failure. Ensure `DefaultGLMBaseURL` passes.

- **M6 — Test coverage of credential-routing invariants (REQ-CGH-009)**. Add a production-path test asserting leader-strips-GLM + teammate-gets-GLM, plus the `injectTmuxSessionEnv`↔`clearTmuxSessionEnv` key-parity test (excluding the documented `ANTHROPIC_AUTH_TOKEN` retention). Use the recording-fake session-manager seam.

- **M7 — Docs + regression close (REQ-CGH-004, REQ-CGH-010)**. Correct `CLAUDE.local.md §22.3` (`"tmux"`/`""`; disambiguate from `llm.yaml team_mode`). If the user-facing root `CLAUDE.md §15` CG description is also edited, sync `internal/template/templates/CLAUDE.md` + run `make build`. Run the full `internal/cli` + `internal/tmux` suites + `golangci-lint`.

## §G. Anti-Patterns to Avoid

- Deleting the `REQ-WTL-009` drift warning instead of reconciling it (violates C-7).
- Refactoring the atomicity helper before characterization tests exist (violates DDD PRESERVE; risks silent user-key loss).
- Adding an `api.z.ai`-only hard allowlist that breaks legitimate user overrides (R5).
- Editing root `CLAUDE.md §15` without the template-sync + `make build` branch (violates Template-First, REQ-CGH-004 / C-6).
- Inlining `os.Getenv("ANTHROPIC_*")` raw strings instead of `envkeys.go` constants (C-2).
- Running `moai cg` against the real dev project in tests (C-5).

## §H. Cross-References

- `internal/cli/launcher.go:143-204` (`applyCGMode`), `:209-268` (`removeGLMEnv`), `:541` (`syscall.Exec`), `:624-669` (`syncPermissionModeToSettingsLocal`)
- `internal/cli/glm.go:379-429` (`injectTmuxSessionEnv`), `:436-470` (`clearTmuxSessionEnv`), `:492-529` (`ensureSettingsLocalJSON`), `:680-698` (`saveLLMSection` atomic pattern), `:742-778` (`loadGLMConfig`)
- `internal/cli/update.go:461` (`reexecNewBinary` Windows pattern)
- `internal/cli/worktree/team_launch_posix.go` + `team_launch_windows.go` (build-tag split precedent)
- `internal/cli/team_spawn_lock_unix.go` (flock infra: `lockFile`/`unlockFile`, `syscall.Flock LOCK_EX/LOCK_UN`, `//go:build !windows`) + `internal/cli/team_spawn_lock_windows.go` (companion)
- `internal/tmux/cg_detect.go:41-87` (`IsCGMode`, `hasGLMEnv`), `internal/tmux/detector.go:60-84` (`IsAvailable`/`Version`)
- `internal/config/validation.go:349-352` (LLM section), `internal/config/defaults.go:41` (`DefaultGLMBaseURL`)
- Sibling SPEC: `.moai/specs/SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001/`
- `CLAUDE.local.md §2` (Template-First), §13 (GLM test isolation), §14 (hardcoding), §22.3 (`teammateMode` intent), §25 (template internal isolation)
- `internal/cli/CLAUDE.md`, `internal/config/CLAUDE.md` (module conventions)
