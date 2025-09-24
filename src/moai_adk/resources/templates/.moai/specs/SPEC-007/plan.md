# SPEC-007 Implementation Plan: .claude/hooks 최적화

## @TASK:PLAN-OVERVIEW-001 구현 계획 개요

### 목표

- 전체 hook 코드량 3,853줄 → ~1,000줄 (74% 감소)
- 세션 시작 시간 50% 단축 (3-5초 → 1.5-2.5초)
- TRUST R-Readable 원칙 준수 (모든 파일 300줄 이하)

### 접근 전략

1. **Big Bang 방식 대신 점진적 리팩토링** 채택
2. **성능 측정 기반 의사결정** - 각 단계마다 벤치마크 실시
3. **백워드 호환성 유지** - 기존 에이전트 동작 보장
4. **TDD 적용** - 리팩토링 전 테스트 작성 필수

## @TASK:MILESTONE-001 우선순위별 마일스톤

### 1차 목표: 핵심 성능 병목 해결 (Critical Path)

**session_start_notice.py 최적화** (@TASK:SESSION-START-001)

- 현재 상태: 2,133줄 (전체의 55%)
- 목표 상태: ~200줄
- 예상 성능 향상: 세션 시작 시간 40% 단축
- 접근 방법:
  - 상세 분석 로직 제거/외부화
  - 중복 검사 로직 통합
  - 로깅 간소화
  - 핵심 알림만 유지

**우선 순위**: High
**리스크**: 중간 (핵심 기능 보존 필수)
**의존성**: 없음

### 2차 목표: 중복 기능 통합

**file_monitor.py 통합 모듈 생성** (@TASK:HOOKS-MERGE-001)

- 통합 대상: auto_checkpoint.py (222줄) + file_watcher.py (323줄)
- 목표 상태: ~150줄
- 예상 효과: 메모리 사용량 20% 감소, 파일 I/O 최적화
- 접근 방법:
  - 공통 이벤트 리스너 통합
  - 설정 파일 단일화
  - 중복 파일 시스템 접근 제거

**우선 순위**: Medium
**리스크**: 낮음
**의존성**: 1차 목표 완료 후

### 3차 목표: 불필요한 기능 제거

**hook 파일 제거** (@TASK:HOOKS-REMOVE-001)

- 제거 대상: tag_validator.py (430줄), check_style.py (241줄)
- 대체 방안: CI/CD 파이프라인으로 이전
- 예상 효과: 즉시 671줄 감소, 세션 로딩 시간 10% 단축

**우선 순위**: High (간단하지만 큰 효과)
**리스크**: 낮음 (기능 이전 후 제거)
**의존성**: CI/CD 대체 기능 준비 필요

### 최종 목표: 나머지 최적화

**pre_write_guard.py 경량화** (@TASK:GUARD-OPTIMIZE-001)

- 현재: 131줄 → 목표: ~50줄
- 핵심 보안 기능 유지하며 불필요한 복잡성 제거

**유지 파일들 점검**

- language_detector.py (108줄): 유지
- policy_block.py (95줄): 유지
- run_tests_and_report.py (91줄): 유지
- steering_guard.py (79줄): 유지

## @DESIGN:ARCHITECTURE-001 아키텍처 설계 방향

### 리팩토링 설계 원칙

#### 1. Single Responsibility Principle (SRP) 강화

```python
# Before: session_start_notice.py (2,133줄)
# - 세션 시작 알림 + 프로젝트 분석 + 설정 검증 + 상세 리포팅

# After: 책임 분리
session_start_notice.py  # 핵심 알림만 (~200줄)
project_analyzer.py      # 프로젝트 분석 (별도 모듈, 필요시 호출)
config_validator.py      # 설정 검증 (별도 모듈)
```

#### 2. Lazy Loading 패턴 적용

- 세션 시작 시 필수 기능만 로딩
- 상세 분석은 필요시에만 수행
- 메모리 사용량 최적화

#### 3. Configuration-Driven Behavior

- 사용자별 hook 활성화/비활성화 설정
- 세션 성능 vs 기능 상세도 trade-off 선택 가능

### 통합 모듈 설계 (file_monitor.py)

#### 통합 전략

```python
# 통합 설계 패턴
class FileMonitor:
    def __init__(self):
        self.file_watcher = FileWatcher()  # 기존 file_watcher.py 로직
        self.checkpoint_manager = CheckpointManager()  # 기존 auto_checkpoint.py 로직

    def on_file_change(self, event):
        # 통합된 이벤트 처리
        if self.should_checkpoint(event):
            self.checkpoint_manager.create_checkpoint()

    def should_checkpoint(self, event):
        # 조건부 체크포인트 로직 (설정 기반)
        pass
```

#### 성능 최적화 요소

- 단일 파일 감시 루프
- 이벤트 debouncing (과도한 체크포인트 방지)
- 메모리 효율적인 파일 캐싱

## @TASK:TECH-APPROACH-001 기술적 접근 방법

### 성능 측정 체계 구축

#### 벤치마크 스크립트 개발

```python
# hooks_benchmark.py
import time
import memory_profiler

def measure_session_start_time():
    """세션 시작 시간 측정"""
    start = time.time()
    # Hook 로딩 시뮬레이션
    load_all_hooks()
    return time.time() - start

def measure_memory_usage():
    """Hook 로딩 시 메모리 사용량 측정"""
    return memory_profiler.memory_usage(load_all_hooks)
```

#### 성능 기준치 설정

- **세션 시작 시간**: 현재 평균 4초 → 목표 2초
- **메모리 사용량**: 현재 대비 30% 감소
- **파일 I/O 횟수**: 50% 감소

### 리팩토링 안전 장치

#### 기능 보존 테스트

```python
# test_hook_compatibility.py
def test_session_start_notice_compatibility():
    """기존 세션 시작 알림 기능 호환성"""
    old_behavior = capture_old_session_start()
    new_behavior = capture_new_session_start()
    assert_equivalent_functionality(old_behavior, new_behavior)

def test_agent_communication():
    """에이전트와의 통신 호환성"""
    pass
```

#### 점진적 배포 전략

1. **Feature Flag**: 새로운 hook과 기존 hook 동시 실행 가능
2. **A/B Testing**: 사용자별 최적화 버전 선택
3. **Rollback Plan**: 문제 발생 시 즉시 이전 버전 복구

### 코드 품질 유지

#### TRUST 원칙 적용

**T - Test First**

- 모든 리팩토링 전 테스트 작성
- TDD Red-Green-Refactor 사이클 적용

**R - Readable**

- 모든 파일 300줄 이하 유지
- 명확한 함수명과 변수명 사용
- 과도한 추상화 방지

**U - Unified**

- 모듈 간 명확한 인터페이스 정의
- 순환 의존성 방지
- 계층화된 아키텍처 유지

**S - Secured**

- 기존 보안 검사 기능 보존
- 민감정보 로깅 방지
- 입력 검증 유지

**T - Trackable**

- 모든 변경사항 Git 추적
- 성능 벤치마크 히스토리 기록
- 16-Core TAG 시스템으로 추적성 유지

## @TASK:RISK-MITIGATION-001 리스크 및 대응 방안

### 주요 리스크 분석

#### 1. 기능 호환성 문제 (높음)

**리스크**: 기존 에이전트가 새로운 hook 인터페이스와 호환되지 않음
**대응방안**:

- 인터페이스 호환성 테스트 필수 실행
- 단계적 배포로 문제 조기 발견
- 기존 인터페이스 Deprecated 방식 적용

#### 2. 성능 개선 효과 미달 (중간)

**리스크**: 예상한 50% 성능 향상이 달성되지 않음
**대응방안**:

- 각 최적화 단계별 성능 측정
- 병목 지점 프로파일링으로 추가 최적화 포인트 발견
- 사용자별 설정으로 성능 vs 기능 trade-off 선택권 제공

#### 3. 복잡한 의존성 문제 (중간)

**리스크**: Hook 간 예상치 못한 의존성으로 통합 실패
**대응방안**:

- 의존성 그래프 분석 도구 활용
- 모듈 간 인터페이스 명확히 정의
- 단위 테스트로 각 모듈 독립성 검증

#### 4. 개발 일정 지연 (낮음)

**리스크**: 리팩토링 복잡도로 인한 일정 지연
**대응방안**:

- 우선순위 기반 단계적 접근
- MVP 먼저 구현 후 점진적 개선
- 정기적인 진행상황 체크포인트

### 품질 게이트 체크리스트

#### 코드 품질

- [ ] 모든 파일 300줄 이하
- [ ] 테스트 커버리지 85% 이상
- [ ] 정적 분석 도구 통과 (mypy, flake8)
- [ ] 코드 포매터 적용 (black, isort)

#### 성능 품질

- [ ] 세션 시작 시간 50% 단축 달성
- [ ] 메모리 사용량 30% 감소 달성
- [ ] 파일 I/O 50% 감소 달성

#### 호환성 품질

- [ ] 기존 에이전트와 100% 호환
- [ ] Claude Code API 변경 없음
- [ ] 설정 파일 호환성 유지

## @TASK:IMPLEMENTATION-SEQUENCE-001 구현 순서

### Phase 1: 준비 단계

1. 성능 벤치마크 스크립트 개발
2. 현재 상태 측정 및 베이스라인 설정
3. 호환성 테스트 스위트 작성
4. CI/CD 대체 기능 준비 (tag validation, style checking)

### Phase 2: 핵심 최적화

1. session_start_notice.py 리팩토링 시작
2. 핵심 기능 분리 및 간소화
3. 성능 측정 및 검증
4. 기능 호환성 테스트

### Phase 3: 통합 및 제거

1. file_monitor.py 통합 모듈 개발
2. 불필요한 hook 파일 제거
3. pre_write_guard.py 경량화
4. 전체 성능 벤치마크 실행

### Phase 4: 검증 및 배포

1. 종합 품질 게이트 검증
2. 사용자 테스트 (opt-in beta)
3. 피드백 수집 및 개선
4. 프로덕션 배포

---

_이 구현 계획은 TRUST 5원칙을 준수하며, 단계적 접근을 통해 안정성을 보장하면서도 목표 성능 향상을 달성하는 것을 목표로 합니다._
