---
spec_id: SPEC-007
status: active
priority: medium
dependencies: []
tags:
  - claude-hooks
  - optimization
  - refactoring
  - performance
  - cleanup
  - session-startup
---

# SPEC-007: .claude/hooks 파일 최적화 및 불필요한 것들 제거

## @SPEC:HOOKS-OPTIMIZE-001 Environment (환경 및 가정사항)

### 운영 환경

- **Claude Code 환경**: 훅 시스템 기반 자동화
- **Python 런타임**: 3.11+ 지원
- **파일 시스템**: Unix/Windows 호환 경로 처리
- **Session 시작**: 세션당 평균 3-5초 소요 (현재)

### 현재 문제 상황

- **Hook 파일 현황**: 10개 파일, 총 3,853줄
- **최대 문제점**: session_start_notice.py (2,133줄, 전체의 55%)
- **성능 이슈**: 세션 시작 지연, 메모리 사용량 증가
- **유지보수 복잡성**: 중복 기능, 불필요한 복잡성

### 기술적 제약사항

- Claude Code 내장 훅 시스템 호환성 유지
- 기존 에이전트와의 인터페이스 호환성
- TRUST 5원칙 R-Readable 준수 (파일당 300줄 이하)

## @SPEC:HOOKS-REFACTOR-001 Assumptions (전제 조건)

### 리팩토링 가정

1. **핵심 기능 보존**: 세션 시작 알림, 보안 검사, 언어 감지 등 필수 기능
2. **성능 우선**: 코드 가독성보다 세션 시작 성능 우선
3. **단계적 접근**: 점진적 리팩토링으로 안정성 확보
4. **백워드 호환성**: 기존 설정과 에이전트 동작 유지

### 기능 분류 기준

- **유지 필수**: policy_block.py, steering_guard.py, language_detector.py
- **통합 대상**: auto_checkpoint.py + file_watcher.py
- **제거 대상**: tag_validator.py, check_style.py
- **최적화 대상**: session_start_notice.py, pre_write_guard.py

## @CODE:HOOKS-CLEANUP-001 Requirements (기능 요구사항)

### 1. session_start_notice.py 최적화 (@CODE:SESSION-START-001)

- **현재**: 2,133줄 → **목표**: ~200줄 (90% 감소)
- **핵심 기능만 유지**: 프로젝트 상태 확인, 가이드 위반 알림
- **제거 대상**: 상세 분석 로직, 중복 검사, 과도한 로깅
- **성능 목표**: 세션 시작 시간 50% 단축

### 2. 중복 기능 통합 (@CODE:HOOKS-MERGE-001)

- **통합**: auto_checkpoint.py + file_watcher.py → file_monitor.py
- **기능 집약**: 파일 변경 감지 + 자동 체크포인트 생성
- **코드 감소**: 222 + 323 = 545줄 → ~150줄

### 3. 불필요한 Hook 제거 (@CODE:HOOKS-REMOVE-001)

- **제거**: tag_validator.py (430줄) - 사용 빈도 낮음, CI/CD로 이전
- **제거**: check_style.py (241줄) - linter/formatter로 대체
- **총 제거**: 671줄

### 4. pre_write_guard.py 경량화 (@CODE:GUARD-OPTIMIZE-001)

- **현재**: 131줄 → **목표**: ~50줄
- **핵심만 유지**: 민감정보 검사, 위험 파일 차단
- **제거**: 상세 로깅, 복잡한 분석

## @TEST:HOOKS-VALIDATION-001 Specifications (상세 명세)

### 성능 명세

- **세션 시작 시간**: 현재 3-5초 → 목표 1.5-2.5초 (50% 단축)
- **메모리 사용량**: Hook 로딩 시 현재 대비 30% 감소
- **파일 I/O**: 불필요한 파일 읽기 최소화

### 코드 품질 명세

- **TRUST R-Readable**: 모든 파일 300줄 이하
- **단일 책임 원칙**: 각 Hook 파일의 역할 명확화
- **의존성 최소화**: 외부 라이브러리 의존성 감소

### 호환성 명세

- **Claude Code API**: 기존 훅 인터페이스 100% 호환
- **에이전트 통신**: JSON 기반 메시지 포맷 유지
- **설정 파일**: .claude/settings.json 변경 없음

### 최적화된 Hook 구조 설계

```python
# 목표 Hook 구조 (총 ~1,000줄)
hooks/moai/
├── session_start_notice.py    # ~200줄 (현재: 2,133줄)
├── file_monitor.py           # ~150줄 (통합: auto_checkpoint + file_watcher)
├── pre_write_guard.py        # ~50줄  (현재: 131줄)
├── language_detector.py      # ~108줄 (유지)
├── policy_block.py          # ~95줄  (유지)
├── run_tests_and_report.py  # ~91줄  (유지)
├── steering_guard.py        # ~79줄  (유지)
└── [제거]
    ├── tag_validator.py      # 430줄 제거
    └── check_style.py        # 241줄 제거
```

### Hook별 상세 최적화 전략

#### session_start_notice.py 리팩토링 계획

**현재 문제**: 2,133줄의 과도한 복잡성
**최적화 접근**:

- 중복 분석 로직 제거 (약 1,200줄 감소)
- 과도한 로깅/리포팅 간소화 (약 500줄 감소)
- 프로젝트 상태 확인 로직 최적화 (약 233줄 감소)
- 핵심 알림 기능만 유지 (약 200줄 목표)

**유지할 핵심 기능**:

- MoAI 개발 가이드 위반 감지
- 프로젝트 초기화 상태 확인
- 중요 설정 누락 알림
- 간단한 세션 시작 메시지

#### file_monitor.py 통합 설계

**통합 대상**:

- auto_checkpoint.py (222줄): Git 체크포인트 자동 생성
- file_watcher.py (323줄): 파일 변경 감지

**통합 효과**:

- 중복 파일 I/O 로직 제거
- 이벤트 리스너 통합
- 설정 파일 공유로 복잡성 감소

**핵심 기능**:

- 파일 변경 이벤트 통합 감지
- 조건부 체크포인트 생성 (설정 가능)
- 메모리 효율적인 파일 감시

#### 제거 대상 분석

**tag_validator.py (430줄) 제거 근거**:

- 사용 빈도 낮음 (월 1회 미만)
- CI/CD 단계에서 대체 가능
- 세션 시작 성능에 부정적 영향
- 복잡한 정규식 패턴으로 메모리 사용량 증가

**check_style.py (241줄) 제거 근거**:

- black, isort, flake8 등 전용 도구가 더 효율적
- 중복 검사로 인한 성능 저하
- 설정 파일(pyproject.toml)과 불일치 위험

## @DOC:TRACEABILITY-001 Traceability (추적성 태그)

### Primary Chain

- `@SPEC:HOOKS-OPTIMIZE-001` → `@SPEC:HOOKS-REFACTOR-001` → `@CODE:HOOKS-CLEANUP-001` → `@TEST:HOOKS-VALIDATION-001`

### Implementation Chain

- `@CODE:SESSION-OPTIMIZE` → `@CODE:HOOKS-INTERFACE` → `@CODE:STARTUP-TIME` → `@CODE:CONFIG-MAINTAIN`

### Quality Chain

- `@CODE:MEMORY-REDUCE` → `@CODE:GUARD-MAINTAIN` → `@DOC:HOOKS-GUIDE` → `@CODE:CODE-CLEANUP`

### 연관 SPEC 및 요구사항

- **@SPEC:USER-001**: 개인 개발자의 "즉시 얻고 싶은 결과" - 빠른 세션 시작
- **@DOC:STRATEGY-001**: 자동화 심도 - 복잡성 제거를 통한 사용자 경험 개선
- **@DOC:MODULES-001**: Claude Extensions 계층 최적화

---

_이 명세는 TRUST 5원칙의 R-Readable(가독성)과 T-Trackable(추적성) 강화를 목표로 하며, 세션 시작 성능 50% 단축을 통해 개발자 생산성을 향상시킵니다._
