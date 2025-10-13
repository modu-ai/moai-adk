# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_system_checker.py
"""시스템 요구사항 검증 모듈"""

import shutil


class SystemChecker:
    """시스템 요구사항 검증

    필수/선택 도구의 설치 여부를 확인합니다.
    """

    REQUIRED_TOOLS: dict[str, str] = {
        "git": "git --version",
        "python": "python3 --version",
    }

    OPTIONAL_TOOLS: dict[str, str] = {
        "gh": "gh --version",
        "docker": "docker --version",
    }

    def check_all(self) -> dict[str, bool]:
        """모든 시스템 요구사항 검증

        Returns:
            도구명: 설치 여부 딕셔너리
        """
        results = {}

        for tool, command in {**self.REQUIRED_TOOLS, **self.OPTIONAL_TOOLS}.items():
            results[tool] = self._check_tool(command)

        return results

    def _check_tool(self, command: str) -> bool:
        """개별 도구 확인

        Args:
            command: 확인할 명령어

        Returns:
            설치 여부
        """
        try:
            # shutil.which()를 사용하면 더 안전
            tool_name = command.split()[0] if command else ""
            if not tool_name:
                return False
            return shutil.which(tool_name) is not None
        except Exception:
            return False
