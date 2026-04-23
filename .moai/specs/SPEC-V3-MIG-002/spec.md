---
id: SPEC-V3-MIG-002
title: v2-to-v3 Migration Content (M01-M05 concrete steps)
version: 0.1.0
status: draft
created: 2026-04-22
updated: 2026-04-22
author: manager-spec
priority: High
phase: "Phase 1 — Foundation"
module: "internal/core/migration/steps/"
dependencies:
  - SPEC-V3-MIG-001
  - SPEC-V3-SCH-001
related_gap: [183, 184, 185, 186, 187, 188, 190, 191]
related_theme: "Theme 2 — Migration Framework + Theme 9 — Internal Cleanup"
breaking: false
bc_id: []
lifecycle: spec-first
tags: "v3, migration, content, cleanup, drift, v2-to-v3, P1"
---

# SPEC-V3-MIG-002: v2-to-v3 Migration Content (M01-M05)

## HISTORY

- 2026-04-22 v0.1.0: 최초 작성. master-v3 §3.2 design approach step 6, §3.9 cleanup, gap-matrix #183-#191, W1.6 §15 self-identified issues 근거. SPEC-V3-MIG-001 프레임워크 위에서 구체 step 5개 구현.

---

## 1. Goal (목적)

SPEC-V3-MIG-001이 제공한 versioned migration framework에 v2.x → v3.0 전환을 위한 **구체 step 5개** (M01-M05)를 등록한다. 각 step은 master-v3 §3.2에서 enumerate된 debt 항목(§2.3, §3.9)과 1:1로 매핑된다.

이는 단순 프레임워크 테스트가 아니라 W1.6 §15.1-§15.10에서 자체 식별된 moai-adk-go의 실제 drift/cruft를 migration으로 제거하는 "dogfooding" SPEC이다. `moai migrate v2-to-v3` 단일 엔트리 포인트(SPEC-V3-MIGRATE-001 소관 — 다른 writer)가 이 5 step을 순차 실행한다.

### 1.1 배경

W1.6 §15 자체 식별 issue 중 migration 대상 enumerate:

| ID | Issue | gap row | 해결 migration |
|----|-------|---------|----------------|
| §15.4/§15.8 | `project.yaml:template_version: v2.7.22` stale (system.yaml: v2.12.0) | #183 (Critical) | **M01** |
| §15.6 | `.agency/` 잔재 (8 redirect + stub constitution + `.moai-backups/`) | #190, #191 | **M02** |
| §15.1/§15.7/§5.5 | `handle-permission-denied.sh` 로컬 누락 | #185 (Medium) | **M03** |
| §15.1/§7.1 | 3개 템플릿 전용 skill 로컬 누락 (`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`) | #184 (Medium) | **M04** |
| §15.10/§13.2 | `.go.bak` (42KB 합) + `coverage.out/html` stale | #186, #187, #188 | **M05** |

다음 항목은 **본 SPEC 범위 외** (직접 소스 수정 또는 별도 SPEC):
- `internal/template/embed.go:8-12` ADR-011 주석 drift (#189) → 직접 소스 edit (SPEC-V3-CLN-002 소관 — other writer)
- `.mcp.json` pencil 누락 (#192) → 직접 템플릿 소스 edit
- Handler count doc (#193) → 주석 추가
- docs-site locale lag (#194) → manager-docs 위임

### 1.2 비목표 (Non-Goals)

- **프레임워크 자체 구현 금지**: SPEC-V3-MIG-001 완성 필수. 본 SPEC은 5 step만 제공.
- **`moai migrate v2-to-v3` CLI tool 구현 금지**: SPEC-V3-MIGRATE-001 관할. 본 SPEC은 step 레지스트리만.
- **신규 feature migration 금지**: M01-M05는 debt resolution만. v3 신규 기능의 config migration은 해당 SPEC(HOOKS-001, SCH-002 등)에서 별도 step 등록 시 Version()을 100+로 할당.
- **Backup 중복 방지**: SPEC-V3-MIG-001의 backup 메커니즘 의존. 각 step은 자체 backup 로직 구현 안 함.
- **Agency 시스템 재통합 금지**: SPEC-AGENCY-ABSORB-001 완료된 상태 전제. M02는 archive만.

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/core/migration/steps/` 디렉터리에 5 파일:
  - `m01_template_version.go`
  - `m02_agency_archive.go`
  - `m03_hook_wrapper_drift.go`
  - `m04_skill_drift.go`
  - `m05_legacy_cleanup.go`
- 각 파일은 `MigrationStep` 인터페이스 구현 + `init()`에서 `registry.Register()` 호출

- **M01 — Template Version Sync** (Version=1, Idempotent):
  - 읽기: `.moai/config/sections/project.yaml` → `template_version`
  - 읽기: `.moai/config/sections/system.yaml` → `moai.version`
  - PreConditionsMet: project.template_version != system.moai.version
  - Apply: project.template_version ← system.moai.version
  - Rollback: 백업 YAML 복원

- **M02 — .agency/ Archive** (Version=2, One-shot):
  - 대상: `.claude/commands/agency/*.md` (8 redirect 파일), `.claude/rules/agency/constitution.md` (695B stub), `.moai-backups/` (폴더)
  - PreConditionsMet: 위 경로 중 1개 이상 존재
  - Apply: `~/.moai/history/v2.12/{commands,rules,backups}/`로 이동 (delete 아님)
  - Rollback: archive에서 원 위치로 복원
  - [HARD] Never delete; move only

- **M03 — Hook Wrapper Drift** (Version=3, Idempotent):
  - 대상: `.claude/hooks/moai/handle-permission-denied.sh` (로컬 누락)
  - PreConditionsMet: 해당 파일 부재 AND 템플릿 `handle-permission-denied.sh.tmpl` 존재
  - Apply: 템플릿 렌더 후 로컬 배포 (기존 deployer 재사용)
  - Rollback: 배포한 파일 제거

- **M04 — Skill Drift** (Version=4, Idempotent):
  - 대상: `.claude/skills/` 하위 3개 skill (`moai-domain-db-docs/`, `moai-workflow-design-context/`, `moai-workflow-pencil-integration/`)
  - PreConditionsMet: 해당 skill 디렉터리 중 1개 이상 부재 AND 템플릿에 존재
  - Apply: 각 skill 디렉터리를 템플릿에서 재귀 배포
  - Rollback: 배포한 skill 디렉터리 제거

- **M05 — Legacy Code Cleanup** (Version=5, One-shot):
  - 대상:
    - `internal/cli/glm.go.bak` (28,567 B)
    - `internal/cli/worktree/new_test.go.bak` (13,700 B)
    - `coverage.out` (mtime > 30일 전)
    - `coverage.html` (mtime > 30일 전)
  - PreConditionsMet: 위 파일 중 1개 이상 존재
  - Apply: 백업 후 제거
  - Rollback: backup에서 복원
  - Note: `.go.bak` 파일은 git tracked이므로 `git rm` 대신 파일 삭제 + `.gitignore` 추가는 별도 PR

- Step 메타데이터 CSV: `internal/core/migration/steps/registry.go`에 5 step 자체 등록

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- ADR-011 주석 수정 (gm#189) — 직접 소스 edit (other writer: CLN-002)
- `.mcp.json.tmpl` pencil 추가 (gm#192) — 직접 템플릿 edit
- Handler count 주석 추가 (gm#193) — 직접 소스 edit
- docs-site locale sync (gm#194) — manager-docs 위임, not migration
- `git rm` 자동화 — moai는 git 상태 직접 수정하지 않음
- 새로운 v3 feature config migration (hooks.yaml 스키마 확장 등) — 해당 기능 SPEC에서 별도 step
- Agency 시스템 재활성화 또는 `.agency/` 복원 로직 — 복구 경로는 manual rollback only
- Cross-version fast-forward (v2.10 → v3.0 단일 migration) — 순차 누적 실행 전제
- Incremental M01 optimization (big YAML) — template_version은 단일 scalar field

---

## 3. Environment (환경)

- 런타임: Go 1.26, moai-adk-go v3.0.0-alpha.1+
- 의존성:
  - SPEC-V3-MIG-001 (Runner, MigrationStep interface, backup 메커니즘)
  - SPEC-V3-SCH-001 (project.yaml, system.yaml 스키마)
- 기존 활용 컴포넌트:
  - `internal/template/deployer.go` (M03, M04에서 재사용)
  - `internal/template/embed.go` (embed FS에서 파일 조회)
  - `internal/template/manifest.go` (provenance 추적)
  - `internal/config/manager.go` (YAML read/write)
- 영향 디렉터리 (실행 시점):
  - `.moai/config/sections/project.yaml` (M01 수정)
  - `.claude/commands/agency/`, `.claude/rules/agency/`, `.moai-backups/` (M02 이동 대상)
  - `~/.moai/history/v2.12/` (M02 archive 대상)
  - `.claude/hooks/moai/handle-permission-denied.sh` (M03 생성)
  - `.claude/skills/moai-domain-db-docs/`, `.claude/skills/moai-workflow-design-context/`, `.claude/skills/moai-workflow-pencil-integration/` (M04 배포)
  - `internal/cli/glm.go.bak`, `internal/cli/worktree/new_test.go.bak`, `coverage.out`, `coverage.html` (M05 삭제 대상)
- 대상 OS: macOS / Linux / Windows 동등

---

## 4. Assumptions (가정)

- A-001 (High): `internal/template/embed.go`의 embedded FS에 `handle-permission-denied.sh.tmpl` 및 3개 skill (`moai-domain-db-docs/`, `moai-workflow-design-context/`, `moai-workflow-pencil-integration/`)이 이미 존재한다. (W1.6 §7.1 / §5.5 확인)
- A-002 (High): `~/.moai/history/v2.12/` 디렉터리는 M02 실행 시 자동 생성되며, 기존 내용이 있으면 append (never overwrite).
- A-003 (Medium): M05 대상 `.go.bak` 파일은 git tracked 이지만, 파일 삭제만으로 다음 commit 시 git status가 변경 감지. 사용자에게 `git add -A && git commit -m "chore: remove pre-v3 .bak files"` 안내.
- A-004 (High): `project.yaml` YAML 주석은 M01 적용 후 보존되어야 한다. `yaml.v3`의 `Node` API로 주석 유지 처리.
- A-005 (Medium): 사용자가 M02 archive 위치(`~/.moai/history/v2.12/`)를 직접 참조하지 않는다. 문서화된 복구 경로 외 의존 금지.
- A-006 (Medium): Coverage 파일의 mtime이 migration 실행 시점의 30일 전 기준으로 평가된다 (SPEC-V3-MIG-002가 v3.0 릴리스 시점에 실행되는 경우 대부분 대상 포함).
- A-007 (High): 5개 step 사이 의존성이 없다 (순차 실행 가능하나 ordering 강제는 Version() 숫자로만). M01만 YAML 수정, 나머지는 파일/디렉터리 조작 — 서로 독립.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MIG-002-001 (Ubiquitous)**
시스템은 **항상** 5개 step 모두 `MigrationStep` 인터페이스 (SPEC-V3-MIG-001) 전체를 구현한다.

**REQ-MIG-002-002 (Ubiquitous)**
시스템은 **항상** 각 step의 Version() 값을 다음과 같이 고정한다: M01=1, M02=2, M03=3, M04=4, M05=5.

**REQ-MIG-002-003 (Ubiquitous)**
시스템은 **항상** 각 step ID를 다음 문자열로 설정한다: M01="M01-template-version-sync", M02="M02-agency-archive", M03="M03-hook-wrapper-drift", M04="M04-skill-drift", M05="M05-legacy-cleanup".

**REQ-MIG-002-004 (Ubiquitous)**
시스템은 **항상** 각 step의 `IsIdempotent()`를 다음과 같이 반환한다: M01=true, M02=false (one-shot archive), M03=true, M04=true, M05=false (one-shot delete).

**REQ-MIG-002-005 (Ubiquitous)**
시스템은 **항상** 각 step의 Description()을 한글 + 영문 병기로 제공해 사용자가 dry-run 출력에서 목적을 즉시 이해하도록 한다.

### 5.2 Event-Driven Requirements

**REQ-MIG-002-010 (Event-Driven)**
**When** M01 `Apply()`가 호출되면, the 시스템 **shall** `.moai/config/sections/project.yaml`의 `template_version` 필드를 `.moai/config/sections/system.yaml`의 `moai.version` 값으로 동기화하고, YAML 주석을 보존한다.

**REQ-MIG-002-011 (Event-Driven)**
**When** M02 `Apply()`가 호출되면, the 시스템 **shall** `.claude/commands/agency/*.md` (8 파일), `.claude/rules/agency/constitution.md`, `.moai-backups/`를 `~/.moai/history/v2.12/`로 **이동**(move)하고, 원 위치에는 파일/디렉터리를 남기지 않는다.

**REQ-MIG-002-012 (Event-Driven)**
**When** M03 `Apply()`가 호출되면, the 시스템 **shall** embedded template의 `handle-permission-denied.sh.tmpl`을 렌더해 `.claude/hooks/moai/handle-permission-denied.sh`에 배포하고, 실행 권한(`0755`)을 설정한다.

**REQ-MIG-002-013 (Event-Driven)**
**When** M04 `Apply()`가 호출되면, the 시스템 **shall** embedded template의 3개 skill 디렉터리 (`moai-domain-db-docs/`, `moai-workflow-design-context/`, `moai-workflow-pencil-integration/`)를 로컬 `.claude/skills/` 아래에 재귀 배포한다.

**REQ-MIG-002-014 (Event-Driven)**
**When** M05 `Apply()`가 호출되면, the 시스템 **shall** `internal/cli/glm.go.bak`, `internal/cli/worktree/new_test.go.bak`, 그리고 mtime이 30일 이상 경과한 `coverage.out` / `coverage.html` 파일을 backup 스냅샷 생성 후 제거한다.

**REQ-MIG-002-015 (Event-Driven)**
**When** 각 step의 `Rollback()`이 호출되면, the 시스템 **shall** SPEC-V3-MIG-001의 backup 스냅샷 메커니즘을 활용해 파일을 원 상태로 복원한다.

### 5.3 State-Driven Requirements

**REQ-MIG-002-020 (State-Driven)**
**While** M01의 `PreConditionsMet()` 평가 중, the 시스템 **shall** project.template_version == system.moai.version인 경우 false를 반환해 step이 skip 되도록 한다.

**REQ-MIG-002-021 (State-Driven)**
**While** M02의 `PreConditionsMet()` 평가 중, the 시스템 **shall** 대상 3개 경로 중 하나라도 존재하면 true, 모두 부재하면 false를 반환한다.

**REQ-MIG-002-022 (State-Driven)**
**While** M03의 `PreConditionsMet()` 평가 중, the 시스템 **shall** 로컬 `handle-permission-denied.sh`가 이미 존재하면 false를 반환한다 (재배포 방지).

**REQ-MIG-002-023 (State-Driven)**
**While** M04의 `PreConditionsMet()` 평가 중, the 시스템 **shall** 3개 skill 모두 이미 로컬에 존재하면 false를 반환하고, 1개 이상 부재 시 true를 반환한다.

**REQ-MIG-002-024 (State-Driven)**
**While** M05의 `PreConditionsMet()` 평가 중, the 시스템 **shall** 대상 4개 파일 중 1개 이상 실제 존재하면 true를 반환하고, 모두 부재하면 false (skip) 반환한다.

### 5.4 Optional Requirements

**REQ-MIG-002-030 (Optional)**
**Where** M05 대상 `.bak` 파일이 git tracked 상태인 경우, the 시스템 **shall** post-migration report에 `git add -A && git commit -m "chore: remove pre-v3 .bak files"` 명령 안내를 포함한다.

**REQ-MIG-002-031 (Optional)**
**Where** M02 archive 후 `~/.moai/history/v2.12/`에 이미 동일 이름 파일이 존재하면, the 시스템 **shall** `.{N}.bak` suffix를 붙여 충돌 회피한다 (never overwrite).

**REQ-MIG-002-032 (Optional)**
**Where** `MOAI_MIGRATION_VERBOSE=1` 환경변수가 설정되면, the 시스템 **shall** 각 step의 DryRun/Apply 과정에서 파일 단위 diff를 stdout에 출력한다.

### 5.5 Unwanted Behavior (Must Not)

**REQ-MIG-002-040 (Unwanted)**
시스템은 M02 실행 시 `.agency/` 경로에서 파일을 **삭제하지 않아야 한다**. 이동(move)만 허용, 원 위치는 empty 디렉터리 또는 완전 제거.

**REQ-MIG-002-041 (Unwanted)**
시스템은 M04 실행 시 사용자가 수정한 `UserModified` 상태의 skill 파일을 **덮어쓰지 않아야 한다**. 기존 파일이 `UserModified`이면 해당 skill 배포 skip 후 경고 표시.

**REQ-MIG-002-042 (Unwanted)**
시스템은 M01 실행 시 project.yaml의 사용자 주석 및 anchor alias를 **손상시키지 않아야 한다**.

**REQ-MIG-002-043 (Unwanted)**
시스템은 M05 실행 시 mtime이 30일 미만인 coverage 파일을 **삭제하지 않아야 한다** (CI로 방금 생성된 코드 커버리지 보호).

**REQ-MIG-002-044 (Unwanted)**
시스템은 M03 배포 후 기존 권한 `0755`가 아닌 값으로 **설정하지 않아야 한다** (Claude Code의 훅 실행 요구사항).

### 5.6 Complex Requirements

**REQ-MIG-002-050 (Complex)**
**While** M04 배포 수행 중, **when** 3개 skill 중 하나라도 `internal/template/manifest` 기준으로 `UserModified` 상태라면, the 시스템 **shall** 해당 skill만 skip하고 나머지 skill은 배포 진행하며, post-migration report에 skip된 skill 목록과 이유를 명시한다.

**REQ-MIG-002-051 (Complex)**
**While** v3.0 first-run 사용자가 v2.x에서 migrate 중이며 `.agency/` + `.moai-backups/` + `internal/cli/*.bak` + `coverage.out` 모두 존재하는 상태에서, **when** `moai migrate v2-to-v3 --yes` 실행 시, the 시스템 **shall** M01→M02→M03→M04→M05 순차 실행하고, 각 step별 backup 서브디렉터리를 `.moai/backups/{timestamp}/M0N/` 경로로 분리 저장해 rollback granularity를 보장한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

**AC-MIG-002-01** (M01): project.yaml에 `template_version: v2.7.22`, system.yaml에 `moai.version: v2.12.0` 인 상태에서 M01 Apply 후 project.yaml의 `template_version: v2.12.0`으로 변경되고 주석 보존.

**AC-MIG-002-02** (M02): `.claude/commands/agency/migrate.md` 존재 상태에서 M02 Apply 후 해당 파일이 `~/.moai/history/v2.12/commands/agency/migrate.md`로 이동, 원 위치 파일 부재.

**AC-MIG-002-03** (M02 idempotency): M02는 one-shot이지만 재실행 시 PreConditionsMet=false 반환해 skip. 정상 종료 코드 0.

**AC-MIG-002-04** (M03): `.claude/hooks/moai/handle-permission-denied.sh` 부재 + template 존재 상태에서 M03 Apply 후 파일 생성, `stat -c "%a"` 결과 `755`.

**AC-MIG-002-05** (M04): 3 skill 중 2개만 부재인 상태에서 M04 Apply 후 부재했던 2개만 배포, 존재하던 1개는 변경 없음 (manifest `UserCreated` 보호).

**AC-MIG-002-06** (M05): `.bak` 2개 파일 + 30일+ coverage 2개 존재 상태에서 M05 Apply 후 4개 모두 제거, backup 디렉터리에서 복구 가능 확인.

**AC-MIG-002-07** (M05 recent coverage): 5일 전 생성된 `coverage.out` 존재 상태에서 M05 Apply 시 해당 파일 보존, 오직 30일+ 파일만 제거.

**AC-MIG-002-08** (Rollback): M01 Apply 후 `moai migrate --rollback {timestamp}` 실행 시 project.yaml이 원래 `v2.7.22`로 복원되고 `migration_version=0` 복귀.

**AC-MIG-002-09** (Non-destructive archive): M02 Apply 이후 `~/.moai/history/v2.12/`에 모든 archived 내용이 존재하고, 원본 소실 없음 (전체 바이트 합 일치).

**AC-MIG-002-10** (End-to-end): v2.12 corpus 10개 프로젝트에서 `moai migrate v2-to-v3 --yes` 실행 후 `.moai/reports/migration-{timestamp}.md`에 M01-M05 5개 모두 applied 또는 skipped (idempotency) 명시, exit code 0.

**AC-MIG-002-11** (UserModified protection): 사용자가 `moai-domain-db-docs/SKILL.md`를 수정한 상태(`UserModified`)에서 M04 실행 시 해당 skill skip, 다른 2개 skill만 배포, report에 skip 이유 기록.

**AC-MIG-002-12** (Version assignment): `moai doctor migration --validate` 출력에 M01=Version 1, M02=Version 2, M03=Version 3, M04=Version 4, M05=Version 5로 정확히 표시.

---

## 7. Constraints (제약)

- **[HARD] SPEC-V3-MIG-001 선행**: Runner, MigrationStep 인터페이스, backup 메커니즘 의존.
- **[HARD] SPEC-V3-SCH-001 선행**: project.yaml, system.yaml 스키마 의존 (특히 `migration_version` 필드).
- **[HARD] Version() 고정**: M01=1 ~ M05=5. 후속 migration은 Version() 6+ 할당.
- **[HARD] Non-destructive M02**: archive = move, never delete.
- **[HARD] UserModified protection**: manifest provenance `UserModified` 파일은 모든 step에서 건드리지 않음.
- **[HARD] Idempotency per step**: IsIdempotent()=true인 M01, M03, M04는 재실행 안전.
- **[HARD] One-shot M02, M05**: IsIdempotent()=false이지만 PreConditionsMet()이 두 번째 실행에서 false 반환하므로 효과적으로 safe.
- Cross-platform path: `os.UserHomeDir()` + `filepath.Join` 엄수.
- Binary size 증가 ≤ 50 KB (5개 step 코드 합).

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk ID | 설명 | 확률 | 영향 | 완화 |
|---------|------|------|------|------|
| R-MIG-002-01 | M01이 project.yaml 주석 소실 | Medium | Medium | yaml.v3 `Node` API 사용; test case로 주석 보존 검증 |
| R-MIG-002-02 | M02 archive 대상 이미 `~/.moai/history/`에 존재 시 충돌 | Low | Medium | `.{N}.bak` suffix 추가 (REQ-MIG-002-031) |
| R-MIG-002-03 | M04 템플릿 skill 누락 (embed FS 이슈) | Low | High | Build-time assert: `go:embed` 후 `internal/template/` 초기화에서 3개 skill 존재 확인 |
| R-MIG-002-04 | M05가 CI-generated 최신 coverage 삭제 | Low | High | 30일 mtime 경계 (REQ-MIG-002-043) + `.gitignore`에 coverage 포함 권장 |
| R-MIG-002-05 | M03 파일 권한 오류 (Windows vs Unix) | Medium | Low | Windows는 permission 개념 차이 — `os.Chmod(0755)` best-effort, 실패 시 warning만 |
| R-MIG-002-06 | `.bak` git tracked 상태에서 파일 삭제 후 git status dirty | Medium | Low | Post-migration report에 `git commit` 안내 (REQ-MIG-002-030) |
| R-MIG-002-07 | 사용자가 `.agency/` 관련 자동화를 직접 작성한 상태 | Low | Medium | M02 archive 경로 docs-site에 명시; `moai doctor` 에서 archive 참조 경고 |
| R-MIG-002-08 | M01 후 template_version과 moai.version 재분기 | Low | Low | M01 idempotent 재실행으로 회복 가능; `moai doctor config --fix`에서도 감지 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-MIG-001** (Versioned Migration Framework) — Runner, MigrationStep, backup 의존
- **SPEC-V3-SCH-001** (Formal Config Schema Framework) — project.yaml, system.yaml 스키마 의존

### 9.2 Blocks

- **SPEC-V3-MIGRATE-001** (`moai migrate v2-to-v3` CLI tool) — 본 SPEC의 5 step을 순차 실행하는 wrapper
- **SPEC-V3-CLN-001/002/003** (Internal Cleanup SPECs — other writer) — 중복 항목 (M03/M04/M05와 overlap); cleanup SPEC은 본 migration을 reference

### 9.3 Related

- gap-matrix #183 (M01 Critical), #184 (M04 Medium), #185 (M03 Medium), #186/#187 (M05 Low), #188 (M05 Low), #190/#191 (M02 Low)
- W1.6 §15.1, §15.4, §15.6, §15.7, §15.8, §15.10 (self-identified issues)
- master-v3 §3.2 M01-M05 list, §5.1 `moai migrate v2-to-v3` 출력 예시
- CLAUDE.local.md §2 (Protected Directories) — manifest 기반 `UserCreated`/`UserModified` 보호 규칙

---

## 10. Traceability (추적성)

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| REQ-MIG-002-001 | `internal/core/migration/steps/m0{1..5}_*.go` | `TestAllStepsSatisfyInterface` |
| REQ-MIG-002-002, 012 | Each step's `Version()` / `ID()` const | `TestStepVersionsUnique`, `TestStepIDs` |
| REQ-MIG-002-003 | Each step's `ID()` method | compile-time const |
| REQ-MIG-002-004 | Each step's `IsIdempotent()` | `TestIdempotencyFlags` |
| REQ-MIG-002-005 | Each step's `Description()` | `TestDescriptionBilingual` |
| REQ-MIG-002-010 | `m01_template_version.go Apply()` | `TestM01SyncPreservesComments` |
| REQ-MIG-002-011, 040 | `m02_agency_archive.go Apply()` | `TestM02MoveNotDelete` |
| REQ-MIG-002-012, 044 | `m03_hook_wrapper_drift.go Apply()` | `TestM03Permission0755` |
| REQ-MIG-002-013, 041, 050 | `m04_skill_drift.go Apply()` | `TestM04RespectsUserModified` |
| REQ-MIG-002-014, 043 | `m05_legacy_cleanup.go Apply()` | `TestM05Respects30DayMtime` |
| REQ-MIG-002-015 | Each step's `Rollback()` | `TestAllStepsRollback` |
| REQ-MIG-002-020 | `m01.PreConditionsMet()` | `TestM01SkipWhenVersionsMatch` |
| REQ-MIG-002-021 | `m02.PreConditionsMet()` | `TestM02SkipWhenAgencyAbsent` |
| REQ-MIG-002-022 | `m03.PreConditionsMet()` | `TestM03SkipWhenScriptExists` |
| REQ-MIG-002-023 | `m04.PreConditionsMet()` | `TestM04SkipWhenAllSkillsPresent` |
| REQ-MIG-002-024 | `m05.PreConditionsMet()` | `TestM05SkipWhenNoTargets` |
| REQ-MIG-002-030 | `internal/core/migration/report.go` post-migration guidance | `TestM05ReportIncludesGitCommand` |
| REQ-MIG-002-031 | `m02_agency_archive.go` suffix logic | `TestM02SuffixCollision` |
| REQ-MIG-002-032 | Env check in each Apply | `TestVerboseMigrationOutput` |
| REQ-MIG-002-042 | `m01_template_version.go` comment preservation | `TestM01CommentsPreserved` |
| REQ-MIG-002-051 | `moai migrate v2-to-v3` e2e | `TestV2ToV3EndToEnd` |
| AC-MIG-002-01 ~ 12 | 해당 REQ 커버리지 전체 | `go test ./internal/core/migration/steps/...` + 10-project corpus e2e |

---

End of SPEC-V3-MIG-002.
