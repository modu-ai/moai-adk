---
id: SPEC-DB-RETIRE-001
title: "DB 문서화 서브시스템 전면 제거 — 진행 기록"
version: "0.1.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
---

# SPEC-DB-RETIRE-001 — Progress (§E)

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 생성 + plan-auditor D1-D5 반영 완료.
- 22 REQ ↔ 24 AC 매핑(커버리지 22/22). 코드 위치 콘텐츠-앵커 실측 완료(DeprecatedPaths·audit_loader·PhaseDB 추가).
- clarification RESOLVED(db.yaml 능동 삭제=REQ-DBR-022). 미해소 clarification 마커(콜론-정규형) **0건**.
- SPEC ID pre-write self-check PASS. Out of Scope 5항목 명시(MCP backend-db / deployment migration / settings testdata db.yaml / `.moai/project/db/` preserve-root / PhaseDB).

## §E.2 Run-phase Evidence

24/24 AC PASS (0 FAIL). All verification commands run in the worktree at run-phase completion.

| AC | Verification | Actual Output | Status |
|----|--------------|---------------|--------|
| AC-DBR-001 | `moai hook --help` lacks `db-schema-sync` | not listed (39 subcommands) | PASS |
| AC-DBR-002 | `test -d internal/hook/dbsync` | ABSENT | PASS |
| AC-DBR-003 | grep dbsync/handler/loader/helpers in hook.go | 0 | PASS |
| AC-DBR-004 | grep `db-schema-sync` in hook_e2e_test.go | 0 | PASS |
| AC-DBR-005 | grep `loadMigrationPatterns` dirs.go + comment | 0 (db.yaml DeprecatedPaths literal retained) | PASS |
| AC-DBR-006 | grep `db.yaml`/`db-schema-sync` v2_detection.go | 0 | PASS |
| AC-DBR-007 | audit_registry `"db"`=0, audit_loader `"db"`=0, TestAuditParity/TestAuditLoaderCompleteness exit 0 | pass | PASS |
| AC-DBR-008 | leak-test dbsync path=0, replacement `internal/hook/security/scan.go` exists + matches C7 regex | pass | PASS |
| AC-DBR-009 | grep `db-schema-sync`/`internal/hook/dbsync` settings_test.go | 0 | PASS |
| AC-DBR-010 | `test -d templates/.moai/project/db` | ABSENT | PASS |
| AC-DBR-011 | `test -f templates/.moai/config/sections/db.yaml` | ABSENT | PASS |
| AC-DBR-012 | `make build` + `moai init --non-interactive --all` deploys no db scaffold/db.yaml | 33 yaml deployed, 0 db assets | PASS |
| AC-DBR-013 | grep db patterns in doc-generation.md (local+template) | 0 both | PASS |
| AC-DBR-014 | grep db patterns in quality-gates-context.md (local+template) | 0 both | PASS |
| AC-DBR-015 | `diff -q` both guidance files + no SPEC-DB-RETIRE-001 in template skills | IDENTICAL, neutral | PASS |
| AC-DBR-016 | `git ls-files .moai/project/db/ .moai/config/sections/db.yaml` | 0 (git rm) | PASS |
| AC-DBR-017 | CHANGELOG auto-remove(db.yaml) + manual-delete(.moai/project/db/) phrases | both present | PASS |
| AC-DBR-018 | `go build ./...` | exit 0 | PASS |
| AC-DBR-019 | `go test ./...` | exit 0 (106 ok, 0 FAIL) | PASS |
| AC-DBR-020 | `make build` | clean (exit 0) | PASS |
| AC-DBR-021 | pre-removal dbsync importer count | 1 (hook.go), now 0 post-removal | PASS |
| AC-DBR-022 | closure grep `db-schema-sync\|dbsync\|db.enabled` (excl. internal/settings/) | 0 | PASS |
| AC-DBR-023 | dirs.go db.yaml DeprecatedPaths entry (Category D) + clean-reinstall removal test + TestDeprecatedPaths | pass (count 39→40) | PASS |
| AC-DBR-024 | `TestDeprecatedPaths_NoTemplateCollision` | exit 0 (∩ = ∅) | PASS |

Category-split decision (plan §B.8): chose option (b) — new `DeprecatedSince: "SPEC-DB-RETIRE-001"` bucket (Category D). Added a matching `case "SPEC-DB-RETIRE-001":` to `TestDeprecatedPathsDeprecatedByConsistency` (which carries the `default: t.Errorf`) AND a `wantCategoryD=1` counter to `TestDeprecatedPathsCategorySplit`. Count atomically raised 39→40 (`const want`), split now 9/27/3/1 (A/B/C/D).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25
run_commit_sha: <backfill>
run_status: PASS
ac_pass_count: 24
ac_fail_count: 0
preserve_list_post_run_count: 5   # MCP backend-db, deployment Migration Check, settings testdata db.yaml, .moai/project/db preserve-root note, session.PhaseDB
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues; go vet clean
cross_platform_build:
  darwin_arm64: exit 0
  windows_amd64: exit 0
  linux_amd64: exit 0
total_run_phase_files: 37   # 36 removal-scope + progress.md
m1_to_mN_commit_strategy: single M1 commit (removal-only; no milestone split)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소유>_
