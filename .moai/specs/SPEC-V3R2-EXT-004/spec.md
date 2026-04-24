---
id: SPEC-V3R2-EXT-004
title: Versioned Migration Framework (idempotent, ordered, rollback, preAction)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P0 Critical
phase: "v3.0.0 — Phase 1 — Constitution & Foundation"
module: "internal/core/migration/, internal/cli/migrate.go, internal/cli/update.go"
dependencies:
  - SPEC-V3R2-CON-001
related_gap:
  - r3-cc-architecture-preaction-hook
  - problem-catalog-ad-hoc-migration
related_theme: "Theme 7 — Extension"
breaking: true
bc_id: [BC-V3R2-019]
lifecycle: spec-anchored
tags: "migration, framework, idempotent, rollback, ordered, preaction, v3, breaking"
---

# SPEC-V3R2-EXT-004: Versioned Migration Framework

## HISTORY

| Version | Date       | Author | Description                                                        |
|---------|------------|--------|--------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — idempotent migration steps, rollback, preAction     |

---

## 1. Goal (목적)

MoAI-ADK v2.x에는 산발적이고 중앙화되지 않은 migration 로직(e.g., `migrate_agency.go`, drift repair ad-hoc)이 존재한다. v3.0.0은 **통합 migration framework**을 도입하여 (a) 버전별 migration step을 단일 레지스트리로 관리, (b) idempotent 실행 보장, (c) order-dependency 명시화, (d) rollback 경로 제공, (e) Claude Code의 **preAction hook** 패턴에서 영감을 받은 pre-condition evaluation을 표준화한다. 본 framework은 SPEC-V3R2-MIG-001 (v2→v3 migrator), SPEC-V3R2-MIG-002 (hook cleanup), SPEC-V3R2-MIG-003 (config loader completeness), 그리고 legacy SPEC-V3-CLN-001 (template drift)의 공통 runtime 기반이 된다.

### 1.1 배경

R3 §CC-architecture-reread §preAction hook: Claude Code는 hook 실행 전 pre-condition 평가를 통해 무의미한 실행을 스킵한다 (e.g., 파일 변경이 없으면 훅 실행 안 함). 동일한 패턴을 migration에 적용하면 idempotency 보장에 매우 유용하다 (REQ-MIGRATE-012 재사용 가능). 현재 `migrate_agency.go`는 ad-hoc으로 작성되어 있으며 `/moai update` 실행 시 재실행 안전성이 명시적으로 테스트되지 않는다. problem-catalog: "Migration logic is scattered across `internal/core/migration/` files without a unified framework; each migration reinvents backup/rollback/idempotency".

### 1.2 비목표 (Non-Goals)

- 기존 migration steps(migrate_agency) rewrite
- Migration DSL 도입 (Go struct + 인터페이스로 충분)
- Migration DB / SQLite persistence (`.moai/state/migration.json` 단일 파일로 충분)
- Multi-project migration (단일 프로젝트 scope)
- Cross-version migration chaining 자동화 (v2 → v3 path는 SPEC-V3R2-MIG-001 담당)
- Migration UI
- Live migration (실행 중 app 업그레이드)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `internal/core/migration/framework.go` (신규), `internal/core/migration/runner.go` (신규), `internal/core/migration/step.go` (인터페이스 정의), `.moai/state/migration.json` 상태 파일 스키마.
- `MigrationStep` interface 정의: `Version() string`, `ID() string`, `Description() string`, `IsIdempotent() bool`, `PreConditionsMet(ctx) (bool, error)`, `DryRun(ctx) (Diff, error)`, `Apply(ctx) error`, `Rollback(ctx) error`.
- `MigrationRunner` 구현: step 등록, 순서 결정(version + dep graph), `.moai/backups/<timestamp>/` 스냅샷, dry-run, idempotent 재실행 안전성.
- `moai migrate` CLI 서브커맨드: `--list`, `--apply`, `--dry-run`, `--only <ID>`, `--rollback <timestamp>`.
- `.moai/state/migration.json`: applied migration ID 목록, 각 step의 마지막 실행 시간, rollback 이력.
- preAction hook 패턴: `PreConditionsMet`이 `false` 반환 시 Apply 실행 skip.
- 단일 step 실패 시 본 step만 Rollback; 전체 chain은 non-atomic (실행 완료된 이전 step은 유지).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 기존 migrate_agency 재작성
- Cross-project migration
- Migration DSL
- Migration UI
- SQL-style migration (up/down 스크립트 언어)
- Rollback of rollback (second-level undo)
- Network-based migration downloads
- Concurrent migration execution (단일 실행 보장)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/core/migration/`, `.moai/state/`, `.moai/backups/`
- 영향 디렉터리:
  - 신설: `internal/core/migration/framework.go`, `runner.go`, `step.go`
  - 수정: `internal/cli/migrate.go` (신규 or 확장), `internal/cli/update.go` (runner 연동)
  - 참조: `internal/core/migration/steps/` (SPEC-V3R2-MIG-001/002/003의 step 파일들)
- 외부 레퍼런스: R3 §preAction hook, SPEC-V3-CLN-001 MigrationStep 원형

---

## 4. Assumptions (가정)

- 사용자 프로젝트에는 `.moai/state/`와 `.moai/backups/` 디렉터리 생성 권한이 있다.
- 디스크 여유 공간은 원본의 최소 2배 (backup + temp).
- `gopkg.in/yaml.v3`과 표준 라이브러리 외 추가 의존성은 필요 없다.
- 모든 migration step은 Go struct로 구현되며 `internal/core/migration/steps/` 아래에 등록된다.
- `time.Time` ISO-8601 포맷이 timestamp의 표준이다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-EXT004-001**
The framework **shall** define interface `MigrationStep` with 8 methods: `Version()`, `ID()`, `Description()`, `IsIdempotent()`, `PreConditionsMet()`, `DryRun()`, `Apply()`, `Rollback()`.

**REQ-EXT004-002**
The framework **shall** define `MigrationRunner` with methods: `Register(step)`, `ListPending(ctx)`, `DryRun(ctx, ids...)`, `Apply(ctx, ids...)`, `Rollback(ctx, timestamp)`.

**REQ-EXT004-003**
Every migration step **shall** be idempotent: calling `Apply` twice on the same initial state produces the same final state.

**REQ-EXT004-004**
The runner **shall** persist applied migration IDs in `.moai/state/migration.json` with timestamp, status, and affected files.

**REQ-EXT004-005**
The runner **shall** snapshot affected files to `.moai/backups/<ISO-8601-timestamp>/` before calling `Apply`.

**REQ-EXT004-006**
Step ordering **shall** be determined by: primary by `Version()` (semver), secondary by `ID()` lexicographic ascending.

**REQ-EXT004-007**
`moai migrate` CLI **shall** support flags: `--list`, `--apply`, `--dry-run`, `--only <ID>`, `--rollback <timestamp>`.

### 5.2 Event-Driven Requirements

**REQ-EXT004-008**
**When** `moai update` completes template refresh, it **shall** invoke `MigrationRunner.Apply(ctx, pending...)` with `--dry-run` preview as default; actual apply requires user confirmation or `--yes` flag.

**REQ-EXT004-009**
**When** `PreConditionsMet()` returns `false`, the runner **shall** skip `Apply` and record status `skipped (preconditions)` in state file.

**REQ-EXT004-010**
**When** `Apply` fails mid-execution, the runner **shall** call `Rollback()` for that step and abort the chain; previously applied steps remain applied.

**REQ-EXT004-011**
**When** the user runs `moai migrate --rollback <timestamp>`, the runner **shall** restore files from `.moai/backups/<timestamp>/` and update state file.

### 5.3 State-Driven Requirements

**REQ-EXT004-012**
**While** a migration step's `IsIdempotent()` returns `false`, the runner **shall** refuse to re-apply it and emit `STEP_NOT_IDEMPOTENT_SKIP`.

**REQ-EXT004-013**
**While** `.moai/backups/` has insufficient disk space (< 2x original size), `Apply` **shall** abort before modification with `MIGRATION_BACKUP_DISK_LOW`.

**REQ-EXT004-014**
**While** the runner is executing, concurrent `moai migrate` invocations **shall** be prevented via a file-lock at `.moai/state/migration.lock`.

### 5.4 Optional Requirements

**REQ-EXT004-015**
**Where** a user sets `.moai/config/sections/system.yaml: migration.skip_steps: [<ID>]`, the runner **shall** permanently skip those steps.

**REQ-EXT004-016**
**Where** a step declares a `DependsOn` list, the runner **shall** ensure all dependencies are applied before the dependent step.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-EXT004-017 (Unwanted Behavior)**
**If** a step's `Apply` modifies files outside the `PreConditionsMet` declared impact set, **then** CI integration tests **shall** fail with `STEP_SCOPE_VIOLATION`.

**REQ-EXT004-018 (Unwanted Behavior)**
**If** the backup snapshot write fails, **then** `Apply` **shall not** proceed; the failure **shall** be surfaced as `MIGRATION_BACKUP_FAILED`.

**REQ-EXT004-019 (Complex: State + Event)**
**While** `moai update` is in progress, **when** `Ctrl+C` is received mid-Apply, the runner **shall** attempt `Rollback()` for the current step and exit with non-zero code, preserving state consistency.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-EXT004-01**: Given a MigrationStep stub When `Version()`, `ID()`, `Description()` are called Then all return non-empty strings (maps REQ-EXT004-001).
- **AC-EXT004-02**: Given 2 pending migrations with versions v3.0.1 and v3.0.2 When `ListPending` is called Then order is [v3.0.1, v3.0.2] (maps REQ-EXT004-006).
- **AC-EXT004-03**: Given an idempotent step applied once When applied again Then final state is identical to after first apply (maps REQ-EXT004-003).
- **AC-EXT004-04**: Given `Apply` succeeds When state file is inspected Then step ID, timestamp, affected files are recorded (maps REQ-EXT004-004).
- **AC-EXT004-05**: Given `Apply` is called When execution starts Then `.moai/backups/<timestamp>/` contains pre-change snapshots (maps REQ-EXT004-005).
- **AC-EXT004-06**: Given `PreConditionsMet` returns false When runner iterates Then step is skipped with status `skipped (preconditions)` (maps REQ-EXT004-009).
- **AC-EXT004-07**: Given step 2 of 3 fails When runner executes Then step 2 rollbacks; steps 1 remains applied; step 3 not attempted (maps REQ-EXT004-010).
- **AC-EXT004-08**: Given `moai migrate --rollback <ts>` When executed Then files are restored from `.moai/backups/<ts>/` (maps REQ-EXT004-011).
- **AC-EXT004-09**: Given disk space < 2x source When `Apply` is called Then abort with `MIGRATION_BACKUP_DISK_LOW` (maps REQ-EXT004-013).
- **AC-EXT004-10**: Given 2 concurrent `moai migrate` invocations When the second starts Then it errors on file-lock collision (maps REQ-EXT004-014).
- **AC-EXT004-11**: Given `migration.skip_steps: [M01]` in system.yaml When runner executes Then M01 is skipped (maps REQ-EXT004-015).
- **AC-EXT004-12**: Given step B declares `DependsOn: [A]` but A not applied When runner tries B Then B waits for A or aborts (maps REQ-EXT004-016).
- **AC-EXT004-13**: Given Ctrl+C during Apply When interrupted Then current step rollbacks and exit code is non-zero (maps REQ-EXT004-019).
- **AC-EXT004-14**: Given a non-idempotent step declared When re-applied Then `STEP_NOT_IDEMPOTENT_SKIP` emitted (maps REQ-EXT004-012).
- **AC-EXT004-15**: Given backup snapshot write failure When Apply called Then `MIGRATION_BACKUP_FAILED` with no file modifications (maps REQ-EXT004-018).
- **AC-EXT004-16**: Given a step violates declared impact set When CI integration test runs Then `STEP_SCOPE_VIOLATION` surfaces (maps REQ-EXT004-017).

---

## 7. Constraints (제약)

- 9-direct-dep 정책 준수 (Go 표준 라이브러리 + `yaml.v3`만 사용).
- 단일 프로젝트 scope (multi-project migration 금지).
- Windows 호환: POSIX permission은 no-op (SPEC-AGENCY-ABSORB-001 REQ-MIGRATE-012a/b 패턴 재사용).
- `.moai/backups/`, `.moai/state/`는 protected directories이나 본 framework의 소유 영역.
- 파일-lock은 OS 수준 file-locking이 아닌 `.moai/state/migration.lock` 파일 존재 여부로 판정 (stale lock detection 필요).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Idempotency 실수로 데이터 손실 | 심각 | `IsIdempotent()` 자기 선언 + CI integration test 의무화 |
| 백업 크기가 디스크 포화 | Apply 실패 | REQ-EXT004-013의 pre-check + `--skip-backup` 금지 |
| Stale lock으로 사용자 차단 | UX 저하 | lock 파일에 PID + timestamp 저장, stale(10min 이상) 자동 무효화 |
| Rollback 실패로 반파 상태 | 데이터 무결성 | `Rollback()`이 실패하면 명시적 에러 + `.moai/reports/rollback-<ts>-error.md` 생성 |
| `DependsOn` 순환 의존 | deadlock | topological sort + cycle detection at registration time |
| Migration chain이 non-atomic | 부분 성공 혼란 | state 파일에 partial apply 기록 + 보고서 생성 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- 없음 (신규 framework).

### 9.2 Blocks

- SPEC-V3R2-MIG-001 (v2→v3 migrator): 본 framework 위에서 step 구현.
- SPEC-V3R2-MIG-002 (hook cleanup): hook registration cleanup step 구현.
- SPEC-V3R2-MIG-003 (config loader completeness): config 관련 migration step.
- SPEC-V3-CLN-001 (legacy): 기존 M01/M03/M04가 신규 framework로 이관.

### 9.3 Related

- R3 §preAction hook pattern.

---

## 10. Traceability (추적성)

- REQ 총 19개: Ubiquitous 7, Event-Driven 4, State-Driven 3, Optional 2, Complex 3.
- AC 총 16개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R3 §preAction hook pattern; SPEC-V3-CLN-001 MigrationStep 원형; problem-catalog ad-hoc migration.
- BC 영향: **BC-V3R2-019** — `moai update`가 migration runner를 자동 invoke하므로 기존 v2 사용자 워크플로우에 영향 (dry-run preview가 기본이므로 사용자 확인 거치지 않고 적용되지 않음).
- 구현 경로 예상:
  - `internal/core/migration/framework.go`
  - `internal/core/migration/runner.go`
  - `internal/core/migration/step.go`
  - `internal/cli/migrate.go`
  - `internal/cli/update.go` (runner 연동)
  - `.moai/state/migration.json` schema docs
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1078` (§11.7 EXT-004 definition)
  - `docs/design/major-v3-master.md:L979` (§8 BC-V3R2-019 — migration framework auto-run)
  - `docs/design/major-v3-master.md:L988` (§9 Phase 1 Constitution & Foundation)

---

End of SPEC.
