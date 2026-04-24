---
id: SPEC-V3-CLN-002
title: Legacy Code Removal & ADR-011 Comment Fix
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: Wave 4 SPEC writer
priority: P3 Low
phase: "v3.0.0 — Phase 1 — Foundation (ships via M05) + direct source edits"
module: "internal/cli/, internal/template/embed.go, coverage artifacts"
dependencies:
  - SPEC-V3-MIG-001
related_gap:
  - gm#186
  - gm#187
  - gm#188
  - gm#189
  - gm#193
related_theme: "Theme 9 — Internal Cleanup & Template Drift Resolution"
breaking: false
bc_id: []
lifecycle: spec-first
tags: "cleanup, legacy, code-debt, comment-drift, v3, housekeeping"
---

# SPEC-V3-CLN-002: Legacy Code Removal & ADR-011 Comment Fix

## HISTORY

| Version | Date       | Author | Description                                                |
|---------|------------|--------|------------------------------------------------------------|
| 0.1.0   | 2026-04-22 | Wave 4 | Initial SPEC draft covering legacy .bak, stale coverage, and ADR-011 comment drift |

---

## 1. Goal (목적)

moai-adk-go 저장소에 체크인된 pre-refactor 백업 파일(`.go.bak`)과 stale coverage 아티팩트를 제거하며, `internal/template/embed.go:8-12`의 ADR-011 관련 stale 주석을 현실과 일치하도록 수정한다. 추가로 Wave 1.6 §15.5에서 발견된 `InitDependencies()`의 핸들러 카운트 불일치(28 handler 호출 vs 27 EventType 상수)에 대한 ADR-style 주석을 `internal/cli/deps.go:151-186`에 추가하여 의도된 동작(AutoUpdateHandler가 SessionStart의 두 번째 handler compose)임을 명시한다. **이 SPEC은 사용자 가시 기능이 아닌 리포지토리 위생(housekeeping)이며, 대부분 직접 소스 편집으로 처리되고 일부만 M05(migration step)로 이전된다.**

### 1.1 배경

Wave 1.6 §15.10 및 §15.2에 따르면 다음 legacy 파일이 저장소에 체크인된 채 유지되고 있다:
- `internal/cli/glm.go.bak` (28,567 bytes, pre-April refactor backup, gm#186)
- `internal/cli/worktree/new_test.go.bak` (13,700 bytes, pre-April test backup, gm#187)
- `coverage.out` (16K, dated 2026-03-11, predates many SPEC changes, gm#188)
- `coverage.html` (77K, dated 2026-03-11)

또한 Wave 1.6 §4.6 및 §15.3은 `internal/template/embed.go:8-12`의 ADR-011 주석이 "runtime-generated files are excluded from embedded templates"라고 서술하지만 실제로는 `.claude/settings.json.tmpl`이 embedded되어 있고 runtime에 rendered 된다고 지적한다(gm#189). 마지막으로 §15.5(gm#193)는 `InitDependencies()`가 28개 handler 호출을 하는데 EventType 상수가 27개라 혼란을 줄 수 있으며, `AutoUpdateHandler`가 SessionStart의 두 번째 handler로 compose되는 패턴이 문서화되지 않았다고 지적한다.

### 1.2 비목표 (Non-Goals)

- ADR-011 자체(Zero Runtime Template Expansion)의 설계 재검토 — 주석만 실제 동작과 일치하도록 수정
- `internal/template/settings.go`의 SettingsGenerator 로직 변경 — 이미 동작 중
- `coverage.out` / `coverage.html` 생성 정책(언제 만들어야 하는지) 변경 — `.gitignore` 처리로 충분
- `InitDependencies()`의 handler registration 아키텍처 리팩토링 — compose pattern 의도 그대로 유지하고 문서화만
- `.moai-backups/` 폴더 제거 — SPEC-V3-MIG-002 (M02 step)(M02 agency archival)이 담당
- `.agency/` 관련 정리 — SPEC-V3-MIG-002 (M02 step)이 담당

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `moai doctor legacy-cleanup` CLI, direct source edits in files OUTSIDE `internal/core/migration/steps/` (e.g., .go.bak removal, stale ADR-011 comment fix in embed.go), deploy-target cleanup hooks.

- Tooling layer: `moai doctor legacy-cleanup` CLI + direct source edits outside `internal/core/migration/steps/` (M05 migration step Go file owned by SPEC-V3-MIG-002)

**직접 소스 편집 (non-migration, Phase 1 첫 PR에 포함):**
- `internal/template/embed.go:8-12`의 ADR-011 주석을 현재 동작과 일치하도록 재작성
- `internal/cli/deps.go:151-186`의 `InitDependencies()`에 AutoUpdateHandler compose pattern 설명 주석 추가
- `.gitignore`에 `coverage.out`, `coverage.html`, `*.go.bak` 패턴 추가 (재발 방지)
- `.gitignore`가 없는 경우 신설

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- ADR-011 설계 변경 또는 `internal/template/settings.go`의 SettingsGenerator 리팩토링
- `InitDependencies()`의 handler registration 구조 변경 (compose pattern 유지, 주석만 추가)
- `coverage.out`을 CI에서 자동 업로드/삭제하는 워크플로우 — 별도 SPEC 대상
- Go 모듈 버전 bump, dependency 갱신 — 본 SPEC 대상 아님
- 다른 `*.bak` 파일이 미래에 다시 생길 경우의 자동 방지 (linting rule) — pre-commit hook 확장은 별도 SPEC
- `.moai-backups/` 폴더 및 `.agency/` 관련 파일 정리 (SPEC-V3-MIG-002 (M02 step) 담당)
- **Migration step Go implementations** — owned by SPEC-V3-MIG-002.

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+)
- Claude Code v2.1.111+
- 영향 파일 (삭제):
  - `internal/cli/glm.go.bak`
  - `internal/cli/worktree/new_test.go.bak`
  - `coverage.out` (stale일 경우만)
  - `coverage.html` (stale일 경우만)
- 영향 파일 (수정):
  - `internal/template/embed.go` (lines 8-12 comment)
  - `internal/cli/deps.go` (lines 151-186, AutoUpdateHandler compose 주석 추가)
  - `.gitignore` (legacy 패턴 추가)
- 영향 파일 (참조): `internal/core/migration/steps/m05_*.go` (owned by SPEC-V3-MIG-002)
- 외부 레퍼런스: Wave 1.6 §4.6, §15.2, §15.3, §15.5, §15.10, gm#186-#189, gm#193, master §3.9

---

## 4. Assumptions (가정)

- SPEC-V3-MIG-001의 MigrationStep 인터페이스가 선행 구현되어 있다 (M05가 해당 인터페이스를 구현).
- `*.go.bak` 파일은 실제로 빌드에 포함되지 않는다 (`.bak` 확장자는 Go 툴체인이 ignore).
- `coverage.out` / `coverage.html` 삭제가 CI/개발자 워크플로우에 부정적 영향을 주지 않는다 (생성 즉시 덮어쓰기됨).
- ADR-011 주석 수정은 의미 변경이 아닌 정확성 개선이므로 별도 ADR 수정이 불필요하다.
- `InitDependencies()`의 compose pattern은 의도된 설계이다 (Wave 1.6 §15.5가 "intentional but undocumented"로 명시).
- `.gitignore`에 패턴 추가는 기존에 tracked된 파일(`coverage.out` 등)을 자동으로 untrack 하지 않으므로, M05의 삭제가 선행되어야 한다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-CLN-002-001**
The `moai doctor legacy-cleanup` subcommand **shall** invoke SPEC-V3-MIG-002's M05 migration step to scan 프로젝트 루트 기준 다음 파일 목록과 그 삭제 후보 상태를 사용자에게 보고한다:
- `internal/cli/glm.go.bak`
- `internal/cli/worktree/new_test.go.bak`
- `coverage.out` (30일 이상 오래된 경우)
- `coverage.html` (30일 이상 오래된 경우)

**REQ-CLN-002-002**
The `moai doctor legacy-cleanup` subcommand **shall** ensure SPEC-V3-MIG-002's M05 step이 삭제 전에 각 대상 파일을 `.moai/backups/<ISO-8601-timestamp>/legacy/`로 복사하여 rollback 경로를 확보하도록 하고, 백업 경로를 사용자에게 보고한다.

**REQ-CLN-002-003**
The `moai doctor legacy-cleanup` subcommand **shall** invoke M05 migration step (owned by SPEC-V3-MIG-002) which implements SPEC-V3-MIG-001's `MigrationStep` 인터페이스 (`Version()`, `ID() == "M05-legacy-cleanup"`, `Description()`, `IsIdempotent() bool = true`, `PreConditionsMet()`, `DryRun()`, `Apply()`, `Rollback()`) and surface its execution status + rollback handle to the user.

**REQ-CLN-002-004**
The direct source edit (non-migration) **shall** `internal/template/embed.go:8-12`의 주석을 다음 정확한 서술로 교체한다: "`.claude/settings.json.tmpl` and similar `.tmpl` files ARE embedded and rendered at deploy time via `internal/template/deployer.go`. Only runtime-mutated files (e.g., `settings.local.json`, manifest.json) are excluded from embedding. See `internal/template/settings.go` for the template-expansion path."

**REQ-CLN-002-005**
The direct source edit (non-migration) **shall** `internal/cli/deps.go:151-186`에 ADR-style 블록 주석을 추가하여 AutoUpdateHandler가 SessionStart의 두 번째 handler로 compose되는 패턴을 문서화한다 (CLAUDE.md Section 5 Agent Chain 구조와 일관).

**REQ-CLN-002-006**
The direct source edit (non-migration) **shall** `.gitignore`에 다음 패턴을 추가하여 legacy 파일 재발을 방지한다:
- `*.go.bak`
- `coverage.out`
- `coverage.html`
- `coverage.txt` (future-proofing)

**REQ-CLN-002-007**
The `.gitignore` 패턴 추가 **shall** 기존 `.gitignore` 내용을 보존하며, idempotent 하게 동작한다 (중복 라인 금지, 이미 존재하면 no-op).

### 5.2 Event-Driven Requirements

**REQ-CLN-002-008**
**When** M05가 각 파일을 삭제하면, the step **shall** "Removed: <path> (<size> bytes)" 라인을 stdout에 기록하고, 누적 삭제 바이트와 파일 수를 실행 완료 시 요약으로 emit한다.

**REQ-CLN-002-009**
**When** M05가 `coverage.out` 또는 `coverage.html`의 `mtime`을 확인할 때, the step **shall** 30일 이상 경과한 경우에만 삭제하고, 그렇지 않으면 "Skipped (fresh): <path> (mtime: <ISO-8601>)"을 stdout에 기록한다.

**REQ-CLN-002-010**
**When** 직접 소스 편집(ADR-011 comment, deps.go compose 주석) PR이 CI에서 실행되면, the test suite **shall** `internal/template/embed_test.go`에 새 주석이 기대 문구를 포함하는지 검증하는 간단한 grep-style 테스트를 포함한다 (regression prevention).

### 5.3 State-Driven Requirements

**REQ-CLN-002-011**
**While** 프로젝트에 `.bak` 파일이 존재하지 않는 상태에서, the M05 `PreConditionsMet()` **shall** false를 반환하여 step을 skip 한다 (idempotent 보장).

**REQ-CLN-002-012**
**While** `coverage.out` / `coverage.html`이 30일 미만으로 최근 생성된 상태에서, the M05 step **shall** 해당 파일을 건드리지 않고 legacy `.bak` 파일만 처리한다 (partial cleanup은 허용).

**REQ-CLN-002-013**
**While** M05가 dry-run 모드로 실행 중인 상태에서, the step **shall** 실제 파일 삭제를 수행하지 않고 "Would remove: <path> (<size> bytes)" 라인만 stdout에 emit한다.

### 5.4 Optional Requirements

**REQ-CLN-002-014**
**Where** 사용자가 `--keep-coverage` 플래그를 `moai migrate` 커맨드에 전달한 환경에서, the M05 step **shall** `coverage.out`과 `coverage.html`의 삭제를 skip하고 `.bak` 파일만 처리한다.

**REQ-CLN-002-015**
**Where** `.moai/config/sections/system.yaml`의 `migration.skip_steps`에 `M05`가 포함된 환경에서, the runner **shall** M05 step을 영구 skip한다.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-CLN-002-016 (Unwanted Behavior)**
**If** M05가 대상 파일 삭제 중 권한 오류(EACCES, EPERM)를 받으면, **then** step **shall** 즉시 실패로 기록하고 이미 삭제된 파일의 rollback을 시도하며, `MIGRATE_CLN002_PERMISSION_DENIED` 오류를 반환한다.

**REQ-CLN-002-017 (Unwanted Behavior)**
**If** `.moai/backups/<timestamp>/legacy/` 쓰기가 실패하면, **then** M05 step **shall** 어떤 파일도 삭제하지 않고 `MIGRATE_CLN002_BACKUP_FAILED` 오류를 반환한다 (rollback path가 없으면 apply하지 않음).

**REQ-CLN-002-018 (Unwanted Behavior)**
**If** Rollback이 호출되었지만 backup 디렉터리가 손상되거나 없으면, **then** step **shall** `MIGRATE_CLN002_ROLLBACK_FAILED` 오류를 반환하고 수동 복구 가이드를 stderr에 emit한다.

**REQ-CLN-002-019 (Complex: State + Unwanted)**
**While** M05 Apply 도중 SIGINT/SIGTERM이 수신되면 (SPEC-AGENCY-ABSORB-001 REQ-MIGRATE-013 패턴), **then** step **shall** 현재 파일 처리를 완료한 후 partial state를 `.moai/backups/<timestamp>/legacy/.tx-state.json`에 flush하고 종료 코드 130(SIGINT) 또는 143(SIGTERM)으로 종료한다. 재실행 시 `moai migrate --resume`으로 이어서 처리 가능하다.

**REQ-CLN-002-020 (Unwanted Behavior)**
**If** 직접 소스 편집 PR이 ADR-011 주석의 의미를 완전히 뒤집는 방향으로 수정되면 (예: "settings.json.tmpl is not embedded" 같은 잘못된 주장), **then** CI의 `internal/template/embed_test.go` **shall** 실패하여 머지를 차단한다 (REQ-CLN-002-010의 regression prevention 테스트 연동).

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-CLN-002-01**: Given v3.0 release 이전 저장소에 `internal/cli/glm.go.bak`과 `internal/cli/worktree/new_test.go.bak`이 tracked 상태 When v3.0 release PR merge 후 M05 Apply 실행 Then 두 파일이 저장소에서 제거되고 `.moai/backups/<timestamp>/legacy/`에 snapshot됨 (maps REQ-CLN-002-001, REQ-CLN-002-002).
- **AC-CLN-002-02**: Given `coverage.out`과 `coverage.html`이 30일 이상 경과된 상태 When M05 Apply 실행 Then 두 파일이 제거되고 `.gitignore`에 `coverage.out`, `coverage.html` 패턴이 추가됨 (maps REQ-CLN-002-001, REQ-CLN-002-006).
- **AC-CLN-002-03**: Given `internal/template/embed.go:8-12`의 기존 ADR-011 주석 상태 When 직접 소스 편집 PR merge Then 주석이 REQ-CLN-002-004의 정확한 문구로 교체됨 (maps REQ-CLN-002-004).
- **AC-CLN-002-04**: Given `internal/cli/deps.go:151-186`에 AutoUpdateHandler compose 주석 부재 상태 When 직접 소스 편집 PR merge Then ADR-style 블록 주석이 추가됨 (maps REQ-CLN-002-005).
- **AC-CLN-002-05**: Given `.gitignore` 부재 또는 legacy 패턴 미포함 상태 When 직접 소스 편집 PR merge Then `.gitignore`에 `*.go.bak`, `coverage.out`, `coverage.html`, `coverage.txt` 패턴이 추가됨 (maps REQ-CLN-002-006, REQ-CLN-002-007).
- **AC-CLN-002-06**: Given M05가 `--dry-run` 모드로 호출된 상태 When M05가 실행됨 Then 실제 파일 삭제 없이 "Would remove: <path> (<size> bytes)" 라인만 stdout에 출력됨 (maps REQ-CLN-002-013).
- **AC-CLN-002-07**: Given M05 Apply 후 `.moai/backups/<timestamp>/legacy/`에 snapshot 존재 When `moai migrate --rollback <timestamp>` 실행 Then 제거된 파일들이 원위치에 복원됨 (maps REQ-CLN-002-002).
- **AC-CLN-002-08**: Given `internal/template/embed.go`에 새 ADR-011 주석이 적용된 상태 When `go test ./internal/template/...` 실행 Then `internal/template/embed_test.go`의 regression test가 새 문구 존재를 확인하고 통과 (maps REQ-CLN-002-010, REQ-CLN-002-020).
- **AC-CLN-002-09**: Given `.bak` 파일 부재 상태 (M05 이미 적용) When M05 재실행 Then PreConditionsMet=false로 skip되고 migration_version counter 변동 없음 (maps REQ-CLN-002-011).
- **AC-CLN-002-10**: Given M05 구현체가 작성된 상태 When `go test ./internal/core/migration/steps/...` 실행 Then M05가 `MigrationStep` 인터페이스 8개 메서드(`Version`, `ID`, `Description`, `IsIdempotent`, `PreConditionsMet`, `DryRun`, `Apply`, `Rollback`) 모두 구현했음이 검증됨 (maps REQ-CLN-002-003).
- **AC-CLN-002-11**: Given M05가 `internal/cli/glm.go.bak` (28,567 bytes)를 삭제한 상태 When Apply 완료 Then stdout에 "Removed: internal/cli/glm.go.bak (28567 bytes)" 라인이 기록되고 누적 삭제 바이트/파일 수 요약이 emit됨 (maps REQ-CLN-002-008).
- **AC-CLN-002-12**: Given `coverage.out`의 mtime이 현재로부터 5일 전 상태 When M05 Apply 실행 Then 해당 파일은 삭제되지 않고 stdout에 "Skipped (fresh): coverage.out (mtime: <ISO-8601>)" 라인이 기록됨 (maps REQ-CLN-002-009).
- **AC-CLN-002-13**: Given `coverage.out`과 `coverage.html` 둘 다 30일 미만 상태 When M05 실행 Then coverage 파일은 건드리지 않고 `.bak` 파일만 처리됨 (maps REQ-CLN-002-012).
- **AC-CLN-002-14**: Given 사용자가 `--keep-coverage` 플래그를 전달한 상태 When M05 실행 Then coverage 파일 삭제가 skip되고 `.bak` 파일만 처리됨 (maps REQ-CLN-002-014).
- **AC-CLN-002-15**: Given `.moai/config/sections/system.yaml`에 `migration.skip_steps: [M05]` 설정된 상태 When runner 실행 Then M05가 영구 skip됨 (maps REQ-CLN-002-015).
- **AC-CLN-002-16**: Given M05가 대상 파일 삭제 중 EACCES 오류 수신 상태 When step 실행 Then 즉시 실패 기록, 이미 삭제된 파일 rollback 시도, `MIGRATE_CLN002_PERMISSION_DENIED` 오류 반환 (maps REQ-CLN-002-016).
- **AC-CLN-002-17**: Given `.moai/backups/<timestamp>/legacy/` 쓰기 실패 상태 When M05 실행 Then 어떤 파일도 삭제되지 않고 `MIGRATE_CLN002_BACKUP_FAILED` 오류 반환 (maps REQ-CLN-002-017).
- **AC-CLN-002-18**: Given Rollback 호출되었으나 backup 디렉터리 손상/부재 상태 When rollback 시도 Then `MIGRATE_CLN002_ROLLBACK_FAILED` 오류 반환 및 수동 복구 가이드가 stderr에 emit됨 (maps REQ-CLN-002-018).
- **AC-CLN-002-19**: Given M05 Apply 도중 SIGINT 수신 상태 When signal 처리 Then 현재 파일 처리 완료 후 partial state를 `.moai/backups/<timestamp>/legacy/.tx-state.json`에 flush하고 exit code 130으로 종료되며 `moai migrate --resume`으로 재개 가능함 (maps REQ-CLN-002-019).

---

## 7. Constraints (제약)

- 본 SPEC은 두 실행 경로로 분리: (a) M05 migration step, (b) 직접 소스 편집.
- 직접 소스 편집은 Phase 1의 첫 PR(v3.0.0-alpha.1 이전)에 포함되어 베이스라인을 정확히 맞춘다.
- M05는 SPEC-V3-MIG-001 framework의 일부로 실행되며 독립 실행 불가.
- Rollback은 파일 단위 restore만 지원 (`.moai/backups/<timestamp>/legacy/`에 snapshot).
- 플랫폼 호환: Windows에서 파일 삭제 시 `os.Remove` semantics 준수 (permission bits는 no-op).
- 9-direct-dep 정책: 신규 외부 의존성 금지.
- 주석 편집 시 HTML/markdown 포맷 깨짐 없이 Go block comment 유지 (gofmt 검증).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Fresh coverage artifact를 잘못 삭제 | 최신 coverage 손실 | REQ-CLN-002-009의 30일 check safeguard |
| `.bak` 파일이 현재 개발 중 branch에서 활발히 사용됨 | Dev workflow 방해 | Pre-condition check로 파일 존재만 탐지. 의도적으로 유지하려는 경우 `--keep-coverage` 또는 `skip_steps: [M05]` |
| ADR-011 주석 수정이 설계 변경으로 오해 | 리뷰어 혼란 | PR description에 Non-Goal(§1.2) 명시, 주석은 문서화만 |
| Compose pattern 주석 추가가 코드 아키텍처 변경으로 오해 | 리뷰어 혼란 | ADR-style 블록 주석 (`// ADR: AutoUpdateHandler composes ...`) 명확히 |
| `.gitignore` 중복 패턴 추가 시 linter 경고 | 사소함 | REQ-CLN-002-007의 idempotent 로직 |
| SIGINT/SIGTERM 처리가 SPEC-V3-MIG-001 framework 책임인지 M05 자체인지 모호 | 중복 구현 | REQ-CLN-002-019은 framework 패턴을 재사용한다고 명시 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-MIG-001 (Versioned migration framework): M05 step의 MigrationStep 인터페이스, Runner, backup/rollback 경로, dry-run 지원이 모두 SPEC-V3-MIG-001에 의존.

### 9.2 Blocks

- SPEC-V3-MIGRATE-001 (`moai migrate v2-to-v3` tool): v2→v3 전환 도구가 M05를 실행.
- Phase 1 첫 PR (v3.0.0-alpha.1): 직접 소스 편집 부분(embed.go comment, deps.go comment, .gitignore)은 Phase 1 초기에 반드시 포함.

### 9.3 Related

- SPEC-V3-CLN-001 (Template drift resolution): M01/M03/M04와 동일한 runner 사이클에서 M05가 실행됨.
- SPEC-V3-MIG-002 (M02 step) (Agency archival + docs drift): M02와 함께 실행되지만 scope가 분리됨 — MIG-002 (M02 step)은 `.agency/` 및 docs-site 정리 담당.
- SPEC-HOOK-001 (Compiled Hook System): `InitDependencies()`의 compose pattern 문서화는 해당 SPEC의 REQ-HOOK-030 (SessionStart handler) 문맥에서 이해됨.

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 요구사항 ID는 `plan.md` 마일스톤(Wave 5)과 §6 Acceptance Criteria 시나리오로 역참조된다.
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-CLN-002:REQ-CLN-002-<NNN>` 주석 부착.
- 총 REQ 개수: 20개 (Ubiquitous 7, Event-Driven 3, State-Driven 3, Optional 2, Complex 5).
- 예상 코드 구현 경로:
  - `internal/cli/doctor_legacy_cleanup.go` (서브커맨드)
  - `internal/template/embed.go` (주석 수정, lines 8-12)
  - `internal/template/embed_test.go` (regression test for comment content)
  - `internal/cli/deps.go` (주석 추가, lines 151-186)
  - `.gitignore` (신설 또는 수정)
- Gap matrix 추적: gm#186 (glm.go.bak), gm#187 (new_test.go.bak), gm#188 (stale coverage), gm#189 (ADR-011 comment drift), gm#193 (handler count 불일치 문서화).
- v3-master §3.9 Theme 9 §4 (ADR-011 comment fix is direct source edit, ships in v3.0 first PR), §5.1 M05 cleanup, §8.7 SPEC-V3-CLN-002.
- Wave 1.6 source anchors: §4.6 (ADR-011 comment drift evidence), §15.2 (stale coverage), §15.3 (ADR-011 stale comment), §15.5 (handler count 28 vs 27), §15.10 (glm.go.bak / new_test.go.bak).

---

End of SPEC.
