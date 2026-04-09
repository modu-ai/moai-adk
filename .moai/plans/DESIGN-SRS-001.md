# DESIGN-SRS-001: MoAI Self-Research System (SRS)

> **Status**: Draft
> **Author**: MoAI Orchestrator
> **Date**: 2026-04-09
> **Source**: autoresearch-skill (olelehmann100kMRR) + Agency Learner 패턴

---

## 1. Executive Summary

MoAI Self-Research System(SRS)은 moai-adk의 스킬, 에이전트, 룰, CLAUDE.md 등 모든 지침 파일을 **자율적으로 실험하고 개선**하는 시스템이다.

기존 Agency Learner의 **수동적 관찰**(실제 사용에서 패턴 수집)과 autoresearch의 **능동적 실험**(의도적 변이-테스트-보존 루프)을 결합하여 **Two-Loop 자가 개선 아키텍처**를 구현한다.

```
┌─────────────────────────────────────────────────────────────┐
│                    MoAI Self-Research System                 │
│                                                             │
│  Loop 1: Passive Observation (Always-On)                    │
│  ┌─────────┐    ┌──────────┐    ┌────────────┐             │
│  │ Hook    │───>│ Observe  │───>│ Pattern    │──┐          │
│  │ Events  │    │ Collect  │    │ Detect     │  │          │
│  └─────────┘    └──────────┘    └────────────┘  │          │
│                                                  │ Trigger  │
│  Loop 2: Active Research (On-Demand)             ▼          │
│  ┌─────────┐    ┌──────────┐    ┌────────────┐             │
│  │ Read    │───>│ Baseline │───>│ Experiment │──┐          │
│  │ Target  │    │ Establish│    │ Loop       │  │          │
│  └─────────┘    └──────────┘    └────────────┘  │          │
│                                                  │          │
│  ┌─────────┐    ┌──────────┐    ┌────────────┐  │          │
│  │ Accept/ │<───│ Canary   │<───│ Score &    │<─┘          │
│  │ Discard │    │ Validate │    │ Compare    │             │
│  └─────────┘    └──────────┘    └────────────┘             │
│                                                             │
│  Flywheel: Usage → Observe → Detect → Research → Improve   │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 분석: autoresearch-skill 핵심 패턴

### 2.1 SKILL.md 핵심 원리

| 원리 | 설명 | MoAI 적용 |
|------|------|-----------|
| Binary Evals | Yes/No만, 척도 없음 | 모든 eval은 pass/fail |
| One Change at a Time | 변경 1개씩, 원인 추적 가능 | worktree에서 단일 변이 |
| Baseline First | 개선 전 현재 상태 측정 | 실험 전 baseline 필수 |
| Never Stop | 수동 중단/예산/95%까지 계속 | 설정 가능한 종료 조건 |
| Full Logging | 모든 실험 기록 | changelog + dashboard |

### 2.2 eval-guide.md 핵심 원리

| 원리 | 설명 |
|------|------|
| 3-Question Test | 1) 두 에이전트가 동의할까? 2) 게이밍 가능한가? 3) 사용자가 관심있는가? |
| Max 6 Evals | 6개 초과 시 게이밍 위험 |
| No Overlap | 각 eval은 독립적 차원 측정 |
| Measurable by Agent | 주관적 판단 → 관찰 가능한 신호로 변환 |

---

## 3. Research Target 분류 (4-Tier)

### Tier 1: Skills (직접 연구 가능)

가장 자연스러운 autoresearch 대상. 입력 → 출력이 명확.

```
Target: .claude/skills/moai-lang-go/SKILL.md
Method: 테스트 프롬프트로 스킬 실행 → 출력물에 binary eval 적용
Eval 예시:
  - "생성된 Go 코드가 컴파일되는가?"
  - "Go 1.23+ 기능을 적절히 사용하는가?"
  - "모든 에러를 context와 함께 wrap하는가?"
```

### Tier 2: Agents (태스크 기반 연구)

에이전트에 참조 태스크를 주고 결과물 평가.

```
Target: .claude/agents/moai/expert-backend.md
Method: 참조 태스크 실행 → 결과물에 binary eval 적용
Eval 예시:
  - "생성된 코드가 컴파일되고 테스트를 통과하는가?"
  - "요청한 파일만 수정했는가?"
  - "새 기능에 대한 테스트를 포함하는가?"
Cost: 높음 (에이전트 실행당 토큰 소비 큼) → runs_per_experiment 줄임
```

### Tier 3: Rules (행동 기반 연구)

규칙 변경이 워크플로우 행동에 미치는 영향 측정.

```
Target: .claude/rules/moai/development/coding-standards.md
Method: A/B 테스트 - 동일 워크플로우를 규칙 변경 전후로 실행 비교
Eval 예시 (Proxy Evals):
  - "이 규칙 적용 후 사용자 교정 횟수가 줄었는가?"
  - "토큰 소비가 10% 이상 증가하지 않았는가?"
  - "에이전트 에러율이 감소했는가?"
```

### Tier 4: Configuration (설정 연구)

설정값 변경의 시스템 전체 영향 측정.

```
Target: .moai/config/sections/quality.yaml
Method: 설정 변경 후 참조 워크플로우 실행 → 품질 지표 비교
제한: 안전 임계값(TRUST 5 최소값)은 FROZEN
```

---

## 4. 아키텍처 설계

### 4.1 시스템 구성요소

```
┌──────────────────────────────────────────────────────┐
│                  /moai research                       │
│              (Entry Point - Skill/CLI)                │
└──────────────┬───────────────────────────────────────┘
               │
┌──────────────▼───────────────────────────────────────┐
│           Researcher Agent (moai/researcher)          │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Reader  │ │ Eval    │ │Experiment│ │ Reporter │ │
│  │ Phase   │ │ Builder │ │ Runner   │ │ Phase    │ │
│  └────┬────┘ └────┬────┘ └────┬─────┘ └────┬─────┘ │
└───────┼──────────┼──────────┼────────────┼──────────┘
        │          │          │            │
┌───────▼──────────▼──────────▼────────────▼──────────┐
│              Infrastructure Layer                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│  │ Eval     │ │ Worktree │ │ Canary   │ │Results │ │
│  │ Engine   │ │ Sandbox  │ │ Guard    │ │ Store  │ │
│  └──────────┘ └──────────┘ └──────────┘ └────────┘ │
└─────────────────────────────────────────────────────┘
        │                                      │
┌───────▼──────────────────────────────────────▼──────┐
│           Existing MoAI Infrastructure               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │ Agency   │ │evaluator │ │ Auto     │            │
│  │ Learner  │ │ -active  │ │ Memory   │            │
│  └──────────┘ └──────────┘ └──────────┘            │
└─────────────────────────────────────────────────────┘
```

### 4.2 Researcher Agent 정의

```yaml
# .claude/agents/moai/researcher.md
name: researcher
description: >
  Active self-research agent that optimizes moai-adk components
  (skills, agents, rules, CLAUDE.md) through iterative experimentation
  with binary eval criteria. Uses worktree isolation for safe mutation.
tools: Read, Write, Edit, Grep, Glob, Bash, Agent
model: opus
permissionMode: acceptEdits
maxTurns: 200
memory: project
skills:
  - moai-workflow-research
```

### 4.3 Research Workflow Skill

```yaml
# .claude/skills/moai-workflow-research/SKILL.md
name: moai-workflow-research
description: >
  Self-research workflow for optimizing moai-adk components through
  binary eval experimentation loops. Implements the autoresearch pattern
  adapted for multi-tier component types.
```

---

## 5. Eval Suite 스키마

### 5.1 Eval Definition (YAML)

```yaml
# .moai/research/evals/skills/moai-lang-go.eval.yaml
target:
  path: .claude/skills/moai-lang-go/SKILL.md
  type: skill                    # skill | agent | rule | config

test_inputs:
  - name: rest-api-endpoint
    prompt: |
      Create a REST API endpoint for user CRUD operations
      using Go Fiber framework with proper error handling.
  - name: concurrent-worker-pool
    prompt: |
      Implement a worker pool pattern in Go with graceful
      shutdown and context cancellation support.
  - name: database-repository
    prompt: |
      Write a PostgreSQL repository pattern in Go using
      sqlx with transaction support and proper error wrapping.

evals:
  - name: compiles
    question: "Does the generated Go code compile without errors?"
    pass: "go build succeeds with zero errors"
    fail: "Any compilation error occurs"
    weight: must_pass              # must_pass | nice_to_have

  - name: modern-go
    question: "Does the code use Go 1.23+ features where appropriate?"
    pass: "Uses modern patterns (range-over-func, enhanced ServeMux, etc.)"
    fail: "Uses deprecated patterns when modern alternatives exist"
    weight: nice_to_have

  - name: error-handling
    question: "Does every fallible function return and handle errors properly?"
    pass: "All errors checked, wrapped with fmt.Errorf context"
    fail: "Any unchecked error or bare error without context"
    weight: must_pass

  - name: no-placeholders
    question: "Is the code free of TODO, FIXME, or placeholder comments?"
    pass: "Zero placeholder comments in output"
    fail: "Any TODO, FIXME, or 'implement this' comment exists"
    weight: nice_to_have

  - name: test-included
    question: "Does the output include unit tests for the main functionality?"
    pass: "At least one test function with meaningful assertions"
    fail: "No tests or only empty test stubs"
    weight: nice_to_have

settings:
  runs_per_experiment: 3          # 실험당 실행 횟수
  max_experiments: 20             # 최대 실험 횟수
  pass_threshold: 0.80            # 통과 기준 (must_pass는 항상 1.0)
  target_score: 0.95              # 목표 점수 (도달 시 자동 종료)
  budget_cap_tokens: 500000       # 토큰 예산 상한
  stagnation_threshold: 0.03     # 정체 감지 임계값
  stagnation_patience: 2          # 정체 허용 연속 횟수
```

### 5.2 Agent Eval Definition

```yaml
# .moai/research/evals/agents/expert-backend.eval.yaml
target:
  path: .claude/agents/moai/expert-backend.md
  type: agent

test_inputs:
  - name: jwt-auth-endpoint
    prompt: |
      Implement JWT authentication middleware and login/logout
      endpoints for a Go Fiber application. Include token
      refresh logic and proper error responses.
    reference_files:
      - tests/fixtures/auth-expected-structure.yaml

  - name: crud-with-validation
    prompt: |
      Create a complete CRUD API for a 'Product' entity with
      input validation, pagination, and proper HTTP status codes.

evals:
  - name: code-compiles
    question: "Does all generated code compile without errors?"
    pass: "go build ./... succeeds"
    fail: "Any compilation error"
    weight: must_pass

  - name: tests-pass
    question: "Do the generated tests pass?"
    pass: "go test ./... exits with 0"
    fail: "Any test failure"
    weight: must_pass

  - name: focused-changes
    question: "Did the agent modify only files relevant to the task?"
    pass: "All modified files are within the task's domain scope"
    fail: "Unrelated files were modified"
    weight: nice_to_have

  - name: includes-tests
    question: "Did the agent create test files for new functionality?"
    pass: "At least one _test.go file created with test functions"
    fail: "No test files created"
    weight: nice_to_have

settings:
  runs_per_experiment: 2          # 에이전트는 비용이 높으므로 적게
  max_experiments: 10
  budget_cap_tokens: 1000000
```

---

## 6. Experiment Loop 프로토콜

### 6.1 실험 루프 (autoresearch 적응)

```
Step 1: READ TARGET
  → 대상 컴포넌트 전체 읽기
  → 핵심 기능, 프로세스, 기존 품질 체크 식별
  → 수정 가능한 영역 식별 (FROZEN 제외)

Step 2: LOAD/BUILD EVAL SUITE
  → .moai/research/evals/ 에서 기존 eval 로드
  → 없으면: 사용자에게 eval 정의 요청 (3-6개 binary checks)
  → 3-Question Test로 eval 품질 검증

Step 3: ESTABLISH BASELINE
  → 대상 컴포넌트를 변경 없이 실행
  → 모든 test_input에 대해 eval 실행
  → 결과를 .moai/research/baselines/ 에 저장
  → [REQUIRED] 사용자에게 baseline 확인 요청

Step 4: EXPERIMENT LOOP
  ┌─ ANALYZE: 마지막 실행의 실패 패턴 분석
  │  → 어떤 eval이 실패했는가?
  │  → 왜 실패했는가? (출력물 분석)
  │
  ├─ HYPOTHESIZE: 개선 가설 수립
  │  → 하나의 구체적 변경 제안
  │  → 예: "Step 3에 에러 핸들링 예시를 추가하면 error-handling eval 통과율이 오른다"
  │
  ├─ MUTATE: worktree에서 ONE change 적용
  │  → Agent(isolation: "worktree")로 격리된 환경에서 수정
  │  → 변경은 반드시 1개만
  │
  ├─ TEST: eval suite 실행
  │  → runs_per_experiment 횟수만큼 반복
  │  → 각 test_input에 대해 모든 eval 실행
  │
  ├─ SCORE: 결과 비교
  │  → 이전 best score와 비교
  │  → must_pass eval은 개별 통과 필수
  │
  ├─ DECIDE: keep or discard
  │  → 개선 시: KEEP (worktree 변경 merge)
  │  → 동일/퇴보 시: DISCARD (worktree 폐기)
  │
  └─ LOG: 실험 결과 기록
     → .moai/research/experiments/{target}/exp-NNN.json
     → changelog.md에 추가

  종료 조건:
  - 사용자 중단
  - max_experiments 도달
  - budget_cap_tokens 소진
  - target_score (95%) 달성 3회 연속
  - stagnation_patience 초과 (개선 < stagnation_threshold)

Step 5: DELIVER RESULTS
  → 점수 요약 (baseline → final)
  → 실험 수, keep 비율
  → 상위 3개 효과적 변경
  → 남은 실패 패턴
  → 개선된 컴포넌트 파일
  → dashboard.html 생성
```

### 6.2 Passive Observation 수집

```
Hook Events → Observation Collector → Pattern Detector → Research Trigger

수집 대상:
┌─────────────────────┬───────────────────────────────────┐
│ Event               │ Observation Signal                │
├─────────────────────┼───────────────────────────────────┤
│ SubagentStop        │ 에이전트 성공/실패, 소요 시간      │
│ PostToolUse (Edit)  │ 사용자가 에이전트 출력 직후 수정   │
│                     │ = 교정 신호                        │
│ TaskCompleted       │ 태스크 완료 품질 점수              │
│ User feedback       │ /moai feedback 통한 명시적 피드백  │
│ Build errors        │ 에이전트 생성 코드의 빌드 실패     │
│ Test failures       │ 에이전트 생성 테스트의 실패         │
└─────────────────────┴───────────────────────────────────┘

패턴 감지 임계값:
  3x 동일 패턴 → Heuristic (연구 후보 등록)
  5x 동일 패턴 → Rule (자동 연구 트리거)
  1x 치명적 실패 → Anti-Pattern (즉시 기록)
```

---

## 7. Safety Architecture (5 Layers)

### Layer 1: Frozen Guard

```yaml
frozen_targets:                           # 절대 수정 불가
  - .claude/rules/moai/core/moai-constitution.md
  - .claude/rules/agency/constitution.md
  - .claude/agents/moai/researcher.md     # 자기 자신
  - .moai/research/config.yaml            # 자기 설정
  sections:
    - "## FROZEN ZONE"                    # 모든 파일의 FROZEN 섹션
    - "### Safety Rails"
    - "### Ethical Boundaries"
```

### Layer 2: Worktree Sandbox

```
모든 실험은 worktree 격리 환경에서 실행:

Main Repo ─────────────────────────────
    │
    ├── .claude/worktrees/research-exp-001/   ← 실험 1
    ├── .claude/worktrees/research-exp-002/   ← 실험 2
    └── (main은 절대 직접 수정하지 않음)

규칙:
- [HARD] 실험 변이는 반드시 worktree에서 실행
- [HARD] 성공한 변경만 main에 merge (사용자 승인 후)
- [HARD] 실패한 worktree는 즉시 폐기
```

### Layer 3: Canary Regression

```
변경 수락 전 회귀 테스트:

1. 제안된 변경을 메모리에 적용 (디스크 아님)
2. 최근 3개 baseline에 대해 재평가
3. 어떤 baseline이라도 0.10 이상 하락 시 거부
4. 결과 기록 (통과/거부 무관)
```

### Layer 4: Rate Limiter

```yaml
rate_limits:
  max_experiments_per_session: 20        # 세션당 최대 실험
  max_accepted_per_session: 5            # 세션당 최대 수락
  max_research_per_week: 3               # 주간 자동 연구 최대
  cooldown_between_auto_research: 24h    # 자동 연구 간 쿨다운
  max_concurrent_targets: 3              # 동시 연구 대상 최대
```

### Layer 5: Human Approval

```
승인 필수 지점:

1. [REQUIRED] Baseline 확인 - 실험 루프 시작 전
2. [CONFIGURABLE] 변경 수락 - 개선 발견 시
   - auto_accept: true → improvement > threshold 시 자동 수락
   - auto_accept: false → 매번 사용자 확인
3. [REQUIRED] 최종 배포 - main 브랜치 merge 전
4. [REQUIRED] CLAUDE.md/rules 변경 - Tier 3-4는 항상 수동 승인
```

---

## 8. 기존 시스템 통합

### 8.1 Agency Learner와의 관계

```
Agency Learner (수동)          Researcher (능동)
─────────────────────         ─────────────────────
실제 프로젝트에서 관찰 수집     의도적 실험으로 개선 탐색
패턴 임계값 도달 시 졸업        실험 결과로 즉시 개선
느린 피드백 루프 (주/월)        빠른 피드백 루프 (분/시간)
모든 세션에서 자동 수집         명시적 트리거로 실행

통합:
Learner의 관찰 → Researcher의 연구 트리거
Researcher의 결과 → Learner의 졸업 프로토콜로 최종 적용
```

### 8.2 evaluator-active와의 관계

```
evaluator-active의 4차원 평가 → Research eval 기준의 소스
  - Functionality → "기능이 정상 동작하는가?"
  - Security → "보안 취약점이 없는가?"
  - Craft → "코딩 표준을 따르는가?"
  - Consistency → "기존 패턴과 일관적인가?"
```

### 8.3 SPEC Workflow 통합

```
/moai plan "research moai-lang-go skill"
  → SPEC-RESEARCH-001 생성

/moai run SPEC-RESEARCH-001
  → Researcher 에이전트 실행
  → 실험 루프 수행
  → 결과 .moai/research/에 저장

/moai sync SPEC-RESEARCH-001
  → changelog 생성
  → 개선된 파일 merge
  → PR 생성
```

### 8.4 Auto-Memory 통합

```
연구에서 발견된 핵심 인사이트 → auto-memory에 저장:
  - "moai-lang-go 스킬에서 에러 핸들링 예시가 3개 이상일 때 eval 통과율 85%→95%"
  - "expert-backend 에이전트에 table-driven test 지시를 추가하면 테스트 품질 향상"

향후 세션에서 유사 작업 시 자동 참조
```

---

## 9. 파일 구조

```
.moai/research/                              # Research 데이터 디렉토리
├── config.yaml                              # Research 시스템 설정
├── evals/                                   # Eval suite 정의
│   ├── skills/                              # 스킬별 eval
│   │   ├── moai-lang-go.eval.yaml
│   │   ├── moai-lang-typescript.eval.yaml
│   │   ├── moai-domain-backend.eval.yaml
│   │   └── ...
│   ├── agents/                              # 에이전트별 eval
│   │   ├── expert-backend.eval.yaml
│   │   ├── manager-tdd.eval.yaml
│   │   └── ...
│   └── rules/                               # 룰별 eval
│       ├── coding-standards.eval.yaml
│       └── ...
├── baselines/                               # Baseline 스냅샷
│   ├── moai-lang-go.baseline.json
│   └── ...
├── experiments/                             # 실험 로그
│   ├── moai-lang-go/
│   │   ├── exp-001.json                    # 개별 실험 결과
│   │   ├── exp-002.json
│   │   └── changelog.md                    # 변경 이력
│   └── ...
├── observations/                            # 수동 관찰 데이터
│   ├── corrections.jsonl                    # 사용자 교정 기록
│   ├── failures.jsonl                       # 에이전트 실패 기록
│   └── patterns.yaml                        # 감지된 패턴
├── results/                                 # 집계 결과
│   ├── results.json
│   └── results.tsv
└── dashboard.html                           # 라이브 대시보드

.claude/agents/moai/researcher.md            # Researcher 에이전트
.claude/skills/moai-workflow-research/
    └── SKILL.md                             # Research 워크플로우 스킬
```

---

## 10. Research Config 스키마

```yaml
# .moai/research/config.yaml
research:
  version: "1.0.0"
  enabled: true

  # 기본 설정
  defaults:
    runs_per_experiment: 3
    max_experiments: 20
    pass_threshold: 0.80
    target_score: 0.95
    budget_cap_tokens: 500000
    stagnation_threshold: 0.03
    stagnation_patience: 2

  # 안전 설정
  safety:
    worktree_isolation: true           # [HARD] 항상 true
    canary_regression_threshold: 0.10
    canary_baseline_count: 3
    auto_accept_improvement: false     # true시 threshold 이상 자동 수락
    auto_accept_threshold: 0.15        # auto_accept 시 최소 개선폭
    require_approval_for_rules: true   # Tier 3-4는 항상 수동 승인

  # 속도 제한
  rate_limits:
    max_experiments_per_session: 20
    max_accepted_per_session: 5
    max_auto_research_per_week: 3
    cooldown_hours: 24
    max_concurrent_targets: 3

  # 수동 관찰 설정
  passive:
    enabled: true
    correction_signal_window: 60       # 에이전트 출력 후 N초 내 수정 = 교정
    pattern_trigger_threshold: 5       # N회 동일 패턴 → 자동 연구 트리거
    observation_retention_days: 90

  # 대시보드 설정
  dashboard:
    auto_refresh_seconds: 10
    chart_library: "chart.js"

  # Tier별 설정 오버라이드
  tier_overrides:
    agent:
      runs_per_experiment: 2           # 에이전트는 비용 높음
      max_experiments: 10
      budget_cap_tokens: 1000000
    rule:
      require_approval: true           # 항상 수동 승인
      max_experiments: 5
    config:
      require_approval: true
      max_experiments: 5
```

---

## 11. 실험 결과 스키마

### 11.1 개별 실험 (exp-NNN.json)

```json
{
  "experiment_id": "exp-001",
  "target": ".claude/skills/moai-lang-go/SKILL.md",
  "timestamp": "2026-04-09T14:30:00Z",
  "hypothesis": "Add concrete error wrapping examples in Step 4 to improve error-handling eval",
  "change": {
    "type": "addition",
    "section": "Step 4: Error Handling Patterns",
    "diff": "--- a/SKILL.md\n+++ b/SKILL.md\n@@ -45,6 +45,12 @@\n+Always wrap errors with context:\n+```go\n+if err != nil {\n+    return fmt.Errorf(\"fetchUser: %w\", err)\n+}\n+```"
  },
  "results": {
    "runs": 3,
    "scores": [0.85, 0.90, 0.85],
    "average": 0.867,
    "per_eval": {
      "compiles": {"pass": 3, "fail": 0},
      "modern-go": {"pass": 2, "fail": 1},
      "error-handling": {"pass": 3, "fail": 0},
      "no-placeholders": {"pass": 2, "fail": 1},
      "test-included": {"pass": 2, "fail": 1}
    }
  },
  "comparison": {
    "previous_best": 0.733,
    "improvement": 0.134,
    "decision": "KEEP"
  },
  "tokens_used": 25000
}
```

### 11.2 Changelog (changelog.md)

```markdown
# Research Changelog: moai-lang-go

## Experiment 001 — KEEP (+0.134)
- **Score**: 0.733 → 0.867
- **Change**: Added concrete error wrapping examples in Step 4
- **Reasoning**: error-handling eval was failing because skill lacked specific patterns
- **Result**: error-handling pass rate 1/3 → 3/3
- **Remaining failures**: modern-go (1/3), no-placeholders (1/3), test-included (1/3)

## Experiment 002 — DISCARD (-0.033)
- **Score**: 0.867 → 0.834
- **Change**: Added mandatory test generation instruction
- **Reasoning**: Attempted to improve test-included eval
- **Result**: test-included improved but compiles regressed (forced tests broke build)
- **Decision**: Discarded — regression in must_pass eval
```

---

## 12. /moai research 커맨드

### 12.1 서브커맨드

```
/moai research <target>              # 대상에 대한 능동적 연구 시작
/moai research --eval <target>       # eval suite만 정의/편집
/moai research --baseline <target>   # baseline만 측정/갱신
/moai research --report              # 전체 연구 대시보드 표시
/moai research --status              # 현재 연구 상태 표시
/moai research --observe             # 수동 관찰 데이터 요약
```

### 12.2 사용 예시

```
사용자: /moai research moai-lang-go

MoAI:
🤖 MoAI ★ Research Start ─────────────────────
📋 대상: .claude/skills/moai-lang-go/SKILL.md
🔬 Eval Suite: 5개 binary checks 로드됨
📊 Baseline 측정 중...

Baseline 결과:
  compiles:       2/3 (66.7%)  ← must_pass
  modern-go:      1/3 (33.3%)
  error-handling:  1/3 (33.3%)  ← must_pass
  no-placeholders: 3/3 (100%)
  test-included:   1/3 (33.3%)
  Overall: 0.533

이 baseline으로 실험을 시작할까요?
────────────────────────────────────────────
```

---

## 13. 구현 로드맵

### Phase 1: Foundation (Week 1-2)

```
- [ ] .moai/research/config.yaml 스키마 정의
- [ ] Eval suite YAML 스키마 및 파서
- [ ] Baseline 측정 및 저장 로직
- [ ] 결과 스키마 (exp-NNN.json, changelog.md)
```

### Phase 2: Skill Research (Week 3-4)

```
- [ ] Researcher 에이전트 정의 (.claude/agents/moai/researcher.md)
- [ ] Research 워크플로우 스킬 (.claude/skills/moai-workflow-research/SKILL.md)
- [ ] Experiment loop 구현 (worktree 격리)
- [ ] /moai research 슬래시 커맨드
- [ ] 2-3개 스킬에 대한 eval suite 작성 (moai-lang-go, moai-domain-backend)
- [ ] 첫 번째 end-to-end 연구 실행 및 검증
```

### Phase 3: Agent Research (Week 5-6)

```
- [ ] Agent eval 전략 구현 (태스크 기반)
- [ ] 참조 태스크 픽스처 시스템
- [ ] Agent 실행 비용 최적화 (토큰 예산 관리)
- [ ] expert-backend, manager-tdd에 대한 eval suite
```

### Phase 4: Passive Loop (Week 7-8)

```
- [ ] Hook 기반 관찰 수집기 (corrections.jsonl, failures.jsonl)
- [ ] 패턴 감지 엔진 (임계값 기반 트리거)
- [ ] Agency Learner 통합 (관찰 → 연구 트리거)
- [ ] Auto-memory 연동 (핵심 인사이트 저장)
```

### Phase 5: Rules & CLAUDE.md Research (Week 9-10)

```
- [ ] A/B 테스트 프레임워크 (규칙 변경 전후 비교)
- [ ] Proxy eval 시스템 (간접 측정)
- [ ] CLAUDE.md 섹션별 eval suite
- [ ] Frozen guard 강화 검증
```

### Phase 6: Dashboard & Reporting (Week 11-12)

```
- [ ] Chart.js 기반 HTML 대시보드
- [ ] 자동 갱신 (10초)
- [ ] 점수 추이, 실험 상태, eval별 breakdown
- [ ] TSV/JSON 내보내기
- [ ] /moai research --report 통합
```

---

## 14. autoresearch와의 차이점

| 측면 | autoresearch (원본) | MoAI SRS (적응) |
|------|-------------------|----------------|
| **대상** | 단일 스킬 | 스킬 + 에이전트 + 룰 + 설정 (4-Tier) |
| **트리거** | 사용자 수동 | 수동 + 수동 관찰 자동 트리거 |
| **격리** | 없음 (직접 수정) | Worktree 격리 필수 |
| **안전** | 없음 | 5-Layer 안전 아키텍처 |
| **종료** | Never Stop | 설정 가능한 종료 조건 |
| **Eval 범위** | 직접 출력 | 직접 + Proxy eval (행동 기반) |
| **학습 통합** | 없음 | Agency Learner + Auto-Memory |
| **비용 관리** | 예산 상한만 | Tier별 차등 + 토큰 추적 |
| **롤백** | .baseline 파일 | Git commit + worktree 폐기 |

---

## 15. Open Questions

1. **Eval 자동 생성**: 사용자가 eval을 직접 정의하는 대신, 컴포넌트를 분석하여 eval을 자동 제안할 수 있는가?
2. **Cross-Component 영향**: 스킬 A를 개선했을 때 에이전트 B의 성능이 변하는 교차 영향을 어떻게 측정할 것인가?
3. **비용 최적화**: Agent 연구(Tier 2)의 토큰 비용을 줄이기 위해 CG Mode(GLM 워커)를 활용할 수 있는가?
4. **Eval Overfitting**: autoresearch의 "6개 이상 eval 시 게이밍" 문제를 어떻게 방지할 것인가?
5. **Passive Observer Hook**: 기존 hook 시스템에 관찰 수집을 추가할 때 성능 영향은?

---

**Version**: 1.0.0 (Draft)
**Classification**: Design Proposal
**Next Step**: GOOS행님 검토 후 Phase 1 구현 시작
