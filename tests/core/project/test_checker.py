"""Tests for checker module (SystemChecker)."""

import shutil
import subprocess
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.project.checker import (
    SystemChecker,
    check_environment,
    get_permission_fix_message,
    get_platform_specific_message,
)


@pytest.fixture
def system_checker():
    """Create SystemChecker instance."""
    return SystemChecker()


class TestSystemChecker:
    """Test SystemChecker functionality."""

    def test_check_all_required_tools(self, system_checker):
        """Test checking all required tools."""
        result = system_checker.check_all()

        # Should check git and python
        assert "git" in result
        assert "python" in result
        assert isinstance(result["git"], bool)
        assert isinstance(result["python"], bool)

    def test_check_all_optional_tools(self, system_checker):
        """Test checking optional tools."""
        result = system_checker.check_all()

        # Should check optional tools
        assert "gh" in result
        assert "docker" in result
        assert isinstance(result["gh"], bool)
        assert isinstance(result["docker"], bool)

    def test_check_tool_empty_command(self, system_checker):
        """Test checking tool with empty command."""
        result = system_checker._check_tool("")

        assert result is False

    def test_check_tool_exception(self, system_checker):
        """Test checking tool that raises exception."""
        result = system_checker._check_tool("invalid tool name with spaces and special chars")

        # Should handle exception gracefully
        assert isinstance(result, bool)

    def test_check_language_tools_none(self, system_checker):
        """Test checking tools when language is None."""
        result = system_checker.check_language_tools(None)

        # Should return empty dict
        assert result == {}

    def test_check_language_tools_unsupported(self, system_checker):
        """Test checking tools for unsupported language."""
        result = system_checker.check_language_tools("unsupported_language")

        # Should return empty dict
        assert result == {}

    def test_check_language_tools_python(self, system_checker):
        """Test checking Python tools."""
        result = system_checker.check_language_tools("python")

        # Should check Python tools
        assert "python3" in result
        assert "pip" in result
        assert isinstance(result["python3"], bool)

    def test_check_language_tools_case_insensitive(self, system_checker):
        """Test language name is case-insensitive."""
        result1 = system_checker.check_language_tools("Python")
        result2 = system_checker.check_language_tools("PYTHON")

        assert "python3" in result1
        assert "python3" in result2

    def test_check_language_tools_typescript(self, system_checker):
        """Test checking TypeScript tools."""
        result = system_checker.check_language_tools("typescript")

        # Should check TypeScript tools
        assert "node" in result
        assert "npm" in result

    def test_check_language_tools_all_categories(self, system_checker):
        """Test that all tool categories are checked."""
        result = system_checker.check_language_tools("python")

        # Should include required, recommended, and optional tools
        # Python has: python3, pip (required), pytest, mypy, ruff (recommended), black, pylint (optional)
        assert len(result) >= 3  # At least some tools from each category

    def test_is_tool_available(self, system_checker):
        """Test tool availability check."""
        # Python should be available (we're running tests with it)
        assert system_checker._is_tool_available("python3") is True

        # Nonexistent tool should not be available
        assert system_checker._is_tool_available("nonexistent_tool_xyz") is False

    def test_get_tool_version_none(self, system_checker):
        """Test getting version of None."""
        result = system_checker.get_tool_version(None)

        assert result is None

    def test_get_tool_version_unavailable(self, system_checker):
        """Test getting version of unavailable tool."""
        result = system_checker.get_tool_version("nonexistent_tool_xyz")

        assert result is None

    def test_get_tool_version_timeout(self, system_checker):
        """Test timeout handling in version check."""
        with patch("subprocess.run", side_effect=subprocess.TimeoutExpired("cmd", 2)):
            result = system_checker.get_tool_version("python3")

            assert result is None

    def test_get_tool_version_os_error(self, system_checker):
        """Test OS error handling in version check."""
        with patch("subprocess.run", side_effect=OSError("error")):
            result = system_checker.get_tool_version("python3")

            assert result is None

    def test_get_tool_version_success(self, system_checker):
        """Test successful version retrieval."""
        result = system_checker.get_tool_version("python3")

        if shutil.which("python3"):
            # If python3 is available, should return version
            assert result is not None
            assert isinstance(result, str)
        else:
            # If not available, should return None
            assert result is None

    def test_get_tool_version_non_zero_returncode(self, system_checker):
        """Test version check with non-zero return code."""
        mock_result = MagicMock()
        mock_result.returncode = 1
        mock_result.stdout = ""

        with patch("subprocess.run", return_value=mock_result):
            result = system_checker.get_tool_version("python3")

            assert result is None

    def test_get_tool_version_empty_stdout(self, system_checker):
        """Test version check with empty stdout."""
        mock_result = MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = ""

        with patch("subprocess.run", return_value=mock_result):
            result = system_checker.get_tool_version("python3")

            assert result is None

    def test_extract_version_line(self, system_checker):
        """Test version line extraction."""
        version_output = "Python 3.11.5\nOther info\nMore lines"

        result = system_checker._extract_version_line(version_output)

        assert result == "Python 3.11.5"

    def test_extract_version_line_single_line(self, system_checker):
        """Test version extraction from single line."""
        version_output = "Python 3.11.5"

        result = system_checker._extract_version_line(version_output)

        assert result == "Python 3.11.5"

    def test_extract_version_line_with_whitespace(self, system_checker):
        """Test version extraction with whitespace."""
        version_output = "  Python 3.11.5  \n"

        result = system_checker._extract_version_line(version_output)

        assert result == "Python 3.11.5"


class TestEnvironmentCheck:
    """Test environment check functions."""

    def test_check_environment(self):
        """Test overall environment check."""
        result = check_environment()

        # Should check all required items
        assert "Python >= 3.11" in result
        assert "Git installed" in result
        assert "Project structure (.moai/)" in result
        assert "Config file (.moai/config/config.json)" in result

        # All values should be boolean
        for value in result.values():
            assert isinstance(value, bool)

    @patch("sys.version_info", (3, 11, 0))
    def test_check_environment_python_version_ok(self):
        """Test environment check with Python 3.11."""
        result = check_environment()

        assert result["Python >= 3.11"] is True

    @patch("sys.version_info", (3, 10, 0))
    def test_check_environment_python_version_old(self):
        """Test environment check with old Python."""
        result = check_environment()

        assert result["Python >= 3.11"] is False


class TestPlatformSpecificMessages:
    """Test platform-specific message functions."""

    def test_get_platform_specific_message_unix(self):
        """Test platform-specific message on Unix."""
        with patch("platform.system", return_value="Darwin"):
            result = get_platform_specific_message("chmod 755", "Check permissions")

            assert result == "chmod 755"

    def test_get_platform_specific_message_linux(self):
        """Test platform-specific message on Linux."""
        with patch("platform.system", return_value="Linux"):
            result = get_platform_specific_message("chmod 755", "Check permissions")

            assert result == "chmod 755"

    def test_get_platform_specific_message_windows(self):
        """Test platform-specific message on Windows."""
        with patch("platform.system", return_value="Windows"):
            result = get_platform_specific_message("chmod 755", "Check permissions")

            assert result == "Check permissions"

    def test_get_permission_fix_message_unix(self):
        """Test permission fix message on Unix."""
        with patch("platform.system", return_value="Darwin"):
            result = get_permission_fix_message("/path/to/dir")

            assert "chmod 755 /path/to/dir" in result

    def test_get_permission_fix_message_windows(self):
        """Test permission fix message on Windows."""
        with patch("platform.system", return_value="Windows"):
            result = get_permission_fix_message("C:\\path\\to\\dir")

            assert "administrator" in result or "permissions" in result


class TestLanguageToolsComprehensive:
    """Test comprehensive language tool configurations."""

    def test_all_supported_languages(self, system_checker):
        """Test that all configured languages can be checked."""
        languages = [
            "python",
            "typescript",
            "javascript",
            "java",
            "go",
            "rust",
            "dart",
            "swift",
            "kotlin",
            "csharp",
            "php",
            "ruby",
            "elixir",
            "scala",
            "clojure",
            "haskell",
            "c",
            "cpp",
            "lua",
            "ocaml",
        ]

        for language in languages:
            result = system_checker.check_language_tools(language)

            # Should return non-empty dict for each language
            assert isinstance(result, dict)
            # Each language should have at least one required tool
            assert len(result) > 0

    def test_language_tools_structure(self, system_checker):
        """Test that language tools have proper structure."""
        # Test Python as example
        tools_config = system_checker.LANGUAGE_TOOLS["python"]

        # Should have all categories
        assert "required" in tools_config
        assert "recommended" in tools_config
        assert "optional" in tools_config

        # Each category should be a list
        assert isinstance(tools_config["required"], list)
        assert isinstance(tools_config["recommended"], list)
        assert isinstance(tools_config["optional"], list)

        # Required tools should not be empty
        assert len(tools_config["required"]) > 0
