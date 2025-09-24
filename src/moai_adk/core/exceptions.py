"""
@FEATURE:EXCEPTIONS-001 MoAI-ADK Core Exceptions

Git 잠금 및 전략 관련 예외 클래스들
"""


class GitLockedException(Exception):
    """@TASK:GIT-LOCKED-001 Git 작업 잠금 예외

    다른 Git 작업이 진행 중일 때 발생하는 예외
    """

    def __init__(self, message: str = "Git 작업이 이미 진행 중입니다"):
        self.message = message
        super().__init__(self.message)


class GitModeException(Exception):
    """Git 모드 설정 예외

    올바르지 않은 Git 모드가 설정되었을 때 발생하는 예외
    """

    def __init__(self, mode: str, supported_modes: list):
        self.mode = mode
        self.supported_modes = supported_modes
        message = f"지원하지 않는 Git 모드입니다: {mode}. 지원되는 모드: {supported_modes}"
        super().__init__(message)