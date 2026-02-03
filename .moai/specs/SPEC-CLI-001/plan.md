# SPEC-CLI-001: Implementation Plan

---
spec_id: SPEC-CLI-001
title: CLI Command Composition & Integration - Implementation Plan
phase: "Phase 5 - CLI (Final, composition root)"
status: Planned
priority: Medium
tags: [cli, composition-root, di-wiring, integration, migration, cobra, pkg]
---

## 1. Implementation Strategy (구현 전략)

### 1.1 전략 요약

이 SPEC은 Composition Root로서 모든 선행 SPEC 모듈을 통합하는 최종 단계이다. 따라서 **Bottom-Up 통합 전략**을 채택한다:

1. `pkg/` 공개 패키지를 먼저 완성 (다른 모든 패키지의 기반)
2. `internal/core/migration/` 마이그레이션 시스템 구현 (config 호환성 보장)
3. CLI 커맨드를 하나씩 Stub에서 Full Implementation으로 전환
4. `internal/core/integration/` 통합 테스트 엔진 구현 (최종 검증)

### 1.2 Composition Root 철학

```
"Composition Root는 유일하게 구체 타입을 알 수 있는 장소이다.
 나머지 코드베이스는 오직 인터페이스를 통해서만 의존성에 접근한다."
```

이 원칙에 따라:
- `internal/cli/root.go`의 `init()` 체인이 Composition Root 역할
- 각 CLI 커맨드 파일은 도메인 인터페이스만 참조
- 구체 타입 생성은 Composition Root에서만 수행

---

## 2. Milestones (마일스톤)

### Milestone 1: pkg/ 공개 패키지 완성 (Primary Goal)

현재 `pkg/` 패키지는 기본 골격만 존재한다. 외부 도구가 import할 수 있는 안정적 공개 API로 완성한다.

#### 작업 목록

| 작업 | 파일 | 설명 | 의존성 |
|------|------|------|--------|
| M1-1 | `pkg/version/version.go` | 완료 상태 확인, 변경 불필요 | 없음 |
| M1-2 | `pkg/models/config.go` | WorkflowConfig, PluginConfig 등 누락 타입 추가 | 없음 |
| M1-3 | `pkg/models/project.go` | SPECStatus, SPECPriority enum 추가 | 없음 |
| M1-4 | `pkg/utils/logger.go` | 완료 상태 확인, 변경 불필요 | 없음 |
| M1-5 | `pkg/utils/path.go` | ClaudeConfigPath, SpecsPath 등 경로 헬퍼 추가 | 없음 |

#### 완료 기준

- `go vet ./pkg/...` 에러 0건
- `go test ./pkg/...` 커버리지 95% 이상
- 모든 공개 함수에 godoc 주석 작성

---

### Milestone 2: Version Migration 시스템 (Primary Goal)

`.moai/config/` 디렉토리의 YAML 설정 파일을 ADK 버전 간 안전하게 마이그레이션하는 시스템을 구현한다.

#### 작업 목록

| 작업 | 파일 | 설명 | 의존성 |
|------|------|------|--------|
| M2-1 | `internal/core/migration/backup.go` | BackupManager 구현 (백업 생성, 복원, 목록, 정리) | `pkg/utils/path.go` |
| M2-2 | `internal/core/migration/migrator.go` | Migrator 구현 (마이그레이션 스텝 체인, 실행, 롤백) | `backup.go`, `internal/config/` |
| M2-3 | `internal/core/migration/migrator_test.go` | 단위 테스트 (table-driven) | M2-1, M2-2 |
| M2-4 | `internal/core/migration/backup_test.go` | 백업/복원 단위 테스트 | M2-1 |

#### 기술적 접근

```go
// Migration Step 패턴
type MigrationStep struct {
    FromVersion string
    ToVersion   string
    Description string
    Migrate     func(cfg config.Manager) error
}

// 실행 흐름
// 1. 현재 버전 감지
// 2. 타겟 버전까지 필요한 스텝 필터링
// 3. 백업 생성
// 4. 순차 실행 (스텝별 에러 처리)
// 5. 실패 시 백업에서 복원
```

#### 완료 기준

- 마이그레이션 실패 시 자동 롤백 동작 검증
- Python ADK YAML 형식에서 Go ADK YAML 형식으로 마이그레이션 테스트 통과
- 백업 디렉토리 `.moai/backup/{timestamp}/` 정상 생성 확인

---

### Milestone 3: CLI 커맨드 구현 -- 기존 Stub 전환 (Primary Goal)

현재 Stub 상태인 `init.go`, `doctor.go`, `status.go`를 완전한 구현으로 전환한다.

#### 작업 목록

| 작업 | 파일 | 설명 | 의존성 |
|------|------|------|--------|
| M3-1 | `internal/cli/init.go` | SPEC-INIT-001 위임 로직 구현, Cobra flag 정의 | SPEC-INIT-001 모듈 |
| M3-2 | `internal/cli/doctor.go` | 진단 로직 구현, flag 정의 (--verbose, --fix, --export) | `internal/config/` |
| M3-3 | `internal/cli/status.go` | 상태 표시 로직 구현 | `internal/config/`, `internal/core/git/` |

#### 기술적 접근

**init.go 패턴:**
```go
var initCmd = &cobra.Command{
    Use:   "init [path]",
    Short: "Initialize a new MoAI project",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
    // 1. Flag 파싱
    nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
    // 2. 의존성에서 Initializer 가져오기
    // 3. Initializer.Run(ctx, opts) 호출
    return nil
}
```

#### 완료 기준

- 각 커맨드의 `--help` 출력이 Python 에디션과 동등한 정보 제공
- 각 커맨드에 대한 스모크 테스트 통과

---

### Milestone 4: CLI 커맨드 구현 -- 신규 커맨드 (Secondary Goal)

아직 존재하지 않는 커맨드 파일들을 생성하고 구현한다.

#### 작업 목록

| 작업 | 파일 | 설명 | 의존성 |
|------|------|------|--------|
| M4-1 | `internal/cli/update.go` | SPEC-UPDATE-001 위임 로직 + flag 정의 | SPEC-UPDATE-001 모듈 |
| M4-2 | `internal/cli/hook.go` | SPEC-HOOK-001 위임 로직 + 7개 서브커맨드 | SPEC-HOOK-001 모듈 |
| M4-3 | `internal/cli/cc.go` | Claude 백엔드 전환 (top-level) | `internal/config/` |
| M4-4 | `internal/cli/glm.go` | GLM 백엔드 전환 (top-level) | `internal/config/` |
| M4-5 | `internal/cli/rank.go` | 7개 서브커맨드 + SPEC-TOOL-001 위임 | SPEC-TOOL-001 모듈 |
| M4-6 | `internal/cli/worktree/*.go` | 6개 서브커맨드 + SPEC-GIT-001 위임 | SPEC-GIT-001 모듈 |

#### 기술적 접근

**cc.go / glm.go 패턴 (top-level 커맨드):**
```go
// internal/cli/cc.go
var ccCmd = &cobra.Command{
    Use:   "cc",
    Short: "Switch to Claude backend",
    RunE:  runCC,
}

func init() {
    rootCmd.AddCommand(ccCmd)  // top-level에 직접 등록
}

// internal/cli/glm.go
var glmCmd = &cobra.Command{
    Use:   "glm [api-key]",
    Short: "Switch to GLM backend",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runGLM,
}

func init() {
    rootCmd.AddCommand(glmCmd)  // top-level에 직접 등록
}
```

**worktree 서브커맨드 패턴:**
```go
// internal/cli/worktree/root.go
var WorktreeCmd = &cobra.Command{
    Use:     "worktree",
    Aliases: []string{"wt"},
    Short:   "Git worktree management",
}

func init() {
    WorktreeCmd.AddCommand(newCmd, listCmd, switchCmd, syncCmd, removeCmd, cleanCmd)
}
```

#### 완료 기준

- 모든 커맨드의 `moai <command> --help` 정상 출력
- `cc`, `glm`이 각각 독립 top-level 커맨드로 동작 확인
- worktree 서브커맨드가 `moai worktree <sub>` 및 `moai wt <sub>` 양쪽으로 접근 가능

---

### Milestone 5: DI Wiring 완성 (Secondary Goal)

모든 커맨드와 도메인 모듈을 Composition Root에서 연결한다.

#### 작업 목록

| 작업 | 파일 | 설명 | 의존성 |
|------|------|------|--------|
| M5-1 | `internal/cli/root.go` | DI Container 또는 init() 체인으로 의존성 구성 | 모든 도메인 모듈 |
| M5-2 | `internal/cli/deps.go` (신규) | 의존성 구조체 정의 (선택적) | M5-1 |
| M5-3 | 전체 CLI | 각 커맨드에 의존성 전달 (global var 또는 closure) | M5-1, M5-2 |

#### 기술적 접근

**옵션 A: Global Dependencies struct**
```go
// internal/cli/deps.go
type Dependencies struct {
    Config     config.Manager
    Git        git.Repository
    Hook       hook.Registry
    Template   template.Deployer
    Manifest   manifest.Manager
    Migrator   migration.Migrator
    Rank       rank.Client
}

var deps *Dependencies

func InitDependencies() error {
    deps = &Dependencies{
        Config:   config.New(),
        // ...
    }
    return nil
}
```

**옵션 B: Cobra PersistentPreRunE 기반**
```go
rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    // 공통 의존성 초기화
    // cmd.Context()에 저장
    return nil
}
```

#### 완료 기준

- `moai init` 실행 시 config, template, manifest 모듈이 정상 주입 확인
- `moai hook session-start` 실행 시 hook 모듈이 정상 주입 확인
- 순환 의존성 없음 (`go vet ./...` 통과)

---

### Milestone 6: Integration Test Engine (Final Goal)

크로스 패키지 검증을 위한 통합 테스트 엔진을 구현한다.

#### 작업 목록

| 작업 | 파일 | 설명 | 의존성 |
|------|------|------|--------|
| M6-1 | `internal/core/integration/models.go` | TestSuite, TestResult, IntegrationReport 모델 | `pkg/models/` |
| M6-2 | `internal/core/integration/engine.go` | Engine 구현 (RegisterSuite, RunAll, RunSuite) | M6-1 |
| M6-3 | `internal/core/integration/engine_test.go` | 엔진 자체 테스트 | M6-1, M6-2 |
| M6-4 | 통합 테스트 스위트 등록 | 크로스 패키지 시나리오 정의 | 모든 도메인 모듈 |

#### 검증 시나리오

| 시나리오 | 검증 내용 |
|---------|----------|
| Config -> CLI | Config 로드 후 CLI 커맨드 정상 실행 |
| Hook -> Dispatch | Hook 등록 -> 이벤트 디스패치 -> 결과 반환 |
| Template -> Manifest | 템플릿 배포 -> 매니페스트 추적 일관성 |
| Version -> Update | 버전 체크 -> 다운로드 -> Merge 파이프라인 |
| Migration -> Config | 구 버전 config -> 마이그레이션 -> 신 버전 config |

#### 완료 기준

- 통합 테스트 스위트 5개 이상 등록
- `go test -tags=integration ./internal/core/integration/...` 전체 통과
- 크로스 패키지 파이프라인 검증 커버리지 확보

---

## 3. Technical Approach (기술적 접근)

### 3.1 Cobra 커맨드 구조

```
moai (root)
  +-- version         (Leaf, 완료)
  +-- init            (Leaf, Stub -> Full)
  +-- doctor          (Leaf, Stub -> Full)
  +-- status          (Leaf, Stub -> Full)
  +-- update          (Leaf, 신규)
  +-- hook            (Group, 신규)
  |   +-- session-start
  |   +-- pre-tool
  |   +-- post-tool
  |   +-- session-end
  |   +-- stop
  |   +-- compact
  |   +-- list
  +-- cc              (Leaf, top-level, 신규)
  +-- glm             (Leaf, top-level, 신규)
  +-- rank            (Group, 신규)
  |   +-- login
  |   +-- status
  |   +-- logout
  |   +-- sync
  |   +-- exclude
  |   +-- include
  |   +-- register
  +-- worktree        (Group, alias: wt, 신규)
      +-- new
      +-- list
      +-- switch
      +-- sync
      +-- remove
      +-- clean
```

### 3.2 테스트 전략

| 테스트 유형 | 위치 | 목적 |
|-----------|------|------|
| Unit Test | `*_test.go` (same package) | 개별 함수/메서드 검증 |
| Smoke Test | `cli/*_test.go` | Cobra 커맨드 실행 + stdout 캡처 |
| Integration Test | `integration/*_test.go` | 크로스 패키지 파이프라인 검증 |
| Benchmark | `*_bench_test.go` | CLI 커맨드 응답 시간 P95 < 200ms 확인 |

**Cobra 커맨드 테스트 패턴:**
```go
func TestVersionCommand(t *testing.T) {
    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetErr(buf)
    rootCmd.SetArgs([]string{"version"})

    err := rootCmd.Execute()
    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "moai-adk")
}
```

### 3.3 에러 처리 전략

모든 CLI 커맨드는 `RunE` (error-returning) 패턴을 사용한다:

```go
var exampleCmd = &cobra.Command{
    Use:  "example",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 에러 발생 시 Cobra가 자동으로 사용법과 함께 출력
        return fmt.Errorf("context: %w", err)
    },
}
```

### 3.4 Context 전파 패턴

모든 장기 실행 작업은 `context.Context`를 전파하여 타임아웃과 취소를 지원한다:

```go
func runUpdate(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
    defer cancel()
    return deps.Update.Execute(ctx, opts)
}
```

---

## 4. Architecture Design (아키텍처 설계 방향)

### 4.1 레이어드 아키텍처

```
Layer 1: Entry Point     -- cmd/moai/main.go
Layer 2: CLI Commands     -- internal/cli/ (Cobra commands, flag parsing)
Layer 3: Domain Services  -- internal/core/, internal/hook/, internal/update/
Layer 4: Infrastructure   -- internal/config/, internal/template/, internal/manifest/
Layer 5: Public API       -- pkg/ (version, models, utils)
```

**의존성 규칙**: 상위 레이어만 하위 레이어를 참조할 수 있다. 역방향 참조는 금지.

### 4.2 패키지 의존성 방향

```
cmd/moai/   -->  internal/cli/   (유일한 의존)
internal/cli/ -->  internal/*/   (도메인 서비스)
internal/*/   -->  pkg/          (공개 모델, 유틸리티)
pkg/          -->  stdlib only   (외부 의존 없음)
```

### 4.3 ADR 참조

| ADR | 적용 영역 |
|-----|----------|
| ADR-001 (Modular Monolithic) | 전체 CLI 구조 |
| ADR-003 (Interface-Based DDD) | DI wiring, 커맨드-서비스 경계 |
| ADR-004 (go:embed) | init 커맨드의 템플릿 배포 |
| ADR-005 (log/slog) | 모든 CLI 커맨드 로깅 |
| ADR-006 (Hooks as Subcommands) | hook 커맨드 구조 |

---

## 5. Risks and Mitigations (위험 및 대응)

| 위험 | 영향 | 발생 확률 | 대응 방안 |
|------|------|----------|----------|
| 선행 SPEC 모듈 인터페이스 불일치 | High | Medium | 인터페이스 계약 사전 정의, 컴파일 타임 검증 |
| Cobra init() 순서 의존성 | Medium | Low | init() 대신 명시적 AddCommand() 체인 사용 |
| 통합 테스트 환경 차이 | Medium | Medium | CI에서 동일 환경 보장, Docker 기반 테스트 |
| 마이그레이션 데이터 손실 | High | Low | 백업 필수, 롤백 자동화, dry-run 모드 제공 |
| worktree go-git 한계 | Medium | High | 시스템 Git fallback (ADR-007) 적극 활용 |
| DI 복잡도 증가 | Low | Medium | 의존성 수 최소화, Composition Root 단일 지점 유지 |

---

## 6. Out of Scope (범위 외)

- 개별 도메인 모듈의 내부 구현 (각 선행 SPEC에서 담당)
- LSP 통합 (`internal/lsp/`)
- Ralph 피드백 루프 (`internal/loop/`, `internal/ralph/`)
- AST-Grep 통합 (`internal/astgrep/`)
- Statusline 렌더링 (`internal/statusline/`)
- Charmbracelet TUI 컴포넌트 (`internal/ui/`)
- goreleaser 크로스 컴파일 설정

---

## 7. Definition of Done (완료 정의)

- [ ] 모든 CLI 커맨드가 `moai <command> --help`로 도움말 출력
- [ ] DI Wiring이 완성되어 모든 커맨드가 도메인 서비스에 접근 가능
- [ ] `cc`, `glm`이 top-level 커맨드로 정상 등록 (`switch` 아님)
- [ ] 버전 마이그레이션 시스템이 백업 + 롤백 동작 검증
- [ ] 통합 테스트 엔진이 5개 이상 크로스 패키지 시나리오 검증
- [ ] `go vet ./...` 에러 0건
- [ ] `go test -race ./...` 전체 통과
- [ ] CLI 레이어 테스트 커버리지 70% 이상
- [ ] `pkg/` 패키지 테스트 커버리지 95% 이상
- [ ] `internal/core/migration/` 테스트 커버리지 85% 이상
- [ ] 순환 의존성 0건
