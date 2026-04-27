---
id: SPEC-SRS-002
version: "1.0.0"
status: completed
created: "2026-04-09"
updated: "2026-04-09"
author: GOOS
priority: high
issue_number: 0
---

# SPEC-SRS-002: Experiment Loop + Passive Observation

## Overview

Self-Research System의 실험 루프 상태 머신과 수동 관찰 수집 시스템을 구현한다.
SPEC-SRS-001에서 구현한 eval/safety 패키지 위에 구축.

Source: `.moai/plans/DESIGN-SRS-002-unified.md` Phase 2 + Phase 3

## What NOT to Build (Exclusions)

- Dashboard (`internal/research/dashboard/`) - Phase 5
- Agency 통합 (Learner 리팩토링) - Phase 4
- Researcher 에이전트/스킬 정의 (.claude/agents/moai/researcher.md) - Phase 5
- `moai research` CLI 서브커맨드의 전체 구현 - 기반 인프라만 (CLI wiring은 Phase 5)
- Hook handler의 실제 품질 평가 로직 - 관찰 기록 인프라만

---

## Requirements

### REQ-1: Experiment 타입 정의 (`internal/research/experiment/types.go`)

**[EARS: Ubiquitous]**

The system SHALL define types for experiment lifecycle management.

```go
type ExperimentState string
const (
    StateIdle      ExperimentState = "idle"
    StateBaseline  ExperimentState = "baseline"
    StateMutating  ExperimentState = "mutating"
    StateEvaluating ExperimentState = "evaluating"
    StateScoring   ExperimentState = "scoring"
    StateComplete  ExperimentState = "complete"
)

type Experiment struct {
    ID         string
    Target     string
    Hypothesis string
    Change     ChangeRecord
    Results    []eval.EvalResult
    Decision   Decision  // keep | discard
    Timestamp  time.Time
}

type Decision string
const (
    DecisionKeep    Decision = "keep"
    DecisionDiscard Decision = "discard"
    DecisionPending Decision = "pending"
)

type ChangeRecord struct {
    Type    string // addition | modification | deletion
    Section string
    Diff    string
}
```

Acceptance:
- 모든 타입이 정의되고 테스트에서 사용 가능
- JSON 직렬화/역직렬화 테스트 통과

### REQ-2: Baseline 관리자 (`internal/research/experiment/baseline.go`)

**[EARS: Event-driven]**

WHEN a baseline measurement is requested, the system SHALL run the eval suite on the unchanged target and store results as JSON.

```go
type BaselineManager struct {
    storeDir string
}

func NewBaselineManager(storeDir string) *BaselineManager
func (m *BaselineManager) Save(target string, result *eval.EvalResult) error
func (m *BaselineManager) Load(target string) (*eval.EvalResult, error)
func (m *BaselineManager) Exists(target string) bool
```

Acceptance:
- Save는 `{storeDir}/{sanitized_target}.baseline.json`에 저장
- Load는 저장된 baseline을 읽어 EvalResult로 반환
- 파일이 없으면 명확한 에러 반환
- t.TempDir()로 격리된 테스트

### REQ-3: Experiment 결과 저장소 (`internal/research/experiment/store.go`)

**[EARS: Event-driven]**

WHEN an experiment completes, the system SHALL persist the result as numbered JSON files and maintain a changelog.

```go
type ResultStore struct {
    baseDir string
}

func NewResultStore(baseDir string) *ResultStore
func (s *ResultStore) SaveExperiment(target string, exp *Experiment) error
func (s *ResultStore) LoadExperiments(target string) ([]*Experiment, error)
func (s *ResultStore) AppendChangelog(target string, entry ChangelogEntry) error
```

Acceptance:
- 실험 결과를 `{baseDir}/{target}/exp-001.json`, `exp-002.json` 순번으로 저장
- 순번은 기존 파일 수 + 1로 자동 증가
- Changelog는 `{baseDir}/{target}/changelog.md`에 append
- 테스트 85%+ 커버리지

### REQ-4: Experiment Loop 상태 머신 (`internal/research/experiment/loop.go`)

**[EARS: State-driven]**

WHILE the experiment loop is running, the system SHALL manage state transitions and enforce termination conditions.

```go
type LoopConfig struct {
    MaxExperiments      int
    TargetScore         float64
    StagnationThreshold float64
    StagnationPatience  int
    BudgetCapTokens     int
}

type Loop struct {
    config    LoopConfig
    state     ExperimentState
    baseline  *eval.EvalResult
    bestScore float64
    history   []*Experiment
    stagnationCount int
}

func NewLoop(config LoopConfig) *Loop
func (l *Loop) SetBaseline(result *eval.EvalResult)
func (l *Loop) ShouldContinue() bool
func (l *Loop) RecordExperiment(exp *Experiment) Decision
func (l *Loop) BestScore() float64
func (l *Loop) ExperimentCount() int
func (l *Loop) State() ExperimentState
```

ShouldContinue 종료 조건:
- MaxExperiments 도달 → false
- TargetScore 3회 연속 달성 → false
- StagnationPatience 초과 (개선 < StagnationThreshold 연속) → false
- 명시적 Complete 상태 → false

RecordExperiment 결정 로직:
- 새 점수 > bestScore → DecisionKeep, bestScore 업데이트
- 새 점수의 MustPassOK == false → DecisionDiscard (must_pass 실패)
- 새 점수 <= bestScore → DecisionDiscard

Acceptance:
- 상태 전이 테스트 (idle→baseline→mutating→evaluating→scoring→complete)
- 종료 조건 4가지 각각 테스트
- keep/discard 결정 로직 테스트
- 정체 감지 테스트
- 테이블 드리븐 테스트

### REQ-5: Observation 타입 (`internal/research/observe/types.go`)

**[EARS: Ubiquitous]**

The system SHALL define types for passive observation collection and pattern detection.

```go
type ObservationType string
const (
    ObsCorrection ObservationType = "correction"
    ObsFailure    ObservationType = "failure"
    ObsSuccess    ObservationType = "success"
)

type Observation struct {
    Type      ObservationType
    Agent     string
    Target    string    // file path or task name
    Detail    string
    Timestamp time.Time
}

type PatternClassification string
const (
    ClassObservation   PatternClassification = "observation"
    ClassHeuristic     PatternClassification = "heuristic"
    ClassRule          PatternClassification = "rule"
    ClassHighConfidence PatternClassification = "high_confidence"
    ClassAntiPattern   PatternClassification = "anti_pattern"
)

type Pattern struct {
    Key            string
    Classification PatternClassification
    Count          int
    Observations   []Observation
    FirstSeen      time.Time
    LastSeen       time.Time
}
```

Acceptance:
- 모든 타입 정의
- JSON 직렬화/역직렬화 테스트

### REQ-6: Observation 저장소 (`internal/research/observe/storage.go`)

**[EARS: Event-driven]**

WHEN an observation is recorded, the system SHALL append it to a JSONL file.

```go
type Storage struct {
    baseDir string
}

func NewStorage(baseDir string) *Storage
func (s *Storage) Append(obs *Observation) error
func (s *Storage) LoadAll() ([]*Observation, error)
func (s *Storage) LoadSince(since time.Time) ([]*Observation, error)
```

Acceptance:
- JSONL 형식 (한 줄에 하나의 JSON 객체)
- Append는 observations.jsonl에 추가
- LoadAll은 전체 읽기
- LoadSince는 시간 필터
- 빈 파일/존재하지 않는 파일 처리
- t.TempDir()로 격리

### REQ-7: Pattern Detector (`internal/research/observe/pattern.go`)

**[EARS: Event-driven]**

WHEN observations are analyzed, the system SHALL detect patterns and classify them by frequency.

```go
type PatternDetector struct {
    thresholds PatternThresholds
}

type PatternThresholds struct {
    Heuristic      int // default 3
    Rule           int // default 5
    HighConfidence int // default 10
}

func NewPatternDetector(thresholds PatternThresholds) *PatternDetector
func (d *PatternDetector) Detect(observations []*Observation) []*Pattern
```

패턴 감지 로직:
- Agent+Target 조합을 key로 그룹핑
- 1x = Observation
- count >= Heuristic threshold → Heuristic
- count >= Rule threshold → Rule
- count >= HighConfidence threshold → HighConfidence

Acceptance:
- 빈 관찰 목록 → 빈 패턴 목록
- 1개 관찰 → 1개 Observation 패턴
- 3개 동일 패턴 → Heuristic 분류
- 5개 동일 패턴 → Rule 분류
- 10개 동일 패턴 → HighConfidence 분류
- 서로 다른 Agent+Target → 별도 패턴
- 테이블 드리븐 테스트

---

## Technical Approach

### 파일 생성 범위

**Phase 2 (experiment):**
- `internal/research/experiment/types.go` + `types_test.go`
- `internal/research/experiment/baseline.go` + `baseline_test.go`
- `internal/research/experiment/store.go` + `store_test.go`
- `internal/research/experiment/loop.go` + `loop_test.go`

**Phase 3 (observe):**
- `internal/research/observe/types.go` + `types_test.go`
- `internal/research/observe/storage.go` + `storage_test.go`
- `internal/research/observe/pattern.go` + `pattern_test.go`

### 의존성

```
internal/research/experiment/ → internal/research/eval (EvalResult 참조)
internal/research/observe/    → 독립 (외부 의존성 없음)
```

### 기존 코드 수정 없음

이 SPEC은 새 패키지만 생성. 기존 hook 통합(Phase 3의 나머지)은 별도 SPEC으로.
