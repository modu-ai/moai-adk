# moai_adk.core.git_lock_manager

Git Lock Manager

Git 작업 동시 실행 방지를 위한 잠금 관리 시스템

@FEATURE:GIT-LOCK-001 - 동시 Git 작업 방지
@PERF:LOCK-100MS - 잠금 성능 최적화 (100ms 이내 응답)
@SEC:LOCK-MED - 잠금 파일 보안 강화

## Functions

### __init__

Initialize GitLockManager

Args:
    project_dir: 프로젝트 루트 디렉토리
    lock_dir: 잠금 파일이 저장될 디렉토리 경로 (project_dir 기준)

```python
__init__(self, project_dir, lock_dir)
```

### _ensure_lock_dir

잠금 디렉토리가 존재하지 않으면 생성

강화된 에러 처리 및 권한 검사 포함 (@SEC:LOCK-MED)

```python
_ensure_lock_dir(self)
```

### _get_lock_info

잠금 파일 정보 획득

캐시된 정보를 사용하여 성능 최적화 (@PERF:LOCK-100MS)

Returns:
    잠금 파일 정보 딕셔너리 또는 None

```python
_get_lock_info(self)
```

### _parse_legacy_lock_format

레거시 잠금 파일 형식 파싱

Args:
    content: 잠금 파일 내용

Returns:
    파싱된 정보 딕셔너리

```python
_parse_legacy_lock_format(self, content)
```

### _is_process_running

프로세스 실행 상태 확인

Args:
    pid: 프로세스 ID

Returns:
    프로세스가 실행 중이면 True

```python
_is_process_running(self, pid)
```

### _is_lock_valid

잠금 파일 유효성 검사

Args:
    lock_info: 잠금 파일 정보

Returns:
    유효한 잠금이면 True

```python
_is_lock_valid(self, lock_info)
```

### _cleanup_stale_lock

무효한 잠금 파일 정리

자동 정리 기능 (@FEATURE:AUTO-CLEANUP-001)

```python
_cleanup_stale_lock(self)
```

### is_locked

현재 잠금 상태 확인

성능 최적화된 잠금 상태 검사 (@PERF:LOCK-100MS)

Returns:
    잠금 파일이 존재하고 유효하면 True, 아니면 False

```python
is_locked(self)
```

### release_lock

잠금 파일 삭제

강화된 잠금 해제 메커니즘 (@SEC:LOCK-MED)

```python
release_lock(self)
```

### _create_lock_info

잠금 정보 생성

Returns:
    잠금 파일에 저장할 정보 딕셔너리

```python
_create_lock_info(self)
```

### acquire_lock

잠금 획득

컨텍스트 매니저로 사용하면 자동으로 잠금이 해제되고,
직접 호출하면 검증만 수행합니다.

Args:
    wait: 잠금 대기 여부
    timeout: 대기 시간 (초)

Returns:
    contextmanager 또는 None

Raises:
    GitLockedException: 잠금 획득에 실패한 경우

```python
acquire_lock(self, wait, timeout)
```

### _acquire_lock_context

실제 잠금 획득 컨텍스트 매니저

성능 최적화된 대기 메커니즘 (@PERF:LOCK-100MS)

```python
_acquire_lock_context(self, wait, timeout)
```

### acquire_lock_direct

잠금 직접 획득 (컨텍스트 매니저 없이)

Args:
    wait: 잠금 대기 여부
    timeout: 대기 시간 (초)

Returns:
    bool: 잠금 획득 성공 여부

Raises:
    GitLockedException: 잠금 획득에 실패한 경우

```python
acquire_lock_direct(self, wait, timeout)
```

### get_lock_status

잠금 상태 정보 반환

모니터링 및 디버깅을 위한 상세 정보 제공

Returns:
    잠금 상태 정보 딕셔너리

```python
get_lock_status(self)
```

## Classes

### GitLockManager

Git 작업 동시 실행 방지를 위한 잠금 관리자

.moai/locks/git.lock 파일을 사용하여 Git 작업의 동시 실행을 방지합니다.

Features:
- 성능 최적화된 잠금 메커니즘 (@PERF:LOCK-100MS)
- 강화된 에러 처리 및 복구 (@SEC:LOCK-MED)
- 구조화된 로깅 (@TASK:LOG-001)
- 자동 정리 및 모니터링 (@FEATURE:AUTO-CLEANUP-001)
