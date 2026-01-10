---
id: SPEC-MODEL-001
type: plan
version: "1.0.0"
created: "2026-01-09"
updated: "2026-01-09"
---

# SPEC-MODEL-001: Implementation Plan

Multi-Model Router 구현 계획서

## Traceability

- SPEC: [spec.md](./spec.md)
- Acceptance: [acceptance.md](./acceptance.md)
- Depends on: SPEC-WEB-001

---

## 마일스톤

### Primary Goal: 핵심 라우팅 시스템

TaskClassifier와 ModelRouter 기본 구현

**구현 항목:**

1. `task_classifier.py` 생성
   - Rule-based 패턴 매칭 로직
   - Tier 1/Tier 2 분류 함수
   - 신뢰도 점수 계산

2. `model_router.py` 생성
   - ModelRouter 클래스 구현
   - Claude Opus 클라이언트 통합
   - GLM-4.7 클라이언트 통합

3. Provider 클라이언트 구현
   - `opus_client.py`: Anthropic SDK 래퍼
   - `glm_client.py`: OpenAI-compatible API 래퍼

**완료 기준:**
- 메시지 분류 정확도 90% 이상
- 기본 라우팅 동작 확인

---

### Secondary Goal: 비용 추적 시스템

CostTracker 구현 및 통계 API

**구현 항목:**

1. `cost_tracker.py` 생성
   - 토큰 사용량 기록
   - 비용 계산 로직
   - 일일/월간 집계

2. 데이터베이스 모델
   - `CostRecord` 테이블
   - 인덱스 최적화

3. API 엔드포인트
   - `/api/costs/daily`
   - `/api/costs/breakdown`

**완료 기준:**
- 비용 추적 정확도 95% 이상
- 실시간 통계 조회 가능

---

### Tertiary Goal: 컨텍스트 전환

모델 간 컨텍스트 압축 및 전달

**구현 항목:**

1. 컨텍스트 압축기
   - 대화 히스토리 요약
   - 핵심 정보 추출

2. 모델 전환 핸들러
   - 상태 동기화
   - 세션 연속성 유지

3. Worktree 연동
   - `model_config.py` 생성
   - Worktree별 모델 설정

**완료 기준:**
- 컨텍스트 전환 시 정보 손실 최소화
- 세션 연속성 유지

---

### Final Goal: Fallback 및 안정화

에러 처리 및 Fallback 시스템

**구현 항목:**

1. Fallback 전략 구현
   - 모델 대체 로직
   - 재시도 메커니즘
   - 캐시 응답 반환

2. 에러 핸들링
   - API 타임아웃 처리
   - Rate limit 대응
   - 네트워크 에러 복구

3. 모니터링 통합
   - 라우팅 결정 로깅
   - 성능 메트릭 수집

**완료 기준:**
- Fallback 시나리오 100% 처리
- 시스템 가용성 99.9%

---

## Technical Approach

### 아키텍처 설계

```
┌─────────────────────────────────────────────────────────┐
│                    Model Router                         │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐ │
│  │   Task      │    │   Router    │    │    Cost     │ │
│  │ Classifier  │───▶│    Core     │───▶│   Tracker   │ │
│  └─────────────┘    └─────────────┘    └─────────────┘ │
│         │                 │                   │         │
│         ▼                 ▼                   ▼         │
│  ┌─────────────────────────────────────────────────────┤
│  │              Provider Clients                       │ │
│  │  ┌──────────────┐     ┌──────────────┐             │ │
│  │  │ Opus Client  │     │ GLM Client   │             │ │
│  │  │ (Tier 1)     │     │ (Tier 2)     │             │ │
│  │  └──────────────┘     └──────────────┘             │ │
│  └─────────────────────────────────────────────────────┤
└─────────────────────────────────────────────────────────┘
```

### 분류 알고리즘

```python
def classify(message: str) -> ModelTier:
    # Phase 1: Command Detection
    if "/moai:1-plan" in message:
        return ModelTier.TIER1
    if "/moai:2-run" in message:
        return ModelTier.TIER2

    # Phase 2: Keyword Matching
    tier1_score = match_patterns(message, TIER1_PATTERNS)
    tier2_score = match_patterns(message, TIER2_PATTERNS)

    # Phase 3: Confidence Threshold
    if tier1_score > CONFIDENCE_THRESHOLD:
        return ModelTier.TIER1

    # Default to Tier 2 (cost-effective)
    return ModelTier.TIER2
```

### 비용 계산 모델

```
Claude Opus 4.5:
  - Input: $0.015 / 1K tokens
  - Output: $0.075 / 1K tokens

GLM-4.7 via Z.AI:
  - Input: $0.001 / 1K tokens
  - Output: $0.002 / 1K tokens

예상 비용 절감:
  - Tier 1 (10%): ~$10/일
  - Tier 2 (90%): ~$2/일
  - 총 절감: ~80%
```

### 환경 변수 우선순위

```
1. Force Model (force_model 파라미터)
2. Command Override (/moai:1-plan, /moai:2-run)
3. Worktree Config (model_config.py)
4. Classifier Decision
5. Default Tier (MODEL_ROUTER_DEFAULT_TIER)
```

---

## 리스크 및 대응

### 기술적 리스크

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| 분류 정확도 부족 | High | 키워드 패턴 지속 개선, 피드백 루프 |
| API Rate Limit | Medium | 요청 큐잉, 백오프 전략 |
| 컨텍스트 손실 | Medium | 핵심 정보 태깅, 요약 알고리즘 개선 |
| 비용 폭증 | High | 일일 한도 설정, 알림 시스템 |

### 의존성 리스크

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| Z.AI API 불안정 | Medium | Claude를 Fallback으로 사용 |
| Anthropic API 변경 | Low | SDK 버전 고정, 마이그레이션 계획 |

---

## 구현 순서

```
Phase 1: Foundation (Primary Goal)
  └── TaskClassifier → ModelRouter → Provider Clients

Phase 2: Tracking (Secondary Goal)
  └── CostTracker → Database → Statistics API

Phase 3: Context (Tertiary Goal)
  └── Context Compressor → Model Switcher → Worktree Integration

Phase 4: Hardening (Final Goal)
  └── Fallback System → Error Handling → Monitoring
```

---

## 테스트 전략

### Unit Tests

- TaskClassifier 분류 로직
- CostTracker 계산 정확도
- Provider 클라이언트 모킹

### Integration Tests

- End-to-end 라우팅 흐름
- 모델 전환 시나리오
- Fallback 동작

### Load Tests

- 동시 요청 처리
- Rate limit 시나리오
- 비용 추적 성능

---

## 성공 지표

| 지표 | 목표값 | 측정 방법 |
|------|--------|-----------|
| 분류 정확도 | 95%+ | 수동 검증 샘플링 |
| 비용 절감율 | 70%+ | 일일 비용 비교 |
| 라우팅 지연 | <100ms | P95 응답 시간 |
| Fallback 성공률 | 99%+ | 에러 로그 분석 |

---

## 태그

- `SPEC-MODEL-001`
- `implementation-plan`
- `multi-model`
- `router`
