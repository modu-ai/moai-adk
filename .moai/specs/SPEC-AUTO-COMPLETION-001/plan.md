---
  "id": "PLAN-AUTO-COMPLETION-001",
  "spec_id": "SPEC-AUTO-COMPLETION-001",
  "title": "자동 SPEC 생성 완성화 시스템 구현 계획",
  "title_en": "Automated SPEC Completion System Implementation Plan",
  "version": "1.0.0",
  "status": "draft",
  "created": "2025-11-11",
  "author": "@user"
}
---

## Implementation Plan

### 구현 단계 (Implementation Phases)

#### 1단계: 핵심 Hook 시스템 구현 (Primary Goal)

**1.1 PostToolUse Hook 기반 구축**
- `post_tool__auto_spec_completion.py` 파일 생성
- PreToolUse 이벤트 분석 로직 재사용
- 툴 실행 결과 기반의 후처리 로직 구현

**1.2 파일 변경 감지 로직**
```python
def detect_code_changes(tool_name: str, tool_args: Dict, result: Any) -> List[str]:
    """
    실제 변경된 코드 파일 목록 추출
    - Write 신규 파일 생성 감지
    - Edit 파일 수정 감지
    - MultiEdit 다중 파일 변경 감지
    """
```

**1.3 SPEC 미존재 확인**
- `.moai/specs/` 디렉토리 스캔
- 파일 경로 기반 SPEC ID 매칭
- 중복 생성 방지 로직

#### 2단계: 지능형 분석 엔진 (Primary Goal)

**2.1 코드 구조 분석 강화**
- 기존 `spec_generator.py` 활용
- 함수/클래스/임포트 분석
- 도메인 키워드 추출 개선

**2.2 Confidence Scoring 시스템**
```python
def calculate_completion_confidence(analysis: CodeAnalysis) -> float:
    """
    SPEC 완성도 신뢰도 점수 계산
    - 구조 명확성 (30%)
    - 도메인 추론 정확도 (40%)
    - 문서화 수준 (30%)
    """
```

**2.3 자동 SPEC ID 생성**
- 파일 경로 기반 ID 추론
- 중복 방지 넘버링
- 의미론적 ID 생성

#### 3단계: 완전한 SPEC 문서 생성 (Primary Goal)

**3.1 EARS 형식 템플릿 엔진**
```python
def generate_ears_spec(analysis: CodeAnalysis, file_path: str) -> Dict[str, str]:
    """
    완전한 EARS 형식 SPEC 생성
    - 자동 환경 설정 추론
    - 가정사항 자동 도출
    - 요구사항 자동 분류
    - 명세사항 상세화
    """
```

**3.2 3종 파일 동시 생성**
- MultiEdit 활용한 효율적 파일 생성
- `spec.md`: 핵심 명세서
- `plan.md`: 구현 계획서
- `acceptance.md`: 검수 기준서

**3.3 품질 보증 시스템**
```python
def validate_generated_spec(spec_content: Dict) -> QualityReport:
    """
    생성된 SPEC 품질 검증
    - EARS 형식 준수 여부
    - 내용 completeness 검사
    - 편집 가이드 생성
    """
```

#### 4단계: 사용자 경험 최적화 (Secondary Goal)

**4.1 실시간 알림 시스템**
- 성공/실패 명확한 피드백
- 진행 상태 표시
- 편집 제안 상세 표시

**4.2 자동 에디터 열기**
```python
def auto_open_editor(spec_path: str, confidence: float) -> bool:
    """
    생성된 SPEC 자동 열기
    - confidence 높으면 바로 열기
    - editor 설정 기반 실행
    ```

**4.3 편집 가이드 통합**
- TODO 체크리스트 자동 생성
- 추천 편집 항목 제안
- 품질 개선 가이드

#### 5단계: 시스템 통합 및 최적화 (Final Goal)

**5.1 Configuration 시스템 연동**
```json
{
  "tags": {
    "policy": {
      "auto_spec_completion": {
        "enabled": true,
        "min_confidence": 0.7,
        "auto_open_editor": true,
        "supported_languages": ["python", "javascript", "typescript", "go"],
        "excluded_patterns": ["test_", "spec_"]
      }
    }
  }
}
```

**5.2 성능 최적화**
- 캐싱 전략 구현
- 비동기 처리 지원
- 메모리 사용 최적화

**5.3 에러 핸들링**
- 그레이스풀 데그레이이션
- 상세 에러 로깅
- 사용자 친화적 에러 메시지

### 기술적 접근 방식 (Technical Approach)

#### 아키텍처 설계

```
PostToolUse Event
    ↓
post_tool__auto_spec_completion.py
    ↓
File Change Detection ← Confidence Scoring
    ↓                      ↓
Code Analysis ← → Domain Inference
    ↓
EARS Template Engine
    ↓
MultiEdit (3 files created)
    ↓
Quality Validation → User Notification
    ↓
Auto Editor Open (optional)
```

#### 핵심 컴포넌트

1. **Hook Controller**: `post_tool__auto_spec_completion.py`
   - 이벤트 수신 및 처리
   - 조건 분기 로직
   - 결과 반환

2. **Analysis Engine**: 기존 `spec_generator.py` 확장
   - 코드 구조 분석
   - 도메인 추론
   - 신뢰도 계산

3. **Template Engine**: 신규 구현
   - EARS 형식 템플릿
   - 동적 내용 채우기
   - 품질 검증

4. **File Manager**: MultiEdit 기반
   - 3종 파일 동시 생성
   - 디렉토리 구조 관리
   - Git 준비

#### 의존성 관리

**기존 모듈 활용:**
- `moai_adk.core.tags.spec_generator.SpecGenerator`
- `moai_adk.core.tags.policy_validator.TagPolicyValidator`
- `moai_adk.core.tags.auto_corrector.AutoCorrection`

**신규 모듈 추가:**
- `moai_adk.core.hooks.post_tool_auto_spec_completion`
- `moai_adk.core.spec.ears_template_engine`
- `moai_adk.core.spec.quality_validator`

### 위험 및 대응 계획 (Risks & Mitigation)

#### 기술적 리스크

**1. 성능 저하 위험**
- *위험*: Hook 실행 시간 초과 (2초 제한)
- *대응*: 비동기 처리, confidence 기필터링, 캐싱

**2. 의존성 충돌**
- *위험*: 기존 모듈과의 호환성 문제
- *대응*: 하위 호환성 유지, 점진적 업그레이드

**3. 파일 시스템 오류**
- *위험*: 디렉토리 생성 실패, 권한 문제
- *대응*: 예외 처리, 롤백 메커니즘

#### 사용자 경험 리스크

**1. 과도한 자동화**
- *위험*: 사용자 제어권 상실
- *대응*: 설정 기반 제어, opt-out 옵션

**2. 품질 저하**
- *위험*: 저품질 SPEC 자동 생성
- *대응*: confidence 기반 필터링, 품질 검증

### 성공 기준 (Success Criteria)

#### 기능적 기준

- ✅ 코드 파일 생성/수정 시 100% SPEC 생성 시도
- ✅ 생성된 SPEC의 EARS 형식 준수율 95% 이상
- ✅ 중복 생성 방지율 100%
- ✅ 사용자 알림 전달률 100%

#### 성능 기준

- ✅ Hook 실행 시간 1.5초 이내 (2초 제한 대비 25% 마진)
- ✅ 메모리 사용량 50MB 이내
- ✅ 캐시 적중률 70% 이상

#### 사용자 경험 기준

- ✅ 자동 생성 활성화 시 사용자 만족도 80% 이상
- ✅ 생성된 SPEC의 즉시 사용 가능성 85% 이상
- ✅ 편집 가이드 유용성 평점 4/5 이상

### 다음 단계 (Next Steps)

1. **즉시 실행**: Hook 기본 구조 구현 (1-2일)
2. **주간 목표**: 완전한 기능 구현 (5-7일)
3. **2주 목표**: 통합 테스트 및 최적화 (10-14일)
4. **배포 준비**: 문서화 및 설정 안내 (14-16일)