---
id: SPEC-MODEL-001
version: "1.0.0"
status: "draft"
created: "2026-01-09"
updated: "2026-01-09"
author: "MoAI-ADK"
priority: "medium"
depends_on: ["SPEC-WEB-001"]
---

# SPEC-MODEL-001: Multi-Model Router - Claude Opus + GLM-4.7

멀티 모델 라우팅 시스템으로 작업 유형에 따라 최적의 모델로 자동 라우팅

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-01-09 | MoAI-ADK | 초기 SPEC 작성 |

---

## 개요

### 목적

작업 유형(Planning vs Implementation)에 따라 Claude Opus 4.5와 GLM-4.7 모델을 자동으로 라우팅하여 비용 효율성을 극대화하면서 최적의 품질을 유지하는 시스템을 구현한다.

### 범위

- 작업 유형 분류 (Rule-based classifier)
- 모델 자동 라우팅 (Tier 1/Tier 2)
- 환경 변수 기반 Provider 전환
- 비용 추적 (모델별 토큰 사용량)
- Fallback 처리 및 에러 복구
- 컨텍스트 압축 및 전달

---

## Environment (환경)

### 기술 스택

```yaml
Models:
  Tier 1 (Planner):
    - Claude Opus 4.5
    - claude-agent-sdk: ">=0.1.19"
  Tier 2 (Implementer):
    - GLM-4.7 via Z.AI
    - OpenAI-compatible API

Router:
  - Task classifier (rule-based)
  - Context compressor
  - Response aggregator

Dependencies:
  - FastAPI: ">=0.115.0"
  - httpx: ">=0.27.0"
  - pydantic: ">=2.9.0"
```

### 시스템 요구사항

- Python: 3.11 이상
- 환경 변수: `ANTHROPIC_API_KEY`, `ZAI_API_KEY`
- 네트워크: HTTPS 지원

---

## Assumptions (가정)

### 기술적 가정

1. Claude Opus 4.5가 Planning 작업에서 우수한 성능을 발휘함
2. GLM-4.7이 Implementation 작업에서 충분한 품질을 제공함
3. OpenAI-compatible API 형식으로 GLM-4.7에 접근 가능
4. 컨텍스트 압축으로 모델 간 전환 시 정보 손실 최소화 가능

### 비즈니스 가정

1. Tier 2 모델 사용으로 90% 토큰 비용 절감 가능
2. Planning 작업은 전체 토큰의 10% 미만 사용
3. 작업 유형 분류 정확도 95% 이상 달성 가능

---

## Requirements (요구사항)

### Ubiquitous (항상 활성)

| ID | 요구사항 | 검증 방법 |
|----|----------|-----------|
| R-UBI-001 | 시스템은 항상 모든 요청에 대한 비용을 추적해야 한다 | 비용 로그 검증 |
| R-UBI-002 | 시스템은 항상 모델 응답에 사용된 모델 정보를 포함해야 한다 | 응답 메타데이터 검증 |
| R-UBI-003 | 시스템은 항상 라우팅 결정 로그를 기록해야 한다 | 로그 파일 분석 |

### Event-Driven (이벤트 기반)

| ID | 트리거 | 동작 | 검증 방법 |
|----|--------|------|-----------|
| R-EVT-001 | WHEN "planning" 작업이 감지되면 | THEN Claude Opus로 라우팅해야 한다 | 라우팅 로그 확인 |
| R-EVT-002 | WHEN "implementation" 작업이 감지되면 | THEN GLM-4.7로 라우팅해야 한다 | 라우팅 로그 확인 |
| R-EVT-003 | WHEN 모델 전환이 발생하면 | THEN 컨텍스트를 압축하여 전달해야 한다 | 컨텍스트 전달 검증 |
| R-EVT-004 | WHEN `/moai:1-plan` 명령이 실행되면 | THEN Claude Opus를 사용해야 한다 | 모델 선택 로그 |
| R-EVT-005 | WHEN `/moai:2-run` 명령이 실행되면 | THEN GLM-4.7을 사용해야 한다 | 모델 선택 로그 |
| R-EVT-006 | WHEN 아키텍처 결정 (5+ 파일 영향)이 필요하면 | THEN Claude Opus를 사용해야 한다 | 분류기 결정 검증 |

### State-Driven (상태 기반)

| ID | 조건 | 동작 | 검증 방법 |
|----|------|------|-----------|
| R-STA-001 | IF 현재 모델이 응답 중이면 | THEN 전환 요청을 대기시켜야 한다 | 동시성 테스트 |
| R-STA-002 | IF API 에러가 발생하면 | THEN Fallback 모델을 사용해야 한다 | 에러 시나리오 테스트 |
| R-STA-003 | IF 일일 비용 한도에 도달하면 | THEN 경고를 발생시키고 계속해야 한다 | 비용 한도 테스트 |

### Unwanted (금지 사항)

| ID | 금지 동작 | 이유 | 검증 방법 |
|----|----------|------|-----------|
| R-UNW-001 | 시스템은 설정되지 않은 모델에 요청을 전송하지 않아야 한다 | 에러 방지 | 미설정 모델 테스트 |
| R-UNW-002 | 시스템은 API 키를 로그에 기록하지 않아야 한다 | 보안 | 로그 검사 자동화 |
| R-UNW-003 | 시스템은 분류 없이 임의로 모델을 선택하지 않아야 한다 | 일관성 | 분류 로직 검증 |

### Optional (선택 사항)

| ID | 요구사항 | 조건 |
|----|----------|------|
| R-OPT-001 | 가능하면 응답 품질 피드백을 수집하여 분류기 개선 | 사용자 피드백 수집 기능 |
| R-OPT-002 | 가능하면 모델별 응답 시간 통계 제공 | 모니터링 대시보드 연동 시 |

---

## Specifications (상세 명세)

### Model Tier 규칙

#### Tier 1 (Claude Opus 4.5) - 10% 토큰

Planning 작업:
- `/moai:1-plan` 실행
- 아키텍처 결정 (5+ 파일 영향)
- 기술 선택 트레이드오프 분석
- 복잡한 설계 패턴 결정

분류 키워드:
- `plan`, `design`, `architecture`, `strategy`
- `trade-off`, `decision`, `spec`, `specification`
- `compare`, `evaluate`, `analyze`

#### Tier 2 (GLM-4.7) - 90% 토큰

Implementation 작업:
- `/moai:2-run` 실행
- 코드 생성 및 수정
- 테스트 작성
- 문서화

분류 키워드:
- `implement`, `code`, `build`, `write`
- `test`, `fix`, `refactor`, `update`
- `document`, `create file`

### 생성할 파일 구조

```
src/moai_adk/web/services/
├── model_router.py         # ModelRouter 클래스
├── task_classifier.py      # 작업 분류 로직
└── cost_tracker.py         # 모델별 비용 추적

src/moai_adk/cli/worktree/
└── model_config.py         # Worktree 모델 설정
```

### 핵심 클래스 설계

#### ModelRouter

```python
class ModelRouter:
    """멀티 모델 라우팅 핵심 클래스"""

    def __init__(
        self,
        classifier: TaskClassifier,
        cost_tracker: CostTracker,
        opus_client: OpusClient,
        glm_client: GLMClient,
    ):
        self.classifier = classifier
        self.cost_tracker = cost_tracker
        self.opus_client = opus_client
        self.glm_client = glm_client

    async def route(
        self,
        message: str,
        context: RouterContext,
    ) -> RouterResponse:
        """작업 분류 후 적절한 모델로 라우팅"""
        ...

    async def switch_model(
        self,
        from_model: str,
        to_model: str,
        context: RouterContext,
    ) -> CompressedContext:
        """모델 전환 시 컨텍스트 압축 및 전달"""
        ...
```

#### TaskClassifier

```python
class TaskClassifier:
    """Rule-based 작업 분류기"""

    TIER1_PATTERNS: ClassVar[list[str]] = [
        r"/moai:1-plan",
        r"(architecture|design|strategy)",
        r"(trade-off|decision)",
        r"(compare|evaluate|analyze)",
    ]

    TIER2_PATTERNS: ClassVar[list[str]] = [
        r"/moai:2-run",
        r"(implement|code|build|write)",
        r"(test|fix|refactor)",
        r"(document|create)",
    ]

    def classify(self, message: str) -> ModelTier:
        """메시지를 분석하여 Tier 결정"""
        ...

    def get_confidence(self, message: str) -> float:
        """분류 신뢰도 반환 (0.0 ~ 1.0)"""
        ...
```

#### CostTracker

```python
class CostTracker:
    """모델별 비용 추적"""

    def __init__(self, db_session: AsyncSession):
        self.db = db_session

    async def track(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
        cost_usd: float,
    ) -> None:
        """비용 기록"""
        ...

    async def get_daily_summary(self) -> CostSummary:
        """일일 비용 요약"""
        ...

    async def get_model_breakdown(self) -> dict[str, ModelCost]:
        """모델별 비용 분석"""
        ...
```

### 데이터 모델

#### RouterContext

```python
class RouterContext(BaseModel):
    """라우터 컨텍스트"""
    session_id: str
    conversation_history: list[Message]
    current_spec_id: str | None = None
    worktree_path: str | None = None
    force_model: str | None = None  # 강제 모델 지정
```

#### RouterResponse

```python
class RouterResponse(BaseModel):
    """라우터 응답"""
    content: str
    model_used: str  # "claude-opus" | "glm-4.7"
    tier: int  # 1 | 2
    input_tokens: int
    output_tokens: int
    cost_usd: float
    classification_confidence: float
```

#### CostSummary

```python
class CostSummary(BaseModel):
    """비용 요약"""
    date: date
    total_cost_usd: float
    tier1_cost_usd: float
    tier2_cost_usd: float
    tier1_percentage: float
    tier2_percentage: float
    total_requests: int
```

### API Endpoints

| Method | Path | 설명 |
|--------|------|------|
| POST | /api/router/route | 메시지 라우팅 |
| GET | /api/router/stats | 라우팅 통계 |
| GET | /api/costs/daily | 일일 비용 조회 |
| GET | /api/costs/breakdown | 모델별 비용 분석 |
| POST | /api/router/force-model | 강제 모델 지정 |

### 환경 변수 설정

```bash
# Tier 1 - Claude Opus
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-opus-4-5-20251101

# Tier 2 - GLM-4.7 via Z.AI
ZAI_API_KEY=your-zai-key
ZAI_BASE_URL=https://api.z.ai/v1
ZAI_MODEL=glm-4.7

# Router Configuration
MODEL_ROUTER_DEFAULT_TIER=2
MODEL_ROUTER_CONFIDENCE_THRESHOLD=0.8
MODEL_ROUTER_DAILY_BUDGET_USD=100.0
```

### Fallback 전략

```
Primary: 분류된 모델 사용
  ↓ API 에러 발생
Fallback 1: 대체 모델 시도 (Tier1 ↔ Tier2)
  ↓ 대체 모델도 실패
Fallback 2: 로컬 캐시된 응답 반환 (가능한 경우)
  ↓ 캐시 없음
Error: 사용자에게 에러 메시지 반환
```

---

## Dependencies (의존성)

### 의존하는 SPEC

- SPEC-WEB-001: MoAI Web Backend (FastAPI 서버 기반)

### 이 SPEC에 의존하는 SPEC

- SPEC-WEB-002: MoAI Web Frontend (모델 선택 UI)
- SPEC-COST-001: Cost Management Dashboard

---

## Traceability (추적성)

### 관련 문서

- `.moai/specs/SPEC-WEB-001/spec.md`: 백엔드 기반
- `CLAUDE.md`: Alfred 실행 지시

### 태그

- `SPEC-MODEL-001`
- `multi-model`
- `router`
- `claude-opus`
- `glm-4.7`
- `cost-optimization`
