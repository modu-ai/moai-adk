---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Progress — Template Internal-Content Isolation"
version: "0.1.0"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, internal-content, progress"
tier: M
---

# Progress — Template Internal-Content Isolation

## §A. Lifecycle Sync

| Field | Value |
|-------|-------|
| `plan_commit_sha` | `b7d1528c8` (canonical plan-phase anchor; spec.md §A.3 row) |
| `iter1_amend_commit_sha` | `5ff9da7d2` (iter-1 amendment, 4 SHOULD-FIX 해소) |
| `run_phase_entry_head` | `5ff9da7d2` (M1 시작 HEAD) |
| `run_phase_branch` | `main` (Hybrid Trunk 1-person OSS Tier M 직진) |
| `run_status` | `in-progress` (M1 시작; M6 종료 시 `PASS` 또는 `PASS-WITH-DEBT` 전환) |
| `run_complete_at` | _(M-final commit 후 ISO-8601 timestamp 기재)_ |
| `run_commit_sha` | _(M-final commit SHA 기재)_ |

## §B. Milestones Progress

| Milestone | Description | Status | Commit SHA |
|-----------|-------------|--------|------------|
| M1 | Status transition + D-008 hyphen fix + progress.md initial | in-progress | _(M1 commit pending)_ |
| M2 | CLAUDE.local.md §25 NEW HARD rule | pending | _(M2 commit pending)_ |
| M3 | Go lint test `TestTemplateNoInternalContentLeak` + D-007 short-sha extension | pending | _(M3 commit pending)_ |
| M4.1 | `.claude/agents/` sub-batch cleanup (5 files) | pending | _(M4.1 commit pending)_ |
| M4.2 | `.claude/rules/moai/` sub-batch cleanup (10 files) | pending | _(M4.2 commit pending)_ |
| M4.3 | `.claude/skills/` sub-batch cleanup (~14 files) | pending | _(M4.3 commit pending)_ |
| M4.4 | Other singleton files cleanup (~5 files) | pending | _(M4.4 commit pending)_ |
| M5 | CI workflow policy anchor + `go test ./internal/template/...` invocation | pending | _(M5 commit pending)_ |
| M6 | Maintainer-only audit + `.gitignore` guard (conditional) | pending | _(M6 commit pending)_ |
| Post-M6 | Run-phase audit-ready signal + progress.md final | pending | _(Post-M6 commit pending)_ |

## §C. Pre-flight Verification (M1 시작 시 ground truth 재확인)

| Check | Expected | Actual | Status |
|-------|----------|--------|--------|
| HEAD SHA | `5ff9da7d2` (iter-1 amendment) | `5ff9da7d2bd981b33ebafb3af7df13d17fc4fcfd` | PASS |
| origin/main sync | `0 0` | `0 0` | PASS |
| Leak file count (4-class prose) | `35` | `35` | PASS (ground truth 일치) |
| CLAUDE.local.md §25 부재 | `0` | `0` | PASS |
| Go build (host) | exit 0 | exit 0 | PASS |
| Go build (windows/amd64) | exit 0 | exit 0 | PASS |
| embedded.go 존재 여부 | N/A (per `internal/template/embed.go` 직접 `//go:embed all:templates` directive — generated 별도 파일 없음) | N/A | INFO (plan.md §C #3 무효; AC-TII-003 검증 방식 조정: shasum 대신 `go build` + `EmbeddedTemplates()` 접근 가능성으로 mirror parity 증명) |

## §D. Run-phase Evidence Table (E.2)

_M6 종료 후 12/12 AC PASS/FAIL/PASS-WITH-DEBT 매트릭스 + Actual Output 기재._

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|---------------|
| AC-TII-001 | pending | _(see acceptance.md AC-TII-001)_ | _(M4.4 종료 후 기재)_ |
| AC-TII-002 | pending | _(see acceptance.md AC-TII-002)_ | _(M4 종료 후 기재)_ |
| AC-TII-003 | pending | _(see acceptance.md AC-TII-003)_ | _(M4 sub-batch 별 기재)_ |
| AC-TII-004 | pending | _(see acceptance.md AC-TII-004)_ | _(M2 종료 후 기재)_ |
| AC-TII-005 | pending | _(see acceptance.md AC-TII-005)_ | _(M2 종료 후 기재)_ |
| AC-TII-006 | pending | _(see acceptance.md AC-TII-006)_ | _(M3 종료 후 기재)_ |
| AC-TII-007 | pending | _(see acceptance.md AC-TII-007)_ | _(M3 종료 후 기재)_ |
| AC-TII-008 | pending | _(see acceptance.md AC-TII-008)_ | _(M5 종료 후 기재)_ |
| AC-TII-009 | pending | _(see acceptance.md AC-TII-009)_ | _(M6 종료 후 기재)_ |
| AC-TII-010 | pending | _(see acceptance.md AC-TII-010)_ | _(M6 종료 후 기재 — 조건부)_ |
| AC-TII-011 | pending | _(see acceptance.md AC-TII-011)_ | _(M2 종료 후 기재)_ |
| AC-TII-012 | pending | _(see acceptance.md AC-TII-012)_ | _(M4/M5 종료 후 기재)_ |

## §E. Audit-Ready Signal (E.3)

```yaml
run_phase_audit_ready_signal:
  run_complete_at: "_(Post-M6 commit timestamp)_"
  run_commit_sha: "_(Post-M6 commit SHA)_"
  run_status: "_(PASS | PASS-WITH-DEBT | FAIL)_"
  ac_pass_count: 0
  ac_fail_count: 0
  ac_pending_count: 12
  preserve_list_post_run_count: 0
  l44_pre_commit_fetch: "0 0 (pre-M1 verified)"
  l44_post_push_fetch: "_(M-final post-push)_"
  new_warnings_or_lints_introduced: "_(M-final 측정)_"
  cross_platform_build:
    host_amd64: "_(M-final 측정)_"
    windows_amd64: "_(M-final 측정)_"
  total_run_phase_files: 0
  m1_to_mN_commit_strategy: "M-별 atomic commit (M1, M2, M3, M4.1-4, M5, M6, progress.md) on main 직진 — Hybrid Trunk 1-person OSS Tier M; pre/post-commit fetch verified per L44 HARD"
```

## §F. Cross-References

- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/spec.md` v0.1.2 — REQ-TII-001~013
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/plan.md` v0.1.1 — M1-M6 milestones
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/acceptance.md` v0.1.2 — AC-TII-001~012
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md` v0.1.1 — Substitution Dictionary + Allowlist + CI Hook placement
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/research.md` v0.1.1 — Predecessor cleanup pattern 분석

## §G. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| v0.1.0 | 2026-05-25 | manager-develop | M1 initial — §A Lifecycle Sync table + §B milestones progress board + §C pre-flight verification (5/6 PASS, embedded.go 사전 가정 무효 INFO) + §D AC pending matrix + §E audit-ready signal skeleton |
