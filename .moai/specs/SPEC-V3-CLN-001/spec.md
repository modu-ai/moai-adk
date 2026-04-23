---
id: SPEC-V3-CLN-001
title: Template Version Sync & Skill Drift Resolution
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: Wave 4 SPEC writer
priority: High
phase: "Phase 1 - Foundation (ships via M01/M03/M04)"
module: "internal/cli/update.go, internal/core/migration/steps/, internal/template/"
dependencies:
  - SPEC-V3-MIG-001
related_gap: [gm#183, gm#184, gm#185]
related_theme: "Theme 9 — Internal Cleanup & Template Drift Resolution"
breaking: false
bc_id: null
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

- `MigrationStep` 구현체 M01 (`m01_template_version_sync.go`): `project.yaml.template_version` ← `system.yaml.moai.version` 자동 설정
- `MigrationStep` 구현체 M03 (`m03_hook_wrapper_drift.go`): 누락된 hook wrapper를 `internal/template/templates/`에서 `.claude/hooks/moai/`로 배포 (템플릿 provenance = `TemplateManaged`)
- `MigrationStep` 구현체 M04 (`m04_skill_drift.go`): 누락된 skill directory를 `internal/template/templates/.claude/skills/`에서 `.claude/skills/`로 배포
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
- 과거 v2.x에서 삭제된 skill 파일의 archive 정책 — M02(SPEC-V3-CLN-003)가 담당

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/core/migration/steps/` 하위 신설 파일들
- Claude Code v2.1.111+
- 영향 디렉터리:
  - 신설: `internal/core/migration/steps/m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go`
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
The M01 migration step **shall** `project.yaml.template_version` 값을 읽고, `system.yaml.moai.version`과 비교하여 다를 경우 전자를 후자의 값으로 덮어쓴다.

**REQ-CLN-001-002**
The M01 migration step **shall** idempotent 동작을 보장하며, pre-condition check에서 `project.yaml.template_version == system.yaml.moai.version`일 경우 no-op으로 skip 한다.

**REQ-CLN-001-003**
The M03 migration step **shall** `internal/template/templates/.claude/hooks/moai/*.sh.tmpl`에 존재하되 `.claude/hooks/moai/*.sh`에 없는 모든 wrapper를 탐지하고, 존재하지 않는 파일만 배포한다 (`UserModified`, `UserCreated` 파일은 건드리지 않음).

**REQ-CLN-001-004**
The M04 migration step **shall** `internal/template/templates/.claude/skills/*/`에 존재하되 `.claude/skills/*/`에 없는 모든 skill directory를 탐지하고, 존재하지 않는 디렉터리만 배포한다.

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

상세 Given-When-Then 시나리오는 `acceptance.md` 참조 (본 SPEC의 Wave 4 scope에서는 spec.md만 생성).

핵심 기준:
- M01 실행 후 `.moai/config/sections/project.yaml:template_version` == `system.yaml:moai.version`.
- M01 재실행(두 번째 호출) 시 no-op으로 동작하고 migration_version counter가 증가하지 않는다.
- M03 실행 후 `.claude/hooks/moai/handle-permission-denied.sh`가 존재한다.
- M04 실행 후 `.claude/skills/moai-domain-db-docs/`, `moai-workflow-design-context/`, `moai-workflow-pencil-integration/`이 모두 존재한다.
- `moai doctor template-drift`가 드리프트 없는 상태에서 "no drift detected" 리포트 출력.
- `moai doctor template-drift`가 드리프트 발생 상태에서 pending 항목 수와 파일 경로를 출력.
- Rollback 시 `.moai/backups/<timestamp>/`에서 원본 복원 가능.
- User-modified wrapper가 M03에 의해 덮어쓰여지지 않는다.
- `make build` 성공 및 `go test ./...` 통과.
- Dry-run이 실제 파일 변경 없이 diff만 출력.

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
- SPEC-V3-CLN-003 (Agency archival + docs drift): M02와 본 SPEC은 독립적이지만 같은 runner cycle에서 실행.
- SPEC-V3-SCH-001 (Formal config schemas): `project.yaml`, `system.yaml` 스키마가 본 SPEC의 읽기/쓰기 대상이며, 추후 schema validation이 M01의 pre-condition에 통합될 수 있다.

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 요구사항 ID는 `plan.md` 마일스톤(Wave 5)과 `acceptance.md` 시나리오로 역참조된다.
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-CLN-001:REQ-CLN-001-<NNN>` 주석 부착.
- 총 REQ 개수: 23개 (Ubiquitous 8, Event-Driven 5, State-Driven 3, Optional 2, Complex 5).
- 예상 코드 구현 경로:
  - `internal/core/migration/steps/m01_template_version_sync.go`
  - `internal/core/migration/steps/m03_hook_wrapper_drift.go`
  - `internal/core/migration/steps/m04_skill_drift.go`
  - `internal/core/migration/steps/m01_test.go`, `m03_test.go`, `m04_test.go`
  - `internal/cli/doctor.go` 확장 (template-drift 서브커맨드)
  - `internal/cli/update.go` 수정 (migration runner 연동)
- Gap matrix 추적: gm#183 (template_version 드리프트), gm#184 (skill drift 3건), gm#185 (hook wrapper drift).
- v3-master §3.2 Theme 2 (Migration Framework), §3.9 Theme 9 (Internal Cleanup), §5.1 (`moai migrate v2-to-v3` tool이 M01-M05를 호출), §8.7 SPEC-V3-CLN-001.
- Wave 1.6 source anchors: §5.5 (hook wrapper count), §7.1 (skill count), §15.4 (template_version 드리프트 증거), §15.7 (hook wrapper drift 증거), §15.8 (template_version 12-버전 lag 증거).

---

End of SPEC.
