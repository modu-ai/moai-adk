# SPEC-011: @TAG 추적성 체계 강화

## @REQ:TAG-SYSTEM-011 Environment & Context

### 현재 환경
- **MoAI-ADK 프로젝트**: Spec-First TDD 개발 키트
- **전체 Python 파일**: 100개 (src/ 디렉토리 기준)
- **@TAG 적용 현황**: 88개 파일 (88% 완료)
- **16-Core TAG 시스템**: SPEC/PROJECT/IMPLEMENTATION/QUALITY 4개 카테고리

### 기존 성과
- **Core 모듈**: 48/48 파일 (100% 완료)
- **CLI 모듈**: 6/7 파일 (85.7% 완료)
- **Install/Utils 모듈**: 100% 완료
- **자동화 시스템**: TAG 검증 스크립트 일부 구현

## @DESIGN:TAG-AUDIT-011 Assumptions & Constraints

### 기본 가정
1. **16-Core TAG 체계**: @REQ → @DESIGN → @TASK → @TEST Primary Chain 유지
2. **추적성 원칙**: 모든 코드는 요구사항부터 테스트까지 연결 가능해야 함
3. **자동화 우선**: 수동 검증보다는 자동화된 검증 시스템 구축
4. **점진적 개선**: 기존 @TAG를 유지하면서 누락된 부분만 보완

### 제약사항
- **호환성 유지**: 기존 88개 파일의 @TAG 패턴 변경 최소화
- **성능 요구**: TAG 검증이 개발 워크플로우를 지연시키지 않아야 함
- **표준 준수**: MoAI-ADK TRUST 5원칙과 16-Core TAG 규칙 준수

## @TASK:TAG-COMPLETION-011 Requirements

### 기능 요구사항

#### R1: 누락된 @TAG 완전 적용
- **대상**: 17개 누락 파일 (cli/__main__.py, templates 스크립트들)
- **적용 범위**: 모든 Python 파일에 적절한 @TAG 할당
- **Primary Chain**: @REQ → @DESIGN → @TASK → @TEST 연결 완성

#### R2: TAG 일관성 및 품질 향상
- **표준화**: TAG 네이밍 규칙 통일 (예: TAG-SYSTEM-011 패턴)
- **중복 제거**: 동일 기능에 대한 중복 TAG 정리
- **계층 구조**: 모듈별/기능별 TAG 계층 체계 정립

#### R3: 자동화 검증 시스템 구축
- **실시간 검증**: pre-commit hook을 통한 @TAG 필수성 검사
- **리포트 생성**: TAG 적용 현황 및 Primary Chain 완성도 리포트
- **CI/CD 통합**: GitHub Actions에서 TAG 완성도 검증

### 비기능 요구사항

#### NR1: 성능
- TAG 검증 시간 < 5초 (전체 codebase 기준)
- 개발 워크플로우 지연 최소화

#### NR2: 유지보수성
- TAG 규칙 변경 시 일괄 적용 가능한 구조
- 새로운 모듈 추가 시 TAG 자동 제안 시스템

## @TEST:TAG-VALIDATION-011 Specifications

### 16-Core TAG 카테고리별 완성도

#### SPEC Category (요구사항)
- @REQ: 비즈니스 요구사항 및 사용자 스토리
- @DESIGN: 설계 및 아키텍처 결정사항
- @TASK: 구현 작업 단위

#### PROJECT Category (프로젝트 관리)
- @VISION: 제품 비전 및 전략
- @STRUCT: 구조 설계 및 모듈 정의
- @TECH: 기술 스택 및 도구 선택
- @ADR: 아키텍처 결정 기록

#### IMPLEMENTATION Category (구현)
- @FEATURE: 기능 구현
- @API: API 설계 및 인터페이스
- @TEST: 테스트 케이스 및 검증
- @DATA: 데이터 모델 및 스키마

#### QUALITY Category (품질)
- @PERF: 성능 최적화
- @SEC: 보안 요구사항
- @DEBT: 기술 부채 및 리팩토링
- @TODO: 향후 개선 계획

### Primary Chain 완성 전략

```
@REQ:TAG-SYSTEM-011 (요구사항)
  ↓
@DESIGN:TAG-AUDIT-011 (설계)
  ↓
@TASK:TAG-COMPLETION-011 (작업)
  ↓
@TEST:TAG-VALIDATION-011 (검증)
```

### 추적성 매트릭스

| 모듈 | 현재 완성도 | 목표 | Primary Chain |
|------|-------------|------|---------------|
| CLI | 85.7% (6/7) | 100% | 부분 완성 |
| Core | 100% (48/48) | 100% | 완성 |
| Install | 100% | 100% | 완성 |
| Utils | 100% | 100% | 완성 |
| Templates | 30% | 100% | 미완성 |

## @DEBT:TAG-TECHNICAL-011 Technical Approach

### 4단계 TDD 접근법

#### 1단계: RED - 검증 테스트 작성
- TAG 누락 파일 검출 테스트
- Primary Chain 완성도 검증 테스트
- TAG 네이밍 규칙 준수 테스트

#### 2단계: GREEN - 최소 구현
- 17개 누락 파일에 기본 @TAG 추가
- 자동화 스크립트로 일괄 적용
- Primary Chain 연결 최소 구현

#### 3단계: REFACTOR - 품질 개선
- TAG 일관성 정리 및 중복 제거
- 계층 구조 최적화
- 자동화 시스템 성능 튜닝

#### 4단계: INTEGRATE - 시스템 통합
- CI/CD 파이프라인 통합
- 개발 워크플로우와 연동
- 문서화 및 가이드 완성

### 기술적 구현 방안

#### 자동화 도구 스택
```python
# TAG 검증 엔진
class TagValidator:
    def __init__(self):
        self.patterns = self._load_tag_patterns()
        self.primary_chain = ['REQ', 'DESIGN', 'TASK', 'TEST']

    def validate_file(self, filepath: str) -> TagValidationResult:
        """개별 파일의 TAG 검증"""

    def validate_primary_chain(self, tag_id: str) -> bool:
        """Primary Chain 완성도 검증"""

    def generate_report(self) -> TagReport:
        """전체 TAG 현황 리포트 생성"""
```

#### 검증 스크립트 확장
- **기존 활용**: `validate_tags.py`, `check_traceability.py`
- **신규 개발**: `tag_completion_checker.py`, `primary_chain_validator.py`
- **통합 도구**: `tag_system_manager.py`

## @TODO:TAG-NEXT-STEPS-011 Implementation Roadmap

### 단계별 완성 목표

#### Phase 1: 기본 완성 (1주차)
- 17개 누락 파일에 @TAG 적용
- 기본 검증 테스트 작성 및 통과
- TAG 적용률 100% 달성

#### Phase 2: 품질 향상 (2주차)
- Primary Chain 완성도 80% 달성
- TAG 네이밍 규칙 표준화
- 자동화 검증 시스템 베타 버전

#### Phase 3: 시스템 통합 (3주차)
- CI/CD 파이프라인 통합
- 실시간 검증 시스템 구축
- 성능 최적화 (< 5초 검증 시간)

#### Phase 4: 완전 자동화 (4주차)
- 새 모듈 자동 TAG 제안
- 개발 워크플로우 완전 통합
- 문서화 및 가이드 완성

---

**@REQ:TAG-SYSTEM-011 연결**: 이 명세는 MoAI-ADK의 16-Core TAG 추적성 체계를 완성하여 Spec-First TDD 개발의 품질을 보장합니다.