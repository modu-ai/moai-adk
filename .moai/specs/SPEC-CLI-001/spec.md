# SPEC-CLI-001: CLI Command Composition & Integration

---
spec_id: SPEC-CLI-001
title: CLI Command Composition & Integration
phase: "Phase 5 - CLI (Final, composition root)"
status: Completed
priority: Medium
created: 2026-02-03
lifecycle: spec-anchored
dependencies:
  - SPEC-INIT-001
  - SPEC-UPDATE-001
  - SPEC-HOOK-001
  - SPEC-TOOL-001
  - SPEC-GIT-001
modules:
  - internal/cli/
  - internal/core/integration/
  - internal/core/migration/
  - pkg/
estimated_loc: ~1,500
tags: [cli, composition-root, di-wiring, integration, migration, cobra, pkg]
---

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|----------|
| 2026-02-03 | 1.0.0 | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 시스템 컨텍스트

MoAI-ADK (Go Edition)는 단일 바이너리로 배포되는 개발 도구로, Cobra CLI 프레임워크 기반의 커맨드 라우팅과 의존성 주입(DI)을 통해 모든 도메인 모듈을 통합한다. 이 SPEC은 **Composition Root** -- 모든 모듈이 결합되는 최종 통합 지점을 정의한다.

### 1.2 기술 스택

| 구성 요소 | 기술 | 버전 |
|----------|------|------|
| CLI 프레임워크 | Cobra | v1.10.2 |
| 설정 관리 | Viper | v1.18+ |
| 언어 | Go | 1.22+ |
| 로깅 | log/slog (stdlib) | Go 1.22 |
| 테스팅 | testing (stdlib) + testify | v1.9+ |

### 1.3 현재 구현 상태

| 파일 | LOC | 상태 |
|------|-----|------|
| `cmd/moai/main.go` | 14 | 완료 (thin entry point) |
| `internal/cli/root.go` | 30 | 완료 (root command) |
| `internal/cli/version.go` | 27 | 완료 |
| `internal/cli/init.go` | 23 | Stub (TODO) |
| `internal/cli/doctor.go` | 23 | Stub (TODO) |
| `internal/cli/status.go` | 23 | Stub (TODO) |
| `pkg/version/version.go` | 31 | 완료 |
| `pkg/models/config.go` | 25 | 완료 (기본 타입) |
| `pkg/models/project.go` | 21 | 완료 (기본 타입) |
| `pkg/utils/logger.go` | 48 | 완료 |
| `pkg/utils/path.go` | 52 | 완료 |

### 1.4 선행 SPEC 의존성

이 SPEC은 **Composition Root**로서 모든 이전 Phase의 도메인 모듈에 의존한다.

| 선행 SPEC | 제공 모듈 | CLI 연결 지점 |
|----------|----------|-------------|
| SPEC-INIT-001 | `internal/config/`, `internal/template/`, `internal/manifest/`, `internal/ui/` | `init.go` |
| SPEC-UPDATE-001 | `internal/update/`, `internal/merge/` | `update.go` |
| SPEC-HOOK-001 | `internal/hook/` | `hook.go` |
| SPEC-TOOL-001 | `internal/rank/` | `rank.go` |
| SPEC-GIT-001 | `internal/core/git/` | `worktree/*.go` |

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- **A1**: 모든 선행 SPEC의 도메인 모듈이 Go 인터페이스를 export하며, 구체 타입은 `internal/` 내부에 캡슐화되어 있다.
- **A2**: 각 도메인 모듈은 생성자 함수(`New*()`)를 통해 인스턴스화되고, 의존성은 생성자 파라미터로 주입된다.
- **A3**: `go:embed`로 번들된 템플릿 파일시스템(`embed.FS`)이 `internal/template/` 에서 제공된다.
- **A4**: Cobra 커맨드 등록은 Go의 `init()` 함수를 통해 수행되며, 순환 참조가 발생하지 않는다.
- **A5**: `pkg/` 패키지는 외부 도구에서도 import 가능한 안정적 공개 API를 제공한다.

### 2.2 비즈니스 가정

- **B1**: `cc`와 `glm`은 별도의 top-level 커맨드로 유지된다 (`switch` 커맨드 사용하지 않음).
- **B2**: Python 에디션의 `language`, `analyze` 커맨드는 드롭된다.
- **B3**: worktree 서브커맨드는 Python의 10개에서 7개로 통합된다 (recover, done, config 제거).
- **B4**: 버전 마이그레이션은 `.moai/config/` YAML 형식 간 하위 호환성을 유지한다.

### 2.3 신뢰도

| 가정 | 신뢰도 | 검증 방법 |
|------|--------|----------|
| A1 | High | 이전 SPEC 코드 리뷰로 인터페이스 존재 확인 |
| A2 | High | Go DI 패턴은 생성자 주입이 관례 |
| A3 | High | ADR-003에 embed 결정 명시 |
| A4 | Medium | init() 순서에 따른 등록 순서 의존성 존재 가능 |
| B1 | High | cli-comparison.md에 명시 |

---

## 3. Requirements (요구사항)

### 3.1 DI Wiring (의존성 주입 구성) -- Composition Root

#### REQ-CLI-001: 의존성 주입 초기화

시스템은 **항상** `cmd/moai/main.go` 또는 `internal/cli/root.go`의 `init()` 체인에서 모든 도메인 의존성을 생성자 주입 방식으로 구성해야 한다.

```
DI 구성 순서:
1. pkg/utils/logger.go     -- 로거 초기화
2. internal/config/         -- 설정 매니저 (Viper + typed struct)
3. internal/hook/           -- Hook 레지스트리 + 프로토콜
4. internal/template/       -- 템플릿 배포자 (embeddedFS)
5. internal/manifest/       -- 매니페스트 매니저
6. internal/core/git/       -- Git 리포지토리
7. internal/core/project/   -- 프로젝트 초기화자
8. internal/update/         -- 업데이트 오케스트레이터
9. internal/rank/           -- 랭킹 클라이언트
10. internal/core/migration/ -- 버전 마이그레이터
11. internal/core/integration/ -- 통합 테스트 엔진
```

#### REQ-CLI-002: 의존성 주입 패턴

시스템은 **항상** 인터페이스 기반 의존성 주입을 사용해야 한다. 구체 타입에 대한 직접 의존은 Composition Root에서만 허용된다.

```go
// 패턴 예시 (Composition Root에서만)
cfg := config.New()
hookRegistry := hook.NewRegistry()
hookProtocol := hook.NewProtocol()
templateDeployer := template.NewDeployer(embeddedFS)
manifestMgr := manifest.NewManager()
gitRepo := git.NewRepository(projectRoot)
migrator := migration.NewMigrator(cfg, manifestMgr)
integrationEngine := integration.NewEngine(cfg)
```

### 3.2 CLI 커맨드 라우팅

#### REQ-CLI-010: Root 커맨드

시스템은 **항상** `moai` 루트 커맨드를 제공해야 하며, 서브커맨드 없이 실행 시 도움말을 표시해야 한다.

#### REQ-CLI-011: Version 커맨드 (완료)

**WHEN** 사용자가 `moai version` 또는 `moai --version`을 실행하면 **THEN** 시스템은 버전, 커밋 해시, 빌드 날짜를 포함한 버전 정보를 출력해야 한다.

#### REQ-CLI-012: Init 커맨드

**WHEN** 사용자가 `moai init [path]`를 실행하면 **THEN** 시스템은 SPEC-INIT-001에서 정의된 프로젝트 초기화 로직을 위임 실행해야 한다.

- Flag: `--non-interactive, -y` (non-interactive 모드)
- Flag: `--mode [personal|team]` (프로젝트 모드)
- Flag: `--locale [ko|en|ja|zh]` (언어 설정)
- Flag: `--language TEXT` (프로그래밍 언어)
- Flag: `--force` (재초기화 강제)

#### REQ-CLI-013: Doctor 커맨드

**WHEN** 사용자가 `moai doctor`를 실행하면 **THEN** 시스템은 시스템 진단을 수행하고 결과를 출력해야 한다.

- Flag: `--verbose, -v` (상세 출력)
- Flag: `--fix` (수정 제안)
- Flag: `--export PATH` (JSON 내보내기)
- Flag: `--check TEXT` (특정 도구 점검)

#### REQ-CLI-014: Status 커맨드

**WHEN** 사용자가 `moai status`를 실행하면 **THEN** 시스템은 프로젝트 상태 (이름, 타입, 버전, config 상태, SPEC 진행 상황)를 출력해야 한다.

#### REQ-CLI-015: Update 커맨드

**WHEN** 사용자가 `moai update`를 실행하면 **THEN** 시스템은 SPEC-UPDATE-001에서 정의된 self-update 워크플로우를 위임 실행해야 한다.

- Flag: `--check` (버전만 확인)
- Flag: `--force` (강제 업데이트)
- Flag: `--templates-only` (템플릿만 동기화)
- Flag: `--yes` (자동 수락)

#### REQ-CLI-016: Hook 커맨드

**WHEN** Claude Code가 `moai hook <event>`를 실행하면 **THEN** 시스템은 SPEC-HOOK-001에서 정의된 Hook 디스패처를 위임 실행해야 한다.

- 서브커맨드: `session-start`, `pre-tool`, `post-tool`, `session-end`, `stop`, `compact`, `list`

#### REQ-CLI-017: CC 커맨드 (top-level)

**WHEN** 사용자가 `moai cc`를 실행하면 **THEN** 시스템은 Claude 백엔드로 전환하고 config를 업데이트해야 한다.

> 참고: `cc`는 별도의 top-level 커맨드이며, `switch` 서브커맨드가 아니다.

#### REQ-CLI-018: GLM 커맨드 (top-level)

**WHEN** 사용자가 `moai glm [api-key]`를 실행하면 **THEN** 시스템은 GLM 백엔드로 전환하고, 선택적으로 API 키를 저장해야 한다.

> 참고: `glm`은 별도의 top-level 커맨드이며, `switch` 서브커맨드가 아니다.

#### REQ-CLI-019: Rank 커맨드

**WHEN** 사용자가 `moai rank <subcommand>`를 실행하면 **THEN** 시스템은 SPEC-TOOL-001에서 정의된 랭킹 기능을 위임 실행해야 한다.

- 서브커맨드: `login`, `status`, `logout`, `sync`, `exclude`, `include`, `register`

#### REQ-CLI-020: Worktree 커맨드

**WHEN** 사용자가 `moai worktree <subcommand>`를 실행하면 **THEN** 시스템은 SPEC-GIT-001에서 정의된 Git worktree 관리 기능을 위임 실행해야 한다.

- Alias: `wt`
- 서브커맨드: `new`, `list`, `switch`, `sync`, `remove`, `clean`

### 3.3 Integration Test Engine (통합 테스트 엔진)

#### REQ-CLI-030: 통합 테스트 엔진 초기화

시스템은 **항상** `internal/core/integration/engine.go`에서 크로스 패키지 검증을 수행하는 통합 테스트 엔진을 제공해야 한다.

#### REQ-CLI-031: 통합 테스트 모델

시스템은 **항상** `internal/core/integration/models.go`에서 테스트 결과 모델을 정의해야 한다.

- `TestSuite`: 테스트 스위트 메타데이터
- `TestResult`: 개별 테스트 결과 (pass/fail/skip, 소요 시간, 에러 메시지)
- `IntegrationReport`: 전체 통합 보고서 (스위트 목록, 총 통계, 실행 시간)

#### REQ-CLI-032: 크로스 패키지 검증

**WHEN** 통합 테스트 엔진이 실행되면 **THEN** 시스템은 다음 크로스 패키지 시나리오를 검증해야 한다:

1. Config 로드 -> CLI 커맨드 실행 파이프라인
2. Hook 등록 -> Hook 디스패치 -> 결과 반환 파이프라인
3. Template 배포 -> Manifest 추적 파이프라인
4. Version 체크 -> Update 다운로드 -> Merge 파이프라인

### 3.4 Version Migration System (버전 마이그레이션)

#### REQ-CLI-040: 버전 마이그레이터

**WHEN** ADK 버전이 업그레이드되면 **THEN** 시스템은 `internal/core/migration/migrator.go`를 통해 config 형식을 자동 마이그레이션해야 한다.

- 마이그레이션 단계: 버전별 migration 함수 체인
- 실패 시: 자동 롤백으로 원본 config 복원
- 로깅: 모든 마이그레이션 단계를 slog로 기록

#### REQ-CLI-041: 백업 매니저

**WHEN** 마이그레이션이 시작되면 **THEN** 시스템은 `internal/core/migration/backup.go`를 통해 현재 config를 백업해야 한다.

- 백업 위치: `.moai/backup/{timestamp}/`
- 백업 대상: `.moai/config/sections/` 전체
- 복원: 마이그레이션 실패 시 자동 복원

#### REQ-CLI-042: 하위 호환성

시스템은 **항상** Python MoAI-ADK의 `.moai/config/sections/` YAML 형식과 호환되는 마이그레이션 경로를 제공해야 한다.

### 3.5 Public Packages (pkg/)

#### REQ-CLI-050: version 패키지

시스템은 **항상** `pkg/version/version.go`에서 빌드 타임 메타데이터 (Version, Commit, Date)를 제공해야 한다.

- `GetVersion()`, `GetCommit()`, `GetDate()`, `GetFullVersion()` 함수 export
- 빌드 타임 주입: `-ldflags "-X .../pkg/version.Version=..."` 패턴

#### REQ-CLI-051: models 패키지

시스템은 **항상** `pkg/models/`에서 공유 데이터 구조를 제공해야 한다.

- `config.go`: UserConfig, LanguageConfig, QualityConfig
- `project.go`: ProjectConfig, ProjectType enum

#### REQ-CLI-052: utils 패키지

시스템은 **항상** `pkg/utils/`에서 공유 유틸리티를 제공해야 한다.

- `logger.go`: slog 기반 구조화 로거 (MOAI_LOG_LEVEL, MOAI_LOG_FORMAT 환경변수)
- `path.go`: 프로젝트 루트 탐색 (FindProjectRoot), .moai 경로 해석 (GetMoAIConfigPath)

### 3.6 Unwanted Behavior (금지 동작)

#### REQ-CLI-090: 순환 의존성 금지

시스템은 패키지 간 순환 의존성을 가져서는 **안 된다**. 의존성은 항상 하향(cli/ -> core/ -> pkg/) 방향으로만 흘러야 한다.

#### REQ-CLI-091: 직접 구체 타입 참조 금지

CLI 커맨드 핸들러는 도메인 모듈의 구체 타입을 직접 참조해서는 **안 된다** (Composition Root 제외). 항상 인터페이스를 통해 접근해야 한다.

#### REQ-CLI-092: 전역 상태 금지

시스템은 전역 변수를 의존성 전달 수단으로 사용해서는 **안 된다** (Cobra 커맨드 등록용 `init()` 제외).

#### REQ-CLI-093: pkg에서 internal 참조 금지

`pkg/` 패키지는 `internal/` 패키지를 import해서는 **안 된다**. Go 컴파일러가 이를 강제한다.

---

## 4. Specifications (명세)

### 4.1 파일 구조

```
internal/cli/
  root.go              (30 LOC, 완료)      -- Root 커맨드, 버전 템플릿
  version.go           (27 LOC, 완료)      -- 버전 정보 출력
  init.go              (~300 LOC)          -- 프로젝트 초기화 (SPEC-INIT-001 위임)
  doctor.go            (~150 LOC)          -- 시스템 진단
  status.go            (~80 LOC)           -- 프로젝트 상태 표시
  update.go            (~800 LOC)          -- Self-update (SPEC-UPDATE-001 위임)
  hook.go              (~200 LOC)          -- Hook 디스패처 (SPEC-HOOK-001 위임)
  cc.go                (~60 LOC)           -- Claude 백엔드 전환 (top-level)
  glm.go               (~60 LOC)           -- GLM 백엔드 전환 (top-level)
  rank.go              (~200 LOC)          -- 7 서브커맨드 (SPEC-TOOL-001 위임)
  worktree/
    new.go             (~80 LOC)           -- 새 worktree 생성
    list.go            (~40 LOC)           -- worktree 목록
    switch.go          (~50 LOC)           -- worktree 전환
    sync.go            (~60 LOC)           -- worktree 동기화
    remove.go          (~50 LOC)           -- worktree 제거
    clean.go           (~50 LOC)           -- worktree 정리

internal/core/integration/
  engine.go            (~200 LOC)          -- 통합 테스트 엔진
  models.go            (~100 LOC)          -- 테스트 결과 모델

internal/core/migration/
  migrator.go          (~200 LOC)          -- 버전 마이그레이션 오케스트레이터
  backup.go            (~150 LOC)          -- 백업 생성 및 복원

pkg/
  version/version.go   (31 LOC, 완료)      -- 빌드 타임 버전 상수
  models/config.go     (25 LOC, 완료)      -- 설정 데이터 모델
  models/project.go    (21 LOC, 완료)      -- 프로젝트 데이터 모델
  utils/logger.go      (48 LOC, 완료)      -- 구조화 로거
  utils/path.go        (52 LOC, 완료)      -- 경로 유틸리티
```

### 4.2 DI Wiring 상세 설계

#### Composition Root 패턴

```go
// cmd/moai/main.go 또는 internal/cli/root.go init() 체인

func initDependencies() {
    // Layer 1: Infrastructure (의존성 없음)
    logger := utils.InitLogger()
    cfg, err := config.New()

    // Layer 2: Domain Services (config 의존)
    hookRegistry := hook.NewRegistry()
    hookProtocol := hook.NewProtocol()
    templateDeployer := template.NewDeployer(embeddedFS)
    manifestMgr := manifest.NewManager(cfg)

    // Layer 3: Core Domains (서비스 의존)
    gitRepo := git.NewRepository(projectRoot)
    projectInit := project.NewInitializer(cfg, templateDeployer, manifestMgr)

    // Layer 4: Cross-cutting (모든 레이어 의존)
    migrator := migration.NewMigrator(cfg, manifestMgr)
    integrationEngine := integration.NewEngine(cfg)

    // Layer 5: CLI Commands (의존성 주입)
    // 각 커맨드에 필요한 의존성을 주입
}
```

#### 의존성 그래프

```
cmd/moai/main.go
    |
    v
internal/cli/root.go
    |
    +-- init.go -----> config/, template/, manifest/, ui/, core/project/
    +-- doctor.go ---> config/, core/project/
    +-- status.go ---> config/, core/git/
    +-- update.go ---> update/, merge/, manifest/, template/
    +-- hook.go -----> hook/
    +-- cc.go -------> config/
    +-- glm.go ------> config/
    +-- rank.go -----> rank/
    +-- worktree/ ---> core/git/
    |
    v
pkg/ (version/, models/, utils/)  -- 모든 패키지에서 접근 가능
```

### 4.3 커맨드 등록 명세

| 커맨드 | 유형 | 등록 위치 | 부모 커맨드 |
|--------|------|----------|-----------|
| `moai` | Root | `root.go` | - |
| `moai version` | Leaf | `version.go` | `rootCmd` |
| `moai init` | Leaf | `init.go` | `rootCmd` |
| `moai doctor` | Leaf | `doctor.go` | `rootCmd` |
| `moai status` | Leaf | `status.go` | `rootCmd` |
| `moai update` | Leaf | `update.go` | `rootCmd` |
| `moai hook` | Group | `hook.go` | `rootCmd` |
| `moai hook session-start` | Leaf | `hook.go` | `hookCmd` |
| `moai hook pre-tool` | Leaf | `hook.go` | `hookCmd` |
| `moai hook post-tool` | Leaf | `hook.go` | `hookCmd` |
| `moai hook session-end` | Leaf | `hook.go` | `hookCmd` |
| `moai hook stop` | Leaf | `hook.go` | `hookCmd` |
| `moai hook compact` | Leaf | `hook.go` | `hookCmd` |
| `moai hook list` | Leaf | `hook.go` | `hookCmd` |
| `moai cc` | Leaf | `cc.go` | `rootCmd` |
| `moai glm` | Leaf | `glm.go` | `rootCmd` |
| `moai rank` | Group | `rank.go` | `rootCmd` |
| `moai rank login` | Leaf | `rank.go` | `rankCmd` |
| `moai rank status` | Leaf | `rank.go` | `rankCmd` |
| `moai rank logout` | Leaf | `rank.go` | `rankCmd` |
| `moai rank sync` | Leaf | `rank.go` | `rankCmd` |
| `moai rank exclude` | Leaf | `rank.go` | `rankCmd` |
| `moai rank include` | Leaf | `rank.go` | `rankCmd` |
| `moai rank register` | Leaf | `rank.go` | `rankCmd` |
| `moai worktree` | Group | `worktree/` | `rootCmd` |
| `moai worktree new` | Leaf | `worktree/new.go` | `worktreeCmd` |
| `moai worktree list` | Leaf | `worktree/list.go` | `worktreeCmd` |
| `moai worktree switch` | Leaf | `worktree/switch.go` | `worktreeCmd` |
| `moai worktree sync` | Leaf | `worktree/sync.go` | `worktreeCmd` |
| `moai worktree remove` | Leaf | `worktree/remove.go` | `worktreeCmd` |
| `moai worktree clean` | Leaf | `worktree/clean.go` | `worktreeCmd` |

### 4.4 Integration Test Engine 명세

#### engine.go

```go
type Engine struct {
    cfg    config.Manager
    suites []TestSuite
}

func NewEngine(cfg config.Manager) *Engine
func (e *Engine) RegisterSuite(suite TestSuite)
func (e *Engine) RunAll(ctx context.Context) (*IntegrationReport, error)
func (e *Engine) RunSuite(ctx context.Context, name string) (*TestResult, error)
```

#### models.go

```go
type TestSuite struct {
    Name        string
    Description string
    Tests       []TestCase
}

type TestCase struct {
    Name     string
    TestFunc func(ctx context.Context) error
}

type TestResult struct {
    Suite    string
    Name     string
    Status   TestStatus  // Pass, Fail, Skip
    Duration time.Duration
    Error    string
}

type IntegrationReport struct {
    Suites    []SuiteResult
    TotalPass int
    TotalFail int
    TotalSkip int
    Duration  time.Duration
}
```

### 4.5 Version Migration 명세

#### migrator.go

```go
type Migrator struct {
    cfg     config.Manager
    backup  *BackupManager
    steps   []MigrationStep
}

type MigrationStep struct {
    FromVersion string
    ToVersion   string
    Migrate     func(cfg config.Manager) error
}

func NewMigrator(cfg config.Manager, backup *BackupManager) *Migrator
func (m *Migrator) NeedsMigration(currentVersion, targetVersion string) bool
func (m *Migrator) Migrate(ctx context.Context, targetVersion string) error
```

#### backup.go

```go
type BackupManager struct {
    baseDir string
}

func NewBackupManager(moaiDir string) *BackupManager
func (b *BackupManager) Create(ctx context.Context) (string, error)
func (b *BackupManager) Restore(ctx context.Context, backupPath string) error
func (b *BackupManager) List() ([]BackupInfo, error)
func (b *BackupManager) Cleanup(keepCount int) error
```

---

## 5. Traceability (추적성)

| 요구사항 ID | 관련 파일 | 테스트 시나리오 |
|------------|----------|--------------|
| REQ-CLI-001 | `root.go`, `main.go` | AC-001 DI 구성 검증 |
| REQ-CLI-002 | `root.go` | AC-002 인터페이스 주입 검증 |
| REQ-CLI-010 | `root.go` | AC-010 Root 커맨드 테스트 |
| REQ-CLI-011 | `version.go` | AC-011 Version 출력 테스트 |
| REQ-CLI-012 | `init.go` | AC-012 Init 라우팅 테스트 |
| REQ-CLI-013 | `doctor.go` | AC-013 Doctor 라우팅 테스트 |
| REQ-CLI-014 | `status.go` | AC-014 Status 표시 테스트 |
| REQ-CLI-015 | `update.go` | AC-015 Update 라우팅 테스트 |
| REQ-CLI-016 | `hook.go` | AC-016 Hook 디스패치 테스트 |
| REQ-CLI-017 | `cc.go` | AC-017 CC 전환 테스트 |
| REQ-CLI-018 | `glm.go` | AC-018 GLM 전환 테스트 |
| REQ-CLI-019 | `rank.go` | AC-019 Rank 라우팅 테스트 |
| REQ-CLI-020 | `worktree/*.go` | AC-020 Worktree 라우팅 테스트 |
| REQ-CLI-030 | `engine.go` | AC-030 통합 엔진 테스트 |
| REQ-CLI-031 | `models.go` | AC-031 모델 직렬화 테스트 |
| REQ-CLI-032 | `engine.go` | AC-032 크로스 패키지 검증 |
| REQ-CLI-040 | `migrator.go` | AC-040 마이그레이션 테스트 |
| REQ-CLI-041 | `backup.go` | AC-041 백업/복원 테스트 |
| REQ-CLI-042 | `migrator.go` | AC-042 하위 호환성 테스트 |
| REQ-CLI-050 | `version.go` | AC-050 version 패키지 테스트 |
| REQ-CLI-051 | `models/*.go` | AC-051 models 직렬화 테스트 |
| REQ-CLI-052 | `utils/*.go` | AC-052 utils 기능 테스트 |
| REQ-CLI-090 | 전체 | AC-090 go vet 순환 참조 검사 |
| REQ-CLI-091 | `cli/*.go` | AC-091 코드 리뷰로 검증 |
| REQ-CLI-092 | 전체 | AC-092 전역 상태 검사 |
| REQ-CLI-093 | `pkg/` | AC-093 Go 컴파일러 강제 |

---

## 6. Expert Consultation (전문가 자문 권장)

### Backend Expert (expert-backend)

- CLI 커맨드와 도메인 서비스 간 DI 패턴 검증
- Go 인터페이스 설계 리뷰 (implicit satisfaction 활용)
- context.Context 전파 패턴 검토

### DevOps Expert (expert-devops)

- goreleaser 크로스 컴파일 설정 검증
- CI/CD 파이프라인에서의 통합 테스트 실행 전략

### Testing Expert (expert-testing)

- 통합 테스트 엔진 설계 리뷰
- Cobra 커맨드 테스팅 패턴 (stdout/stderr 캡처)
- Mock 생성 전략 (mockery + 인터페이스)

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 92.0% (cli) / 100.0% (worktree)

### Summary

CLI integration layer implemented using Cobra command framework with dependency injection pattern. Root command wires all subcommands (init, doctor, status, update, hook, cc, glm, rank, version, statusline, worktree) with shared dependencies injected via a Deps struct. Each subcommand delegates to its corresponding internal package through interfaces, enabling comprehensive unit testing with mocks. Worktree subcommands (new, list, switch, remove, clean, sync) provide git worktree management. Integration tests verify end-to-end command routing and output formatting.

### Files Created

- `internal/cli/cc.go`
- `internal/cli/cc_test.go`
- `internal/cli/coverage_test.go`
- `internal/cli/deps.go`
- `internal/cli/deps_test.go`
- `internal/cli/doctor.go`
- `internal/cli/doctor_test.go`
- `internal/cli/glm.go`
- `internal/cli/glm_test.go`
- `internal/cli/hook.go`
- `internal/cli/hook_test.go`
- `internal/cli/init.go`
- `internal/cli/init_test.go`
- `internal/cli/integration_test.go`
- `internal/cli/mock_test.go`
- `internal/cli/rank.go`
- `internal/cli/rank_test.go`
- `internal/cli/root.go`
- `internal/cli/root_test.go`
- `internal/cli/status.go`
- `internal/cli/status_test.go`
- `internal/cli/statusline.go`
- `internal/cli/statusline_test.go`
- `internal/cli/update.go`
- `internal/cli/update_test.go`
- `internal/cli/version.go`
- `internal/cli/version_test.go`
- `internal/cli/worktree/clean.go`
- `internal/cli/worktree/list.go`
- `internal/cli/worktree/new.go`
- `internal/cli/worktree/remove.go`
- `internal/cli/worktree/root.go`
- `internal/cli/worktree/root_test.go`
- `internal/cli/worktree/subcommands_test.go`
- `internal/cli/worktree/switch.go`
- `internal/cli/worktree/sync.go`
