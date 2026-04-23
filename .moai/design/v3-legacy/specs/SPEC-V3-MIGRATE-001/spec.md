---
id: SPEC-V3-MIGRATE-001
title: moai migrate v2-to-v3 Unified Migration Tool
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: Wave 4 SPEC writer
priority: P1 High
phase: "v3.0.0 — Phase 7 — Migration Tool + User Docs"
module: "internal/cli/migrate_v2_to_v3.go, docs-site/content/{ko,en,ja,zh}/migration/v3.md"
dependencies:
  - SPEC-V3-MIG-001
  - SPEC-V3-MIG-002
  - SPEC-V3-CLN-001
  - SPEC-V3-CLN-002
  - SPEC-V3-HOOKS-001
  - SPEC-V3-SCH-001
  - SPEC-V3-SCH-002
related_gap:
  - gm#149
  - gm#150
  - gm#151
related_theme: "Theme 2 — Migration Framework (user-facing tool)"
breaking: true
bc_id: [BC-004]
lifecycle: spec-anchored
tags: "migration, v2-to-v3, tool, cli, user-facing, breaking, v3"
---

# SPEC-V3-MIGRATE-001: `moai migrate v2-to-v3` Unified Migration Tool

## HISTORY

| Version | Date       | Author | Description                                                     |
|---------|------------|--------|-----------------------------------------------------------------|
| 0.1.0   | 2026-04-22 | Wave 4 | Initial SPEC draft covering v2→v3 user-facing migration command |

---

## 1. Goal (목적)

v2.x 프로젝트를 v3.0으로 안전하게 전환하기 위한 **단일 사용자 가시 커맨드** `moai migrate v2-to-v3`를 제공한다. 이 커맨드는 SPEC-V3-MIG-001이 정의한 versioned migration framework 위에서 동작하며, SPEC-V3-MIG-002의 초기 migration set(M01-M05)을 일괄 적용하고, `moai update`와 `moai doctor config --fix`를 순차 실행하여 v3.0 스키마를 완성한다. **`--dry-run`을 기본으로 사용자 확인 단계를 거치고, 모든 변경은 자동 백업과 rollback 경로를 갖추어 데이터 손실 위험을 최소화한다.**

### 1.1 배경

v3-master §5.1에 따르면 v2.x → v3.0 마이그레이션은 여러 단계(migration steps M01-M05, template refresh, schema auto-repair)를 포함하며, 사용자가 이를 개별적으로 호출하면 순서 오류나 누락이 발생할 수 있다. Claude Code의 `migrate` 커맨드는 `preAction` 훅에서 자동 실행되지만(Wave 1.5 §5.2), moai-adk-go는 더 보수적인 전략(명시적 커맨드에서만 실행, master §3.2 open question #10)을 채택하고 있다. 사용자 관점에서는 "하나의 커맨드로 전체 v2→v3 변환"이 필요하며, 이것이 본 SPEC의 목표다.

현재 `moai migrate agency`(SPEC-AGENCY-ABSORB-001, `internal/cli/migrate_agency.go:569`)가 유일한 migration 커맨드이며, 이 패턴을 일반화하여 `moai migrate v2-to-v3` subcommand를 추가한다.

### 1.2 비목표 (Non-Goals)

- v3→v4 또는 이후 버전 마이그레이션 (스코프 밖)
- 개별 migration step (M01-M05)의 로직 구현 — 각각 별도 SPEC(V3-CLN-001, V3-CLN-002, V3-MIG-002)이 담당
- `preAction` 훅에서 자동 실행 (master §3.2 open question #10 보수적 선택: 명시적 커맨드만)
- v2.x 이전 버전(v1.x) 사용자 지원 — v2.x 진입이 선결
- Rollback 후 프로젝트를 v2.x 동작 상태로 완전 복원 — 백업에서 파일 restore는 가능하나 v2.x 바이너리 다운그레이드는 사용자 책임
- GUI/TUI 진행률 표시 — stdout 기반 progress만 제공
- Claude Code 측 설정(`~/.claude/`, `settings.local.json`) 마이그레이션 — moai scope 밖

---

## 2. Scope (범위)

### 2.1 In Scope

- `moai migrate v2-to-v3` 신규 subcommand (`internal/cli/migrate_v2_to_v3.go`)
- 플래그: `--dry-run` (기본 true for interactive, false for `--yes`), `--yes` (skip confirmations), `--no-backup` (NOT recommended), `--only <step>` (개별 step 실행), `--rollback <ts>` (과거 백업에서 복원)
- Pre-flight check: moai 바이너리 버전이 v3.0.0+ 인지 검증
- Snapshot 생성: `.moai/backups/v2-to-v3-<ISO-8601-timestamp>/`
- Migration runner invocation: SPEC-V3-MIG-001 runner를 통해 M01-M05 순차 실행
- Template refresh: `moai update` 로직 호출 (migration 이후)
- Schema auto-repair: `moai doctor config --fix` 로직 호출 (template refresh 이후)
- 요약 리포트 생성: `.moai/reports/migration-v2-to-v3-<timestamp>.md`
- Rollback 커맨드: `moai migrate v2-to-v3 --rollback <ts>`로 과거 상태 복원
- 4-locale 사용자 문서: `docs-site/content/{ko,en,ja,zh}/migration/v3.md` (CLAUDE.local.md §17 준수)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 개별 migration step의 내부 로직 (SPEC-V3-CLN-001/002, SPEC-V3-MIG-002 담당)
- SPEC-V3-MIG-001 framework 자체 (본 SPEC은 consumer일 뿐)
- v3.0 새 기능 추가 (Hooks v2, Schema, Agent frontmatter 등은 개별 SPEC 담당)
- 자동 실행 (`preAction` 훅 통합) — master §3.2 open question #10의 conservative 선택
- 실시간 TUI progress bar — stdout 라인 형식만 지원
- 여러 프로젝트 일괄 마이그레이션 — 단일 프로젝트 CWD 기준만
- Git branch/commit 자동 생성 — 사용자가 직접 VCS 작업 수행 (주의 안내만 제공)
- v3.0 이전 alpha/beta 빌드에서 v3.0 stable로의 전환 — alpha/beta는 자체 규칙 따름
- 마이그레이션 분석/텔레메트리 전송 — moai는 텔레메트리 보내지 않음(기존 정책 유지)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+) v3.0.0+
- Claude Code v2.1.111+
- OS: macOS, Linux, Windows (평가 대상)
- 영향 디렉터리:
  - 신설: `internal/cli/migrate_v2_to_v3.go`, `internal/cli/migrate_v2_to_v3_test.go`
  - 신설: `docs-site/content/ko/migration/v3.md`, `en/migration/v3.md`, `ja/migration/v3.md`, `zh/migration/v3.md`
  - 수정: `internal/cli/migrate_agency.go` 주변의 cobra 등록부 (v2-to-v3 subcommand 추가)
- 대상 프로젝트: `.moai/config/sections/system.yaml:moai.version >= v2.0.0 && < v3.0.0`
- 외부 레퍼런스: Wave 1.5 §5.1-§5.2 (Claude Code migration framework), master §5.1 (tool design), §5.2 (shims), §5.3 (user-facing guide sketch), CLAUDE.local.md §17 (docs-site 4-locale rules)

---

## 4. Assumptions (가정)

- SPEC-V3-MIG-001(framework)과 SPEC-V3-MIG-002(initial migration set M01-M05)가 Phase 1에서 선행 완료된다.
- SPEC-V3-CLN-001/002/003가 모두 Phase 1에서 동반 완료되어 M01-M05 구현체가 존재한다.
- 사용자는 migration 실행 전 VCS commit을 수행한다 (우리가 대신 commit하지 않음).
- `.moai/backups/` 디렉터리에 대한 쓰기 권한이 있고 디스크 여유 공간이 프로젝트 크기의 최소 2배이다.
- Pre-v2.0 프로젝트(v1.x 이하)는 사전에 v2.x로 업그레이드되어 있다 (§1.2 Non-Goal 준수).
- `moai migrate` 최상위 커맨드는 SPEC-AGENCY-ABSORB-001의 migrate_agency.go 패턴으로 이미 cobra에 등록되어 있으며, `v2-to-v3`가 subcommand로 붙는다.
- 4-locale 문서 작성은 manager-docs agent와 expert-docs(ja/zh 번역)에 위임 가능하다 (CLAUDE.local.md §17.5).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MIGRATE-001-001**
The `moai migrate v2-to-v3` 커맨드 **shall** cobra subcommand로 `moai migrate`의 하위에 등록된다 (`migrate.AddCommand(v2ToV3Cmd)` 패턴).

**REQ-MIGRATE-001-002**
The 커맨드 **shall** 다음 flag 집합을 지원한다:
- `--dry-run` (bool, default: interactive 시 true, `--yes` 지정 시 false)
- `--yes` / `-y` (bool, default: false) — 확인 프롬프트 skip
- `--no-backup` (bool, default: false) — NOT recommended 안내와 함께 백업 skip
- `--only <step-id>` (string) — e.g., `M01`, `M03`. 단일 step만 실행
- `--rollback <timestamp>` (string) — 과거 백업 복원 (다른 flag와 상호 배타적)

**REQ-MIGRATE-001-003**
The 커맨드 **shall** 실행 전 pre-flight check로 현재 `moai` 바이너리 버전(`pkg/version.GetVersion()`)이 `v3.0.0` 이상인지 확인하고, 미만이면 `MIGRATE_V2V3_BINARY_TOO_OLD` 오류와 함께 "upgrade moai binary first: `brew upgrade moai` or equivalent" 안내를 출력하고 exit code 2로 종료한다.

**REQ-MIGRATE-001-004**
The 커맨드 **shall** 현재 프로젝트가 v2.x인지 `.moai/config/sections/system.yaml:moai.version`에서 확인하며, v3.x 이상이면 `MIGRATE_V2V3_ALREADY_V3` 정보 메시지를 출력하고 exit code 0으로 종료한다 (이미 마이그레이션 완료 상태는 오류가 아님).

**REQ-MIGRATE-001-005**
The 커맨드 **shall** 다음 순서로 v2→v3 전환 단계를 실행한다:
1. Pre-flight checks (REQ-MIGRATE-001-003, REQ-MIGRATE-001-004, REQ-MIGRATE-001-007)
2. Backup snapshot 생성 (`--no-backup` 미지정 시)
3. Migration runner invocation: SPEC-V3-MIG-001 Runner.Run()으로 M01-M05 순차 실행
4. Template refresh: `moai update` 로직 (사용자 파일은 건드리지 않음, `TemplateManaged` 파일만)
5. Schema auto-repair: `moai doctor config --fix` 로직 (recoverable errors)
6. Summary report 생성

**REQ-MIGRATE-001-006**
The 커맨드 **shall** Migration runner(SPEC-V3-MIG-001)와 동일한 `.moai/backups/v2-to-v3-<ISO-8601-timestamp>/` 디렉터리에 마이그레이션 전 스냅샷을 생성하며, 이 타임스탬프는 summary report와 rollback 가이드에 포함된다.

**REQ-MIGRATE-001-007**
The 커맨드 **shall** 실행 전 디스크 공간을 검증하며, 프로젝트 `.moai/`와 `.claude/` 크기 합의 최소 2배의 여유 공간을 요구한다. 부족 시 `MIGRATE_V2V3_DISK_FULL`로 종료.

**REQ-MIGRATE-001-008**
The 커맨드 **shall** Summary report를 `.moai/reports/migration-v2-to-v3-<timestamp>.md`에 저장하며 다음 내용을 포함한다:
- Migration 시작/종료 시각
- 적용된 step 목록 (M01, M03, M04, M05 각각의 상태: applied / skipped / rolled-back)
- 변경된 파일 경로 목록
- Backup 디렉터리 절대 경로
- Rollback 가이드 ("`moai migrate v2-to-v3 --rollback <timestamp>`"
- 후속 안내 ("run `moai doctor` to verify", 4-locale docs URL)

### 5.2 Event-Driven Requirements

**REQ-MIGRATE-001-009**
**When** 커맨드가 interactive 모드(`--yes` 미지정)로 실행되면, the 커맨드 **shall** migration plan(각 step의 dry-run diff)을 출력한 후 사용자에게 확인 프롬프트 "Apply these changes? (y/N)"를 제시하고, "y" / "Y" / "yes" 이외의 응답 시 중단한다.

**REQ-MIGRATE-001-010**
**When** 사용자가 `--rollback <timestamp>` flag를 제공하면, the 커맨드 **shall** `.moai/backups/v2-to-v3-<timestamp>/`의 존재를 확인하고, 해당 백업에서 파일들을 원 위치로 복원하며, `.moai/config/sections/system.yaml:migration_version`을 백업 시점 값으로 되돌린다.

**REQ-MIGRATE-001-011**
**When** migration 실행 중 임의 step이 실패하면, the 커맨드 **shall** Runner의 step-level rollback을 신뢰하고, 실패한 step 이후의 단계는 실행하지 않으며, 사용자에게 "Step <ID> failed: <reason>. Run `moai migrate v2-to-v3 --rollback <timestamp>` to revert earlier steps if needed." 안내를 출력한다.

**REQ-MIGRATE-001-012**
**When** `--dry-run` 모드로 실행되면, the 커맨드 **shall** 실제 파일 변경을 수행하지 않고 모든 step의 예상 변경 내역을 master §5.1의 dry-run output 예시와 동일한 형식으로 stdout에 출력한다.

**REQ-MIGRATE-001-013**
**When** `--only <step-id>`가 지정되면, the 커맨드 **shall** SPEC-V3-MIG-001 runner의 단일 step 실행 모드를 사용하여 해당 step만 적용하고, template refresh와 schema auto-repair 단계(step 4, 5)는 skip한다.

### 5.3 State-Driven Requirements

**REQ-MIGRATE-001-014**
**While** 커맨드가 interactive 모드에서 dry-run preview를 출력 중인 상태에서, the 시스템 **shall** terminal width를 감지하고 각 step의 파일 경로가 terminal column을 초과하면 truncation(`...` with middle ellipsis)을 적용한다.

**REQ-MIGRATE-001-015**
**While** `--rollback <timestamp>`가 지정된 상태에서, the 커맨드 **shall** migration 실행(step 2-6)을 수행하지 않고 오직 복원만 수행한다 (mutually exclusive와의 조합 금지).

**REQ-MIGRATE-001-016**
**While** `--no-backup` flag가 지정된 상태에서, the 커맨드 **shall** 실행 시작 시 "WARNING: --no-backup is NOT recommended. Rollback will not be available." 경고를 stderr에 emit하고 10초 대기 후 진행한다 (`--yes`가 동시 지정되어도 대기 없이 진행은 허용).

### 5.4 Optional Requirements

**REQ-MIGRATE-001-017**
**Where** 환경변수 `MOAI_MIGRATE_NO_BACKUP=1`가 설정된 환경에서, the 커맨드 **shall** `--no-backup` flag와 동일하게 동작하며 경고를 emit한다 (CI/airgap 자동화 지원).

**REQ-MIGRATE-001-018**
**Where** 사용자가 `--verbose` flag를 제공한 환경에서, the 커맨드 **shall** 각 step의 내부 상태(pre-condition 결과, 개별 파일 변경, 체크섬 비교)를 stdout에 상세 출력한다.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-MIGRATE-001-019 (Unwanted Behavior)**
**If** `--rollback <timestamp>`와 다른 실행 flag(`--yes`, `--no-backup`, `--only` 등)가 함께 지정되면, **then** the 커맨드 **shall** `MIGRATE_V2V3_INCOMPATIBLE_FLAGS` 오류를 반환하고 exit code 2로 종료한다.

**REQ-MIGRATE-001-020 (Unwanted Behavior)**
**If** `.moai/backups/v2-to-v3-<timestamp>/`가 존재하지 않는 상태에서 `--rollback`이 지정되면, **then** the 커맨드 **shall** `MIGRATE_V2V3_BACKUP_NOT_FOUND` 오류를 반환하고 `.moai/backups/` 하위의 사용 가능한 백업 타임스탬프 목록을 stderr에 표시한다.

**REQ-MIGRATE-001-021 (Complex: State + Event, SIGINT/SIGTERM)**
**While** 커맨드가 실행 중이고, **when** SIGINT 또는 SIGTERM 시그널이 수신되면, the 커맨드 **shall** 현재 진행 중인 step을 완료하도록 Runner에 시그널을 전파하며(SPEC-V3-MIG-001 책임), 개별 step의 transaction checkpoint flush와 rollback 시도를 Runner에 위임하고, exit code 130(SIGINT) 또는 143(SIGTERM)으로 종료한다. 재실행 시 `moai migrate v2-to-v3 --resume <tx-id>` 플래그로 이어서 복구 가능하다.

**REQ-MIGRATE-001-022 (Unwanted Behavior)**
**If** template refresh 단계(step 4)에서 사용자가 수정한 파일(`UserModified` provenance)이 존재하면, **then** the 커맨드 **shall** 해당 파일을 건드리지 않고 summary report에 "User-modified files preserved: <list>"를 기록하며, 사용자에게 "These files were not refreshed; review the template changes at docs-site/content/<locale>/migration/v3.md" 안내를 제공한다.

**REQ-MIGRATE-001-023 (Unwanted Behavior)**
**If** schema auto-repair(step 5)가 unrecoverable error를 반환하면, **then** the 커맨드 **shall** 전체 마이그레이션을 실패로 표시하되 앞선 step들의 rollback은 시도하지 않고 (step-level 성공 유지), 사용자에게 `moai doctor config` 수동 실행으로 상세 오류 확인을 안내한다.

**REQ-MIGRATE-001-024 (Ubiquitous, docs consistency)**
The 4-locale 사용자 문서(`docs-site/content/{ko,en,ja,zh}/migration/v3.md`) **shall** CLAUDE.local.md §17.1(URL blacklist), §17.2(강조 표기 공백, Mermaid TD only), §17.3(ko canonical, 48h en SLA, 72h zh/ja SLA) 규칙을 준수하며 다음 섹션을 포함한다:
- "Why v3?" (1 paragraph summary)
- "Who needs to migrate?"
- "Before you start" (VCS commit 권장, backup 확인)
- "The one command" (`moai migrate v2-to-v3 --dry-run && moai migrate v2-to-v3 --yes`)
- "What changes" (affected files table per migration step)
- "Hook authors" / "Config maintainers" / "Team mode users" (BC 별 상세)
- "Rollback"
- "Troubleshooting"

**REQ-MIGRATE-001-025 (Unwanted Behavior, docs gate)**
**If** `moai migrate v2-to-v3`가 Phase 7 릴리스 검증에서 실행될 때 `docs-site/content/{ko,en,ja,zh}/migration/v3.md` 중 어느 하나라도 누락 또는 빈 상태이면, **then** release CI **shall** tag 푸시를 차단한다 (CLAUDE.local.md §17.6 release checklist 준수).

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-MIGRATE-001-01**: Given v2.x 프로젝트와 `moai v3.0.0+` 바이너리 When `moai migrate v2-to-v3 --dry-run` 실행 Then 실제 파일 변경 없이 M01-M05 예상 변경 내역이 master §5.1 예시와 동일 형식으로 stdout에 출력됨 (maps REQ-MIGRATE-001-012).
- **AC-MIGRATE-001-02**: Given v2.x 프로젝트 When `moai migrate v2-to-v3 --yes` 실행 Then backup 생성 → M01-M05 순차 적용 → template refresh → schema auto-repair → summary report 순서로 실행되고 exit code 0 반환 (maps REQ-MIGRATE-001-005, REQ-MIGRATE-001-006).
- **AC-MIGRATE-001-03**: Given `moai` 바이너리 버전이 v2.x (< v3.0.0) 인 상태 When `moai migrate v2-to-v3` 실행 Then `MIGRATE_V2V3_BINARY_TOO_OLD` 오류와 업그레이드 안내가 출력되고 exit code 2 반환 (maps REQ-MIGRATE-001-003).
- **AC-MIGRATE-001-04**: Given 이미 v3.x인 프로젝트 When `moai migrate v2-to-v3` 재실행 Then `MIGRATE_V2V3_ALREADY_V3` 정보 메시지가 출력되고 exit code 0 반환 (maps REQ-MIGRATE-001-004).
- **AC-MIGRATE-001-05**: Given `.moai/backups/v2-to-v3-<ts>/` 백업 디렉터리가 존재 When `moai migrate v2-to-v3 --rollback <ts>` 실행 Then 백업에서 파일들이 원위치로 복원되고 `migration_version`이 백업 시점 값으로 되돌려짐 (maps REQ-MIGRATE-001-010).
- **AC-MIGRATE-001-06**: Given v2.x 프로젝트 When `moai migrate v2-to-v3 --only M01` 실행 Then M01만 적용되고 M02-M05, template refresh, schema auto-repair 단계는 skip됨 (maps REQ-MIGRATE-001-013).
- **AC-MIGRATE-001-07**: Given `--rollback <ts>`와 `--yes`가 동시에 지정된 상태 When 커맨드 실행 Then `MIGRATE_V2V3_INCOMPATIBLE_FLAGS` 오류를 반환하고 exit code 2 (maps REQ-MIGRATE-001-019).
- **AC-MIGRATE-001-08**: Given migration 실행 중인 상태 When SIGINT 또는 SIGTERM 시그널 수신 Then SPEC-V3-MIG-001 runner에 시그널 전파되고 각각 exit code 130 / 143으로 종료 (maps REQ-MIGRATE-001-021).
- **AC-MIGRATE-001-09**: Given migration이 성공 완료된 상태 When summary report 파일을 검사 Then `.moai/reports/migration-v2-to-v3-<timestamp>.md`에 REQ-MIGRATE-001-008의 모든 필드가 포함됨 (maps REQ-MIGRATE-001-008).
- **AC-MIGRATE-001-10**: Given release CI가 Phase 7 release 검증 중 When `docs-site/content/{ko,en,ja,zh}/migration/v3.md` 중 하나라도 누락 또는 빈 상태 Then release CI가 tag 푸시를 차단 (maps REQ-MIGRATE-001-025).
- **AC-MIGRATE-001-11**: Given migration의 template refresh 단계가 실행 중 When `UserModified` provenance 파일이 존재 Then 해당 파일은 건드려지지 않고 summary report에 "User-modified files preserved: <list>"가 기록됨 (maps REQ-MIGRATE-001-022).
- **AC-MIGRATE-001-12**: Given `.moai/backups/v2-to-v3-<ts>/`가 존재하지 않는 상태 When `--rollback <ts>` 실행 Then `MIGRATE_V2V3_BACKUP_NOT_FOUND` 오류와 사용 가능한 백업 목록이 stderr에 출력됨 (maps REQ-MIGRATE-001-020).
- **AC-MIGRATE-001-13**: Given moai CLI가 cobra 트리로 구성된 상태 When `moai migrate --help` 실행 Then `v2-to-v3` 서브커맨드가 `moai migrate` 하위에 나열됨 (maps REQ-MIGRATE-001-001).
- **AC-MIGRATE-001-14**: Given `moai migrate v2-to-v3 --help` 실행 상태 When help 출력 검사 Then `--dry-run`, `--yes`, `--no-backup`, `--only`, `--rollback` 5개 플래그 모두 표시됨 (maps REQ-MIGRATE-001-002).
- **AC-MIGRATE-001-15**: Given `.moai/`와 `.claude/` 합계 크기의 2배 미만 디스크 여유 상태 When `moai migrate v2-to-v3` 실행 Then `MIGRATE_V2V3_DISK_FULL` 오류 반환 (maps REQ-MIGRATE-001-007).
- **AC-MIGRATE-001-16**: Given interactive 모드(`--yes` 미지정) When `moai migrate v2-to-v3` 실행 Then dry-run diff 출력 후 "Apply these changes? (y/N)" 프롬프트 제시되고, "n" 응답 시 중단 (maps REQ-MIGRATE-001-009).
- **AC-MIGRATE-001-17**: Given M03 step이 Apply 중 실패 상태 When runner가 보고함 Then "Step M03 failed: <reason>. Run `moai migrate v2-to-v3 --rollback <timestamp>` to revert earlier steps if needed." 안내가 stderr에 출력되고 이후 step은 실행되지 않음 (maps REQ-MIGRATE-001-011).
- **AC-MIGRATE-001-18**: Given terminal width 80 컬럼 + 긴 파일 경로 상태 When dry-run preview가 출력됨 Then 각 경로가 middle ellipsis (`...`)로 truncation되어 컬럼 초과 방지 (maps REQ-MIGRATE-001-014).
- **AC-MIGRATE-001-19**: Given `--rollback <ts>` 지정 + 유효한 백업 존재 상태 When 커맨드 실행 Then 오직 복원만 수행되고 step 2-6(migration runner 등)은 실행되지 않음 (maps REQ-MIGRATE-001-015).
- **AC-MIGRATE-001-20**: Given `--no-backup` 플래그 지정 상태 When 커맨드 시작 Then stderr에 "WARNING: --no-backup is NOT recommended. Rollback will not be available." 경고 emit, `--yes` 병기 시 대기 없이 진행 (maps REQ-MIGRATE-001-016).
- **AC-MIGRATE-001-21**: Given 환경변수 `MOAI_MIGRATE_NO_BACKUP=1` 설정된 상태 When 커맨드 실행 Then `--no-backup`과 동일하게 동작하며 경고 emit (maps REQ-MIGRATE-001-017).
- **AC-MIGRATE-001-22**: Given `--verbose` 플래그 지정 상태 When migration 실행 Then 각 step의 pre-condition 결과, 파일별 변경, 체크섬 비교가 stdout에 상세 출력됨 (maps REQ-MIGRATE-001-018).
- **AC-MIGRATE-001-23**: Given schema auto-repair가 unrecoverable error 반환 상태 When 커맨드 평가 Then 전체 migration이 실패 표시되고 앞선 step rollback 미시도, `moai doctor config` 실행 안내 제공 (maps REQ-MIGRATE-001-023).
- **AC-MIGRATE-001-24**: Given 4-locale 파일(`ko/en/ja/zh/migration/v3.md`) 모두 작성된 상태 When 각 파일 검사 Then CLAUDE.local.md §17.1/§17.2/§17.3 규칙 준수(URL blacklist, Mermaid TD, 강조 공백)하며 "Why v3?", "The one command", "Rollback", "Troubleshooting" 등 8개 섹션 모두 포함됨 (maps REQ-MIGRATE-001-024).

---

## 7. Constraints (제약)

- 본 SPEC은 SPEC-V3-MIG-001 framework의 **consumer**이며, runner/backup/rollback 내부 로직을 재구현하지 않는다.
- SPEC-AGENCY-ABSORB-001 `moai migrate agency` 패턴(`internal/cli/migrate_agency.go:569-`)을 재사용하여 subcommand를 일관된 구조로 구현한다.
- 9-direct-dep 정책: 신규 외부 의존성 금지.
- 플랫폼 호환: Windows에서도 rollback이 정상 동작해야 한다 (POSIX permission bits는 no-op, SPEC-AGENCY-ABSORB-001 REQ-MIGRATE-012a/b 패턴 재사용).
- Dry-run은 interactive 모드의 기본 동작 (`--yes` 미지정 시 자동 활성).
- Rollback은 파일 단위 restore만 지원 (VCS state, external config는 사용자 책임).
- 4-locale 문서는 CLAUDE.local.md §17 준수 (canonical ko, 48h en SLA, 72h zh/ja SLA, URL blacklist).
- Claude Code 측 설정은 건드리지 않는다 (moai scope 밖).
- `preAction` 자동 실행 미지원 (master §3.2 open question #10 보수적 선택).
- 텔레메트리 전송 없음 (기존 moai 정책 유지).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| 사용자가 `--no-backup`으로 백업 없이 실행 후 실패 | 데이터 복구 불가 | REQ-MIGRATE-001-016의 10초 경고. 문서에 "NOT recommended" 강조 |
| Rollback 중 `.moai/backups/` 손상 | 복원 실패 | REQ-MIGRATE-001-020의 상세 오류 및 사용 가능한 백업 목록 제공 |
| SIGINT 수신 시 partial state 잔존 | 프로젝트 일관성 저하 | REQ-MIGRATE-001-021의 transaction checkpoint, `--resume` flag (SPEC-V3-MIG-001 framework 책임) |
| User-modified 파일이 template refresh로 덮어씌워짐 | 사용자 작업 손실 | REQ-MIGRATE-001-022로 provenance 기반 보호, summary report에 목록화 |
| v3.0 이전 alpha/beta 빌드 사용자 혼란 | UX 저하 | Pre-flight check의 version 비교에서 semver pre-release 태그 처리. 문서에 "alpha/beta는 자체 규칙" 명시 |
| 4-locale 문서 번역 lag로 릴리스 지연 | 릴리스 지연 | CLAUDE.local.md §17.3의 48h/72h SLA 준수, CI에서 95% parity로 gate |
| Template refresh 중 기존 `.claude/settings.json` 덮어쓰기 | 사용자 설정 손실 | `.claude/settings.json`은 `.tmpl` 렌더링으로 생성되며, 이미 존재하는 경우 기존 deployer 로직(UserModified 보호)이 적용됨 |
| 사용자가 VCS commit 없이 실행 후 파일 손실 | 데이터 복구 불가 | "Before you start" 문서 섹션에서 "commit pending changes first" 강조. `moai migrate v2-to-v3` 시작 시 git status warning 자동 표시(optional 확장) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-MIG-001** (Versioned migration framework): 본 SPEC의 runner invocation, backup/rollback, dry-run, SIGINT handling은 모두 SPEC-V3-MIG-001에 의존.
- **SPEC-V3-MIG-002** — owns M01-M05 migration step Go implementations (including M02 agency archival and docs-site 4-locale sync, absorbed); this SPEC (MIGRATE-001) invokes MIG-002's runner via its public API.
- **SPEC-V3-SCH-001** (Formal config schemas): step 5(schema auto-repair)의 `moai doctor config --fix` 로직 제공.
- **SPEC-V3-SCH-002** (Settings source layering): schema auto-repair가 3-tier config를 인식해야 함.
- **SPEC-V3-HOOKS-001** (Hook Protocol v2): BC-001 breaking change 안내의 일부. Release CI에서 hook validation이 유효하려면 HOOKS-001 선행 완료 필요.

### 9.2 Blocks

- Phase 8 Release (v3.0.0 stable tag): 본 SPEC 완료 없이 v3.0.0 stable 태그를 release할 수 없음 (master §6.1 Phase 8).

### 9.3 Related

- **SPEC-AGENCY-ABSORB-001** (`moai migrate agency`): 동일한 `moai migrate` 최상위 커맨드 하위 subcommand 패턴. 두 subcommand가 공존하며 cobra 등록 구조를 공유.
- **SPEC-V3-CLN-001** — tooling dependency only (diagnostic commands `moai doctor template-drift` / `moai doctor skill-drift` used during migration UX; M01/M03/M04 Go file ownership lives in SPEC-V3-MIG-002).
- **SPEC-V3-CLN-002** — tooling dependency only (diagnostic command `moai doctor legacy-cleanup` and direct source edits used during migration UX; M05 Go file ownership lives in SPEC-V3-MIG-002).
- **SPEC-V3-AGT-001** (Agent frontmatter v2 bundle): Template refresh step에서 agent 파일의 v3 frontmatter schema 검증이 추가될 수 있음.
- **SPEC-V3-PLG-001** (Plugin system v1): Plugin manifests는 v3.0에서 새로 도입되므로, v2→v3 마이그레이션이 plugin 관련 변경을 요구하지 않음 (v2에 plugin 없음).
- **SPEC-V3-TEAM-001** (Teammate mailbox v2): Team mode 사용자는 BC-008 안내를 4-locale 문서에서 확인.

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 요구사항 ID는 `plan.md` 마일스톤(Wave 5)과 §6 Acceptance Criteria 시나리오로 역참조된다.
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-MIGRATE-001:REQ-MIGRATE-001-<NNN>` 주석 부착.
- 총 REQ 개수: 25개 (Ubiquitous 8, Event-Driven 5, State-Driven 3, Optional 2, Complex 7).
- 예상 코드 구현 경로:
  - `internal/cli/migrate_v2_to_v3.go` (cobra subcommand + orchestration)
  - `internal/cli/migrate_v2_to_v3_test.go` (단위 테스트)
  - `internal/cli/migrate_v2_to_v3_integration_test.go` (통합 테스트, v2.12 corpus 기반)
  - `internal/cli/migrate_v2_to_v3_windows_test.go` (Windows 전용)
  - `internal/cli/migrate_agency.go` 주변 cobra 등록부 수정 (subcommand 추가)
  - `docs-site/content/ko/migration/v3.md` (canonical)
  - `docs-site/content/en/migration/v3.md`, `ja/migration/v3.md`, `zh/migration/v3.md` (번역)
- Gap matrix 추적:
  - gm#149 (`CURRENT_MIGRATION_VERSION` counter) — SPEC-V3-MIG-001에 위임
  - gm#150 (idempotent migrations) — SPEC-V3-MIG-001에 위임
  - gm#151 (preAction auto-invocation) — **명시적으로 거부**(master §3.2 open question #10, conservative). 본 SPEC은 explicit 커맨드로만 실행.
- v3-master §5.1 (tool design), §5.2 (compat shims table), §5.3 (user-facing guide), §6.1 Phase 7 (tool + docs), §8.8 SPEC-V3-MIGRATE-001.
- Breaking change: BC-004 (Migration auto-run) — 본 SPEC은 auto-run을 "explicit command only"로 한정하며, 사용자가 `moai init` / `moai update` / `moai doctor` / `moai migrate`에서 migration runner를 의식적으로 호출해야 한다.
- Wave 1.5 source anchors: §5.1 (CC migration registry 패턴 참조), §5.2 (preAction invocation 패턴 — moai 채택하지 않음).
- CLAUDE.local.md anchors: §17.1 (URL blacklist), §17.2 (Markdown/Mermaid TD), §17.3 (4-locale SLA), §17.4 (버전 스냅샷 정책), §17.5 (manager-docs 위임), §17.6 (release checklist).

---

End of SPEC.
