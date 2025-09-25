# moai_adk.core.exceptions

@FEATURE:EXCEPTIONS-001 MoAI-ADK Core Exceptions

Git 잠금 및 전략 관련 예외 클래스들

## Functions

### __init__

```python
__init__(self, mode, supported_modes)
```

## Classes

### GitLockedException

@TASK:GIT-LOCKED-001 Git 작업 잠금 예외

다른 Git 작업이 진행 중일 때 발생하는 예외

### GitModeException

Git 모드 설정 예외

올바르지 않은 Git 모드가 설정되었을 때 발생하는 예외
