# SPEC-UI-002: Statusline Rendering - 수락 기준

---
spec_id: SPEC-UI-002
document: acceptance
created: 2026-02-03
tags: statusline, ui, git, metrics, theme, claude-code
---

## 1. 수락 기준 개요

이 문서는 SPEC-UI-002 (Statusline Rendering) 구현의 완료를 검증하기 위한 상세 수락 기준을 정의한다. 모든 시나리오는 Given-When-Then 형식으로 작성되었다.

---

## 2. Feature: Statusline 기본 렌더링

### Scenario 1: 정상적인 statusline 출력 생성

```gherkin
Given Claude Code가 statusline 명령을 호출한다
  And stdin으로 유효한 JSON이 전달된다:
    """json
    {
      "hook_event_name": "statusLine",
      "session_id": "test-session",
      "cwd": "/path/to/project",
      "model": "claude-sonnet-4",
      "cost": { "total_usd": 0.05, "input_tokens": 1000, "output_tokens": 500 },
      "context_window": { "used": 50000, "total": 200000 }
    }
    """
When Builder.Build(ctx)가 실행된다
Then stdout에 단일 줄 텍스트가 출력된다
  And 출력에 개행 문자(\n)가 포함되지 않는다
  And 출력이 UTF-8 인코딩이다
  And stderr에는 아무것도 출력되지 않는다
```

**관련 요구사항:** REQ-U-001, REQ-U-004, REQ-N-001, REQ-N-002

### Scenario 2: 응답 시간 300ms 이내 보장

```gherkin
Given 유효한 stdin JSON이 제공된다
When Builder.Build(ctx)가 1000회 반복 실행된다
Then P95 응답 시간이 100ms 이하이다
  And P99 응답 시간이 300ms 이하이다
  And 모든 실행이 타임아웃 없이 완료된다
```

**관련 요구사항:** REQ-U-002

### Scenario 3: stdin JSON 파싱 실패 시 안전한 기본 출력

```gherkin
Given stdin으로 잘못된 JSON이 전달된다:
    """
    { invalid json content
    """
When Builder.Build(ctx)가 실행된다
Then stdout에 기본 출력이 생성된다
  And 출력에 최소한 버전 정보 또는 "MoAI" 텍스트가 포함된다
  And panic이 발생하지 않는다
  And exit code가 0이다
```

**관련 요구사항:** REQ-U-003, REQ-N-003

### Scenario 4: 빈 stdin 처리

```gherkin
Given stdin이 비어 있다 (EOF)
When Builder.Build(ctx)가 실행된다
Then stdout에 기본 출력이 생성된다
  And panic이 발생하지 않는다
```

**관련 요구사항:** REQ-U-003, REQ-N-003

---

## 3. Feature: Git 상태 표시

### Scenario 5: Git 저장소에서 브랜치 및 변경사항 표시

```gherkin
Given Git 저장소가 초기화되어 있다
  And 현재 브랜치가 "feature/auth"이다
  And 3개의 modified 파일이 있다
  And 2개의 staged 파일이 있다
  And 1개의 untracked 파일이 있다
  And remote에 대해 1 ahead, 0 behind이다
When GitCollector.CollectGit(ctx)가 실행된다
Then GitStatusData.Branch가 "feature/auth"이다
  And GitStatusData.Modified가 3이다
  And GitStatusData.Staged가 2이다
  And GitStatusData.Untracked가 1이다
  And GitStatusData.Ahead가 1이다
  And GitStatusData.Behind이 0이다
```

**관련 요구사항:** REQ-E-002

### Scenario 6: Default 모드에서 Git 정보 렌더링

```gherkin
Given StatuslineMode가 "default"이다
  And GitStatusData가 { Branch: "main", Modified: 3, Staged: 2 }이다
When Renderer.Render(data, ModeDefault)가 실행된다
Then 출력에 "main" 브랜치명이 포함된다
  And 출력에 수정 파일 수가 포함된다
```

**관련 요구사항:** REQ-E-002, REQ-S-005

### Scenario 7: Git 저장소가 없는 환경

```gherkin
Given 현재 디렉토리에 Git 저장소가 없다
When GitCollector.CollectGit(ctx)가 실행된다
Then 에러가 반환되지 않는다
  And GitStatusData가 빈 기본값으로 반환된다
When Builder.Build(ctx)가 실행된다
Then statusline이 Git 섹션 없이 정상 출력된다
  And 나머지 섹션(context, cost)은 정상 표시된다
```

**관련 요구사항:** REQ-E-002, REQ-N-003

---

## 4. Feature: 컨텍스트 윈도우 표시

### Scenario 8: 컨텍스트 사용률 계산 및 표시

```gherkin
Given context_window.used가 50000이다
  And context_window.total이 200000이다
When MemoryCollector.CollectMemory(input)가 실행된다
Then MemoryData.TokensUsed가 50000이다
  And MemoryData.TokenBudget이 200000이다
When Renderer가 사용률을 렌더링한다
Then "25%" 텍스트가 출력에 포함된다
```

**관련 요구사항:** REQ-E-003

### Scenario 9: 컨텍스트 사용률에 따른 색상 코딩 - 녹색

```gherkin
Given context_window 사용률이 25% (50K/200K)이다
  And MOAI_NO_COLOR가 설정되지 않았다
When Renderer.Render(data, mode)가 실행된다
Then 컨텍스트 비율 섹션에 녹색 ANSI 코드(\033[32m)가 적용된다
```

**관련 요구사항:** REQ-S-001

### Scenario 10: 컨텍스트 사용률에 따른 색상 코딩 - 황색

```gherkin
Given context_window 사용률이 65% (130K/200K)이다
  And MOAI_NO_COLOR가 설정되지 않았다
When Renderer.Render(data, mode)가 실행된다
Then 컨텍스트 비율 섹션에 황색 ANSI 코드(\033[33m)가 적용된다
```

**관련 요구사항:** REQ-S-002

### Scenario 11: 컨텍스트 사용률에 따른 색상 코딩 - 적색

```gherkin
Given context_window 사용률이 90% (180K/200K)이다
  And MOAI_NO_COLOR가 설정되지 않았다
When Renderer.Render(data, mode)가 실행된다
Then 컨텍스트 비율 섹션에 적색 ANSI 코드(\033[31m)가 적용된다
```

**관련 요구사항:** REQ-S-003

### Scenario 12: context_window 필드 누락

```gherkin
Given stdin JSON에 context_window 필드가 없다:
    """json
    {
      "hook_event_name": "statusLine",
      "model": "claude-sonnet-4",
      "cost": { "total_usd": 0.05 }
    }
    """
When Builder.Build(ctx)가 실행된다
Then statusline이 정상 출력된다
  And 컨텍스트 섹션이 "N/A" 또는 빈 상태로 표시된다
  And panic이 발생하지 않는다
```

**관련 요구사항:** REQ-U-003, REQ-N-003

---

## 5. Feature: 업데이트 알림

### Scenario 13: 새 버전 가용 시 알림 표시

```gherkin
Given 현재 ADK 버전이 "1.2.0"이다
  And 최신 버전이 "1.3.0"으로 확인되었다
When UpdateChecker.CheckUpdate(ctx)가 실행된다
Then VersionData.Current가 "1.2.0"이다
  And VersionData.Latest가 "1.3.0"이다
  And VersionData.UpdateAvailable이 true이다
When Verbose 모드에서 Renderer가 실행된다
Then 출력에 업데이트 가능 알림이 포함된다
```

**관련 요구사항:** REQ-E-004

### Scenario 14: 최신 버전 사용 중일 때

```gherkin
Given 현재 ADK 버전이 "1.3.0"이다
  And 최신 버전이 "1.3.0"이다
When UpdateChecker.CheckUpdate(ctx)가 실행된다
Then VersionData.UpdateAvailable이 false이다
When Renderer가 실행된다
Then 출력에 업데이트 알림이 포함되지 않는다
```

**관련 요구사항:** REQ-E-004

### Scenario 15: 업데이트 확인 캐싱

```gherkin
Given UpdateChecker의 cacheTTL이 1시간으로 설정되어 있다
  And 10분 전에 업데이트 확인이 수행되었다
When UpdateChecker.CheckUpdate(ctx)가 재호출된다
Then 네트워크 요청 없이 캐시된 결과가 반환된다
  And 응답 시간이 1ms 이하이다
```

**관련 요구사항:** REQ-N-005

### Scenario 16: 업데이트 확인 네트워크 오류

```gherkin
Given 네트워크가 사용 불가능하다
  And 캐시된 버전 정보가 없다
When UpdateChecker.CheckUpdate(ctx)가 실행된다
Then 에러가 반환되지만 panic은 발생하지 않는다
When Builder.Build(ctx)가 실행된다
Then statusline이 버전 섹션 없이 정상 출력된다
```

**관련 요구사항:** REQ-N-003, REQ-N-005

---

## 6. Feature: 테마 전환

### Scenario 17: 기본 테마 적용

```gherkin
Given statusline 설정의 theme이 "default"이다
When Renderer가 초기화된다
Then 녹색/황색/적색 lipgloss 스타일이 적용된다
  And 구분자가 " | "이다
```

**관련 요구사항:** REQ-E-006

### Scenario 18: 테마 변경 적용

```gherkin
Given 현재 테마가 "default"이다
When 사용자가 설정에서 테마를 "nerd"로 변경한다
  And 다음 statusline 호출이 발생한다
When Builder.Build(ctx)가 실행된다
Then Nerd Font 아이콘이 포함된 출력이 생성된다
  And 이전 "default" 테마의 아이콘은 사용되지 않는다
```

**관련 요구사항:** REQ-E-006

### Scenario 19: MOAI_NO_COLOR 환경변수

```gherkin
Given MOAI_NO_COLOR 환경변수가 "1"로 설정되어 있다
When Renderer.Render(data, mode)가 실행된다
Then 출력에 ANSI 이스케이프 코드(\033[)가 포함되지 않는다
  And 모든 텍스트가 plain text로 출력된다
```

**관련 요구사항:** REQ-S-007

---

## 7. Feature: 디스플레이 모드

### Scenario 20: Minimal 모드 출력

```gherkin
Given StatuslineMode가 "minimal"이다
  And model이 "claude-sonnet-4"이다
  And context_window 사용률이 25%이다
When Builder.Build(ctx)가 실행된다
Then 출력이 모델명과 컨텍스트 비율만 포함한다
  And Git 정보가 포함되지 않는다
  And 비용 정보가 포함되지 않는다
  And 예시: "sonnet-4 | Ctx: 25%"
```

**관련 요구사항:** REQ-S-004

### Scenario 21: Default 모드 출력

```gherkin
Given StatuslineMode가 "default"이다
  And 브랜치가 "main"이고 modified 파일이 2개이다
  And context_window 사용률이 25%이다
  And 비용이 $0.05이다
When Builder.Build(ctx)가 실행된다
Then 출력에 Git 상태가 포함된다
  And 출력에 컨텍스트 비율이 포함된다
  And 출력에 비용이 포함된다
  And 예시: "main ~2 | Ctx: 25% | $0.05"
```

**관련 요구사항:** REQ-S-005

### Scenario 22: Verbose 모드 출력

```gherkin
Given StatuslineMode가 "verbose"이다
  And 모든 데이터 소스가 사용 가능하다
When Builder.Build(ctx)가 실행된다
Then 출력에 Git 상태(브랜치, modified, staged, ahead, behind)가 포함된다
  And 출력에 컨텍스트 사용량(절대값 + 비율)이 포함된다
  And 출력에 비용이 포함된다
  And 출력에 버전 정보가 포함된다
  And 업데이트 가능 시 알림이 포함된다
  And 예시: "main +3 ~2 ^1 v0 | Ctx: 50K/200K (25%) | $0.05 | v1.2.0 (update!)"
```

**관련 요구사항:** REQ-S-006

---

## 8. Feature: 에러 복원력

### Scenario 23: 개별 Collector 타임아웃

```gherkin
Given GitCollector가 2초 이상 응답하지 않는다
  And 다른 Collector들은 정상이다
When Builder.Build(ctx)가 실행된다
Then Git 섹션이 빈 상태로 표시된다
  And 나머지 섹션(context, cost, version)은 정상 출력된다
  And 전체 응답 시간이 1초 이내이다
  And slog에 debug 레벨 로그가 기록된다
```

**관련 요구사항:** REQ-U-002, REQ-U-005, REQ-N-003

### Scenario 24: 모든 Collector 동시 실패

```gherkin
Given 모든 데이터 Collector가 에러를 반환한다
When Builder.Build(ctx)가 실행된다
Then 최소 기본 출력이 생성된다 (예: "MoAI" 또는 빈 줄)
  And panic이 발생하지 않는다
  And exit code가 0이다
```

**관련 요구사항:** REQ-U-003, REQ-N-003

### Scenario 25: Context 취소 처리

```gherkin
Given context.WithCancel로 생성된 ctx가 있다
  And Builder.Build(ctx)가 실행 중이다
When ctx가 취소된다
Then Build가 즉시 반환된다
  And 부분 데이터 또는 기본 출력이 반환된다
  And goroutine 누수가 없다
```

**관련 요구사항:** REQ-U-005

---

## 9. Feature: 비용 표시

### Scenario 26: 세션 비용 표시

```gherkin
Given stdin JSON의 cost.total_usd가 0.15이다
When MetricsCollector.CollectMetrics(input)가 실행된다
Then 비용 데이터가 정상 파싱된다
When Default 모드에서 Renderer가 실행된다
Then 출력에 "$0.15" 형식의 비용이 포함된다
```

**관련 요구사항:** REQ-E-005

### Scenario 27: cost 필드 누락

```gherkin
Given stdin JSON에 cost 필드가 없다
When Builder.Build(ctx)가 실행된다
Then statusline이 비용 섹션 없이 정상 출력된다
  And panic이 발생하지 않는다
```

**관련 요구사항:** REQ-U-003

---

## 10. Feature: 병렬 실행 안전성

### Scenario 28: 경쟁 조건 없음

```gherkin
Given 4개의 Collector가 errgroup.Group으로 병렬 실행된다
When go test -race 플래그로 테스트가 실행된다
Then race condition이 감지되지 않는다
  And 모든 테스트가 통과한다
```

**관련 요구사항:** REQ-U-005

### Scenario 29: Goroutine 누수 없음

```gherkin
Given Builder.Build(ctx)가 1000회 반복 호출된다
When 모든 호출이 완료된 후 runtime.NumGoroutine()을 확인한다
Then 활성 goroutine 수가 호출 전과 유사하다 (허용 차이: 2개 이내)
  And goroutine 누수가 없다
```

**관련 요구사항:** REQ-U-005, REQ-N-003

---

## 11. Quality Gate 기준

### 11.1 TRUST 5 체크리스트

| 원칙 | 기준 | 검증 방법 |
|------|------|----------|
| **Tested** | 85%+ 코드 커버리지 | `go test -coverprofile` |
| **Tested** | 모든 에러 경로에 테스트 존재 | 리뷰 확인 |
| **Tested** | 벤치마크 P95 < 100ms | `go test -bench` |
| **Readable** | 모든 exported 타입/함수에 godoc 주석 | `golangci-lint` |
| **Readable** | 함수 길이 80줄 이하 | `golangci-lint` (funlen) |
| **Unified** | gofumpt 포맷팅 통과 | `golangci-lint` |
| **Unified** | golangci-lint 경고 0개 | CI 파이프라인 |
| **Secured** | 파일 시스템 쓰기 없음 | 코드 리뷰 |
| **Secured** | stdin 외 외부 입력 없음 | 코드 리뷰 |
| **Trackable** | 모든 커밋에 SPEC-UI-002 참조 | Git log 확인 |

### 11.2 Definition of Done

- [ ] 6개 파일(builder.go, git.go, metrics.go, memory.go, renderer.go, update.go) 구현 완료
- [ ] 모든 Scenario (1~29) 통과
- [ ] `go test -race ./internal/statusline/...` 통과
- [ ] `go test -coverprofile` 결과 85% 이상
- [ ] `golangci-lint run ./internal/statusline/...` 경고 0개
- [ ] 벤치마크 P95 응답 시간 100ms 이하
- [ ] MOAI_NO_COLOR 모드 정상 동작 확인
- [ ] 3개 테마(default, minimal, nerd) 렌더링 확인
- [ ] 3개 모드(minimal, default, verbose) 출력 검증
- [ ] godoc 주석 완료 (모든 exported 심볼)
- [ ] SPEC-GIT-001 인터페이스 Mock으로 테스트 통과 (실제 연동은 해당 SPEC 완료 후)
