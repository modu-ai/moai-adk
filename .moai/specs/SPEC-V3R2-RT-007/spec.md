---
id: SPEC-V3R2-RT-007
title: "Hardcoded Path Fix + Versioned Migration"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 — Runtime Hardening"
module: "internal/template/ + internal/migration/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-001
  - SPEC-V3R2-RT-006
bc_id: [BC-V3R2-008]
related_principle: [P11 File-First Primitives, P12 Constitutional Governance]
related_pattern: [X-5, X-2]
related_problem: [P-H04, P-C06]
related_theme: "Layer 3: Runtime"
breaking: true
lifecycle: spec-anchored
tags: "path, hardcoded, migration, v3r2, breaking, runtime, shell-wrappers"
---

# SPEC-V3R2-RT-007: Hardcoded Path Fix + Versioned Migration

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft. New SPEC — no v3-legacy predecessor. Addresses CRITICAL path hardcoding P-H04 across all 26 shell wrappers, plus silent migration framework P-C06. |

---

## 1. Goal (목적)

Eliminate the CRITICAL hardcoded absolute path `/Users/goos/go/bin/moai` present in all 26 shell hook wrappers at `.claude/hooks/moai/handle-*.sh` (r6 §2.1 finding: "Breaks for every user except `goos` on every machine") and simultaneously ship the versioned migration framework (pattern X-5) that retrofits existing installs. The current fallback chain `moai in PATH → /Users/goos/go/bin/moai → $HOME/go/bin/moai → silent exit` orders the hardcoded value before the portable environment variable — this SPEC inverts that order and also regenerates the templates so new installs never see the hardcoded literal.

The migration framework half of this SPEC is pattern X-5 from pattern-library.md (CC `CURRENT_MIGRATION_VERSION = 11` adoption candidate, r3 §2 Decision 10). moai's current `moai migrate agency` is explicit and users must remember to run it; v3 makes migrations silent via session-start hook preAction, idempotent via version guards, and observable via `moai doctor migration-status`. This closes problem-catalog.md P-C06 ("No silent migration pattern") and becomes the mechanism by which P-H04 retroactively reaches v2.x installs upgrading to v3.

Master §8 BC-V3R2-008 commits to AUTO migration: `make build` regenerates embedded templates via the updated GoBinPath resolver; existing user shells receive the fix on next `moai update`. The migration framework SPEC-V3R2-EXT-004 consumes the infrastructure declared here and adds a registry of concrete migrations.

## 2. Scope (범위)

In-scope:

**Path fix half:**

- `internal/template/render.go` GoBinPath resolver: detect `go env GOBIN` → `go env GOPATH/bin` → `$HOME/go/bin` → compile-time fallback. Never emit absolute user paths.
- 26 shell wrapper template files at `internal/template/templates/.claude/hooks/moai/handle-*.sh` rewritten to use `$HOME/go/bin/moai` as primary fallback (second in the chain after PATH lookup).
- New fallback chain: `moai in PATH → $HOME/go/bin/moai → {{posixPath .GoBinPath}}/moai → silent exit`. `.GoBinPath` uses `posixPath` template helper per CLAUDE.local.md §14 ".sh.tmpl" rule (`.HomeDir` is forbidden in fallback paths; only `$HOME` allowed).
- `make build` regenerates all embedded shell templates through the updated resolver.
- Migration step applies the fix retroactively to existing `.claude/hooks/moai/*.sh` in user projects when migration 1 runs.

**Migration framework half:**

- `internal/migration/runner.go`: `Migration` struct with `Version int`, `Name string`, `Apply func(projectRoot string) error`.
- Idempotent `Apply` contract: every migration SHALL be safe to re-run; version guard file `.moai/state/migration-version` tracks `CURRENT_MIGRATION_VERSION`.
- `MigrationRunner.Apply(current int) error` walks registry, applies pending migrations in order, updates version file atomically on each success.
- Session-start hook (SPEC-V3R2-RT-006 SessionStart handler) invokes `MigrationRunner.Apply` silently; errors surface via `HookResponse.SystemMessage` but do not block the session.
- `moai doctor migration-status` prints current version, pending migrations, last-applied migration's timestamp and log.
- Migration 1 (this SPEC) fixes hardcoded path in existing user shell wrappers; migration 2+ ship in SPEC-V3R2-EXT-004.
- Migration rollback: each migration declares an optional `Rollback(projectRoot string) error` function. Only invoked manually via `moai migration rollback <version>`; never auto-invoked.
- Migration log: every applied migration writes a structured entry to `.moai/logs/migrations.log` with `{version, name, started, completed, result: success|failed|rolled-back, details}`.

Out-of-scope (addressed by other SPECs):

- Migration 2+ (v2→v3 SPEC agent rename, flat permission-allow rewrite, etc.) — SPEC-V3R2-MIG-001.
- Session-start hook upgrade semantics — SPEC-V3R2-RT-006.
- Config loader additions for 5 missing yaml sections — SPEC-V3R2-MIG-003.
- Settings-file multi-layer merge — SPEC-V3R2-RT-005.
- Constitutional FROZEN zone codification — SPEC-V3R2-CON-001.

## 3. Environment (환경)

Current moai-adk state per r6-commands-hooks-style-rules.md §2.1:

- 26 shell wrapper scripts at `.claude/hooks/moai/handle-*.sh` all use fallback chain: `moai` in PATH → `/Users/goos/go/bin/moai` → `$HOME/go/bin/moai` → silent exit.
- The absolute path `/Users/goos/go/bin/moai` is hardcoded — a `.HomeDir`-style anti-pattern per CLAUDE.local.md §14.
- Templates are embedded via `go:embed` in `internal/template/templates/`; regeneration requires `make build`.
- No silent migration mechanism exists per problem-catalog.md P-C06: "moai migrate is explicit CLI command users must remember."
- Existing `moai migrate agency` exists as a one-shot command; v3 generalizes this into a framework.

CLAUDE.local.md constraints:

- §14 [HARD] `.HomeDir` forbidden in fallback paths; `$HOME` required.
- §14 Primary: `{{posixPath .GoBinPath}}/moai` (OK at init-time).
- §14 Fallback: `$HOME/go/bin/moai` (MUST use `$HOME`).
- §14 `renderer.go`: `$HOME` is already registered in `claudeCodePassthroughTokens`.

Claude Code reference (X-5 pattern):

- r3 §2 Decision 10: "Versioned migrations with preAction auto-apply. `CURRENT_MIGRATION_VERSION = 11` + preAction hook + per-migration idempotency guards." Silently rolls forward every user.
- pattern-library.md X-5: "Versioned migration auto-apply at preAction. ADOPT — moai's current `moai migrate agency` is explicit; silent preAction align with CC and reduce support burden."

Affected modules:

- `internal/template/render.go` — GoBinPath resolver.
- `internal/template/templates/.claude/hooks/moai/handle-*.sh` — 26 template files rewritten.
- `internal/migration/runner.go` — new file, Migration + MigrationRunner.
- `internal/migration/migrations/m001_hardcoded_path.go` — new file, migration 1.
- `internal/migration/registry.go` — new file, migration registry.
- `internal/hook/session_start.go` — add MigrationRunner invocation.
- `internal/cli/doctor.go` — `migration-status` sub-subcommand.
- `internal/cli/migration.go` — new file, `moai migration` commands (rollback).
- `.moai/state/migration-version` — new file, version guard (at user project root).
- `.moai/logs/migrations.log` — new append-only log.

## 4. Assumptions (가정)

- Shell wrappers currently parse via POSIX `sh` compatibility (bash, zsh, dash all supported); `$HOME` is universally available.
- Migration 1 idempotency: detecting the hardcoded `/Users/goos/go/bin/moai` literal in a shell wrapper is the trigger; absence means migration already applied.
- `make build` is run at release time so new installs never see the literal; existing installs get the fix via migration 1.
- Session-start migration invocation adds under 50 ms p99 on no-op (version already current) and under 500 ms on single-migration apply (typical migration 1 rewrite of 26 files).
- Migration registry is compile-time; adding a migration requires code change + `make build` + release.
- `os.Rename` provides atomic version-file updates on all supported platforms.
- Rollback semantics are opt-in; not every migration declares a Rollback (CRITICAL bug-fix migrations may be non-rollback-able).
- `.moai/state/migration-version` is excluded from template sync (runtime-managed) per CLAUDE.local.md §2 protected directories.
- Migration apply errors surface via HookResponse.SystemMessage but do not block session — a CRITICAL-bug migration failure is itself logged and retried next session.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements — Path Fix

- REQ-V3R2-RT-007-001: The `internal/template/render.go` GoBinPath resolver SHALL detect the user's Go bin directory via `go env GOBIN` → `go env GOPATH/bin` → `$HOME/go/bin` fallback chain.
- REQ-V3R2-RT-007-002: Generated shell wrapper scripts SHALL NOT contain absolute user paths (no `/Users/{username}/`, no `/home/{username}/` literals).
- REQ-V3R2-RT-007-003: The fallback chain in every generated wrapper SHALL be `moai` in PATH → `$HOME/go/bin/moai` → `{{posixPath .GoBinPath}}/moai` → silent exit.
- REQ-V3R2-RT-007-004: `.HomeDir` template variable SHALL be forbidden in fallback path context; CI lint SHALL reject templates using `{{posixPath .HomeDir}}` in fallback chains.
- REQ-V3R2-RT-007-005: The 26 shell wrapper template files SHALL be regenerated at `make build` time via the updated resolver.
- REQ-V3R2-RT-007-006: `$HOME` environment variable SHALL be registered in `claudeCodePassthroughTokens` per CLAUDE.local.md §14 (already present; this SPEC affirms and validates).

### 5.2 Ubiquitous Requirements — Migration Framework

- REQ-V3R2-RT-007-010: The `Migration` struct SHALL expose `Version int`, `Name string`, `Apply func(projectRoot string) error`, optional `Rollback func(projectRoot string) error`.
- REQ-V3R2-RT-007-011: Every migration SHALL be idempotent: re-applying a migration whose effect is already present SHALL NOT produce an error or duplicate work.
- REQ-V3R2-RT-007-012: `MigrationRunner.Apply(current int)` SHALL walk the registry in ascending version order and apply only migrations whose `Version > current`.
- REQ-V3R2-RT-007-013: On each successful Apply, `.moai/state/migration-version` SHALL be updated atomically (write to `*.tmp`, `os.Rename` to final) with the new version.
- REQ-V3R2-RT-007-014: Every applied migration SHALL write a structured entry to `.moai/logs/migrations.log` with version, name, started-at, completed-at, result, and details.
- REQ-V3R2-RT-007-015: `moai doctor migration-status` SHALL print current version, list of pending migrations, and last-applied migration entry from the log.
- REQ-V3R2-RT-007-016: The migration registry SHALL be compile-time static; no dynamic migration loading from disk.

### 5.3 Event-Driven Requirements

- REQ-V3R2-RT-007-020: WHEN the session-start hook fires, the handler SHALL invoke `MigrationRunner.Apply(current_version)` before any user-facing output.
- REQ-V3R2-RT-007-021: WHEN a migration fails with a non-nil error, the runner SHALL NOT advance the version file, SHALL write the failure to `.moai/logs/migrations.log`, and SHALL emit `HookResponse.SystemMessage` naming the migration number and error.
- REQ-V3R2-RT-007-022: WHEN migration 1 runs on a user project whose `.claude/hooks/moai/handle-*.sh` contains the literal `/Users/goos/go/bin/moai`, the migration SHALL rewrite the file replacing the literal with `$HOME/go/bin/moai`, preserving all other content.
- REQ-V3R2-RT-007-023: WHEN migration 1 runs on a project where shell wrappers do NOT contain the hardcoded literal (already migrated or fresh install), the migration SHALL treat it as a no-op and write a log entry `result: success, details: "already migrated"`.
- REQ-V3R2-RT-007-024: WHEN `moai migration rollback <version>` is invoked, the system SHALL call the migration's `Rollback` function if declared, update the version file to `version-1`, and log the rollback; if no Rollback is declared, return error `MigrationNotRollbackable`.
- REQ-V3R2-RT-007-025: WHEN `make build` runs, the template regenerator SHALL emit the shell wrappers with the updated fallback chain; generated files SHALL contain no hardcoded user paths.

### 5.4 State-Driven Requirements

- REQ-V3R2-RT-007-030: WHILE `.moai/state/migration-version` is absent (fresh v3 install or v2→v3 transition), `MigrationRunner` SHALL treat the current version as 0 and apply all registered migrations in order.
- REQ-V3R2-RT-007-031: WHILE a migration is mid-apply (in-flight), the version file SHALL NOT yet reflect the new number; crash-recovery on next session detects in-flight state via presence of `migration-version.tmp` file and re-applies the migration (idempotency protects from partial writes).
- REQ-V3R2-RT-007-032: WHILE `.moai/config/sections/system.yaml` has `migrations.disabled: true`, the session-start migration invocation SHALL be skipped with a log entry but no SystemMessage.

### 5.5 Optional Features

- REQ-V3R2-RT-007-040: WHERE `moai migration run` is invoked manually, the runner SHALL apply pending migrations regardless of the session-start trigger.
- REQ-V3R2-RT-007-041: WHERE `moai migration status --json` is invoked, the output SHALL be machine-readable JSON containing `{current_version, registered_versions, pending_versions, last_applied}`.
- REQ-V3R2-RT-007-042: WHERE a migration declares `Rollback`, `moai migration rollback <version>` SHALL succeed; `MigrationNotRollbackable` is returned only for migrations without declared Rollback.

### 5.6 Unwanted Behavior

- REQ-V3R2-RT-007-050: IF the `posixPath` template helper is invoked with `.HomeDir` in a fallback chain context, THEN the CI lint in `internal/template/lint_test.go` SHALL fail the build naming the template file and line.
- REQ-V3R2-RT-007-051: IF generated shell wrappers are found to contain `/Users/*/go/bin/moai` or `/home/*/go/bin/moai` literals (absolute user paths), THEN the CI lint SHALL fail the build naming the file.
- REQ-V3R2-RT-007-052: IF migration 1 Apply fails on a read-only `.claude/hooks/moai/` directory, THEN the error SHALL be logged as `MigrationReadOnly` and the SystemMessage SHALL instruct the user to check permissions; the version file SHALL NOT advance.
- REQ-V3R2-RT-007-053: IF two migrations declare the same `Version int`, THEN `go test ./internal/migration/...` SHALL fail at init time with error `DuplicateMigrationVersion`.
- REQ-V3R2-RT-007-054: IF `.moai/state/migration-version` contains a version greater than any registered migration version (user downgraded moai), THEN the runner SHALL emit `SystemMessage: "Version file ahead of known migrations; treating as no-op"` and not attempt to apply anything.

### 5.7 Complex Requirements

- REQ-V3R2-RT-007-060: WHILE the user is on Windows without WSL, WHEN migration 1 runs, THEN the migration SHALL still rewrite `$HOME/go/bin/moai` in `*.sh` files (these are executed by Git Bash or the Claude Code shell shim); `$HOME` is expanded by the shell at runtime, not the migration.
- REQ-V3R2-RT-007-061: WHILE a migration both modifies templates AND emits a new artifact file, WHEN Apply is interrupted mid-write, THEN on next session the idempotent check detects the partial state and completes the remaining work without duplicating already-applied parts.

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-007-01: Given `make build` runs on any developer machine, When generated shell wrappers are inspected, Then none contains `/Users/goos/go/bin/moai` or any other absolute user path. (maps REQ-V3R2-RT-007-002, -025, -051)
- AC-V3R2-RT-007-02: Given the fallback chain in a generated wrapper, When grep runs for path patterns, Then the chain order is `moai` in PATH → `$HOME/go/bin/moai` → `{{posixPath .GoBinPath}}/moai`. (maps REQ-V3R2-RT-007-003)
- AC-V3R2-RT-007-03: Given a v2.x user project with hardcoded `/Users/goos/go/bin/moai` in `.claude/hooks/moai/handle-session-start.sh`, When migration 1 runs at session-start, Then the file is rewritten with `$HOME/go/bin/moai` preserving all other content. (maps REQ-V3R2-RT-007-022)
- AC-V3R2-RT-007-04: Given migration 1 has already been applied (no hardcoded literal), When re-run, Then it logs `"already migrated"` and succeeds as no-op. (maps REQ-V3R2-RT-007-011, -023)
- AC-V3R2-RT-007-05: Given `.moai/state/migration-version` is absent, When session-start hook fires, Then all registered migrations are applied in ascending order. (maps REQ-V3R2-RT-007-030)
- AC-V3R2-RT-007-06: Given a migration fails with error, When runner completes, Then version file is unchanged and `.moai/logs/migrations.log` contains the failure entry with error text. (maps REQ-V3R2-RT-007-021, -014)
- AC-V3R2-RT-007-07: Given `moai doctor migration-status`, When invoked, Then stdout shows current version, pending versions, and last-applied log entry. (maps REQ-V3R2-RT-007-015)
- AC-V3R2-RT-007-08: Given `moai migration rollback 1` is invoked, When migration 1 declares Rollback, Then rollback runs, version file goes to 0, and log entry records the rollback. (maps REQ-V3R2-RT-007-024, -042)
- AC-V3R2-RT-007-09: Given migration 1 does NOT declare Rollback, When `moai migration rollback 1` is invoked, Then error `MigrationNotRollbackable` is returned. (maps REQ-V3R2-RT-007-024)
- AC-V3R2-RT-007-10: Given a template author adds `{{posixPath .HomeDir}}/go/bin/moai` to a fallback chain, When `go test ./internal/template/... -run TestLint` runs, Then the test fails naming the offending template and line. (maps REQ-V3R2-RT-007-050)
- AC-V3R2-RT-007-11: Given two migrations with the same Version, When `go test` init runs, Then panic with `DuplicateMigrationVersion`. (maps REQ-V3R2-RT-007-053)
- AC-V3R2-RT-007-12: Given `.moai/state/migration-version` contains version 99, When runner starts and registry max version is 5, Then runner logs "Version file ahead" and no migrations apply. (maps REQ-V3R2-RT-007-054)
- AC-V3R2-RT-007-13: Given `system.yaml` has `migrations.disabled: true`, When session-start fires, Then migration runner is skipped and no SystemMessage appears. (maps REQ-V3R2-RT-007-032)
- AC-V3R2-RT-007-14: Given migration Apply writes `migration-version.tmp` then crashes before `os.Rename`, When next session starts, Then runner detects in-flight state, re-applies the migration (idempotent), and completes the rename. (maps REQ-V3R2-RT-007-031, -061)
- AC-V3R2-RT-007-15: Given Windows Git Bash environment, When migration 1 runs on a user project, Then shell wrappers are rewritten with `$HOME/go/bin/moai` (shell expands at runtime). (maps REQ-V3R2-RT-007-060)
- AC-V3R2-RT-007-16: Given `moai migration run` is invoked manually on a project with no session active, When executed, Then pending migrations apply. (maps REQ-V3R2-RT-007-040)

## 7. Constraints (제약)

- Technical: Go 1.22+; no new external dependencies. Atomic file rename via `os.Rename`; advisory lock via `golang.org/x/sys` for the version file.
- Backward compat: v2.x users upgrading to v3.0 receive migration 1 silently at first session-start per master §8 BC-V3R2-008. The migration framework itself is additive — v2 installs had no migration-version file, which is interpreted as version 0.
- Platform: macOS / Linux / Windows (Git Bash or WSL for shell wrappers). Shell expansion of `$HOME` happens at runtime, not at migration time.
- Performance: Session-start migration invocation MUST complete in under 50 ms p99 on no-op and under 500 ms p99 on single-migration apply (typical migration 1 rewrite of 26 files).
- Idempotency: REQ-V3R2-RT-007-011 is non-negotiable; every migration MUST be safe to re-run. Test harness verifies idempotency by running each migration twice and comparing results.
- Template lint: REQ-V3R2-RT-007-050/051 enforce via CI; the project already has `internal/template/commands_audit_test.go` pattern to extend.
- Observable failure: migration failures never silently halt the session; SystemMessage is mandatory (REQ-V3R2-RT-007-021).
- $HOME registration: `renderer.go`'s `claudeCodePassthroughTokens` already includes `$HOME` per CLAUDE.local.md §14 — this SPEC affirms without duplication.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Users ignore the session-start warning when migration 1 fails, leaving hardcoded path indefinitely | M | H | Retry on every session until success; `moai doctor migration-status` surfaces pending state; manual `moai migration run` is always available. |
| Shell wrapper rewrite corrupts a custom user modification | L | H | Migration 1 only replaces the exact literal string; if the user hand-edited their wrappers, their changes are preserved. |
| Idempotency bug in a future migration advances the version file but leaves work incomplete | M | M | REQ-V3R2-RT-007-011 enforced via test harness running each migration twice; CI gates this. |
| Concurrent sessions race on migration-version file | L | M | Advisory lock during Apply; atomic rename; REQ-V3R2-RT-007-031 handles crash recovery. |
| Windows path handling differs between WSL, Git Bash, MSYS2 | M | L | `$HOME` is shell-expanded, not Go-expanded; all three shells handle `$HOME` identically. |
| Rollback of a CRITICAL bug-fix migration reintroduces the bug | H | H | Migration 1 does not declare Rollback (bug-fix migrations SHOULD be non-rollback-able); if rollback is ever added, documentation explicitly names the risk. |
| User disables migrations via system.yaml and never upgrades | M | M | `moai doctor` surfaces pending state; release notes call out the flag. |
| `go env GOBIN` may be empty on some machines, falling through to GOPATH detection | L | L | Multi-step fallback chain in REQ-V3R2-RT-007-001 handles all three cases. |
| Hardcoded path is actually legitimate for dev-local dogfooding (moai project itself) | L | L | CLAUDE.local.md explicitly permits hardcoded paths in dev-local `.claude/settings.local.json` + `settings.local.json`; this SPEC scopes to template + user-project wrappers, not dev-local overrides. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-CON-001 (FROZEN-zone codification — the migration framework is a constitutional mechanism).
- SPEC-V3R2-RT-001 (HookResponse for SystemMessage surfacing of migration results).
- SPEC-V3R2-RT-006 (SessionStart handler consumes MigrationRunner.Apply invocation).

### 9.2 Blocks

- SPEC-V3R2-EXT-004 (versioned migration auto-apply general framework — this SPEC provides the base, EXT-004 adds the migration catalog for v2→v3).
- SPEC-V3R2-MIG-001 (v2→v3 user migrator tool consumes the MigrationRunner pattern for SPEC agent renames, flat permission rewrites, etc.).
- SPEC-V3R2-MIG-002 (hook registration cleanup uses migration framework to retire `notification`, `elicitation`, `elicitationResult`, `taskCreated` from user settings.json).
- SPEC-V3R2-MIG-003 (config loader addition may use migration to retrofit missing `.moai/config/sections/` files).

### 9.3 Related

- SPEC-V3R2-RT-005 (migration can update settings files; provenance tags apply to migrated values).
- SPEC-V3R2-RT-004 (SessionStore's migration-version file lives in `.moai/state/`).
- SPEC-V3R2-CON-003 (consolidation pass may emit migration to move misfiled rules — see P-H10 lsp-client.md move).
- SPEC-V3R2-WF-001 (skill consolidation 48→24 ships migrations in SPEC-V3R2-MIG-001).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §8 BC-V3R2-008; §7.3 hook audit (subagent-stop pairing); CLAUDE.local.md §14 hardcoding rules.
- Principle: P11 (File-First Primitives — version-file on disk); P12 (Constitutional Governance — silent migrations under explicit rule set).
- Pattern: X-5 (Versioned Migration Auto-Apply); X-2 (Multi-Layer Settings — migrations work across tiers).
- Problem: P-H04 (hardcoded absolute path in 26 shell wrappers, CRITICAL); P-C06 (no silent migration pattern, LOW-but-leverage).
- Master Appendix A: Principle P11 → secondary SPEC-V3R2-RT-007.
- Master Appendix C: Pattern X-5 → primary SPEC-V3R2-RT-007 + SPEC-V3R2-EXT-004.
- Wave 1 sources: r6-commands-hooks-style-rules.md §2.1 (hardcoded path finding); r3-cc-architecture-reread.md §2 Decision 10 (versioned migrations).
- Wave 2 sources: problem-catalog.md P-H04 (CRITICAL), P-C06 (LOW but adoption); pattern-library.md X-5 (ADOPT).
- BC-ID: BC-V3R2-008 (hardcoded path removed from 26 shell wrappers, AUTO via `make build` regeneration + migration 1).
- Priority: P0 Critical — P-H04 is classified CRITICAL in problem-catalog.md (breaks every user except `goos`). The migration framework half is strategic but the path fix is urgent.
