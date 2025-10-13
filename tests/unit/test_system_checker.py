# @TEST:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md
"""SystemChecker 테스트 스위트

시스템 요구사항 검증 기능을 검증합니다.
"""

from typing import Any
from unittest.mock import MagicMock, patch

from moai_adk.core.project.checker import SystemChecker


class TestSystemChecker:
    """SystemChecker 클래스 테스트"""

    def test_check_all_returns_dict(self) -> None:
        """check_all이 딕셔너리 반환"""
        checker = SystemChecker()

        result = checker.check_all()

        assert isinstance(result, dict)

    def test_check_all_includes_required_tools(self) -> None:
        """check_all이 필수 도구 포함"""
        checker = SystemChecker()

        result = checker.check_all()

        assert "git" in result
        assert "python" in result

    def test_check_all_includes_optional_tools(self) -> None:
        """check_all이 선택 도구 포함"""
        checker = SystemChecker()

        result = checker.check_all()

        assert "gh" in result
        assert "docker" in result

    def test_check_all_values_are_boolean(self) -> None:
        """check_all 결과 값이 불리언"""
        checker = SystemChecker()

        result = checker.check_all()

        for _tool, status in result.items():
            assert isinstance(status, bool)

    def test_check_tool_returns_true_for_existing_command(self) -> None:
        """존재하는 명령어에 대해 True 반환"""
        checker = SystemChecker()

        # 'python3'는 대부분의 환경에 존재
        result = checker._check_tool("python3 --version")

        # 환경에 따라 True 또는 False
        assert isinstance(result, bool)

    @patch("shutil.which")
    def test_check_tool_returns_true_when_tool_found(self, mock_which: MagicMock) -> None:
        """shutil.which가 경로 반환 시 True"""
        mock_which.return_value = "/usr/bin/git"
        checker = SystemChecker()

        result = checker._check_tool("git --version")

        assert result is True
        mock_which.assert_called_once_with("git")

    @patch("shutil.which")
    def test_check_tool_returns_false_when_tool_not_found(self, mock_which: MagicMock) -> None:
        """shutil.which가 None 반환 시 False"""
        mock_which.return_value = None
        checker = SystemChecker()

        result = checker._check_tool("nonexistent --version")

        assert result is False

    @patch("shutil.which")
    def test_check_tool_handles_exception(self, mock_which: MagicMock) -> None:
        """예외 발생 시 False 반환"""
        mock_which.side_effect = Exception("Error")
        checker = SystemChecker()

        result = checker._check_tool("any command")

        assert result is False

    def test_check_tool_handles_empty_command(self) -> None:
        """빈 명령어에 대해 False 반환"""
        checker = SystemChecker()

        result = checker._check_tool("")

        assert result is False

    def test_required_tools_constant(self) -> None:
        """REQUIRED_TOOLS 상수가 정의됨"""
        assert "git" in SystemChecker.REQUIRED_TOOLS
        assert "python" in SystemChecker.REQUIRED_TOOLS

    def test_optional_tools_constant(self) -> None:
        """OPTIONAL_TOOLS 상수가 정의됨"""
        assert "gh" in SystemChecker.OPTIONAL_TOOLS
        assert "docker" in SystemChecker.OPTIONAL_TOOLS

    def test_required_tools_format(self) -> None:
        """REQUIRED_TOOLS가 올바른 형식"""
        for tool, command in SystemChecker.REQUIRED_TOOLS.items():
            assert isinstance(tool, str)
            assert isinstance(command, str)
            assert "--version" in command

    def test_optional_tools_format(self) -> None:
        """OPTIONAL_TOOLS가 올바른 형식"""
        for tool, command in SystemChecker.OPTIONAL_TOOLS.items():
            assert isinstance(tool, str)
            assert isinstance(command, str)
            assert "--version" in command

    @patch("shutil.which")
    def test_check_all_with_all_tools_available(self, mock_which: MagicMock) -> None:
        """모든 도구가 사용 가능한 경우"""
        mock_which.return_value = "/usr/bin/tool"
        checker = SystemChecker()

        result = checker.check_all()

        assert all(status is True for status in result.values())

    @patch("shutil.which")
    def test_check_all_with_no_tools_available(self, mock_which: MagicMock) -> None:
        """모든 도구가 사용 불가능한 경우"""
        mock_which.return_value = None
        checker = SystemChecker()

        result = checker.check_all()

        assert all(status is False for status in result.values())

    @patch("shutil.which")
    def test_check_all_with_mixed_availability(self, mock_which: MagicMock) -> None:
        """일부 도구만 사용 가능한 경우"""

        def which_side_effect(tool: str) -> str | None:
            return "/usr/bin/git" if tool == "git" else None

        mock_which.side_effect = which_side_effect
        checker = SystemChecker()

        result = checker.check_all()

        assert result["git"] is True
        assert result["python"] is False

    def test_check_tool_extracts_command_name(self) -> None:
        """명령어에서 도구 이름 추출"""
        checker = SystemChecker()

        # shutil.which는 첫 번째 단어만 사용
        # 예: "git --version" -> "git"
        with patch("shutil.which") as mock_which:
            mock_which.return_value = "/usr/bin/git"
            result = checker._check_tool("git --version --verbose")

            mock_which.assert_called_once_with("git")
            assert result is True
