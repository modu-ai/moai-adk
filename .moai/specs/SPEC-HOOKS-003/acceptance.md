# @SPEC:HOOKS-003 인수 기준

> **TRUST 원칙 자동 검증 (PostToolUse 통합)**
>
> Given-When-Then 시나리오 및 품질 게이트

---

## 📋 인수 테스트 시나리오

### Scenario 1: TDD GREEN 커밋 후 자동 검증

**Given**:
- Git 저장소가 초기화되어 있음
- `scripts/validate_trust.py`가 설치되어 있음
- 최근 커밋 메시지가 `🟢 GREEN: JWT 인증 구현`임

**When**:
- PostToolUse 이벤트가 발생함
- `handlers/tool.py`의 `handle_post_tool_use()`가 호출됨

**Then**:
- `detect_tdd_completion()`이 `True`를 반환함
- `trigger_trust_validation()`이 백그라운드 프로세스를 시작함
- PostToolUse 핸들러가 100ms 이내에 반환함 (blocked=False)
- HookResult.message에 "🔍 TRUST 원칙 검증 중..." 포함됨
- PID가 `.moai/.cache/validation_pids.json`에 저장됨

**검증 방법**:
```bash
# 1. 테스트 커밋 생성
git commit -m "🟢 GREEN: JWT 인증 구현"

# 2. PostToolUse 핸들러 실행 (통합 테스트)
pytest tests/integration/test_hooks_post_tool_use.py -v

# 3. PID 파일 확인
cat .moai/.cache/validation_pids.json
# 예상 출력: [12345]

# 4. 프로세스 실행 확인
ps aux | grep validate_trust.py
```

---

### Scenario 2: REFACTOR 커밋 후 자동 검증

**Given**:
- Git 저장소가 초기화되어 있음
- `scripts/validate_trust.py`가 설치되어 있음
- 최근 커밋 메시지가 `♻️ REFACTOR: 코드 품질 개선`임

**When**:
- PostToolUse 이벤트가 발생함

**Then**:
- `detect_tdd_completion()`이 `True`를 반환함
- TRUST 검증이 백그라운드에서 실행됨
- 알림 메시지 "🔍 TRUST 원칙 검증 중..." 출력됨

**검증 방법**:
```bash
# 1. 테스트 커밋 생성
git commit -m "♻️ REFACTOR: 코드 품질 개선"

# 2. 핸들러 실행 및 메시지 확인
pytest tests/integration/test_hooks_refactor_commit.py -v
```

---

### Scenario 3: RED 커밋 무시 (검증 미실행)

**Given**:
- Git 저장소가 초기화되어 있음
- 최근 커밋 메시지가 `🔴 RED: 실패하는 테스트 작성`임

**When**:
- PostToolUse 이벤트가 발생함

**Then**:
- `detect_tdd_completion()`이 `False`를 반환함
- TRUST 검증이 실행되지 않음
- HookResult.blocked = False (알림 메시지 없음)

**검증 방법**:
```bash
# 1. 테스트 커밋 생성
git commit -m "🔴 RED: 실패하는 테스트 작성"

# 2. 핸들러 실행 및 검증 미실행 확인
pytest tests/unit/test_detect_tdd_completion.py::test_ignore_red_commit -v

# 3. PID 파일이 생성되지 않았는지 확인
test ! -f .moai/.cache/validation_pids.json
```

---

### Scenario 4: 검증 결과 수집 (통과 시나리오)

**Given**:
- TRUST 검증 프로세스가 백그라운드에서 실행 완료됨
- 검증 결과가 "통과" (status: passed)
- 결과 파일 `.moai/.cache/validation_result_12345.json` 존재

**When**:
- SessionStart 또는 UserMessage 이벤트가 발생함
- `handlers/notification.py`의 `collect_pending_validation_results()`가 호출됨

**Then**:
- 결과 파일이 읽혀짐
- 알림 메시지가 생성됨:
  ```
  ✅ **TRUST 원칙 검증 통과**
  - 테스트 커버리지: 87%
  - 코드 제약 준수: 45/45
  - TAG 체인 무결성: OK
  ```
- 결과 파일이 자동 삭제됨
- PID가 `validation_pids.json`에서 제거됨

**검증 방법**:
```bash
# 1. 테스트 결과 파일 생성 (mock)
echo '{"status":"passed","test_coverage":87,"code_constraints_passed":45,"code_constraints_total":45}' \
  > .moai/.cache/validation_result_12345.json

# 2. 알림 수집 함수 실행
pytest tests/unit/test_collect_validation_result.py::test_format_passed_message -v

# 3. 결과 파일 삭제 확인
test ! -f .moai/.cache/validation_result_12345.json
```

---

### Scenario 5: 검증 결과 수집 (실패 시나리오)

**Given**:
- TRUST 검증 프로세스가 백그라운드에서 실행 완료됨
- 검증 결과가 "실패" (status: failed)
- 실패 원인: 테스트 커버리지 78% (목표 85%)

**When**:
- SessionStart 이벤트가 발생함

**Then**:
- 알림 메시지가 생성됨:
  ```
  ❌ **TRUST 원칙 검증 실패**
  - 실패 원인: 테스트 커버리지 부족
  - 테스트 커버리지: 78% (목표 85%)
  - 권장 조치: 추가 테스트 케이스 작성 권장
  ```

**검증 방법**:
```bash
# 1. 테스트 결과 파일 생성 (실패 시나리오)
echo '{"status":"failed","error":"테스트 커버리지 부족","test_coverage":78,"recommendation":"추가 테스트 케이스 작성 권장"}' \
  > .moai/.cache/validation_result_12345.json

# 2. 알림 메시지 포맷 확인
pytest tests/unit/test_format_validation_result.py::test_format_failed_message -v
```

---

### Scenario 6: 검증 도구 없음 처리

**Given**:
- Git 저장소가 초기화되어 있음
- `scripts/validate_trust.py`가 설치되지 않음
- 최근 커밋 메시지가 `🟢 GREEN: 기능 구현`임

**When**:
- PostToolUse 이벤트가 발생함

**Then**:
- `detect_tdd_completion()`이 `True`를 반환함
- 검증 도구 존재 확인 실패
- HookResult.message에 "ℹ️ TRUST 검증 도구가 없습니다. scripts/validate_trust.py 설치 필요" 포함됨
- 검증 프로세스가 시작되지 않음

**검증 방법**:
```bash
# 1. 검증 도구 임시 제거 (백업)
mv scripts/validate_trust.py scripts/validate_trust.py.bak

# 2. 핸들러 실행 및 Info 메시지 확인
pytest tests/unit/test_handle_post_tool_use.py::test_handler_no_validation_tool -v

# 3. 검증 도구 복원
mv scripts/validate_trust.py.bak scripts/validate_trust.py
```

---

## 🎯 품질 게이트 기준

### 1. 테스트 커버리지

**목표**: ≥85%

**측정 방법**:
```bash
pytest --cov=.claude/hooks/alfred/handlers/tool \
       --cov=.claude/hooks/alfred/handlers/notification \
       --cov=.claude/hooks/alfred/core/validation \
       --cov-report=term-missing \
       --cov-fail-under=85
```

**커버리지 대상**:
- `handlers/tool.py`: detect_tdd_completion, is_alfred_build_command, handle_post_tool_use
- `handlers/notification.py`: collect_pending_validation_results, format_validation_result
- `core/validation.py`: trigger_trust_validation, collect_validation_result, PID 관리

---

### 2. 성능 기준

| 항목               | 목표   | 측정 방법                            |
| ------------------ | ------ | ------------------------------------ |
| Git 로그 파싱      | <10ms  | `time git log -5 --pretty=format:%s` |
| PostToolUse 핸들러 | <100ms | Hooks 시스템 타이머                  |
| 검증 프로세스 시작 | <50ms  | `subprocess.Popen()` 호출 시간       |

**측정 코드**:
```python
import time

def test_handler_performance():
    start = time.perf_counter()
    result = handle_post_tool_use(mock_payload)
    elapsed = (time.perf_counter() - start) * 1000  # ms

    assert elapsed < 100, f"Handler took {elapsed:.2f}ms (expected <100ms)"
```

---

### 3. 코드 품질

**린트 도구**: Ruff

**실행 명령**:
```bash
ruff check .claude/hooks/alfred/handlers/tool.py \
           .claude/hooks/alfred/handlers/notification.py \
           .claude/hooks/alfred/core/validation.py
```

**기준**:
- 경고 0개
- 에러 0개
- 복잡도 ≤10

---

### 4. 타입 검증

**도구**: mypy

**실행 명령**:
```bash
mypy --strict .claude/hooks/alfred/handlers/tool.py \
              .claude/hooks/alfred/handlers/notification.py \
              .claude/hooks/alfred/core/validation.py
```

**기준**:
- 타입 에러 0개
- `Any` 타입 사용 최소화

---

## 🔧 검증 방법 및 도구

### 단위 테스트 실행

```bash
# 전체 단위 테스트 실행
pytest tests/unit/test_hooks_trust_validation.py -v

# 커버리지 포함 실행
pytest tests/unit/test_hooks_trust_validation.py \
       --cov=.claude/hooks/alfred \
       --cov-report=term-missing

# 특정 시나리오만 실행
pytest tests/unit/test_hooks_trust_validation.py::test_detect_green_commit -v
```

---

### 통합 테스트 실행

```bash
# End-to-End 시나리오
pytest tests/integration/test_hooks_e2e.py -v

# 성능 테스트
pytest tests/integration/test_hooks_performance.py -v

# 에러 시나리오
pytest tests/integration/test_hooks_error_scenarios.py -v
```

---

### 수동 검증 절차

#### 1. TDD 워크플로우 검증

```bash
# 1. 새 SPEC 생성
/alfred:1-plan "테스트 기능"

# 2. TDD 구현
/alfred:2-run TEST-001

# 3. REFACTOR 커밋 생성 (수동)
git add src/ tests/
git commit -m "♻️ REFACTOR: 코드 품질 개선"

# 4. PostToolUse 이벤트 트리거 (자동)
# → 검증 백그라운드 실행 시작

# 5. 다음 세션 시작 또는 사용자 메시지 입력
# → 검증 결과 알림 확인

# 예상 출력:
# ✅ **TRUST 원칙 검증 통과**
# - 테스트 커버리지: 87%
# - 코드 제약 준수: 45/45
# - TAG 체인 무결성: OK
```

#### 2. 검증 도구 없음 시나리오

```bash
# 1. 검증 도구 임시 제거
mv scripts/validate_trust.py scripts/validate_trust.py.bak

# 2. TDD 구현 및 REFACTOR 커밋
git commit -m "♻️ REFACTOR: 리팩토링"

# 3. 알림 메시지 확인
# 예상 출력:
# ℹ️ TRUST 검증 도구가 없습니다. scripts/validate_trust.py 설치 필요

# 4. 검증 도구 복원
mv scripts/validate_trust.py.bak scripts/validate_trust.py
```

#### 3. 검증 실패 시나리오

```bash
# 1. 테스트 커버리지를 의도적으로 낮춤 (일부 테스트 주석 처리)

# 2. GREEN 커밋
git commit -m "🟢 GREEN: 구현 완료"

# 3. 검증 결과 확인
# 예상 출력:
# ❌ **TRUST 원칙 검증 실패**
# - 실패 원인: 테스트 커버리지 부족
# - 테스트 커버리지: 78% (목표 85%)
# - 권장 조치: 추가 테스트 케이스 작성 권장
```

---

## ✅ Definition of Done (최종 인수 기준)

### 필수 요구사항

- [ ] **6개 Given-When-Then 시나리오 통과**:
  - [ ] Scenario 1: GREEN 커밋 후 자동 검증
  - [ ] Scenario 2: REFACTOR 커밋 후 자동 검증
  - [ ] Scenario 3: RED 커밋 무시
  - [ ] Scenario 4: 검증 결과 수집 (통과)
  - [ ] Scenario 5: 검증 결과 수집 (실패)
  - [ ] Scenario 6: 검증 도구 없음 처리

- [ ] **테스트 커버리지 ≥85%**:
  - [ ] handlers/tool.py: ≥85%
  - [ ] handlers/notification.py: ≥85%
  - [ ] core/validation.py: ≥85%

- [ ] **성능 기준 충족**:
  - [ ] Git 로그 파싱 <10ms
  - [ ] PostToolUse 핸들러 <100ms
  - [ ] 검증 프로세스 시작 <50ms

- [ ] **코드 품질 통과**:
  - [ ] Ruff 린트 경고/에러 0개
  - [ ] mypy 타입 검증 통과
  - [ ] 복잡도 ≤10

---

### 선택 요구사항 (권장)

- [ ] **문서화 완료**:
  - [ ] README.md 업데이트 (TRUST 자동 검증 안내)
  - [ ] validation-logic-migration.md Phase 2 계획 작성

- [ ] **수동 검증 완료**:
  - [ ] End-to-End 시나리오 3회 이상 테스트
  - [ ] 다양한 Git 저장소 상태에서 테스트 (Detached HEAD, Bare repo 등)

- [ ] **에러 처리 강건성**:
  - [ ] Git 저장소 없음
  - [ ] 검증 도구 없음
  - [ ] 의존성 미설치
  - [ ] 프로세스 타임아웃
  - [ ] 중복 실행 방지

---

## 📊 인수 체크리스트 (실행 순서)

1. **[ ] 단위 테스트 실행** (15개 테스트 통과)
2. **[ ] 통합 테스트 실행** (3개 시나리오 통과)
3. **[ ] 테스트 커버리지 확인** (≥85%)
4. **[ ] 성능 테스트 실행** (3개 기준 충족)
5. **[ ] 코드 품질 검증** (Ruff + mypy 통과)
6. **[ ] 수동 검증** (3개 시나리오 확인)
7. **[ ] 문서 업데이트** (README.md, Phase 2 계획)
8. **[ ] 최종 리뷰** (코드 리뷰, SPEC 준수 확인)

---

**Last Updated**: 2025-10-16
**Author**: @Goos
**Status**: Draft (v0.0.1)
