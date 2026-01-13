---
id: SPEC-HOOK-001
version: "1.0.0"
status: "draft"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [hook-system, quality-improvement, type-safety, file-operations, logging]
spec_id: SPEC-HOOK-001
---

# 인수 기준: SPEC-HOOK-001 Hook System 통합 품질 개선

## 테스트 전략 개요

이 문서는 SPEC-HOOK-001의 5가지 핵심 이슈(H1-H5)에 대한 Given-When-Then 형식의 테스트 시나리오와 Edge case, 성공 기준을 정의합니다.

### 테스트 커버리지 목표

- 전체 테스트 커버리지: 85% 이상
- 핵심 모듈 커버리지: 90% 이상
- Edge case 커버리지: 80% 이상

---

## H1: 타입 어노테이션 통합

### 시나리오 1.1: 함수 타입 어노테이션 검증

**Given** Hook 함수가 Python 3.13+ 문법을 사용하는 환경에서
**When** 함수가 `str | None` 형식의 타입 어노테이션을 포함하여 정의될 때
**Then** mypy strict mode가 통과해야 한다

```python
# 예시
def read_config(file_path: Path) -> dict[str, str] | None:
    """설정 파일을 읽고 딕셔너리를 반환하거나 None을 반환"""
    if not file_path.exists():
        return None
    return {"key": "value"}

# 테스트
assert mypy.run(["--strict", "hook_file.py"]) == 0
```

### 시나리오 1.2: typing 모듈 의존성 제거

**Given** 기존 코드가 `typing.Optional[str]`을 사용하는 환경에서
**When** 타입 어노테이션이 `str | None` 형식으로 리팩토링될 때
**Then** `from typing import Optional` import가 제거되어야 한다

```python
# Before (거부)
from typing import Optional
def get_user() -> Optional[str]:
    return "user"

# After (허용)
def get_user() -> str | None:
    return "user"
```

### Edge Cases

**EC-1.1:** 제네릭 타입의 builtin 변환
```python
# Before (거부)
from typing import List, Dict
def process(items: List[str]) -> Dict[str, int]:
    return {"item": 1}

# After (허용)
def process(items: list[str]) -> dict[str, int]:
    return {"item": 1}
```

**EC-1.2:** 중첩된 nullable 타입
```python
def nested_type() -> list[str] | None:
    """중첩된 타입에서도 올바른 형식 사용"""
    return None
```

### 성공 기준

- [ ] 모든 Hook 파일이 mypy strict mode 통과
- [ ] `from typing import Optional, List, Dict` import 제거
- [ ] 모든 함수 파라미터와 반환값에 타입 어노테이션 존재
- [ ] 0개의 타입 관련 lint 경고

---

## H2: 파일 크기 검증

### 시나리오 2.1: 정상 크기 파일 읽기

**Given** 1MB 크기의 설정 파일이 존재하는 환경에서
**When** Hook이 파일을 읽을 때
**Then** 정상적으로 파일 내용이 반환되어야 한다

```python
def test_read_normal_file():
    # Given
    file_path = create_test_file(size_mb=1)

    # When
    content = safe_read_file(file_path)

    # Then
    assert content is not None
    assert len(content) > 0
```

### 시나리오 2.2: 크기 초과 파일 거부

**Given** 15MB 크기의 로그 파일이 존재하는 환경에서
**When** Hook이 10MB 제한으로 파일을 읽으려 할 때
**Then** `ValueError`가 발생하고 ERROR 로그가 기록되어야 한다

```python
def test_reject_large_file():
    # Given
    file_path = create_test_file(size_mb=15)

    # When & Then
    with pytest.raises(ValueError, match="File size exceeds 10MB limit"):
        safe_read_file(file_path)

    # 로그 검증
    assert_log_entry(
        level="ERROR",
        message="File size exceeds limit",
        file_size_mb=15
    )
```

### Edge Cases

**EC-2.1:** 정확히 10MB인 파일 (경계값)
```python
def test_exact_limit_file():
    # Given
    file_path = create_test_file(size_mb=10)

    # When
    content = safe_read_file(file_path)

    # Then - 경계값은 허용
    assert content is not None
```

**EC-2.2:** 크기 확인 불가능한 파일 (권한 없음)
```python
def test_unreadable_file():
    # Given
    file_path = Path("/root/.ssh/private_key")

    # When & Then
    with pytest.raises(PermissionError):
        safe_read_file(file_path)
```

**EC-2.3:** 빈 파일
```python
def test_empty_file():
    # Given
    file_path = create_test_file(size_mb=0)

    # When
    content = safe_read_file(file_path)

    # Then
    assert content == ""
```

### 성공 기준

- [ ] 모든 파일 읽기가 크기 제한 준수
- [ ] 크기 초과 시 명확한 에러 메시지
- [ ] 경계값 (10MB) 올바르게 처리
- [ ] 파일 크기 확인 관련 테스트 커버리지 90%+

---

## H3: 원자적 파일 연산

### 시나리오 3.1: 정상 상태 파일 업데이트

**Given** 상태 파일이 존재하는 환경에서
**When** Hook이 새로운 상태를 원자적으로 쓸 때
**Then** 임시 파일이 생성되고 fsync 후 rename이 완료되어야 한다

```python
def test_atomic_write_success():
    # Given
    state_file = Path("/tmp/state.json")
    state_file.write_text('{"version": 1}')

    # When
    atomic_write(state_file, '{"version": 2}')

    # Then
    assert state_file.read_text() == '{"version": 2}'
    assert not Path(f"{state_file}.tmp").exists()  # 임시 파일 정리
```

### 시나리오 3.2: 쓰기 중단 시 원본 보존

**Given** 상태 파일이 존재하는 환경에서
**When** 쓰기 작업이 중단될 때 (디스크 부족, 시스템 오류)
**Then** 원본 파일이 손상되지 않아야 한다

```python
def test_write_failure_preserves_original():
    # Given
    state_file = Path("/tmp/state.json")
    state_file.write_text('{"version": 1}')

    # When - 쓰기 실패 시뮬레이션
    with patch("builtins.open", side_effect=OSError("Disk full")):
        with pytest.raises(OSError):
            atomic_write(state_file, '{"version": 2}')

    # Then - 원본 보존
    assert state_file.read_text() == '{"version": 1}'
```

### Edge Cases

**EC-3.1:** 동시 쓰기 (경쟁 조건)
```python
def test_concurrent_writes():
    # Given
    state_file = Path("/tmp/state.json")

    # When - 두 프로세스가 동시에 쓰기 시도
    def write_worker(value: str):
        atomic_write(state_file, f'{{"value": "{value}"}}')

    with ThreadPoolExecutor(max_workers=2) as executor:
        futures = [executor.submit(write_worker, v) for v in ["a", "b"]]
        wait(futures)

    # Then - 하나의 유효한 상태만 존재
    content = state_file.read_text()
    assert content in ['{"value": "a"}', '{"value": "b"}']
    assert is_valid_json(content)  # 손상되지 않음
```

**EC-3.2:** Rename 실패 시 롤백
```python
def test_rename_failure_rollback():
    # Given
    state_file = Path("/tmp/state.json")

    # When - rename 실패 시뮬레이션
    with patch("pathlib.Path.rename", side_effect=OSError("Cross-device link")):
        with pytest.raises(OSError):
            atomic_write(state_file, '{"version": 2}')

    # Then - 임시 파일 정리
    assert not Path(f"{state_file}.tmp").exists()
```

**EC-3.3:** 파일 시스템에서 원자적 연산 미지원
```python
def test_non_atomic_filesystem():
    # Given - 네트워크 파일 시스템 (NFS)
    state_file = Path("/nfs/share/state.json")

    # When
    atomic_write(state_file, '{"version": 2}')

    # Then - 가능한 원자적 시도, 실패 시 명확한 에러
    # (구현에 따라 다름)
```

### 성공 기준

- [ ] 모든 상태 파일 쓰기가 원자적임
- [ ] 쓰기 실패 시 원본 손상 0건
- [ ] 임시 파일이 정상적으로 정리
- [ ] 동시 실행 테스트 통과
- [ ] 원자적 연산 테스트 커버리지 90%+

---

## H4: 구조화된 로깅

### 시나리오 4.1: Hook 실행 로그 출력

**Given** Hook이 실행되는 환경에서
**When** Hook이 시작될 때
**Then** INFO 레벨의 구조화된 JSON 로그가 출력되어야 한다

```python
def test_hook_start_logging():
    # Given
    hook_name = "session_start__show_project_info"

    # When
    with capture_logs() as logs:
        run_hook(hook_name)

    # Then
    assert logs[0] == {
        "level": "INFO",
        "hook_name": hook_name,
        "message": "Hook execution started",
        "timestamp": "<iso8601>",
        "execution_id": "<uuid>"
    }
```

### 시나리오 4.2: 에러 발생 시 스택 트레이스 로깅

**Given** Hook 실행 중 에러가 발생하는 환경에서
**When** 예외가 catch될 때
**Then** ERROR 레벨 로그와 스택 트레이스가 포함되어야 한다

```python
def test_error_logging():
    # Given
    hook_name = "failing_hook"

    # When - 에러 발생
    with capture_logs() as logs:
        try:
            raise ValueError("Hook failed")
        except Exception as e:
            log_error(hook_name, e)

    # Then
    assert logs[-1] == {
        "level": "ERROR",
        "hook_name": hook_name,
        "message": "Hook execution failed",
        "error_type": "ValueError",
        "error_message": "Hook failed",
        "stack_trace": "<traceback string>"
    }
```

### Edge Cases

**EC-4.1:** 비-ASCII 문자 로깅
```python
def test_non_ascii_logging():
    # Given
    message = "한글 메시지 🎉"

    # When
    log_info("test_hook", message)

    # Then - UTF-8로 인코딩된 JSON
    assert json.loads(log_output)["message"] == message
```

**EC-4.2:** 대용량 로그 메시지
```python
def test_large_log_message():
    # Given
    large_message = "x" * 10000

    # When
    log_info("test_hook", large_message)

    # Then - 로그가 잘리지 않고 완전히 출력
    assert len(json.loads(log_output)["message"]) == 10000
```

**EC-4.3:** 중첩된 구조 로깅
```python
def test_nested_structured_logging():
    # Given
    context = {
        "user": {"id": 123, "name": "Alice"},
        "session": {"id": "abc-123", "start_time": "2026-01-13T10:00:00"}
    }

    # When
    log_info("test_hook", "User session started", **context)

    # Then - 중첩 구조 보존
    log_entry = json.loads(log_output)
    assert log_entry["context"]["user"]["id"] == 123
```

### 성공 기준

- [ ] 모든 로그가 구조화된 JSON 형식
- [ ] 로그 레벨 (INFO, WARNING, ERROR) 명확히 구분
- [ ] 모든 에러에 스택 트레이스 포함
- [ ] UTF-8 문자 지원
- [ ] print() 문 사용 0건

---

## H5: 코드 중복 제거

### 시나리오 5.1: 헬퍼 함수 재사용

**Given** 3개의 Hook이 유사한 파일 읽기 패턴을 사용하는 환경에서
**When** 공통 헬퍼 함수로 리팩토링될 때
**Then** 모든 Hook이 헬퍼 함수를 호출하고 중복이 제거되어야 한다

```python
# Before (중복)
def hook1():
    with open("file.txt") as f:
        return f.read()

def hook2():
    with open("file.txt") as f:
        return f.read()

# After (헬퍼 함수 재사용)
def hook1():
    return safe_read_file(Path("file.txt"))

def hook2():
    return safe_read_file(Path("file.txt"))
```

### 시나리오 5.2: 코드 중복 감소 측정

**Given** 리팩토링 전 코드베이스에서
**When** 헬퍼 함수 추출이 완료될 때
**Then** 코드 중복이 30% 이상 감소해야 한다

```python
def test_duplication_reduction():
    # Given - 중복 코드 라인 수 측정
    before_duplication = measure_duplication()

    # When - 리팩토링
    refactor_to_helpers()

    # Then
    after_duplication = measure_duplication()
    reduction_rate = (before_duplication - after_duplication) / before_duplication
    assert reduction_rate >= 0.30  # 30% 이상 감소
```

### Edge Cases

**EC-5.1:** 너무 구체적인 헬퍼 함수 (과도한 추상화)
```python
# Bad - 너무 구체적
def hook1_specific_helper():
    pass

# Good - 범용적
def generic_helper(data: str) -> str:
    return data.strip().lower()
```

**EC-5.2:** 헬퍼 함수 순환 의존
```python
# Bad - 순환 의존
# helper_a.py: from helper_b import func_b
# helper_b.py: from helper_a import func_a

# Good - 단방향 의존
# hook files → helper modules
```

**EC-5.3:** 헬퍼 함수 테스트 커버리지
```python
def test_helper_function_coverage():
    # Given
    helpers = discover_helper_functions()

    # When
    for helper in helpers:
        coverage = calculate_test_coverage(helper)

        # Then - 각 헬퍼 함수의 커버리지 90%+
        assert coverage >= 0.90
```

### 성공 기준

- [ ] 코드 중복 30% 이상 감소
- [ ] 헬퍼 함수 재사용률 80%+
- [ ] 헬퍼 함수 평균 길이 20줄 이내
- [ ] 순환 의존성 0건
- [ ] 헬퍼 함수 테스트 커버리지 90%+

---

## 통합 테스트 시나리오

### IT-1: 전체 Hook 실행 파이프라인

**Given** 모든 품질 개선이 적용된 Hook 시스템에서
**When** 세션 시작 Hook이 실행될 때
**Then** 다음 조건이 모두 충족되어야 한다

```python
def test_full_hook_pipeline():
    # Given
    hook_name = "session_start__show_project_info"

    # When
    result = execute_hook(hook_name)

    # Then
    # 1. 타입 안전성
    assert isinstance(result, str | None)

    # 2. 로그 출력
    assert log_contains("INFO", hook_name, "started")

    # 3. 파일 연산 안전성
    assert no_file_size_violations()
    assert no_file_corruption()

    # 4. 실행 시간
    assert execution_time < 5.0  # 5초 이내
```

### IT-2: 동시 Hook 실행

**Given** 여러 Hook이 동시에 실행되는 환경에서
**When** 세션 시작 및 종료 Hook이 병렬 실행될 때
**Then** 데이터 손상이나 경쟁 조건이 발생하지 않아야 한다

```python
def test_concurrent_hook_execution():
    # Given
    hooks = [
        "session_start__show_project_info",
        "session_start__load_config",
        "session_end__cleanup"
    ]

    # When - 병렬 실행
    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = [executor.submit(execute_hook, h) for h in hooks]
        results = wait(futures)

    # Then
    assert not any(f.exception() for f in results.done)
    assert no_file_corruption()
    assert no_race_conditions()
```

---

## Performance 테스트

### PT-1: Hook 실행 시간 벤치마크

**Given** 최적화된 Hook 시스템에서
**When** Hook이 100회 실행될 때
**Then** 평균 실행 시간이 2초 이내여야 한다

```python
def test_hook_execution_time():
    # Given
    hook_name = "session_start__show_project_info"
    iterations = 100

    # When
    times = []
    for _ in range(iterations):
        start = time.time()
        execute_hook(hook_name)
        times.append(time.time() - start)

    # Then
    avg_time = sum(times) / len(times)
    assert avg_time < 2.0  # 2초 이내
    assert max(times) < 5.0  # 최대 5초
```

### PT-2: 대용량 파일 처리 성능

**Given** 10MB 크기의 제한에 가까운 파일이 존재하는 환경에서
**When** 파일이 읽힐 때
**Then** 크기 확인 및 읽기가 100ms 이내에 완료되어야 한다

```python
def test_large_file_performance():
    # Given
    file_path = create_test_file(size_mb=9.9)

    # When
    start = time.time()
    content = safe_read_file(file_path)
    elapsed = time.time() - start

    # Then
    assert elapsed < 0.1  # 100ms 이내
    assert content is not None
```

---

## 보안 테스트

### ST-1: 경로 탐색 방지

**Given** 악의적인 파일 경로가 제공되는 환경에서
**When** Hook이 파일을 읽으려 할 때
**Then** 경로 탐색 (path traversal) 공격이 방지되어야 한다

```python
def test_path_traversal_prevention():
    # Given - 악의적인 경로
    malicious_paths = [
        "../../../etc/passwd",
        "/absolute/path/to/sensitive",
        "C:\\Windows\\System32\\config"
    ]

    # When & Then
    for path in malicious_paths:
        with pytest.raises(ValueError, match="Path traversal detected"):
            safe_read_file(Path(path))
```

### ST-2: 로그에 민감 정보 미포함

**Given** 사용자 개인정보를 처리하는 Hook에서
**When** 에러가 발생하여 로깅될 때
**Then** 로그에 비밀번호, 토큰 등 민감 정보가 포함되지 않아야 한다

```python
def test_no_sensitive_data_in_logs():
    # Given - 민감 정보 포함 데이터
    sensitive_data = {
        "password": "super_secret",
        "api_token": "abc123xyz",
        "user": "alice@example.com"
    }

    # When - 에러 로깅
    with capture_logs() as logs:
        try:
            process_sensitive_data(sensitive_data)
        except Exception:
            log_error("hook", Exception("Failed"))

    # Then - 로그에 민정 정보 제거
    log_str = json.dumps(logs)
    assert "super_secret" not in log_str
    assert "abc123xyz" not in log_str
```

---

## 최종 인수 기준

### 품질 게이트 (Quality Gates)

#### TRUST-5 Framework 준수

**Test-first Pillar**
- [ ] 전체 테스트 커버리지 85%+
- [ ] 핵심 모듈 커버리지 90%+
- [ ] 모든 기능에 단위 테스트 존재

**Readable Pillar**
- [ ] ruff linter 통과 (zero warnings)
- [ ] 함수 평균 길이 20줄 이내
- [ ] 명확한 변수 및 함수命名

**Unified Pillar**
- [ ] 일관된 코드 포맷 (black)
- [ ] 일관된 import 순서 (isort)
- [ ] 통합된 로그 형식

**Secured Pillar**
- [ ] 경로 탐색 방지
- [ ] 민감 정보 로깅 제거
- [ ] 파일 크기 검증 (DoS 방지)

**Trackable Pillar**
- [ ] 명확한 커밋 메시지
- [ ] 테스트 실행 추적 가능
- [ ] 로그로부터 실행 추적 가능

### Definition of Done

- [ ] 모든 H1-H5 이슈 해결
- [ ] acceptance 기준 100% 충족
- [ ] mypy strict mode 통과
- [ ] ruff linter zero warnings
- [ ] 테스트 커버리지 85%+
- [ ] 성능 기준 충족 (평균 2초 이내)
- [ ] 보안 테스트 통과
- [ ] 문서화 완료 (docstring)

---

## 다음 단계

```bash
# TDD 실행
/moai:2-run SPEC-HOOK-001

# 문서 동기화
/moai:3-sync SPEC-HOOK-001
```
