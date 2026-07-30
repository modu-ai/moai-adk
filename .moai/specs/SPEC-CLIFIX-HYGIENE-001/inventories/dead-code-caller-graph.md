# SPEC-CLIFIX-HYGIENE-001 M1 — Dead-Code Caller-Graph Inventory (ANALYZE)

> Frozen by manager-develop M1 (ANALYZE) on 2026-07-30 against worktree HEAD
> `359e887b9`. This file is the FOUNDATION that determines what M5 can safely
> delete. M1 deletes NOTHING — per plan.md §D, deletion is M5 scope.
>
> Method per plan.md §D + moai-workflow-ddd: grep + per-symbol caller-graph +
> reflection/registration/tag-gated-caller check + PRESERVE-list cross-check.

## Verdict legend

- **DEAD (M5-deletable)** — zero production callers; test-only callers do NOT count.
- **DEFERRED** — explicitly tagged for a follow-up SPEC; do NOT delete in M5 without resolving the tag.
- **LIVE (PRESERVE-KEEP)** — has a production caller; MUST NOT be deleted.

## Per-symbol verdicts

### 1. `buildGLMEnvVars` (`internal/cli/glm.go:917`)

- **Verdict: DEAD (M5-deletable)**
- Definition: `internal/cli/glm.go:917` `func buildGLMEnvVars(glmConfig *GLMConfigFromYAML, apiKey string) map[string]string`
- Production (non-test) callers: **0**
  - grep `buildGLMEnvVars internal/ cmd/ pkg/ --include='*.go'` excluding `_test.go` and the definition lines → empty result.
- Test-only callers (do NOT block deletion; tests get deleted alongside in M5):
  - `internal/cli/glm_team_test.go:144,148,154,161,163`
  - `internal/cli/coverage_improvement_test.go:1303,5840,5850`
- Tag-gated / reflection / registration check: none observed (no `//go:build` gate, no `reflect` call, no init/registry wiring).
- PRESERVE cross-check: not in the P0-wired lock path; not `scanDeprecatedPaths`/`cleanup_old_backups`; not exported with external callers (it IS exported, but `internal/cli` is an internal package — no external import path).
- M5 action: delete `buildGLMEnvVars` + its test callers in `glm_team_test.go` + `coverage_improvement_test.go`. Re-run `go build ./...` and `go test ./internal/cli/...` to confirm.

### 2. `ttyConfirmer` (`internal/cli/branch_protection.go:39-55`)

- **Verdict: DEFERRED — SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred-pairing**
- Definition: `internal/cli/branch_protection.go:39` `type ttyConfirmer struct{}` + method `Confirm` at `:43`.
- Production (non-test) callers: **0**
  - grep excluding `_test.go` and `branch_protection.go` itself → empty.
- Self-attested deferred status (verbatim from the source):
  - `branch_protection.go:42` `// nolint:unused // SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (paired with ttyConfirmer)`
  - `branch_protection.go:43` `// nolint:unused // SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (interactive prompt path)`
  - header note: "Currently unwired (CLI path uses yesConfirmer + --yes-branch-protection flag); kept for follow-up interactive prompt SPEC."
- The live production confirmer is `yesConfirmer` (`branch_protection.go` ~line 56); `ttyConfirmer` is the interactive-prompt counterpart that was never wired.
- M5 action: **DO NOT delete unconditionally.** M5 must first resolve the SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred-pairing decision (confirm with the orchestrator whether the pairing SPEC has landed or `ttyConfirmer` is still reserved). Only if the pairing is resolved as "delete" does M5 remove the type + method + their `nolint:unused` directives.

### 3. `worktree_validation.go` (whole-file candidate, `internal/cli/worktree_validation.go`)

- **Verdict: DEAD in production — but DEFERRED-pending a wire-up decision (self-attested follow-up SPEC)**
- Symbols defined: `worktreeReturn` (struct), `WorktreePathInvalidError` (exported type), `ErrWorktreePathInvalid` (exported var), `validateWorktreeReturn` (func).
- Production (non-test) callers across `internal/ cmd/ pkg/`: **0**
  - grep for `ErrWorktreePathInvalid | WorktreePathInvalidError | validateWorktreeReturn | worktreeReturn` excluding the defining file → matches ONLY in `internal/cli/launcher_worktree_validation_test.go` (test file).
  - NOTE: `internal/core/quality/worktree_validator_test.go` references `ValidateWorktreePath` — a DIFFERENT function in a DIFFERENT package (`internal/core/quality`), NOT a caller of anything in `internal/cli/worktree_validation.go`.
- Self-attested deferred status (verbatim from the file header + the `@MX:NOTE` on `validateWorktreeReturn`):
  - file header: "currently has no direct callsites (plan-stage expectation: 3-5 callsites → actual grep result: 0, only behavior validated in tests. actual wire-up to be performed in follow-up SPEC."
  - `@MX:NOTE` on `validateWorktreeReturn`: "current callsite: none (standalone helper); wire-up planned in follow-up SPEC."
- PRESERVE cross-check: the symbols are exported, but `internal/cli` is an internal package — no external import path exists. The exported-ness does not save them from being dead-in-production.
- M5 action: **record as a strong M5 candidate, but surface the wire-up-planned status to the orchestrator before deleting.** M5 may delete the whole file + `launcher_worktree_validation_test.go` (which only exercises the dead helper), contingent on confirming the "follow-up SPEC" was not secretly depending on the file existing. Per plan.md §D anti-pattern: "deleting 'dead' code by grep absence alone — reflection/registration/tag-gated callers must be checked" — checked, none found.

## PRESERVE-KEEP (confirmed LIVE — MUST NOT be deleted in M5)

These were flagged in the prompt's PRESERVE list and are re-verified LIVE here:

| Symbol | Definition | Live production caller(s) | Status |
|---|---|---|---|
| `acquireUpdateLock` | `internal/cli/update_cleanup.go:55` | `internal/cli/update.go:264` (`releaseLock, lockErr := acquireUpdateLock(lockRoot)`) — P0-wired by SPEC-CLIFIX-CRITICAL-001 | LIVE — P0-wired |
| `cleanStaleLock` | `internal/cli/update_cleanup.go:92` | `internal/cli/update_cleanup.go:59` (called from inside `acquireUpdateLock`) — P0-wired by SPEC-CLIFIX-CRITICAL-001 | LIVE — P0-wired |
| `scanDeprecatedPaths` | `internal/cli/update_cleanup.go:126` | `internal/cli/update_clean_install.go:189,257,265,278` (4 call sites) | LIVE |
| `cleanup_old_backups` | (audit-flagged) | (per prompt: confirmed-LIVE; re-verify exact callers at M5) | LIVE (re-verify at M5) |

## Summary counts

- DEAD (M5-deletable now): **1** (`buildGLMEnvVars`)
- DEFERRED (resolve tag, then decide): **2** (`ttyConfirmer`, `worktree_validation.go` whole file)
- LIVE (PRESERVE-KEEP): **4+** (the lock path + scanDeprecatedPaths + cleanup_old_backups)

## Realistic net-LOC delta for M5

- `buildGLMEnvVars` definition (~30 lines incl. comment) + its test callers (~30-40 lines across two test files) ≈ **−60 to −70 lines**.
- `ttyConfiermer` (~10 lines) — contingent on the §D.1 pairing resolution.
- `worktree_validation.go` whole file (~100 lines) + its test file (~150 lines) ≈ **−250 lines**, contingent on confirming the wire-up-planned follow-up SPEC status.

This walks back the original audit's ~500-line figure honestly (per acceptance.md §D.5): realistic delta is on the order of **−60 to −320 lines**, contingent on the two DEFERRED resolutions.
