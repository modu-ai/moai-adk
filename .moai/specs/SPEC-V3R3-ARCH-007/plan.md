---
spec_id: SPEC-V3R3-ARCH-007
title: Implementation Plan — Token Circuit Breaker
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-ARCH-007/spec.md
---

# Plan — SPEC-V3R3-ARCH-007

## 1. Objectives

- `.moai/config/sections/runtime.yaml` 신설 (per-agent budget + circuit breaker config)
- Template 동기화 (`internal/template/templates/.moai/config/sections/runtime.yaml`)
- `internal/runtime/budget.go` Tracker 모듈 구현
- SessionStart 훅에서 runtime.yaml 로드
- 75% 도달 시 progress.md 자동 저장 + resume message 출력
- BC-V3R3-006 warning-first 정책 명시

## 2. Technical Approach

### 2.1 runtime.yaml schema

```yaml
# .moai/config/sections/runtime.yaml
runtime:
  context_window:
    pre_clear_threshold: 0.75
    hard_clear_threshold: 0.90

  per_agent_budget:
    default: 30000
    manager-strategy: 60000
    manager-spec: 40000
    expert-backend: 40000
    expert-frontend: 40000
    expert-security: 40000
    expert-testing: 40000
    expert-debug: 40000
    expert-performance: 40000
    expert-refactoring: 40000
    expert-devops: 40000
    expert-mobile: 40000
    evaluator-active: 20000
    plan-auditor: 20000

  circuit_breaker:
    stall_detection_seconds: 60
    retry_max: 3
    fallback: split_into_waves

  progress_persistence:
    auto_save_at_threshold: true
    save_path_template: ".moai/specs/{SPEC_ID}/progress.md"
    resume_message_format: "ultrathink. {wave_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}."
```

### 2.2 internal/runtime/budget.go

```go
// Package runtime provides the Token Circuit Breaker for MoAI-ADK.
//
// SPEC-V3R3-ARCH-007: per-agent token budget tracking + stall detection.
// Warning-first policy (BC-V3R3-006); P5에서 hard-fail로 전환 예정.
package runtime

import (
    "fmt"
    "sync"
    "time"
)

type Tracker struct {
    mu              sync.RWMutex
    config          RuntimeConfig
    perAgentUsage   map[string]int
    perAgentLastTs  map[string]time.Time
    perAgentRetries map[string]int
}

// RuntimeConfig represents the parsed runtime.yaml schema.
type RuntimeConfig struct {
    PreClearThreshold      float64
    HardClearThreshold     float64
    PerAgentBudget         map[string]int
    StallDetectionSeconds  int
    RetryMax               int
    Fallback               string
    AutoSaveAtThreshold    bool
    SavePathTemplate       string
    ResumeMessageFormat    string
}

// NewTracker creates a Tracker with the given config (or defaults if nil).
func NewTracker(cfg *RuntimeConfig) *Tracker { /* ... */ }

// RecordCall records token usage for an agent invocation.
func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int) { /* ... */ }

// Usage returns current usage, budget, and ratio for an agent.
func (t *Tracker) Usage(agentName string) (current int, budget int, ratio float64) { /* ... */ }

// IsApproachingLimit returns true if usage >= 75% of budget.
func (t *Tracker) IsApproachingLimit(agentName string) bool { /* ... */ }

// IsAtHardLimit returns true if usage >= 90% of budget.
func (t *Tracker) IsAtHardLimit(agentName string) bool { /* ... */ }

// DetectStall returns true if no RecordCall received within stall_detection_seconds.
func (t *Tracker) DetectStall(agentName string) bool { /* ... */ }

// PersistProgress writes progress.md and returns the resume message.
func (t *Tracker) PersistProgress(specID, waveLabel, approach, nextStep string) (string, error) { /* ... */ }
```

### 2.3 SessionStart 훅 통합

`internal/cli/hook_session_start.go` (또는 동등 경로)에서:

```go
// Load runtime.yaml at session start
cfg, err := config.LoadRuntime(".moai/config/sections/runtime.yaml")
if err != nil {
    // Use defaults per REQ-ARCH007-011
    cfg = runtime.DefaultRuntimeConfig()
}

// Initialize tracker; expose via context for downstream agents
tracker := runtime.NewTracker(cfg)
sessionState.SetTracker(tracker)
```

### 2.4 75% trigger flow

```
[Tracker.RecordCall]
    → [Tracker.IsApproachingLimit] returns true
        → [Tracker.PersistProgress]
            → write .moai/specs/<SPEC>/progress.md (atomic)
            → emit resume message to stderr (or hook output)
        → log WARN level (not interrupting agent execution)
```

## 3. Wave / Phase 설계

### Wave A.1 — runtime.yaml schema 작성 (2 tasks)

- T-A1-1: `.moai/config/sections/runtime.yaml` 작성
- T-A1-2: `internal/template/templates/.moai/config/sections/runtime.yaml` 동기화

### Wave A.2 — Go runtime 모듈 구현 (4 tasks)

- T-A2-1: `internal/runtime/budget.go` Tracker type + 5 메서드 골격
- T-A2-2: `internal/runtime/config.go` runtime.yaml 파싱
- T-A2-3: `internal/runtime/persist.go` progress.md 저장 + resume message 생성
- T-A2-4: `internal/runtime/budget_test.go` unit tests (각 메서드 별 1 + Edge case 3)

### Wave A.3 — SessionStart 훅 통합 (1 task)

- T-A3-1: SessionStart 훅 호출 지점에 runtime.yaml 로드 + Tracker 초기화 추가

### Wave A.4 — Verification (3 tasks)

- T-A4-1: AC verification 스크립트 실행
- T-A4-2: `make build && make install` + Claude Code 재시작 안내
- T-A4-3: CHANGELOG에 BC-V3R3-006 entry 추가

## 4. File 영향 요약

| File | Change Type |
|------|-------------|
| `.moai/config/sections/runtime.yaml` | NEW |
| `internal/template/templates/.moai/config/sections/runtime.yaml` | NEW |
| `internal/runtime/budget.go` | NEW |
| `internal/runtime/config.go` | NEW |
| `internal/runtime/persist.go` | NEW |
| `internal/runtime/budget_test.go` | NEW |
| `internal/cli/hook_session_start.go` (또는 hook 진입점) | EDIT (SessionStart 통합) |
| `internal/template/embedded.go` | AUTO-REGEN (`make build`) |
| `CHANGELOG.md` | EDIT (BC-V3R3-006 entry) |

총 신규 6 + 수정 2 + 자동 생성 1.

## 5. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| token estimation heuristic이 실제와 큰 차이 | Medium | ±10% 허용, 향후 calibration SPEC에서 정밀화 |
| progress.md 동시 쓰기 race | Medium | atomic write (`os.Rename` 패턴) |
| SessionStart 훅 호출 진입점 미식별 | High | grep으로 hook 등록 위치 사전 확인 (T-A3-1 전) |
| Go runtime 메모리 leak (per-agent map 무한 증가) | Low | session 단위 lifecycle, session 종료 시 cleanup |
| runtime.yaml 누락 시 default fallback 동작 실패 | Medium | DefaultRuntimeConfig() 함수 + REQ-ARCH007-011 |
| Binary 미빌드로 변경 미반영 | High | DoD에 `make build && make install` 명시; `MEMORY.md` Hard Constraint |

## 6. Open Questions

- OQ1: per-agent budget의 단위는 input / output / total tokens? → **Decision**: total (input + output) tokens. Heuristic estimation 사용.
- OQ2: stall detection 시점 — RecordCall 마지막 timestamp 기반? → **Decision**: Yes. RecordCall 호출 간 간격이 60s 초과 시 stall로 판정.
- OQ3: progress.md 자동 저장 시 기존 내용 보존? → **Decision**: 보존 + append section "## Auto-saved at <timestamp> (75% threshold)" 추가
- OQ4: hook integration — SessionStart 또는 PreToolUse? → **Decision**: SessionStart 1회 + PreToolUse에서 RecordCall 트리거 (별도 SPEC에서 추가 가능, 본 SPEC은 인프라만)
- OQ5: SessionStart 훅 진입점 위치? → **Decision**: T-A3-1 시작 전 `grep -r "session-start" internal/cli/` 로 식별

## 7. Milestones

- M1: runtime.yaml schema 신설 (local + template)
- M2: Go Tracker 모듈 구현 + unit tests 통과
- M3: SessionStart 통합 완료
- M4: `make build && make install` 후 dogfood 실행 (75% trigger 동작 확인)
- M5: CHANGELOG BC-V3R3-006 entry 추가

## 8. Definition of Done

- [ ] `.moai/config/sections/runtime.yaml` + template pair 모두 존재 + schema 일치
- [ ] `internal/runtime/budget.go`의 Tracker type 5 메서드 모두 구현
- [ ] `internal/runtime/budget_test.go` 모든 unit test 통과 (`go test -count=1 -race ./internal/runtime/...`)
- [ ] SessionStart 훅에서 Tracker 초기화 동작 확인
- [ ] 75% 도달 시 progress.md 저장 + resume message 출력 (manual smoke test)
- [ ] `/clear` 자동 트리거 부재 검증 (코드 grep으로 `os.Exec("/clear")` 등 부재)
- [ ] `make build && make install` 성공
- [ ] `go test ./...` 회귀 통과
- [ ] CHANGELOG에 BC-V3R3-006 entry 추가
