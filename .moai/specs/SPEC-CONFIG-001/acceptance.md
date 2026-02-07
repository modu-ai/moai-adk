---
id: SPEC-CONFIG-001
version: "1.0.0"
status: Planned
created: "2026-02-03"
updated: "2026-02-03"
author: GOOS
priority: P0-Critical
---

# SPEC-CONFIG-001: Acceptance Criteria

## 1. Test Scenario Overview

### 1.1 테스트 범위

| 카테고리 | 시나리오 수 | 커버리지 목표 |
|----------|------------|--------------|
| Configuration Loading | 6 | 90%+ |
| Thread-Safe Access | 4 | 90%+ |
| File Watching & Events | 3 | 85%+ |
| Validation & Error Handling | 5 | 90%+ |
| State Machine | 3 | 85%+ |
| Legacy Migration | 3 | 85%+ |
| Performance | 4 | Benchmark |
| Edge Cases | 5 | 85%+ |
| Integration | 3 | 85%+ |
| Methodology Validation | 6 | 90%+ |
| **합계** | **42** | **85%+ 전체** |

---

## 2. Core Test Scenarios

### Scenario 1: 정상 설정 파일 로드

YAML section 파일이 올바르게 존재할 때, typed Config struct로 정확히 로드되는지 검증한다.

```gherkin
Given .moai/config/sections/ 디렉토리에 유효한 YAML section 파일들이 존재하고
  And user.yaml에 "name: GOOS"가 포함되어 있고
  And language.yaml에 "conversation_language: ko"가 포함되어 있고
  And quality.yaml에 "test_coverage_target: 85"가 포함되어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then error는 nil이어야 하고
  And 반환된 Config.User.Name은 "GOOS"이어야 하고
  And 반환된 Config.Language.ConversationLanguage는 "ko"이어야 하고
  And 반환된 Config.Quality.TestCoverageTarget은 85이어야 한다
```

### Scenario 2: 누락된 Section 파일에 Default 값 적용

일부 section 파일이 존재하지 않을 때, compiled default 값이 올바르게 적용되는지 검증한다.

```gherkin
Given .moai/config/sections/ 디렉토리에 user.yaml만 존재하고
  And language.yaml, quality.yaml 등 나머지 section 파일이 존재하지 않을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then error는 nil이어야 하고
  And Config.Language.ConversationLanguage는 default 값 "en"이어야 하고
  And Config.Language.AgentPromptLanguage는 default 값 "en"이어야 하고
  And Config.Quality.TestCoverageTarget은 default 값 85이어야 하고
  And Config.Quality.EnforceQuality는 default 값 true이어야 하고
  And Config.LLM.DefaultModel은 default 값 "sonnet"이어야 하고
  And Config.Workflow.PlanTokens는 default 값 30000이어야 한다
```

### Scenario 3: 환경 변수 오버라이드

환경 변수가 YAML 파일 값보다 높은 우선순위로 적용되는지 검증한다.

```gherkin
Given .moai/config/sections/system.yaml에 "log_level: info"가 설정되어 있고
  And 환경 변수 MOAI_SYSTEM_LOG_LEVEL="debug"가 설정되어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then Config.System.LogLevel은 "debug"이어야 한다 (환경 변수 우선)
```

### Scenario 4: Thread-Safe Concurrent Read

다수의 goroutine이 동시에 `Get()`을 호출할 때 data race가 발생하지 않는지 검증한다.

```gherkin
Given ConfigManager가 Load()를 통해 initialized 상태이고
When 100개의 goroutine이 동시에 Get()을 호출하면
Then Go race detector (-race flag)에서 data race가 감지되지 않아야 하고
  And 모든 goroutine이 동일한 Config 값을 반환해야 한다
```

### Scenario 5: Thread-Safe Concurrent Read/Write

동시 read와 write 작업이 충돌 없이 처리되는지 검증한다.

```gherkin
Given ConfigManager가 initialized 상태이고
When 50개의 goroutine이 Get()을 호출하면서 동시에
  And 10개의 goroutine이 SetSection("language", newLangConfig)을 호출하면
Then Go race detector에서 data race가 감지되지 않아야 하고
  And 모든 SetSection 호출이 순서대로 처리되어야 하고
  And Get() 호출은 항상 일관된 Config snapshot을 반환해야 한다
```

### Scenario 6: Invalid YAML 파일 Graceful Handling

YAML 구문 오류가 있는 section 파일을 만났을 때 전체 로딩이 실패하지 않고 default로 대체하는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 유효하지 않은 YAML 구문이 포함되어 있고
  "quality:\n  enforce_quality: [invalid yaml"
  And 나머지 section 파일들은 정상적일 때
When ConfigManager.Load(projectRoot)가 호출되면
Then error는 nil이어야 하고 (전체 로딩 실패가 아닌 graceful degradation)
  And Config.Quality는 NewDefaultQualityConfig() 값이어야 하고
  And slog warning 로그에 "quality.yaml" 파일명과 parsing error가 기록되어야 한다
```

### Scenario 7: 필수 필드 누락 Validation Error

`validate:"required"` tag가 지정된 필수 필드가 누락되었을 때 actionable error를 반환하는지 검증한다.

```gherkin
Given .moai/config/sections/user.yaml에 "user:" 만 존재하고 name 필드가 비어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then error가 반환되어야 하고
  And error message에 "user.name" 필드명이 포함되어야 하고
  And error message에 "required" 키워드가 포함되어야 하고
  And error message에 예시 값 또는 수정 방법 안내가 포함되어야 한다
```

### Scenario 8: Dynamic Token 감지 및 거부

설정 값에 unexpanded dynamic token이 포함되어 있을 때 거부하는지 검증한다.

```gherkin
Given .moai/config/sections/system.yaml에 다음 내용이 포함되어 있을 때
  "system:\n  version: ${MOAI_VERSION}"
When ConfigManager.Load(projectRoot)가 호출되면
Then error가 반환되어야 하고
  And error는 ErrDynamicToken 타입이어야 하고
  And error message에 "system.version" 필드명이 포함되어야 하고
  And error message에 "${MOAI_VERSION}" 토큰이 식별되어야 한다
```

### Scenario 9: Atomic Save 동작

`Save()` 호출 시 temp file + rename 패턴으로 atomic하게 저장되는지 검증한다.

```gherkin
Given ConfigManager가 initialized 상태이고
  And Config.Language.ConversationLanguage를 "ja"로 SetSection()으로 변경했을 때
When ConfigManager.Save()가 호출되면
Then .moai/config/sections/language.yaml 파일이 업데이트되어야 하고
  And 파일 내용에 "conversation_language: ja"가 포함되어야 하고
  And temp 파일(.moai-config-*.tmp)이 남아있지 않아야 하고
  And 새 ConfigManager로 같은 경로를 Load()하면 동일한 값이 반환되어야 한다
```

### Scenario 10: File Watcher Callback 호출

설정 파일이 외부에서 변경되었을 때 registered callback이 호출되는지 검증한다.

```gherkin
Given ConfigManager가 initialized 상태이고
  And Watch(callback)으로 callback 함수가 등록되어 있을 때
When 외부 프로세스가 .moai/config/sections/language.yaml 파일을 직접 수정하면
Then 1초 이내에 callback 함수가 호출되어야 하고
  And callback에 전달된 Config에 변경된 값이 반영되어야 한다
```

### Scenario 11: Uninitialized State 보호

`Load()`를 호출하기 전에 다른 메서드를 호출하면 적절한 error가 반환되는지 검증한다.

```gherkin
Given ConfigManager가 New()로 생성되었지만 Load()가 호출되지 않은 상태일 때
When Get()을 호출하면
Then nil이 반환되어야 하고

When SetSection("user", newValue)를 호출하면
Then ErrNotInitialized error가 반환되어야 하고

When Save()를 호출하면
Then ErrNotInitialized error가 반환되어야 하고

When Watch(callback)를 호출하면
Then ErrNotInitialized error가 반환되어야 하고

When Reload()를 호출하면
Then ErrNotInitialized error가 반환되어야 한다
```

### Scenario 12: GetSection 정상 반환 및 오류 처리

존재하는 section과 존재하지 않는 section에 대한 GetSection 동작을 검증한다.

```gherkin
Given ConfigManager가 initialized 상태일 때
When GetSection("user")를 호출하면
Then UserConfig 타입의 값이 반환되어야 하고
  And error는 nil이어야 하고

When GetSection("nonexistent")를 호출하면
Then nil이 반환되어야 하고
  And error는 ErrSectionNotFound이어야 한다
```

---

## 3. Legacy Migration Scenarios

### Scenario 13: JSON에서 YAML Sections로 Migration

Legacy JSON 설정 파일이 있을 때 YAML sections 형식으로 정상 변환되는지 검증한다.

```gherkin
Given 프로젝트 루트에 config.json 파일이 존재하고
  And config.json에 {"user": {"name": "GOOS"}, "language": {"conversation_language": "ko"}} 내용이 있고
  And .moai/config/sections/ 디렉토리가 존재하지 않을 때
When DetectLegacy(projectRoot)가 호출되면
Then FormatJSON이 반환되어야 하고

When Migrate(projectRoot, FormatJSON)가 호출되면
Then .moai/config/sections/user.yaml이 생성되어야 하고
  And .moai/config/sections/language.yaml이 생성되어야 하고
  And user.yaml에 "name: GOOS"가 포함되어야 하고
  And language.yaml에 "conversation_language: ko"가 포함되어야 하고
  And config.json.bak 백업 파일이 생성되어야 한다
```

### Scenario 14: Migration 없는 프로젝트

Legacy 파일이 없는 프로젝트에서 migration 감지가 정상 동작하는지 검증한다.

```gherkin
Given 프로젝트 루트에 config.json이 존재하지 않고
  And .moai/config/sections/ 디렉토리에 YAML 파일들이 정상 존재할 때
When DetectLegacy(projectRoot)가 호출되면
Then FormatNone이 반환되어야 한다
```

### Scenario 15: Migration 중 원본 보존

Migration 실패 시 원본 파일이 보존되는지 검증한다.

```gherkin
Given 프로젝트 루트에 config.json이 존재하고
  And 디스크에 쓰기 권한이 제한된 상태일 때
When Migrate(projectRoot, FormatJSON)가 호출되고 파일 쓰기에 실패하면
Then error가 반환되어야 하고
  And 원본 config.json 파일이 변경되지 않아야 하고
  And 부분적으로 생성된 section 파일이 cleanup 되어야 한다
```

---

## 4. Edge Case Scenarios

### Scenario 16: 빈 디렉토리에서 로드

`.moai/config/sections/` 디렉토리가 비어 있을 때 전체 default 값으로 초기화되는지 검증한다.

```gherkin
Given .moai/config/sections/ 디렉토리가 존재하지만 YAML 파일이 하나도 없을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then error는 nil이어야 하고
  And 반환된 Config는 NewDefaultConfig()와 동일해야 한다
```

### Scenario 17: 디렉토리 자체가 없는 경우

`.moai/config/sections/` 디렉토리가 존재하지 않을 때의 동작을 검증한다.

```gherkin
Given 프로젝트 루트에 .moai/config/sections/ 디렉토리가 존재하지 않을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then error는 nil이어야 하고
  And 반환된 Config는 NewDefaultConfig()와 동일해야 한다
```

### Scenario 18: 매우 큰 설정 파일

비정상적으로 큰 section 파일을 처리할 수 있는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml 파일 크기가 1MB를 초과할 때
When ConfigManager.Load(projectRoot)가 호출되면
Then 로딩이 10ms 이내에 완료되거나
  Or 적절한 error message와 함께 실패해야 한다
```

### Scenario 19: YAML key 대소문자 호환성

Python MoAI-ADK의 YAML key 명칭과 정확히 일치하는지 검증한다.

```gherkin
Given Python MoAI-ADK에서 생성된 .moai/config/sections/quality.yaml 파일이 있고
  And 파일에 "constitution:\n  development_mode: ddd\n  enforce_quality: true" 형식이 사용되었을 때
When Go ConfigManager.Load(projectRoot)가 호출되면
Then error는 nil이어야 하고
  And 모든 필드가 올바르게 매핑되어야 한다
```

### Scenario 20: Reload 후 Watch Callback 유지

Reload() 후에도 기존 Watch callback이 유효한지 검증한다.

```gherkin
Given ConfigManager가 initialized 상태이고
  And Watch(callback)으로 callback이 등록되어 있고
When Reload()가 호출된 후
  And 외부에서 설정 파일이 변경되면
Then callback이 여전히 호출되어야 한다
```

---

## 5. Performance Criteria

### 5.1 Benchmark 기준

| 메트릭 | 목표 | 테스트 방법 |
|--------|------|-------------|
| `Load()` cold start | < 10ms | `BenchmarkLoad` (9개 section 파일 로드) |
| `Load()` cached | < 1ms | `BenchmarkLoadCached` (Viper 캐시 활용) |
| `Get()` read | < 1us (1 microsecond) | `BenchmarkGet` (RLock 오버헤드만) |
| `GetSection()` | < 1us | `BenchmarkGetSection` (RLock + reflect 조회) |
| `Save()` atomic write | < 5ms | `BenchmarkSave` (temp file + rename) |
| `Reload()` full | < 10ms | `BenchmarkReload` (write lock + 전체 re-read) |
| Memory footprint | < 1MB | `TestMemoryUsage` (runtime.ReadMemStats) |

### 5.2 Benchmark Test 템플릿

```go
func BenchmarkLoad(b *testing.B) {
    // Setup: create test config directory with all section files
    dir := setupTestConfigDir(b)

    mgr := config.New()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := mgr.Load(dir)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGet(b *testing.B) {
    mgr := setupInitializedManager(b)
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _ = mgr.Get()
        }
    })
}
```

### Scenario 21: Load 성능 검증

```gherkin
Given .moai/config/sections/에 9개의 YAML section 파일이 존재할 때
When BenchmarkLoad를 실행하면
Then 평균 실행 시간이 10ms 미만이어야 한다
```

### Scenario 22: Get 성능 검증

```gherkin
Given ConfigManager가 initialized 상태일 때
When BenchmarkGet을 100개 goroutine으로 병렬 실행하면
Then 평균 실행 시간이 1 microsecond 미만이어야 한다
```

### Scenario 23: Save 성능 검증

```gherkin
Given ConfigManager에 변경된 설정이 있을 때
When BenchmarkSave를 실행하면
Then 평균 실행 시간이 5ms 미만이어야 한다
```

### Scenario 24: Memory 사용량 검증

```gherkin
Given ConfigManager가 9개 section 전체를 로드한 상태일 때
When runtime.ReadMemStats()로 메모리 사용량을 측정하면
Then Config 관련 heap 할당이 1MB 미만이어야 한다
```

---

## 6. Quality Gate Criteria

### 6.1 TRUST 5 Validation

| 원칙 | 기준 | 검증 방법 |
|------|------|-----------|
| **Tested** | >= 85% line coverage (전체), >= 90% (`manager.go`, `validation.go`) | `go test -coverprofile` |
| **Tested** | `go test -race` zero data races | CI 파이프라인 필수 |
| **Readable** | 모든 exported 함수에 godoc 주석 | `golangci-lint` (exported check) |
| **Readable** | 함수 길이 50줄 이내, cyclomatic complexity 10 이하 | `gocyclo`, `funlen` linter |
| **Unified** | `gofumpt` 포매팅 준수 | `golangci-lint` (gofumpt) |
| **Unified** | 모든 error는 `fmt.Errorf("context: %w", err)` 래핑 | Code review, linter |
| **Secured** | `gosec` zero findings | `golangci-lint` (gosec) |
| **Secured** | filepath.Clean() 적용 확인 | Code review |
| **Trackable** | Conventional commit 메시지 | Git hook validation |

### 6.2 Python 호환성 검증

```gherkin
Given 현행 Python MoAI-ADK 프로젝트의 .moai/config/sections/ 디렉토리를 test fixture로 복사했을 때
When Go ConfigManager.Load()로 로드하면
Then 모든 section이 올바르게 파싱되어야 하고
  And Save()로 다시 저장한 후
  And 다시 Load()하면 원래 값과 동일해야 한다 (round-trip 검증)
```

### 6.3 Regression Test

| 테스트 | 검증 대상 | Python Issue |
|--------|-----------|--------------|
| `TestLoadRecovery` | 로딩 실패 시 graceful degradation | #315 |
| `TestConcurrentAccess` | Race condition 없음 (`-race` flag) | #283 |
| `TestTypeSafety` | Typed struct deserialization | #206 |
| `TestDefaultValues` | 누락 필드에 default 적용 | #245 |
| `TestSectionConsistency` | Section 간 값 일관성 | #243 |
| `TestNoDynamicTokens` | Dynamic token 감지 및 거부 | #304 |

---

## 7. Integration Test Scenarios

### Scenario 25: CLI -> Config 통합

```gherkin
Given moai binary가 빌드되어 있고
  And 테스트 프로젝트에 .moai/config/sections/ 설정이 존재할 때
When "moai status" CLI 커맨드를 실행하면
Then exit code 0으로 완료되어야 하고
  And 출력에 설정 값이 올바르게 표시되어야 한다
```

### Scenario 26: Config -> Hook 통합

```gherkin
Given ConfigManager가 initialized 상태이고
  And Hook handler가 config.Get()을 호출할 때
When Hook이 실행되면
Then ConfigManager.Get()이 nil이 아닌 유효한 Config를 반환해야 하고
  And Hook handler가 설정 값을 기반으로 정상 동작해야 한다
```

### Scenario 27: Config -> Template 통합

```gherkin
Given ConfigManager가 initialized 상태이고
  And Template renderer가 Config 값을 참조할 때
When Template rendering이 실행되면
Then Config의 Language, System 설정이 올바르게 template에 반영되어야 하고
  And 렌더링된 출력에 unexpanded token이 존재하지 않아야 한다
```

---

## 8. Methodology Validation Scenarios

### Scenario 28: DDD 모드 설정 및 동작 검증

`development_mode`가 `ddd`일 때 DDD 관련 설정이 활성화되는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 다음 내용이 설정되어 있을 때
  "development_mode: ddd"
  And ddd_settings에 "characterization_tests: true"가 설정되어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then Config.Quality.DevelopmentMode는 "ddd"이어야 하고
  And Config.Quality.DDDSettings.CharacterizationTests는 true이어야 하고
  And Config.Quality.DDDSettings.BehaviorSnapshots는 true이어야 한다
```

### Scenario 29: TDD 모드 설정 및 동작 검증

`development_mode`가 `tdd`일 때 TDD 관련 설정이 활성화되는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 다음 내용이 설정되어 있을 때
  "development_mode: tdd"
  And tdd_settings에 "test_first_required: true"가 설정되어 있고
  And tdd_settings에 "min_coverage_per_commit: 80"이 설정되어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then Config.Quality.DevelopmentMode는 "tdd"이어야 하고
  And Config.Quality.TDDSettings.TestFirstRequired는 true이어야 하고
  And Config.Quality.TDDSettings.RedGreenRefactor는 true이어야 하고
  And Config.Quality.TDDSettings.MinCoveragePerCommit은 80이어야 한다
```

### Scenario 30: Hybrid 모드 설정 및 동작 검증

`development_mode`가 `hybrid`일 때 Hybrid 관련 설정이 활성화되는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 다음 내용이 설정되어 있을 때
  "development_mode: hybrid"
  And hybrid_settings에 "new_features: tdd"가 설정되어 있고
  And hybrid_settings에 "legacy_refactoring: ddd"가 설정되어 있고
  And hybrid_settings에 "min_coverage_new: 90"이 설정되어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then Config.Quality.DevelopmentMode는 "hybrid"이어야 하고
  And Config.Quality.HybridSettings.NewFeatures는 "tdd"이어야 하고
  And Config.Quality.HybridSettings.LegacyRefactoring는 "ddd"이어야 하고
  And Config.Quality.HybridSettings.MinCoverageNew는 90이어야 한다
```

### Scenario 31: 잘못된 Development Mode 검증 에러

유효하지 않은 `development_mode` 값이 설정되었을 때 validation error가 반환되는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 다음 내용이 설정되어 있을 때
  "development_mode: waterfall"
When ConfigManager.Load(projectRoot)가 호출되면
Then error가 반환되어야 하고
  And error는 ErrInvalidDevelopmentMode 타입이어야 하고
  And error message에 "ddd", "tdd", "hybrid" 유효 옵션이 포함되어야 한다
```

### Scenario 32: 환경 변수로 Development Mode 오버라이드

`MOAI_DEVELOPMENT_MODE` 환경 변수가 YAML 설정보다 우선 적용되는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 "development_mode: ddd"가 설정되어 있고
  And 환경 변수 MOAI_DEVELOPMENT_MODE="tdd"가 설정되어 있을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then Config.Quality.DevelopmentMode는 "tdd"이어야 한다 (환경 변수 우선)
```

### Scenario 33: 모드별 설정 독립성 검증

활성화된 모드의 설정만 적용되고 다른 모드의 설정은 default 값을 유지하는지 검증한다.

```gherkin
Given .moai/config/sections/quality.yaml에 다음 내용이 설정되어 있을 때
  "development_mode: tdd"
  And tdd_settings에 커스텀 값이 설정되어 있고
  And ddd_settings와 hybrid_settings는 설정되지 않았을 때
When ConfigManager.Load(projectRoot)가 호출되면
Then Config.Quality.DevelopmentMode는 "tdd"이어야 하고
  And Config.Quality.TDDSettings에 커스텀 값이 적용되어야 하고
  And Config.Quality.DDDSettings는 default 값이어야 하고
  And Config.Quality.HybridSettings는 default 값이어야 한다
```

---

## 9. Test File Structure

```
internal/config/
  manager.go
  manager_test.go          # Scenario 1-6, 9-12, 20-24
  types.go
  types_test.go            # Struct tag 검증, YAML round-trip
  defaults.go
  defaults_test.go         # Default 값 완전성 검증
  migration.go
  migration_test.go        # Scenario 13-15
  validation.go
  validation_test.go       # Scenario 7-8, dynamic token 감지
  testdata/
    valid/                 # 정상 YAML fixture files
      user.yaml
      language.yaml
      quality.yaml
      ...
    invalid/               # 비정상 YAML fixture files
      broken_syntax.yaml
      missing_required.yaml
      dynamic_tokens.yaml
    legacy/                # Legacy JSON fixture files
      config.json
      unified_config.json
    methodology/           # 개발 방법론 모드별 fixture files
      quality_ddd.yaml
      quality_tdd.yaml
      quality_hybrid.yaml
      quality_invalid.yaml
    python_compat/         # Python MoAI-ADK 호환성 fixture
      sections/            # 실제 Python 프로젝트에서 복사한 파일
```

---

## 10. Definition of Done

SPEC-CONFIG-001은 다음 조건이 모두 충족될 때 완료된 것으로 간주한다:

- [ ] 5개 파일 모두 구현 완료 (`types.go`, `defaults.go`, `validation.go`, `manager.go`, `migration.go`)
- [ ] `ConfigManager` interface의 모든 메서드 구현 완료
- [ ] 전체 test coverage >= 85%
- [ ] `go test -race ./internal/config/...` 통과 (zero data races)
- [ ] `golangci-lint run ./internal/config/...` zero errors
- [ ] Benchmark: `Load()` < 10ms, `Get()` < 1us, `Save()` < 5ms
- [ ] Python MoAI-ADK YAML section 파일 round-trip test 통과
- [ ] Dynamic token 감지 test 통과 (`${VAR}`, `{{VAR}}` 패턴)
- [ ] `development_mode` validation test 통과 (`ddd`, `tdd`, `hybrid` 허용 값)
- [ ] `MOAI_DEVELOPMENT_MODE` 환경 변수 오버라이드 test 통과
- [ ] TDDSettings, HybridSettings, CoverageExemptions default 값 test 통과
- [ ] 모든 exported 함수에 godoc 주석 존재
- [ ] Error wrapping: 모든 error에 `fmt.Errorf("context: %w", err)` 패턴 적용
- [ ] 33개 acceptance scenario 중 31개 이상 통과 (Optional scenario 제외)
