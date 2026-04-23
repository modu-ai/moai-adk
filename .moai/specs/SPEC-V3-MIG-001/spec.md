---
id: SPEC-V3-MIG-001
title: Versioned Migration Framework (CURRENT_MIGRATION_VERSION counter + ordered runner)
version: 0.1.0
status: draft
created: 2026-04-22
updated: 2026-04-22
author: manager-spec
priority: Critical
phase: "Phase 1 — Foundation"
module: "internal/core/migration/"
dependencies:
  - SPEC-V3-SCH-001
related_gap: [149, 150, 151, 152, 153, 154, 155]
related_theme: "Theme 2 — Migration Framework"
breaking: true
bc_id: [BC-004]
lifecycle: spec-anchored
tags: "v3, migration, framework, versioned, idempotent, foundation, P0"
---

# SPEC-V3-MIG-001: Versioned Migration Framework

## HISTORY

- 2026-04-22 v0.1.0: 최초 작성. master-v3 §3.2 Theme 2, gap-matrix #149-#155 (Critical), W1.5 §5.1-§5.4, W1.2 §7.1-§7.6 근거. 모든 v3 release의 breaking change migration 인프라. master-v3 BC-004 기반.

---

## 1. Goal (목적)

moai-adk-go는 현재 단 하나의 one-shot migration만 보유한다: `moai migrate agency` (SPEC-AGENCY-ABSORB-001, `internal/cli/migrate_agency.go:569`). 이에 반해 Claude Code는 `CURRENT_MIGRATION_VERSION=11` 카운터 + ordered 순차 runner + `preAction` hook 자동 발사 (W1.5 §5.1-§5.2)를 통해 10+ migration을 무중단 관리한다.

104 SPEC / 50 skills / 22 agents / 22 YAML section 규모에 도달한 moai는 ad-hoc 스크립트 방식을 유지할 수 없다. W1.6 §15.4가 보고한 `project.yaml.template_version: v2.7.22` (12 minor 지연)는 프레임워크 부재의 직접 증거. 본 SPEC은 CC의 프레임워크를 Go-idiomatic하게 포팅해 모든 v3 이후 release의 breaking change migration 기반을 수립한다.

### 1.1 배경

W1.5 §5.1 `runMigrations()` 구조:
```typescript
const CURRENT_MIGRATION_VERSION = 11;
if (getGlobalConfig().migrationVersion !== CURRENT_MIGRATION_VERSION) {
  migrateAutoUpdatesToSettings();
  migrateBypassPermissionsAcceptedToSettings();
  // ... 10 more migrations in order
  saveGlobalConfig(prev => ({...prev, migrationVersion: CURRENT_MIGRATION_VERSION}));
}
```

Invocation (W1.5 §5.2): Commander `preAction` hook — 모든 CLI 실행 전 자동 발사.

gap-matrix evidence:
- #149 (Critical, L): 카운터 + runner 부재
- #150 (High, M): idempotency 부분적
- #151 (High, S): preAction auto-fire 미지원 — 사용자 explicit `moai migrate` 필요
- #152 (Medium): source specificity (user/project/local)
- #153 (Low): async fire-and-forget (e.g., `migrateChangelogFromConfig`)
- #154 (Low): analytics events (moai는 N/A)
- #155 (Low): timestamp-based notifications

본 SPEC은 **#149-#151 핵심 3개**에 집중, #152는 SPEC-V3-SCH-002 tier-awareness로 해결, #153 optional, #154 N/A, #155 optional.

### 1.2 비목표 (Non-Goals)

- **구체 마이그레이션 구현 금지**: M01-M05 concrete content는 **SPEC-V3-MIG-002** 소관. 본 SPEC은 프레임워크만.
- **Agency migration 리팩토링 금지**: 기존 `moai migrate agency`는 유지. v3 프레임워크에 포함될 때 호환 래퍼로 감싸거나 별도 유지.
- **Auto-rollback 금지**: 각 step에 Rollback() 제공하되 자동 감지형 rollback은 없음 (CC와 일치).
- **Cross-platform consistency over perfection**: Windows에서 `~/.moai/backups/` 경로는 `%USERPROFILE%\.moai\backups\`로 해석 — OS 차이 수용.
- **Analytics 금지**: moai는 telemetry opt-in이며 이벤트 emit 안 함.
- **Async migration 금지 (v3.0)**: fire-and-forget은 디버깅 난이도 상승. 모든 v3.0 migration은 동기.

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/core/migration/` 신규 패키지: `runner.go`, `context.go`, `step.go`, `registry.go`, `backup.go`
- `MigrationStep` 인터페이스 (8 메서드: Version/ID/Description/IsIdempotent/PreConditionsMet/DryRun/Apply/Rollback)
- `Runner` 타입: 등록된 step을 `Version()` 오름차순 정렬 후 pending 필터 → dry-run/apply 실행
- `CURRENT_MIGRATION_VERSION` 카운터: `.moai/config/sections/system.yaml`의 `migration_version: N` 필드
- `preAction` equivalent: 4개 선별된 cobra 명령(`moai init`, `moai update`, `moai doctor`, `moai migrate`)에 PersistentPreRunE 훅 부착 (conservative trigger list)
- `MOAI_DISABLE_MIGRATIONS=1` 환경변수 opt-out (master-v3 BC-004 escape)
- Backup snapshot: `.moai/backups/{ISO-8601-timestamp}/` 경로에 영향 파일 복사
- `moai migrate` subcommand 확장: `--dry-run`, `--yes`, `--only <step-id>`, `--rollback <timestamp>`, `--no-backup`
- Post-migration report: `.moai/reports/migration-{timestamp}.md` with before/after diff 요약
- `moai doctor migration --validate`: 등록된 step 인벤토리, pending 개수, 마지막 실행 이력
- Step 등록 방식: Go `init()` 함수에서 registry에 자체 등록 (registry.Register(step))

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 구체 마이그레이션 (M01-M05) — SPEC-V3-MIG-002
- Auto-rollback on failure detection — manual rollback only
- Parallel migration execution — 순차만 (ordering 안전성)
- Remote migration script download — local registered only
- Plugin-authored migrations — SPEC-V3-PLG-001 이후 고려
- Agent/Skill frontmatter migrations — SPEC-V3-AGT-001/SKL-001에서 별도
- Migration UI (GUI tool) — CLI only
- Analytics / telemetry event emission — out of scope
- Async / fire-and-forget migrations — v3.0에선 동기만

---

## 3. Environment (환경)

- 런타임: Go 1.26, moai-adk-go v3.0.0-alpha.1+
- 의존성: SPEC-V3-SCH-001 (schema + registry)
- Cobra: `PersistentPreRunE` hook 사용 (이미 사용 중, `internal/cli/worktree.go`)
- YAML: `gopkg.in/yaml.v3` (기존)
- Time: `time.Now().UTC().Format(time.RFC3339)` → backup dir name
- File ops: `internal/template/manifest` (기존 manifest system 활용 가능)
- 영향 디렉터리:
  - `internal/core/migration/` (신규)
  - `internal/core/migration/steps/` (SPEC-V3-MIG-002에서 M01-M05 배치)
  - `internal/cli/deps.go` (runner 등록)
  - `internal/cli/migrate.go` (기존 migrate_agency.go 옆에 v3 프레임워크 추가)
  - `internal/cli/{init,update,doctor}.go` (PersistentPreRunE hook)
  - `.moai/config/sections/system.yaml` (migration_version 필드 추가)
  - `.moai/backups/` (runtime, gitignored)
  - `.moai/reports/` (runtime, gitignored)

---

## 4. Assumptions (가정)

- A-001 (High): 모든 migration step은 순차 (sorted by Version()) 실행 시 이전 step의 결과물을 전제로 한다. 의존성 DAG는 선형으로 충분.
- A-002 (High): `system.yaml.migration_version` 필드는 신규이며 기본값 0. 미존재 시 모든 step을 pending으로 간주 후 순차 실행.
- A-003 (High): Idempotency는 `PreConditionsMet()` 메서드로 보장. "이미 적용됨" 상태는 skip (에러 아님).
- A-004 (Medium): Dry-run 모드에서 `Apply()`는 호출되지 않으며 `DryRun()`만 호출됨. `DryRun()`은 side-effect가 없어야 한다.
- A-005 (High): Backup 스냅샷은 touched 파일만 저장 (not entire project). 복구 시 동일 경로로 복원.
- A-006 (Medium): 사용자는 migration을 skip하고 싶을 때 `MOAI_DISABLE_MIGRATIONS=1`을 설정한다. CI/airgap 환경 주 고객.
- A-007 (High): Rollback은 backup 스냅샷을 단순 파일 복원으로 수행. 복잡한 semantic reverse 로직 없음.
- A-008 (Medium): `preAction` equivalent은 Cobra의 `PersistentPreRunE`로 4개 명령에 국한. `moai version`, `moai status` 등은 side-effect-free 유지 (master-v3 §9 open question #10).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MIG-001-001 (Ubiquitous)**
시스템은 **항상** `internal/core/migration/step.go`에 `MigrationStep` 인터페이스를 정의한다:
```go
type MigrationStep interface {
    Version() int
    ID() string
    Description() string
    IsIdempotent() bool
    PreConditionsMet(*MigrationContext) (bool, error)
    DryRun(*MigrationContext) (MigrationDiff, error)
    Apply(*MigrationContext) error
    Rollback(*MigrationContext) error
}
```

**REQ-MIG-001-002 (Ubiquitous)**
시스템은 **항상** `Version()`이 단조 증가 정수(1부터)이며, 전역 레지스트리 내에서 유일함을 보증한다.

**REQ-MIG-001-003 (Ubiquitous)**
시스템은 **항상** `.moai/config/sections/system.yaml`에 `migration_version: N` 필드를 유지하며, 적용된 마지막 step의 Version()을 기록한다.

**REQ-MIG-001-004 (Ubiquitous)**
시스템은 **항상** Runner 실행 전 `.moai/backups/{ISO-8601-UTC-timestamp}/` 디렉터리를 생성하고, 영향받을 파일의 백업을 저장한다.

**REQ-MIG-001-005 (Ubiquitous)**
시스템은 **항상** Post-migration report를 `.moai/reports/migration-{timestamp}.md`에 markdown으로 기록한다 (applied step list + diff summary).

**REQ-MIG-001-006 (Ubiquitous)**
시스템은 **항상** step 등록을 Go `init()` 함수 + `registry.Register(step)` 패턴으로 수행한다 (컴파일 타임 결정).

### 5.2 Event-Driven Requirements

**REQ-MIG-001-010 (Event-Driven)**
**When** `moai init`, `moai update`, `moai doctor`, 또는 `moai migrate` 중 하나가 실행되면, the 시스템 **shall** Cobra `PersistentPreRunE`를 통해 Runner를 호출하고 pending step 개수를 사용자에게 표시한다.

**REQ-MIG-001-011 (Event-Driven)**
**When** pending step이 1개 이상 존재하면, the 시스템 **shall** 사용자에게 dry-run 결과를 표시하고 `--yes` 플래그 없이는 확인을 요구한다 (interactive 환경) 또는 중단(non-interactive).

**REQ-MIG-001-012 (Event-Driven)**
**When** step의 `Apply()`가 에러를 반환하면, the 시스템 **shall** 즉시 `Rollback()`을 호출하고, 실패 이후 step은 실행하지 않으며, RunReport를 생성해 에러를 포함한다.

**REQ-MIG-001-013 (Event-Driven)**
**When** 모든 pending step이 성공하면, the 시스템 **shall** `system.yaml.migration_version`을 마지막 step의 Version()으로 갱신한다.

**REQ-MIG-001-014 (Event-Driven)**
**When** 사용자가 `moai migrate --rollback <timestamp>`를 실행하면, the 시스템 **shall** 해당 timestamp 백업 디렉터리에서 파일을 복원하고, `system.yaml.migration_version`을 백업 시점 값으로 되돌린다.

**REQ-MIG-001-015 (Event-Driven)**
**When** 사용자가 `moai migrate --only <step-id>`를 실행하면, the 시스템 **shall** 지정된 step만 (선행 step 무시) pre-condition 확인 후 실행한다.

**REQ-MIG-001-016 (Event-Driven)**
**When** 사용자가 `moai doctor migration --validate`를 실행하면, the 시스템 **shall** 등록된 step inventory, 현재 `migration_version`, pending 개수, 최근 실행 리포트 링크를 표시한다.

### 5.3 State-Driven Requirements

**REQ-MIG-001-020 (State-Driven)**
**While** `MOAI_DISABLE_MIGRATIONS=1` 환경변수가 설정된 동안, the 시스템 **shall** Runner를 호출하지 않으며, pending step 존재 시 "Migrations disabled by MOAI_DISABLE_MIGRATIONS" 경고를 한 번 표시한다.

**REQ-MIG-001-021 (State-Driven)**
**While** dry-run 모드(`--dry-run`)가 활성화된 동안, the 시스템 **shall** `Apply()` / 파일 쓰기를 호출하지 않으며, `DryRun()` 결과만 표시한다.

**REQ-MIG-001-022 (State-Driven)**
**While** interactive 환경에서 `--yes` 없이 Runner가 실행되는 동안, the 시스템 **shall** 모든 pending step의 dry-run 결과 표시 후 사용자 확인을 대기한다 (Claude Code 환경에서는 AskUserQuestion 위임 — subagent 제약 준수).

**REQ-MIG-001-023 (State-Driven)**
**While** fresh project (첫 `moai init`)인 동안, the 시스템 **shall** `migration_version`을 현재 레지스트리 최대 Version() 값으로 초기화하고 step을 실행하지 않는다 (새 프로젝트는 이미 "최신" 상태).

**REQ-MIG-001-024 (State-Driven)**
**While** step의 `IsIdempotent()`가 true이고 `PreConditionsMet()`이 false인 동안, the 시스템 **shall** 해당 step을 skip으로 마킹하고 다음 step으로 진행한다 (idempotent 완료로 간주).

### 5.4 Optional Requirements

**REQ-MIG-001-030 (Optional)**
**Where** 사용자가 `--no-backup` 플래그를 지정하면, the 시스템 **shall** backup 스냅샷 생성을 생략한다 (NOT RECOMMENDED 경고 표시).

**REQ-MIG-001-031 (Optional)**
**Where** 사용자가 `moai migrate --json`을 실행하면, the 시스템 **shall** RunReport를 JSON 형식으로 출력해 CI 파이프라인 통합을 지원한다.

**REQ-MIG-001-032 (Optional)**
**Where** step이 `PostApplyHook(ctx) error`를 구현하면 (optional 메서드), the 시스템 **shall** `Apply()` 성공 후 해당 hook을 호출한다 (e.g., touched file re-validation).

**REQ-MIG-001-033 (Optional)**
**Where** backup 스냅샷이 30일 이상 오래된 경우, the 시스템 **shall** `moai doctor migration --cleanup` 서브명령으로 정리 옵션을 제공한다.

### 5.5 Unwanted Behavior (Must Not)

**REQ-MIG-001-040 (Unwanted)**
시스템은 pending step을 실행하기 전에 backup 스냅샷을 **생략하지 않아야 한다**. `--no-backup` 명시적 지정 시에만 예외.

**REQ-MIG-001-041 (Unwanted)**
시스템은 `Apply()` 실패 시 `Rollback()` 호출을 **생략하지 않아야 한다**. Rollback 자체 실패는 critical error로 사용자에게 명시적 경고.

**REQ-MIG-001-042 (Unwanted)**
시스템은 `moai version`, `moai status`, `moai hook`, `moai glm`, `moai cg`, `moai cc`, `moai worktree` 등 **read-only 또는 session-level 명령에서는 migration을 실행하지 않아야 한다** (conservative trigger list per master-v3 §9 #10).

**REQ-MIG-001-043 (Unwanted)**
시스템은 동일 Version()을 가진 2개 이상의 step 등록을 **허용하지 않아야 한다**. 레지스트리 초기화 시점에 panic으로 실패 (compile-time 대비 runtime fail-fast).

**REQ-MIG-001-044 (Unwanted)**
시스템은 step의 `DryRun()`이 side-effect(파일 쓰기 / 외부 호출)를 발생시키는 것을 **허용하지 않아야 한다**. DryRun은 pure함수여야 함.

**REQ-MIG-001-045 (Unwanted)**
시스템은 Runner를 parallel goroutine으로 실행해서는 **안 된다**. 순차 실행만 (ordering 안전성).

### 5.6 Complex Requirements

**REQ-MIG-001-050 (Complex)**
**While** v3.0 first-run(`migration_version` 0 또는 미존재) 상태에서, **when** 사용자가 `moai update`를 실행하면, the 시스템 **shall**:
(a) Dry-run으로 모든 pending step 요약 출력,
(b) 사용자에게 "N migrations pending. Run with --yes to apply, --dry-run to preview. Or set MOAI_DISABLE_MIGRATIONS=1 to skip." 안내,
(c) `--yes` 후 backup 스냅샷 생성,
(d) 순차 Apply 실행,
(e) 성공 시 `migration_version` 갱신 + report 생성,
(f) 실패 시 Rollback + RunReport에 에러 포함 + exit code 1.

**REQ-MIG-001-051 (Complex)**
**While** Runner가 순차 Apply 실행 중, **when** step N이 실패하면, the 시스템 **shall**:
(a) step N의 `Rollback()` 호출,
(b) step 1..N-1은 그대로 유지 (선행 step은 성공으로 간주; partial apply 허용),
(c) `migration_version` = 마지막 성공 step의 Version() (N-1),
(d) RunReport에 partial-apply 명시 + 복구 명령 `moai migrate --rollback <timestamp>` 제안.

---

## 6. Acceptance Criteria (수용 기준 요약)

**AC-MIG-001-01**: Dummy step (Version=1, Description="test") 등록 후 `moai doctor migration --validate` 실행 시 inventory에 표시되고 `migration_version=0`, pending=1로 보고된다.

**AC-MIG-001-02**: `moai migrate --dry-run` 실행 시 `DryRun()` 결과만 출력되고, 실제 파일 변경은 발생하지 않음 (filesystem diff 0).

**AC-MIG-001-03**: `moai migrate --yes` 성공 후 `.moai/backups/{timestamp}/`가 생성되고, `.moai/reports/migration-{timestamp}.md` 파일이 존재. `system.yaml.migration_version=1`로 갱신.

**AC-MIG-001-04**: Rollback 테스트: apply 후 `moai migrate --rollback {timestamp}` 실행 시 파일이 원복되고 `migration_version=0` 복귀.

**AC-MIG-001-05**: `MOAI_DISABLE_MIGRATIONS=1` 설정 시 pending step 존재해도 Runner skip, 경고 1회 표시.

**AC-MIG-001-06**: `moai version` / `moai status` / `moai hook session-start` 실행 시 Runner 호출되지 않음 (goroutine trace / side-effect 없음 검증).

**AC-MIG-001-07**: Fresh `moai init` 프로젝트에서 `migration_version`이 현재 레지스트리 최대 Version() 값으로 초기화되고, 어떤 step도 실행되지 않음.

**AC-MIG-001-08**: 실패 주입 테스트: step 2에서 고의로 에러 반환 → step 1은 유지, step 2 Rollback 호출, `migration_version=1`, 이후 step 3 실행되지 않음. RunReport에 partial-apply 명시.

**AC-MIG-001-09**: 동일 Version 중복 등록 시 panic 발생 (compile/init-time fail-fast).

**AC-MIG-001-10**: `moai migrate --only M01` 실행 시 선행 step 무시 후 M01만 실행. PreConditionsMet false 시 skip.

---

## 7. Constraints (제약)

- **[HARD] SPEC-V3-SCH-001 선행**: `system.yaml.migration_version` 필드는 system_schema.go에 정의 필수.
- **[HARD] Conservative trigger list**: master-v3 §9 open question #10에 따라 init/update/doctor/migrate 4개만 PersistentPreRunE 훅 부착. 추가 확장은 SPEC 재검토 필요.
- **[HARD] Backup 의무**: `--no-backup` 명시 없이는 반드시 backup.
- **[HARD] 순차 실행**: parallel 금지. Ordering invariant 보장.
- **[HARD] Idempotency**: 모든 step은 재실행 안전. `PreConditionsMet()` 구현 필수.
- **[HARD] Cross-platform backup path**: macOS/Linux는 `~/.moai/backups/`, Windows는 `%USERPROFILE%\.moai\backups\`. `os.UserHomeDir()` 사용.
- **[HARD] MOAI_DISABLE_MIGRATIONS escape**: CI/airgap 환경을 위한 공식 opt-out. 문서화 필수.
- Binary size 영향 ≤ 200 KB (프레임워크 자체. 구체 step은 SPEC-V3-MIG-002 별도 계산).

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk ID | 설명 | 확률 | 영향 | 완화 |
|---------|------|------|------|------|
| R-MIG-001-01 | Migration 자동 실행으로 사용자 예상 밖 파일 변경 | Medium | High | Dry-run default in interactive, `--yes` 필수, systemMessage 안내; `MOAI_DISABLE_MIGRATIONS=1` escape |
| R-MIG-001-02 | Backup 디렉터리 디스크 용량 누적 | Medium | Low | `moai doctor migration --cleanup` 30일 이상 자동 정리 옵션 |
| R-MIG-001-03 | Rollback 실패 (e.g., 백업 파일 소실) | Low | Critical | Rollback 실패 시 stderr critical error + exit code 2; Rollback 자체는 단순 파일 복원이라 실패 가능성 매우 낮음 |
| R-MIG-001-04 | Cross-platform path 불일치 (Windows `\` vs Unix `/`) | Medium | Medium | `filepath.Join` 엄격 사용; path literal 금지 |
| R-MIG-001-05 | v2 `moai migrate agency` 레거시와 충돌 | Low | Medium | 기존 agency migration은 v3 프레임워크에서 Version=0으로 가짜 등록 또는 별도 유지. SPEC-V3-MIG-002에서 결정 |
| R-MIG-001-06 | `PreConditionsMet()` 구현 오류로 idempotency 깨짐 | Medium | High | 테스트 2회 연속 실행 시 동일 결과 보증 강제 (integration test); `moai doctor migration --validate`에서 step별 자가진단 |
| R-MIG-001-07 | Step 순서 의존성 위반 (v3.1에서 step 추가 시 v3.0 step 뒤에만 가능) | Low | Medium | Version() 단조 증가 원칙 문서화; SPEC review에서 체크 |
| R-MIG-001-08 | User가 `--no-backup` 습관화로 실패 복구 불가 | Low | High | CLI에 `WARNING: Proceeding without backup is NOT RECOMMENDED` 경고; docs-site 명시 |
| R-MIG-001-09 | PersistentPreRunE 추가로 `moai init`/`update` 콜드 스타트 지연 | Low | Low | pending=0 fast path: registry 스캔만 (< 5ms) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** (Formal Config Schema Framework) — `system.yaml.migration_version` 필드 스키마 정의 필요

### 9.2 Blocks

- **SPEC-V3-MIG-002** (v2-to-v3 Migration Content: M01-M05) — 본 프레임워크 위에서 구체 step 구현
- **SPEC-V3-HOOKS-001~006** — Hook Protocol v2 배포 시 기존 hook wrapper 업그레이드 migration 잠재 필요
- **SPEC-V3-CLI-001** (CLI Subcommand Restructure) — `moai migrate` 서브커맨드 플래그 (`--dry-run`, `--yes`, `--only`, `--rollback`) 정의 본 SPEC 기반
- **SPEC-V3-PLG-001** (Plugin Manifest) — 플러그인 uninstall 시 migration 로직 공유
- **SPEC-V3-SCH-002** (Settings 3-tier Source Layering) — tier 이전 시 migration framework 재사용
- **SPEC-V3-CLN-001/002/003** (내부 cleanup — other writer) — cleanup은 M03/M04/M05로 migration 형태로 배포
- 미래 모든 v3.x / v4.x breaking change

### 9.3 Related

- gap-matrix #149, #150, #151, #152, #153, #154, #155
- W1.5 §5.1-§5.4 (CC migration framework)
- W1.2 §7.1-§7.6 (migration patterns)
- W1.6 §15.4, §15.8, §15.10 (drift examples)
- CLAUDE.local.md §2 (Protected Directories), §14 (하드코딩 방지)
- master-v3 §5.1 (`moai migrate v2-to-v3` tool), §9 open question #10 (trigger aggression)
- 기존 `internal/cli/migrate_agency.go:569-595`

---

## 10. Traceability (추적성)

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| REQ-MIG-001-001, 002 | `internal/core/migration/step.go`, `registry.go` | `TestStepInterfaceCompliance`, `TestVersionUnique` |
| REQ-MIG-001-003 | `internal/config/schema/system_schema.go` (migration_version 필드) | `TestMigrationVersionField` |
| REQ-MIG-001-004 | `internal/core/migration/backup.go` | `TestBackupSnapshot` |
| REQ-MIG-001-005 | `internal/core/migration/report.go` | `TestReportMarkdownOutput` |
| REQ-MIG-001-006 | `registry.Register()` + `init()` 패턴 | `TestRegistryInitTime` |
| REQ-MIG-001-010 | `internal/cli/{init,update,doctor,migrate}.go` PersistentPreRunE | `TestPreRunEHookFires` |
| REQ-MIG-001-011 | `internal/core/migration/runner.go` interactive prompt (AskUserQuestion delegation 경로) | `TestDryRunPromptInteractive` |
| REQ-MIG-001-012, 051 | `Runner.Run()` apply → rollback 경로 | `TestApplyFailureTriggersRollback` |
| REQ-MIG-001-013 | `Runner.Run()` setVer call | `TestVersionCounterAdvances` |
| REQ-MIG-001-014 | `internal/cli/migrate.go --rollback` | `TestRollbackRestoresBackup` |
| REQ-MIG-001-015 | `internal/cli/migrate.go --only` | `TestOnlyFlagRunsSingleStep` |
| REQ-MIG-001-016 | `internal/cli/doctor_migration.go --validate` | `TestDoctorMigrationValidate` |
| REQ-MIG-001-020 | Env check in `Runner.Run()` | `TestDisableMigrationsEnvVar` |
| REQ-MIG-001-021, 044 | `--dry-run` 경로; `DryRun()` side-effect test | `TestDryRunNoSideEffects` |
| REQ-MIG-001-022 | Interactive confirm flow | `TestYesFlagBypassPrompt` |
| REQ-MIG-001-023 | `internal/cli/init.go` 첫 실행 경로 | `TestFreshInitSkipsMigrations` |
| REQ-MIG-001-024 | `Runner.Run()` skip path | `TestSkipWhenPreConditionsFalse` |
| REQ-MIG-001-030 | `--no-backup` flag | `TestNoBackupFlagWarns` |
| REQ-MIG-001-031 | `--json` flag in migrate | `TestMigrateJSONOutput` |
| REQ-MIG-001-032 | Optional interface check in runner | `TestPostApplyHookOptional` |
| REQ-MIG-001-033 | `doctor migration --cleanup` | `TestBackupCleanup30Days` |
| REQ-MIG-001-040 | Code review + test | `TestBackupRequiredDefault` |
| REQ-MIG-001-041 | Rollback call path | `TestRollbackInvokedOnFailure` |
| REQ-MIG-001-042 | `PersistentPreRunE` 등록 대상 4개 명령 한정 | `TestReadOnlyCommandsNoMigration` |
| REQ-MIG-001-043 | `registry.Register()` panic | `TestDuplicateVersionPanics` |
| REQ-MIG-001-045 | Unit test for sequential runner | `TestNoParallelExecution` |
| REQ-MIG-001-050 | End-to-end v3.0 first-run flow | `TestV3FirstRunFlow` |
| AC-MIG-001-01 ~ 10 | 해당 REQ 커버리지 전체 | `go test ./internal/core/migration/...` + CLI e2e |
| BC-004 (Breaking) | Auto-run on init/update/doctor/migrate | master-v3 Breaking Changes Catalog §4 |

---

End of SPEC-V3-MIG-001.
