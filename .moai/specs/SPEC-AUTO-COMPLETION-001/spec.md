---
@META: {
  "id": "SPEC-AUTO-COMPLETION-001",
  "title": "자동 SPEC 생성 완성화 시스템",
  "title_en": "Automated SPEC Completion System",
  "version": "1.0.0",
  "status": "draft",
  "created": "2025-11-11",
  "author": "@user",
  "reviewer": "",
  "category": "FEATURE",
  "priority": "HIGH",
  "tags": ["automation", "spec", "hooks", "completion"],
  "language": "ko",
  "estimated_complexity": "medium"
}
---

# @SPEC:AUTO-COMPLETION-001: 자동 SPEC 생성 완성화 시스템
## Automated SPEC Completion System

### 개요 (Overview)

MoAI-ADK의 현재 자동 SPEC 생성 기능을 완전히 자동화하는 시스템을 구현합니다. 기존의 `pre_tool__auto_spec_proposal.py`가 단순 제안에 그치는 것을 넘어, `post_tool__auto_spec_completion.py`를 통해 실제 완성된 SPEC 문서를 자동으로 생성하여 사용자 경험을 극대화합니다.

### 환경 (Environment)

- **프로젝트**: MoAI-ADK v0.23.0
- **언어**: Python 3.13+
- **모드**: Personal Mode (Team Mode 지원)
- **의존성**:
  - `spec_generator.py` (기존)
  - `auto_corrector.py` (기존)
  - `policy_validator.py` (기존)
- **Hooks 시스템**: PostToolUse 이벤트 기반

### 가정 (Assumptions)

1. 사용자가 코드를 작성할 때 SPEC을 먼저 작성하는 것을 잊어버리는 경우가 많음
2. 기존의 단순 제안 방식은 사용자에게 추가적인 작업 부담을 줌
3. 자동화된 SPEC 생성은 개발 생산성을 크게 향상시킬 수 있음
4. 사용자는 생성된 SPEC을 검토하고 수정할 수 있어야 함
5. 완전 자동화지만 사용자 제어권은 보장되어야 함

### 요구사항 (Requirements)

#### 보편적 요구사항 (Ubiquitous Requirements)

- **REQ-001**: 시스템은 코드 파일 생성/수정 시 자동으로 SPEC 문서를 생성해야 함
- **REQ-002**: 생성된 SPEC은 EARS 형식을 따라야 함
- **REQ-003**: 사용자는 생성된 SPEC을 즉시 검토하고 수정할 수 있어야 함
- **REQ-004**: 시스템은 기존의 SPEC이 있는지 확인하고 중복 생성을 방지해야 함
- **REQ-005**: 자동 생성은 사용자 설정에 따라 활성화/비활성화될 수 있어야 함

#### 상태 기반 요구사항 (State-driven Requirements)

- **REQ-006**: 코드 파일이 생성되었지만 해당 SPEC이 없는 상태에서 자동 생성을触发해야 함
- **REQ-007**: 코드 파일이 수정되었지만 SPEC이 없는 상태에서도 생성을触发해야 함
- **REQ-008**: 일정 confidence score 이상일 때만 자동 생성을 실행해야 함
- **REQ-009**: 생성된 SPEC은 `pending` 상태로 시작해야 함

#### 이벤트 기반 요구사항 (Event-driven Requirements)

- **REQ-010**: `Write` 툴 실행 후 코드 파일 생성 이벤트를 감지해야 함
- **REQ-011**: `Edit` 툴 실행 후 코드 파일 수정 이벤트를 감지해야 함
- **REQ-012**: `MultiEdit` 툴 실행 후 여러 파일 변경 이벤트를 처리해야 함
- **REQ-013**: SPEC 생성 완료 시 사용자에게 알림을 보내야 함
- **REQ-014**: 생성 실패 시 에러 로그를 기록하고 사용자에게 알려야 함

#### 선택적 요구사항 (Optional Requirements)

- **REQ-015**: 생성된 SPEC의 품질을 자동으로 검증하고 점수를 매길 수 있음
- **REQ-016**: 사용자의 피드백을 기반으로 SPEC 생성 품질을 개선할 수 있음
- **REQ-017**: 템플릿 기반의 다국어 SPEC 생성을 지원할 수 있음
- **REQ-018**: Git 커밋 메시지를 기반으로 컨텍스트를 추론할 수 있음

#### 바람직하지 않은 요구사항 (Unwanted Requirements)

- **REQ-019**: 사용자 동의 없이 기존 SPEC을 덮어쓰지 않아야 함
- **REQ-020**: 테스트 파일에 대한 SPEC을 생성하지 않아야 함
- **REQ-021**: 설정 파일이나 문서 파일에 대한 SPEC을 생성하지 않아야 함
- **REQ-022**: 시스템 성능에 큰 영향을 주어서는 안 됨

### 명세 (Specifications)

#### 1. PostToolUse Hook 구현

**Trigger Events:**
```python
# post_tool__auto_spec_completion.py
def should_trigger_spec_completion(tool_name: str, tool_args: Dict[str, Any]) -> bool:
    """
    SPEC 자동 생성을触发할 조건 확인
    - Write/Edit/MultiEdit 툴 실행
    - 코드 파일 대상 (.py, .js, .ts, .jsx, .tsx, .go)
    - 테스트 파일 제외
    - SPEC 미존재 확인
    """
```

**File Analysis:**
```python
def analyze_code_file(file_path: str) -> CodeAnalysis:
    """
    코드 파일 분석 및 SPEC 생성 정보 추출
    - AST 기반 구조 분석
    - 도메인 키워드 추출
    - confidence score 계산
    - 추천 SPEC ID 생성
    """
```

#### 2. SPEC 자동 생성 로직

**Template Generation:**
```python
def generate_complete_spec(analysis: CodeAnalysis) -> SpecDocument:
    """
    완전한 SPEC 문서 생성
    - EARS 형식 구조
    - 환경, 가정, 요구사항, 명세 섹션
    - 자동 추론된 내용 채우기
    - 편집 가이드 포함
    """
```

**File Creation:**
```python
def create_spec_files(spec_id: str, content: Dict[str, str]) -> bool:
    """
    SPEC 파일 3종 생성 (spec.md, plan.md, acceptance.md)
    - .moai/specs/SPEC-{ID}/ 디렉토리 생성
    - 템플릿 기반 내용 채우기
    - Git tracking 준비
    """
```

#### 3. 사용자 인터페이스

**Success Notification:**
```
✅ 자동 SPEC 생성 완료
📁 위치: .moai/specs/SPEC-AUTO-001/
📊 신뢰도: 85% (높음)
📝 편집 제안: 3가지 항목 검토 권장
```

**Quality Report:**
```
📋 SPEC 품질 리포트
🟢 구조: 완벽 (100%)
🟡 내용: 양호 (75%) - 사용자 검토 필요
🔧 추천 편집: 도메인 전문 용어 추가, 요구사항 구체화
```

#### 4. Configuration 통합

**config.json 확장:**
```json
{
  "tags": {
    "policy": {
      "auto_spec_completion": {
        "enabled": true,
        "min_confidence": 0.7,
        "auto_open_editor": true,
        "require_user_review": true,
        "supported_languages": ["python", "javascript", "typescript", "go"],
        "excluded_patterns": ["test_", "spec_", "__tests__"]
      }
    }
  }
}
```

#### 5. 성능 최적화

**Caching Strategy:**
- 파일 분석 결과 캐싱 (TTL: 1시간)
- 도메인 추론 결과 재사용
- 중복 SPEC 생성 방지

**Async Processing:**
- 백그라운드 SPEC 생성 (2초 타임아웃 내)
- 진행 상태 표시
- 부하 분산

### 추적성 (Traceability)

- **@SPEC:AUTO-COMPLETION-001** ← **@SPEC:TAG-SPEC-GENERATION-001** (기존 spec_generator.py 확장)
- **@SPEC:AUTO-COMPLETION-001** ← **@SPEC:TAG-AUTO-SPEC-PROPOSAL-001** (pre_tool hook 확장)
- **@SPEC:AUTO-COMPLETION-001** → **@TEST:AUTO-COMPLETION-001** (테스트)
- **@SPEC:AUTO-COMPLETION-001** → **@CODE:HOOK-POST-AUTO-SPEC-001** (구현)

### SUMMARY (English Summary)

The Automated SPEC Completion System enhances MoAI-ADK's current SPEC generation capabilities by implementing a fully automated workflow that creates complete SPEC documents when code files are created or modified. This system extends the existing `pre_tool__auto_spec_proposal.py` (which only suggests SPEC creation) with a new `post_tool__auto_spec_completion.py` hook that actually generates complete SPEC documents in EARS format.

Key features include intelligent code analysis, confidence scoring, automatic file creation (spec.md, plan.md, acceptance.md), user notifications, and seamless integration with the existing configuration system. The system maintains user control while significantly improving development productivity through intelligent automation, with built-in quality assessment and editing guidance to ensure generated SPECs are immediately useful.

### HISTORY

**v1.0.0** (2025-11-11): Initial SPEC draft for automated SPEC completion system