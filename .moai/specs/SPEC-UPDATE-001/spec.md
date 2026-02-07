# SPEC-UPDATE-001: Self-Update & 3-Way Merge System

---
id: SPEC-UPDATE-001
title: Self-Update & 3-Way Merge System
status: Completed
priority: High
phase: "Phase 3 - Automation"
created: 2026-02-03
modules:
  - internal/update/
  - internal/merge/
estimated_loc: ~1,800
dependencies:
  - SPEC-TEMPLATE-001 (manifest)
  - SPEC-CONFIG-001
adr_references:
  - ADR-007 (File Manifest Provenance)
  - ADR-008 (3-Way Merge for Template Updates)
  - ADR-009 (Self-Update via Binary Replacement)
resolves_issues:
  - "#253"
  - "#296"
  - "#159"
  - "#312"
  - "#246"
  - "#187"
  - "#318"
tags: SPEC-UPDATE-001
lifecycle: spec-anchored
---

## HISTORY

| 날짜 | 버전 | 변경사항 |
|------|------|----------|
| 2026-02-03 | 1.0.0 | 최초 SPEC 작성 |

---

## 1. 개요

### 1.1 배경

Python MoAI-ADK의 가장 복잡한 모듈인 `update.py`(3,162 LOC)는 38회 수정되었음에도 불구하고 파괴적 덮어쓰기, 패키지 매니저 의존성 실패, 마이그레이션 충돌 등 15건의 이슈를 해결하지 못했다. 이 SPEC은 Go Edition에서 바이너리 자체 업데이트 시스템과 Git 스타일 3-Way Merge 엔진을 완전히 재설계하여 해당 문제를 근본적으로 해결한다.

### 1.2 목적

- PyPI/uv/pipx 의존성 체인을 제거하고 GitHub Releases 기반 바이너리 자체 업데이트로 대체
- 파일 매니페스트(`.moai/manifest.json`)와 3-Way Merge를 결합하여 사용자 커스터마이징을 보존하면서 템플릿 업데이트 적용
- 실패 시 원자적 롤백으로 업데이트 안정성 보장

### 1.3 범위

**포함 범위:**

- `internal/update/`: checker.go, updater.go, rollback.go, orchestrator.go
- `internal/merge/`: three_way.go, strategies.go, conflict.go, differ.go
- `moai update` CLI 커맨드와의 통합

**제외 범위:**

- `internal/manifest/` (SPEC-TEMPLATE-001에서 다룸)
- `internal/template/` (SPEC-TEMPLATE-001에서 다룸)
- Homebrew tap 자동 업데이트
- `go install` 경로를 통한 업데이트

---

## 2. 환경 (Environment)

### 2.1 시스템 환경

| 항목 | 사양 |
|------|------|
| 언어 | Go 1.22+ |
| 배포 형태 | 단일 바이너리 (CGO_ENABLED=0) |
| 대상 플랫폼 | darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64, windows/arm64 |
| 외부 API | GitHub Releases API (REST, HTTPS) |
| 파일시스템 | 프로젝트 루트 내 `.moai/manifest.json` |

### 2.2 의존 모듈

| 모듈 | 역할 | SPEC |
|------|------|------|
| `internal/manifest/` | 파일 출처 추적, 해시 비교 | SPEC-TEMPLATE-001 |
| `internal/template/` | go:embed 템플릿 추출 | SPEC-TEMPLATE-001 |
| `internal/config/` | 업데이트 설정 (자동 확인 주기, 프록시 등) | SPEC-CONFIG-001 |
| `internal/ui/` | 진행 표시 및 결과 요약 | SPEC-UI-001 |
| `pkg/version/` | 현재 바이너리 버전 | - |

### 2.3 외부 의존성

- GitHub Releases API (`api.github.com/repos/modu-ai/moai-adk-go/releases`)
- 네트워크 접근 (HTTPS)
- 파일시스템 쓰기 권한 (바이너리 교체, 임시 파일)

---

## 3. 가정 (Assumptions)

### 3.1 기술 가정

- [A-01] GitHub Releases API는 안정적으로 사용 가능하며, 릴리즈에는 플랫폼별 바이너리와 체크섬이 포함된다.
- [A-02] `internal/manifest/`가 먼저 구현되어 `.moai/manifest.json`의 읽기/쓰기가 가능하다 (SPEC-TEMPLATE-001 선행 조건).
- [A-03] goreleaser가 릴리즈 에셋에 SHA-256 체크섬 파일을 포함한다.
- [A-04] 바이너리 교체 시 OS별 원자적 rename 연산이 지원된다 (Unix: `os.Rename`, Windows: `MoveFileEx`).

### 3.2 비즈니스 가정

- [A-05] 사용자는 인터넷 연결이 있는 환경에서 업데이트를 실행한다.
- [A-06] 사용자가 수정한 템플릿 파일은 업데이트 과정에서 보존되어야 한다.
- [A-07] 충돌이 발생한 파일은 `.conflict` 파일로 생성되어 사용자가 수동 해결할 수 있어야 한다.

### 3.3 위험 가정

- [A-08] 바이너리 교체 중 프로세스가 중단될 경우 롤백 메커니즘이 필요하다.
- [A-09] 네트워크 불안정으로 다운로드가 실패할 수 있으며, 부분 다운로드 상태를 정리해야 한다.

---

## 4. 요구사항 (Requirements) - EARS 형식

### 4.1 Self-Update 모듈 (`internal/update/`)

#### REQ-UPD-001: 버전 확인 (Checker)

**WHEN** 사용자가 `moai update`를 실행하거나 자동 확인 주기가 도래하면 **THEN** 시스템은 GitHub Releases API에서 최신 버전 메타데이터(버전, URL, 체크섬, 날짜)를 조회해야 한다.

#### REQ-UPD-002: 업데이트 가용성 판단

**WHEN** 최신 버전 정보가 조회되면 **THEN** 시스템은 현재 바이너리 버전(`pkg/version`)과 시맨틱 버전 비교를 수행하여 업데이트 가용 여부를 판단해야 한다.

#### REQ-UPD-003: 플랫폼 바이너리 다운로드

**WHEN** 업데이트가 가용하고 사용자가 업데이트를 승인하면 **THEN** 시스템은 현재 플랫폼(`runtime.GOOS`, `runtime.GOARCH`)에 해당하는 바이너리를 임시 디렉토리에 다운로드해야 한다.

#### REQ-UPD-004: 체크섬 검증

**WHEN** 바이너리 다운로드가 완료되면 **THEN** 시스템은 SHA-256 체크섬을 검증하여 다운로드 무결성을 확인해야 한다.

시스템은 **항상** 체크섬이 불일치할 경우 다운로드된 파일을 삭제하고 `ErrChecksumMismatch` 오류를 반환해야 한다.

#### REQ-UPD-005: 원자적 바이너리 교체

**WHEN** 체크섬 검증이 완료되면 **THEN** 시스템은 다음 순서로 바이너리를 교체해야 한다:
1. 새 바이너리를 임시 파일에 기록
2. `os.Rename`을 통한 원자적 교체 (Unix) 또는 `MoveFileEx` (Windows)

시스템은 교체 **하지 않아야 한다** -- 단, 임시 파일이 실행 가능하고 `moai version` 서브커맨드가 정상 실행되는 것이 확인된 경우에만 교체를 수행해야 한다.

#### REQ-UPD-006: 롤백 메커니즘

**WHEN** 바이너리 교체 또는 검증이 실패하면 **THEN** 시스템은 백업된 이전 바이너리로 자동 롤백해야 한다.

**IF** 롤백 자체도 실패하면 **THEN** 시스템은 백업 경로를 포함한 복구 안내 메시지를 사용자에게 출력해야 한다.

#### REQ-UPD-007: 업데이트 전 백업 생성

시스템은 **항상** 바이너리 교체를 시도하기 전에 현재 바이너리의 백업을 생성해야 한다. 백업 경로는 `{binary_path}.backup.{timestamp}` 형식을 따른다.

#### REQ-UPD-008: 업데이트 오케스트레이션

**WHEN** 사용자가 `moai update`를 실행하면 **THEN** Orchestrator는 다음 전체 파이프라인을 순차적으로 실행해야 한다:
1. 버전 확인 (Checker)
2. 사용자 승인 (CLI 인터랙션)
3. 백업 생성 (Rollback)
4. 바이너리 다운로드 (Updater)
5. 체크섬 검증 (Updater)
6. 템플릿 추출 및 매니페스트 기반 머지 (Template + Manifest + Merge)
7. 원자적 바이너리 교체 (Updater)
8. 새 바이너리 검증 (`moai version`)
9. 결과 요약 출력 (UI)

#### REQ-UPD-009: 컨텍스트 기반 취소 및 타임아웃

시스템은 **항상** 모든 네트워크 작업(API 호출, 다운로드)에 `context.Context`를 전파하여 취소와 타임아웃을 지원해야 한다.

**가능하면** 다운로드 진행률을 UI에 표시한다.

#### REQ-UPD-010: 임시 파일 정리

시스템은 **항상** 업데이트 성공/실패 여부와 관계없이 임시 다운로드 파일을 `defer`로 정리해야 한다.

---

### 4.2 3-Way Merge 모듈 (`internal/merge/`)

#### REQ-MRG-001: 3-Way Merge 엔진

**WHEN** 매니페스트에서 `user_modified` 또는 `template_managed`(변경 감지) 파일이 식별되면 **THEN** 시스템은 base(마지막 배포 템플릿), current(사용자의 현재 파일), updated(새 템플릿) 세 버전을 입력으로 3-Way Merge를 수행해야 한다.

#### REQ-MRG-002: 파일 타입별 머지 전략 선택

**WHEN** `MergeFile`이 호출되면 **THEN** `StrategySelector`는 파일 경로와 확장자를 기반으로 다음 6가지 전략 중 하나를 선택해야 한다:

| 전략 | 대상 파일 | 설명 |
|------|-----------|------|
| `LineMerge` | `.md`, `.txt` (기본) | 라인 단위 diff 머지 |
| `YAMLDeep` | `.yaml`, `.yml` | YAML 구조 보존 deep merge |
| `JSONMerge` | `.json` | JSON 객체 구조 보존 merge |
| `SectionMerge` | `CLAUDE.md` | 섹션 단위 머지 (사용자 섹션 보존) |
| `EntryMerge` | `.gitignore` | 엔트리 기반 머지 (신규 추가, 기존 유지) |
| `Overwrite` | 바이너리, 비머지 파일 | 전체 교체 |

#### REQ-MRG-003: LineMerge 전략

**WHEN** `LineMerge` 전략이 선택되면 **THEN** 시스템은 라인 단위 diff 알고리즘을 적용하여:
- base와 current 간의 사용자 변경을 감지
- base와 updated 간의 템플릿 변경을 감지
- 양측 변경이 겹치지 않는 경우 자동 머지
- 양측 변경이 동일 라인에서 충돌하는 경우 `Conflict`를 기록

#### REQ-MRG-004: YAMLDeep 전략

**WHEN** `YAMLDeep` 전략이 선택되면 **THEN** 시스템은:
- YAML을 파싱하여 키-값 트리 구조로 비교
- 사용자가 추가/수정한 키를 보존
- 템플릿에서 추가된 새 키를 반영
- 동일 키에 대한 양측 변경이 충돌할 경우 `Conflict`를 기록

#### REQ-MRG-005: JSONMerge 전략

**WHEN** `JSONMerge` 전략이 선택되면 **THEN** 시스템은:
- JSON을 파싱하여 객체/배열 구조를 비교
- `json.Marshal`/`json.Unmarshal`을 사용하여 유효한 JSON을 보장
- 사용자 변경 키를 보존하고 새 키를 반영
- 배열 머지는 중복 제거 후 합집합 적용

#### REQ-MRG-006: SectionMerge 전략

**WHEN** `SectionMerge` 전략이 선택되고 대상이 `CLAUDE.md`이면 **THEN** 시스템은:
- Markdown 헤딩(`##`, `###`)을 기준으로 섹션을 분리
- 사용자가 추가/수정한 섹션을 보존
- 템플릿에서 추가된 새 섹션을 적절한 위치에 삽입
- 동일 섹션 내 양측 변경은 `LineMerge`로 처리

#### REQ-MRG-007: EntryMerge 전략

**WHEN** `EntryMerge` 전략이 선택되면 **THEN** 시스템은:
- 각 라인을 독립 엔트리로 취급 (`.gitignore` 패턴)
- 사용자가 추가한 엔트리를 보존
- 템플릿에서 추가된 새 엔트리를 추가
- 중복 엔트리는 제거
- 사용자가 삭제한 엔트리는 재추가하지 않음

#### REQ-MRG-008: Overwrite 전략

**WHEN** `Overwrite` 전략이 선택되면 **THEN** 시스템은 updated 내용으로 전체 교체하고, 기존 파일은 `.backup` 확장자로 보존해야 한다.

#### REQ-MRG-009: 충돌 감지 및 보고

**IF** 3-Way Merge 중 자동 해결이 불가능한 충돌이 발생하면 **THEN** 시스템은:
- `MergeResult.HasConflict`를 `true`로 설정
- `Conflict` 구조체에 시작/끝 라인, base/current/updated 내용을 기록
- 충돌 파일 경로에 `.conflict` 확장자 파일을 생성하여 Git 스타일 충돌 마커를 포함

시스템은 충돌 발생 시 원본 파일을 수정**하지 않아야 한다**. 충돌 해결은 사용자가 `.conflict` 파일을 검토한 후 수동으로 수행한다.

#### REQ-MRG-010: Diff 생성기

시스템은 **항상** 머지 결과에 대해 라인 수준 diff를 생성할 수 있어야 한다. Diff 출력은 unified diff 형식을 따른다.

---

### 4.3 통합 요구사항 (Update + Merge)

#### REQ-INT-001: 매니페스트 기반 업데이트 결정

**WHEN** 업데이트 오케스트레이션 중 템플릿 파일 처리 단계에 도달하면 **THEN** 시스템은 매니페스트의 `provenance`를 기준으로 다음과 같이 처리해야 한다:

| Provenance | 조건 | 동작 |
|-----------|------|------|
| `template_managed` | 해시 변경 없음 | 안전 덮어쓰기 |
| `template_managed` | 해시 변경 감지 | `user_modified`로 승격 후 3-way merge |
| `user_modified` | - | 3-way merge |
| `user_created` | - | 건너뜀 (절대 수정하지 않음) |
| `deprecated` | - | 사용자 알림, 파일 유지 |

#### REQ-INT-002: 업데이트 결과 요약

**WHEN** 업데이트가 완료되면 **THEN** 시스템은 `UpdateResult`를 통해 다음 통계를 보고해야 한다:
- `FilesUpdated`: 안전 덮어쓰기된 파일 수
- `FilesMerged`: 3-way merge 성공 파일 수
- `FilesConflicted`: 충돌 발생 파일 수
- `FilesSkipped`: 건너뛴 파일 수 (user_created)
- `RollbackPath`: 롤백 백업 경로

---

## 5. 명세 (Specifications)

### 5.1 인터페이스 정의

#### Update 모듈 인터페이스

```go
// Checker: GitHub Releases API 버전 조회
type Checker interface {
    CheckLatest(ctx context.Context) (*VersionInfo, error)
    IsUpdateAvailable(current string) (bool, *VersionInfo, error)
}

// Updater: 바이너리 다운로드 및 교체
type Updater interface {
    Download(ctx context.Context, version *VersionInfo) (string, error)
    Replace(ctx context.Context, newBinaryPath string) error
}

// Rollback: 백업 및 복원
type Rollback interface {
    CreateBackup() (string, error)
    Restore(backupPath string) error
}

// Orchestrator: 전체 업데이트 워크플로우 조율
type Orchestrator interface {
    Update(ctx context.Context) (*UpdateResult, error)
}
```

#### Merge 모듈 인터페이스

```go
// Engine: 3-Way Merge 연산
type Engine interface {
    ThreeWayMerge(base, current, updated []byte) (*MergeResult, error)
    MergeFile(ctx context.Context, path string, base, current, updated []byte) (*MergeResult, error)
}

// StrategySelector: 파일 경로 기반 전략 선택
type StrategySelector interface {
    SelectStrategy(path string) MergeStrategy
}
```

### 5.2 데이터 구조

```go
// MergeStrategy: 6가지 머지 알고리즘
type MergeStrategy string

const (
    LineMerge    MergeStrategy = "line_merge"
    YAMLDeep     MergeStrategy = "yaml_deep"
    JSONMerge    MergeStrategy = "json_merge"
    SectionMerge MergeStrategy = "section_merge"
    EntryMerge   MergeStrategy = "entry_merge"
    Overwrite    MergeStrategy = "overwrite"
)

// MergeResult: 머지 연산 결과
type MergeResult struct {
    Content     []byte
    HasConflict bool
    Conflicts   []Conflict
    Strategy    MergeStrategy
}

// Conflict: 충돌 영역 상세
type Conflict struct {
    StartLine int
    EndLine   int
    Base      string
    Current   string
    Updated   string
}

// VersionInfo: GitHub Release 메타데이터
type VersionInfo struct {
    Version  string    `json:"version"`
    URL      string    `json:"url"`
    Checksum string    `json:"checksum"`
    Date     time.Time `json:"date"`
}

// UpdateResult: 업데이트 결과 요약
type UpdateResult struct {
    PreviousVersion string
    NewVersion      string
    FilesUpdated    int
    FilesMerged     int
    FilesConflicted int
    FilesSkipped    int
    RollbackPath    string
}
```

### 5.3 오류 유형

```go
// internal/merge/
var (
    ErrMergeConflict    = errors.New("merge: unresolvable conflict detected")
    ErrMergeUnsupported = errors.New("merge: file type not supported for merge")
)

// internal/update/
var (
    ErrUpdateNotAvail   = errors.New("update: no update available")
    ErrDownloadFailed   = errors.New("update: binary download failed")
    ErrChecksumMismatch = errors.New("update: checksum verification failed")
    ErrReplaceFailed    = errors.New("update: binary replacement failed")
    ErrRollbackFailed   = errors.New("update: rollback restoration failed")
)
```

### 5.4 모듈 의존성 그래프

```
internal/cli/update.go
    |
    v
internal/update/orchestrator.go
    |
    +---> internal/update/checker.go     (GitHub API)
    +---> internal/update/updater.go     (다운로드/교체)
    +---> internal/update/rollback.go    (백업/복원)
    +---> internal/manifest/             (파일 출처 조회)
    +---> internal/template/deployer.go  (새 템플릿 추출)
    +---> internal/merge/three_way.go    (3-way merge)
    |         |
    |         +---> internal/merge/strategies.go  (전략 선택)
    |         +---> internal/merge/conflict.go    (충돌 처리)
    |         +---> internal/merge/differ.go      (diff 생성)
    +---> internal/config/               (업데이트 설정)
    +---> internal/ui/progress.go        (진행률 표시)
    +---> pkg/version/                   (현재 버전)
```

### 5.5 비기능 요구사항

| 항목 | 목표 |
|------|------|
| 버전 확인 시간 | < 2초 |
| 바이너리 다운로드 | 네트워크 속도 의존, 타임아웃 300초 |
| 3-Way Merge (단일 파일) | < 100ms |
| 3-Way Merge (전체 프로젝트, ~200 파일) | < 5초 |
| 바이너리 교체 | < 1초 (원자적 rename) |
| 롤백 | < 1초 |
| 메모리 사용 | 파일당 < 10MB (대용량 파일은 스트리밍) |
| 테스트 커버리지 | >= 85% |

---

## 6. 추적성 (Traceability)

| 요구사항 ID | 해결 이슈 | ADR | 구현 파일 |
|------------|-----------|-----|-----------|
| REQ-UPD-001 | #253, #296 | ADR-009 | checker.go |
| REQ-UPD-002 | #253 | ADR-009 | checker.go |
| REQ-UPD-003 | #159, #312 | ADR-009 | updater.go |
| REQ-UPD-004 | #159 | ADR-009 | updater.go |
| REQ-UPD-005 | #312 | ADR-009 | updater.go |
| REQ-UPD-006 | #312 | ADR-009 | rollback.go |
| REQ-UPD-007 | #312 | ADR-009 | rollback.go |
| REQ-UPD-008 | #253, #296, #159, #312 | ADR-009 | orchestrator.go |
| REQ-MRG-001 | #246, #187, #318 | ADR-008 | three_way.go |
| REQ-MRG-002 | #246, #187 | ADR-008 | strategies.go |
| REQ-MRG-003 | #246 | ADR-008 | three_way.go, differ.go |
| REQ-MRG-004 | #187, #318 | ADR-008 | strategies.go |
| REQ-MRG-005 | #318 | ADR-008 | strategies.go |
| REQ-MRG-006 | #246, #187 | ADR-008 | strategies.go |
| REQ-MRG-007 | #246 | ADR-008 | strategies.go |
| REQ-MRG-008 | #187 | ADR-008 | strategies.go |
| REQ-MRG-009 | #246, #187, #318 | ADR-007, ADR-008 | conflict.go |
| REQ-MRG-010 | - | ADR-008 | differ.go |
| REQ-INT-001 | #246, #187, #318 | ADR-007 | orchestrator.go |
| REQ-INT-002 | - | ADR-009 | orchestrator.go |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 87.6% (update) / 90.3% (merge)

### Summary

Self-update system implemented with GitHub API-based version checking, binary download with checksum verification, atomic binary replacement, and automatic rollback on failure. 3-way merge engine supports YAML, JSON, and TOML file types with strategy-based conflict resolution (ours, theirs, union). Orchestrator coordinates the full update lifecycle: check, download, backup, deploy templates, merge configurations, and rollback if needed. Integrated with manifest module for file provenance tracking during template updates.

### Files Created

- `internal/update/checker.go`
- `internal/update/checker_test.go`
- `internal/update/orchestrator.go`
- `internal/update/orchestrator_test.go`
- `internal/update/rollback.go`
- `internal/update/rollback_test.go`
- `internal/update/types.go`
- `internal/update/updater.go`
- `internal/update/updater_test.go`
- `internal/merge/conflict.go`
- `internal/merge/conflict_test.go`
- `internal/merge/coverage_extra_test.go`
- `internal/merge/differ.go`
- `internal/merge/differ_test.go`
- `internal/merge/strategies.go`
- `internal/merge/strategies_test.go`
- `internal/merge/three_way.go`
- `internal/merge/three_way_test.go`
- `internal/merge/types.go`
