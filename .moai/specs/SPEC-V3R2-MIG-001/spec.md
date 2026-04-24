---
id: SPEC-V3R2-MIG-001
title: v2 to v3 Migrator (moai migrate v2-to-v3)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P0 Critical
phase: "v3.0.0 — Phase 8 — Migration Tool + Docs"
module: "internal/cli/migrate.go, internal/core/migration/steps/v2_to_v3/, internal/core/migration/detector.go"
dependencies:
  - SPEC-V3R2-EXT-004
  - SPEC-V3R2-MIG-002
  - SPEC-V3R2-MIG-003
related_gap:
  - problem-catalog-v2-migration
related_theme: "Theme 8 — Migration Tool"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "migration, v2-to-v3, migrator, detector, report, dry-run, v3"
---

# SPEC-V3R2-MIG-001: v2 to v3 Migrator

## HISTORY

| Version | Date       | Author | Description                                                             |
|---------|------------|--------|-------------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — moai migrate v2-to-v3 CLI subcommand + detector + report |

---

## 1. Goal (목적)

사용자 프로젝트가 MoAI-ADK v2.x에서 v3.0.0으로 이행할 때 단일 명령(`moai migrate v2-to-v3`)으로 자동 마이그레이션하는 도구를 제공한다. 본 SPEC은 (a) v2 프로젝트 자동 감지, (b) SPEC-V3R2-EXT-004의 migration framework을 통한 step 일괄 실행, (c) human-readable 리포트 생성, (d) `--dry-run` 프리뷰 모드를 규정한다. Migration step들(24-skill 축소, hook cleanup, config loader completeness 등)의 개별 구현은 각 하위 SPEC(WF-001, MIG-002, MIG-003, 기존 SPEC-V3-CLN-001)에 위임하고, 본 SPEC은 오케스트레이션 레이어만 담당한다.

### 1.1 배경

problem-catalog: v2 사용자가 v3.0.0 업그레이드 시 skill 13개 삭제, hook 재등록, config loader 확장, 템플릿 drift 해결 등 복수 migration을 독립적으로 적용해야 한다. 각 step의 dependency를 수동으로 파악하는 부담은 upgrade adoption을 저해한다. SPEC-V3-CLN-001 legacy 경험에 따르면 `moai migrate v2-to-v3` 엔트리포인트가 migration runner를 chained invoke하는 것이 가장 안전한 UX다.

### 1.2 비목표 (Non-Goals)

- Migration step 구현 (본 SPEC은 오케스트레이션만)
- v3.x → v4 migration (v3 scope 외)
- v2 미만 (v1.x) 지원
- Cross-project batch migration
- 자동 git commit (사용자가 commit 여부 결정)
- Rollback을 위한 full undo script 생성 (EXT-004 rollback 메커니즘 재사용)
- GUI / web interface

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `moai migrate v2-to-v3` CLI 서브커맨드, v2 프로젝트 detector, migration orchestration, 리포트 생성.
- v2 프로젝트 감지 heuristic: `.moai/config/sections/system.yaml: moai.version == 2.x`, `.claude/skills/` 48개 skill 존재, `.claude/hooks/moai/` 26개 wrapper, legacy `.agency/` 디렉터리 존재.
- Migration step 일괄 호출: SPEC-V3R2-WF-001 (skill consolidation), SPEC-V3R2-MIG-002 (hook cleanup), SPEC-V3R2-MIG-003 (config loader completeness), SPEC-V3-CLN-001 (template drift M01/M03/M04).
- `--dry-run` 모드: 모든 예정 변경을 리포트로만 출력.
- 실행 리포트: `.moai/reports/migration-v3-<timestamp>.md`, 각 step의 applied/skipped/rolled-back 상태 + affected files.
- SPEC-V3R2-EXT-004 runner 재사용.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Migration step 자체 구현
- v2 미만 버전 지원
- v3.x → v4 migration path
- GitHub Actions 자동화
- Slack/email 알림
- Multi-project migration
- 프로젝트 fork + 병렬 migration

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/cli/migrate.go`, `internal/core/migration/`
- 영향 디렉터리:
  - 수정: `internal/cli/migrate.go` (신규 혹은 확장), `internal/core/migration/detector.go` (신규)
  - 참조: SPEC-V3R2-EXT-004 framework, 각 하위 SPEC의 migration steps
- 외부 레퍼런스: SPEC-V3-CLN-001 (기존 M01/M03/M04 오케스트레이션 사례)

---

## 4. Assumptions (가정)

- 사용자 프로젝트는 `.moai/`와 `.claude/` 디렉터리를 포함한다.
- `.moai/config/sections/system.yaml`의 `moai.version` 필드가 v2임을 식별 가능하다.
- SPEC-V3R2-EXT-004의 MigrationRunner가 완료되어 있다 (blocked-by 관계).
- 사용자가 git branch 분기 상태에서 migration을 실행한다 (권장; 강제 아님).
- 프로젝트 크기가 1GB 미만이며 backup이 디스크 여유 내에 들어간다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MIG001-001**
The CLI **shall** provide a subcommand `moai migrate v2-to-v3` that orchestrates v2→v3 migration.

**REQ-MIG001-002**
The subcommand **shall** detect v2 projects via the heuristic described in §2.1 (version string + skill count + hook count + optional `.agency/`).

**REQ-MIG001-003**
The subcommand **shall** default to `--dry-run` mode; actual application requires `--apply` flag or interactive confirmation.

**REQ-MIG001-004**
The subcommand **shall** invoke migration steps via SPEC-V3R2-EXT-004 MigrationRunner in the order: template drift (CLN-001 M01/M03/M04) → skill consolidation (WF-001) → hook cleanup (MIG-002) → config loader completeness (MIG-003).

**REQ-MIG001-005**
The subcommand **shall** produce a report at `.moai/reports/migration-v3-<timestamp>.md` including each step's status, affected files, and before/after summary.

**REQ-MIG001-006**
The subcommand **shall** exit with code 0 on full success, non-zero on any step failure.

### 5.2 Event-Driven Requirements

**REQ-MIG001-007**
**When** detector determines the project is already v3 or later, the subcommand **shall** exit with message "Project already on v3+" and exit code 0.

**REQ-MIG001-008**
**When** a migration step fails, the subcommand **shall** halt the chain, invoke `Rollback()` for the failed step, and produce a partial-success report.

**REQ-MIG001-009**
**When** `--dry-run` is active, the subcommand **shall** not modify any files; it **shall** emit predicted diffs to stdout + report.

**REQ-MIG001-010**
**When** user confirmation is required (interactive mode without `--apply`), the subcommand **shall** present the dry-run summary and prompt `[y/N]`.

### 5.3 State-Driven Requirements

**REQ-MIG001-011**
**While** migration runs, `.moai/state/migration.json` **shall** be updated per SPEC-V3R2-EXT-004 REQ-EXT004-004 for each applied step.

**REQ-MIG001-012**
**While** a v3 project is detected, no-op path **shall** be taken (no step execution).

### 5.4 Optional Requirements

**REQ-MIG001-013**
**Where** the user passes `--only <step-id>` (e.g., `--only WF-001`), the subcommand **shall** execute only that step and skip others.

**REQ-MIG001-014**
**Where** the user passes `--skip <step-id>`, the subcommand **shall** skip the named step(s) and continue with others.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-MIG001-015 (Unwanted Behavior)**
**If** the user's project has uncommitted git changes, **then** the subcommand **shall** warn "Uncommitted changes detected; recommend commit or stash before migration" (non-blocking; `--force` to bypass).

**REQ-MIG001-016 (Unwanted Behavior)**
**If** Claude Code is actively running a session in the project directory, **then** the subcommand **shall** warn but allow continuation (non-blocking).

**REQ-MIG001-017 (Complex: State + Event)**
**While** the migration chain is executing, **when** the user presses Ctrl+C, the subcommand **shall** invoke Rollback for the current step, finalize the report with partial status, and exit with code 130.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-MIG001-01**: Given a v2 project When `moai migrate v2-to-v3 --dry-run` runs Then dry-run report is produced without file modification (maps REQ-MIG001-009).
- **AC-MIG001-02**: Given a v2 project When `moai migrate v2-to-v3 --apply` runs Then all 4 migration stages execute in order (maps REQ-MIG001-004).
- **AC-MIG001-03**: Given a v3 project When `moai migrate v2-to-v3` runs Then "Project already on v3+" message + exit code 0 (maps REQ-MIG001-007, REQ-MIG001-012).
- **AC-MIG001-04**: Given a step failure mid-chain When chain runs Then partial-success report + rollback of failed step (maps REQ-MIG001-008).
- **AC-MIG001-05**: Given `--apply` executed successfully When inspected Then `.moai/reports/migration-v3-<ts>.md` exists with per-step status (maps REQ-MIG001-005).
- **AC-MIG001-06**: Given detector with version 2.13.2 + 48 skills + 26 hooks When invoked Then project identified as v2 (maps REQ-MIG001-002).
- **AC-MIG001-07**: Given interactive mode without `--apply` When executed Then dry-run summary shown + [y/N] prompt (maps REQ-MIG001-010).
- **AC-MIG001-08**: Given `--only WF-001` When executed Then only skill consolidation runs (maps REQ-MIG001-013).
- **AC-MIG001-09**: Given `--skip MIG-002` When executed Then hook cleanup step is skipped (maps REQ-MIG001-014).
- **AC-MIG001-10**: Given uncommitted git changes When migration starts Then warning is emitted but continues unless `--force` rejected (maps REQ-MIG001-015).
- **AC-MIG001-11**: Given Ctrl+C during Apply When received Then current step rollbacks and exit code 130 (maps REQ-MIG001-017).
- **AC-MIG001-12**: Given `--apply` with any step failure When run ends Then exit code is non-zero (maps REQ-MIG001-006).
- **AC-MIG001-13**: Given state file `.moai/state/migration.json` When inspected after run Then each applied step has entry with timestamp + affected files (maps REQ-MIG001-011).

---

## 7. Constraints (제약)

- 9-direct-dep 정책 준수 (본 SPEC은 orchestration만, 신규 외부 의존성 없음).
- `--dry-run`을 기본값으로 (destructive default 금지, auto mode protection Section 6 §6).
- Protected directories(CLAUDE.local.md §2.1) 준수: `.moai/project/`, `.moai/specs/` 건드리지 않음 (템플릿 drift migrate는 skills/hooks만 타겟).
- SPEC-V3R2-EXT-004의 MigrationRunner를 재사용 (중복 구현 금지).
- 언어 중립성 유지.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| v2 detector false-positive (이미 v3) | 무효한 migration 시도 | REQ-MIG001-007의 no-op path + 명확한 메시지 |
| 4단계 migration chain 중 2단계 실패 | 반파 상태 | REQ-MIG001-008의 rollback + partial report |
| 사용자 git branch 관리 미숙 | 변경 추적 어려움 | REQ-MIG001-015의 warning + 문서 권장 |
| Claude Code 세션 중 migration 실행 | 파일 충돌 | REQ-MIG001-016의 warning + `.moai/state/migration.lock` 사용 |
| 대규모 프로젝트 migration 시간 | UX 저하 | 각 step의 진행 상황을 stdout에 실시간 스트림; 누적 progress bar |
| 사용자가 `--apply` 없이 실행 후 혼란 | UX 저하 | 첫 줄에 "DRY RUN MODE — no files will change" 명시 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-EXT-004: migration framework이 선행.
- SPEC-V3R2-WF-001: skill consolidation step.
- SPEC-V3R2-MIG-002: hook cleanup step.
- SPEC-V3R2-MIG-003: config loader completeness step.

### 9.2 Blocks

- v3.0.0 공식 릴리스 (사용자가 v2→v3 migrate 없이는 전환 불가).

### 9.3 Related

- SPEC-V3-CLN-001 (legacy): 기존 M01/M03/M04 step을 본 SPEC가 orchestrate.

---

## 10. Traceability (추적성)

- REQ 총 17개: Ubiquitous 6, Event-Driven 4, State-Driven 2, Optional 2, Complex 3.
- AC 총 13개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: SPEC-V3-CLN-001 오케스트레이션 사례; problem-catalog ad-hoc migration.
- BC 영향: 없음 (dry-run 기본 + 사용자 confirmation 필수; SPEC-V3R2-EXT-004의 BC-V3R2-019만 적용).
- 구현 경로 예상:
  - `internal/cli/migrate.go` (확장: `v2-to-v3` 서브커맨드)
  - `internal/core/migration/detector.go` (신규)
  - `.moai/reports/migration-v3-<timestamp>.md` (런타임 생성)
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1082` (§11.8 MIG-001 definition)
  - `docs/design/major-v3-master.md:L958-978` (§8 BC catalog — all AUTO BCs executed by migrator)
  - `docs/design/major-v3-master.md:L994` (§9 Phase 8 Migration Tool + Docs)

---

End of SPEC.
