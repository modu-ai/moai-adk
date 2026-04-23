---
id: SPEC-V3-CLN-001
title: Template Version Sync & Skill Drift Resolution
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: Wave 4 SPEC writer
priority: P1 High
phase: "v3.0.0 — Phase 1 — Foundation (ships via M01/M03/M04)"
module: "internal/cli/update.go, internal/core/migration/steps/, internal/template/"
dependencies:
  - SPEC-V3-MIG-001
related_gap:
  - gm#183
  - gm#184
  - gm#185
related_theme: "Theme 9 — Internal Cleanup & Template Drift Resolution"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "cleanup, template-drift, version-sync, reliability, migration, v3"
---

# SPEC-V3-CLN-001: Template Version Sync & Skill Drift Resolution

## HISTORY

| Version | Date       | Author | Description                                                      |
|---------|------------|--------|------------------------------------------------------------------|
| 0.1.0   | 2026-04-22 | Wave 4 | Initial SPEC draft covering template drift resolution (M01/M03/M04) |

---

## 1. Goal (목적)

`moai update` 실행 시 `project.yaml.template_version`이 `system.yaml.moai.version`과 자동 동기화되도록 만들고, 템플릿 소스(`internal/template/templates/`)와 로컬 배포본(`.claude/skills/`, `.claude/hooks/moai/`) 간의 파일 드리프트를 탐지·해결하는 신뢰성 메커니즘을 도입한다. **이 SPEC은 사용자 가시 기능이 아니라 내부 일관성(reliability) 개선이다.** 구현은 SPEC-V3-MIG-001(versioned migration framework)이 정의한 MigrationStep 인터페이스를 사용하며, M01(template_version sync), M03(hook wrapper drift), M04(skill drift)로 실제 코드가 쪼개진다.

### 1.1 배경

Wave 1.6 §15.4, §15.8에 따르면 `.moai/config/sections/project.yaml:template_version: v2.7.22`가 `system.yaml:moai.version: v2.12.0`보다 약 12 minor 버전 뒤처져 있으며, 현재 `moai update`는 이 필드를 자동 동기화하지 않는다(gm#183). 또한 템플릿에는 있으나 로컬에는 배포되지 않은 skill 3개(`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`, Wave 1.6 §7.1, gm#184)와 missing hook wrapper(`handle-permission-denied.sh`, Wave 1.6 §5.5, gm#185)가 발견되었다. 이러한 드리프트는 silent 실패이므로 사용자가 인지할 수 없고, `moai doctor`도 탐지하지 못한다.

### 1.2 비목표 (Non-Goals)

- `moai update`의 기존 template refresh 로직의 대규모 재작성 (additive only)
- 템플릿 파일 내용의 의미론적 diff/merge (기존 3-way merge 로직 유지)
- Protected directories(`.moai/project/`, `.moai/specs/`)의 자동 덮어쓰기 (CLAUDE.local.md §2.1 준수)
- 새로운 skill/hook 추가를 위한 generator 도구 (SPEC-V3-PLG-001의 plugin scope)
- `project.yaml.template_version`을 사용자가 수동 고정(pin)하는 기능 (단일 source of truth 원칙)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `moai doctor template-drift` and `moai doctor skill-drift` CLI commands, diagnostic reporting, user-facing orchestration of template_version sync + skill catalog drift.

- Tooling layer: `moai doctor template-drift` CLI + orchestration of M01/M03/M04 (migration step Go files owned by SPEC-V3-MIG-002)
- `moai doctor template-drift` 신규 서브커맨드: 현재 드리프트 상태를 리포트 (dry-run-like)
- `moai update` 통합: v3.0부터 `moai update` 실행 시 migration runner (SPEC-V3-MIG-001)가 pending M01/M03/M04를 자동 적용
- 각 migration step의 pre-condition, dry-run diff, idempotency, rollback 구현
- `.moai/backups/<ISO-8601-timestamp>/` 백업 (rollback 경로)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 사용자가 의도적으로 수정한 `UserModified` skill/hook wrapper 파일의 자동 덮어쓰기 — 기존 deployer의 provenance 로직(template deployer:121-137) 준수
- 삭제된 skill(사용자가 일부러 제거한 경우) 자동 복구 — 사용자가 명시적으로 `--force`를 전달해야 함 (deferred to SPEC-V3-MIG-001)
- Skill 버전 diff / 내용 검증(content hash) — Wave 1.6 §4 manifest.Hasher는 사용하나, 내용 품질 검사는 범위 밖
- `project.yaml`의 다른 필드(예: `name`, `mode`) 자동 수정
- `~/.moai/config/` 사용자 레벨 설정의 수정 (§3.3 settings source layering은 SPEC-V3-SCH-002가 담당)
- 과거 v2.x에서 삭제된 skill 파일의 archive 정책 — M02(MIG-002 (M02 step))가 담당
- **Migration step Go implementations** in `internal/core/migration/steps/` — owned by SPEC-V3-MIG-002. CLN-001 calls into MIG-002's migrate functions but does not implement them.

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/core/migration/steps/` 하위 신설 파일들
- Claude Code v2.1.111+
- 영향 디렉터리:
  - 참조: `internal/core/migration/steps/m01_*.go`, `m03_*.go`, `m04_*.go` (owned by SPEC-V3-MIG-002)
  - 수정: `internal/cli/update.go` (migration runner trigger), `internal/cli/doctor.go` (template-drift 서브커맨드)
  - 읽기 전용 참조: `internal/template/embed.go`, `internal/template/deployer.go`, `internal/manifest/manager.go`
- 템플릿 소스 기준: `internal/template/templates/.claude/skills/`에 50개 skill (Wave 1.6 §7.1), `internal/template/templates/.claude/hooks/moai/`에 27개 wrapper
- 로컬 현재 상태: skills 47개 (3개 누락), hook wrappers 26개 (1개 누락)
- 외부 레퍼런스: Wave 1.6 §15.1, §15.4, §15.7, §15.8, gm#183-#185, master §3.2, §3.9

---

## 4. Assumptions (가정)

- SPEC-V3-MIG-001의 MigrationStep 인터페이스와 Runner가 선행 구현되어 있다 (Phase 1 ordering).
- `internal/template/embed.go`의 `go:embed all:templates`가 모든 템플릿 파일을 포함한다 (Wave 1.6 §4.1 evidence).
- 사용자가 의도적으로 skill이나 hook wrapper를 삭제한 경우는 rare edge case로, default 정책은 "템플릿 상태로 복구"이다.
- `project.yaml.template_version`은 사용자가 수동 pin하지 않는다 (단일 source of truth 가정).
- `.moai/backups/`에 쓰기 권한이 있고 디스크 여유 공간이 원본 크기의 최소 2배이다.
- `system.yaml.moai.version`은 릴리스 프로세스에서 정확히 유지된다 (ground truth).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-CLN-001-001**
The `moai doctor template-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M01 migration step (template_version sync) and report its diagnostic status to the user.

**REQ-CLN-001-002**
The `moai doctor template-drift` subcommand **shall** treat M01 as idempotent and surface "no drift" when pre-conditions are not met (i.e., `project.yaml.template_version == system.yaml.moai.version`).

**REQ-CLN-001-003**
The `moai doctor template-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M03 migration step (hook wrapper drift) and report its diagnostic status — identifying missing wrappers and pending deployments.

**REQ-CLN-001-004**
The `moai doctor skill-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M04 migration step (skill drift) and report its diagnostic status — enumerating missing skill directories and pending deployments.

**REQ-CLN-001-005**
The `moai doctor template-drift` 서브커맨드 **shall** 현재 프로젝트의 드리프트 상태(누락된 skill 목록, 누락된 hook wrapper 목록, template_version 불일치)를 구조화된 리포트로 출력하고, pending migration step을 안내한다.

**REQ-CLN-001-006**
Each migration step (M01, M03, M04) **shall** SPEC-V3-MIG-001의 `MigrationStep` 인터페이스를 구현하며 `Version()`, `ID()`, `Description()`, `IsIdempotent()`, `PreConditionsMet()`, `DryRun()`, `Apply()`, `Rollback()` 메서드를 모두 제공한다.

**REQ-CLN-001-007**
Each migration step **shall** Apply 실행 전 영향받는 파일을 `.moai/backups/<ISO-8601-timestamp>/`에 snapshot 하며, Rollback 시 이 스냅샷을 사용해 원복한다.

**REQ-CLN-001-008**
The M04 skill drift migration **shall** 템플릿의 skill directory 트리 전체(하위 `SKILL.md`, `modules/`, `references/` 등 모든 bundled resources 포함)를 재귀적으로 복사하며, file permission bits를 REQ-MIGRATE-012a/b(SPEC-AGENCY-ABSORB-001) 패턴에 따라 plaform-aware하게 처리한다.

### 5.2 Event-Driven Requirements

**REQ-CLN-001-009**
**When** 사용자가 `moai update`를 실행하면, the 시스템 **shall** template refresh 직후 SPEC-V3-MIG-001 migration runner를 invoke하여 pending M01/M03/M04를 순차 실행하며, dry-run preview를 기본 동작으로 제시한다.

**REQ-CLN-001-010**
**When** M01이 `project.yaml.template_version`을 변경하면, the step **shall** 변경 전후 값을 stdout에 표시한다 (예: `template_version: v2.7.22 → v2.12.0`).

**REQ-CLN-001-011**
**When** M03 또는 M04가 파일을 배포하면, the step **shall** 각 배포된 파일 경로를 stdout에 "Deployed: <relative-path>" 형식으로 기록하고, 누적 카운트를 요약에 포함한다.

**REQ-CLN-001-012**
**When** `moai doctor template-drift`가 드리프트를 발견하면, the 서브커맨드 **shall** exit code 0으로 종료하지만 stdout에 pending 항목을 "PENDING:" 접두사로 강조한다 (doctor가 리포트용이지 실패 채널이 아니므로).

**REQ-CLN-001-013**
**When** M01/M03/M04 중 하나라도 apply 실패하면, the runner **shall** 해당 step의 Rollback()을 호출하고, 이미 성공한 이전 step들의 변경 사항은 유지한다 (individual step transactional, 전체 chain은 atomic하지 않음).

### 5.3 State-Driven Requirements

**REQ-CLN-001-014**
**While** `.claude/skills/<skill-name>/`이 이미 존재하는 상태에서, the M04 step **shall** 해당 skill 디렉터리를 건드리지 않고 다음 skill로 진행한다 (partial skill 배포는 수동 개입 대상).

**REQ-CLN-001-015**
**While** `.claude/hooks/moai/<wrapper>.sh`의 manifest provenance가 `UserModified`인 상태에서, the M03 step **shall** 해당 wrapper를 덮어쓰지 않고 stderr에 경고 "User-modified wrapper preserved: <path>"를 emit한다.

**REQ-CLN-001-016**
**While** migration runner가 dry-run 모드로 실행 중인 상태에서, the M01/M03/M04 step **shall** 실제 파일 수정을 수행하지 않고 예상 변경만 리포트한다.

### 5.4 Optional Requirements

**REQ-CLN-001-017**
**Where** 사용자가 `--only M01` / `--only M03` / `--only M04` 플래그를 `moai migrate` 커맨드에 전달한 환경에서, the runner **shall** 지정된 step만 실행하고 나머지는 skip 한다 (SPEC-V3-MIG-001 CLI 연동).

**REQ-CLN-001-018**
**Where** 사용자가 `.moai/config/sections/system.yaml`에 `migration.skip_steps: [M04]`를 명시한 환경에서, the runner **shall** 해당 step을 영구 skip한다 (자동 복구가 원치 않는 고급 사용자용 옵트아웃).

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-CLN-001-019 (Unwanted Behavior)**
**If** `system.yaml.moai.version`이 비어 있거나 semver 형식을 따르지 않으면, **then** M01 step **shall** apply를 거부하고 `MIGRATE_CLN001_INVALID_VERSION` 오류를 반환한다 (rollback 불필요, apply 전 중단).

**REQ-CLN-001-020 (Unwanted Behavior)**
**If** `project.yaml` 파일이 존재하지 않으면, **then** M01 step **shall** skip 하고 stderr에 "project.yaml not found; skipping template_version sync"를 emit한다 (`moai init`이 선행되지 않은 프로젝트에서 오류 방지).

**REQ-CLN-001-021 (Complex: State + Event)**
**While** M04가 실행 중이고, **when** 복사하려는 skill 디렉터리의 총 크기가 100MB를 초과하면, the step **shall** apply를 중단하고 `MIGRATE_CLN001_SKILL_TOO_LARGE` 오류를 반환한다 (비정상 템플릿 탐지 safeguard).

**REQ-CLN-001-022 (Unwanted Behavior)**
**If** M03/M04 apply 중 `.moai/backups/` 쓰기가 실패하면 (권한, 디스크), **then** 해당 step **shall** apply를 시작하기 전에 중단하고 `MIGRATE_CLN001_BACKUP_FAILED` 오류를 반환한다 (rollback path 없이 변경 없음 보장).

**REQ-CLN-001-023 (Ubiquitous, cross-step consistency)**
After all of M01/M03/M04 complete (either successfully or rolled-back), the runner **shall** `.moai/reports/migration-v3-cln001-<timestamp>.md` 리포트를 생성하고 각 step의 status (applied, rolled-back, skipped), affected files, before/after diff summary를 포함한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-CLN-001-01**: Given project.yaml의 `template_version: v2.7.22`, system.yaml의 `moai.version: v2.12.0` 상태 When M01 Apply 실행 Then project.yaml의 `template_version`이 `v2.12.0`으로 덮어써지고 system.yaml과 일치 (maps REQ-CLN-001-001).
- **AC-CLN-001-02**: Given M01이 이미 적용되어 project.template_version == system.moai.version인 상태 When M01 재실행 Then PreConditionsMet=false로 skip되고 migration_version counter가 증가하지 않음 (maps REQ-CLN-001-002).
- **AC-CLN-001-03**: Given `.claude/hooks/moai/handle-permission-denied.sh` 부재 상태 When M03 Apply 실행 Then `.claude/hooks/moai/handle-permission-denied.sh` 파일이 템플릿으로부터 배포됨 (maps REQ-CLN-001-003).
- **AC-CLN-001-04**: Given `.claude/skills/`에 `moai-domain-db-docs/`, `moai-workflow-design-context/`, `moai-workflow-pencil-integration/` 3개 skill 부재 상태 When M04 Apply 실행 Then 3개 skill 디렉터리 모두 배포됨 (maps REQ-CLN-001-004).
- **AC-CLN-001-05**: Given 드리프트가 없는 프로젝트 When `moai doctor template-drift` 실행 Then exit code 0과 "no drift detected" 리포트 출력 (maps REQ-CLN-001-005, REQ-CLN-001-012).
- **AC-CLN-001-06**: Given 드리프트가 존재하는 프로젝트 When `moai doctor template-drift` 실행 Then pending 항목 수와 각 항목의 파일 경로가 "PENDING:" 접두사로 출력됨 (maps REQ-CLN-001-005, REQ-CLN-001-012).
- **AC-CLN-001-07**: Given M01/M03/M04가 Apply되고 backup이 `.moai/backups/<timestamp>/`에 snapshot된 상태 When `moai migrate --rollback <timestamp>` 실행 Then 원본 파일들이 복원됨 (maps REQ-CLN-001-007).
- **AC-CLN-001-08**: Given `.claude/hooks/moai/handle-permission-denied.sh`의 manifest provenance가 `UserModified`인 상태 When M03 Apply 실행 Then 해당 파일은 건드려지지 않고 stderr에 경고가 emit됨 (maps REQ-CLN-001-015).
- **AC-CLN-001-09**: Given 본 SPEC의 모든 구현 파일이 작성된 상태 When `make build && go test ./...` 실행 Then 전체 빌드/테스트 통과 (maps REQ-CLN-001-006).
- **AC-CLN-001-10**: Given migration runner가 `--dry-run` 모드로 호출된 상태 When M01/M03/M04가 평가됨 Then 실제 파일 변경 없이 예상 diff만 stdout에 출력됨 (maps REQ-CLN-001-016).
- **AC-CLN-001-11**: Given 템플릿 skill 디렉터리가 bundled resources (`modules/`, `references/`, `SKILL.md`)를 포함한 상태 When M04 Apply가 실행됨 Then 전체 트리가 재귀 복사되고 permission bits가 platform-aware 로직으로 적용됨 (maps REQ-CLN-001-008).
- **AC-CLN-001-12**: Given `moai update`가 실행된 상태 When template refresh가 완료됨 Then 시스템이 SPEC-V3-MIG-001 runner를 자동 invoke하고 pending M01/M03/M04를 dry-run preview 형태로 제시함 (maps REQ-CLN-001-009).
- **AC-CLN-001-13**: Given `project.yaml.template_version: v2.7.22`, `system.yaml.moai.version: v2.12.0` 상태 When M01 Apply 수행 Then stdout에 `template_version: v2.7.22 → v2.12.0` 라인이 출력됨 (maps REQ-CLN-001-010).
- **AC-CLN-001-14**: Given M03 또는 M04가 각각 2개의 파일을 배포한 상태 When Apply 완료 Then stdout에 "Deployed: <relative-path>" 라인이 파일별로 기록되고 누적 카운트가 최종 요약에 포함됨 (maps REQ-CLN-001-011).
- **AC-CLN-001-15**: Given M03 Apply 중 `.moai/backups/` 쓰기가 실패하는 상태 When step이 시작됨 Then apply 시작 전 중단되고 `MIGRATE_CLN001_BACKUP_FAILED` 오류가 반환되며 파일 변경 없음 (maps REQ-CLN-001-013, REQ-CLN-001-022).
- **AC-CLN-001-16**: Given `.claude/skills/foo/`가 이미 존재하는 상태 When M04가 실행됨 Then 해당 skill 디렉터리는 건드려지지 않고 다음 skill로 진행됨 (maps REQ-CLN-001-014).
- **AC-CLN-001-17**: Given 사용자가 `moai migrate --only M03`을 실행한 상태 When runner가 평가됨 Then M03만 실행되고 M01/M04는 skip됨 (maps REQ-CLN-001-017).
- **AC-CLN-001-18**: Given `.moai/config/sections/system.yaml`에 `migration.skip_steps: [M04]` 설정된 상태 When runner가 실행됨 Then M04가 영구 skip됨 (maps REQ-CLN-001-018).
- **AC-CLN-001-19**: Given `system.yaml.moai.version`이 빈 문자열이거나 semver 형식 위반 상태 When M01 Apply 시도 Then `MIGRATE_CLN001_INVALID_VERSION` 오류를 반환하고 apply 전 중단됨 (maps REQ-CLN-001-019).
- **AC-CLN-001-20**: Given `project.yaml` 파일 부재 상태 When M01 Apply 시도 Then step이 skip되고 stderr에 "project.yaml not found; skipping template_version sync" 메시지 emit (maps REQ-CLN-001-020).
- **AC-CLN-001-21**: Given M04 대상 skill 디렉터리 총 크기 > 100MB 상태 When M04 Apply 시도 Then `MIGRATE_CLN001_SKILL_TOO_LARGE` 오류를 반환하고 중단됨 (maps REQ-CLN-001-021).
- **AC-CLN-001-22**: Given M03 Apply 중 `.moai/backups/` 쓰기 권한 부족 상태 When step 시작 Then apply 전 중단되고 `MIGRATE_CLN001_BACKUP_FAILED` 오류 반환, 파일 변경 없음 (maps REQ-CLN-001-022).
- **AC-CLN-001-23**: Given M01/M03/M04 모두 실행 완료(성공 또는 rolled-back) 상태 When runner가 종료됨 Then `.moai/reports/migration-v3-cln001-<timestamp>.md` 리포트가 생성되고 step별 status/affected files/diff summary가 포함됨 (maps REQ-CLN-001-023).

---

## 7. Constraints (제약)

- 본 SPEC은 SPEC-V3-MIG-001의 MigrationStep 인터페이스를 **구현**만 하고, migration framework의 수정은 포함하지 않는다.
- Protected directories 정책 준수(CLAUDE.local.md §2.1): `.moai/project/`, `.moai/specs/`는 절대 건드리지 않는다.
- Template-First 원칙(CLAUDE.local.md §2.2): 새로운 skill/hook wrapper 추가는 `internal/template/templates/`가 우선이며, M04/M03은 **deployment**만 담당한다.
- 9-direct-dep 정책 준수: 신규 외부 의존성 금지.
- 언어 중립성(CLAUDE.local.md §15): 복사되는 skill 템플릿은 특정 언어에 편향되지 않는다.
- 롤백은 파일 단위 restore만 지원 (`.moai/backups/`에 snapshot으로 보존).
- 플랫폼 호환: Windows에서 POSIX permission bits는 no-op 처리 (SPEC-AGENCY-ABSORB-001 REQ-MIGRATE-012a/b 패턴 재사용).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| User가 의도적으로 삭제한 skill이 M04에 의해 자동 복구 | UX 저하 (의도 위반) | REQ-CLN-001-018의 `migration.skip_steps: [M04]` 옵트아웃 제공. 첫 실행 시 "this will deploy missing skills; run with --dry-run to preview" 안내 |
| `project.yaml`의 `template_version`을 사용자가 수동 pin한 경우 덮어쓰기 | 의도 파괴 | Non-Goal로 명시(§1.2). 수동 pin 사용 케이스는 §9.5 open question으로 에스컬레이션 |
| 큰 skill 디렉터리(미래에 bundled binaries 포함) 복사 시간 | 성능 저하 | REQ-CLN-001-021의 100MB safeguard |
| `.moai/backups/` 디스크 가득 | Apply 실패 | REQ-CLN-001-022의 사전 검증. 추가로 SPEC-V3-MIG-001이 글로벌 디스크 검사 책임 |
| `internal/template/embed.go`의 `go:embed`가 새로 추가된 skill을 포함하지 않음 (빌드 실수) | 드리프트 영구화 | `make build` 후 CI에서 embedded 파일 카운트 검증. M04 실행 시 embedded vs local 비교로 탐지 |
| User-modified wrapper가 template 변경 누락으로 drift 지속 | 일관성 저하 | `moai doctor template-drift`가 UserModified 파일도 리포트 (information only) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-MIG-001 (Versioned migration framework): 본 SPEC의 M01/M03/M04는 SPEC-V3-MIG-001의 MigrationStep 인터페이스, Runner, `.moai/backups/`, dry-run 지원, rollback에 전적으로 의존한다. SPEC-V3-MIG-001이 먼저 완료되어야 한다.

### 9.2 Blocks

- SPEC-V3-MIGRATE-001 (`moai migrate v2-to-v3` tool): v2→v3 전환 도구가 M01/M03/M04를 실행해야 하므로 본 SPEC이 선행되어야 한다.

### 9.3 Related

- SPEC-V3-CLN-002 (Legacy code removal): M05와 본 SPEC의 M01/M03/M04는 같은 migration runner에서 순차 실행됨.
- SPEC-V3-MIG-002 (M02 step) (Agency archival + docs drift): M02와 본 SPEC은 독립적이지만 같은 runner cycle에서 실행.
- SPEC-V3-SCH-001 (Formal config schemas): `project.yaml`, `system.yaml` 스키마가 본 SPEC의 읽기/쓰기 대상이며, 추후 schema validation이 M01의 pre-condition에 통합될 수 있다.

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 요구사항 ID는 `plan.md` 마일스톤(Wave 5)과 §6 Acceptance Criteria 시나리오로 역참조된다.
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-CLN-001:REQ-CLN-001-{REQ_NUM}` 주석 부착 (REQ_NUM은 구체 REQ ID 치환).
- 총 REQ 개수: 23개 (Ubiquitous 8, Event-Driven 5, State-Driven 3, Optional 2, Complex 5).
- 예상 코드 구현 경로:
  - `internal/cli/doctor_template_drift.go`
  - `internal/cli/doctor_skill_drift.go`
  - `internal/cli/doctor.go` 확장 (template-drift / skill-drift 서브커맨드 등록)
  - `internal/cli/update.go` 수정 (migration runner 연동)
- Gap matrix 추적: gm#183 (template_version 드리프트), gm#184 (skill drift 3건), gm#185 (hook wrapper drift).
- v3-master §3.2 Theme 2 (Migration Framework), §3.9 Theme 9 (Internal Cleanup), §5.1 (`moai migrate v2-to-v3` tool이 M01-M05를 호출), §8.7 SPEC-V3-CLN-001.
- Wave 1.6 source anchors: §5.5 (hook wrapper count), §7.1 (skill count), §15.4 (template_version 드리프트 증거), §15.7 (hook wrapper drift 증거), §15.8 (template_version 12-버전 lag 증거).

---

End of SPEC.
