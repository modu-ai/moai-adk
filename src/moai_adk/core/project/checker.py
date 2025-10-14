# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""시스템 요구사항 검증 모듈

필수 및 선택 도구의 설치 여부를 확인합니다.
"""

import shutil
import sys
from pathlib import Path


class SystemChecker:
    """시스템 요구사항 검증"""

    REQUIRED_TOOLS: dict[str, str] = {
        "git": "git --version",
        "python": "python3 --version",
    }

    OPTIONAL_TOOLS: dict[str, str] = {
        "gh": "gh --version",
        "docker": "docker --version",
    }

    def check_all(self) -> dict[str, bool]:
        """모든 도구 검증

        Returns:
            도구명: 사용가능 여부 딕셔너리
        """
        result = {}

        # 필수 도구 확인
        for tool, command in self.REQUIRED_TOOLS.items():
            result[tool] = self._check_tool(command)

        # 선택 도구 확인
        for tool, command in self.OPTIONAL_TOOLS.items():
            result[tool] = self._check_tool(command)

        return result

    def _check_tool(self, command: str) -> bool:
        """개별 도구 확인

        Args:
            command: 확인할 명령어 (예: "git --version")

        Returns:
            도구가 사용 가능하면 True
        """
        if not command:
            return False

        try:
            # 명령어에서 도구 이름 추출 (첫 단어)
            tool_name = command.split()[0]
            # shutil.which로 도구 존재 확인
            return shutil.which(tool_name) is not None
        except Exception:
            return False


def check_environment() -> dict[str, bool]:
    """전체 환경 검증 (CLI doctor 명령어용)

    Returns:
        각 체크 항목의 결과 딕셔너리
    """
    return {
        "Python >= 3.13": sys.version_info >= (3, 13),
        "Git installed": shutil.which("git") is not None,
        "Project structure (.moai/)": Path(".moai").exists(),
        "Config file (.moai/config.json)": Path(".moai/config.json").exists(),
    }
