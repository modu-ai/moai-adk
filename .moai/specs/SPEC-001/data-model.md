# Data Model: 마법사 UX 시스템 @DATA:WIZARD-UX

> **@DATA:WIZARD-UX** "Claude Code 명령어 기반 데이터 구조 설계"

## 📊 핵심 데이터 구조 설계

### 1. WizardController 데이터 모델

#### WizardSession (.moai/indexes/state.json) - 최신 계약 반영
```json
{
  "wizard_session": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "dev_user_001",
    "command": "/moai:1-project",
    "wizard_version": "1.0.0",
    "current_step": 3,
    "total_steps": 10,
    "status": "ACTIVE",
    "started_at": "2025-01-20T14:30:22Z",
    "last_activity": "2025-01-20T14:35:15Z",
    "completed_at": null,

    "answers": {
      "Q1_problem_definition": "개발자들이 코드 리뷰 시 일관성 없는 품질로 인해 반복적인 지적 발생",
      "Q2_target_users": ["시니어 개발자", "팀 리더", "주니어 개발자"],
      "Q3_success_metrics": {
        "p95_response_time_ms": 300,
        "error_rate_percent": 1,
        "test_coverage_percent": 85
      }
    },

    "step_states": [
      {
        "step_number": 1,
        "question_id": "Q1_PROBLEM",
        "status": "COMPLETED",
        "started_at": "2025-01-20T14:30:25Z",
        "completed_at": "2025-01-20T14:32:10Z",
        "duration_ms": 105000,
        "retry_count": 0,
        "metadata": {
          "keywords_detected": ["코드", "리뷰", "품질"],
          "user_agent": "Claude Code 1.0.0"
        }
      }
    ],

    "dynamic_questions": ["AI_ML_DETECTED", "PERFORMANCE_CRITICAL"],
    "project_metadata": {
      "claude_code_version": "1.0.0",
      "platform": "macOS",
      "project_directory": "/Users/dev/my-project",
      "environment_variables": {
        "NODE_ENV": "development",
        "PYTHON_VERSION": "3.11.0"
      }
    }
  }
}
```

#### WizardStep (명령어 내장 구조)
```markdown
## 10단계 질문 체계
1. **Q1 문제 정의**: 핵심 해결 문제 (30자 이상, 대상/원인/빈도 포함)
2. **Q2 목표 사용자**: 역할별 사용자 그룹 (복수 선택 가능)
3. **Q3 성공 지표**: 6개월 후 KPI (측정 가능한 수치 필수)
4. **Q4 핵심 기능**: Top-3 우선순위 기능 (1→2→3 순서)
5. **Q5 화면 구성**: 주요 페이지/화면 구조
6. **Q6 UI/UX**: 컴포넌트 및 디자인 토큰
7. **Q7 기능 트리**: 3레벨 계층 구조 (@REQ 자동 분류)
8. **Q8 기술 스택**: 웹/모바일/백엔드/DB 선택
9. **Q9 팀 숙련도**: 기술 복잡도 조정 기준
10. **Q10 품질 목표**: 테스트 커버리지, 성능 목표

## 동적 분기 로직
- **AI/ML 키워드** → 데이터·모델·추론 경로 추가 질문
- **보안/PII** → 보관기간·삭제·암호화 추가 질문
- **성능/실시간** → 수치 목표 재확인
```

#### SessionState (Pydantic V2 모델)
```python
from pydantic import BaseModel, Field, UUID4
from datetime import datetime
from typing import Optional, List, Dict, Any, Literal
from enum import Enum

class SessionStatus(str, Enum):
    """세션 상태 열거형"""
    INITIALIZING = "INITIALIZING"
    ACTIVE = "ACTIVE"
    PAUSED = "PAUSED"
    COMPLETING = "COMPLETING"
    COMPLETED = "COMPLETED"
    ABANDONED = "ABANDONED"
    ERROR = "ERROR"

class StepStatus(str, Enum):
    """단계 상태 열거형"""
    NOT_STARTED = "NOT_STARTED"
    IN_PROGRESS = "IN_PROGRESS"
    VALIDATING = "VALIDATING"
    COMPLETED = "COMPLETED"
    ERROR = "ERROR"
    SKIPPED = "SKIPPED"

class StepState(BaseModel):
    """개별 단계 상태"""
    step_number: int = Field(ge=1, le=20)
    question_id: str = Field(min_length=1, max_length=50)
    status: StepStatus = StepStatus.NOT_STARTED
    user_answer: Optional[Any] = None
    validation_result: Optional['ValidationResult'] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    duration_ms: Optional[int] = Field(default=None, ge=0)
    retry_count: int = Field(default=0, ge=0)
    metadata: Dict[str, Any] = Field(default_factory=dict)

class WizardSession(BaseModel):
    """마법사 세션 상태"""
    session_id: UUID4
    user_id: str = Field(min_length=1, max_length=100)
    command: str = "/moai:1-project"
    wizard_version: str = Field(default="1.0.0")

    current_step: int = Field(default=0, ge=0, le=20)
    total_steps: int = Field(default=10, ge=1, le=20)
    status: SessionStatus = SessionStatus.INITIALIZING

    started_at: datetime = Field(default_factory=datetime.utcnow)
    last_activity: datetime = Field(default_factory=datetime.utcnow)
    completed_at: Optional[datetime] = None

    answers: Dict[str, Any] = Field(default_factory=dict)
    step_states: List[StepState] = Field(default_factory=list)
    dynamic_questions: List[str] = Field(default_factory=list)

    project_metadata: Dict[str, Any] = Field(default_factory=dict)

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat(),
            UUID4: lambda v: str(v)
        }
        use_enum_values = True

class ValidationResult(BaseModel):
    """입력 검증 결과"""
    is_valid: bool
    error_message: Optional[str] = None
    error_code: Optional[str] = None
    suggestions: List[str] = Field(default_factory=list)

    # 성능 메트릭
    validation_time_ms: float = Field(ge=0)
    retry_count: int = Field(default=0, ge=0)

    # 보안 검사
    security_flags: List[str] = Field(default_factory=list)
    sanitized_input: Optional[str] = None

    # 추가된 필드들
    confidence_score: float = Field(default=1.0, ge=0.0, le=1.0)
    validation_rules_applied: List[str] = Field(default_factory=list)
```

### 2. OutputRenderer 데이터 모델

#### ProgressDisplay (마크다운 출력 템플릿)
```markdown
## 진행 상태 표시 형식

### 단계별 진행바
🗿 MoAI-ADK 프로젝트 초기화 마법사

[3/10] 🎯 성공 지표 설정
███████░░░░░░░░░░░ 30% 완료

✅ 완료: Q1 문제 정의, Q2 목표 사용자
🔄 진행중: Q3 성공 지표 설정
⏳ 대기: Q4~Q10

### 질문 표시 형식
**Q3. 6개월 후 달성하고 싶은 구체적인 목표는?**

💡 **예시 답변:**
- p95 응답시간 300ms 이하
- 에러율 1% 미만
- 테스트 커버리지 85% 이상

📝 **입력 가이드:**
측정 가능한 KPI를 포함해주세요 (응답시간, 에러율, 커버리지 등)

> [사용자 입력 대기]

### 에러 메시지 형식
❌ **입력이 불완전합니다**

문제: 성공 지표가 측정 불가능합니다
입력: "사용자가 만족하는 서비스"

💡 **개선 제안:**
- p95 응답시간: 300ms 이하
- 사용자 만족도: 4.5/5.0 이상
- 월간 활성 사용자: 1,000명 이상

다시 입력해주세요.
```

#### DisplayPanel
```python
class DisplayPanel(BaseModel):
    """화면 표시 패널"""

    panel_type: Literal["progress", "question", "summary", "error"]
    title: str = Field(min_length=1, max_length=100)
    content: str = Field(min_length=1, max_length=2000)

    # 스타일링
    border_color: Optional[str] = None
    background_color: Optional[str] = None
    text_color: Optional[str] = None

    # 상호작용
    interactive: bool = False
    actions: List[str] = Field(default_factory=list)

    # 메타데이터
    created_at: datetime = Field(default_factory=datetime.utcnow)
    render_time_ms: Optional[float] = None
```

### 3. AgentOrchestrator 데이터 모델

#### TaskCallPattern (에이전트 호출 구조)
```markdown
## steering-architect 호출
Task 도구를 사용하여 steering-architect 에이전트 호출:

### 입력 데이터
- 수집된 10단계 답변 정보
- 동적 질문 결과 (AI/ML, 보안, 성능 관련)
- 프로젝트 환경 정보 (디렉토리, 플랫폼)

### 생성 결과
- .moai/steering/product.md (제품 비전과 목표)
- .moai/steering/structure.md (코드 구조 원칙)
- .moai/steering/tech.md (기술 스택 결정)
- .moai/config.json (MoAI 설정 및 Constitution)

## spec-manager 호출
Top-3 기능에 대한 SPEC 시드 생성:

### 입력 데이터
- Q4에서 수집된 핵심 기능 3가지
- 전체 프로젝트 맥락 정보

### 생성 결과
- .moai/specs/SPEC-001/spec.md (1순위 기능 명세)
- .moai/specs/SPEC-002/spec.md (2순위 기능 명세)
- .moai/specs/SPEC-003/spec.md (3순위 기능 명세)
- [NEEDS CLARIFICATION] 마커 자동 삽입

## tag-indexer 호출
16-Core TAG 시스템 초기화:

### 생성 결과
- .moai/indexes/tags.json (TAG 인덱스 구조)
- @VISION, @STRUCT, @TECH 태그 생성
- 추적성 매트릭스 기본 구조
```

#### KeywordDetection (동적 분기 감지)
    """키워드 감지 결과"""

    detected_keywords: List[str] = Field(default_factory=list)
    confidence_scores: Dict[str, float] = Field(default_factory=dict)
    suggested_questions: List[str] = Field(default_factory=list)

    # AI/ML 관련
    ai_ml_detected: bool = False
    framework_suggestions: List[str] = Field(default_factory=list)

    # 보안 관련
    security_context: bool = False
    compliance_requirements: List[str] = Field(default_factory=list)

    # 성능 관련
    performance_critical: bool = False
    scalability_concerns: List[str] = Field(default_factory=list)
```

#### DynamicQuestionRule
```python
class DynamicQuestionRule(BaseModel):
    """동적 질문 생성 규칙"""

    rule_id: str = Field(min_length=1, max_length=50)
    trigger_pattern: str = Field(min_length=1, max_length=100)
    priority: int = Field(default=1, ge=1, le=10)

    # 조건부 로직
    conditions: Dict[str, Any] = Field(default_factory=dict)
    generated_questions: List[Question] = Field(default_factory=list)

    # 메타데이터
    created_by: str = "system"
    last_used: Optional[datetime] = None
    usage_count: int = 0
    effectiveness_score: float = Field(default=0.0, ge=0.0, le=1.0)
```

## 🔄 데이터 흐름 설계

### 1. 입력 검증 파이프라인
```python
class InputValidationPipeline(BaseModel):
    """입력 데이터 검증 파이프라인"""

    # 단계별 검증
    stage_1_sanitization: bool = True    # 입력 정제
    stage_2_type_validation: bool = True # 타입 검증
    stage_3_business_rules: bool = True  # 비즈니스 규칙
    stage_4_security_scan: bool = True   # 보안 검사

    # 결과 로깅
    validation_log: List[ValidationResult] = Field(default_factory=list)
    total_validation_time_ms: float = 0.0

    def validate_pipeline(self, input_data: str, step: WizardStep) -> ValidationResult:
        """전체 검증 파이프라인 실행"""
        pass
```

### 2. 상태 변환 모델
```python
class StateTransition(BaseModel):
    """상태 전환 추적"""

    from_step: int = Field(ge=0, le=10)
    to_step: int = Field(ge=0, le=10)
    transition_type: Literal["next", "back", "skip", "restart"] = "next"

    # 전환 조건
    conditions_met: List[str] = Field(default_factory=list)
    user_action: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)

    # 메트릭
    duration_ms: float
    success: bool = True
    error_message: Optional[str] = None
```

### 3. 프로젝트 설정 출력 모델
```python
class ProjectConfig(BaseModel):
    """최종 프로젝트 설정"""

    # 기본 정보
    project_name: str = Field(min_length=1, max_length=100)
    description: str = Field(min_length=10, max_length=500)
    author: str = Field(min_length=1, max_length=100)

    # 기술 스택
    language: str = Field(min_length=1, max_length=50)
    framework: Optional[str] = None
    database: Optional[str] = None
    deployment_target: Optional[str] = None

    # 프로젝트 속성
    project_type: Literal["web", "mobile", "desktop", "api", "library", "other"]
    team_size: int = Field(ge=1, le=50)
    timeline_months: int = Field(ge=1, le=24)

    # 동적 설정 (키워드 기반)
    ai_ml_features: Optional[Dict[str, Any]] = None
    security_requirements: Optional[Dict[str, Any]] = None
    performance_targets: Optional[Dict[str, Any]] = None

    # 메타데이터
    config_version: str = "1.0.0"
    generated_at: datetime = Field(default_factory=datetime.utcnow)
    wizard_session_id: UUID

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat(),
            UUID: lambda v: str(v)
        }
```

## 🔒 보안 및 검증 모델

### InputSanitization
```python
class InputSanitization(BaseModel):
    """입력 정제 및 보안 검사"""

    # 기본 정제
    strip_whitespace: bool = True
    normalize_unicode: bool = True
    remove_control_chars: bool = True

    # 보안 필터
    html_escape: bool = True
    sql_injection_check: bool = True
    xss_prevention: bool = True
    command_injection_check: bool = True

    # 허용 목록
    allowed_characters: str = r"[a-zA-Z0-9\s\-_\.@#]"
    max_input_length: int = 1000
    blocked_patterns: List[str] = Field(default_factory=list)

    def sanitize(self, raw_input: str) -> Tuple[str, List[str]]:
        """입력 정제 및 보안 플래그 반환"""
        pass
```

## 📈 성능 및 모니터링 모델

### PerformanceMetrics
```python
class PerformanceMetrics(BaseModel):
    """성능 지표 수집 - 최신 계약 반영"""

    # 응답시간 메트릭 (NFR-1 목표: < 500ms)
    step_render_time_ms: float = Field(ge=0)
    validation_time_ms: float = Field(ge=0)
    state_transition_time_ms: float = Field(ge=0)
    total_response_time_ms: float = Field(ge=0)

    # 단계별 세부 타이밍
    question_generation_time_ms: float = Field(ge=0)
    ui_rendering_time_ms: float = Field(ge=0)
    file_io_time_ms: float = Field(ge=0)
    keyword_detection_time_ms: float = Field(ge=0)

    # 메모리 사용량 (NFR-1 목표: < 100MB)
    memory_usage_mb: float = Field(ge=0)
    peak_memory_mb: float = Field(ge=0)
    memory_usage_by_component: Dict[str, float] = Field(default_factory=dict)

    # 사용자 경험 메트릭
    user_wait_time_ms: float = Field(ge=0)
    user_think_time_ms: Optional[float] = Field(default=None, ge=0)
    abandonment_risk_score: float = Field(default=0.0, ge=0.0, le=1.0)
    step_retry_count: int = Field(default=0, ge=0)

    # 시스템 자원 사용량
    cpu_usage_percent: float = Field(ge=0.0, le=100.0)
    disk_usage_mb: float = Field(ge=0)
    io_operations: int = Field(default=0, ge=0)
    network_requests: int = Field(default=0, ge=0)

    # 캐시 효율성
    cache_hit_ratio: float = Field(default=0.0, ge=0.0, le=1.0)
    cache_memory_mb: float = Field(default=0.0, ge=0)

    # 타임스탬프
    measurement_timestamp: datetime = Field(default_factory=datetime.utcnow)
    session_id: Optional[UUID4] = None
    step_number: Optional[int] = Field(default=None, ge=0, le=20)

    def is_performance_acceptable(self) -> bool:
        """성능 목표 달성 여부 확인 (Constitution & NFR 기준)"""
        return (
            self.total_response_time_ms < 500 and          # NFR-1: 응답시간
            self.memory_usage_mb < 100 and                 # NFR-1: 메모리
            self.cpu_usage_percent < 20 and                # 시스템 부하 고려
            self.abandonment_risk_score < 0.3              # 사용성 목표
        )

    def get_performance_grade(self) -> Literal["A", "B", "C", "D", "F"]:
        """성능 등급 계산"""
        score = 0
        if self.total_response_time_ms < 200: score += 25
        elif self.total_response_time_ms < 500: score += 15

        if self.memory_usage_mb < 50: score += 25
        elif self.memory_usage_mb < 100: score += 15

        if self.cpu_usage_percent < 10: score += 25
        elif self.cpu_usage_percent < 20: score += 15

        if self.abandonment_risk_score < 0.1: score += 25
        elif self.abandonment_risk_score < 0.3: score += 15

        if score >= 90: return "A"
        elif score >= 80: return "B"
        elif score >= 70: return "C"
        elif score >= 60: return "D"
        else: return "F"

class SecurityMetrics(BaseModel):
    """보안 지표 및 감사 추적"""

    # 입력 보안 검사
    input_sanitization_checks: int = Field(default=0, ge=0)
    potential_threats_detected: List[str] = Field(default_factory=list)
    sanitization_time_ms: float = Field(default=0, ge=0)

    # 파일 시스템 보안
    file_access_violations: int = Field(default=0, ge=0)
    path_traversal_attempts: int = Field(default=0, ge=0)
    unauthorized_access_attempts: int = Field(default=0, ge=0)

    # 데이터 보호
    sensitive_data_redactions: int = Field(default=0, ge=0)
    encryption_operations: int = Field(default=0, ge=0)
    data_validation_failures: int = Field(default=0, ge=0)

    # 세션 보안
    session_hijacking_attempts: int = Field(default=0, ge=0)
    invalid_session_access: int = Field(default=0, ge=0)
    session_timeout_enforcements: int = Field(default=0, ge=0)

    # 감사 추적
    security_events_logged: int = Field(default=0, ge=0)
    compliance_violations: List[str] = Field(default_factory=list)
    audit_log_size_mb: float = Field(default=0, ge=0)

    # 타임스탬프 및 식별자
    measurement_timestamp: datetime = Field(default_factory=datetime.utcnow)
    session_id: Optional[UUID4] = None
    security_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"] = "MEDIUM"

    def get_security_score(self) -> float:
        """보안 점수 계산 (0.0 = 위험, 1.0 = 안전)"""
        threats = len(self.potential_threats_detected)
        violations = self.file_access_violations + self.unauthorized_access_attempts

        if threats == 0 and violations == 0:
            return 1.0
        elif threats <= 2 and violations <= 1:
            return 0.8
        elif threats <= 5 and violations <= 3:
            return 0.6
        else:
            return 0.3
```

### UsageAnalytics
```python
class UsageAnalytics(BaseModel):
    """사용 패턴 분석"""

    # 완료율 메트릭
    completion_rate: float = Field(ge=0.0, le=1.0)
    average_completion_time_min: float
    most_common_abandonment_step: int

    # 에러 분석
    validation_error_rate: float = Field(ge=0.0, le=1.0)
    common_error_types: Dict[str, int] = Field(default_factory=dict)
    user_retry_patterns: List[int] = Field(default_factory=list)

    # 만족도 지표
    user_satisfaction_score: float = Field(ge=0.0, le=5.0)
    feature_usage_stats: Dict[str, int] = Field(default_factory=dict)

    # 시간대별 분석
    peak_usage_hours: List[int] = Field(default_factory=list)
    session_duration_distribution: Dict[str, int] = Field(default_factory=dict)
```

---

## 🗄️ 데이터 저장 전략

### 1. 세션 상태 저장
```python
# ~/.moai/wizard_sessions/
{session_id}.json  # 진행 중인 세션
{session_id}_completed.json  # 완료된 세션
{session_id}_abandoned.json  # 중단된 세션
```

### 2. 설정 템플릿 저장
```python
# ~/.moai/project_templates/
template_001.json  # 웹 애플리케이션 템플릿
template_002.json  # AI/ML 프로젝트 템플릿
template_003.json  # 모바일 앱 템플릿
```

### 3. 사용자 선호도 저장
```python
# ~/.moai/user_preferences/
defaults.json      # 기본 설정
shortcuts.json     # 자주 사용하는 설정
history.json       # 이전 프로젝트 기록
```

---

## 🔗 연관 태그 시스템

**@DATA:WIZARD-UX**와 연결된 주요 태그들:
- **@REQ:WIZARD-UX-001** → 요구사항 추적
- **@DESIGN:WIZARD-UX** → 설계 문서 연결
- **@TASK:DATA-MODEL** → 구현 작업 연결
- **@TEST:DATA-VALIDATION** → 테스트 케이스 연결

---

> **@DATA:WIZARD-UX** 를 통해 이 데이터 모델이 전체 시스템에서 완벽하게 추적됩니다.
>
> **Pydantic V2 기반의 강력한 타입 안전성과 런타임 검증을 보장합니다.**