# SPEC-GLM-KEY-INPUT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-29

## §E.2 Run-phase Evidence

Milestone evidence (M1–M6). Each AC in `acceptance.md` maps to a test below.

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-GKI-001-01 | PASS | `go test ./internal/web/ -run TestGLMKeyField_Renders -count=1` | `--- PASS: TestGLMKeyField_Renders` |
| AC-GKI-001-02 | PASS | `go test ./internal/web/ -run TestGLMKeyField_SecretClass -count=1` | `--- PASS: TestGLMKeyField_SecretClass` |
| AC-GKI-001-03 | PASS | `go test ./internal/web/ -run TestGLMKeySave_Persists -count=1` | `--- PASS: TestGLMKeySave_Persists` |
| AC-GKI-001-04 | PASS | `go test ./internal/web/ -run TestGLMKeySave_BannerNoEcho -count=1` | `--- PASS: TestGLMKeySave_BannerNoEcho` |
| AC-GKI-002-01 | PASS | `go test ./internal/web/ -run TestGLMKeySave_OnlyCredentialFileHoldsKey -count=1` | `--- PASS: TestGLMKeySave_OnlyCredentialFileHoldsKey` |
| AC-GKI-002-02 | PASS | `go test ./internal/glmcred/ -run TestSave_FileMode0600 -count=1` | `--- PASS: TestSave_FileMode0600` |
| AC-GKI-002-03 | PASS | structural — single writer. `grep -rn 'os.WriteFile.*env\.glm\|GLM_API_KEY=' internal/ --include='*.go' \| grep -v _test \| grep -v glmcred.go` returns only `internal/glmcred/glmcred.go` | one writer; cli delegates |
| AC-GKI-002-04 | PASS | `go test ./internal/web/ -run TestGLMKeySave_AtomicRejectLeavesCredentialUntouched -count=1` | `--- PASS` |
| AC-GKI-002-05 | PASS | `go test ./internal/web/ -run TestGLMKeySave_FailureSurfaced -count=1` | `--- PASS` |
| AC-GKI-003-01 | PASS | covered by TestGLMKeySave_OnlyCredentialFileHoldsKey (profile store tree scanned, no match) | `--- PASS` |
| AC-GKI-003-02 | PASS | covered by TestGLMKeySave_OnlyCredentialFileHoldsKey (project config sections scanned, no match) | `--- PASS` |
| AC-GKI-003-03 | PASS | `go test ./internal/web/ -run TestGLMKeyField_AbsentFromSchema -count=1` | `--- PASS` |
| AC-GKI-004-01 | PASS | `go test ./internal/web/ -run TestGLMKeySave_NoFullKeyInResponse -count=1` | `--- PASS` |
| AC-GKI-004-02 | PASS | `go test ./internal/web/ -run TestGLMKeyHint_TrailingFourOnly -count=1` | `--- PASS` (differential window scan) |
| AC-GKI-004-03 | PASS | covered by TestGLMKeySave_BannerNoEcho + TestGLMKeySave_FailureSurfaced (no key material in either response) | `--- PASS` |
| AC-GKI-004-04 | PASS | `go test ./internal/web/ -run TestGLMKeyHint_ShortKeyDisclosesNothing -count=1` | `--- PASS` (differential window) |
| AC-GKI-005-01 | PASS | `go test ./internal/web/ -run TestGLMKeySave_EmptyPreserves -count=1` | `--- PASS` (mtime unchanged) |
| AC-GKI-005-02 | PASS | `go test ./internal/web/ -run TestGLMKeySave_TrimsSurroundingWhitespace -count=1` | `--- PASS` |
| AC-GKI-005-03 | PASS | `go test ./internal/web/ -run 'TestValidateGLMKey/newline_in_body_rejected' -count=1` | `--- PASS` |
| AC-GKI-006-01 | PASS | `go test ./internal/glmcred/ -run TestSave_ReplaceInPlace -count=1` | `--- PASS` |
| AC-GKI-006-02 | PASS | `go test ./internal/glmcred/ -run TestSave_NarrowsExisting0644to0600 -count=1` + `go test ./internal/web/ -run TestGLMKeySave_NarrowsExistingWideMode -count=1` | `--- PASS` (both paths) |
| AC-GKI-006-03 | PASS | covered by TestGLMKeySave_EmptyPreserves (file still exists, content intact) | `--- PASS` |

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-07-29
- run_commit_sha: pending-backfill
- run_status: audit-ready
- ac_pass_count: 22
- ac_fail_count: 0
- new_warnings_or_lints_introduced: 0 (golangci-lint clean on internal/glmcred, internal/web, internal/cli)
- cross_platform_build.linux_amd64: pass
- cross_platform_build.windows_amd64: pass
- total_run_phase_files: 9 (glmcred.go, glmcred_test.go, cli/glm.go, web/glmkey.go, web/glmkey_test.go, web/handlers.go, web/fieldsets.templ, web/fieldsets_templ.go, web/assets/i18n.js)
- m1_to_mN_commit_strategy: single feat commit

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: sync-complete (manager-docs sync-phase, single sync commit 3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 / Status Transition Ownership Matrix)
- sync_complete_at: 2026-07-30
- sync_commit_sha: pending-backfill-SPEC-GLM-KEY-INPUT-001 (SHA cannot be in its own commit; backfilled in the follow-up backfill commit of this batch)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Added` — SPEC-GLM-KEY-INPUT-001 entry (22 ACs, Tier M, M1–M6 single feat commit)
- frontmatter_status_transitions: spec.md `in-progress → completed` atomic on this single sync commit; `updated: 2026-07-30` refreshed
- run_phase_pr: #1227 (merge commit 5832f0671 — GLM API key field in web console + shared credential writer)
- note: §25 template neutrality N/A — this SPEC touched `internal/glmcred` (new stdlib-only leaf pkg), `internal/cli`, `internal/web` Go code + templ; no template mirror edits. This sync commit carries only frontmatter + CHANGELOG + this §E.4 block (no template edits, no code changes)

## §F Phase 4 Mode Selection

- tier: M
- scope: 9 files (1 new Go pkg + 1 new web module + cli delegation + templ + i18n)
- domain count: 3 (glm credential pkg, web console, cli delegation)
- concurrency benefit: LOW (coding-heavy, brownfield integration)
- Decision: sub-agent (Mode 5) — single sequential implementer (this agent)
- Justification: coding-heavy brownfield integration touching a live CLI path (R-3); sequential sub-agent per Anthropic's coding-task parallelism caveat. The work is a single cohesive feature with tight coupling between the credential extraction and the console wiring — fanning out would create file-write conflicts on handlers.go and fieldsets.templ.
