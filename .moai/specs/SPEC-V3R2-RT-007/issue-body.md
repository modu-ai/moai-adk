# SPEC-V3R2-RT-007 — Hardcoded Path Fix + Versioned Migration

> Companion GitHub Issue body for tracking the SPEC across plan / run / sync phases. Mirrors `spec.md`, `plan.md`, `acceptance.md` summaries and serves as the public-facing entry-point for collaborators.

## Summary

Two reinforcing halves bundled into a single SPEC:

1. **Path-fix half (affirm + harden + retroactive)** — Lock in regression prevention for the absolute-path anti-pattern in shell wrapper templates. The 28 hook wrapper templates under `internal/template/templates/.claude/hooks/moai/` are *already free* of the historical `/Users/goos/go/bin/moai` literal (verified 2026-05-10), but no CI lint exists today to prevent regression. This SPEC adds (a) a CI audit test rejecting any future absolute-path emission in `*.sh.tmpl`, (b) a CI audit test rejecting `.HomeDir` in fallback-chain context (CLAUDE.local.md §14 enforcement), (c) an affirm test that locks `$HOME` registration in `claudeCodePassthroughTokens`, and (d) consolidation of two duplicate `detectGoBinPath` implementations (initializer.go:286 + update.go:2515) into a single `internal/runtime/gobin/resolver.go::Detect` helper.

2. **Migration framework half (new construction)** — Build the versioned migration auto-apply infrastructure under `internal/migration/` (new package). Closes problem-catalog P-C06 ("no silent migration pattern"). Every session-start silently advances pending migrations under a compile-time static registry, with idempotency guards, atomic version-file updates, advisory-lock concurrency safety, JSONL structured logging, and `moai migration {run,status,rollback}` + `moai doctor --check migration` user surfaces. **Migration 1** retroactively rewrites the absolute-path literal in v2.x user projects' on-disk wrappers to `$HOME/go/bin/moai` — idempotent (no-op on already-clean projects) and non-rollback-able by design (CRITICAL bug-fix migrations decline rollback per REQ-024).

## Why this matters

- **P-H04 (CRITICAL)**: Per `master.md` §8 BC-V3R2-008, `/Users/goos/go/bin/moai` literal "breaks every user except 'goos' on every machine". Templates are clean; **user disks are not**. Migration 1 is the only mechanism to retroactively reach v2.x installs.
- **P-C06 (LEVERAGE)**: Without a migration framework, every future `.moai/` schema or asset shape change requires explicit user CLI invocation. Existing `moai migrate agency` is one-shot; v3 needs a generalizable platform.
- **X-5 pattern adoption** (per Claude Code reference, r3 §2 Decision 10): silent preAction migrations align moai with CC's `CURRENT_MIGRATION_VERSION = 11` model and reduce support burden.

## Breaking Change — BC-V3R2-008 (AUTO migration)

This SPEC carries the `breaking: true` flag and consumes BC ID **BC-V3R2-008**. The breaking-change classification covers:

### What changes for users

| Audience | Impact | Mitigation |
|----------|--------|------------|
| **Fresh v3.0 install user** | None observable. Templates already clean; m001 detects "0 rewrites needed" no-op. | Automatic — no action required. |
| **v2.x → v3.0 upgrade user** (most affected) | First session-start after upgrade silently rewrites `.claude/hooks/moai/handle-*.sh` containing the literal. ~5-10 wrapper files modified per project. Atomic; no risk of partial state. | Automatic via session-start hook. Manual fallback: `moai migration run`. Disable: `migrations.disabled: true` in `.moai/config/sections/system.yaml` (NOT recommended). |
| **CI / non-interactive environment** | If session-start hook does not fire (no Claude Code session), wrappers stay at v2.x state until the user opens an interactive session. | `moai migration run` invocation in CI scripts or migration-aware `moai update` post-step. |
| **Power user with custom wrapper modifications** | Migration replaces only the exact literal `/Users/goos/go/bin/moai`; other modifications preserved. | None needed; intentional design. Verify post-migration with `moai doctor --check migration`. |
| **Windows Git Bash / WSL / MSYS2 user** | Wrapper rewrite is a string-replace (platform-independent). `$HOME` is shell-expanded at runtime, identical across Bash environments. | None needed (REQ-V3R2-RT-007-060). |
| **Rollback requester** | `moai migration rollback 1` returns `MigrationNotRollbackable`. m001 is intentionally non-rollback-able (CRITICAL bug-fix). | Manual edit of wrapper files if regression to v2.x state is required (not supported). |

### Communication plan

1. **CHANGELOG entry** under `## [Unreleased] / ### Added` per `plan.md` §M5d.
2. **Release notes** for v3.0.0 explicitly call out: `migrations.disabled: true` opt-out flag, manual `moai migration run` for CI, and the non-rollback-able nature of m001.
3. **doctor surface**: `moai doctor --check migration` shows applied migration history and current version.
4. **Backward compatibility window**: BC-V3R2-008 commits to "AUTO migration" — no manual user action required for the typical case. v2.x installs receive the fix on first v3 session-start.

## Acceptance Criteria (16 ACs)

Full G/W/T scenarios in `acceptance.md`. Headline summary:

- **AC-01**: Generated wrappers contain no absolute user paths (CI lint).
- **AC-02**: Fallback chain order verified in 28 wrappers + status_line.
- **AC-03**: Migration 1 rewrites hardcoded literal in v2.x projects (idempotent).
- **AC-04**: Migration 1 no-op on already-clean projects.
- **AC-05**: Fresh install applies all registered migrations in order.
- **AC-06**: Failed migration: version-file unchanged, log entry, SystemMessage surfaced.
- **AC-07**: `moai doctor --check migration` reports current version + pending + last log.
- **AC-08**: `moai migration rollback <v>` succeeds when `Rollback` declared.
- **AC-09**: `moai migration rollback 1` returns `MigrationNotRollbackable` for m001.
- **AC-10**: CI lint catches `.HomeDir` in fallback-chain context.
- **AC-11**: Two migrations with same Version panic at init time.
- **AC-12**: Version-file ahead of registry: graceful no-op + warning.
- **AC-13**: `migrations.disabled: true` skips runner with no SystemMessage.
- **AC-14**: Crash mid-Apply: re-run completes idempotently.
- **AC-15**: Windows Git Bash rewrite works (platform-independent string-replace).
- **AC-16**: `moai migration run` manually applies pending migrations.

## Dependencies

### Blocked by

- **SPEC-V3R2-CON-001** (FROZEN-zone codification — migration framework is constitutional).
- **SPEC-V3R2-RT-001** (HookResponse with SystemMessage — migration error surfacing).
- **SPEC-V3R2-RT-006** (SessionStart handler completeness — RT-007 Apply call inserted in extended handler).

### Blocks

- **SPEC-V3R2-EXT-004** (versioned migration general framework).
- **SPEC-V3R2-MIG-001** (v2→v3 user migrator).
- **SPEC-V3R2-MIG-002** (hook registration cleanup).
- **SPEC-V3R2-MIG-003** (config loader retrofit).

## Implementation Plan

5 milestones M1-M5 with 40 tasks (T-RT007-01..40). See `plan.md` §2 and `tasks.md`. Critical path: 18 sequential gates from M1 RED → M5 closure.

- **M1 (P0)**: ~22 RED tests + 3 audit lint tests across 12 new test files.
- **M2 (P0)**: `gobin.Detect` helper + 2 call-site refactors (DRY consolidation) + lint tests GREEN.
- **M3 (P0)**: `internal/migration/{runner,registry,version,log}.go` core.
- **M4 (P0)**: m001 + session-start hook integration + `migrations.disabled` config.
- **M5 (P1)**: `moai migration {run,status,rollback}` CLI + `moai doctor --check migration` + 7 MX tags + CHANGELOG + final verification.

Estimated scope: ~1,530 LOC new + ~30 LOC modified across 20 new files + 6 modified files. No new direct module dependencies.

## Plan-time facts and verifications

- 28 hook wrapper templates (verified `ls internal/template/templates/.claude/hooks/moai/*.tmpl | wc -l`).
- 0 hardcoded `/Users/goos/go/bin/moai` literals in 28 wrappers (verified by grep).
- `internal/migration/` does not yet exist (new package per plan.md §3.2).
- Two duplicate `detectGoBinPath` implementations (initializer.go:286 + update.go:2515) — to be unified.
- 32 EARS REQs across 7 categories; 16 ACs all map to ≥1 REQ.
- 24 file:line anchors in plan.md; 29 in research.md; all verified via Grep/Read.

See `progress.md` "Plan-time evidence checks" for the verification table.

## Spec-vs-reality drift acknowledgement

`spec.md` (drafted 2026-04-23) cites "26 wrappers" (actual 28) and "all wrappers contain hardcoded literal" (actual 0 hits). The SPEC's *intent* — migration framework + retroactive v2.x fix + CI lint — remains valid. Spec text correction is recommended as a separate patch SPEC to avoid plan-audit drift rejection. See `research.md` §2.2 and `progress.md` "Spec drift acknowledgement".

## Risks and mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| RT-001 / RT-006 unmerged at run-phase entry | M | M | plan-auditor verifies; stub SystemMessage with `slog.Warn` until merge. |
| User overrides `migrations.disabled: true` and never upgrades | M | M | doctor surface + release notes. |
| m001 corrupts custom user modifications | L | H | exact-literal match only; preserves all other content. |
| Concurrent session-start race on version-file | L | M | advisory lock + 3-retry / 10ms backoff (RT-004 primitive reuse). |
| Cobra command group naming clash (`migrate` vs `migration`) | L | L | cross-reference in `--help` text. |

Full risk table in `spec.md` §8 (9 risks) and `plan.md` §5 (15 file-anchored mitigations).

## Worktree mode

Plan branch: `plan/SPEC-V3R2-RT-007`
Worktree: `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007`
Base commit: `496595c3f` (origin/main HEAD at plan-time)
Run-phase branch (proposed): `feat/SPEC-V3R2-RT-007` per CLAUDE.local.md §18.2 + spec-workflow.md updated convention (Step 2: SPEC worktree, branch `feat/SPEC-XXX`).

## Out of scope (explicit exclusions)

Per `spec.md` §2:

- **Migration 2+** (v2→v3 SPEC agent rename, flat permission rewrite) — owned by SPEC-V3R2-MIG-001.
- **Session-start hook upgrade semantics** — SPEC-V3R2-RT-006.
- **Config loader for 5 missing yaml sections** — SPEC-V3R2-MIG-003.
- **Settings-file multi-layer merge** — SPEC-V3R2-RT-005.
- **Constitutional FROZEN zone codification** — SPEC-V3R2-CON-001.
- **`internal/template/templates/.claude/...` content changes** — none in this SPEC; only audit tests added in `internal/template/`.

## Tracker

- [ ] plan-auditor PASS (Phase 0.5)
- [ ] RT-001 dependency verified
- [ ] RT-006 dependency verified
- [ ] CON-001 dependency verified
- [ ] M1 RED tests added (~22 tests, 12 new files)
- [ ] M2 path-fix consolidation GREEN (AC-01, -02, -10)
- [ ] M3 migration core GREEN (AC-04, -05, -11, -12, -14)
- [ ] M4 m001 + session-start hook GREEN (AC-03, -06, -13, -15)
- [ ] M5 CLI + doctor + MX + CHANGELOG GREEN (AC-07, -08, -09, -16)
- [ ] `make build` clean (no embedded.go diff)
- [ ] `go test ./...` 0 failures
- [ ] `go vet ./...` + `golangci-lint run` 0 warnings
- [ ] PR opened with required CI checks green
- [ ] PR merged to main

---

End of issue-body.md.
