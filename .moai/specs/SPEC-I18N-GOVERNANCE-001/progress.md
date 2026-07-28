# SPEC-I18N-GOVERNANCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-25
plan_auditor_verdict: PASS (epic autonomous approval — "미구현 15 SPEC 전부 completed")

## §E.2 Run-phase Evidence

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-I18NGOV-001 | PASS | `go test -run TestI18nUntranslatedValues ./internal/web/ -v` | `--- PASS: TestI18nUntranslatedValues (0.00s)` — detector green on shipped catalogue after 3 translation rulings + 23-entry allowlist |
| AC-I18NGOV-002 | PASS | `go test -run TestI18nUntranslatedDetectorNegativeControl ./internal/web/ -v` | `--- PASS` — synthetic cpl.placeholder: violation without allowlist, none with allowlist |
| AC-I18NGOV-003 | PASS | `go test -run TestI18nAllowlistNoOrphans ./internal/web/ -v` | `--- PASS` — every entry resolves to a real, still-identical key |
| AC-I18NGOV-004 | PASS | `go test -run TestI18nAllowlistShape ./internal/web/ -v` + `go vet ./internal/web/` | `--- PASS`; vet exit 0 (closed taxonomy compile-time-enforced) |
| AC-I18NGOV-005 | PASS | `go test -run TestI18nEndonymInvariants ./internal/web/ -v` | `--- PASS` — self-equality + exonym distinctness hold |
| AC-I18NGOV-006 | PASS | `go test -run TestI18nKeyCoverageForward ./internal/web/ -v` | `--- PASS` — en=340 keys, 0 missing from ko/ja/zh |
| AC-I18NGOV-007 | PASS | `go test -run TestI18nKeyCoverageReverse ./internal/web/ -v` | `--- PASS` — 10 agentdesc.* per locale match the registry |
| AC-I18NGOV-008 | PASS | `go test -run TestI18nParserSpecialKeys ./internal/web/ -v` | `--- PASS` — html+md + 3×[1m] keys parsed with non-empty values |
| AC-I18NGOV-009 | PASS | `go test -run TestI18nGovernanceContractPresent ./internal/web/ -v` + `grep -n "i18n_untranslated_allowlist" internal/web/assets/i18n.js` | `--- PASS`; grep → `21:// internal/web/i18n_untranslated_allowlist_test.go (SPEC-I18N-GOVERNANCE-001).` |
| AC-I18NGOV-010 | PASS | `go test ./internal/web/...` + `go test -run 'TestD3AgentDesc' ./internal/web/ -v` + `git diff --stat -- internal/web/webux_followup_test.go` | suite ok; both D3 tests PASS unmodified; webux_followup diff empty (C1 preserved) |
| AC-I18NGOV-011 | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + `git diff --name-only \| grep -v _test\.go$` | both builds exit 0; only non-test file changed = `internal/web/assets/i18n.js` |

### M1 rulings (judgment calls)

Every member of the re-measured identity set received an explicit ruling. The re-measurement at run-phase entry (en=340, ko/ja/zh=350; exact-identical set = 26 keys per locale) confirmed the plan-phase baseline modulo one drift: the per-locale count is 26, not 25 (the spec.md §A.1 baseline of 25 was stale by one — B1 flagged this).

- **`board.badge.mustfix`** = `"MUST-FIX"` → **allowlist, `technical-identifier`**. The literal is the `Severity` value emitted by `internal/spec/audit.go` and matched verbatim by `internal/web/board.go` and `internal/cli/spec_audit.go`; the badge must read the same token the CLI and its JSON output emit so a user grepping `moai spec audit --json` finds the badge. Translating would break that correspondence.
- **`mp.tier.empty`** = `"(runtime default: medium)"` → **translate (fix)**. No taxonomy reason applies (English prose). Its sibling `mp.tier.default` is translated in all three locales (`ko` = `(런타임 기본값: medium — llm.performance_tier 미설정)`, etc.), so the omission is an inconsistency, not a decision. Fixed to `(런타임 기본값: medium)` / `(ランタイムデフォルト: medium)` / `(运行时默认: medium)`, matching the sibling's locale pattern.

Two additional translation rulings (mechanical — no taxonomy reason applies):

- **`mp.col.effort`** — en `"Effort"`, all three non-en `"effort"` (case-only divergence, caught by normalization). Siblings `mp.col.tier/phase/model` are translated (`티어/ティア/层级`, etc.). Fixed to `에포트` / `エフォート` / `努力度` following the sibling transliteration pattern.
- **`f.model.desc`** — `"The specific model to run."` verbatim in all three locales. All other `*.desc` fields are translated (`f.report.format.desc`, `agentfm.tier.desc`). Fixed to `실행할 특정 모델.` / `実行する特定のモデル。` / `要运行的具体模型。`.

The remaining 23 identity-set members (enum literals, model names, `LLM`, `MoAI-Loop`) received allowlist entries with taxonomy reasons and justifications (see `i18n_untranslated_allowlist_test.go`). The `lang.opt.*` family is excluded structurally by the endonym invariant (REQ-I18NGOV-011/012/013); the `agentdesc.*` reverse gap is governed by the exempt-prefix registry (REQ-I18NGOV-019/020).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-07-29
run_commit_sha: pending-backfill-run-phase
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build.host: PASS (go build ./... exit 0)
cross_platform_build.windows_amd64: PASS (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 3 (2 new _test.go + 1 edited i18n.js)
m1_to_mN_commit_strategy: single run-phase PR (squash)

Evidence redirected to `.moai/state/verify/spec-i18n-governance-001/`:
- web-suite.log, d3.log, full-test.log, vet.log, build-host.log, build-windows.log, lint.log

## §E.4 Sync-phase Audit-Ready Signal

(owned by manager-docs on the single sync commit; placeholder)

sync_commit_sha: ""

## §F Phase 4 Mode Selection

Decision: sub-agent (Mode 5, solo-sequential)

Input parameters:
- tier: M
- scope (file count): 3 (2 new _test.go + 1 edited i18n.js)
- domain count: 1 (internal/web i18n catalogue governance, tests-only)
- file language mix: Go (_test.go) + JS (i18n.js header/value edits)
- concurrency benefit: LOW (single-package, tightly-coupled test infrastructure)

Justification: single-domain, 3-file, tests-only change with no inter-file dependency that would benefit from parallel fan-out. The detector, allowlist, and rulings are logically interdependent (the allowlist is derived from the same measurement that drives the rulings) and land most cleanly in one sequential pass. Mode 4 (parallel) and Mode 6 (workflow) are not indicated: the work is not multi-domain research nor a high-volume mechanical transform.

## §G Implementation Files

- `internal/web/i18n_untranslated_allowlist_test.go` — NEW (M1): reason enum, 23-entry allowlist, en-exempt prefix registry, inline governance contract.
- `internal/web/i18n_governance_test.go` — NEW (M2–M4): parser, pure detector, orphan check, endonym invariants, bidirectional coverage, shape check, negative control, governance-contract self-inspection.
- `internal/web/assets/i18n.js` — EDITED (M1 rulings + M5): header owner line; 3 translation fixes (mp.tier.empty, mp.col.effort, f.model.desc × 3 locales = 9 value edits).
