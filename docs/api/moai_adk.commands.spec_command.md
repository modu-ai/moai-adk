# moai_adk.commands.spec_command

@FEATURE:SPEC-COMMAND-001 SPEC Command Implementation
@REQ:SPEC-CREATION-001 /moai:1-spec 명령어의 브랜치 스킵 옵션을 포함한 구현

@API:POST-SPEC - SPEC 생성 API 인터페이스
@PERF:CMD-FAST - 명령어 실행 최적화
@SEC:INPUT-MED - 입력 검증 보안 강화

## Functions

### __init__

Initialize SpecCommand

Args:
    project_dir: 프로젝트 디렉토리
    config: 설정 관리자 인스턴스
    skip_branch: 브랜치 생성 스킵 옵션

```python
__init__(self, project_dir, config, skip_branch)
```

### _get_git_strategy

Git 전략 획득

Returns:
    현재 모드에 적합한 Git 전략 인스턴스

```python
_get_git_strategy(self)
```

### _get_current_mode

현재 작업 모드 확인

Returns:
    작업 모드 ('personal' 또는 'team')

```python
_get_current_mode(self)
```

### execute

@TASK:SPEC-EXECUTE-001 SPEC 명령어 실행

Args:
    spec_name: 명세 이름
    description: 명세 설명
    skip_branch: 브랜치 생성 스킵 여부 (None이면 기본값 사용)

Raises:
    ValueError: 유효하지 않은 입력
    GitLockedException: Git 작업 충돌

```python
execute(self, spec_name, description, skip_branch)
```

### execute_with_mode

모드별 실행 전략

Args:
    mode: 실행 모드 (personal/team)
    spec_name: 명세 이름
    description: 명세 설명

```python
execute_with_mode(self, mode, spec_name, description)
```

### _validate_spec_name

명세 이름 검증 및 정규화

Args:
    spec_name: 검증할 명세 이름

Returns:
    정규화된 명세 이름

Raises:
    ValueError: 유효하지 않은 명세 이름

```python
_validate_spec_name(self, spec_name)
```

### _validate_description

설명 검증 및 정규화

Args:
    description: 검증할 설명

Returns:
    정규화된 설명

Raises:
    ValueError: 유효하지 않은 설명

```python
_validate_description(self, description)
```

### _create_spec_file

@TASK:SPEC-FILE-CREATE-001 SPEC 파일 생성

Args:
    spec_name: 명세 이름
    description: 명세 설명

```python
_create_spec_file(self, spec_name, description)
```

### _generate_spec_content

SPEC 파일 내용 생성 (간결한 버전)

Args:
    spec_name: 명세 이름
    description: 명세 설명

Returns:
    생성된 SPEC 내용

```python
_generate_spec_content(self, spec_name, description)
```

### _get_current_timestamp

현재 타임스탬프 반환

Returns:
    포맷된 현재 시간

```python
_get_current_timestamp(self)
```

### _should_create_branch

브랜치 생성 여부 결정

Returns:
    브랜치를 생성해야 하면 True, 아니면 False

```python
_should_create_branch(self)
```

### _execute_git_workflow

@TASK:SPEC-GIT-WORKFLOW-001 Git 워크플로우 실행

Args:
    spec_name: 명세 이름 (브랜치명에 사용)

```python
_execute_git_workflow(self, spec_name)
```

### _log_execution_start

실행 시작 로깅

```python
_log_execution_start(self, spec_name, description)
```

### _log_execution_success

실행 성공 로깅

```python
_log_execution_success(self, spec_name)
```

### _log_execution_error

실행 오류 로깅

```python
_log_execution_error(self, spec_name, error)
```

### get_command_status

명령어 상태 정보 반환

Returns:
    현재 명령어 상태 정보 딕셔너리

```python
get_command_status(self)
```

## Classes

### SpecCommand

@TASK:SPEC-MAIN-001 개선된 SPEC 명령어 - 브랜치 스킵 옵션 및 사용자 경험 향상

TRUST 원칙 적용:
- T: 테스트 가능한 구조 설계
- R: 명확한 사용자 피드백
- U: Git 전략 패턴 활용
- S: 입력 검증 및 에러 처리
- T: 상세한 실행 추적
