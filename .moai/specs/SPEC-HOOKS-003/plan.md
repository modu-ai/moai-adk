# SPEC-HOOKS-003 구현 계획서

> **TRUST 원칙 자동 검증 (PostToolUse 통합)**
>
> Phase 1 (validation-logic-migration.md) 구현 계획

---

## 📋 구현 개요

### 목표
- `/alfred:2-build` 완료 후 TRUST 원칙 검증 자동 실행
- PostToolUse Hook을 통한 비동기 검증 트리거
- 검증 결과를 사용자에게 알림 메시지로 전달

### 범위
- **포함**: PostToolUse 핸들러 확장, TDD 감지, 비동기 실행, 결과 수집
- **제외**: TRUST 검증 도구 자체 구현 (SPEC-TRUST-001에서 완료됨)

### 의존성
- ✅ SPEC-HOOKS-001: Hooks 시스템 아키텍처 (v0.1.0 완료)
- ✅ SPEC-TRUST-001: TRUST 검증 시스템 (scripts/validate_trust.py 구현 완료)
- ⏳ SPEC-HOOKS-003: 본 SPEC (작성 중)

---

## 🎯 3단계 마일스톤

### 1차 목표: TDD 완료 감지 로직

**우선순위**: High
**의존성**: HOOKS-001 완료

**구현 대상**:
1. Git 로그 분석 함수 (`detect_tdd_completion`)
   - 최근 5개 커밋 메시지 파싱
   - `🟢 GREEN:` 또는 `♻️ REFACTOR:` 키워드 감지
   - 성능: <10ms

2. Alfred 2-build 명령 감지 (`is_alfred_build_command`)
   - PostToolUse payload 분석
   - Bash 명령어 또는 Agent 설명에서 `alfred:2-build` 검색
   - False Positive 최소화

3. Git 저장소 상태 확인
   - `.git/` 디렉토리 존재 확인
   - Detached HEAD 상태 처리
   - Bare repository 제외

**테스트 케이스**:
- ✅ GREEN 커밋 감지 성공
- ✅ REFACTOR 커밋 감지 성공
- ✅ RED 커밋 무시 (검증 트리거 안 함)
- ✅ alfred:2-build 명령 감지
- ✅ Git 저장소 없음 처리
- ✅ Detached HEAD 상태 처리

**완료 기준**:
- [ ] `detect_tdd_completion()` 함수 구현 및 테스트 통과
- [ ] `is_alfred_build_command()` 함수 구현 및 테스트 통과
- [ ] Git 로그 파싱 성능 <10ms 확인

---

### 2차 목표: 비동기 TRUST 검증 실행

**우선순위**: High
**의존성**: 1차 목표 완료, TRUST-001 검증 도구 설치

**구현 대상**:
1. 백그라운드 프로세스 실행 (`trigger_trust_validation`)
   - `subprocess.Popen()` 사용
   - stdout/stderr 파이프 설정
   - 타임아웃: 30초

2. 프로세스 ID 관리
   - 임시 파일에 PID 저장 (`.moai/.cache/validation_pids.json`)
   - 프로세스 완료 후 결과 파일 경로 매핑
   - 좀비 프로세스 방지 (psutil 사용)

3. PostToolUse 핸들러 통합 (`handlers/tool.py`)
   - TDD 완료 감지 → 비동기 검증 트리거
   - 100ms 이내 반환 보장 (blocked=False)
   - 초기 알림 메시지: "🔍 TRUST 원칙 검증 중..."

**테스트 케이스**:
- ✅ subprocess.Popen() 호출 성공
- ✅ 프로세스 ID 저장/로드
- ✅ PostToolUse 핸들러 <100ms 반환
- ✅ 검증 도구 없음 처리 (ℹ️ Info 메시지)
- ✅ 의존성 미설치 처리 (❌ Critical 메시지)

**완료 기준**:
- [ ] `trigger_trust_validation()` 함수 구현 및 테스트 통과
- [ ] PID 저장/로드 유틸리티 구현
- [ ] PostToolUse 핸들러 통합 완료
- [ ] 성능 테스트: 핸들러 실행 <100ms

---

### 3차 목표: 검증 결과 수집 및 알림

**우선순위**: Medium
**의존성**: 2차 목표 완료

**구현 대상**:
1. 검증 결과 수집 (`collect_validation_result`)
   - 프로세스 완료 대기 (최대 30초)
   - stdout JSON 파싱
   - stderr 에러 메시지 추출

2. 결과 파일 관리
   - 결과 파일 경로: `.moai/.cache/validation_result_{pid}.json`
   - 파일 읽기 후 자동 삭제
   - 파일 누적 방지

3. 알림 메시지 생성 (`format_validation_result`)
   - 통과 시: ✅ 커버리지, 제약 준수, TAG 체인 무결성 표시
   - 실패 시: ❌ 실패 원인, 권장 조치 표시
   - Markdown 형식

4. SessionStart/UserMessage 핸들러 통합 (`handlers/notification.py`)
   - 대기 중인 검증 결과 자동 수집
   - 다음 Hook 이벤트에서 알림 메시지 추가
   - 중복 알림 방지

**테스트 케이스**:
- ✅ JSON 파싱 성공 (통과/실패 시나리오)
- ✅ 알림 메시지 포맷 검증
- ✅ 결과 파일 자동 삭제
- ✅ 중복 알림 방지
- ✅ 프로세스 타임아웃 처리

**완료 기준**:
- [ ] `collect_validation_result()` 함수 구현 및 테스트 통과
- [ ] `format_validation_result()` 함수 구현 및 테스트 통과
- [ ] SessionStart 핸들러 통합 완료
- [ ] End-to-End 통합 테스트 통과

---

## 🔧 기술적 접근 방법

### 1. Git 로그 분석

**도구**: `subprocess.run()` + Git CLI

**구현 전략**:
```python
# 최근 5개 커밋 메시지 가져오기
git log -5 --pretty=format:%s

# 출력 예시:
# ♻️ REFACTOR: 코드 품질 개선
# 🟢 GREEN: 테스트 통과 구현
# 🔴 RED: 실패하는 테스트 작성
```

**장점**:
- 빠름 (<10ms)
- 표준 Git 명령어 사용
- 추가 의존성 없음

**단점**:
- Git 저장소 필수
- Detached HEAD 상태 처리 필요

---

### 2. 비동기 프로세스 실행

**도구**: `subprocess.Popen()` + 임시 파일

**구현 전략**:
```python
# 1. 백그라운드 프로세스 시작
process = subprocess.Popen(
    ["python", "scripts/validate_trust.py", "--json"],
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# 2. PID 저장
save_pid(process.pid)

# 3. 다음 Hook 이벤트에서 결과 수집
if process.poll() is not None:  # 프로세스 완료됨
    stdout, stderr = process.communicate()
    result = json.loads(stdout)
```

**장점**:
- 100ms 제약 준수
- 사용자 워크플로우 중단 없음
- 표준 라이브러리만 사용

**단점**:
- 프로세스 관리 복잡도 증가
- PID 영속화 필요

---

### 3. 검증 결과 포맷

**도구**: Jinja2 템플릿 또는 f-string

**구현 전략**:
```python
def format_validation_result(result: dict) -> str:
    if result["status"] == "passed":
        return f"""
✅ **TRUST 원칙 검증 통과**
- 테스트 커버리지: {result["test_coverage"]}%
- 코드 제약 준수: {result["code_constraints_passed"]}/{result["code_constraints_total"]}
- TAG 체인 무결성: OK
"""
    else:
        return f"""
❌ **TRUST 원칙 검증 실패**
- 실패 원인: {result["error"]}
- 권장 조치: {result["recommendation"]}
"""
```

**장점**:
- Markdown 형식 (Claude Code 지원)
- 사용자 친화적
- 심각도 아이콘 사용 (✅ ❌ ⚠️ ℹ️)

---

## 🏗️ 아키텍처 설계

### 디렉토리 구조

```
.claude/hooks/alfred/
├── handlers/
│   ├── tool.py              # @CODE:HOOKS-003 (PostToolUse 핸들러)
│   │   ├── detect_tdd_completion()
│   │   ├── is_alfred_build_command()
│   │   └── handle_post_tool_use()
│   └── notification.py      # @CODE:HOOKS-003 (결과 알림)
│       ├── collect_pending_validation_results()
│       └── format_validation_result()
├── core/
│   └── validation.py        # @CODE:HOOKS-003 (검증 유틸리티)
│       ├── trigger_trust_validation()
│       ├── collect_validation_result()
│       ├── save_validation_pid()
│       └── load_validation_pids()
└── tests/
    └── unit/
        └── test_hooks_trust_validation.py  # @TEST:HOOKS-003
```

### 데이터 흐름

```
1. PostToolUse 이벤트 발생
   ↓
2. handlers/tool.py: TDD 완료 감지
   ↓
3. core/validation.py: 비동기 검증 실행
   ↓
4. PID 저장 (.moai/.cache/validation_pids.json)
   ↓
5. 다음 Hook 이벤트 (SessionStart/UserMessage)
   ↓
6. handlers/notification.py: 결과 수집
   ↓
7. 알림 메시지 추가 (HookResult.message)
   ↓
8. 사용자에게 전달
```

---

## ⚠️ 리스크 및 대응 방안

### 리스크 1: 프로세스 관리 복잡도

**설명**: 비동기 프로세스의 PID를 추적하고 결과를 수집하는 로직이 복잡함.

**영향**: Medium
- 좀비 프로세스 발생 가능
- 결과 누락 가능

**대응 방안**:
1. **psutil 사용**: 프로세스 상태 확인 및 종료
2. **타임아웃 설정**: 30초 이상 소요 시 자동 종료
3. **결과 파일 TTL**: 24시간 후 자동 삭제

---

### 리스크 2: Git 로그 파싱 성능

**설명**: 대규모 저장소에서 Git 로그 파싱이 느릴 수 있음.

**영향**: Low
- 최근 5개 커밋만 조회 (빠름)
- 일반적으로 <10ms

**대응 방안**:
1. **캐싱**: 마지막 커밋 SHA를 저장하고 변경 시만 재파싱
2. **타임아웃**: 100ms 초과 시 건너뜀

---

### 리스크 3: 의존성 누락

**설명**: TRUST 검증 도구 또는 의존성이 설치되지 않았을 수 있음.

**영향**: High
- 검증 실행 실패
- 사용자 혼란

**대응 방안**:
1. **선제적 확인**: PostToolUse 핸들러에서 `scripts/validate_trust.py` 존재 확인
2. **명확한 에러 메시지**: "❌ TRUST 검증 도구가 없습니다. scripts/validate_trust.py 설치 필요"
3. **설치 가이드**: 문서에 의존성 설치 방법 명시

---

### 리스크 4: False Positive/Negative

**설명**: TDD 완료를 잘못 감지하거나 놓칠 수 있음.

**영향**: Medium
- 불필요한 검증 실행 (False Positive)
- 필요한 검증 누락 (False Negative)

**대응 방안**:
1. **다중 감지 전략**: Git 로그 + Alfred 명령어 분석 병행
2. **수동 트리거**: 사용자가 `/trust-check` 명령으로 수동 실행 가능 (SPEC-TRUST-001)
3. **로깅**: 감지 로직 디버깅을 위한 상세 로그

---

## 📊 성능 목표

| 항목 | 목표 | 측정 방법 |
|------|------|-----------|
| Git 로그 파싱 | <10ms | `time git log -5 --pretty=format:%s` |
| PostToolUse 핸들러 | <100ms | Hooks 시스템 타이머 |
| 검증 프로세스 시작 | <50ms | `subprocess.Popen()` 호출 시간 |
| 전체 비동기 실행 | <100ms | PostToolUse 시작 → 반환 시간 |
| 검증 완료 | <30초 | 백그라운드 프로세스 종료 시간 |

---

## 🧪 테스트 전략

### 단위 테스트 (15개 이상, ≥85% 커버리지)

| 테스트 케이스 | 대상 함수 | 시나리오 |
|---------------|-----------|----------|
| test_detect_green_commit | detect_tdd_completion() | GREEN 커밋 감지 성공 |
| test_detect_refactor_commit | detect_tdd_completion() | REFACTOR 커밋 감지 성공 |
| test_ignore_red_commit | detect_tdd_completion() | RED 커밋 무시 |
| test_no_git_repo | detect_tdd_completion() | Git 저장소 없음 처리 |
| test_alfred_build_detection | is_alfred_build_command() | alfred:2-build 명령 감지 |
| test_trigger_validation | trigger_trust_validation() | 프로세스 시작 성공 |
| test_save_load_pid | save_validation_pid() | PID 영속화 |
| test_collect_result_success | collect_validation_result() | JSON 파싱 성공 (통과) |
| test_collect_result_failure | collect_validation_result() | JSON 파싱 성공 (실패) |
| test_format_passed_message | format_validation_result() | 통과 메시지 포맷 |
| test_format_failed_message | format_validation_result() | 실패 메시지 포맷 |
| test_handler_performance | handle_post_tool_use() | <100ms 반환 확인 |
| test_handler_no_validation_tool | handle_post_tool_use() | 도구 없음 처리 |
| test_handler_duplicate_trigger | handle_post_tool_use() | 중복 실행 방지 |
| test_notification_collect_pending | collect_pending_validation_results() | 대기 결과 수집 |

### 통합 테스트 (3개)

1. **End-to-End 시나리오**:
   - `/alfred:2-build SPEC-XXX` 실행
   - REFACTOR 커밋 생성
   - PostToolUse 트리거 확인
   - 검증 결과 알림 확인

2. **성능 테스트**:
   - 100개 커밋 히스토리에서 Git 로그 파싱 <10ms
   - PostToolUse 핸들러 <100ms 반환

3. **에러 시나리오**:
   - Git 저장소 없음
   - 검증 도구 없음
   - 의존성 미설치

---

## 📝 Definition of Done

- [ ] 3개 파일 (spec.md, plan.md, acceptance.md) 작성 완료
- [ ] spec.md에 EARS 구문 충실히 적용 (10개 요구사항)
- [ ] plan.md에 3단계 마일스톤, 기술적 접근, 아키텍처, 리스크 명시
- [ ] acceptance.md에 6개 Given-When-Then 시나리오 작성
- [ ] TAG 체인 명확히 표기 (@SPEC, @TEST, @CODE)
- [ ] 의존성 TAG 명시 (HOOKS-001, TRUST-001)
- [ ] 테스트 커버리지 ≥85% 목표 명시
- [ ] 성능 목표 명시 (<100ms PostToolUse, <10ms Git 파싱)

---

**Last Updated**: 2025-10-16
**Author**: @Goos
**Status**: Draft (v0.0.1)
