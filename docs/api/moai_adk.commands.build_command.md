# moai_adk.commands.build_command

@FEATURE:BUILD-COMMAND-001 BUILD Command Implementation
@REQ:TDD-AUTOMATION-001 /moai:2-build 명령어의 Git 잠금 확인 로직 구현

@API:POST-BUILD - BUILD 실행 API 인터페이스
@PERF:TDD-FAST - TDD 프로세스 실행 최적화
@SEC:LOCK-MED - Git 잠금 보안 강화

## Functions

### __init__

Initialize BuildCommand

Args:
    project_dir: 프로젝트 디렉토리
    config: 설정 관리자 인스턴스

```python
__init__(self, project_dir, config)
```

### execute

@TASK:BUILD-EXECUTE-001 BUILD 명령어 실행

Args:
    spec_name: 빌드할 명세 이름
    wait_for_lock: 잠금 대기 여부

Raises:
    GitLockedException: 잠금 파일이 존재하고 대기하지 않는 경우
    ValueError: 유효하지 않은 입력

```python
execute(self, spec_name, wait_for_lock)
```

### execute_with_lock_check

잠금 확인 후 실행

Args:
    spec_name: 빌드할 명세 이름
    wait_for_lock: 잠금 대기 여부

```python
execute_with_lock_check(self, spec_name, wait_for_lock)
```

### _validate_spec_name

명세 이름 검증

Args:
    spec_name: 검증할 명세 이름

Returns:
    검증된 명세 이름

Raises:
    ValueError: 유효하지 않은 명세 이름

```python
_validate_spec_name(self, spec_name)
```

### _execute_tdd_process

@TASK:TDD-PROCESS-001 TDD 프로세스 실행 (RED-GREEN-REFACTOR)

성능 최적화된 TDD 사이클 (@PERF:TDD-FAST)

Args:
    spec_name: 빌드할 명세 이름

```python
_execute_tdd_process(self, spec_name)
```

### _execute_red_phase

@TASK:TDD-RED-001 RED Phase: 실패하는 테스트 작성

Args:
    spec_name: 명세 이름

```python
_execute_red_phase(self, spec_name)
```

### _execute_green_phase

@TASK:TDD-GREEN-001 GREEN Phase: 최소 구현

Args:
    spec_name: 명세 이름

```python
_execute_green_phase(self, spec_name)
```

### _execute_refactor_phase

@TASK:TDD-REFACTOR-001 REFACTOR Phase: 코드 리팩터링

Args:
    spec_name: 명세 이름

```python
_execute_refactor_phase(self, spec_name)
```

### _write_failing_tests

실패하는 테스트 작성

Args:
    spec_name: 명세 이름

```python
_write_failing_tests(self, spec_name)
```

### _implement_minimum_code

최소 구현

Args:
    spec_name: 명세 이름

```python
_implement_minimum_code(self, spec_name)
```

### _refactor_code

코드 리팩터링

Args:
    spec_name: 명세 이름

```python
_refactor_code(self, spec_name)
```

### _log_execution_start

실행 시작 로깅

```python
_log_execution_start(self, spec_name, wait_for_lock)
```

### _log_execution_success

실행 성공 로깅

```python
_log_execution_success(self, spec_name)
```

### _log_execution_error

실행 오류 로깅

```python
_log_execution_error(self, spec_name, error_message)
```

### get_build_status

빌드 상태 정보 반환

Returns:
    현재 빌드 상태 정보 딕셔너리

```python
get_build_status(self)
```

## Classes

### BuildCommand

@TASK:BUILD-MAIN-001 개선된 BUILD 명령어 - Git 잠금 확인 및 TDD 프로세스 최적화

TRUST 원칙 적용:
- T: TDD 사이클 엄격 준수
- R: 명확한 빌드 단계 피드백
- U: 잠금 시스템 통합 설계
- S: 안전한 동시 작업 방지
- T: 상세한 빌드 과정 추적
