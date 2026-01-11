---
id: SPEC-MODEL-001
type: acceptance
version: "1.0.0"
created: "2026-01-09"
updated: "2026-01-09"
---

# SPEC-MODEL-001: Acceptance Criteria

Multi-Model Router 인수 조건 및 테스트 시나리오

## Traceability

- SPEC: [spec.md](./spec.md)
- Plan: [plan.md](./plan.md)
- Depends on: SPEC-WEB-001

---

## 인수 조건 요약

| ID | 조건 | 우선순위 | 상태 |
|----|------|----------|------|
| AC-001 | "API 설계해줘" → Claude Opus 라우팅 | High | Pending |
| AC-002 | "코드 구현해줘" → GLM-4.7 라우팅 | High | Pending |
| AC-003 | Provider 전환 시 환경 변수 업데이트 | Medium | Pending |
| AC-004 | 비용 추적 정확도 95%+ | High | Pending |
| AC-005 | Fallback 시 대체 모델 동작 | High | Pending |
| AC-006 | 컨텍스트 전환 시 정보 보존 | Medium | Pending |

---

## 테스트 시나리오

### AC-001: Planning 작업 라우팅

**Scenario: "API 설계해줘" 메시지 처리**

```gherkin
Feature: Planning 작업 Claude Opus 라우팅

  Scenario: API 설계 요청 시 Tier 1 모델 사용
    Given 사용자가 인증된 상태이고
    And ModelRouter가 초기화된 상태에서
    When 사용자가 "API 설계해줘" 메시지를 전송하면
    Then TaskClassifier가 해당 메시지를 "planning" 작업으로 분류하고
    And ModelRouter가 Claude Opus 모델을 선택하고
    And 응답의 model_used 필드가 "claude-opus"이어야 한다
    And 응답의 tier 필드가 1이어야 한다

  Scenario: /moai:1-plan 명령 실행 시 Tier 1 모델 사용
    Given 사용자가 인증된 상태이고
    And ModelRouter가 초기화된 상태에서
    When 사용자가 "/moai:1-plan user authentication" 명령을 실행하면
    Then ModelRouter가 명령을 감지하고 Claude Opus를 강제 선택하고
    And classification_confidence가 1.0이어야 한다
```

---

### AC-002: Implementation 작업 라우팅

**Scenario: "코드 구현해줘" 메시지 처리**

```gherkin
Feature: Implementation 작업 GLM-4.7 라우팅

  Scenario: 코드 구현 요청 시 Tier 2 모델 사용
    Given 사용자가 인증된 상태이고
    And ModelRouter가 초기화된 상태에서
    When 사용자가 "코드 구현해줘" 메시지를 전송하면
    Then TaskClassifier가 해당 메시지를 "implementation" 작업으로 분류하고
    And ModelRouter가 GLM-4.7 모델을 선택하고
    And 응답의 model_used 필드가 "glm-4.7"이어야 한다
    And 응답의 tier 필드가 2이어야 한다

  Scenario: /moai:2-run 명령 실행 시 Tier 2 모델 사용
    Given 사용자가 인증된 상태이고
    And SPEC-001이 존재하는 상태에서
    When 사용자가 "/moai:2-run SPEC-001" 명령을 실행하면
    Then ModelRouter가 명령을 감지하고 GLM-4.7을 강제 선택하고
    And 코드 생성 작업이 GLM-4.7로 처리되어야 한다
```

---

### AC-003: Provider 전환

**Scenario: 환경 변수 기반 Provider 전환**

```gherkin
Feature: Provider 전환 및 환경 변수 업데이트

  Scenario: 런타임 Provider 전환
    Given 현재 Provider가 "claude"인 상태에서
    When 관리자가 "/api/router/force-model" API를 호출하여 "glm"으로 전환하면
    Then 환경 변수 ZAI_MODEL이 활성화되고
    And 후속 요청이 GLM-4.7로 라우팅되어야 한다

  Scenario: Worktree별 Provider 설정
    Given Worktree "feature-auth"가 존재하고
    And 해당 Worktree의 model_config가 "claude-opus"로 설정된 상태에서
    When 해당 Worktree에서 작업을 수행하면
    Then 분류 결과와 관계없이 Claude Opus가 사용되어야 한다
```

---

### AC-004: 비용 추적 정확도

**Scenario: 토큰 사용량 및 비용 기록**

```gherkin
Feature: 비용 추적 시스템 정확도

  Scenario: 단일 요청 비용 추적
    Given CostTracker가 초기화된 상태에서
    When Claude Opus로 1000 input tokens, 500 output tokens 요청이 처리되면
    Then CostTracker가 다음을 기록해야 한다:
      | field         | value         |
      | model         | claude-opus   |
      | input_tokens  | 1000          |
      | output_tokens | 500           |
      | cost_usd      | 0.0525        |
    And 기록된 비용이 실제 청구 금액과 95% 이상 일치해야 한다

  Scenario: 일일 비용 요약 조회
    Given 당일 10건의 요청이 처리된 상태에서
    When "/api/costs/daily" API를 호출하면
    Then 응답에 다음 필드가 포함되어야 한다:
      | field            | type   |
      | total_cost_usd   | float  |
      | tier1_cost_usd   | float  |
      | tier2_cost_usd   | float  |
      | tier1_percentage | float  |
      | tier2_percentage | float  |
    And tier1_percentage + tier2_percentage = 100%이어야 한다

  Scenario: 모델별 비용 분석
    Given 다양한 모델로 요청이 처리된 상태에서
    When "/api/costs/breakdown" API를 호출하면
    Then 각 모델별 비용 상세가 반환되어야 한다
    And 총 비용 합계가 daily summary와 일치해야 한다
```

---

### AC-005: Fallback 동작

**Scenario: API 에러 시 대체 모델 사용**

```gherkin
Feature: Fallback 시스템

  Scenario: Tier 1 모델 실패 시 Tier 2로 Fallback
    Given Claude Opus API가 503 에러를 반환하는 상태에서
    When Planning 작업 요청이 발생하면
    Then 시스템이 GLM-4.7로 Fallback하고
    And 응답 메타데이터에 "fallback: true"가 포함되어야 한다
    And 사용자에게 Fallback 사용 알림이 전달되어야 한다

  Scenario: 모든 모델 실패 시 에러 반환
    Given Claude Opus와 GLM-4.7 모두 API 에러 상태에서
    When 요청이 발생하면
    Then 적절한 에러 메시지가 반환되어야 한다
    And 에러 로그에 시도된 모든 모델이 기록되어야 한다

  Scenario: Rate Limit 시 백오프 및 재시도
    Given API가 429 Rate Limit 에러를 반환하는 상태에서
    When 요청이 발생하면
    Then 시스템이 지수 백오프로 재시도하고
    And 최대 3회 재시도 후 Fallback을 시도해야 한다
```

---

### AC-006: 컨텍스트 전환

**Scenario: 모델 전환 시 컨텍스트 보존**

```gherkin
Feature: 컨텍스트 전환 및 정보 보존

  Scenario: Tier 1에서 Tier 2로 전환
    Given Claude Opus로 SPEC 설계가 완료된 상태에서
    And 대화 히스토리에 설계 결정 사항이 포함된 상태에서
    When Implementation 작업으로 GLM-4.7로 전환하면
    Then 핵심 설계 결정 사항이 컨텍스트에 포함되어야 하고
    And GLM-4.7이 설계 내용을 인식하고 구현해야 한다

  Scenario: 컨텍스트 압축 효율성
    Given 50개 메시지의 대화 히스토리가 있는 상태에서
    When 모델 전환이 발생하면
    Then 컨텍스트가 압축되어 핵심 정보만 전달되고
    And 압축된 컨텍스트 크기가 원본의 30% 이하이어야 한다
    And 핵심 결정 사항, SPEC ID, 현재 작업 상태가 보존되어야 한다
```

---

## Quality Gates

### 코드 품질

```yaml
Test Coverage:
  - Unit Tests: >= 85%
  - Integration Tests: >= 70%
  - E2E Tests: 핵심 시나리오 100%

Static Analysis:
  - ruff: No errors
  - mypy: No type errors
  - security: No critical vulnerabilities
```

### 성능 기준

```yaml
Response Time:
  - P50: < 50ms (라우팅 결정)
  - P95: < 100ms (라우팅 결정)
  - P99: < 200ms (라우팅 결정)

Throughput:
  - Min: 100 req/s
  - Target: 500 req/s

Accuracy:
  - Classification: >= 95%
  - Cost Tracking: >= 95%
```

### 보안 기준

```yaml
Authentication:
  - 모든 API 엔드포인트 인증 필요
  - API 키 로그 노출 금지

Data Protection:
  - 비용 데이터 암호화 저장
  - 세션 데이터 24시간 후 자동 삭제
```

---

## Definition of Done

- [ ] 모든 인수 조건(AC-001 ~ AC-006) 통과
- [ ] Unit test coverage >= 85%
- [ ] Integration test 시나리오 모두 통과
- [ ] 성능 기준 충족 (P95 < 100ms)
- [ ] 보안 취약점 스캔 통과
- [ ] API 문서 업데이트 완료
- [ ] 코드 리뷰 승인
- [ ] SPEC-WEB-001 연동 테스트 완료

---

## 테스트 데이터

### 분류 테스트 케이스

```python
TIER1_TEST_CASES = [
    ("API 설계해줘", ModelTier.TIER1),
    ("아키텍처 결정 필요", ModelTier.TIER1),
    ("/moai:1-plan auth system", ModelTier.TIER1),
    ("기술 스택 비교 분석", ModelTier.TIER1),
    ("시스템 설계 전략 수립", ModelTier.TIER1),
]

TIER2_TEST_CASES = [
    ("코드 구현해줘", ModelTier.TIER2),
    ("테스트 작성해줘", ModelTier.TIER2),
    ("/moai:2-run SPEC-001", ModelTier.TIER2),
    ("이 버그 수정해줘", ModelTier.TIER2),
    ("문서화 진행해줘", ModelTier.TIER2),
]

EDGE_CASES = [
    ("", ModelTier.TIER2),  # 빈 문자열 → 기본 Tier
    ("안녕하세요", ModelTier.TIER2),  # 일반 인사 → 기본 Tier
]
```

### 비용 계산 테스트 케이스

```python
COST_TEST_CASES = [
    # (model, input_tokens, output_tokens, expected_cost)
    ("claude-opus", 1000, 500, 0.0525),
    ("claude-opus", 10000, 5000, 0.525),
    ("glm-4.7", 1000, 500, 0.002),
    ("glm-4.7", 10000, 5000, 0.02),
]
```

---

## 태그

- `SPEC-MODEL-001`
- `acceptance-criteria`
- `test-scenarios`
- `gherkin`
